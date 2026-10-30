package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func histogramQuestion(typ string) Query {
	q := basic()
	b := numericBinding()
	b.Type = typ
	q.Numeric = &b
	q.NumericHistogram = &NumericHistogramSpec{Binning: "powers-of-two"}
	return q
}
func histogramValues() []*string {
	out := []*string{nil}
	for _, s := range []string{"", "0", "1", "-1", "-2", "2", "3", "4", "9007199254740993", "-9223372036854775808", "9223372036854775807", "18446744073709551615", "18446744073709551616", "-9223372036854775809", "01", "-0", "+1", "1e3", " 1", "1\n"} {
		v := s
		out = append(out, &v)
	}
	return out
}

func TestNumericHistogramFixedEdgesAndCoverage(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		q := histogramQuestion(typ)
		edges, e := NumericHistogramEdges(typ)
		if e != nil {
			t.Fatal(e)
		}
		want := 66
		first := "0"
		last := "18446744073709551616"
		if typ == "int64" {
			want = 129
			first = "-9223372036854775808"
			last = "9223372036854775808"
		}
		if len(edges) != want || edges[0] != first || edges[len(edges)-1] != last {
			t.Fatal("wrong full-domain geometry", typ, edges)
		}
		for i := 1; i < len(edges); i++ {
			a, _ := new(big.Int).SetString(edges[i-1], 10)
			b, _ := new(big.Int).SetString(edges[i], 10)
			if a.Cmp(b) >= 0 {
				t.Fatal("unordered bins")
			}
		}
		values := histogramValues()
		h, e := EvaluateNumericHistogram(*q.Numeric, values)
		summary, _ := EvaluateNumericValues(*q.Numeric, values)
		if e != nil || h.Validate(*q.Numeric, summary) != nil {
			t.Fatal(h, e)
		}
		for _, raw := range values {
			if raw == nil {
				continue
			}
			v, valid := parseCanonicalInteger(*raw, typ)
			matches := 0
			for _, bin := range h.Bins {
				from, _ := new(big.Int).SetString(bin.From, 10)
				to, _ := new(big.Int).SetString(bin.To, 10)
				if valid && v.Cmp(from) >= 0 && v.Cmp(to) < 0 {
					matches++
				}
			}
			if valid && matches != 1 || !valid && matches != 0 {
				t.Fatal("value not in exactly one valid bin", typ, *raw, matches)
			}
		}
		for _, input := range [][]*string{nil, numericValues("", "bad", "01")} {
			empty, e := EvaluateNumericHistogram(*q.Numeric, input)
			s, _ := EvaluateNumericValues(*q.Numeric, input)
			if e != nil || empty.Validate(*q.Numeric, s) != nil {
				t.Fatal("zero-valid coverage rejected", e)
			}
			for _, bin := range empty.Bins {
				if bin.Count != "0" {
					t.Fatal("invalid values entered histogram")
				}
			}
		}
		// Every canonical in-range edge belongs to its right-hand half-open bin.
		for i, edge := range edges[:len(edges)-1] {
			if _, ok := parseCanonicalInteger(edge, typ); !ok {
				t.Fatal("non-sentinel edge invalid")
			}
			one, e := EvaluateNumericHistogram(*q.Numeric, numericValues(edge))
			if e != nil || one.Bins[i].Count != "1" {
				t.Fatal("edge assigned to wrong bin", typ, i, e)
			}
		}
	}
	if _, e := NumericHistogramEdges("float64"); e == nil {
		t.Fatal("unsupported type accepted")
	}
}

func TestNumericHistogramValidationAndLegacyOmission(t *testing.T) {
	q := histogramQuestion("int64")
	values := histogramValues()
	h, _ := EvaluateNumericHistogram(*q.Numeric, values)
	summary, _ := EvaluateNumericValues(*q.Numeric, values)
	for _, mutate := range []func(*NumericHistogram){func(h *NumericHistogram) { h.Binning = "auto" }, func(h *NumericHistogram) { h.Bins = h.Bins[:1] }, func(h *NumericHistogram) { h.Bins[0].From = "-1" }, func(h *NumericHistogram) { h.Bins[0].To = "0" }, func(h *NumericHistogram) { h.Bins[0].Count = "01" }, func(h *NumericHistogram) { h.Bins[0].Count = "18446744073709551616" }, func(h *NumericHistogram) { h.Bins[0].Count = "-1" }, func(h *NumericHistogram) { h.Bins[0].Count = "0" }, func(h *NumericHistogram) { h.Bins[0], h.Bins[1] = h.Bins[1], h.Bins[0] }} {
		bad := detached(h)
		mutate(&bad)
		if bad.Validate(*q.Numeric, summary) == nil {
			t.Fatal("corrupt histogram accepted")
		}
	}
	for _, mutate := range []func(*Query){func(q *Query) { q.Numeric = nil }, func(q *Query) { q.NumericHistogram.Binning = "linear" }} {
		bad := detached(q)
		mutate(&bad)
		if _, e := Compile(bad); e == nil {
			t.Fatal("invalid histogram request lowered")
		}
	}
	r := Result{Question: q, Count: summary.Count, Numeric: &summary, NumericHistogram: &h}
	if ValidateNumericResult(&r) != nil {
		t.Fatal("valid histogram result rejected")
	}
	r.NumericHistogram = nil
	if ValidateNumericResult(&r) == nil {
		t.Fatal("missing histogram accepted")
	}
	r.NumericHistogram = &h
	r.Question.NumericHistogram = nil
	if ValidateNumericResult(&r) == nil {
		t.Fatal("unrequested histogram accepted")
	}
	old := basic()
	wire, _ := json.Marshal(old)
	if strings.Contains(string(wire), "numericHistogram") {
		t.Fatal("legacy query bytes grew")
	}
	wire, _ = json.Marshal(Result{Question: old})
	if strings.Contains(string(wire), "numericHistogram") {
		t.Fatal("legacy result bytes grew")
	}
}

func TestNumericHistogramCompilePackageAndScope(t *testing.T) {
	q := histogramQuestion("int64")
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.BucketMS = 10000
	q.GroupTimeSeries = true
	legacy := detached(q)
	legacy.NumericHistogram = nil
	before, e := Compile(legacy)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(plan.Count, "AS numeric_histogram") || !strings.Contains(plan.Count, "toInt256('9223372036854775808')") || strings.Contains(plan.Count, "Float") {
		t.Fatal("missing exact histogram lowering")
	}
	for _, sql := range []string{plan.Rows, plan.Fields, plan.Groups, plan.TimeSeries, plan.GroupedTimeSeries} {
		if strings.Contains(sql, "numeric_histogram") {
			t.Fatal("histogram added outside main aggregate")
		}
	}
	oldAgain, e := Compile(legacy)
	if e != nil || !reflect.DeepEqual(before, oldAgain) {
		t.Fatal("nonopt-in SQL changed", e)
	}
	pkg, e := DraftTelemetryPackage(q, "histogram", "Histogram", 1)
	if e != nil {
		t.Fatal(e)
	}
	pp, e := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	if e != nil {
		t.Fatal(e)
	}
	parsed, e := ReadTelemetryPackage(pp.YAML)
	if e != nil || !reflect.DeepEqual(parsed.NumericHistogram, q.NumericHistogram) || !reflect.DeepEqual(parsed.Question(q.FromMS, q.ToMS).NumericHistogram, q.NumericHistogram) {
		t.Fatal("package lost histogram", e)
	}
	if pp.MeasurePlan.NumericHistogram == nil || pp.MeasurePlan.Measures[len(pp.MeasurePlan.Measures)-1].ID != "numeric_histogram" {
		t.Fatal("measure plan omitted histogram")
	}
	population, e := PlanPopulation(PopulationRequest{Question: q})
	if e != nil || population.Queries[0].NumericHistogram == nil || population.Queries[1].NumericHistogram != nil || population.Queries[1].Numeric != nil {
		t.Fatal("companion retained orphan histogram", e)
	}
	transform, e := DescribeTransform(CaptureQuestion{Kind: "explore", Explore: &legacy}, CaptureQuestion{Kind: "explore", Explore: &q}, "same-population")
	if e != nil || transform.MeasurePlan.NumericHistogram == nil {
		t.Fatal("histogram presentation changed population scope", e)
	}
	q.NumericHistogram.Binning = "mutated"
	if parsed.NumericHistogram.Binning != "powers-of-two" || pp.MeasurePlan.NumericHistogram.Binning != "powers-of-two" {
		t.Fatal("planner aliases caller")
	}
}

func TestNumericHistogramSelectionPartitionAndSentinels(t *testing.T) {
	for _, typ := range []string{"int64", "uint64"} {
		q := histogramQuestion(typ)
		q.Where = &Predicate{Field: &Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "orders"}
		q.GroupBy = []Field{{Kind: "column", Key: "service"}}
		q.Summary = "span-duration"
		q.SpanStatus = true
		q.Offset = 50
		edges, _ := NumericHistogramEdges(typ)
		n := len(edges) - 1
		for _, span := range [][2]int{{0, n}, {0, 1}, {n - 1, n}, {n / 2, n/2 + 1}, {1, n - 1}} {
			plan, e := PlanNumericHistogramSelection(q, span[0], span[1])
			if e != nil {
				t.Fatal(e)
			}
			if plan.From != edges[span[0]] || plan.To != edges[span[1]] || !reflect.DeepEqual(plan.Parent, q) || plan.Selected.Offset != 0 || plan.Rest.Offset != 0 {
				t.Fatal("plan changed geometry/parent")
			}
			for _, v := range []Query{plan.Selected, plan.Rest} {
				if v.Numeric == nil || v.NumericHistogram == nil || v.Summary != q.Summary || !v.SpanStatus {
					t.Fatal("exploration measures lost")
				}
				if _, e := Compile(v); e != nil {
					t.Fatal(e)
				}
			}
			for _, v := range []Query{plan.Comparison.Baseline, plan.Comparison.Selection} {
				if v.Numeric != nil || v.NumericHistogram != nil || v.Summary != "" || v.SpanStatus || !reflect.DeepEqual(v.GroupBy, q.GroupBy) {
					t.Fatal("count projection wrong")
				}
				if _, e := Compile(v); e != nil {
					t.Fatal(e)
				}
			}
			for _, raw := range histogramValues() {
				for _, inside := range []bool{false, true} {
					row := map[Field]string{}
					if inside {
						row[*q.Where.Field] = "orders"
					}
					if raw != nil {
						row[q.Numeric.Field] = *raw
					}
					a, b, parent := Match(plan.Selected.Where, row), Match(plan.Rest.Where, row), Match(q.Where, row)
					if a && b || (a || b) != parent {
						t.Fatal("selected/rest fail full parent partition", typ, span, raw, a, b, parent)
					}
					if a != Match(plan.Comparison.Selection.Where, row) || b != Match(plan.Comparison.Baseline.Where, row) {
						t.Fatal("count projection changed population")
					}
				}
			}
			plan.Selected.Numeric.Unit = "mutated"
			plan.Parent.Where.Value = "mutated"
			if q.Numeric.Unit == "mutated" || q.Where.Value != "orders" || plan.Rest.Numeric.Unit == "mutated" {
				t.Fatal("plan aliases parent or sibling")
			}
		}
		for _, span := range [][2]int{{-1, 1}, {1, 1}, {2, 1}, {0, n + 1}} {
			if _, e := PlanNumericHistogramSelection(q, span[0], span[1]); e == nil {
				t.Fatal("invalid bin indexes accepted")
			}
		}
		// A valid parent at depth four must be refused when composition exceeds the
		// fixed budget, rather than dropping or weakening its selector.
		p := Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}
		for i := 0; i < 4; i++ {
			p = Predicate{All: []Predicate{p}}
		}
		q.Where = &p
		if q.Validate() != nil {
			t.Fatal("depth fixture invalid")
		}
		if _, e := PlanNumericHistogramSelection(q, 0, 1); e == nil {
			t.Fatal("selection silently exceeded predicate budget")
		}
	}
}

func TestNumericHistogramSourceCaptureSameRead(t *testing.T) {
	q := histogramQuestion("uint64")
	values := histogramValues()
	summary, _ := EvaluateNumericValues(*q.Numeric, values)
	hist, _ := EvaluateNumericHistogram(*q.Numeric, values)
	reads, countReads := 0, 0
	bad := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			countReads++
			if !strings.Contains(sql, "AS numeric_histogram") {
				t.Error("separate/missing histogram read")
			}
			h := detached(hist)
			if bad {
				h.Bins[0].Count = "99"
			}
			json.NewEncoder(w).Encode(map[string]any{"count": summary.Count, "numeric": summary, "numeric_histogram": h.Bins})
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL}
	r, e := source.Fetch(context.Background(), q)
	if e != nil || reads != 3 || countReads != 1 || !reflect.DeepEqual(r.NumericHistogram, &hist) || r.HistogramDataset == nil || r.HistogramDataset.Completeness == nil {
		t.Fatal("histogram not from main aggregate", reads, countReads, e)
	}
	store := CaptureStore{Dir: t.TempDir()}
	saved, e := store.Create(context.Background(), source, "Exact numeric histogram", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	before := reads
	loaded, e := store.Load(saved.ID)
	if e != nil || reads != before || !reflect.DeepEqual(loaded.Explore.NumericHistogram, &hist) {
		t.Fatal("capture lost or reread histogram", e)
	}
	loaded.Explore.NumericHistogram.Bins[0].Count = "99"
	if _, e := store.publish(*loaded); e == nil {
		t.Fatal("corrupt histogram capture published")
	}
	bad = true
	if _, e := source.Fetch(context.Background(), q); e == nil {
		t.Fatal("same-read histogram disagreement accepted")
	}
}

func TestClickHouseNumericHistogramParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	values := histogramValues()
	raw := []string{}
	for _, v := range values[1:] {
		raw = append(raw, *v)
	}
	data, _ := json.Marshal(raw)
	for _, signal := range []string{"spans", "logs"} {
		for _, kind := range []string{"attribute", "resource"} {
			for _, typ := range []string{"int64", "uint64"} {
				q := histogramQuestion(typ)
				q.Signal = signal
				q.Numeric.Field.Kind = kind
				c := compiler{signal: signal, params: map[string]string{}}
				aggregate := c.numericSQL(*q.Numeric) + ", " + c.numericHistogramSQL(*q.Numeric)
				array := c.param(string(data), "String")
				column := "SpanAttributes"
				if signal == "logs" {
					column = "LogAttributes"
				}
				if kind == "resource" {
					column = "ResourceAttributes"
				}
				sql := "SELECT toString(count()) AS count, " + aggregate + " FROM (SELECT if(number = 0, CAST(map(), 'Map(String, String)'), map(" + c.param(q.Numeric.Field.Key, "String") + ", JSONExtract(" + array + ", 'Array(String)')[number])) AS " + column + " FROM numbers(" + fmt.Sprint(len(values)) + ")) FORMAT JSONEachRow"
				rows, e := source.read(context.Background(), sql, c.params)
				if e != nil || len(rows) != 1 {
					t.Fatal(signal, kind, typ, e)
				}
				summary, e := decodeNumeric(rows[0], *q.Numeric, fmt.Sprint(len(values)))
				if e != nil {
					t.Fatal(e)
				}
				h, e := decodeNumericHistogram(rows[0], q, summary)
				if e != nil {
					t.Fatal(e)
				}
				want, _ := EvaluateNumericHistogram(*q.Numeric, values)
				wantSummary, _ := EvaluateNumericValues(*q.Numeric, values)
				if !reflect.DeepEqual(h, &want) || *summary != wantSummary {
					t.Fatal("real ClickHouse/reference histogram mismatch", signal, kind, typ)
				}
			}
		}
	}
}

func TestNumericHistogramComparisonDimension(t *testing.T) {
	for _, fields := range [][]Field{nil, {{Kind: "attribute", Key: "region"}}, {{Kind: "attribute", Key: "region"}, {Kind: "resource", Key: "service.namespace"}}} {
		q := histogramQuestion("uint64")
		q.GroupBy = fields
		plan, e := PlanNumericHistogramSelection(q, 0, 1)
		if e != nil {
			t.Fatal(e)
		}
		expected := Field{Kind: "column", Key: "service"}
		if len(fields) > 0 {
			expected = fields[0]
		}
		for _, projected := range []Query{plan.Comparison.Baseline, plan.Comparison.Selection} {
			if len(projected.GroupBy) != 1 || projected.GroupBy[0] != expected {
				t.Fatal("comparison requires shared dimension", projected.GroupBy)
			}
		}
		if !reflect.DeepEqual(plan.Selected.GroupBy, fields) || !reflect.DeepEqual(plan.Rest.GroupBy, fields) {
			t.Fatal("range question grouping changed")
		}
	}
}
