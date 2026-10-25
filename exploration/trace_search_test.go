package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func traceSearchTestRequest() TraceSearchRequest {
	return TraceSearchRequest{Version: TraceSearchVersion, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, Conditions: []Predicate{{All: []Predicate{{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "db"}, {Field: &Field{Kind: "column", Key: "duration_ns"}, Op: "gte", Value: "50000000"}}}, {Field: &Field{Kind: "column", Key: "status_class"}, Op: "eq", Value: "error"}}}
}
func traceSearchTestRecords() []TraceSearchRecord {
	row := func(id, service, status, work string, at uint64) TraceSearchRecord {
		return TraceSearchRecord{Environment: "prod", Fields: map[Field]string{{Kind: "column", Key: "trace_id"}: id, {Kind: "column", Key: "span_id"}: "0000000000000001", {Kind: "column", Key: "timestamp_ns"}: fmt.Sprint(at), {Kind: "column", Key: "duration_ns"}: "60000000", {Kind: "column", Key: "service"}: service, {Kind: "column", Key: "status_class"}: status, {Kind: "column", Key: "status"}: map[string]string{"error": "Error", "ok": "Ok"}[status], {Kind: "column", Key: "kind"}: "Server", {Kind: "column", Key: "name"}: "operation", {Kind: "attribute", Key: "work"}: work}}
	}
	a, b, c, d := strings.Repeat("a", 32), strings.Repeat("b", 32), strings.Repeat("c", 32), strings.Repeat("d", 32)
	rows := []TraceSearchRecord{row(a, "db", "ok", "9007199254740993", 1000000000), row(a, "worker", "error", "0", 1100000000), row(b, "db", "error", "9007199254740994", 1200000000), row(c, "db", "ok", "1", 1300000000), row(d, "worker", "error", "1", 1400000000), row(strings.Repeat("0", 32), "db", "error", "9", 1500000000), row(strings.Repeat("A", 32), "db", "error", "9", 1500000000), row(strings.Repeat("a", 32)+"\n", "db", "error", "9", 1500000000), row(strings.Repeat("e", 32), "db", "error", "9", 999999999), row(strings.Repeat("f", 32), "db", "error", "9", 61000000000)}
	wrong := row(strings.Repeat("1", 32), "db", "error", "9", 1500000000)
	wrong.Environment = "other"
	tied := TraceSearchRecord{Environment: rows[0].Environment, Fields: map[Field]string{}}
	for k, v := range rows[0].Fields {
		tied.Fields[k] = v
	}
	tied.Fields[Field{Kind: "attribute", Key: "work"}] = "9007199254740995"
	rows = append(rows, wrong, rows[0], tied)
	return rows
}
func TestTraceSearchReferenceFullTraceExistsAndLimits(t *testing.T) {
	r := traceSearchTestRequest()
	items, count, e := ReferenceTraceSearch(r, traceSearchTestRecords())
	if e != nil || count != "2" || len(items) != 1 || items[0].Witnesses[0].MatchingRecords != "3" {
		t.Fatal(items, count, e)
	}
	r.Limit = 100
	items, count, e = ReferenceTraceSearch(r, traceSearchTestRecords())
	if e != nil || len(items) != 2 || items[1].Witnesses[0].SpanID != items[1].Witnesses[1].SpanID {
		t.Fatal("same record may satisfy both conditions", items, count, e)
	}
	plan, e := CompileTraceSearch(r)
	if e != nil || !strings.Contains(plan.SQL, "count() OVER ()") || !strings.Contains(plan.SQL, "HAVING countIf(") || !strings.Contains(plan.SQL, "length(TraceId) = 32") {
		t.Fatal(plan, e)
	}
	if strings.Contains(plan.SQL, "'db'") {
		t.Fatal("interpolated user predicate")
	}
	for _, n := range []int{0, 5} {
		bad := r
		bad.Conditions = make([]Predicate, n)
		if bad.Validate() == nil {
			t.Fatal("condition limit")
		}
	}
	bad := r
	bad.Conditions = []Predicate{{Field: &Field{Kind: "column", Key: "body"}, Op: "contains", Value: "secret"}}
	if bad.Validate() == nil {
		t.Fatal("log condition accepted")
	}
	if _, _, e := ReferenceTraceSearch(bad, nil); e == nil {
		t.Fatal("invalid reference accepted")
	}
}

type traceSearchTransport func(*http.Request) (*http.Response, error)

func (f traceSearchTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestTraceSearchRejectsMalformedSourceProof(t *testing.T) {
	r := traceSearchTestRequest()
	for _, body := range []string{`{"trace_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`, `{"trace_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","matched_traces":"0"}`, `{"trace_id":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","matched_traces":"1"}`} {
		s := Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
		})}}
		if _, e := s.SearchTraces(context.Background(), r); e == nil {
			t.Fatal("malformed proof accepted", body)
		}
	}
}
func TestClickHouseTraceSearchParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL not configured")
	}
	records := traceSearchTestRecords()
	payload := []map[string]any{}
	for _, r := range records {
		v := r.Fields
		column := func(k string) string { return v[Field{Kind: "column", Key: k}] }
		payload = append(payload, map[string]any{"trace": column("trace_id"), "span": column("span_id"), "parent": column("parent_span_id"), "service": column("service"), "operation": column("name"), "kind": column("kind"), "timestamp": column("timestamp_ns"), "duration": column("duration_ns"), "status": column("status"), "resources": map[string]string{"deployment.environment.name": r.Environment}, "attributes": map[string]string{"work": v[Field{Kind: "attribute", Key: "work"}]}})
	}
	data, _ := json.Marshal(payload)
	inline := " FROM (SELECT JSONExtractString(j,'trace') AS TraceId, JSONExtractString(j,'span') AS SpanId, JSONExtractString(j,'parent') AS ParentSpanId, JSONExtractString(j,'service') AS ServiceName, JSONExtractString(j,'operation') AS SpanName, JSONExtractString(j,'kind') AS SpanKind, JSONExtractString(j,'status') AS StatusCode, toUInt64(JSONExtractString(j,'duration')) AS Duration, fromUnixTimestamp64Nano(toInt64(JSONExtractString(j,'timestamp'))) AS Timestamp, JSONExtract(j,'resources','Map(String,String)') AS ResourceAttributes, JSONExtract(j,'attributes','Map(String,String)') AS SpanAttributes FROM (SELECT JSONExtractArrayRaw({fixture:String})[number+1] AS j FROM numbers(" + fmt.Sprint(len(payload)) + ")))"
	transport := traceSearchTransport(func(req *http.Request) (*http.Response, error) {
		raw, e := io.ReadAll(req.Body)
		if e != nil {
			return nil, e
		}
		req = req.Clone(req.Context())
		sql := strings.Replace(string(raw), " FROM otel_traces", inline, 1)
		req.Body = io.NopCloser(strings.NewReader(sql))
		req.ContentLength = int64(len(sql))
		u := *req.URL
		q := u.Query()
		q.Set("param_fixture", encodeParameter(string(data)))
		u.RawQuery = q.Encode()
		req.URL = &u
		return http.DefaultTransport.RoundTrip(req)
	})
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"), Client: &http.Client{Transport: transport}}
	requests := []TraceSearchRequest{traceSearchTestRequest()}
	numeric := traceSearchTestRequest()
	numeric.Limit = 100
	numeric.Conditions[0] = Predicate{Numeric: &NumericBinding{Field: Field{Kind: "attribute", Key: "work"}, Type: "uint64", Unit: "items", Parse: "canonical-decimal"}, Op: "gte", Value: "9007199254740993"}
	requests = append(requests, numeric)
	empty := traceSearchTestRequest()
	empty.Conditions = []Predicate{{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "absent"}}
	requests = append(requests, empty)
	for _, r := range requests {
		want, count, e := ReferenceTraceSearch(r, records)
		if e != nil {
			t.Fatal(e)
		}
		got, e := source.SearchTraces(context.Background(), r)
		if e != nil {
			t.Fatal(e)
		}
		if got.MatchedTraces != count || !reflect.DeepEqual(got.Items, want) || got.HasMore != (len(want) == r.Limit && count != fmt.Sprint(len(want))) {
			t.Fatalf("Go/ClickHouse trace search differs\ngot %#v\nwant %#v total%s", got, want, count)
		}
	}
}

func TestTraceSearchRejectsContradictoryWitnessMetadata(t *testing.T) {
	request := traceSearchTestRequest()
	items, _, e := ReferenceTraceSearch(request, traceSearchTestRecords())
	if e != nil {
		t.Fatal(e)
	}
	wire := func() map[string]any {
		out := map[string]any{"trace_id": items[0].TraceID, "matched_traces": "1"}
		for _, w := range items[0].Witnesses {
			values := []any{}
			for _, f := range w.FieldValues {
				present := 0
				value := ""
				if f.Value != nil {
					present = 1
					value = *f.Value
				}
				values = append(values, []any{present, value})
			}
			out[fmt.Sprint("c", w.ConditionIndex)] = w.MatchingRecords
			out[fmt.Sprint("w", w.ConditionIndex)] = []any{w.TimestampNS, w.SpanID, w.ParentSpanID, w.Service, w.Operation, w.Kind, w.DurationNS, values}
		}
		return out
	}
	tests := []struct {
		name string
		edit func(map[string]any)
	}{{"positive", func(map[string]any) {}}, {"wrong-service", func(o map[string]any) { o["w0"].([]any)[3] = "wrong" }}, {"outside-window", func(o map[string]any) { o["w0"].([]any)[0] = "999999999" }}, {"end-exclusive", func(o map[string]any) { o["w0"].([]any)[0] = "61000000000" }}, {"null-service", func(o map[string]any) { o["w0"].([]any)[3] = nil }}, {"null-proof", func(o map[string]any) { o["w0"].([]any)[7].([]any)[0].([]any)[1] = nil }}, {"false-proof", func(o map[string]any) { o["w0"].([]any)[7].([]any)[0].([]any)[1] = "1" }}, {"boolean-presence", func(o map[string]any) { o["w0"].([]any)[7].([]any)[0].([]any)[0] = true }}, {"missing-proof", func(o map[string]any) { o["w0"].([]any)[7] = []any{} }}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := wire()
			tc.edit(object)
			raw, _ := json.Marshal(object)
			s := Source{Endpoint: "http://unused", Client: &http.Client{Transport: traceSearchTransport(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(raw))), Header: http.Header{}}, nil
			})}}
			_, e := s.SearchTraces(context.Background(), request)
			if tc.name == "positive" && e != nil {
				t.Fatal(e)
			}
			if tc.name != "positive" && e == nil {
				t.Fatal("malformed witness accepted")
			}
		})
	}
}
func TestTraceSearchCompiledPlanBound(t *testing.T) {
	r := traceSearchTestRequest()
	binding := NumericBinding{Field: Field{Kind: "attribute", Key: "amount"}, Type: "decimal64", Scale: 9, Unit: "USD", Parse: "canonical-decimal"}
	group := Predicate{All: []Predicate{}}
	for i := 0; i < 31; i++ {
		b := binding
		group.All = append(group.All, Predicate{Numeric: &b, Op: "valid"})
	}
	r.Conditions = []Predicate{group, group}
	if e := r.Validate(); e != nil {
		t.Fatal(e)
	}
	p, e := CompileTraceSearch(r)
	if e == nil && len(p.SQL) > 256<<10 {
		t.Fatal("oversize SQL admitted")
	}
	if e != nil && !strings.Contains(e.Error(), "compiled SQL exceeds") {
		t.Fatal(e)
	}
}

func TestClickHouseTraceSearchAbsentDerivedEvidenceTieParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL not configured")
	}
	request := traceSearchTestRequest()
	request.Limit = 20
	request.Conditions = []Predicate{{All: []Predicate{{Field: &Field{Kind: "column", Key: "exception_type"}, Op: "missing"}, {Field: &Field{Kind: "resource", Key: "zone"}, Op: "exists"}}}}
	records := []TraceSearchRecord{}
	for _, zone := range []string{"a", "z"} {
		records = append(records, TraceSearchRecord{Environment: "prod", Fields: map[Field]string{{Kind: "column", Key: "trace_id"}: strings.Repeat("a", 32), {Kind: "column", Key: "span_id"}: "0000000000000001", {Kind: "column", Key: "timestamp_ns"}: "1000000000", {Kind: "column", Key: "duration_ns"}: "100", {Kind: "column", Key: "service"}: "db", {Kind: "column", Key: "name"}: "operation", {Kind: "column", Key: "kind"}: "Server", {Kind: "resource", Key: "zone"}: zone}})
	}
	// Both exception types are unreported as a unique type. Their hidden first
	// values deliberately sort opposite the visible resource-zone evidence.
	inline := " FROM (SELECT ['z','a'][number+1] AS hidden, ['a','z'][number+1] AS zone, 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' AS TraceId, '0000000000000001' AS SpanId, '' AS ParentSpanId, 'db' AS ServiceName, 'operation' AS SpanName, 'Server' AS SpanKind, toUInt64(100) AS Duration, fromUnixTimestamp64Nano(toInt64(1000000000)) AS Timestamp, map('deployment.environment.name','prod','zone',zone) AS ResourceAttributes, map() AS SpanAttributes, ['exception','exception'] AS `Events.Name`, [map('exception.type',hidden),map('exception.type',concat(hidden,'2'))] AS `Events.Attributes` FROM numbers(2))"
	transport := traceSearchTransport(func(req *http.Request) (*http.Response, error) {
		raw, e := io.ReadAll(req.Body)
		if e != nil {
			return nil, e
		}
		req = req.Clone(req.Context())
		sql := strings.Replace(string(raw), " FROM otel_traces", inline, 1)
		req.Body = io.NopCloser(strings.NewReader(sql))
		req.ContentLength = int64(len(sql))
		return http.DefaultTransport.RoundTrip(req)
	})
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"), Client: &http.Client{Transport: transport}}
	want, count, e := ReferenceTraceSearch(request, records)
	if e != nil {
		t.Fatal(e)
	}
	got, e := source.SearchTraces(context.Background(), request)
	if e != nil {
		t.Fatal(e)
	}
	if count != got.MatchedTraces || !reflect.DeepEqual(want, got.Items) {
		w, _ := json.Marshal(want)
		g, _ := json.Marshal(got.Items)
		t.Fatalf("absent derived projection altered visible evidence tie: got %s want %s", g, w)
	}
}

func TestTraceSearchAbsentProofRequiresCanonicalEmptyPayload(t *testing.T) {
	p := Predicate{Field: &Field{Kind: "attribute", Key: "optional"}, Op: "missing"}
	if _, e := decodeTraceSearchValues(json.RawMessage(`[[0,"hidden"]]`), p); e == nil {
		t.Fatal("absent evidence retained hidden tie value")
	}
	out, e := decodeTraceSearchValues(json.RawMessage(`[[0,""]]`), p)
	if e != nil || len(out) != 1 || out[0].Value != nil {
		t.Fatal("canonical absent evidence refused", out, e)
	}
}
