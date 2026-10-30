package exploration

import (
	"math/big"
	"strconv"
)

// ViewHistogram discriminates fixed numeric bins from raw-span duration bins.
// Numeric Support retains whole-population parse coverage; duration Total is
// the sum of the same exhaustive group read, without numeric parsing metadata.
type ViewHistogram struct {
	Kind    string         `json:"kind"`
	Unit    string         `json:"unit,omitempty"`
	Total   string         `json:"total,omitempty"`
	Binning string         `json:"binning"`
	Binding NumericBinding `json:"binding,omitzero"`
	Support ViewSupport    `json:"support,omitzero"`
	Closure string         `json:"closure"`
}

// histogramViewDataset projects only the already validated count aggregate.
// Fixed bins include zeroes and exclusive max+1 boundaries; no record page,
// separately read groups or additional source query establishes its population.
func histogramViewDataset(r *Result) *ViewDataset {
	return buildHistogramViewDataset(r, true)
}

func buildHistogramViewDataset(r *Result, source bool) *ViewDataset {
	q := r.Question
	if q.NumericHistogram == nil || q.Numeric == nil || validateOptionalNumeric(q.Numeric, r.Numeric, r.Count) != nil || validateOptionalNumericHistogram(q, r.NumericHistogram, r.Numeric) != nil {
		return nil
	}
	if r.State != "observed" && r.State != "no_data" && r.State != "changed_between_reads" || r.State == "observed" && r.Count == "0" || r.State == "no_data" && r.Count != "0" || source && r.CapturedAt == "" || !validCohortClock(Cohort{CapturedAt: r.CapturedAt}) {
		return nil
	}
	d := baseViewDataset(r, "histogram", "bins")
	if d == nil {
		return nil
	}
	b := *q.Numeric
	n := r.Numeric
	d.Histogram = &ViewHistogram{Kind: "numeric", Binning: r.NumericHistogram.Binning, Binding: b, Support: ViewSupport{Total: n.Count, Valid: n.Valid, Missing: n.Missing, Empty: n.Empty, Invalid: n.Invalid}, Closure: "left-closed-right-open"}
	d.Fields = []ViewField{{Key: "from", Label: "Bin start (inclusive)", Type: "rational", Unit: b.Unit}, {Key: "to", Label: "Bin end (exclusive)", Type: "rational", Unit: b.Unit}, {Key: "count", Label: "Valid captured records", Type: "number", Unit: "records"}}
	scale := big.NewInt(1)
	if b.Type == "decimal64" {
		scale.Exp(big.NewInt(10), big.NewInt(int64(b.Scale)), nil)
	}
	for i, bin := range r.NumericHistogram.Bins {
		state := "observed"
		if bin.Count == "0" {
			state = "no_data"
		}
		d.Rows = append(d.Rows, ViewRow{Key: "bin:" + strconv.Itoa(i), State: state, Values: ViewValues{"from": viewRational(bin.From, scale), "to": viewRational(bin.To, scale), "count": viewText(bin.Count)}})
	}
	// The histogram and coverage came from one aggregate even when another
	// result aspect disagreed. Conservatively retain the overall disagreement.
	if source && r.State != "changed_between_reads" {
		attestView(d, "bins")
	}
	return d
}
