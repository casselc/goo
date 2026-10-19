package metrics

import (
	"encoding/json"
	"reflect"
	"testing"
)

func windowStream() Stream {
	return Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "one"}, Scope: "otel.http", Name: "requests", Unit: "{request}", Attributes: map[string]string{}}
}
func counterWindowFixture(t *testing.T, from, to, delta string) *CounterResult {
	t.Helper()
	r, e := EvaluateCounters(CounterRequest{FromNS: from, ToNS: to, MaxGapNS: "10000000000", Points: []CounterPoint{{Stream: windowStream(), Temporality: "delta", Monotonic: true, StartNS: from, EndNS: to, Value: delta}}})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestCounterWindowRatesCoverageAndIdentity(t *testing.T) {
	a, b := counterWindowFixture(t, "1000000000", "3000000000", "6"), counterWindowFixture(t, "4000000000", "8000000000", "20")
	before, _ := json.Marshal([]*CounterResult{a, b})
	r := CompareCounterWindows(a, b)
	after, _ := json.Marshal([]*CounterResult{a, b})
	if len(r) != 1 || r[0].RateDifference != "2" || r[0].Baseline.Rate != "3" || r[0].Selection.Rate != "5" || string(before) != string(after) {
		t.Fatal(r)
	}
	b.Series[0].Intervals[0].FromNS = "5000000000"
	r = CompareCounterWindows(a, b)
	if r[0].State != "unavailable" || r[0].RateDifference != "" {
		t.Fatal("gap accepted", r)
	}
	b = counterWindowFixture(t, "4000000000", "8000000000", "20")
	b.Series[0].Stream.Resource["service.instance.id"] = "two"
	r = CompareCounterWindows(a, b)
	if len(r) != 2 || r[0].RateDifference != "" || r[1].RateDifference != "" {
		t.Fatal("instances merged", r)
	}
	b = counterWindowFixture(t, "4000000000", "8000000000", "20")
	b.Series[0].Temporality = "cumulative"
	r = CompareCounterWindows(a, b)
	if r[0].State != "incompatible" {
		t.Fatal("temporality mixed", r)
	}
}
func TestCounterResetAndBoundaryIntervalsRefuse(t *testing.T) {
	for _, reason := range []string{"counter-epoch-change", "counter-decrease", "invalid-counter-epoch", "boundary-crossing", "excessive-gap", "no-recorded-value"} {
		a, b := counterWindowFixture(t, "1000000000", "3000000000", "6"), counterWindowFixture(t, "1000000000", "3000000000", "10")
		b.Series[0].Intervals[0].State = "unknown"
		b.Series[0].Intervals[0].Reason = reason
		b.Series[0].State = "unknown"
		r := CompareCounterWindows(a, b)
		if r[0].RateDifference != "" || !reflect.DeepEqual(r[0].Selection.Reasons, []string{reason, "incomplete-window"}) {
			t.Fatal(r)
		}
	}
}
func histogramWindowFixture(t *testing.T, bounds []float64, buckets []string) *HistogramResult {
	t.Helper()
	s := windowStream()
	s.Unit = "s"
	s.Name = "duration"
	r, e := EvaluateHistograms(HistogramRequest{FromNS: "1000000000", ToNS: "3000000000", MaxGapNS: "10000000000", Points: []HistogramPoint{{Stream: s, Temporality: "delta", StartNS: "1000000000", EndNS: "3000000000", Bounds: bounds, BucketCounts: buckets, Count: "10"}}})
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestHistogramWholePopulationsAndRangeComparison(t *testing.T) {
	a := histogramWindowFixture(t, []float64{1, 2, 3}, []string{"10", "0", "0", "0"})
	b := histogramWindowFixture(t, []float64{1, 2, 3}, []string{"0", "0", "10", "0"})
	r := CompareHistogramWindows(a, b)
	if len(r) != 1 || r[0].P95Relation != "selection-range-higher" || r[0].Buckets[0].DifferencePP != "-100" || r[0].Buckets[2].DifferencePP != "100" {
		t.Fatal(r)
	}
	b = histogramWindowFixture(t, []float64{1, 3, 5}, []string{"0", "0", "10", "0"})
	r = CompareHistogramWindows(a, b)
	if r[0].State != "incompatible" || len(r[0].Buckets) != 0 {
		t.Fatal("rebinned", r)
	}
	b = histogramWindowFixture(t, []float64{1, 2, 3}, []string{"10", "0", "0", "0"})
	r = CompareHistogramWindows(a, b)
	if r[0].P95Relation != "overlap-or-unbounded" {
		t.Fatal(r)
	}
	b.Series[0].Intervals[0].Distribution.Count = "0"
	b.Series[0].Intervals[0].Distribution.BucketCounts = []string{"0", "0", "0", "0"}
	r = CompareHistogramWindows(a, b)
	if r[0].State != "unavailable" || r[0].Selection.State != "no_data" {
		t.Fatal(r)
	}
}
func TestHistogramMultipleIntervalsNeverAverageQuantiles(t *testing.T) {
	a := histogramWindowFixture(t, []float64{1, 2, 3}, []string{"10", "0", "0", "0"})
	s := a.Series[0]
	first := s.Intervals[0]
	first.ToNS = "2000000000"
	second := first
	second.FromNS = "2000000000"
	second.ToNS = "3000000000"
	second.Distribution = distribution("s", []float64{1, 2, 3}, []uint64{0, 0, 90, 0}, 90)
	s.Intervals = []HistogramInterval{first, second}
	a.Series[0] = s
	r := CompareHistogramWindows(a, a)
	d := r[0].Baseline.Distribution
	if d.Count != "100" || d.P95.Lower == nil || *d.P95.Lower != 2 || r[0].Buckets[0].DifferencePP != "0" {
		t.Fatal(r)
	}
}

func TestWindowTotalsOverflowPolicy(t *testing.T) {
	a := counterWindowFixture(t, "1000000000", "3000000000", "18446744073709551615")
	first := a.Series[0].Intervals[0]
	first.ToNS = "2000000000"
	second := first
	second.FromNS = "2000000000"
	second.ToNS = "3000000000"
	a.Series[0].Intervals = []CounterInterval{first, second}
	r := CompareCounterWindows(a, a)
	if r[0].Baseline.Delta != "36893488147419103230" || r[0].RateDifference != "0" {
		t.Fatal("counter total overflow", r)
	}
	h := histogramWindowFixture(t, []float64{1}, []string{"10", "0"})
	hi := h.Series[0].Intervals[0]
	hi.ToNS = "2000000000"
	hi.Distribution = distribution("s", []float64{1}, []uint64{^uint64(0), 0}, ^uint64(0))
	hj := hi
	hj.FromNS = "2000000000"
	hj.ToNS = "3000000000"
	h.Series[0].Intervals = []HistogramInterval{hi, hj}
	rh := CompareHistogramWindows(h, h)
	if rh[0].State != "unavailable" || rh[0].Baseline.Distribution != nil {
		t.Fatal("histogram overflow accepted", rh)
	}
}
