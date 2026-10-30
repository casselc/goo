package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func comparisonNumericCohort(typ string, groups map[string][]*string) Cohort {
	q := histogramQuestion(typ)
	q.NumericHistogram = nil
	q.GroupBy = []Field{{Kind: "attribute", Key: "http.route"}}
	c := Cohort{Question: q, Basis: "numeric-fixture", State: "observed", Groups: []ValueCount{}}
	all := []*string{}
	for name, values := range groups {
		label := name
		s, _ := EvaluateNumericValues(*q.Numeric, values)
		c.Groups = append(c.Groups, ValueCount{Numeric: &s, Value: &label, Count: fmt.Sprint(len(values))})
		all = append(all, values...)
	}
	s, _ := EvaluateNumericValues(*q.Numeric, all)
	c.Numeric = &s
	c.Count = fmt.Sprint(len(all))
	if len(all) == 0 {
		c.State = "no_data"
	}
	return c
}
func stripNumericCohort(c Cohort) Cohort {
	c = detached(c)
	c.Numeric = nil
	for i := range c.Groups {
		c.Groups[i].Numeric = nil
	}
	return c
}

func TestNumericComparisonExactChangesAndCoverage(t *testing.T) {
	a := comparisonNumericCohort("int64", map[string][]*string{"/a": append(numericValues("1", "2", "bad", ""), nil)})
	b := comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("-2", "-2", "1")})
	r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
	if e != nil {
		t.Fatal(e)
	}
	n := r.Measures.Numeric
	if r.State != "ok" || n.State != "comparable" || n.SumDifference != "-6" || n.MinDifference != "-3" || n.MaxDifference != "-1" || n.MeanDifference != "-5/2" {
		t.Fatal("inexact numeric changes", r, n)
	}
	for i, row := range n.Coverage {
		difference, pp := "-1", "-20"
		if i == 0 {
			difference = "1"
			pp = "60"
		}
		if row.Difference != difference || row.DifferencePP != pp {
			t.Fatal("coverage denominator changed", row)
		}
	}
	if len(r.Rows) != 1 || !reflect.DeepEqual(r.Rows[0].Measures.Numeric, n) {
		t.Fatal("group measures not compared")
	}
	for _, tc := range []struct {
		typ  string
		a, b []*string
		want string
	}{
		{"int64", numericValues("0", "1", "0"), numericValues("0", "1"), "1/6"},
		{"int64", numericValues("9007199254740992"), numericValues("9007199254740993"), "1"},
		{"int64", numericValues("-9223372036854775808"), numericValues("9223372036854775807"), "18446744073709551615"},
		{"uint64", numericValues("18446744073709551615"), numericValues("0"), "-18446744073709551615"},
	} {
		a := comparisonNumericCohort(tc.typ, map[string][]*string{"/a": tc.a})
		b := comparisonNumericCohort(tc.typ, map[string][]*string{"/a": tc.b})
		r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
		if e != nil || r.Measures.Numeric.MeanDifference != tc.want {
			t.Fatal(tc, e, r)
		}
	}
	// Fractional coverage differences stay reduced rationals, not rounded decimals.
	a = comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("1", "bad", "bad")})
	b = comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("1", "bad")})
	r, e = CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
	if e != nil || r.Measures.Numeric.Coverage[0].DifferencePP != "50/3" {
		t.Fatal("rounded coverage", r, e)
	}
}

func TestNumericComparisonRequiresIntentBindingAndConservation(t *testing.T) {
	a := comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("1", "2")})
	if _, e := CompareCohorts(ComparisonRequest{Baseline: a, Selection: a}); e == nil {
		t.Fatal("numeric payload silently used in count mode")
	}
	for _, mutate := range []func(*Cohort){func(c *Cohort) { c.Numeric = nil }, func(c *Cohort) { c.Groups[0].Numeric = nil }, func(c *Cohort) { c.Numeric.Count = "99" }, func(c *Cohort) {
		s, _ := EvaluateNumericValues(*c.Question.Numeric, numericValues("100", "200"))
		c.Groups[0].Numeric = &s
	}, func(c *Cohort) { c.Question.Numeric.Field.Kind = "resource" }, func(c *Cohort) { c.Question.Numeric.Field.Key = "other" }, func(c *Cohort) { c.Question.Numeric.Type = "uint64" }, func(c *Cohort) { c.Question.Numeric.Unit = "s" }, func(c *Cohort) { c.Question.Numeric.Parse = "other" }} {
		b := detached(a)
		mutate(&b)
		if _, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b}); e == nil {
			t.Fatal("invalid numeric comparison accepted", b)
		}
	}
	// A declared family must not silently disappear because evidence or opt-in
	// was omitted. Count-only callers need an explicit count-only question.
	missing := stripNumericCohort(a)
	if _, err := CompareCohorts(ComparisonRequest{Baseline: missing, Selection: missing}); err == nil || !strings.Contains(err.Error(), "explicit numeric comparison") {
		t.Fatal("declared numeric profile silently dropped", err)
	}
	if _, err := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: missing, Selection: missing}); err == nil || !strings.Contains(err.Error(), "validated numeric summaries") {
		t.Fatal("missing numeric evidence accepted", err)
	}
	countOnly := detached(missing)
	countOnly.Question.Numeric = nil
	result, err := CompareCohorts(ComparisonRequest{Baseline: countOnly, Selection: countOnly})
	if err != nil || result.Measures != nil || result.ViewComparison == nil || len(result.ViewComparison.Baseline.Fields) != 1 {
		t.Fatal("explicit count-only recipe refused or acquired numeric fields", result, err)
	}
}

func TestNumericComparisonUnknownUnavailableAndMissingGroups(t *testing.T) {
	a := comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("1")})
	for _, tc := range []struct {
		values        []*string
		state, reason string
	}{
		{nil, "unavailable", "empty-population"}, {numericValues("bad", ""), "unavailable", "no-valid-values"},
	} {
		b := comparisonNumericCohort("int64", map[string][]*string{})
		if len(tc.values) > 0 {
			b = comparisonNumericCohort("int64", map[string][]*string{"/a": tc.values})
		}
		r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
		if e != nil {
			t.Fatal(e)
		}
		n := r.Measures.Numeric
		if n.State != tc.state || n.Reason != tc.reason || n.SumDifference != "" || n.MeanDifference != "" || n.MinDifference != "" || n.MaxDifference != "" {
			t.Fatal("unavailable numeric value fabricated", n)
		}
	}
	b := comparisonNumericCohort("int64", map[string][]*string{"/b": numericValues("2")})
	r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
	if e != nil || len(r.Rows) != 2 {
		t.Fatal(r, e)
	}
	for _, row := range r.Rows {
		n := row.Measures.Numeric
		if n.State != "unavailable" || n.Reason != "absent-group" || n.MeanDifference != "" || n.Baseline != nil && n.Selection != nil {
			t.Fatal("absent group fabricated observed summary", n)
		}
	}
	b = detached(a)
	b.State = "changed_between_reads"
	s, _ := EvaluateNumericValues(*b.Question.Numeric, numericValues("9"))
	b.Groups[0].Numeric = &s
	r, e = CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
	if e != nil || r.State != "unknown" {
		t.Fatal(r, e)
	}
	n := r.Measures.Numeric
	if n.State != "unknown" || n.MeanDifference != "" || n.SumDifference != "" || n.Baseline == nil || n.Selection == nil {
		t.Fatal(n)
	}
	for _, v := range n.Coverage {
		if v.Difference != "" || v.DifferencePP != "" {
			t.Fatal("unknown emitted coverage delta")
		}
	}
}

func TestNumericComparisonLargeProductsAndDetachedSnapshots(t *testing.T) {
	c := comparisonNumericCohort("int64", map[string][]*string{"/a": numericValues("0")})
	count := "18446744073709551615"
	makeSide := func(value string) Cohort {
		x := detached(c)
		v, _ := new(big.Int).SetString(value, 10)
		n, _ := new(big.Int).SetString(count, 10)
		s := NumericSummary{State: "observed", Unit: x.Question.Numeric.Unit, Count: count, Valid: count, Missing: "0", Empty: "0", Invalid: "0", Sum: new(big.Int).Mul(v, n).String(), Min: value, Max: value}
		x.Numeric = &s
		x.Count = count
		x.Groups[0].Numeric = detached(&s)
		x.Groups[0].Count = count
		return x
	}
	a, b := makeSide("-9223372036854775808"), makeSide("9223372036854775807")
	r, e := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: b})
	if e != nil {
		t.Fatal(e)
	}
	n := r.Measures.Numeric
	if n.MeanDifference != "18446744073709551615" || n.SumDifference != "340282366920938463426481119284349108225" {
		t.Fatal("bounded intermediate overflow", n)
	}
	a.Numeric.Sum = "mutated"
	if n.Baseline.Sum == "mutated" {
		t.Fatal("comparison aliases caller summary")
	}
}

func TestNumericComparisonCapturesAndHistogramPlans(t *testing.T) {
	q := histogramQuestion("int64")
	q.NumericHistogram = nil
	q.GroupBy = []Field{{Kind: "attribute", Key: "http.route"}}
	values := numericValues("1", "2", "bad")
	summary, _ := EvaluateNumericValues(*q.Numeric, values)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		data, _ := io.ReadAll(r.Body)
		sql := string(data)
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			row := map[string]any{"count": "3", "numeric": summary}
			if strings.Contains(sql, "GROUP BY") {
				row["g0"] = "/a"
			}
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer server.Close()
	store := CaptureStore{Dir: t.TempDir()}
	source := Source{Endpoint: server.URL}
	incompatible := detached(q)
	incompatible.Numeric.Unit = "s"
	if _, e := store.Create(context.Background(), source, "Incompatible", CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: q, Selection: incompatible}}, ""); e == nil || requests != 0 {
		t.Fatal("incompatible numeric capture read the source", e, requests)
	}
	saved, e := store.Create(context.Background(), source, "Numeric comparison", CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Baseline: q, Selection: q}}, "")
	if e != nil {
		t.Fatal(e)
	}
	if !saved.Comparison.Request.Numeric || saved.Comparison.Result.Measures.Numeric.MeanDifference != "0" {
		t.Fatal("new capture discarded numeric aggregates")
	}
	before := requests
	reopened, e := store.Load(saved.ID)
	if e != nil || requests != before || !reflect.DeepEqual(saved.Comparison, reopened.Comparison) {
		t.Fatal("reopen recomputed source or lost numeric payload", e)
	}
	// Refresh follows the current typed question; it never strips a declared
	// family to imitate a previous result's missing numeric opt-in.
	fresh, e := store.Create(context.Background(), source, "Refreshed numeric comparison", saved.Question, saved.ID)
	if e != nil || !fresh.Comparison.Request.Numeric || fresh.Comparison.Result.Measures.Numeric == nil || fresh.Comparison.Result.ViewComparison == nil {
		t.Fatal("numeric refresh dropped the active measurement family", fresh, e)
	}
	children := []*SavedCapture{}
	for _, id := range []string{saved.Comparison.BaselineID, saved.Comparison.SelectionID} {
		v, e := store.Load(id)
		if e != nil {
			t.Fatal(e)
		}
		children = append(children, v)
	}
	ungrouped := detached(*children[1])
	ungrouped.Question.Explore.GroupBy = nil
	ungrouped.Explore.Question.GroupBy = nil
	ungrouped.Explore.Groups = nil
	left := detached(*children[0])
	left.Question.Explore.GroupBy = nil
	left.Explore.Question.GroupBy = nil
	left.Explore.Groups = nil
	evidence, e := CompareEvidence(&left, &ungrouped)
	if e != nil || evidence.State != "comparable" || evidence.Measures.Numeric.MeanDifference != "0" || evidence.Breakdown != nil {
		t.Fatal("ungrouped numeric evidence unsupported", e)
	}
	right := detached(*children[1])
	right.Question.Explore.Where = numericPredicateFixture("int64", "gte", "0")
	right.Explore.Question.Where = detached(right.Question.Explore.Where)
	evidence, e = CompareEvidence(children[0], &right)
	if e != nil || evidence.State != "comparable" || evidence.Breakdown.Measures.Numeric == nil {
		t.Fatal("numeric selector comparison withheld", e)
	}
	right.Question.Explore.Numeric.Unit = "s"
	right.Explore.Question.Numeric.Unit = "s"
	right.Explore.Numeric.Unit = "s"
	for i, raw := range right.Explore.Groups {
		var g ValueCount
		json.Unmarshal(raw, &g)
		g.Numeric.Unit = "s"
		right.Explore.Groups[i], _ = json.Marshal(g)
	}
	evidence, e = CompareEvidence(children[0], &right)
	if e != nil || evidence.State != "incompatible" || evidence.Measures != nil {
		t.Fatal("binding mismatch compared", e)
	}
	histogram := histogramQuestion("int64")
	histogram.GroupBy = q.GroupBy
	plan, e := PlanNumericHistogramSelection(histogram, 64, 66)
	if e != nil || plan.NumericComparison == nil || plan.NumericComparison.Baseline.Numeric == nil || plan.NumericComparison.Selection.Numeric == nil || plan.NumericComparison.Baseline.NumericHistogram != nil || plan.Comparison.Baseline.Numeric != nil {
		t.Fatal("numeric/count histogram intent conflated", e)
	}
}

func TestNumericComparisonHistogramSharedGrouping(t *testing.T) {
	for _, groups := range [][]Field{nil, {{Kind: "attribute", Key: "http.route"}, {Kind: "resource", Key: "service.namespace"}}} {
		q := histogramQuestion("int64")
		q.GroupBy = groups
		p, e := PlanNumericHistogramSelection(q, 64, 66)
		if e != nil {
			t.Fatal(e)
		}
		want := Field{Kind: "column", Key: "service"}
		if len(groups) > 0 {
			want = groups[0]
		}
		for _, v := range []Query{p.NumericComparison.Baseline, p.NumericComparison.Selection} {
			if len(v.GroupBy) != 1 || v.GroupBy[0] != want || v.Numeric == nil || v.NumericHistogram != nil {
				t.Fatal("numeric comparison lacks shared single grouping", v)
			}
		}
	}
}
