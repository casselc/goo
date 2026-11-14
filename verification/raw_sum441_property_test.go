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
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// Supply exporter-shaped source bytes to the actual source adapter. The driver
// checks family and read boundaries; production code still decodes every row.
type sum441Transport struct {
	schema []json.RawMessage
	rows   []json.RawMessage
	calls  int
	sql    []string
}

func (d *sum441Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	d.calls++
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	sql := string(body)
	d.sql = append(d.sql, sql)
	if r.URL.Query().Get("readonly") != "1" {
		return nil, fmt.Errorf("raw sum lost readonly boundary")
	}
	var rows []json.RawMessage
	switch {
	case strings.HasPrefix(sql, "DESCRIBE TABLE "):
		if !strings.Contains(sql, "otel_metrics_sum ") {
			return nil, fmt.Errorf("sum inspected unrelated family")
		}
		rows = d.schema
	case strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow"):
		if !strings.Contains(sql, "otel_metrics_sum") {
			return nil, fmt.Errorf("sum probed unrelated family")
		}
	case strings.Contains(sql, "FROM otel_metrics_sum"):
		for _, part := range []string{"LIMIT 10001", "fromUnixTimestamp64Milli({p0:Int64})", "fromUnixTimestamp64Milli({p1:Int64})", "AggregationTemporality", "IsMonotonic"} {
			if !strings.Contains(sql, part) {
				return nil, fmt.Errorf("sum lost bounded exact projection: %s", part)
			}
		}
		rows = d.rows
	default:
		return nil, fmt.Errorf("unexpected sum source SQL")
	}
	var out bytes.Buffer
	for _, row := range rows {
		compact, _ := json.Marshal(row)
		out.Write(compact)
		out.WriteByte('\n')
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(out.Bytes()))}, nil
}
func sum441Question(temporality int32, monotonic bool) measurements.RawSumQuestion {
	kind := "counter"
	if temporality == 2 && !monotonic {
		kind = "stock"
	}
	return measurements.RawSumQuestion{Question: measurements.DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: measurements.RawSumInstrument{Kind: kind, Metric: "jobs", Unit: "{job}", Temporality: temporality, Monotonic: monotonic}}
}
func sum441Input(q measurements.RawSumQuestion, value any, flags uint32, index int) map[string]any {
	return map[string]any{"resource_schema_url": "", "resource": map[string]string{"service.name": "api", "deployment.environment.name": "prod", "bank": strconv.Itoa(index % 3)}, "scope": "", "version": "", "schema_url": "", "name": q.Instrument.Metric, "unit": q.Instrument.Unit, "attributes": map[string]string{"queue": strconv.Itoa(index % 2), "empty": ""}, "start_ns": "0", "end_ns": "1500000000", "value": value, "flags": flags, "temporality": q.Instrument.Temporality, "monotonic": q.Instrument.Monotonic}
}
func sum441Read(schema []json.RawMessage, rows []json.RawMessage, q measurements.RawSumQuestion) (*measurements.RawSumResult, error, *sum441Transport) {
	driver := &sum441Transport{schema: schema, rows: rows}
	source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: driver}}
	result, err := source.ReadRawSumObservations(context.Background(), q)
	return result, err, driver
}
func TestHegelRawSum441Conservation(t *testing.T) {
	schema := metricScopeSchema432(t)["counter"]
	hegel.Test(t, func(tc *hegel.T) {
		tempo := hegel.Draw(tc, hegel.SampledFrom([]int32{-1, 0, 1, 2, 7, 2147483647}))
		mono := hegel.Draw(tc, hegel.Booleans())
		q := sum441Question(tempo, mono)
		size := hegel.Draw(tc, hegel.Integers(1, 20))
		var rows []json.RawMessage
		var states []string
		var values []*float64
		for i := range size {
			choice := hegel.Draw(tc, hegel.Integers(0, 6))
			var value any
			state := "finite"
			var number *float64
			flags := uint32(0)
			switch choice {
			case 0:
				value = "0"
				n := 0.0
				number = &n
			case 1:
				value = "-3.25"
				n := -3.25
				number = &n
			case 2:
				value = "12.5"
				n := 12.5
				number = &n
			case 3:
				value = "NaN"
				state = "nonfinite"
			case 4:
				value = nil
				state = "absent"
			case 5:
				value = "bad"
				state = "invalid"
			case 6:
				value = "7.25"
				flags = 1
				state = "absent"
			}
			raw, _ := json.Marshal(sum441Input(q, value, flags, i))
			rows = append(rows, raw)
			states = append(states, state)
			values = append(values, number)
			if hegel.Draw(tc, hegel.Booleans()) {
				rows = append(rows, bytes.Clone(raw))
				states = append(states, state)
				values = append(values, number)
			}
		}
		result, err, driver := sum441Read(schema, rows, q)
		if err != nil || result == nil {
			tc.Fatal("sum bank refused", err, driver.calls, driver.sql)
			return
		}
		if driver.calls != 3 || result.Version != "goo/raw-sum-observations/v1" || result.Limit != 10000 || len(result.Rows) != len(rows) || !reflect.DeepEqual(result.Question, q.Question) || result.Instrument != q.Instrument {
			tc.Fatal("sum profile lost exact acquisition", driver.calls)
			return
		}
		for i, row := range result.Rows {
			if row.OccurrenceKey != strconv.Itoa(i+1) || !bytes.Equal(row.RawRow, rows[i]) || row.Temporality != tempo || row.Monotonic != mono || row.ValueState != states[i] || !reflect.DeepEqual(row.Value, values[i]) || row.EndNS != "1500000000" || row.StartNS == nil || *row.StartNS != "0" {
				tc.Fatal("sum occurrence changed", i, row)
				return
			}
			var source map[string]json.RawMessage
			_ = json.Unmarshal(rows[i], &source)
			if !bytes.Equal(row.RawValue, source["value"]) {
				tc.Fatal("sum raw scalar changed")
			}
			var resource, attrs map[string]string
			_ = json.Unmarshal(source["resource"], &resource)
			_ = json.Unmarshal(source["attributes"], &attrs)
			if !reflect.DeepEqual(row.Stream.Resource, resource) || !reflect.DeepEqual(row.Stream.Attributes, attrs) {
				tc.Fatal("sum labels changed")
			}
			diagnostic := false
			for _, d := range row.Diagnostics {
				if d.Field == "temporality" {
					diagnostic = true
				}
			}
			if diagnostic != (tempo != 1 && tempo != 2) {
				tc.Fatal("unknown temporality lost diagnostic", tempo, row.Diagnostics)
			}
		}
	}, options(150)...)
}
func TestRawSum441GoJavaScriptAdmission(t *testing.T) {
	schema := metricScopeSchema432(t)["counter"]
	var banks []any
	for _, tempo := range []int32{-1, 0, 1, 2, 7, 2147483647} {
		for _, mono := range []bool{false, true} {
			q := sum441Question(tempo, mono)
			var rows []json.RawMessage
			for i, value := range []any{"0", "-3.25", "NaN", nil, "bad", "7.25"} {
				flags := uint32(0)
				if i == 5 {
					flags = 1
				}
				raw, _ := json.Marshal(sum441Input(q, value, flags, i))
				rows = append(rows, raw)
			}
			result, err, _ := sum441Read(schema, rows, q)
			if err != nil {
				t.Fatal(err)
			}
			banks = append(banks, map[string]any{"request": q, "result": result})
		}
	}
	input, _ := json.Marshal(banks)
	script := `import fs from 'node:fs';import {checkedResult} from '../ui/raw-observations.mjs';const banks=JSON.parse(fs.readFileSync(0,'utf8'));let rows=0;for(const b of banks){const r=checkedResult(b.result,b.request);if(r.rows.length!==b.result.rows.length)throw Error('sum admission lost rows');rows+=r.rows.length}const x=banks.find(b=>b.request.instrument.temporality===2&&!b.request.instrument.monotonic);const changes=[r=>r.rows[0].temporality=1,r=>r.rows[0].monotonic=true,r=>delete r.rows[0].rawRow.temporality,r=>r.rows[0].rawRow.monotonic=true,r=>r.version='goo/raw-gauge-observations/v1',r=>r.instrument.kind='counter',r=>r.rows[0].occurrenceKey=r.rows[1].occurrenceKey,r=>r.rows[0].value=123];let controls=0;for(const change of changes){const r=structuredClone(x.result);change(r);let rejected=false;try{checkedResult(r,x.request)}catch{rejected=true}if(!rejected)throw Error('sum semantic mutant admitted');controls++}process.stdout.write(JSON.stringify({banks:banks.length,rows,controls}));`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sum Go/JS admission failed: %v %s", err, out)
	}
	t.Log(string(out))
	q := sum441Question(2, false)
	good, _ := json.Marshal(sum441Input(q, "1", 0, 0))
	for _, mutate := range []func(map[string]any){func(r map[string]any) { delete(r, "temporality") }, func(r map[string]any) { r["temporality"] = "2" }, func(r map[string]any) { r["temporality"] = 1 }, func(r map[string]any) { delete(r, "monotonic") }, func(r map[string]any) { r["monotonic"] = 0 }, func(r map[string]any) { r["monotonic"] = true }} {
		bad := sum441Input(q, "1", 0, 0)
		mutate(bad)
		raw, _ := json.Marshal(bad)
		result, err, _ := sum441Read(schema, []json.RawMessage{good, raw}, q)
		if err == nil || result != nil {
			t.Fatal("malformed sum metadata published partial bank")
		}
	}
	q.Instrument.Kind = "counter"
	result, err, driver := sum441Read(schema, []json.RawMessage{good}, q)
	if err == nil || result != nil || driver.calls != 0 {
		t.Fatal("inconsistent stock/counter identity reached source")
	}
}
