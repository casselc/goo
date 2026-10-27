package metrics

import (
	"encoding/json"
	"math/big"
	"sort"
)

// WindowSummary preserves actual bounds and full, validated interval coverage.
// Counter rates are exact rational strings; histogram percentiles remain ranges.
type WindowSummary struct {
	FromNS       string        `json:"fromNS"`
	ToNS         string        `json:"toNS"`
	Temporality  string        `json:"temporality"`
	State        string        `json:"state"`
	Reasons      []string      `json:"reasons"`
	Delta        string        `json:"delta,omitempty"`
	Rate         string        `json:"rate,omitempty"`
	Distribution *Distribution `json:"distribution,omitempty"`
}
type BucketChange struct {
	Baseline     string `json:"baseline"`
	Selection    string `json:"selection"`
	DifferencePP string `json:"differencePP"`
}
type WindowComparison struct {
	Stream         Stream         `json:"stream"`
	State          string         `json:"state"`
	Reasons        []string       `json:"reasons"`
	Baseline       *WindowSummary `json:"baseline,omitempty"`
	Selection      *WindowSummary `json:"selection,omitempty"`
	RateDifference string         `json:"rateDifference,omitempty"`
	Buckets        []BucketChange `json:"buckets,omitempty"`
	P95Relation    string         `json:"p95Relation,omitempty"`
}

func window(from, to, temp string) *WindowSummary {
	return &WindowSummary{FromNS: from, ToNS: to, Temporality: temp, State: "observed", Reasons: []string{}}
}
func (w *WindowSummary) reject(reason string) {
	w.State = "unavailable"
	for _, old := range w.Reasons {
		if old == reason {
			return
		}
	}
	w.Reasons = append(w.Reasons, reason)
}
func boundsOK(w *WindowSummary) bool {
	a, e := nano(w.FromNS)
	b, f := nano(w.ToNS)
	if e != nil || f != nil || b <= a {
		w.reject("invalid-window")
		return false
	}
	return true
}
func coverage(w *WindowSummary, cursor, from, to string, state, reason string) string {
	a, e := nano(from)
	b, f := nano(to)
	lo, _ := nano(w.FromNS)
	hi, _ := nano(w.ToNS)
	if e != nil || f != nil || a >= b || a < lo || b > hi {
		w.reject("invalid-interval")
	}
	if from != cursor {
		w.reject("incomplete-or-overlapping-coverage")
	}
	if state != "ok" {
		if reason == "" {
			reason = "unknown-interval"
		}
		w.reject(reason)
	}
	return to
}
func counterWindow(from, to string, s CounterSeries) *WindowSummary {
	w := window(from, to, s.Temporality)
	if !boundsOK(w) {
		return w
	}
	cursor := from
	sum := new(big.Int)
	for _, i := range s.Intervals {
		cursor = coverage(w, cursor, i.FromNS, i.ToNS, i.State, i.Reason)
		if i.State == "ok" {
			n, e := uintValue(i.Delta)
			if e != nil {
				w.reject("invalid-delta")
			} else {
				sum.Add(sum, new(big.Int).SetUint64(n))
			}
		}
	}
	if cursor != to || len(s.Intervals) == 0 || s.State != "ok" {
		w.reject("incomplete-window")
	}
	if s.Temporality != "delta" && s.Temporality != "cumulative" {
		w.reject("unsupported-temporality")
	}
	if w.State == "observed" {
		a, _ := nano(from)
		b, _ := nano(to)
		w.Delta = sum.String()
		w.Rate = new(big.Rat).SetFrac(new(big.Int).Mul(sum, big.NewInt(1000000000)), big.NewInt(b-a)).RatString()
	}
	return w
}
func histogramWindow(from, to string, s HistogramSeries) *WindowSummary {
	w := window(from, to, s.Temporality)
	if !boundsOK(w) {
		return w
	}
	cursor := from
	var bounds []float64
	var buckets []uint64
	var count uint64
	have := false
	for _, i := range s.Intervals {
		cursor = coverage(w, cursor, i.FromNS, i.ToNS, i.State, i.Reason)
		d := i.Distribution
		if i.State != "ok" || d == nil {
			if d == nil {
				w.reject("missing-distribution")
			}
			continue
		}
		if d.Unit != s.Stream.Unit || validBounds(d.Bounds) != nil || len(d.BucketCounts) != len(d.Bounds)+1 {
			w.reject("invalid-distribution")
			continue
		}
		if !have {
			bounds = append([]float64{}, d.Bounds...)
			buckets = make([]uint64, len(d.BucketCounts))
			have = true
		} else if !sameBounds(bounds, d.Bounds) {
			w.reject("bucket-layout-change")
			continue
		}
		n, e := uintValue(d.Count)
		if e != nil || ^uint64(0)-count < n {
			w.reject("count-overflow")
			continue
		}
		count += n
		var subtotal uint64
		for k, text := range d.BucketCounts {
			n, e := uintValue(text)
			if e != nil || ^uint64(0)-buckets[k] < n || ^uint64(0)-subtotal < n {
				w.reject("bucket-count-overflow")
				continue
			}
			buckets[k] += n
			subtotal += n
		}
		expected, _ := uintValue(d.Count)
		if subtotal != expected {
			w.reject("bucket-count-mismatch")
		}
	}
	if cursor != to || !have || s.State != "ok" {
		w.reject("incomplete-window")
	}
	if s.Temporality != "delta" && s.Temporality != "cumulative" {
		w.reject("unsupported-temporality")
	}
	if w.State == "observed" {
		w.Distribution = distribution(s.Stream.Unit, bounds, buckets, count)
		if count == 0 {
			w.State = "no_data"
			w.Reasons = append(w.Reasons, "empty-distribution")
		}
	}
	return w
}

type windowPair struct {
	stream Stream
	a, b   *WindowSummary
}

func addWindow(m map[string]*windowPair, stream Stream, summary *WindowSummary, side int) {
	raw, _ := json.Marshal(stream)
	key := string(raw)
	p := m[key]
	if p == nil {
		p = &windowPair{stream: stream}
		m[key] = p
	}
	if side == 0 {
		if p.a != nil {
			summary.reject("duplicate-stream")
		}
		p.a = summary
	} else {
		if p.b != nil {
			summary.reject("duplicate-stream")
		}
		p.b = summary
	}
}
func compareWindows(m map[string]*windowPair, hist bool) []WindowComparison {
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := []WindowComparison{}
	for _, k := range keys {
		p := m[k]
		r := WindowComparison{Stream: p.stream, Baseline: p.a, Selection: p.b, State: "comparable", Reasons: []string{}}
		if p.a == nil || p.b == nil {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "missing-exact-stream")
		} else if p.a.State != "observed" || p.b.State != "observed" {
			r.State = "unavailable"
			r.Reasons = append(r.Reasons, "incomplete-or-empty-window")
		} else if p.a.Temporality != p.b.Temporality {
			r.State = "incompatible"
			r.Reasons = append(r.Reasons, "temporality-differs")
		} else if !hist {
			x, _ := new(big.Rat).SetString(p.a.Rate)
			y, _ := new(big.Rat).SetString(p.b.Rate)
			r.RateDifference = new(big.Rat).Sub(y, x).RatString()
		} else {
			a, b := p.a.Distribution, p.b.Distribution
			if !sameBounds(a.Bounds, b.Bounds) {
				r.State = "incompatible"
				r.Reasons = append(r.Reasons, "bucket-layout-differs")
			} else {
				for i := range a.BucketCounts {
					an, _ := new(big.Int).SetString(a.BucketCounts[i], 10)
					bn, _ := new(big.Int).SetString(b.BucketCounts[i], 10)
					ad, _ := new(big.Int).SetString(a.Count, 10)
					bd, _ := new(big.Int).SetString(b.Count, 10)
					change := new(big.Rat).Mul(new(big.Rat).Sub(new(big.Rat).SetFrac(bn, bd), new(big.Rat).SetFrac(an, ad)), big.NewRat(100, 1))
					r.Buckets = append(r.Buckets, BucketChange{Baseline: a.BucketCounts[i], Selection: b.BucketCounts[i], DifferencePP: change.RatString()})
				}
				r.P95Relation = "overlap-or-unbounded"
				if a.P95.Upper != nil && b.P95.Lower != nil && *b.P95.Lower > *a.P95.Upper {
					r.P95Relation = "selection-range-higher"
				}
				if b.P95.Upper != nil && a.P95.Lower != nil && *b.P95.Upper < *a.P95.Lower {
					r.P95Relation = "selection-range-lower"
				}
			}
		}
		out = append(out, r)
	}
	return out
}
func CompareCounterWindows(a, b *CounterResult) []WindowComparison {
	m := map[string]*windowPair{}
	for side, r := range []*CounterResult{a, b} {
		if r != nil {
			for _, s := range r.Series {
				summary := window(r.FromNS, r.ToNS, s.Temporality)
				if r.ValueDomain != "" {
					summary.reject("unsupported-decimal-counter-window-comparison")
				} else {
					summary = counterWindow(r.FromNS, r.ToNS, s)
				}
				addWindow(m, s.Stream, summary, side)
			}
		}
	}
	return compareWindows(m, false)
}
func CompareHistogramWindows(a, b *HistogramResult) []WindowComparison {
	m := map[string]*windowPair{}
	for side, r := range []*HistogramResult{a, b} {
		if r != nil {
			for _, s := range r.Series {
				addWindow(m, s.Stream, histogramWindow(r.FromNS, r.ToNS, s), side)
			}
		}
	}
	return compareWindows(m, true)
}
