package measurements

import (
	"fmt"
	"strconv"
	"time"

	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/scopedview"
)

// HistogramTimeViewSelector identifies one anchored, whole-export time bucket.
// Width and index are policy, not a claim about individual observation times.
type HistogramTimeViewSelector struct {
	Kind        string         `json:"kind"`
	Question    Question       `json:"question"`
	Stream      metrics.Stream `json:"stream"`
	FromNS      string         `json:"fromNS"`
	ToNS        string         `json:"toNS"`
	WidthNS     string         `json:"widthNS"`
	BucketIndex int            `json:"bucketIndex"`
}

// HistogramTimeView projects one retained evaluated fixed-time bucket. It does
// not split exports, recompute aggregates, merge streams or read a source.
func HistogramTimeView(r *Result, seriesIndex, bucketIndex int) (*scopedview.Dataset, error) {
	refuse := func(reason string) (*scopedview.Dataset, error) {
		return nil, fmt.Errorf("native histogram time view: %s", reason)
	}
	if r == nil || r.Question.Definition.Kind != "histogram" || r.Question.Validate() != nil || r.Question.Definition.HistogramBucketMS == nil || !histogramViewBasis.MatchString(r.Basis) {
		return refuse("require a checked retained bucketed measurement")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.ReadAt); err != nil || !histogramViewClock.MatchString(r.ReadAt) {
		return refuse("require a recorded read clock")
	}
	h, ht := r.Histograms, r.HistogramTime
	from, to := r.Question.BoundsNS()
	width := *r.Question.Definition.HistogramBucketMS * 1000000
	widthNS := strconv.FormatInt(width, 10)
	if h == nil || h.APIVersion != "goo/histogram-result/v1alpha1" || !histogramViewBasis.MatchString(h.Basis) || h.FromNS != from || h.ToNS != to || h.MaxGapNS != strconv.FormatInt(r.Question.Definition.MaxGapMS*1000000, 10) || ht == nil || ht.APIVersion != "goo/histogram-time-result/v1alpha1" || !histogramViewBasis.MatchString(ht.Basis) || ht.SourceBasis != h.Basis || ht.FromNS != from || ht.ToNS != to || ht.WidthNS != widthNS || seriesIndex < 0 || seriesIndex >= len(ht.Series) {
		return refuse("require the matching derived time result and source basis")
	}
	s := ht.Series[seriesIndex]
	if (s.Temporality != "delta" && s.Temporality != "cumulative") || (s.Stream.Unit != "s" && s.Stream.Unit != "ms") {
		return refuse("require a supported retained duration stream")
	}
	bound := r.Question.Definition
	if bound.Source.ExactStream != nil && !sameExactStream(*bound.Source.ExactStream, s.Stream) {
		return refuse("stream differs from exact measurement binding")
	}
	bound.Source.ExactStream = &s.Stream
	if bound.Validate() != nil {
		return refuse("stream differs from measurement binding")
	}
	members := 0
	for _, original := range h.Series {
		if sameExactStream(original.Stream, s.Stream) {
			members++
			if original.Temporality != s.Temporality {
				return refuse("temporality differs from original evaluated stream")
			}
		}
	}
	if members != 1 {
		return refuse("derived stream is not a unique original member")
	}
	members = 0
	for _, derived := range ht.Series {
		if sameExactStream(derived.Stream, s.Stream) {
			members++
		}
	}
	if members != 1 {
		return refuse("derived stream is not uniquely retained")
	}
	lo, _ := strconv.ParseInt(from, 10, 64)
	hi, _ := strconv.ParseInt(to, 10, 64)
	count := (hi-lo-1)/width + 1
	if count > metrics.MaxHistogramTimeBuckets || int64(len(s.Buckets)) != count || bucketIndex < 0 || bucketIndex >= len(s.Buckets) {
		return refuse("invalid anchored bucket membership")
	}
	bucket := s.Buckets[bucketIndex]
	start := lo + int64(bucketIndex)*width
	end := start + width
	if end > hi {
		end = hi
	}
	if bucket.FromNS != strconv.FormatInt(start, 10) || bucket.ToNS != strconv.FormatInt(end, 10) || bucket.Reasons == nil || bucket.Distribution == nil || bucket.Distribution.Unit != s.Stream.Unit {
		return refuse("selected bucket is unavailable or has invalid bounds")
	}
	// empty-distribution is the evaluator's informational known-zero reason.
	if bucket.State == "no_data" {
		if bucket.Distribution.Count != "0" || len(bucket.Reasons) != 1 || bucket.Reasons[0] != "empty-distribution" {
			return refuse("invalid known-empty bucket")
		}
	} else if bucket.State != "observed" || bucket.Distribution.Count == "0" || len(bucket.Reasons) != 0 {
		return refuse("selected bucket coverage is unavailable")
	}
	selector := HistogramTimeViewSelector{Kind: "measurement-time-bucket", Question: r.Question, Stream: s.Stream, FromNS: bucket.FromNS, ToNS: bucket.ToNS, WidthNS: widthNS, BucketIndex: bucketIndex}
	scopeID := r.Basis + ":native-time:" + strconv.Itoa(seriesIndex) + ":" + strconv.Itoa(bucketIndex)
	return histogramDistributionView(r, selector, scopeID, ht.Basis, s.Temporality, bucket.Distribution)
}
