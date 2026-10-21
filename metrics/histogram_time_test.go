package metrics

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func timeHistogram(intervals ...HistogramInterval) *HistogramResult {
	stream := testStream()
	stream.Name = "duration"
	stream.Unit = "s"
	return &HistogramResult{APIVersion: "goo/histogram-result/v1alpha1", Basis: "source-basis", FromNS: "0", ToNS: "40", Series: []HistogramSeries{{Stream: stream, Temporality: "delta", State: "ok", Intervals: intervals}}}
}
func timeInterval(from, to, count string, populations ...string) HistogramInterval {
	return HistogramInterval{FromNS: from, ToNS: to, State: "ok", Distribution: &Distribution{Unit: "s", Bounds: []float64{1}, Count: count, BucketCounts: populations}}
}
func mustTime(t *testing.T, r *HistogramResult, width string) *HistogramTimeResult {
	t.Helper()
	out, e := BucketHistograms(r, width)
	if e != nil {
		t.Fatal(e)
	}
	return out
}
func hasTimeReason(b HistogramTimeBucket, reason string) bool {
	for _, r := range b.Reasons {
		if r == reason {
			return true
		}
	}
	return false
}
func TestHistogramTimeWholeIntervalsAndLocalCoverage(t *testing.T) {
	r := timeHistogram(timeInterval("0", "5", "3", "1", "2"), timeInterval("5", "10", "5", "3", "2"), timeInterval("10", "20", "0", "0", "0"), HistogramInterval{FromNS: "20", ToNS: "30", State: "unknown", Reason: "histogram-epoch-change"}, timeInterval("30", "40", "2", "0", "2"))
	r.Series[0].State = "unknown"
	out := mustTime(t, r, "10")
	b := out.Series[0].Buckets
	if len(b) != 4 || b[0].State != "observed" || b[0].Distribution.Count != "8" || !reflect.DeepEqual(b[0].Distribution.BucketCounts, []string{"4", "4"}) || b[0].Distribution.P95.State != "unbounded-upper" {
		t.Fatalf("bad merge: %+v", b[0])
	}
	if b[1].State != "no_data" || b[1].Distribution.Count != "0" || b[1].Distribution.P95.State != "no_data" {
		t.Fatalf("invented empty distribution percentile: %+v", b[1])
	}
	if b[2].State != "unavailable" || b[2].Distribution != nil || !hasTimeReason(b[2], "histogram-epoch-change") {
		t.Fatalf("unknown interval gained distribution: %+v", b[2])
	}
	if b[3].State != "observed" || b[3].Distribution.Count != "2" {
		t.Fatal("global unknown tainted independent local bucket")
	}
	if out.SourceBasis != r.Basis || out.WidthNS != "10" || out.Basis == "" {
		t.Fatal("missing frozen basis/policy")
	}
}
func TestHistogramTimeCrossingGapAndFinalClip(t *testing.T) {
	r := timeHistogram(timeInterval("0", "15", "4", "2", "2"), timeInterval("15", "20", "1", "1", "0"), timeInterval("30", "35", "1", "0", "1"))
	r.ToNS = "35"
	b := mustTime(t, r, "10").Series[0].Buckets
	for _, i := range []int{0, 1} {
		if b[i].State != "unavailable" || b[i].Distribution != nil || !hasTimeReason(b[i], "boundary-crossing") {
			t.Fatalf("split increment in bucket %d: %+v", i, b[i])
		}
	}
	if b[2].State != "unavailable" || b[2].Distribution != nil {
		t.Fatal("missing interval became zero")
	}
	if b[3].FromNS != "30" || b[3].ToNS != "35" || b[3].State != "observed" || b[3].Distribution.Count != "1" {
		t.Fatalf("bad clipped final bucket: %+v", b[3])
	}
}
func TestHistogramTimeRefusesPartialOrIncompatibleDistributions(t *testing.T) {
	cases := []struct {
		name, reason string
		change       func(*HistogramResult)
	}{
		{"layout", "bucket-layout-change", func(r *HistogramResult) { r.Series[0].Intervals[1].Distribution.Bounds = []float64{2} }},
		{"unit", "invalid-distribution", func(r *HistogramResult) { r.Series[0].Intervals[1].Distribution.Unit = "ms" }},
		{"missing", "missing-distribution", func(r *HistogramResult) { r.Series[0].Intervals[1].Distribution = nil }},
		{"counts", "bucket-count-mismatch", func(r *HistogramResult) { r.Series[0].Intervals[1].Distribution.BucketCounts = []string{"1", "0"} }},
		{"noncanonical", "bucket-count-overflow", func(r *HistogramResult) { r.Series[0].Intervals[1].Distribution.BucketCounts = []string{"01", "1"} }},
		{"overlap", "incomplete-or-overlapping-coverage", func(r *HistogramResult) { r.Series[0].Intervals[1].FromNS = "4" }},
		{"overflow", "count-overflow", func(r *HistogramResult) {
			r.Series[0].Intervals[0].Distribution.Count = "18446744073709551615"
			r.Series[0].Intervals[0].Distribution.BucketCounts = []string{"18446744073709551615", "0"}
		}},
		{"unsupported", "unsupported-temporality", func(r *HistogramResult) { r.Series[0].Temporality = "unspecified" }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := timeHistogram(timeInterval("0", "5", "2", "1", "1"), timeInterval("5", "10", "2", "1", "1"))
			r.ToNS = "10"
			c.change(r)
			b := mustTime(t, r, "10").Series[0].Buckets[0]
			if b.State != "unavailable" || b.Distribution != nil || !hasTimeReason(b, c.reason) {
				t.Fatalf("partial/incompatible aggregate: %+v", b)
			}
		})
	}
}
func TestHistogramTimeExactImmutableDeterministicAndIdentity(t *testing.T) {
	r := timeHistogram(timeInterval("0", "10", "9007199254740993", "9007199254740993", "0"), timeInterval("10", "20", "1", "0", "1"))
	r.ToNS = "20"
	before, _ := json.Marshal(r)
	a := mustTime(t, r, "20")
	after, _ := json.Marshal(r)
	if string(before) != string(after) || a.Series[0].Buckets[0].Distribution.Count != "9007199254740994" {
		t.Fatal("input mutated or counts rounded")
	}
	r.Series[0].Intervals[0], r.Series[0].Intervals[1] = r.Series[0].Intervals[1], r.Series[0].Intervals[0]
	b := mustTime(t, r, "20")
	if !reflect.DeepEqual(a, b) {
		t.Fatal("interval arrival order changed result")
	}
	a.Series[0].Stream.Resource["service.name"] = "mutated"
	a.Series[0].Buckets[0].Distribution.Bounds[0] = 100
	if r.Series[0].Stream.Resource["service.name"] == "mutated" || r.Series[0].Intervals[0].Distribution.Bounds[0] != 1 {
		t.Fatal("derived output aliases input")
	}
	c := mustTime(t, r, "30")
	if b.Basis == c.Basis {
		t.Fatal("different widths lost policy identity")
	}
	other := r.Series[0]
	other.Stream = cloneStream(other.Stream)
	other.Stream.Resource["service.instance.id"] = "second"
	r.Series = append(r.Series, other)
	d := mustTime(t, r, "20")
	if len(d.Series) != 2 || d.Series[1].Stream.Resource["service.instance.id"] != "second" || d.Series[0].Buckets[0].Distribution.Count != d.Series[1].Buckets[0].Distribution.Count {
		t.Fatal("streams merged or reordered")
	}
}
func TestHistogramTimeBoundsAndBudgets(t *testing.T) {
	for _, width := range []string{"0", "-1", "01", "1.5", "9223372036854775808"} {
		if _, e := BucketHistograms(timeHistogram(), width); e == nil {
			t.Fatal("accepted invalid width", width)
		}
	}
	if _, e := BucketHistograms(nil, "1"); e == nil {
		t.Fatal("accepted nil result")
	}
	r := timeHistogram()
	r.ToNS = "201"
	if _, e := BucketHistograms(r, "1"); e == nil {
		t.Fatal("truncated bucket budget")
	}
	r = timeHistogram()
	r.Series[0].Intervals = make([]HistogramInterval, 10001)
	if _, e := BucketHistograms(r, "10"); e == nil {
		t.Fatal("accepted interval budget")
	}
	r = timeHistogram(timeInterval("01", "10", "1", "1", "0"))
	if _, e := BucketHistograms(r, "10"); e == nil {
		t.Fatal("accepted noncanonical bounds")
	}
	r = timeHistogram()
	r.Series = append(r.Series, r.Series[0])
	if _, e := BucketHistograms(r, "10"); e == nil {
		t.Fatal("ambiguous duplicate stream")
	}
	r = timeHistogram()
	r.Series[0].Stream.Unit = "bytes"
	if _, e := BucketHistograms(r, "10"); e == nil {
		t.Fatal("accepted unsupported identity")
	}
	r = timeHistogram()
	r.Series = nil
	empty := mustTime(t, r, "10")
	if empty.Series == nil {
		t.Fatal("nil empty output series")
	}
	r = timeHistogram()
	r.ToNS = "9223372036854775807"
	out := mustTime(t, r, "9223372036854775807")
	if out.Series[0].Buckets[0].ToNS != "9223372036854775807" {
		t.Fatal("overflowed bucket endpoint")
	}
	r = timeHistogram(timeInterval("0", "10", "1", "1", "0"))
	r.ToNS = "10"
	r.Series[0].Intervals[0].Distribution.Bounds = []float64{math.NaN()}
	b := mustTime(t, r, "10").Series[0].Buckets[0]
	if b.State != "unavailable" || b.Distribution != nil {
		t.Fatal("invalid boundary accepted")
	}
	r = timeHistogram()
	r.ToNS = "200"
	r.Series = make([]HistogramSeries, 101)
	if _, e := BucketHistograms(r, "1"); e == nil || !strings.Contains(e.Error(), "budget") {
		t.Fatal("output budget not bounded")
	}
}

func TestHistogramTimeOriginAndHalfOpenAdjacency(t *testing.T) {
	r := timeHistogram(timeInterval("7", "12", "1", "1", "0"), timeInterval("12", "17", "2", "0", "2"), timeInterval("17", "27", "1", "1", "0"), timeInterval("27", "32", "1", "0", "1"))
	r.FromNS = "7"
	r.ToNS = "32"
	b := mustTime(t, r, "10").Series[0].Buckets
	want := [][3]string{{"7", "17", "3"}, {"17", "27", "1"}, {"27", "32", "1"}}
	for i, w := range want {
		if b[i].FromNS != w[0] || b[i].ToNS != w[1] || b[i].State != "observed" || b[i].Distribution.Count != w[2] {
			t.Fatalf("origin/adjacency bucket %d: %+v", i, b[i])
		}
	}
}

func TestHistogramTimeUsesDifferencedCumulativeIntervals(t *testing.T) {
	points := []HistogramPoint{hp("0", "1", "3", "1", "1", "1"), hp("0", "2", "7", "2", "3", "2"), hp("0", "3", "9", "2", "4", "3")}
	native, e := EvaluateHistograms(HistogramRequest{FromNS: "1", ToNS: "3", MaxGapNS: "1", Points: points})
	if e != nil {
		t.Fatal(e)
	}
	result, e := BucketHistograms(native, "2")
	if e != nil {
		t.Fatal(e)
	}
	if d := result.Series[0].Buckets[0].Distribution; d == nil || d.Count != "6" || !reflect.DeepEqual(d.BucketCounts, []string{"1", "3", "2"}) {
		t.Fatal("cumulative observations double-counted", d)
	}
	points[2].StartNS = "2"
	points[2].Count = "1"
	points[2].BucketCounts = []string{"0", "0", "1"}
	native, e = EvaluateHistograms(HistogramRequest{FromNS: "1", ToNS: "3", MaxGapNS: "1", Points: points})
	if e != nil {
		t.Fatal(e)
	}
	result, e = BucketHistograms(native, "1")
	if e != nil {
		t.Fatal(e)
	}
	if result.Series[0].Buckets[0].State != "observed" || result.Series[0].Buckets[1].State != "unavailable" || result.Series[0].Buckets[1].Distribution != nil {
		t.Fatal("reset repaired or tainted a neighboring bucket", result)
	}
}
