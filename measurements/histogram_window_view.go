package measurements

import (
	"fmt"
	"strconv"
	"time"

	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/scopedview"
)

// HistogramWindowViewSelector describes one complete captured measurement window.
// It is neither an export interval nor a claim about contributing records.
type HistogramWindowViewSelector struct {
	Kind     string         `json:"kind"`
	Question Question       `json:"question"`
	Stream   metrics.Stream `json:"stream"`
	FromNS   string         `json:"fromNS"`
	ToNS     string         `json:"toNS"`
}

// HistogramWindowView projects an already evaluated whole-window summary. It
// does not aggregate intervals or read a source. Output is always held data;
// complete coverage in an evaluator summary does not grant source authority.
func HistogramWindowView(r *Result, stream metrics.Stream, summary *metrics.WindowSummary) (*scopedview.Dataset, error) {
	refuse := func(reason string) (*scopedview.Dataset, error) {
		return nil, fmt.Errorf("native histogram window view: %s", reason)
	}
	if r == nil || r.Question.Definition.Kind != "histogram" || r.Question.Validate() != nil || !histogramViewBasis.MatchString(r.Basis) {
		return refuse("require a checked retained histogram measurement")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.ReadAt); err != nil || !histogramViewClock.MatchString(r.ReadAt) {
		return refuse("require a recorded read clock")
	}
	h := r.Histograms
	from, to := r.Question.BoundsNS()
	if h == nil || h.APIVersion != "goo/histogram-result/v1alpha1" || !histogramViewBasis.MatchString(h.Basis) || h.FromNS != from || h.ToNS != to || h.MaxGapNS != strconv.FormatInt(r.Question.Definition.MaxGapMS*1000000, 10) {
		return refuse("require the matching evaluated histogram result")
	}
	member, index := 0, -1
	var original *metrics.HistogramSeries
	for i := range h.Series {
		if sameExactStream(h.Series[i].Stream, stream) {
			member++
			index = i
			original = &h.Series[i]
		}
	}
	if member != 1 || original.State != "ok" || (original.Temporality != "delta" && original.Temporality != "cumulative") {
		return refuse("require a unique complete original stream")
	}
	bound := r.Question.Definition
	if bound.Source.ExactStream != nil && !sameExactStream(*bound.Source.ExactStream, stream) {
		return refuse("stream differs from exact measurement binding")
	}
	bound.Source.ExactStream = &stream
	if bound.Validate() != nil {
		return refuse("stream differs from measurement binding")
	}
	if summary == nil || summary.FromNS != from || summary.ToNS != to || summary.Temporality != original.Temporality || summary.Distribution == nil || summary.Reasons == nil || summary.Distribution.Unit != stream.Unit {
		return refuse("require the matching evaluated full-window summary")
	}
	if summary.State == "no_data" {
		if summary.Distribution.Count != "0" || len(summary.Reasons) != 1 || summary.Reasons[0] != "empty-distribution" {
			return refuse("invalid known-empty window")
		}
	} else if summary.State != "observed" || summary.Distribution.Count == "0" || len(summary.Reasons) != 0 {
		return refuse("window coverage is unavailable")
	}
	selector := HistogramWindowViewSelector{Kind: "measurement-window", Question: r.Question, Stream: stream, FromNS: from, ToNS: to}
	scopeID := r.Basis + ":native-window:" + strconv.Itoa(index)
	d, err := histogramDistributionView(r, selector, scopeID, h.Basis, original.Temporality, summary.Distribution)
	if err != nil {
		return nil, err
	}
	d.Origin = "browser-held"
	d.Quality.Extent = "unknown"
	d.Completeness = nil
	return d, nil
}
