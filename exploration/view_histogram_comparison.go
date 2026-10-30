package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// ViewHistogramComparison keeps two independent captured populations. Shares
// remain a display choice over each operand's total or valid support; this is
// never a merged histogram, a pooled statistic or a source authority grant.
type ViewHistogramComparison struct {
	Version   string       `json:"version"`
	Baseline  *ViewDataset `json:"baseline"`
	Selection *ViewDataset `json:"selection"`
	State     string       `json:"state"`
	Reason    string       `json:"reason,omitempty"`
}

func heldHistogramView(c Cohort) *ViewDataset {
	if !validCohortClock(c) {
		return nil
	}
	canonical := detached(c)
	sort.Slice(canonical.Groups, func(i, j int) bool { return valueKey(canonical.Groups[i].Value) < valueKey(canonical.Groups[j].Value) })
	// Capture time remains descriptive metadata, not content identity.
	capturedAt := canonical.CapturedAt
	canonical.CapturedAt = ""
	raw, err := json.Marshal(canonical)
	if err != nil {
		return nil
	}
	h := sha256.Sum256(raw)
	r := &Result{Question: canonical.Question, Basis: hex.EncodeToString(h[:]), CapturedAt: capturedAt, Count: canonical.Count, State: canonical.State, Numeric: canonical.Numeric, NumericHistogram: canonical.NumericHistogram}
	d := buildHistogramViewDataset(r, false)
	if d == nil {
		return nil
	}
	d.Origin = "browser-held"
	d.SourceBasis = canonical.Basis
	d.Quality.Extent = "unknown"
	d.Completeness = nil
	return d
}

// Called only after the existing comparison validator has admitted the entire
// numeric binding, signal, environment and both complete fixed bin layouts.
func cohortHistogramComparison(q ComparisonRequest) *ViewHistogramComparison {
	if !q.Distribution {
		return nil
	}
	a, b := heldHistogramView(q.Baseline), heldHistogramView(q.Selection)
	if a == nil || b == nil {
		return nil
	}
	c := &ViewHistogramComparison{Version: "goo.histogram-comparison/1", Baseline: a, Selection: b, State: "observed"}
	if q.Baseline.State == "changed_between_reads" || q.Selection.State == "changed_between_reads" {
		c.State = "diagnostic"
		c.Reason = "source-reads-disagree"
	} else if a.Histogram.Support.Total == "0" || b.Histogram.Support.Total == "0" {
		c.State = "no_data"
	}
	return c
}
