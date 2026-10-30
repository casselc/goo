package measurements

import (
	"encoding/json"
	"fmt"
	"math/big"
	"math/bits"
	"regexp"
	"strconv"
	"time"

	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/scopedview"
)

// HistogramViewSelector identifies one retained evaluated instrument interval.
// It does not identify contributing records or request another source read.
type HistogramViewSelector struct {
	Kind     string         `json:"kind"`
	Question Question       `json:"question"`
	Stream   metrics.Stream `json:"stream"`
	FromNS   string         `json:"fromNS"`
	ToNS     string         `json:"toNS"`
}

var histogramViewBasis = regexp.MustCompile(`^[a-f0-9]{64}$`)
var histogramViewClock = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$`)

// HistogramView projects exactly one already evaluated histogram interval. It
// performs no evaluation, merging, source access or eager per-interval expansion.
// A known interval remains usable when other intervals in its series are unknown.
func HistogramView(r *Result, seriesIndex, intervalIndex int) (*scopedview.Dataset, error) {
	refuse := func(reason string) (*scopedview.Dataset, error) {
		return nil, fmt.Errorf("native histogram view: %s", reason)
	}
	if r == nil || r.Question.Definition.Kind != "histogram" || r.Question.Validate() != nil || !histogramViewBasis.MatchString(r.Basis) {
		return refuse("require a checked retained measurement result")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.ReadAt); err != nil || !histogramViewClock.MatchString(r.ReadAt) {
		return refuse("require a recorded read clock")
	}
	h := r.Histograms
	from, to := r.Question.BoundsNS()
	if h == nil || h.APIVersion != "goo/histogram-result/v1alpha1" || !histogramViewBasis.MatchString(h.Basis) || h.FromNS != from || h.ToNS != to || h.MaxGapNS != strconv.FormatInt(r.Question.Definition.MaxGapMS*1000000, 10) || seriesIndex < 0 || seriesIndex >= len(h.Series) {
		return refuse("require an evaluated member of the measurement window")
	}
	s := h.Series[seriesIndex]
	if (s.Temporality != "delta" && s.Temporality != "cumulative") || (s.Stream.Unit != "s" && s.Stream.Unit != "ms") || intervalIndex < 0 || intervalIndex >= len(s.Intervals) {
		return refuse("require a supported retained stream interval")
	}
	bound := r.Question.Definition
	if bound.Source.ExactStream != nil && !sameExactStream(*bound.Source.ExactStream, s.Stream) {
		return refuse("stream differs from the exact measurement binding")
	}
	bound.Source.ExactStream = &s.Stream
	if bound.Validate() != nil {
		return refuse("stream differs from the measurement source binding")
	}
	in := s.Intervals[intervalIndex]
	a, e1 := strconv.ParseInt(in.FromNS, 10, 64)
	b, e2 := strconv.ParseInt(in.ToNS, 10, 64)
	lo, _ := strconv.ParseInt(from, 10, 64)
	hi, _ := strconv.ParseInt(to, 10, 64)
	if e1 != nil || e2 != nil || strconv.FormatInt(a, 10) != in.FromNS || strconv.FormatInt(b, 10) != in.ToNS || a < lo || b > hi || a >= b || b-a > r.Question.Definition.MaxGapMS*1000000 || in.State != "ok" || in.Reason != "" || in.Distribution == nil {
		return refuse("selected interval is unavailable or outside the measurement window")
	}
	dist := in.Distribution
	if dist.Unit != s.Stream.Unit || metrics.ValidateDurationBounds(dist.Bounds) != nil || len(dist.BucketCounts) != len(dist.Bounds)+1 {
		return refuse("invalid explicit bucket layout")
	}
	total, err := strconv.ParseUint(dist.Count, 10, 64)
	if err != nil || strconv.FormatUint(total, 10) != dist.Count {
		return refuse("invalid observation count")
	}
	var sum uint64
	for _, raw := range dist.BucketCounts {
		n, e := strconv.ParseUint(raw, 10, 64)
		if e != nil || strconv.FormatUint(n, 10) != raw {
			return refuse("invalid bucket count")
		}
		next, carry := bits.Add64(sum, n, 0)
		if carry != 0 {
			return refuse("bucket count overflow")
		}
		sum = next
	}
	if sum != total {
		return refuse("bucket counts do not conserve observations")
	}
	selector, err := json.Marshal(HistogramViewSelector{Kind: "measurement-interval", Question: r.Question, Stream: s.Stream, FromNS: in.FromNS, ToNS: in.ToNS})
	if err != nil {
		return refuse("invalid scope encoding")
	}
	scopeID := r.Basis + ":native:" + strconv.Itoa(seriesIndex) + ":" + strconv.Itoa(intervalIndex)
	d := &scopedview.Dataset{Version: "goo.dataset/1", ID: scopeID + ":bins", Revision: r.Basis, Scope: scopedview.Scope{ID: scopeID, Generation: r.Basis, Signal: "metrics", Selector: selector}, Shape: "histogram", Origin: "source-read", SourceBasis: h.Basis, CapturedAt: r.ReadAt, Fields: []scopedview.Field{{Key: "from", Label: "Bin start (exclusive)", Type: "histogram-bound", Unit: dist.Unit}, {Key: "to", Label: "Bin end (inclusive)", Type: "histogram-bound", Unit: dist.Unit}, {Key: "count", Label: "Instrument observations", Type: "number", Unit: "observations"}}, Rows: []scopedview.Row{}, Quality: scopedview.Quality{Extent: "complete", Basis: "captured-query", Value: "derived", Sampling: "unknown", Freshness: "fixed"}, Completeness: &scopedview.Completeness{Basis: "captured-query", Revision: r.Basis, ScopeID: scopeID, Aspects: []string{"bins"}}, Histogram: &scopedview.Histogram{Kind: "otel-explicit", Binning: "explicit", Grain: "observations", Temporality: s.Temporality, Unit: dist.Unit, Total: dist.Count, Closure: "left-open-right-closed"}}
	lower := "-inf"
	for i, count := range dist.BucketCounts {
		upper := "+inf"
		if i < len(dist.Bounds) {
			upper = new(big.Rat).SetFloat64(dist.Bounds[i]).RatString()
		}
		state := "observed"
		if count == "0" {
			state = "no_data"
		}
		x, y, n := lower, upper, count
		d.Rows = append(d.Rows, scopedview.Row{Key: "bin:" + strconv.Itoa(i), State: state, Values: scopedview.Values{"from": &x, "to": &y, "count": &n}})
		lower = upper
	}
	return d, nil
}
