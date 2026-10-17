package metrics

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
)

type HistogramPoint struct {
	Stream          Stream    `json:"stream"`
	Temporality     string    `json:"temporality"`
	StartNS         string    `json:"startNS"`
	EndNS           string    `json:"endNS"`
	Bounds          []float64 `json:"bounds"`
	BucketCounts    []string  `json:"bucketCounts"`
	Count           string    `json:"count"`
	NoRecordedValue bool      `json:"noRecordedValue,omitempty"`
}
type HistogramRequest struct {
	FromNS   string           `json:"fromNS"`
	ToNS     string           `json:"toNS"`
	MaxGapNS string           `json:"maxGapNS"`
	Points   []HistogramPoint `json:"points"`
}
type QuantileRange struct {
	State string   `json:"state"`
	Lower *float64 `json:"lower,omitempty"`
	Upper *float64 `json:"upper,omitempty"`
}
type Distribution struct {
	Unit         string        `json:"unit"`
	Bounds       []float64     `json:"bounds"`
	BucketCounts []string      `json:"bucketCounts"`
	Count        string        `json:"count"`
	P95          QuantileRange `json:"p95"`
}
type HistogramInterval struct {
	FromNS       string        `json:"fromNS"`
	ToNS         string        `json:"toNS"`
	State        string        `json:"state"`
	Reason       string        `json:"reason,omitempty"`
	Distribution *Distribution `json:"distribution,omitempty"`
}
type HistogramSeries struct {
	Stream      Stream              `json:"stream"`
	Temporality string              `json:"temporality"`
	State       string              `json:"state"`
	Intervals   []HistogramInterval `json:"intervals"`
}
type HistogramResult struct {
	APIVersion string            `json:"apiVersion"`
	Basis      string            `json:"basis"`
	FromNS     string            `json:"fromNS"`
	ToNS       string            `json:"toNS"`
	MaxGapNS   string            `json:"maxGapNS"`
	Series     []HistogramSeries `json:"series"`
	Warnings   []string          `json:"warnings"`
}
type histogramPoint struct {
	raw        HistogramPoint
	start, end int64
	count      uint64
	buckets    []uint64
}

func ValidateDurationBounds(bounds []float64) error { return validBounds(bounds) }
func validBounds(bounds []float64) error {
	if len(bounds) > 256 {
		return fmt.Errorf("histogram has more than 256 explicit bounds")
	}
	for i, b := range bounds {
		if math.IsNaN(b) || math.IsInf(b, 0) || b < 0 || (i > 0 && b <= bounds[i-1]) {
			return fmt.Errorf("duration bounds must be finite, nonnegative and strictly increasing")
		}
	}
	return nil
}
func uintValue(s string) (uint64, error) {
	n, e := strconv.ParseUint(s, 10, 64)
	if e != nil || strconv.FormatUint(n, 10) != s {
		return 0, fmt.Errorf("require canonical uint64 histogram counts")
	}
	return n, nil
}
func histogramCounts(bounds []float64, raw []string, total string) ([]uint64, uint64, error) {
	if e := validBounds(bounds); e != nil {
		return nil, 0, e
	}
	if len(raw) != len(bounds)+1 {
		return nil, 0, fmt.Errorf("histogram bucket shape mismatch")
	}
	count, e := uintValue(total)
	if e != nil {
		return nil, 0, e
	}
	counts := make([]uint64, len(raw))
	sum := uint64(0)
	for i, s := range raw {
		n, e := uintValue(s)
		if e != nil {
			return nil, 0, e
		}
		if math.MaxUint64-sum < n {
			return nil, 0, fmt.Errorf("histogram count overflow")
		}
		sum += n
		counts[i] = n
	}
	if sum != count {
		return nil, 0, fmt.Errorf("histogram bucket counts do not equal count")
	}
	return counts, count, nil
}
func sameBounds(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func distribution(unit string, bounds []float64, buckets []uint64, count uint64) *Distribution {
	d := &Distribution{Unit: unit, Bounds: append([]float64{}, bounds...), BucketCounts: []string{}, Count: strconv.FormatUint(count, 10), P95: QuantileRange{State: "no_data"}}
	for _, n := range buckets {
		d.BucketCounts = append(d.BucketCounts, strconv.FormatUint(n, 10))
	}
	if count == 0 {
		return d
	}
	// Exact ceil(count*95/100), without overflowing or rounding large counts.
	rank := count/100*95 + (count%100*95+99)/100
	sum := uint64(0)
	for i, n := range buckets {
		sum += n
		if sum >= rank {
			d.P95 = QuantileRange{State: "bucket-bound"}
			if i > 0 {
				lower := bounds[i-1]
				d.P95.Lower = &lower
			} else {
				d.P95.State = "unbounded-lower"
			}
			if i < len(bounds) {
				upper := bounds[i]
				d.P95.Upper = &upper
			} else if i == 0 {
				d.P95.State = "unbounded"
			} else {
				d.P95.State = "unbounded-upper"
			}
			break
		}
	}
	return d
}
func EvaluateHistograms(request HistogramRequest) (*HistogramResult, error) {
	from, e := nano(request.FromNS)
	if e != nil {
		return nil, e
	}
	to, e := nano(request.ToNS)
	if e != nil || to <= from {
		return nil, fmt.Errorf("require nonempty histogram window")
	}
	gap, e := nano(request.MaxGapNS)
	if e != nil || gap == 0 {
		return nil, fmt.Errorf("require positive histogram maxGapNS")
	}
	if len(request.Points) > 10000 {
		return nil, fmt.Errorf("histogram point limit exceeded")
	}
	groups := map[string][]histogramPoint{}
	for _, p := range request.Points {
		s := p.Stream
		if s.Resource["service.name"] == "" || s.Resource["service.instance.id"] == "" || s.Scope == "" || s.Version == "" || s.SchemaURL == "" || s.Name == "" || (s.Unit != "s" && s.Unit != "ms") {
			return nil, fmt.Errorf("require scoped duration histogram identity and s/ms unit")
		}
		if p.Temporality != "cumulative" && p.Temporality != "delta" {
			return nil, fmt.Errorf("unsupported histogram temporality")
		}
		start, e := nano(p.StartNS)
		if e != nil {
			return nil, e
		}
		end, e := nano(p.EndNS)
		if e != nil || end <= start {
			return nil, fmt.Errorf("invalid histogram interval")
		}
		if e = validBounds(p.Bounds); e != nil {
			return nil, e
		}
		var buckets []uint64
		var count uint64
		if !p.NoRecordedValue {
			buckets, count, e = histogramCounts(p.Bounds, p.BucketCounts, p.Count)
			if e != nil {
				return nil, e
			}
		}
		p.Stream = cloneStream(s)
		p.Bounds = append([]float64{}, p.Bounds...)
		p.BucketCounts = append([]string{}, p.BucketCounts...)
		key, _ := json.Marshal(p.Stream)
		groups[string(key)] = append(groups[string(key)], histogramPoint{p, start, end, count, buckets})
	}
	result := &HistogramResult{APIVersion: "goo/histogram-result/v1alpha1", FromNS: request.FromNS, ToNS: request.ToNS, MaxGapNS: request.MaxGapNS, Series: []HistogramSeries{}, Warnings: []string{"Explicit bucket populations only; no interpolation, scalar percentile averaging or completeness inferred.", "P95 is an observed bucket range, not an exact or interpolated latency percentile."}}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	canonical := []HistogramPoint{}
	for _, key := range keys {
		points := groups[key]
		sort.Slice(points, func(i, j int) bool { return points[i].end < points[j].end })
		unique := []histogramPoint{}
		for _, p := range points {
			if len(unique) > 0 {
				prior := unique[len(unique)-1]
				if prior.raw.Temporality != p.raw.Temporality {
					return nil, fmt.Errorf("mixed histogram temporality")
				}
				if prior.end == p.end {
					a, _ := json.Marshal(prior.raw)
					b, _ := json.Marshal(p.raw)
					if string(a) != string(b) {
						return nil, fmt.Errorf("conflicting histogram timestamp")
					}
					continue
				}
				if p.raw.Temporality == "cumulative" && p.start < prior.start {
					return nil, fmt.Errorf("histogram epoch start regressed")
				}
			}
			unique = append(unique, p)
			canonical = append(canonical, p.raw)
		}
		series := HistogramSeries{Stream: unique[0].raw.Stream, Temporality: unique[0].raw.Temporality, State: "no_data", Intervals: []HistogramInterval{}}
		tainted := false
		appendInterval := func(start, end int64, bounds []float64, counts []uint64, count uint64, reason string) {
			if end <= from || start >= to {
				return
			}
			a, b := start, end
			if a < from {
				a = from
			}
			if b > to {
				b = to
			}
			if a != start || b != end {
				reason = "boundary-crossing"
			}
			interval := HistogramInterval{FromNS: strconv.FormatInt(a, 10), ToNS: strconv.FormatInt(b, 10), State: "unknown", Reason: reason}
			if reason == "" {
				interval.State = "ok"
				interval.Distribution = distribution(series.Stream.Unit, bounds, counts, count)
			}
			series.Intervals = append(series.Intervals, interval)
		}
		for i, p := range unique {
			if p.raw.Temporality == "delta" {
				if i > 0 && p.start < unique[i-1].end {
					return nil, fmt.Errorf("overlapping delta histograms")
				}
				reason := ""
				if p.end-p.start > gap {
					reason = "excessive-gap"
				}
				if p.raw.NoRecordedValue {
					reason = "no-recorded-value"
				}
				appendInterval(p.start, p.end, p.raw.Bounds, p.buckets, p.count, reason)
				continue
			}
			if i == 0 {
				continue
			}
			prior := unique[i-1]
			reason := ""
			counts := []uint64{}
			count := uint64(0)
			switch {
			case p.start != prior.start:
				reason = "histogram-epoch-change"
				tainted = false
			case p.raw.NoRecordedValue || prior.raw.NoRecordedValue:
				reason = "no-recorded-value"
			case !sameBounds(p.raw.Bounds, prior.raw.Bounds):
				reason = "bucket-layout-change"
				tainted = true
			case tainted:
				reason = "invalid-histogram-epoch"
			default:
				if p.count < prior.count {
					reason = "histogram-decrease"
					tainted = true
				} else {
					count = p.count - prior.count
				}
				for j, n := range p.buckets {
					if n < prior.buckets[j] {
						reason = "histogram-decrease"
						tainted = true
						break
					}
					counts = append(counts, n-prior.buckets[j])
				}
				if reason == "" && p.end-prior.end > gap {
					reason = "excessive-gap"
				}
			}
			appendInterval(prior.end, p.end, p.raw.Bounds, counts, count, reason)
		}
		if len(series.Intervals) > 0 {
			series.State = "ok"
			cursor := from
			for _, interval := range series.Intervals {
				a, _ := nano(interval.FromNS)
				b, _ := nano(interval.ToNS)
				if a != cursor || interval.State != "ok" {
					series.State = "unknown"
				}
				cursor = b
			}
			if cursor != to {
				series.State = "unknown"
			}
		}
		result.Series = append(result.Series, series)
	}
	payload, _ := json.Marshal(HistogramRequest{request.FromNS, request.ToNS, request.MaxGapNS, canonical})
	sum := sha256.Sum256(payload)
	result.Basis = hex.EncodeToString(sum[:])
	return result, nil
}

// MergeAlignedHistograms combines explicitly selected members only. Other than
// instance ID, identities and exact time/bucket boundaries must agree. It never
// discovers missing fleet members or rebins incompatible histograms.
type HistogramMember struct {
	Stream   Stream            `json:"stream"`
	Interval HistogramInterval `json:"interval"`
}
type HistogramMerge struct {
	State        string        `json:"state"`
	Reason       string        `json:"reason,omitempty"`
	Basis        string        `json:"basis"`
	FromNS       string        `json:"fromNS"`
	ToNS         string        `json:"toNS"`
	Members      []Stream      `json:"members"`
	Distribution *Distribution `json:"distribution,omitempty"`
}

func MergeAlignedHistograms(input []HistogramMember) (*HistogramMerge, error) {
	if len(input) == 0 || len(input) > 1000 {
		return nil, fmt.Errorf("require 1..1000 explicit histogram members")
	}
	members := append([]HistogramMember{}, input...)
	sort.Slice(members, func(i, j int) bool {
		a, _ := json.Marshal(members[i].Stream)
		b, _ := json.Marshal(members[j].Stream)
		return string(a) < string(b)
	})
	out := &HistogramMerge{State: "unknown", FromNS: members[0].Interval.FromNS, ToNS: members[0].Interval.ToNS, Members: []Stream{}}
	payload, err := json.Marshal(members)
	if err != nil {
		return nil, fmt.Errorf("invalid histogram merge payload: %w", err)
	}
	sum := sha256.Sum256(payload)
	out.Basis = hex.EncodeToString(sum[:])
	for _, m := range members {
		out.Members = append(out.Members, cloneStream(m.Stream))
	}
	var base string
	seen := map[string]bool{}
	var bounds []float64
	counts := []uint64{}
	total := uint64(0)
	for i, m := range members {
		s := cloneStream(m.Stream)
		if s.Resource["service.name"] == "" || s.Resource["service.instance.id"] == "" || s.Scope == "" || s.Version == "" || s.SchemaURL == "" || s.Name == "" || (s.Unit != "s" && s.Unit != "ms") {
			return nil, fmt.Errorf("invalid histogram member identity")
		}

		key, _ := json.Marshal(s)
		if seen[string(key)] {
			return nil, fmt.Errorf("duplicate histogram member")
		}
		seen[string(key)] = true
		group := cloneStream(s)
		delete(group.Resource, "service.instance.id")
		g, _ := json.Marshal(group)
		if i == 0 {
			base = string(g)
		} else if string(g) != base {
			out.Reason = "incompatible-stream"
			return out, nil
		}
		if m.Interval.FromNS != out.FromNS || m.Interval.ToNS != out.ToNS {
			out.Reason = "unaligned-interval"
			return out, nil
		}
		a, e := nano(m.Interval.FromNS)
		if e != nil {
			return nil, e
		}
		b, e := nano(m.Interval.ToNS)
		if e != nil || b <= a {
			return nil, fmt.Errorf("invalid merge interval")
		}
		if m.Interval.State != "ok" || m.Interval.Distribution == nil {
			out.Reason = "member-unknown"
			return out, nil
		}
		d := m.Interval.Distribution
		if d.Unit != s.Unit {
			out.Reason = "incompatible-unit"
			return out, nil
		}
		parsed, count, e := histogramCounts(d.Bounds, d.BucketCounts, d.Count)
		if e != nil {
			return nil, e
		}
		if i == 0 {
			bounds = append([]float64{}, d.Bounds...)
			counts = make([]uint64, len(parsed))
		} else if !sameBounds(bounds, d.Bounds) {
			out.Reason = "incompatible-layout"
			return out, nil
		}
		if math.MaxUint64-total < count {
			return nil, fmt.Errorf("merged count overflow")
		}
		total += count
		for j, n := range parsed {
			if math.MaxUint64-counts[j] < n {
				return nil, fmt.Errorf("merged bucket overflow")
			}
			counts[j] += n
		}
	}
	out.State = "ok"
	out.Distribution = distribution(members[0].Stream.Unit, bounds, counts, total)
	return out, nil
}
