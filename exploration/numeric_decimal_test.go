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

func decimalBinding(scale int) NumericBinding {
	return NumericBinding{Field: Field{Kind: "attribute", Key: "cost"}, Type: "decimal64", Scale: scale, Unit: "USD", Parse: "canonical-decimal"}
}
func decimalValues(scale int) []*string {
	out := []*string{nil}
	valid := []string{"0", "1", "-1", "100", "-100", "9007199254740993", "9223372036854775807", "-9223372036854775808"}
	for _, v := range valid {
		n, _ := new(big.Int).SetString(v, 10)
		s := formatFixedCoefficient(n, scale)
		out = append(out, &s)
	}
	invalid := []string{"", "0", "1", "1.", ".1", "+0." + strings.Repeat("0", scale), "-0." + strings.Repeat("0", scale), "00." + strings.Repeat("0", scale), "1e2", formatFixedCoefficient(big.NewInt(1), scale) + "0", formatFixedCoefficient(big.NewInt(1), scale) + "\n", strings.Repeat("1", 1000)}
	lo := new(big.Int).Sub(big.NewInt(-9223372036854775808), big.NewInt(1))
	hi := new(big.Int).Add(big.NewInt(9223372036854775807), big.NewInt(1))
	invalid = append(invalid, formatFixedCoefficient(lo, scale), formatFixedCoefficient(hi, scale))
	for _, s := range invalid {
		v := s
		out = append(out, &v)
	}
	return out
}

func TestNumericDecimalCoefficientGrammarAndLegacyWire(t *testing.T) {
	for scale := 1; scale <= 9; scale++ {
		b := decimalBinding(scale)
		if b.Validate("spans") != nil {
			t.Fatal(b)
		}
		values := decimalValues(scale)
		s, e := EvaluateNumericValues(b, values)
		if e != nil || s.Scale != scale || s.Valid != "8" || s.Missing != "1" || s.Empty != "1" || s.Min != "-9223372036854775808" || s.Max != "9223372036854775807" || s.Validate(b) != nil {
			t.Fatal(scale, s, e)
		}
		for _, raw := range values {
			row := map[Field]string{}
			if raw != nil {
				row[b.Field] = *raw
			}
			matches := 0
			for _, op := range []string{"valid", "missing", "empty", "invalid"} {
				if Match(&Predicate{Numeric: &b, Op: op}, row) {
					matches++
				}
			}
			if matches != 1 {
				t.Fatal("decimal parse coverage does not partition", scale, raw)
			}
		}
		empty, _ := EvaluateNumericValues(b, nil)
		if empty.Scale != scale || empty.Sum != "0" || empty.Validate(b) != nil {
			t.Fatal("empty merge identity lost scale")
		}
		part, _ := EvaluateNumericValues(b, numericValues(formatFixedCoefficient(big.NewInt(1), scale)))
		merged := mergeNumeric(b, []*NumericSummary{&empty, &part})
		if merged.Scale != scale || merged.Sum != "1" || merged.Validate(b) != nil {
			t.Fatal("merged scale missing", merged)
		}
	}
	for _, b := range []NumericBinding{decimalBinding(0), decimalBinding(10), {Field: Field{Kind: "attribute", Key: "x"}, Type: "int64", Scale: 2, Unit: "1", Parse: "canonical-decimal"}} {
		if b.Validate("logs") == nil {
			t.Fatal("invalid scale accepted", b)
		}
	}
	integer := numericBinding()
	wire, _ := json.Marshal(integer)
	if string(wire) != `{"field":{"kind":"attribute","key":"db.response.returned_rows"},"type":"int64","unit":"{row}","parse":"canonical-decimal"}` {
		t.Fatal("legacy integer binding bytes changed", string(wire))
	}
	s, _ := EvaluateNumericValues(integer, numericValues("1"))
	wire, _ = json.Marshal(s)
	if strings.Contains(string(wire), "scale") {
		t.Fatal("integer summary acquired scale")
	}
	var explicit NumericBinding
	if json.Unmarshal([]byte(`{"field":{"kind":"attribute","key":"x"},"type":"int64","scale":0,"unit":"1","parse":"canonical-decimal"}`), &explicit) != nil || explicit.Validate("logs") != nil {
		t.Fatal("explicit scale zero not semantic omission")
	}
	b := decimalBinding(2)
	s, _ = EvaluateNumericValues(b, numericValues("1.20"))
	s.Scale = 0
	if s.Validate(b) == nil {
		t.Fatal("external summary missing decimal scale accepted")
	}
}

func TestNumericDecimalPredicatesHistogramAndComparison(t *testing.T) {
	b := decimalBinding(2)
	q := basic()
	q.Numeric = &b
	q.NumericHistogram = &NumericHistogramSpec{Binning: "powers-of-two"}
	q.Where = &Predicate{Numeric: &b, Op: "gte", Value: "0.01"}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.Count, "toInt64OrNull") || strings.Contains(p.Count, "Float") || len(p.Count) > 250000 {
		t.Fatal("unsafe/unbounded decimal lowering", e, len(p.Count))
	}
	found := false
	for _, value := range p.Parameters {
		found = found || value == "1"
	}
	if !found {
		t.Fatal("threshold was not a typed coefficient")
	}
	for _, threshold := range []string{"1", "0.1", "0.010", "-0.00", "92233720368547758.08"} {
		q.Where.Value = threshold
		if _, e := Compile(q); e == nil {
			t.Fatal("invalid decimal threshold compiled", threshold)
		}
	}
	q.Where = nil
	h, e := EvaluateNumericHistogram(b, decimalValues(2))
	summary, _ := EvaluateNumericValues(b, decimalValues(2))
	if e != nil || h.Validate(b, summary) != nil || len(h.Bins) != 128 || h.Bins[0].From != "-9223372036854775808" || h.Bins[127].To != "9223372036854775808" {
		t.Fatal("histogram coefficient contract changed", e)
	}
	plan, e := PlanNumericHistogramSelection(q, 64, 65)
	if e != nil {
		t.Fatal(e)
	}
	for _, raw := range []string{"0.00", "0.01", "-0.01", "bad"} {
		row := map[Field]string{b.Field: raw}
		a, z := Match(plan.Selected.Where, row), Match(plan.Rest.Where, row)
		if a == z || a != (raw == "0.00") {
			t.Fatal("decimal selected/rest lost coefficient bin", raw, a, z)
		}
	}
	tail, e := PlanNumericHistogramSelection(q, 127, 128)
	if e != nil || !Match(tail.Selected.Where, map[Field]string{b.Field: "92233720368547758.07"}) {
		t.Fatal("sentinel became invalid decimal threshold", e)
	}
	makeSide := func(values []*string) Cohort {
		s, _ := EvaluateNumericValues(b, values)
		label := "one"
		sideQ := q
		sideQ.NumericHistogram = nil
		sideQ.GroupBy = []Field{{Kind: "column", Key: "service"}}
		return Cohort{Question: sideQ, Count: s.Count, State: "observed", Basis: "basis", Numeric: &s, Groups: []ValueCount{{Value: &label, Count: s.Count, Numeric: &s}}}
	}
	a, z := makeSide(numericValues("0.00", "0.01", "0.00")), makeSide(numericValues("0.00", "0.01"))
	r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: z})
	if e != nil || r.Measures.Numeric.MeanDifference != "1/6" || r.Measures.Numeric.Binding.Scale != 2 || r.Measures.Numeric.Selection.Scale != 2 {
		t.Fatal("comparison stopped using coefficient rational", r, e)
	}
	different := detached(z)
	different.Question.Numeric.Scale = 3
	if _, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: different}); e == nil {
		t.Fatal("scale conversion inferred")
	}
}

func TestNumericDecimalSourceScaleAllAggregatePathsAndCapture(t *testing.T) {
	b := decimalBinding(2)
	q := basic()
	q.Numeric = &b
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.BucketMS = 10000
	q.GroupTimeSeries = true
	summary, _ := EvaluateNumericValues(b, numericValues("0.01", "-0.01"))
	withoutScale := summary
	withoutScale.Scale = 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		sql := string(raw)
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			row := map[string]any{"count": "2", "numeric": withoutScale}
			if strings.Contains(sql, "GROUP BY g0") {
				row["g0"] = "svc"
			}
			if strings.Contains(sql, "from_ms") {
				row["from_ms"] = "0"
			}
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL}
	r, e := source.Fetch(context.Background(), q)
	if e != nil || r.Numeric.Scale != 2 || r.State != "observed" {
		t.Fatal("source scale injection failed", r, e)
	}
	var group struct{ Numeric *NumericSummary }
	json.Unmarshal(r.Groups[0], &group)
	if group.Numeric.Scale != 2 {
		t.Fatal("raw group retained missing scale")
	}
	for _, p := range r.TimeSeries {
		if p.Numeric.Scale != 2 {
			t.Fatal("time-cell scale missing")
		}
	}
	for _, g := range r.GroupedTimeSeries {
		for _, p := range g.Points {
			if p.Numeric.Scale != 2 {
				t.Fatal("split-cell scale missing")
			}
		}
	}
	store := CaptureStore{Dir: t.TempDir()}
	saved, e := store.Create(context.Background(), source, "Decimal coefficient capture", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.Load(saved.ID)
	if e != nil || !reflect.DeepEqual(loaded.Explore.Numeric, saved.Explore.Numeric) {
		t.Fatal("capture decimal roundtrip", e)
	}
	pkg, e := DraftTelemetryPackage(q, "decimal", "Decimal binding", 1)
	if e != nil {
		t.Fatal(e)
	}
	pp, e := PlanTelemetryPackage(pkg, q.FromMS, q.ToMS, nil)
	if e != nil {
		t.Fatal(e)
	}
	again, e := ReadTelemetryPackage(pp.YAML)
	if e != nil || again.Numeric.Scale != 2 {
		t.Fatal("package lost scale", e)
	}
	bad := summary
	bad.Scale = 3
	raw, _ := json.Marshal(map[string]any{"numeric": bad})
	if _, e := decodeNumeric(raw, b, bad.Count); e == nil {
		t.Fatal("explicit wrong source scale silently injected")
	}
}

func TestClickHouseNumericDecimalParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL is not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	for _, scale := range []int{1, 2, 9} {
		for _, signal := range []string{"spans", "logs"} {
			for _, kind := range []string{"attribute", "resource"} {
				b := decimalBinding(scale)
				b.Field.Kind = kind
				values := decimalValues(scale)
				raw := []string{}
				for _, v := range values[1:] {
					raw = append(raw, *v)
				}
				data, _ := json.Marshal(raw)
				c := compiler{signal: signal, params: map[string]string{}}
				aggregate := c.numericSQL(b) + ", " + c.numericHistogramSQL(b)
				array := c.param(string(data), "String")
				column := "SpanAttributes"
				if signal == "logs" {
					column = "LogAttributes"
				}
				if kind == "resource" {
					column = "ResourceAttributes"
				}
				key := c.param(b.Field.Key, "String")
				inline := " FROM (SELECT number, if(number=0, CAST(map(), 'Map(String, String)'), map(" + key + ", JSONExtract(" + array + ", 'Array(String)')[number])) AS " + column + " FROM numbers(" + fmt.Sprint(len(values)) + "))"
				rows, e := source.read(context.Background(), "SELECT toString(count()) AS count, "+aggregate+inline+" FORMAT JSONEachRow", c.params)
				if e != nil || len(rows) != 1 {
					t.Fatal(scale, signal, kind, e)
				}
				got, e := decodeNumeric(rows[0], b, fmt.Sprint(len(values)))
				if e != nil {
					t.Fatal(e)
				}
				want, _ := EvaluateNumericValues(b, values)
				hist, e := decodeNumericHistogram(rows[0], Query{Signal: signal, Numeric: &b, NumericHistogram: &NumericHistogramSpec{Binning: "powers-of-two"}}, got)
				wantHist, _ := EvaluateNumericHistogram(b, values)
				if e != nil || *got != want || !reflect.DeepEqual(hist, &wantHist) {
					t.Fatal("ClickHouse decimal/reference disagreement", scale, signal, kind, got, want, e)
				}
				// Typed threshold transport must compare coefficients, including extrema and tiny negatives.
				for _, coefficient := range []string{"-9223372036854775808", "-1", "0", "1", "9223372036854775807"} {
					n, _ := new(big.Int).SetString(coefficient, 10)
					threshold := formatFixedCoefficient(n, scale)
					for _, op := range []string{"eq", "ne", "gte", "lt"} {
						p := Predicate{Numeric: &b, Op: op, Value: threshold}
						expr := c.predicate(&p)
						selected, e := source.read(context.Background(), "SELECT toString(number) AS n, "+expr+" AS matched"+inline+" ORDER BY number FORMAT JSONEachRow", c.params)
						if e != nil || len(selected) != len(values) {
							t.Fatal(scale, op, threshold, e)
						}
						for i, row := range selected {
							var v struct {
								N       string
								Matched *uint8
							}
							if json.Unmarshal(row, &v) != nil || v.Matched == nil {
								t.Fatal("nullable comparison leaked", string(row))
							}
							logical := map[Field]string{}
							if values[i] != nil {
								logical[b.Field] = *values[i]
							}
							if (*v.Matched == 1) != Match(&p, logical) {
								t.Fatal("decimal threshold parity", scale, op, threshold, i, string(row))
							}
						}
					}
				}
			}
		}
	}
}
