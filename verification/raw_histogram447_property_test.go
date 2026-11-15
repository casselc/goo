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

// Feed independent exporter bytes through the actual bounded source adapter.
// The transport checks the authority boundary rather than projecting results.
type histogram447Transport struct {
	schema, rows []json.RawMessage
	calls        int
	sql          []string
}

func (d *histogram447Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	d.calls++
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	sql := string(raw)
	d.sql = append(d.sql, sql)
	if r.URL.Query().Get("readonly") != "1" {
		return nil, fmt.Errorf("histogram source lost readonly boundary")
	}
	var rows []json.RawMessage
	switch {
	case strings.HasPrefix(sql, "DESCRIBE TABLE "):
		if !strings.Contains(sql, "otel_metrics_histogram ") {
			return nil, fmt.Errorf("histogram inspected another family")
		}
		rows = d.schema
	case strings.HasSuffix(sql, " LIMIT 0 FORMAT JSONEachRow"):
		if !strings.Contains(sql, "otel_metrics_histogram") {
			return nil, fmt.Errorf("histogram probed another family")
		}
	case strings.Contains(sql, "FROM otel_metrics_histogram"):
		for _, part := range []string{"LIMIT 10001", "fromUnixTimestamp64Milli({p0:Int64})", "fromUnixTimestamp64Milli({p1:Int64})", "AggregationTemporality", "BucketCounts", "ExplicitBounds"} {
			if !strings.Contains(sql, part) {
				return nil, fmt.Errorf("histogram lost bounded exact source projection: %s", part)
			}
		}
		rows = d.rows
	default:
		return nil, fmt.Errorf("unexpected histogram source SQL")
	}
	var out bytes.Buffer
	for _, row := range rows {
		encoded, _ := json.Marshal(row)
		out.Write(encoded)
		out.WriteByte('\n')
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(bytes.NewReader(out.Bytes()))}, nil
}
func histogram447Schema(t *testing.T) []json.RawMessage {
	t.Helper()
	schema := append([]json.RawMessage(nil), metricScopeSchema432(t)["histogram"]...)
	for _, name := range []string{"Sum", "Min", "Max"} {
		raw, _ := json.Marshal(map[string]string{"name": name, "type": "Float64"})
		schema = append(schema, raw)
	}
	return schema
}
func histogram447Question(tempo int32) measurements.RawHistogramQuestion {
	return measurements.RawHistogramQuestion{Question: measurements.DiscoveryQuestion{Environment: "prod", Service: "api", FromMS: 1000, ToMS: 2000}, Instrument: measurements.RawHistogramInstrument{Kind: "histogram", Metric: "request.duration", Unit: "ms", Temporality: tempo}}
}
func histogram447Input(q measurements.RawHistogramQuestion, kind int, n uint64) map[string]any {
	count := strconv.FormatUint(n, 10)
	row := map[string]any{"resource_schema_url": "", "resource": map[string]string{"service.name": "api", "deployment.environment.name": "prod", "empty": ""}, "scope": "", "version": "", "schema_url": "", "name": q.Instrument.Metric, "unit": q.Instrument.Unit, "attributes": map[string]string{"cohort": "standard", "empty": ""}, "start_ns": "0", "end_ns": "1500000000", "flags": uint32(0), "temporality": q.Instrument.Temporality, "count": count, "bounds": []string{"1"}, "bucket_counts": []string{count, "0"}, "sum": "0", "min": "0", "max": "0"}
	switch kind {
	case 1:
		row["count"] = "18446744073709551615"
		row["bounds"] = []string{}
		row["bucket_counts"] = []string{"18446744073709551615"}
	case 2:
		row["bounds"] = []string{}
		row["bucket_counts"] = []string{}
	case 3:
		row["bounds"] = []string{"0", "1"}
		row["bucket_counts"] = []string{count}
	case 4:
		row["bounds"] = []string{"NaN"}
	case 5:
		row["bounds"] = []string{"1", "1"}
		row["bucket_counts"] = []string{count, "0", "0"}
	case 6:
		row["count"] = "0"
		row["bucket_counts"] = []string{"18446744073709551615", "1"}
	case 7:
		row["count"] = "01"
		row["bounds"] = []string{}
		row["bucket_counts"] = []string{"1"}
	case 8:
		delete(row, "count")
	case 9:
		row["flags"] = uint32(8)
	case 10:
		row["flags"] = uint32(1)
	case 11:
		row["sum"] = "NaN"
		row["min"] = nil
		row["max"] = "bad"
	}
	return row
}
func histogram447Read(schema, rows []json.RawMessage, q measurements.RawHistogramQuestion) (*measurements.RawHistogramResult, error, *histogram447Transport) {
	driver := &histogram447Transport{schema: schema, rows: rows}
	source := measurements.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: driver}}
	result, err := source.ReadRawHistogramObservations(context.Background(), q)
	return result, err, driver
}
func TestHegelRawHistogram447Conservation(t *testing.T) {
	schema := histogram447Schema(t)
	hegel.Test(t, func(tc *hegel.T) {
		q := histogram447Question(hegel.Draw(tc, hegel.SampledFrom([]int32{-1, 0, 1, 2, 7, 2147483647})))
		n := uint64(9007199254740993) + uint64(hegel.Draw(tc, hegel.Integers(0, 100000)))
		kind := hegel.Draw(tc, hegel.Integers(0, 11))
		input := histogram447Input(q, kind, n)
		raw, _ := json.Marshal(input)
		result, err, driver := histogram447Read(schema, []json.RawMessage{raw, bytes.Clone(raw)}, q)
		if err != nil || result == nil {
			tc.Fatal("histogram source refused inspectable points", err, driver.calls)
			return
		}
		if driver.calls != 3 || result.Version != "goo/raw-histogram-observations/v1" || result.Limit != 10000 || len(result.Rows) != 2 || !reflect.DeepEqual(result.Question, q.Question) || result.Instrument != q.Instrument {
			tc.Fatal("histogram acquisition lost exact bounds/tuple/duplicates")
			return
		}
		for i, row := range result.Rows {
			if row.OccurrenceKey != strconv.Itoa(i+1) || !bytes.Equal(row.RawRow, raw) || row.EndNS != "1500000000" || row.StartNS == nil || *row.StartNS != "0" || row.Temporality != q.Instrument.Temporality {
				tc.Fatal("histogram occurrence changed", row)
				return
			}
			if row.Sum.PresenceState != "unavailable" || row.Min.PresenceState != "unavailable" || row.Max.PresenceState != "unavailable" {
				tc.Fatal("flattened optional source statistic gained producer presence")
				return
			}
			wantBuckets := "valid"
			switch kind {
			case 2, 8, 10:
				wantBuckets = "absent"
			case 3, 4, 5, 6, 7:
				wantBuckets = "invalid"
			}
			if row.BucketsState != wantBuckets {
				tc.Fatal("histogram structural state differs", kind, row.BucketsState, wantBuckets)
				return
			}
			if kind == 0 || kind == 9 || kind == 11 {
				if row.Count == nil || *row.Count != strconv.FormatUint(n, 10) || !reflect.DeepEqual(row.BucketCounts, []string{strconv.FormatUint(n, 10), "0"}) {
					tc.Fatal("histogram count rounded above 2^53", row)
					return
				}
			}
			if kind == 1 && (row.Count == nil || *row.Count != "18446744073709551615" || len(row.Bounds) != 0 || !reflect.DeepEqual(row.BucketCounts, []string{"18446744073709551615"})) {
				tc.Fatal("all-values bucket lost UInt64 maximum")
				return
			}
			if kind == 10 && (row.Count != nil || row.Bounds != nil || row.BucketCounts != nil) {
				tc.Fatal("no-recorded-value exposed normalized population")
				return
			}
			encoded, _ := json.Marshal(row)
			var object map[string]any
			json.Unmarshal(encoded, &object)
			state := "recorded"
			if kind == 9 {
				state = "unknown"
			}
			if kind == 10 {
				state = "no-recorded-value"
			}
			if object["recordingState"] != state {
				tc.Fatal("histogram recordedness changed", kind, object["recordingState"])
				return
			}
			diagnostic := false
			for _, d := range row.Diagnostics {
				if d.Field == "temporality" {
					diagnostic = true
				}
			}
			if diagnostic != (q.Instrument.Temporality != 1 && q.Instrument.Temporality != 2) {
				tc.Fatal("unknown temporality gained implicit support")
				return
			}
		}
		result.Rows[0].Stream.Resource["service.name"] = "changed"
		result.Rows[0].RawRow[0] = 'x'
		if result.Rows[1].Stream.Resource["service.name"] != "api" || !bytes.Equal(result.Rows[1].RawRow, raw) || raw[0] != '{' {
			tc.Fatal("duplicate occurrences alias mutable source facts")
		}
	}, options(100)...)
}
func TestRawHistogram447GoJavaScriptAdmission(t *testing.T) {
	schema := histogram447Schema(t)
	var banks []any
	for _, tempo := range []int32{-1, 0, 1, 2, 7, 2147483647} {
		q := histogram447Question(tempo)
		var rows []json.RawMessage
		for kind := range 12 {
			raw, _ := json.Marshal(histogram447Input(q, kind, 9007199254740993))
			rows = append(rows, raw)
		}
		result, err, _ := histogram447Read(schema, rows, q)
		if err != nil {
			t.Fatal(err)
		}
		banks = append(banks, map[string]any{"request": q, "result": result})
	}
	input, _ := json.Marshal(banks)
	script := `import fs from 'node:fs';import {checkedResult} from '../ui/raw-observations.mjs';const banks=JSON.parse(fs.readFileSync(0,'utf8'));let rows=0;for(const b of banks){const result=checkedResult(b.result,b.request);if(result.rows.length!==b.result.rows.length)throw Error('histogram admission lost points');rows+=result.rows.length}const b=banks[2],mutants=[r=>r.rows[0].count='9',r=>r.rows[0].bucketCounts[0]='9',r=>r.rows[0].bounds=[2],r=>r.rows[0].sum.presenceState='present',r=>r.rows[9].recordingState='recorded',r=>r.rows[0].temporality=2,r=>r.rows[0].rawRow.temporality=2,r=>r.rows[0].occurrenceKey=r.rows[1].occurrenceKey,r=>r.version='goo/raw-sum-observations/v1',r=>r.rows[3].bucketsState='valid'];for(const change of mutants){const result=structuredClone(b.result);change(result);let refused=false;try{checkedResult(result,b.request)}catch{refused=true}if(!refused)throw Error('histogram semantic mutant admitted')}process.stdout.write(JSON.stringify({banks:banks.length,rows,controls:mutants.length}));`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("histogram Go/JS admission: %v %s", err, out)
	}
	t.Log(string(out))
	q := histogram447Question(1)
	good, _ := json.Marshal(histogram447Input(q, 0, 9))
	for _, mutate := range []func(map[string]any){func(r map[string]any) { delete(r, "temporality") }, func(r map[string]any) { r["temporality"] = 2 }, func(r map[string]any) { r["name"] = "other" }, func(r map[string]any) { r["end_ns"] = "2000000000" }} {
		bad := histogram447Input(q, 0, 9)
		mutate(bad)
		raw, _ := json.Marshal(bad)
		result, err, _ := histogram447Read(schema, []json.RawMessage{good, raw}, q)
		if err == nil || result != nil {
			t.Fatal("identity/window mismatch published partial histogram bank")
		}
	}
	q.Instrument.Kind = "exponential-histogram"
	result, err, driver := histogram447Read(schema, []json.RawMessage{good}, q)
	if err == nil || result != nil || driver.calls != 0 {
		t.Fatal("explicit histogram accepted another physical family")
	}
}
