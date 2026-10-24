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

func distributionCohort(typ string, groups map[string][]*string) Cohort {
	c := comparisonNumericCohort(typ, groups)
	c.Question.NumericHistogram = &NumericHistogramSpec{Binning: "powers-of-two"}
	all := []*string{}
	for _, v := range groups {
		all = append(all, v...)
	}
	h, _ := EvaluateNumericHistogram(*c.Question.Numeric, all)
	c.NumericHistogram = &h
	return c
}
func distributionRequest(a, b Cohort) ComparisonRequest {
	return ComparisonRequest{Distribution: true, Numeric: true, Baseline: a, Selection: b}
}

func TestDistributionComparisonAlignedPopulationShares(t *testing.T) {
	a := distributionCohort("int64", map[string][]*string{"/a": append(numericValues("1", "", "bad"), nil)})
	b := distributionCohort("int64", map[string][]*string{"/a": numericValues("1", "1", "bad")})
	r, e := CompareCohorts(distributionRequest(a, b))
	if e != nil {
		t.Fatal(e)
	}
	d := r.Measures.Numeric.Distribution
	if d.State != "comparable" || d.Binning != "powers-of-two" || d.Denominator != "population-count" || d.BaselineCount != "4" || d.SelectionCount != "3" || len(d.Bins) != 128 {
		t.Fatal(d)
	}
	sumA, sumB := new(big.Rat), new(big.Rat)
	for _, bin := range d.Bins {
		ap, ok := new(big.Rat).SetString(bin.BaselinePercent)
		if !ok {
			t.Fatal(bin)
		}
		bp, _ := new(big.Rat).SetString(bin.SelectionPercent)
		sumA.Add(sumA, ap)
		sumB.Add(sumB, bp)
		if bin.From == "1" {
			if bin.To != "2" || bin.BaselineCount != "1" || bin.SelectionCount != "2" || bin.CountDifference != "1" || bin.BaselinePercent != "25" || bin.SelectionPercent != "200/3" || bin.DifferencePP != "125/3" {
				t.Fatal("bin comparison rounded or used valid denominator", bin)
			}
		}
	}
	if sumA.RatString() != "25" || sumB.RatString() != "200/3" {
		t.Fatal("population coverage hidden", sumA, sumB)
	}
	if len(r.Measures.Numeric.Coverage) != 4 || r.Measures.Numeric.Baseline.Missing != "1" {
		t.Fatal("parse states lost")
	}
	for _, typ := range []string{"int64", "uint64"} {
		values := numericValues("18446744073709551615")
		if typ == "int64" {
			values = numericValues("-9223372036854775808", "9223372036854775807")
		}
		side := distributionCohort(typ, map[string][]*string{"/a": values})
		r, e := CompareCohorts(distributionRequest(side, side))
		if e != nil {
			t.Fatal(e)
		}
		h := r.Measures.Numeric.Distribution
		index := len(h.Bins) - 1
		if h.Bins[index].SelectionCount != "1" || h.Bins[index].CountDifference != "0" {
			t.Fatal("extreme value lost", typ, h.Bins[index])
		}
		if typ == "int64" && h.Bins[0].BaselineCount != "1" {
			t.Fatal("signed minimum lost")
		}
	}
	a.NumericHistogram.Bins[0].Count = "mutated"
	if d.Bins[0].BaselineCount == "mutated" {
		t.Fatal("comparison aliases source histogram")
	}
}

func TestDistributionComparisonExplicitIntentAndValidation(t *testing.T) {
	a := distributionCohort("int64", map[string][]*string{"/a": numericValues("1")})
	for _, mutate := range []func(*ComparisonRequest){func(r *ComparisonRequest) { r.Numeric = false }, func(r *ComparisonRequest) { r.Distribution = false }, func(r *ComparisonRequest) { r.Selection.NumericHistogram = nil }, func(r *ComparisonRequest) { r.Selection.NumericHistogram.Bins[0].Count = "01" }, func(r *ComparisonRequest) { r.Selection.NumericHistogram.Bins = r.Selection.NumericHistogram.Bins[:1] }, func(r *ComparisonRequest) { r.Selection.NumericHistogram.Bins[0].From = "0" }, func(r *ComparisonRequest) { r.Selection.NumericHistogram.Bins[0].Count = "1" }, func(r *ComparisonRequest) { r.Selection.Question.NumericHistogram = nil }, func(r *ComparisonRequest) { r.Selection.Question.Numeric.Unit = "s" }} {
		request := detached(distributionRequest(a, a))
		mutate(&request)
		if _, e := CompareCohorts(request); e == nil {
			t.Fatal("invalid/implicit distribution accepted")
		}
	}
	// Earlier numeric comparisons may carry histogram questions, but their snapshot
	// did not carry bins. Neither their result shape nor hash acquires a new mode.
	legacy := detached(a)
	legacy.NumericHistogram = nil
	request := ComparisonRequest{Numeric: true, Baseline: legacy, Selection: legacy}
	r, e := CompareCohorts(request)
	if e != nil || r.Measures.Numeric.Distribution != nil {
		t.Fatal("legacy numeric mode upgraded", e)
	}
	wire, _ := json.Marshal(request)
	if strings.HasPrefix(string(wire), `{"distribution"`) {
		t.Fatal("false flag changed old wire bytes")
	}
	rAgain, e := CompareCohorts(detached(request))
	if e != nil || !reflect.DeepEqual(r, rAgain) {
		t.Fatal("legacy recomputation changed", e)
	}
}

func TestDistributionComparisonUnknownEmptyAndInvalidCoverage(t *testing.T) {
	valid := distributionCohort("int64", map[string][]*string{"/a": numericValues("1")})
	empty := distributionCohort("int64", map[string][]*string{})
	r, e := CompareCohorts(distributionRequest(empty, valid))
	if e != nil {
		t.Fatal(e)
	}
	d := r.Measures.Numeric.Distribution
	if d.State != "unavailable" || d.Reason != "empty-population" {
		t.Fatal(d)
	}
	for _, bin := range d.Bins {
		if bin.BaselineCount != "0" || bin.BaselinePercent != "" || bin.SelectionPercent == "" || bin.DifferencePP != "" || bin.CountDifference != bin.SelectionCount {
			t.Fatal("zero denominator treated as observed zero share", bin)
		}
	}
	invalid := distributionCohort("int64", map[string][]*string{"/a": numericValues("bad", "")})
	r, e = CompareCohorts(distributionRequest(invalid, invalid))
	if e != nil || r.Measures.Numeric.State != "unavailable" || r.Measures.Numeric.Distribution.State != "comparable" {
		t.Fatal("parse coverage conflated with scalar arithmetic", r, e)
	}
	for _, bin := range r.Measures.Numeric.Distribution.Bins {
		if bin.BaselinePercent != "0" || bin.CountDifference != "0" {
			t.Fatal("known empty bin not preserved", bin)
		}
	}
	changed := detached(valid)
	changed.State = "changed_between_reads"
	r, e = CompareCohorts(distributionRequest(valid, changed))
	if e != nil || r.State != "unknown" {
		t.Fatal(r, e)
	}
	d = r.Measures.Numeric.Distribution
	for _, bin := range d.Bins {
		if bin.BaselinePercent != "" || bin.SelectionPercent != "" || bin.DifferencePP != "" || bin.CountDifference != "" {
			t.Fatal("unknown distribution emitted changes", bin)
		}
	}
}

func TestDistributionSelectionPreservesSideAndRejectsEmptyUnknown(t *testing.T) {
	a := distributionCohort("uint64", map[string][]*string{"/a": numericValues("18446744073709551615")})
	b := distributionCohort("uint64", map[string][]*string{"/b": numericValues("1")})
	a.Question.Environment = "local"
	a.Question.Where = &Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "baseline-service"}
	b.Question.Where = &Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: "selection-service"}
	request := distributionRequest(a, b)
	p, e := PlanNumericDistributionSelection(request, "baseline", 64, 65)
	if e != nil || !reflect.DeepEqual(p.Parent, a.Question) || p.To != "18446744073709551616" {
		t.Fatal("wrong side/sentinel selection", p, e)
	}
	if p.DistributionComparison == nil || p.DistributionComparison.Baseline.NumericHistogram == nil || p.NumericComparison.Baseline.NumericHistogram != nil {
		t.Fatal("distribution and scalar intent conflated")
	}
	row := map[Field]string{*a.Question.Where.Field: "baseline-service", a.Question.Numeric.Field: "18446744073709551615"}
	if !Match(p.Selected.Where, row) {
		t.Fatal("selected side query lost exact maximum")
	}
	row[*a.Question.Where.Field] = "selection-service"
	if Match(p.Selected.Where, row) {
		t.Fatal("selection leaked other side")
	}
	for _, tc := range []struct {
		side     string
		from, to int
	}{{"selection", 64, 65}, {"other", 0, 1}, {"baseline", 0, 0}, {"baseline", -1, 1}, {"baseline", 0, 66}} {
		if _, e := PlanNumericDistributionSelection(request, tc.side, tc.from, tc.to); e == nil {
			t.Fatal("empty/invalid selection accepted", tc)
		}
	}
	changed := detached(request)
	changed.Selection.State = "changed_between_reads"
	if _, e := PlanNumericDistributionSelection(changed, "baseline", 64, 65); e == nil {
		t.Fatal("unknown comparison permitted selection")
	}
	empty := distributionCohort("uint64", map[string][]*string{})
	request.Baseline = empty
	if _, e := PlanNumericDistributionSelection(request, "baseline", 0, 65); e == nil {
		t.Fatal("empty side selected")
	}
	if _, e := PlanNumericDistributionSelection(request, "selection", 1, 2); e != nil {
		t.Fatal("nonempty opposite side incorrectly withheld", e)
	}
}

func TestDistributionCaptureIntentReopenRefreshAndEvidence(t *testing.T) {
	q := histogramQuestion("int64")
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	values := numericValues("-1", "1", "bad")
	summary, _ := EvaluateNumericValues(*q.Numeric, values)
	hist, _ := EvaluateNumericHistogram(*q.Numeric, values)
	reads := 0
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
			row := map[string]any{"count": "3", "numeric": summary}
			if strings.Contains(sql, "GROUP BY") {
				row["g0"] = "service"
			} else {
				row["numeric_histogram"] = hist.Bins
			}
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer server.Close()
	store := CaptureStore{Dir: t.TempDir()}
	source := Source{Endpoint: server.URL}
	question := CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Distribution: true, Baseline: q, Selection: q}}
	saved, e := store.Create(context.Background(), source, "Aligned distribution", question, "")
	if e != nil {
		t.Fatal(e)
	}
	if !saved.Comparison.Request.Distribution || saved.Comparison.Result.Measures.Numeric.Distribution == nil {
		t.Fatal("distribution intent not retained")
	}
	before := reads
	loaded, e := store.Load(saved.ID)
	if e != nil || reads != before || !reflect.DeepEqual(loaded.Comparison, saved.Comparison) {
		t.Fatal("reopen lost intent or read source", e)
	}
	refresh, e := store.Create(context.Background(), source, "Refreshed aligned distribution", question, saved.ID)
	if e != nil || !refresh.Comparison.Request.Distribution {
		t.Fatal("refresh dropped distribution intent", e)
	}
	oldQ := detached(question)
	oldQ.Comparison.Distribution = false
	old, e := store.Create(context.Background(), source, "Legacy numeric histogram question", oldQ, "")
	if e != nil || old.Comparison.Request.Distribution || old.Comparison.Request.Baseline.NumericHistogram != nil || old.Comparison.Result.Measures.Numeric.Distribution != nil {
		t.Fatal("old numeric mode upgraded", e)
	}
	oldRefresh, e := store.Create(context.Background(), source, "Old numeric refreshed", oldQ, old.ID)
	if e != nil || oldRefresh.Comparison.Request.Distribution || oldRefresh.Comparison.Request.Baseline.NumericHistogram != nil {
		t.Fatal("old refresh upgraded", e)
	}
	a, e := store.Load(saved.Comparison.BaselineID)
	if e != nil {
		t.Fatal(e)
	}
	b, e := store.Load(saved.Comparison.SelectionID)
	if e != nil {
		t.Fatal(e)
	}
	defaultResult, e := CompareEvidence(a, b)
	if e != nil || defaultResult.Distribution || defaultResult.Measures.Numeric.Distribution != nil {
		t.Fatal("default evidence comparison upgraded", e)
	}
	before = reads
	explicit, e := CompareDistributionEvidence(a, b)
	if e != nil || reads != before || !explicit.Distribution || explicit.Measures.Numeric.Distribution == nil {
		t.Fatal("explicit evidence distribution missing/read source", e)
	}
	plain, e := CohortFromResult(a.Explore)
	if e != nil || plain.NumericHistogram != nil {
		t.Fatal("default snapshot copied new payload", e)
	}
	full, e := CohortFromResultDistribution(a.Explore)
	if e != nil || full.NumericHistogram == nil {
		t.Fatal("explicit snapshot dropped payload", e)
	}
	a.Explore.NumericHistogram.Bins[0].Count = "mutated"
	if full.NumericHistogram.Bins[0].Count == "mutated" {
		t.Fatal("snapshot aliases retained histogram")
	}
	// Intent is validated independently of the optional computed result fields.
	tampered := detached(*saved)
	tampered.Question.Comparison.Distribution = false
	if _, e := store.publish(tampered); e == nil {
		t.Fatal("capture mode drift accepted")
	}
}

func TestDistributionAggregateOnlyIsExplicitAndSelectable(t *testing.T) {
	a := distributionCohort("int64", map[string][]*string{"/a": numericValues("-1", "1", "bad")})
	a.Question.GroupBy = nil
	a.Groups = nil
	r, e := CompareCohorts(distributionRequest(a, a))
	if e != nil || !r.AggregateOnly || r.Field != (Field{}) || len(r.Rows) != 0 || r.Measures.Numeric.Distribution == nil {
		t.Fatal("aggregate-only distribution not explicit", r, e)
	}
	if _, e := PlanNumericDistributionSelection(distributionRequest(a, a), "baseline", 63, 64); e != nil {
		t.Fatal("aggregate-only side cannot be opened", e)
	}
	for _, mutate := range []func(*ComparisonRequest){func(r *ComparisonRequest) {
		r.Distribution = false
		r.Baseline.NumericHistogram = nil
		r.Selection.NumericHistogram = nil
	}, func(r *ComparisonRequest) {
		r.Distribution = false
		r.Numeric = false
		r.Baseline = stripNumericCohort(r.Baseline)
		r.Selection = stripNumericCohort(r.Selection)
		r.Baseline.NumericHistogram = nil
		r.Selection.NumericHistogram = nil
	}, func(r *ComparisonRequest) { r.Selection.Question.GroupBy = []Field{{Kind: "column", Key: "service"}} }, func(r *ComparisonRequest) {
		r.Baseline.Groups = []ValueCount{{Value: str("fake"), Count: "3", Numeric: r.Baseline.Numeric}}
	}} {
		request := detached(distributionRequest(a, a))
		mutate(&request)
		if _, e := CompareCohorts(request); e == nil {
			t.Fatal("old mode or fabricated group silently accepted")
		}
	}
}

func TestDistributionAggregateOnlyCaptureSnapshotAndReopen(t *testing.T) {
	q := histogramQuestion("int64")
	values := numericValues("-1", "1", "bad")
	summary, _ := EvaluateNumericValues(*q.Numeric, values)
	h, _ := EvaluateNumericHistogram(*q.Numeric, values)
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		wire, _ := io.ReadAll(r.Body)
		sql := string(wire)
		switch {
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"1"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			return
		default:
			json.NewEncoder(w).Encode(map[string]any{"count": "3", "numeric": summary, "numeric_histogram": h.Bins})
		}
	}))
	defer server.Close()
	store := CaptureStore{Dir: t.TempDir()}
	question := CaptureQuestion{Kind: "comparison", Comparison: &ComparisonQuestion{Distribution: true, Baseline: q, Selection: q}}
	saved, e := store.Create(context.Background(), Source{Endpoint: server.URL}, "Aggregate distribution", question, "")
	if e != nil || !saved.Comparison.Result.AggregateOnly {
		t.Fatal(saved, e)
	}
	before := reads
	loaded, e := store.Load(saved.ID)
	if e != nil || reads != before || !reflect.DeepEqual(saved.Comparison, loaded.Comparison) {
		t.Fatal("aggregate-only capture drift or source read", e)
	}
	child, e := store.Load(saved.Comparison.BaselineID)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := CohortFromResult(child.Explore); e == nil {
		t.Fatal("legacy snapshot silently accepted ungrouped cohort")
	}
	snapshot, e := CohortFromResultDistribution(child.Explore)
	if e != nil || len(snapshot.Groups) != 0 || len(snapshot.Question.GroupBy) != 0 {
		t.Fatal("explicit snapshot invented grouping", e)
	}
	if _, e := PlanNumericDistributionSelection(distributionRequest(snapshot, snapshot), "selection", 65, 66); e != nil {
		t.Fatal("captured aggregate side selection failed", e)
	}
}
