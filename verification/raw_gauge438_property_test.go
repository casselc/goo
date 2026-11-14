package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"strings"
)

// gauge438Transport supplies independent exporter-shaped rows to the real source
// adapter. It does not replace the adapter's decoding or admission logic.
type gauge438Transport struct {
	schema map[string][]json.RawMessage
	rows   []json.RawMessage
	calls  int
	sql    []string
}

func (d *gauge438Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	d.calls++
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	sql := string(raw)
	d.sql = append(d.sql, sql)
	if r.URL.Query().Get("readonly") != "1" {
		return nil, fmt.Errorf("raw gauge source lost readonly contract")
	}
	var rows []json.RawMessage
	switch {
	case strings.HasPrefix(sql, "DESCRIBE TABLE "):
		if !strings.Contains(sql, "otel_metrics_gauge ") {
			return nil, fmt.Errorf("raw gauge inspected unrelated source")
		}
		rows = d.schema["gauge"]
	case strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow"):
	case strings.Contains(sql, "otel_metrics_gauge"):
		if !strings.Contains(sql, "LIMIT 10001") || !strings.Contains(sql, "fromUnixTimestamp64Milli({p0:Int64})") || !strings.Contains(sql, "fromUnixTimestamp64Milli({p1:Int64})") {
			return nil, fmt.Errorf("raw gauge query lost bounded typed window")
		}
		rows = d.rows
	default:
		return nil, fmt.Errorf("unexpected raw gauge SQL")
	}
	var body strings.Builder
	for _, row := range rows {
		compact, _ := json.Marshal(row)
		body.Write(compact)
		body.WriteByte('\n')
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body.String()))}, nil
}

func gauge438Question() measurements.RawGaugeQuestion {
	return measurements.RawGaugeQuestion{Question: measurements.DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: measurements.RawGaugeInstrument{Kind: "gauge", Scope: "", Metric: "temperature", Unit: "Cel"}}
}
func gauge438Input(value any, flags uint32, index int) map[string]any {
	return map[string]any{"resource_schema_url": "", "resource": map[string]string{"service.name": "api", "deployment.environment.name": "prod", "bank": strconv.Itoa(index % 3)}, "scope": "", "version": "", "schema_url": "", "name": "temperature", "unit": "Cel", "attributes": map[string]string{"sensor": strconv.Itoa(index % 2), "empty": ""}, "start_ns": "0", "end_ns": "1500000000", "value": value, "flags": flags}
}
func gauge438Read(t interface{ Fatal(...any) }, schema map[string][]json.RawMessage, rows []json.RawMessage, q measurements.RawGaugeQuestion) (*measurements.RawGaugeResult, error, *gauge438Transport) {
	driver := &gauge438Transport{schema: schema, rows: rows}
	source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: driver}}
	result, err := source.ReadRawGaugeObservations(context.Background(), q)
	return result, err, driver
}
func TestHegelRawGauge438Conservation(t *testing.T) {
	schema := metricScopeSchema432(t)
	hegel.Test(t, func(tc *hegel.T) {
		size := hegel.Draw(tc, hegel.Integers(1, 30))
		q := gauge438Question()
		rows := []json.RawMessage{}
		states := []string{}
		numbers := []*float64{}
		for i := range size {
			kind := hegel.Draw(tc, hegel.Integers(0, 7))
			var value any
			state := "finite"
			var number *float64
			flags := uint32(0)
			switch kind {
			case 0:
				value = "0"
				n := 0.0
				number = &n
			case 1:
				value = "-0.000000000000001"
				n := -0.000000000000001
				number = &n
			case 2:
				value = "12.5"
				n := 12.5
				number = &n
			case 3:
				value = "NaN"
				state = "nonfinite"
			case 4:
				value = "+Inf"
				state = "nonfinite"
			case 5:
				value = nil
				state = "absent"
			case 6:
				value = "bad"
				state = "invalid"
			case 7:
				value = "7.25"
				flags = 1
				state = "absent"
			}
			row := gauge438Input(value, flags, i)
			encoded, _ := json.Marshal(row)
			rows = append(rows, encoded)
			states = append(states, state)
			numbers = append(numbers, number)
			if hegel.Draw(tc, hegel.Booleans()) {
				rows = append(rows, bytes.Clone(encoded))
				states = append(states, state)
				numbers = append(numbers, number)
			}
		}
		start := time.Now()
		result, err, driver := gauge438Read(tc, schema, rows, q)
		if err != nil || result == nil {
			tc.Fatal("raw bank refused", err)
			return
		}
		if driver.calls != 3 || len(result.Rows) != len(rows) || result.Limit != 10000 || result.Version != "goo/raw-gauge-observations/v1" || !reflect.DeepEqual(result.Question, q.Question) || result.Instrument != q.Instrument {
			tc.Fatal("bounded raw profile lost conservation", driver.calls)
			return
		}
		readAt, err := time.Parse(time.RFC3339Nano, result.ReadAt)
		if err != nil || readAt.Before(start) || readAt.After(time.Now()) {
			tc.Fatal("readAt not acquisition clock", result.ReadAt)
		}
		for i, row := range result.Rows {
			if row.OccurrenceKey != strconv.Itoa(i+1) || !gauge438JSONEqual(row.RawRow, rows[i]) || row.ValueState != states[i] || !reflect.DeepEqual(row.Value, numbers[i]) || row.EndNS != "1500000000" || row.StartNS == nil || *row.StartNS != "0" || row.Stream.Resource["service.instance.id"] != "" {
				tc.Fatal("raw occurrence changed", i, row)
				return
			}
			var source map[string]json.RawMessage
			_ = json.Unmarshal(rows[i], &source)
			if !gauge438JSONEqual(row.RawValue, source["value"]) {
				tc.Fatal("raw scalar changed", i)
			}
			var resource, attrs map[string]string
			_ = json.Unmarshal(source["resource"], &resource)
			_ = json.Unmarshal(source["attributes"], &attrs)
			if !reflect.DeepEqual(row.Stream.Resource, resource) || !reflect.DeepEqual(row.Stream.Attributes, attrs) || row.Stream.Scope != q.Instrument.Scope || row.Stream.Unit != q.Instrument.Unit {
				tc.Fatal("reported labels or empty metadata changed", i)
			}
		}
	}, options(150)...)
}

func TestRawGauge438GoJavaScriptAdmission(t *testing.T) {
	schema := metricScopeSchema432(t)
	q := gauge438Question()
	rows := []json.RawMessage{}
	for i, value := range []any{"0", "-0.000000001", "NaN", "nan", "+Inf", "infinity", "1e400", "1e-400", "1_000.25", "0x1.8p2", "0x1.fffffffffffffp1023", nil, "bad", false, map[string]any{"x": 1}} {
		for _, flags := range []uint32{0, 1, 2} {
			raw, _ := json.Marshal(gauge438Input(value, flags, i))
			rows = append(rows, raw)
		}
	}
	for _, start := range []any{nil, "bad", "01500000000", "1600000000", "9223372036854775808", 123} {
		row := gauge438Input("1", 0, 0)
		row["start_ns"] = start
		raw, _ := json.Marshal(row)
		rows = append(rows, raw)
	}
	for _, flags := range []string{"0.0", "1e0", "4294967295.0000001", "1.0000000000000000000001", "4294967296", "1.5", "-1", "1e5000"} {
		row := gauge438Input("7.25", 0, 0)
		row["flags"] = json.RawMessage(flags)
		raw, _ := json.Marshal(row)
		rows = append(rows, raw)
	}
	row := gauge438Input("1", 0, 0)
	resource := row["resource"].(map[string]string)
	for i := range 70 {
		resource[strconv.Itoa(i)] = strings.Repeat("x", 2100)
	}
	raw, _ := json.Marshal(row)
	rows = append(rows, raw)
	result, err, _ := gauge438Read(t, schema, rows, q)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(map[string]any{"request": q, "result": result})
	script := `import fs from 'node:fs';import {checkedResult} from '../ui/raw-gauge-observations.mjs';const x=JSON.parse(fs.readFileSync(0,'utf8'));const valid=checkedResult(x.result,x.request);if(!valid||valid.rows.length!==x.result.rows.length)throw Error('Go/JS admitted bank diverged');const changes=[r=>r.rows[0].occurrenceKey=r.rows[1].occurrenceKey,r=>r.rows[0].endNS='2000000000',r=>r.rows[0].stream.unit='ms',r=>r.rows[0].value=123,r=>r.readAt='invalid',r=>r.limit=100,r=>r.question.service='other',r=>r.rows[0].stream.resource['service.instance.id']='invented'];let rejected=0;for(const change of changes){const r=structuredClone(x.result);change(r);try{if(checkedResult(r,x.request))throw Error('mutation admitted')}catch(e){if(e.message==='mutation admitted')throw e;rejected++}}if(rejected!==changes.length)throw Error('falsey admission boundary not counted');process.stdout.write(JSON.stringify({rows:valid.rows.length,controls:rejected}));`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Go/JS raw admission failed: %v %s", err, out)
	}
	t.Log(string(out))
	// End/identity failure is atomic in the source adapter; invalid point values
	// above remain visible instead. Unsupported kinds must not acquire a source.
	for _, mutate := range []func(map[string]any){func(r map[string]any) { delete(r, "end_ns") }, func(r map[string]any) { r["end_ns"] = "01500000000" }, func(r map[string]any) { r["end_ns"] = "2000000000" }, func(r map[string]any) { r["attributes"] = nil }, func(r map[string]any) { r["unit"] = "ms" }} {
		row := gauge438Input("1", 0, 0)
		mutate(row)
		raw, _ := json.Marshal(row)
		got, e, _ := gauge438Read(t, schema, []json.RawMessage{rows[0], raw}, q)
		if e == nil || got != nil {
			t.Fatal("malformed bank published partially")
		}
	}
	q.Instrument.Kind = "counter"
	got, e, driver := gauge438Read(t, schema, rows, q)
	if e == nil || got != nil || driver.calls != 0 {
		t.Fatal("non-gauge reached source")
	}
}
