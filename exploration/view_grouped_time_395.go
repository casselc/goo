package exploration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strconv"
)

const MaxGroupedTimeSeriesDerivedCells = MaxTimeSeriesGroups * 240

// GroupedCountTimelineQuery is the finite reusable grouped-count profile. The
// grouping and bucket width belong to the query, never to a renderer's labels.
func GroupedCountTimelineQuery(q Query) bool {
	if q.Validate() != nil || !q.GroupTimeSeries || q.BucketMS <= 0 || len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.LogSeverity || q.ReportedEdge != nil {
		return false
	}
	field := q.GroupBy[0]
	return !(field.Kind == "column" && (field.Key == "body_pattern_v1" || field.Key == "duration_bucket"))
}

// GroupedMeanDurationTimelineQuery admits only raw-span additive mean support.
func GroupedMeanDurationTimelineQuery(q Query) bool {
	if q.Signal != "spans" || q.Summary != "span-duration" {
		return false
	}
	q.Summary = ""
	return GroupedCountTimelineQuery(q)
}

// The source summaries retain exact duration sums and contributing counts.
// Validate each before checking agreement; malformed support never becomes a plot.
func groupedDurationSupportValid(r *Result) bool {
	if r.Duration == nil || r.Duration.Count != r.Count || r.Duration.Validate() != nil {
		return false
	}
	for _, raw := range r.Groups {
		var group struct {
			Count    string           `json:"count"`
			Duration *DurationSummary `json:"duration"`
		}
		if json.Unmarshal(raw, &group) != nil || group.Duration == nil || group.Duration.Count != group.Count || group.Duration.Validate() != nil {
			return false
		}
	}
	return true
}

// groupedTimeViewDataset consumes the existing exhaustive bounded source read.
// Only positively observed categories get rows. Missing bins become known zero
// only when the independent aggregate reads agree; disagreement retains positive
// diagnostics while filled absences are unknown, never fabricated zero evidence.
func groupedTimeViewDataset(r *Result, exhaustive bool) *ViewDataset {
	if r == nil || !(GroupedCountTimelineQuery(r.Question) || GroupedMeanDurationTimelineQuery(r.Question)) || ValidateGroupedTimeSeries(r.Question, r.GroupedTimeSeries) != nil || ValidateTimeSeries(r.Question, r.TimeSeries) != nil {
		return nil
	}
	q := r.Question
	mean := GroupedMeanDurationTimelineQuery(q)
	if mean && !groupedDurationSupportValid(r) {
		return nil
	}
	d := baseViewDataset(r, "grouped-time-series", "grouped-intervals")
	if d == nil {
		return nil
	}
	d.Dimensions = []string{"category"}
	d.Interval = &ViewInterval{FromMS: strconv.FormatInt(q.FromMS, 10), ToMS: strconv.FormatInt(q.ToMS, 10), BucketMS: strconv.FormatInt(q.BucketMS, 10)}
	d.Fields = []ViewField{{Key: "category", Label: q.GroupBy[0].Kind + ":" + q.GroupBy[0].Key, Type: "string", Nullable: true}, {Key: "fromMs", Label: "Interval start", Type: "timestamp", Unit: "ms"}, {Key: "toMs", Label: "Interval end", Type: "timestamp", Unit: "ms"}, {Key: "count", Label: "Captured records", Type: "number", Unit: "records", Nullable: true}}
	if mean {
		d.Fields = append(d.Fields,
			ViewField{Key: "duration.sum", Label: "Duration sum", Type: "rational", Unit: "ns", Nullable: true, Aggregation: &ViewAggregation{Family: "duration", Kind: "sum", Method: "exact"}},
			ViewField{Key: "duration.mean", Label: "Mean duration", Type: "rational", Unit: "ns", Nullable: true, Aggregation: &ViewAggregation{Family: "duration", Kind: "mean", Method: "exact-rational"}})
	}
	var duration *DurationSummary
	if mean {
		duration = r.Duration
	}
	coherent := exhaustive && (r.State == "observed" || r.State == "no_data") && groupedSeriesAgree(r) && aggregateSeriesAgree(r.Count, nil, duration, r.TimeSeries)
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
			var cells map[string]ViewCell
			if mean {
				cells = map[string]ViewCell{}
				for _, key := range []string{"duration.sum", "duration.mean"} {
					values[key] = nil
					cells[key] = ViewCell{State: "unknown"}
				}
				if state != "unknown" {
					support := &ViewSupport{Total: point.Count, Valid: point.Count, Missing: "0", Empty: "0", Invalid: "0"}
					cellState := "no_data"
					if point.Count != "0" {
						cellState = "derived"
						count, _ := new(big.Int).SetString(point.Count, 10)
						values["duration.sum"] = viewText(point.Duration.SumNS)
						values["duration.mean"] = viewRational(point.Duration.SumNS, count)
					}
					for _, key := range []string{"duration.sum", "duration.mean"} {
						cells[key] = ViewCell{State: cellState, Support: support}
					}
				}
			}
			d.Rows = append(d.Rows, ViewRow{Key: "group-interval:" + hex.EncodeToString(digest[:]), Values: values, State: state, Cells: cells})
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
