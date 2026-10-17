package metrics

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func hp(start, end, count string, buckets ...string) HistogramPoint {
	s := testStream()
	s.Name = "duration"
	s.Unit = "s"
	return HistogramPoint{Stream: s, Temporality: "cumulative", StartNS: start, EndNS: end, Bounds: []float64{.01, .1}, BucketCounts: buckets, Count: count}
}
func TestHistogramEpochsAndBuckets(t *testing.T) {
	a, b := hp("1", "10", "3", "1", "1", "1"), hp("1", "20", "7", "2", "3", "2")
	r, e := EvaluateHistograms(HistogramRequest{"10", "20", "10", []HistogramPoint{b, a, a}})
	if e != nil {
		t.Fatal(e)
	}
	d := r.Series[0].Intervals[0].Distribution
	if r.Series[0].State != "ok" || d.Count != "4" || !reflect.DeepEqual(d.BucketCounts, []string{"1", "2", "1"}) || d.P95.State != "unbounded-upper" || *d.P95.Lower != .1 {
		t.Fatalf("%+v", r)
	}
	for _, change := range []struct {
		mutate func(*HistogramPoint)
		reason string
	}{
		{func(p *HistogramPoint) { p.StartNS = "11" }, "histogram-epoch-change"},
		{func(p *HistogramPoint) { p.Bounds = []float64{.02, .1} }, "bucket-layout-change"},
		{func(p *HistogramPoint) { p.BucketCounts = []string{"0", "4", "3"} }, "histogram-decrease"},
		{func(p *HistogramPoint) { p.NoRecordedValue = true; p.BucketCounts = nil; p.Count = "" }, "no-recorded-value"},
	} {
		bad := b
		change.mutate(&bad)
		r, e := EvaluateHistograms(HistogramRequest{"10", "20", "10", []HistogramPoint{a, bad}})
		if e != nil || r.Series[0].State != "unknown" || r.Series[0].Intervals[0].Reason != change.reason || r.Series[0].Intervals[0].Distribution != nil {
			t.Fatalf("%+v %v", r, e)
		}
	}
	b.EndNS = "30"
	old, e := EvaluateHistograms(HistogramRequest{"10", "30", "10", []HistogramPoint{a, b}})
	if e != nil || old.Series[0].State != "unknown" {
		t.Fatal("bridged gap", e)
	}
	saved, _ := json.Marshal(old)
	late := hp("1", "20", "5", "2", "2", "1")
	repaired, e := EvaluateHistograms(HistogramRequest{"10", "30", "10", []HistogramPoint{a, b, late}})
	if e != nil || repaired.Series[0].State != "ok" || old.Basis == repaired.Basis {
		t.Fatal("late repair failed", e)
	}
	after, _ := json.Marshal(old)
	if string(saved) != string(after) {
		t.Fatal("changed frozen result")
	}
}
func TestHistogramValidationAndDelta(t *testing.T) {
	p := hp("10", "20", "3", "1", "1", "1")
	p.Temporality = "delta"
	for _, mutate := range []func(*HistogramPoint){func(p *HistogramPoint) { p.Bounds = []float64{.1, .01} }, func(p *HistogramPoint) { p.Bounds = []float64{math.NaN(), .1} }, func(p *HistogramPoint) { p.BucketCounts = []string{"1"} }, func(p *HistogramPoint) { p.Count = "2" }, func(p *HistogramPoint) { p.Stream.Unit = "bytes" }} {
		bad := p
		mutate(&bad)
		if _, e := EvaluateHistograms(HistogramRequest{"10", "20", "10", []HistogramPoint{bad}}); e == nil {
			t.Fatal("accepted invalid histogram")
		}
	}
	r, e := EvaluateHistograms(HistogramRequest{"11", "19", "10", []HistogramPoint{p}})
	if e != nil || r.Series[0].Intervals[0].Reason != "boundary-crossing" {
		t.Fatal("interpolated buckets")
	}
	q := p
	q.StartNS = "19"
	q.EndNS = "30"
	if _, e = EvaluateHistograms(HistogramRequest{"10", "30", "10", []HistogramPoint{p, q}}); e == nil {
		t.Fatal("accepted delta overlap")
	}
	max := hp("10", "20", "18446744073709551615", "0", "18446744073709551615", "0")
	max.Temporality = "delta"
	r, e = EvaluateHistograms(HistogramRequest{"10", "20", "10", []HistogramPoint{max}})
	if e != nil || *r.Series[0].Intervals[0].Distribution.P95.Upper != .1 {
		t.Fatal("lost uint64 quantile rank", e)
	}
}
func TestAlignedHistogramMerge(t *testing.T) {
	p := hp("10", "20", "3", "1", "1", "1")
	p.Temporality = "delta"
	r, e := EvaluateHistograms(HistogramRequest{"10", "20", "10", []HistogramPoint{p}})
	if e != nil {
		t.Fatal(e)
	}
	a := HistogramMember{p.Stream, r.Series[0].Intervals[0]}
	b := a
	b.Stream = cloneStream(a.Stream)
	b.Stream.Resource["service.instance.id"] = "two"
	merged, e := MergeAlignedHistograms([]HistogramMember{b, a})
	if e != nil || merged.State != "ok" || merged.Distribution.Count != "6" {
		t.Fatal(merged, e)
	}
	reversed, e := MergeAlignedHistograms([]HistogramMember{a, b})
	if e != nil || merged.Basis != reversed.Basis {
		t.Fatal("merge order changed basis")
	}
	b.Interval.ToNS = "21"
	bad, e := MergeAlignedHistograms([]HistogramMember{a, b})
	if e != nil || bad.State != "unknown" || bad.Distribution != nil || len(bad.Members) != 2 {
		t.Fatal("merged unaligned intervals", e)
	}
	b.Interval = a.Interval
	d := *b.Interval.Distribution
	d.Bounds = []float64{.02, .1}
	b.Interval.Distribution = &d
	bad, e = MergeAlignedHistograms([]HistogramMember{a, b})
	if e != nil || bad.Reason != "incompatible-layout" {
		t.Fatal("merged layouts", e)
	}
	if _, e = MergeAlignedHistograms([]HistogramMember{a, a}); e == nil {
		t.Fatal("duplicated member")
	}
}
