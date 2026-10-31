package measurements

import (
	"maps"
	"strconv"
	"time"

	"github.com/casselc/goo/metrics"
)

// HistogramWindowStream is the retained stream reference needed to resolve one
// whole-window display. It intentionally carries no observation or bucket data.
type HistogramWindowStream struct {
	Stream      metrics.Stream `json:"stream"`
	Temporality string         `json:"temporality"`
	State       string         `json:"state"`
}

// HistogramWindowSource is compact source context for held window projections.
// It is metadata, not authentication or a source-completeness witness.
type HistogramWindowSource struct {
	MeasurementBasis string                  `json:"measurementBasis"`
	HistogramBasis   string                  `json:"histogramBasis"`
	Streams          []HistogramWindowStream `json:"streams"`
}

// HistogramSource returns bounded metadata once per retained measurement side.
// It never copies rows, intervals or distributions and never reads a source.
func HistogramSource(r *Result) *HistogramWindowSource {
	if r == nil || r.Question.Definition.Kind != "histogram" || r.Question.Validate() != nil || !histogramViewBasis.MatchString(r.Basis) {
		return nil
	}
	if _, err := time.Parse(time.RFC3339Nano, r.ReadAt); err != nil || !histogramViewClock.MatchString(r.ReadAt) {
		return nil
	}
	h := r.Histograms
	from, to := r.Question.BoundsNS()
	if h == nil || len(h.Series) > RowLimit || h.APIVersion != "goo/histogram-result/v1alpha1" || !histogramViewBasis.MatchString(h.Basis) || h.FromNS != from || h.ToNS != to || h.MaxGapNS != strconv.FormatInt(r.Question.Definition.MaxGapMS*1000000, 10) {
		return nil
	}
	out := &HistogramWindowSource{MeasurementBasis: r.Basis, HistogramBasis: h.Basis, Streams: make([]HistogramWindowStream, len(h.Series))}
	for i, s := range h.Series {
		stream := s.Stream
		stream.Resource = maps.Clone(stream.Resource)
		stream.Attributes = maps.Clone(stream.Attributes)
		out.Streams[i] = HistogramWindowStream{Stream: stream, Temporality: s.Temporality, State: s.State}
	}
	return out
}
