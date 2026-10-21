package metrics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

const MaxHistogramTimeBuckets = 200
const maxHistogramTimeOutputBuckets = 20000

// HistogramTimeBucket is a fixed half-open window. Unavailable windows never
// expose a partial distribution. A validated empty distribution is no_data.
type HistogramTimeBucket struct {
	FromNS       string        `json:"fromNS"`
	ToNS         string        `json:"toNS"`
	State        string        `json:"state"`
	Reasons      []string      `json:"reasons"`
	Distribution *Distribution `json:"distribution,omitempty"`
}
type HistogramTimeSeries struct {
	Stream      Stream                `json:"stream"`
	Temporality string                `json:"temporality"`
	Buckets     []HistogramTimeBucket `json:"buckets"`
}
type HistogramTimeResult struct {
	APIVersion  string                `json:"apiVersion"`
	Basis       string                `json:"basis"`
	SourceBasis string                `json:"sourceBasis"`
	FromNS      string                `json:"fromNS"`
	ToNS        string                `json:"toNS"`
	WidthNS     string                `json:"widthNS"`
	Series      []HistogramTimeSeries `json:"series"`
}

type timedHistogramInterval struct {
	interval HistogramInterval
	from, to int64
}

// BucketHistograms combines complete evaluated intervals within explicit time
// windows anchored at r.FromNS. It never splits an increment, merges streams,
// interpolates percentiles or repairs unknown source evidence. The final window
// is clipped to r.ToNS; output budgets reject the entire request, never truncate.
func BucketHistograms(r *HistogramResult, widthNS string) (*HistogramTimeResult, error) {
	if r == nil || r.APIVersion != "goo/histogram-result/v1alpha1" || r.Basis == "" || len(r.Basis) > 256 {
		return nil, fmt.Errorf("require a versioned evaluated histogram result and source basis")
	}
	from, e := nano(r.FromNS)
	to, f := nano(r.ToNS)
	width, g := nano(widthNS)
	if e != nil || f != nil || g != nil || to <= from || width == 0 {
		return nil, fmt.Errorf("require canonical nonempty histogram time window and positive widthNS")
	}
	count := (to-from-1)/width + 1
	if count > MaxHistogramTimeBuckets || int64(len(r.Series)) > maxHistogramTimeOutputBuckets/count {
		return nil, fmt.Errorf("histogram time bucket output budget exceeded")
	}
	out := &HistogramTimeResult{APIVersion: "goo/histogram-time-result/v1alpha1", SourceBasis: r.Basis, FromNS: r.FromNS, ToNS: r.ToNS, WidthNS: widthNS, Series: []HistogramTimeSeries{}}
	streams := make(map[string]bool)
	total := 0
	for _, s := range r.Series {
		stream := cloneStream(s.Stream)
		if stream.Resource["service.name"] == "" || stream.Resource["service.instance.id"] == "" || stream.Scope == "" || stream.Name == "" || (stream.Unit != "s" && stream.Unit != "ms") {
			return nil, fmt.Errorf("require scoped duration histogram identity and s/ms unit")
		}
		encoded, _ := json.Marshal(stream)
		key := string(encoded)
		if streams[key] {
			return nil, fmt.Errorf("duplicate histogram time stream")
		}
		streams[key] = true
		total += len(s.Intervals)
		if total > 10000 {
			return nil, fmt.Errorf("histogram time interval limit exceeded")
		}
		intervals := make([]timedHistogramInterval, 0, len(s.Intervals))
		for _, i := range s.Intervals {
			a, x := nano(i.FromNS)
			b, y := nano(i.ToNS)
			if x != nil || y != nil || a >= b || a < from || b > to {
				return nil, fmt.Errorf("invalid retained histogram interval bounds")
			}
			intervals = append(intervals, timedHistogramInterval{i, a, b})
		}
		sort.SliceStable(intervals, func(i, j int) bool {
			if intervals[i].from != intervals[j].from {
				return intervals[i].from < intervals[j].from
			}
			if intervals[i].to != intervals[j].to {
				return intervals[i].to < intervals[j].to
			}
			a, _ := json.Marshal(intervals[i].interval)
			b, _ := json.Marshal(intervals[j].interval)
			return string(a) < string(b)
		})
		series := HistogramTimeSeries{Stream: stream, Temporality: s.Temporality, Buckets: []HistogramTimeBucket{}}
		for start := from; start < to; {
			step := width
			if step > to-start {
				step = to - start
			}
			end := start + step
			subset := HistogramSeries{Stream: stream, Temporality: s.Temporality, State: "ok", Intervals: []HistogramInterval{}}
			crossing := false
			for _, i := range intervals {
				if i.to <= start {
					continue
				}
				if i.from >= end {
					break
				}
				if i.from < start || i.to > end {
					crossing = true
					continue
				}
				subset.Intervals = append(subset.Intervals, i.interval)
			}
			summary := histogramWindow(strconv.FormatInt(start, 10), strconv.FormatInt(end, 10), subset)
			if crossing {
				summary.reject("boundary-crossing")
			}
			if summary.State == "unavailable" {
				summary.Distribution = nil
			}
			series.Buckets = append(series.Buckets, HistogramTimeBucket{FromNS: summary.FromNS, ToNS: summary.ToNS, State: summary.State, Reasons: summary.Reasons, Distribution: summary.Distribution})
			start = end
		}
		out.Series = append(out.Series, series)
	}
	// Include the complete derived result and the frozen source basis. Width is a
	// policy input even when it happens to produce the same clipped single bucket.
	encoded, _ := json.Marshal(out)
	sum := sha256.Sum256(encoded)
	out.Basis = hex.EncodeToString(sum[:])
	return out, nil
}
