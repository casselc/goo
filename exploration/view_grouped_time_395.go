package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
)

const MaxGroupedTimeSeriesDerivedCells = MaxTimeSeriesGroups * 240

// GroupedCountTimelineQuery is the finite reusable grouped-count profile. The
// grouping and bucket width belong to the query, never to a renderer's labels.
func GroupedCountTimelineQuery(q Query) bool {
	if q.Validate() != nil || !q.GroupTimeSeries || q.BucketMS <= 0 || len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.ReportedEdge != nil {
		return false
	}
	field := q.GroupBy[0]
	return !(field.Kind == "column" && (field.Key == "body_pattern_v1" || field.Key == "duration_bucket"))
}

// groupedTimeViewDataset consumes the existing exhaustive bounded source read.
// Only positively observed categories get rows. Missing bins become known zero
// only when the independent aggregate reads agree; disagreement retains positive
// diagnostics while filled absences are unknown, never fabricated zero evidence.
func groupedTimeViewDataset(r *Result, exhaustive bool) *ViewDataset {
	if r == nil || !GroupedCountTimelineQuery(r.Question) || ValidateGroupedTimeSeries(r.Question, r.GroupedTimeSeries) != nil || ValidateTimeSeries(r.Question, r.TimeSeries) != nil {
		return nil
	}
	q := r.Question
	d := baseViewDataset(r, "grouped-time-series", "grouped-intervals")
	if d == nil {
		return nil
	}
	d.Dimensions = []string{"category"}
	d.Interval = &ViewInterval{FromMS: strconv.FormatInt(q.FromMS, 10), ToMS: strconv.FormatInt(q.ToMS, 10), BucketMS: strconv.FormatInt(q.BucketMS, 10)}
	d.Fields = []ViewField{{Key: "category", Label: q.GroupBy[0].Kind + ":" + q.GroupBy[0].Key, Type: "string", Nullable: true}, {Key: "fromMs", Label: "Interval start", Type: "timestamp", Unit: "ms"}, {Key: "toMs", Label: "Interval end", Type: "timestamp", Unit: "ms"}, {Key: "count", Label: "Captured records", Type: "number", Unit: "records", Nullable: true}}
	coherent := exhaustive && (r.State == "observed" || r.State == "no_data") && groupedSeriesAgree(r) && aggregateSeriesAgree(r.Count, nil, nil, r.TimeSeries)
	for _, group := range r.GroupedTimeSeries {
		for _, point := range group.Points {
			from, to := strconv.FormatInt(point.FromMS, 10), strconv.FormatInt(point.ToMS, 10)
			category := group.Values[0]
			if category != nil {
				category = viewText(*category)
			}
			identity, _ := json.Marshal([]any{category, from, to})
			digest := sha256.Sum256(identity)
			values := ViewValues{"category": category, "fromMs": viewText(from), "toMs": viewText(to), "count": viewText(point.Count)}
			state := point.State
			if !coherent && point.Count == "0" {
				state = "unknown"
				values["count"] = nil
			}
			d.Rows = append(d.Rows, ViewRow{Key: "group-interval:" + hex.EncodeToString(digest[:]), Values: values, State: state})
		}
	}
	if len(d.Rows) > MaxGroupedTimeSeriesDerivedCells {
		return nil
	}
	if coherent {
		attestView(d, "grouped-intervals")
	}
	return d
}
