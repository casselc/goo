package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/scopedview"
	"math/bits"
	"strconv"
)

type ViewDataset = scopedview.Dataset
type ViewPartition = scopedview.Partition
type ViewInterval = scopedview.Interval
type ViewScope = scopedview.Scope
type ViewField = scopedview.Field
type ViewRow = scopedview.Row
type ViewValues = scopedview.Values
type ViewQuality = scopedview.Quality
type ViewCompleteness = scopedview.Completeness

func viewText(v string) *string { return &v }
func supportedViewMeasures(q Query) bool {
	return q.NumericHistogram == nil && !q.GroupTimeSeries
}
func baseViewDataset(r *Result, shape, aspect string) *ViewDataset {
	if !regexpRevision.MatchString(r.Basis) || r.Question.Validate() != nil {
		return nil
	}
	selector, err := json.Marshal(r.Question)
	if err != nil {
		return nil
	}
	h := sha256.Sum256(selector)
	id := hex.EncodeToString(h[:])
	return &ViewDataset{Version: "goo.dataset/1", ID: id + ":" + aspect, Revision: r.Basis, Scope: ViewScope{ID: id, Generation: r.Basis, Signal: r.Question.Signal, Selector: selector}, Shape: shape, Fields: []ViewField{}, Rows: []ViewRow{}, Quality: ViewQuality{Extent: "unknown", Basis: "captured-query", Value: "derived", Sampling: "unknown", Freshness: "fixed"}, CapturedAt: r.CapturedAt, Origin: "source-read"}
}
func attestView(d *ViewDataset, aspect string) {
	d.Quality.Extent = "complete"
	d.Completeness = &ViewCompleteness{Basis: "captured-query", Revision: d.Revision, ScopeID: d.Scope.ID, Aspects: []string{aspect}}
}
func categoryField(i int) string {
	if i == 0 {
		return "category"
	}
	return fmt.Sprintf("category%d", i)
}

// The source's LIMIT1001 sentinel rejects more than1000 groups. Matching a
// separately read count or page length never establishes this attestation.
func groupViewDataset(r *Result, exhaustiveGroups bool) *ViewDataset {
	q := r.Question
	if len(q.GroupBy) < 1 || len(q.GroupBy) > 2 || !supportedViewMeasures(q) || len(r.Groups) > 1000 {
		return nil
	}
	d := baseViewDataset(r, "ranked", "groups")
	if d == nil {
		return nil
	}
	for i, f := range q.GroupBy {
		d.Fields = append(d.Fields, ViewField{Key: categoryField(i), Label: f.Kind + ":" + f.Key, Type: "string", Nullable: true})
		d.Dimensions = append(d.Dimensions, categoryField(i))
	}
	d.Fields = append(d.Fields, ViewField{Key: "count", Label: "Captured records", Type: "number", Unit: "records"})
	d.Fields = append(d.Fields, viewMeasureFields(q)...)
	seen := map[string]bool{}
	var partitionTotal uint64
	partitionOverflow := false
	for _, raw := range r.Groups {
		var g map[string]json.RawMessage
		if json.Unmarshal(raw, &g) != nil {
			return nil
		}
		values := ViewValues{}
		tuple := []*string{}
		for i := range q.GroupBy {
			raw, ok := g[fmt.Sprintf("g%d", i)]
			if !ok {
				return nil
			}
			var v *string
			if json.Unmarshal(raw, &v) != nil || v != nil && len(*v) > 1024 {
				return nil
			}
			values[categoryField(i)] = v
			tuple = append(tuple, v)
		}
		var count string
		if json.Unmarshal(g["count"], &count) != nil {
			return nil
		}
		n, err := strconv.ParseUint(count, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != count {
			return nil
		}
		values["count"] = viewText(count)
		var carry uint64
		partitionTotal, carry = bits.Add64(partitionTotal, n, 0)
		partitionOverflow = partitionOverflow || carry != 0
		canonical, _ := json.Marshal(tuple)
		h := sha256.Sum256(canonical)
		key := "category:" + hex.EncodeToString(h[:])
		if seen[key] {
			return nil
		}
		seen[key] = true
		var summaries struct {
			Numeric    *NumericSummary    `json:"numeric"`
			Duration   *DurationSummary   `json:"duration"`
			SpanStatus *SpanStatusSummary `json:"spanStatus"`
		}
		if json.Unmarshal(raw, &summaries) != nil {
			return nil
		}
		cells, ok := viewMeasureCells(q, count, summaries.Numeric, summaries.Duration, summaries.SpanStatus, false, values)
		if !ok {
			return nil
		}
		d.Rows = append(d.Rows, ViewRow{Key: key, Values: values, Cells: cells})
	}
	if exhaustiveGroups && r.State != "changed_between_reads" {
		attestView(d, "groups")
		// No wrapped or separately read denominator can authorize population
		// shares. Exact individual rows remain usable if the sum is too large.
		if !partitionOverflow {
			d.Partition = &ViewPartition{Kind: "disjoint-record-groups", Measure: "count", Total: strconv.FormatUint(partitionTotal, 10)}
		}
	}
	return d
}

// timeViewDataset exposes the checked count intervals without deriving bins from
// record pages. DecodeTimeSeries alone may fill absent source buckets with zero:
// its exhaustive bounded source read establishes those no_data intervals.
func timeViewDataset(r *Result, exhaustiveIntervals bool) *ViewDataset {
	q := r.Question
	if q.BucketMS <= 0 || !supportedViewMeasures(q) || q.Validate() != nil || len(r.TimeSeries) > 240 {
		return nil
	}
	d := baseViewDataset(r, "time-series", "intervals")
	if d == nil {
		return nil
	}
	d.Interval = &ViewInterval{FromMS: strconv.FormatInt(q.FromMS, 10), ToMS: strconv.FormatInt(q.ToMS, 10), BucketMS: strconv.FormatInt(q.BucketMS, 10)}
	d.Fields = []ViewField{{Key: "fromMs", Label: "Interval start", Type: "timestamp", Unit: "ms"}, {Key: "toMs", Label: "Interval end", Type: "timestamp", Unit: "ms"}, {Key: "count", Label: "Captured records", Type: "number", Unit: "records", Nullable: true}}
	d.Fields = append(d.Fields, viewMeasureFields(q)...)
	previousEnd := q.FromMS
	for _, p := range r.TimeSeries {
		if p.FromMS < previousEnd || p.FromMS < q.FromMS || p.FromMS >= q.ToMS || (p.FromMS-q.FromMS)%q.BucketMS != 0 || p.ToMS != min(p.FromMS+q.BucketMS, q.ToMS) {
			return nil
		}
		values := ViewValues{"fromMs": viewText(strconv.FormatInt(p.FromMS, 10)), "toMs": viewText(strconv.FormatInt(p.ToMS, 10))}
		switch p.State {
		case "observed", "no_data":
			n, e := strconv.ParseUint(p.Count, 10, 64)
			if e != nil || strconv.FormatUint(n, 10) != p.Count || (p.State == "observed") != (n > 0) {
				return nil
			}
			values["count"] = viewText(p.Count)
		case "unknown":
			if p.Count != "" {
				return nil
			}
			values["count"] = nil
		default:
			return nil
		}
		from, to := *values["fromMs"], *values["toMs"]
		cells, ok := viewMeasureCells(q, p.Count, p.Numeric, p.Duration, p.SpanStatus, p.State == "unknown", values)
		if !ok {
			return nil
		}
		d.Rows = append(d.Rows, ViewRow{Key: "interval:" + from + ":" + to, State: p.State, Values: values, Cells: cells})
		previousEnd = p.ToMS
	}
	if exhaustiveIntervals && r.State != "changed_between_reads" && ValidateTimeSeries(q, r.TimeSeries) == nil {
		attestView(d, "intervals")
	}
	return d
}
