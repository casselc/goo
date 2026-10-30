package exploration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestViewHistogramComparisonIndependentOperands(t *testing.T) {
	a := distributionCohort("int64", map[string][]*string{"/a": numericValues("1", "bad", "")})
	b := distributionCohort("int64", map[string][]*string{"/a": numericValues("1", "1")})
	a.CapturedAt = "2026-10-30T01:00:00Z"
	b.CapturedAt = ""
	request := distributionRequest(a, b)
	r, e := CompareCohorts(request)
	if e != nil {
		t.Fatal(e)
	}
	c := r.HistogramComparison
	if c == nil || c.Version != "goo.histogram-comparison/1" || c.State != "observed" || c.Reason != "" {
		t.Fatal(c)
	}
	for _, d := range []*ViewDataset{c.Baseline, c.Selection} {
		if d.Shape != "histogram" || d.Origin != "browser-held" || d.Quality.Extent != "unknown" || d.Completeness != nil {
			t.Fatal("held source authority", d)
		}
	}
	if c.Baseline.Histogram.Support.Total != "3" || c.Baseline.Histogram.Support.Valid != "1" || c.Selection.Histogram.Support.Valid != "2" || c.Baseline.CapturedAt == c.Selection.CapturedAt {
		t.Fatal("operands collapsed")
	}
	oldRevision := c.Baseline.Revision
	request.Baseline.CapturedAt = "2026-10-31T01:00:00Z"
	next, e := CompareCohorts(request)
	if e != nil || next.HistogramComparison.Baseline.Revision != oldRevision {
		t.Fatal("clock changed content revision", e)
	}
	request.Baseline.NumericHistogram.Bins[0].Count = "mutation"
	if *c.Baseline.Rows[0].Values["count"] == "mutation" {
		t.Fatal("operand not detached")
	}
}

func TestViewHistogramComparisonStatesAndWire(t *testing.T) {
	fixtures := map[string]*ViewHistogramComparison{}
	for _, tc := range []struct {
		name   string
		values []*string
		state  string
		want   string
	}{{"observed", numericValues("1"), "observed", "observed"}, {"no-valid", numericValues("bad", ""), "observed", "observed"}, {"empty", nil, "no_data", "no_data"}, {"diagnostic", numericValues("1"), "changed_between_reads", "diagnostic"}} {
		a := distributionCohort("int64", map[string][]*string{"/a": tc.values})
		// Aggregate-only allows a truly empty operand without a fabricated category.
		a.Question.GroupBy = nil
		a.Groups = nil
		a.State = tc.state
		b := distributionCohort("int64", map[string][]*string{"/a": numericValues("1", "2")})
		b.Question.GroupBy = nil
		b.Groups = nil
		r, e := CompareCohorts(distributionRequest(a, b))
		if e != nil {
			t.Fatal(tc.name, e)
		}
		if r.HistogramComparison == nil || r.HistogramComparison.State != tc.want {
			t.Fatal(tc.name, r.HistogramComparison)
		}
		fixtures[tc.name] = r.HistogramComparison
	}
	if path := os.Getenv("GOO_HISTOGRAM_COMPARISON_FIXTURES"); path != "" {
		raw, e := json.Marshal(fixtures)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
}

func TestEvidenceHistogramComparisonTwoGroupSnapshot(t *testing.T) {
	makeCapture := func(values []*string) *SavedCapture {
		q := histogramQuestion("int64")
		q.GroupBy = []Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
		q.Limit = CapturePageSize
		n, e := EvaluateNumericValues(*q.Numeric, values)
		if e != nil {
			t.Fatal(e)
		}
		h, e := EvaluateNumericHistogram(*q.Numeric, values)
		if e != nil {
			t.Fatal(e)
		}
		raw, e := json.Marshal(map[string]any{"g0": "service", "g1": "operation", "count": n.Count, "numeric": n})
		if e != nil {
			t.Fatal(e)
		}
		state := "observed"
		if n.Count == "0" {
			state = "no_data"
		}
		return &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Two group histogram", CapturedAt: "2026-10-30T01:00:00Z", Question: CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &Result{Question: q, Count: n.Count, State: state, Basis: strings.Repeat("a", 64), CapturedAt: "2026-10-30T01:00:00Z", Numeric: &n, NumericHistogram: &h, Groups: []json.RawMessage{raw}}}
	}
	a, b := makeCapture(numericValues("1", "bad")), makeCapture(numericValues("2", "3"))
	before, _ := json.Marshal([]*SavedCapture{a, b})
	r, e := CompareDistributionEvidence(a, b)
	if e != nil {
		t.Fatal(e)
	}
	c := r.HistogramComparison
	if c == nil || c.State != "observed" || c.Baseline.SourceBasis != a.Explore.Basis || c.Baseline.CapturedAt != a.Explore.CapturedAt || len(c.Baseline.Scope.Selector) == 0 || len(c.Baseline.Rows) != 128 || c.Baseline.Completeness != nil || c.Baseline.Origin != "browser-held" {
		t.Fatal("stored histogram missing independent operands", c)
	}
	after, _ := json.Marshal([]*SavedCapture{a, b})
	if string(before) != string(after) {
		t.Fatal("stored captures mutated")
	}
	legacy, e := CompareEvidence(a, b)
	if e != nil || legacy.HistogramComparison != nil {
		t.Fatal("implicit upgrade", e)
	}
	if path := os.Getenv("GOO_EVIDENCE_HISTOGRAM_FIXTURE"); path != "" {
		raw, e := json.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(path, raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
	a.Explore.State = "changed_between_reads"
	diagnostic, e := CompareDistributionEvidence(a, b)
	if e != nil || diagnostic.HistogramComparison == nil || diagnostic.HistogramComparison.State != "diagnostic" || diagnostic.HistogramComparison.Reason != "source-reads-disagree" {
		t.Fatal("stored changed reads hidden", e)
	}
}
