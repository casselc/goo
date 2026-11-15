package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func link453Query() Query {
	return Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, Offset: 1, LinksTo: &SpanLinkTarget{TraceID: strings.Repeat("a", 32), SpanID: strings.Repeat("b", 16)}}
}

func TestSpanLink453ReferenceTupleMembership(t *testing.T) {
	target := link453Query().LinksTo
	otherTrace, otherSpan := strings.Repeat("c", 32), strings.Repeat("d", 16)
	for _, tc := range []struct {
		name          string
		traces, spans []string
		attrs         []map[string]string
		want          bool
	}{
		{"empty", nil, nil, nil, false},
		{"exact", []string{target.TraceID}, []string{target.SpanID}, []map[string]string{{}}, true},
		{"crossed", []string{target.TraceID, otherTrace}, []string{otherSpan, target.SpanID}, []map[string]string{{}, {}}, false},
		{"repeated links", []string{target.TraceID, target.TraceID}, []string{target.SpanID, target.SpanID}, []map[string]string{{}, {}}, true},
		{"unaligned spans", []string{target.TraceID}, nil, []map[string]string{{}}, false},
		{"unaligned attributes", []string{target.TraceID}, []string{target.SpanID}, nil, false},
		{"matching attributes absent", []string{target.TraceID}, []string{target.SpanID}, []map[string]string{nil}, false},
		{"unrelated invalid ID", []string{"invalid", target.TraceID}, []string{"", target.SpanID}, []map[string]string{nil, {}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := MatchSpanLinkTarget(target, tc.traces, tc.spans, tc.attrs); got != tc.want {
				t.Fatal(got, tc.want)
			}
		})
	}
	if MatchSpanLinkTarget(nil, []string{target.TraceID}, []string{target.SpanID}, []map[string]string{{}}) {
		t.Fatal("nil target matched")
	}
}

func TestSpanLink453PlanBindsFullPopulationBeforePaging(t *testing.T) {
	q := link453Query()
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.BucketMS = 1000
	q.GroupTimeSeries = true
	plan, err := Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{plan.Rows, plan.Count, plan.Groups, plan.Fields, plan.TimeSeries, plan.GroupedTimeSeries} {
		if sql == "" || !strings.Contains(sql, "arrayExists(i ->") || !strings.Contains(sql, "length(`Links.TraceId`)=length(`Links.Attributes`)") || strings.Contains(sql, q.LinksTo.TraceID) || strings.Contains(sql, q.LinksTo.SpanID) || strings.Contains(sql, "arrayJoin(`Links.") {
			t.Fatalf("selector lost or unbound: %s", sql)
		}
	}
	if !strings.Contains(plan.Rows, "LIMIT 2 OFFSET 1") || strings.Contains(plan.Count, "LIMIT") || strings.Contains(plan.Count, "OFFSET") {
		t.Fatal("page defines matching count", plan)
	}
	values := []string{}
	for _, v := range plan.Parameters {
		values = append(values, v)
	}
	if !slices.Contains(values, q.LinksTo.TraceID) || !slices.Contains(values, q.LinksTo.SpanID) {
		t.Fatal("target IDs not bound", plan.Parameters)
	}
	if !strings.Contains(plan.Count, "arrayEnumerate(`Links.TraceId`)") {
		t.Fatal("unsafe multi-array lambda")
	}
	q.Signal = "logs"
	if q.Validate() == nil {
		t.Fatal("links admitted for logs")
	}
}

func TestSpanLink453WireAndCustody(t *testing.T) {
	q := link453Query()
	raw, err := json.Marshal(q)
	if err != nil {
		t.Fatal(err)
	}
	var restored Query
	if err = json.Unmarshal(raw, &restored); err != nil || !reflect.DeepEqual(q, restored) {
		t.Fatal("roundtrip lost selector", err)
	}
	for _, target := range []string{`null`, `{}`, `{"traceId":"` + q.LinksTo.TraceID + `"}`, `{"TraceId":"` + q.LinksTo.TraceID + `","spanId":"` + q.LinksTo.SpanID + `"}`, `{"traceId":"` + q.LinksTo.TraceID + `","spanId":null}`, `{"traceId":"` + q.LinksTo.TraceID + `","spanId":"` + q.LinksTo.SpanID + `","extra":true}`, `{"traceId":"` + strings.Repeat("0", 32) + `","spanId":"` + q.LinksTo.SpanID + `"}`, `{"traceId":"` + strings.Repeat("A", 32) + `","spanId":"` + q.LinksTo.SpanID + `"}`} {
		wire := `{"signal":"spans","environment":"prod","fromMs":1000,"toMs":61000,"limit":1,"linksTo":` + target + `}`
		var v Query
		if json.Unmarshal([]byte(wire), &v) == nil && v.Validate() == nil {
			t.Fatal("unsafe target admitted", target)
		}
	}
	clone, err := CloneQuery(q)
	if err != nil {
		t.Fatal(err)
	}
	capture := (CaptureQuestion{Kind: "explore", Explore: &q}).Normalized()
	scope, err := (CaptureQuestion{Kind: "explore", Explore: &q}).Scope()
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := DraftTelemetryPackage(q, "linked", "Spans reporting a link", 1)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ReadTelemetryPackage(plan.YAML)
	if err != nil || !reflect.DeepEqual(decoded.LinksTo, q.LinksTo) {
		t.Fatal("package lost target", err)
	}
	detached := pkg.Question(q.FromMS, q.ToMS)
	detached.LinksTo.SpanID = strings.Repeat("e", 16)
	if pkg.LinksTo.SpanID != q.LinksTo.SpanID {
		t.Fatal("package question aliased target")
	}
	q.LinksTo.TraceID = strings.Repeat("f", 32)
	for _, v := range []*SpanLinkTarget{clone.LinksTo, capture.Explore.LinksTo, scope.LinksTo, pkg.LinksTo} {
		if v.TraceID == q.LinksTo.TraceID {
			t.Fatal("selector aliased caller")
		}
	}
	wider := clone
	wider.LinksTo = nil
	if _, err = DescribeTransform(CaptureQuestion{Kind: "explore", Explore: &clone}, CaptureQuestion{Kind: "explore", Explore: &wider}, "same-population"); err == nil {
		t.Fatal("membership removal claimed same population")
	}
}

func TestSpanLink453SourceMembershipAdmission(t *testing.T) {
	q := link453Query()
	record := map[string]any{"SpanName": "consume", "timestamp_ns": "2000000000", "ResourceAttributes": map[string]string{"deployment.environment.name": "prod"}, "linked_traces": []string{q.LinksTo.TraceID}, "linked_spans": []string{q.LinksTo.SpanID}, "linked_attributes": []map[string]string{{}}}
	raw, _ := json.Marshal(record)
	if err := validateSelectedRows(q, []json.RawMessage{raw}); err != nil {
		t.Fatal(err)
	}
	unrelated := map[string]any{}
	json.Unmarshal(raw, &unrelated)
	unrelated["linked_traces"] = []any{42, q.LinksTo.TraceID}
	unrelated["linked_spans"] = []any{nil, q.LinksTo.SpanID}
	unrelated["linked_attributes"] = []any{map[string]any{"bad": 42}, map[string]string{}}
	kept, _ := json.Marshal(unrelated)
	if err := validateSelectedRows(q, []json.RawMessage{kept}); err != nil {
		t.Fatal("unrelated malformed link hid valid selected entry", err)
	}
	for _, mutate := range []func(map[string]any){func(r map[string]any) { r["linked_spans"] = []string{strings.Repeat("c", 16)} }, func(r map[string]any) { r["linked_attributes"] = nil }, func(r map[string]any) { r["linked_attributes"] = []map[string]any{{"bad": nil}} }, func(r map[string]any) { r["timestamp_ns"] = "61000000000" }} {
		var changed map[string]any
		json.Unmarshal(raw, &changed)
		mutate(changed)
		bad, _ := json.Marshal(changed)
		if validateSelectedRows(q, []json.RawMessage{bad}) == nil {
			t.Fatal("source membership drift admitted", string(bad))
		}
	}
}

func TestClickHouseSpanLink453TupleParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	target := link453Query().LinksTo
	otherTrace, otherSpan := strings.Repeat("c", 32), strings.Repeat("d", 16)
	traces := [][]string{{target.TraceID}, {target.TraceID, otherTrace}, {target.TraceID, target.TraceID}, {target.TraceID}, {target.TraceID}, {}, {"invalid", target.TraceID}, {target.TraceID}}
	spans := [][]string{{target.SpanID}, {otherSpan, target.SpanID}, {target.SpanID, target.SpanID}, {}, {target.SpanID}, {}, {"", target.SpanID}, {target.SpanID}}
	attrs := [][]map[string]string{{{}}, {{}, {}}, {{}, {}}, {{}}, {}, {}, {{}, {}}, {{}}}
	params := map[string]string{}
	for key, value := range map[string]any{"traces": traces, "spans": spans, "attrs": attrs} {
		wire, _ := json.Marshal(value)
		params[key] = string(wire)
	}
	c := compiler{signal: "spans", params: params}
	membership := c.spanLinksTo(target)
	inline := " FROM (SELECT number AS n, JSONExtract({traces:String},'Array(Array(String))')[number+1] AS `Links.TraceId`, JSONExtract({spans:String},'Array(Array(String))')[number+1] AS `Links.SpanId`, JSONExtract({attrs:String},'Array(Array(Map(String,String)))')[number+1] AS `Links.Attributes` FROM numbers(" + fmt.Sprint(len(traces)) + "))"
	parsed, err := url.Parse(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	settings := parsed.Query()
	settings.Set("short_circuit_function_evaluation", "disable")
	parsed.RawQuery = settings.Encode()
	source := Source{Endpoint: parsed.String(), User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	rows, err := source.read(context.Background(), "SELECT toString(n) AS n,toUInt8("+membership+") AS matched"+inline+" ORDER BY n FORMAT JSONEachRow", c.params)
	if err != nil || len(rows) != len(traces) {
		t.Fatal(err, len(rows))
	}
	expected := []string{}
	for i, row := range rows {
		var got struct {
			N       string
			Matched uint8
		}
		if err = json.Unmarshal(row, &got); err != nil {
			t.Fatal(err)
		}
		want := MatchSpanLinkTarget(target, traces[i], spans[i], attrs[i])
		if (got.Matched == 1) != want {
			t.Fatalf("tuple parity differs for %d: %s", i, row)
		}
		if want {
			expected = append(expected, fmt.Sprint(i))
		}
	}
	count, err := source.read(context.Background(), "SELECT toString(count()) AS count"+inline+" WHERE "+membership+" FORMAT JSONEachRow", c.params)
	if err != nil || len(count) != 1 {
		t.Fatal(err)
	}
	var total struct{ Count string }
	json.Unmarshal(count[0], &total)
	if total.Count != fmt.Sprint(len(expected)) {
		t.Fatal("link entries counted instead of records", total, expected)
	}
	page, err := source.read(context.Background(), "SELECT toString(n) AS n"+inline+" WHERE "+membership+" ORDER BY n LIMIT 1 OFFSET 1 FORMAT JSONEachRow", c.params)
	if err != nil || len(page) != 1 {
		t.Fatal(err)
	}
	var second struct{ N string }
	json.Unmarshal(page[0], &second)
	if second.N != expected[1] {
		t.Fatal("paging changed grain", second, expected)
	}
}

func TestSpanLink453UnsupportedSchemaRefusesBeforePopulationReads(t *testing.T) {
	data, err := os.ReadFile("../internal/sourcecap/testdata/schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var families map[string][]map[string]string
	if err = json.Unmarshal(data, &families); err != nil {
		t.Fatal(err)
	}
	for _, column := range families["spans"] {
		if column["name"] == "Links.Attributes" {
			column["type"] = "Array(Map(String, Nullable(String)))"
		}
	}
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		calls = append(calls, sql)
		if sql != "DESCRIBE TABLE otel_traces FORMAT JSONEachRow" {
			t.Error("population source read started despite incompatible link shape", sql)
			http.Error(w, "unexpected read", 500)
			return
		}
		for _, column := range families["spans"] {
			json.NewEncoder(w).Encode(column)
		}
	}))
	defer server.Close()
	q := link453Query()
	result, err := (Source{Endpoint: server.URL}).Fetch(context.Background(), q)
	if err == nil || result != nil || len(calls) != 1 {
		t.Fatal("unsupported nullable link map admitted", result, err, calls)
	}
}
