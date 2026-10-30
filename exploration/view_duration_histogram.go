package exploration

import (
	"encoding/json"
	"math/bits"
	"strconv"
)

// durationHistogramViewDataset interprets the compiler's fixed duration labels,
// never a numeric attribute parser or a native OTel histogram. Only the bounded
// exhaustive GROUP BY read authorizes absent bins to become known zeroes.
func durationHistogramViewDataset(r *Result, exhaustiveGroups bool) *ViewDataset {
	q := r.Question
	if !exhaustiveGroups || r.State == "changed_between_reads" || (r.State != "observed" && r.State != "no_data") || q.Signal != "spans" || q.Numeric != nil || q.NumericHistogram != nil || q.GroupTimeSeries || len(q.GroupBy) != 1 || q.GroupBy[0] != (Field{Kind: "column", Key: "duration_bucket"}) || r.CapturedAt == "" || !validCohortClock(Cohort{CapturedAt: r.CapturedAt}) {
		return nil
	}
	d := baseViewDataset(r, "histogram", "bins")
	if d == nil {
		return nil
	}
	counts := make([]string, len(durationLabels))
	seen := map[string]bool{}
	var total uint64
	for _, raw := range r.Groups {
		var g struct {
			Label *string `json:"g0"`
			Count string  `json:"count"`
		}
		if json.Unmarshal(raw, &g) != nil || g.Label == nil || seen[*g.Label] {
			return nil
		}
		seen[*g.Label] = true
		index := -1
		for i, label := range durationLabels {
			if *g.Label == label {
				index = i
				break
			}
		}
		if index < 0 {
			return nil
		}
		n, e := strconv.ParseUint(g.Count, 10, 64)
		if e != nil || strconv.FormatUint(n, 10) != g.Count {
			return nil
		}
		sum, carry := bits.Add64(total, n, 0)
		if carry != 0 {
			return nil
		}
		total = sum
		counts[index] = g.Count
	}
	if (r.State == "no_data" && total > 0) || (r.State == "observed" && total == 0) {
		return nil
	}
	d.Histogram = &ViewHistogram{Kind: "span-duration", Binning: "goo-duration-v1", Unit: "ns", Total: strconv.FormatUint(total, 10), Closure: "left-closed-right-open"}
	d.Fields = []ViewField{{Key: "from", Label: "Bin start (inclusive)", Type: "histogram-bound", Unit: "ns"}, {Key: "to", Label: "Bin end (exclusive)", Type: "histogram-bound", Unit: "ns"}, {Key: "count", Label: "Captured spans", Type: "number", Unit: "records"}}
	from := "0"
	for i := range durationLabels {
		to := "+inf"
		if i < len(durationBounds) {
			to = strconv.FormatUint(durationBounds[i], 10)
		}
		count := counts[i]
		if count == "" {
			count = "0"
		}
		state := "observed"
		if count == "0" {
			state = "no_data"
		}
		d.Rows = append(d.Rows, ViewRow{Key: "bin:" + strconv.Itoa(i), State: state, Values: ViewValues{"from": viewText(from), "to": viewText(to), "count": viewText(count)}})
		from = to
	}
	attestView(d, "bins")
	return d
}
