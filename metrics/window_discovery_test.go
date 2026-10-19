package metrics

import (
	"encoding/json"
	"strconv"
	"testing"
)

func supportStream(instance string) Stream {
	return Stream{Resource: map[string]string{"service.name": "search", "service.instance.id": instance}, Scope: "otel.http", Name: "calls", Unit: "{request}"}
}
func TestDiscoverWindowsKeepsGapsTaintAndAllMembers(t *testing.T) {
	s := supportStream("one")
	point := func(a, b string, missing bool) CounterPoint {
		return CounterPoint{Stream: s, Temporality: "delta", Monotonic: true, StartNS: a, EndNS: b, Value: "6", NoRecordedValue: missing}
	}
	points := []CounterPoint{point("101", "201", false), point("201", "301", false), point("401", "501", false), point("501", "601", true)}
	r, e := EvaluateCounters(CounterRequest{FromNS: "100", ToNS: "700", MaxGapNS: "1000", Points: points})
	if e != nil {
		t.Fatal(e)
	}
	before, _ := json.Marshal(r)
	d := DiscoverCounterWindows(r)
	after, _ := json.Marshal(r)
	if string(before) != string(after) || len(d.Windows) != 2 || d.Windows[0].FromNS != "101" || d.Windows[0].ToNS != "301" || d.Windows[0].Members[0].Summary.Delta != "12" || d.Windows[1].FromNS != "401" || d.ExclusionCount != 4 {
		t.Fatalf("%+v", d)
	}
	for _, w := range d.Windows {
		check, e := EvaluateCounters(CounterRequest{FromNS: w.FromNS, ToNS: w.ToNS, MaxGapNS: "1000", Points: points})
		if e != nil || check.Series[0].State != "ok" {
			t.Fatal(w, e)
		}
	}
	other := r.Series[0]
	other.Stream = supportStream("two")
	other.Intervals = []CounterInterval{{FromNS: "102", ToNS: "202", State: "ok", Delta: "6"}}
	r.Series = append(r.Series, other)
	if d = DiscoverCounterWindows(r); len(d.Windows) != 0 {
		t.Fatal("interpolated asynchronous boundaries", d)
	}
	cumulative := []CounterPoint{}
	for i, value := range []string{"10", "5", "8", "12"} {
		cumulative = append(cumulative, CounterPoint{Stream: s, Temporality: "cumulative", Monotonic: true, StartNS: "1", EndNS: strconv.Itoa(100 + i*100), Value: value})
	}
	reset, e := EvaluateCounters(CounterRequest{FromNS: "100", ToNS: "400", MaxGapNS: "1000", Points: cumulative})
	if e != nil {
		t.Fatal(e)
	}
	if d = DiscoverCounterWindows(reset); len(d.Windows) != 0 {
		t.Fatal("discarded reset taint", d)
	}
}
func TestDiscoverHistogramSplitsLayoutsAndRefusesOverflow(t *testing.T) {
	s := supportStream("one")
	s.Unit = "s"
	s.Name = "duration"
	p := func(a, b string, bound float64, n string) HistogramPoint {
		return HistogramPoint{Stream: s, Temporality: "delta", StartNS: a, EndNS: b, Bounds: []float64{bound}, Count: n, BucketCounts: []string{n, "0"}}
	}
	r, e := EvaluateHistograms(HistogramRequest{FromNS: "100", ToNS: "400", MaxGapNS: "1000", Points: []HistogramPoint{p("100", "200", 1, "4"), p("200", "300", 1, "6"), p("300", "400", 2, "10")}})
	if e != nil {
		t.Fatal(e)
	}
	d := DiscoverHistogramWindows(r)
	if len(d.Windows) != 2 || d.Windows[0].ToNS != "300" || d.Windows[0].Members[0].Summary.Distribution.Count != "10" || d.Windows[0].Signature == d.Windows[1].Signature {
		t.Fatal(d)
	}
	r, e = EvaluateHistograms(HistogramRequest{FromNS: "100", ToNS: "300", MaxGapNS: "1000", Points: []HistogramPoint{p("100", "200", 1, "18446744073709551615"), p("200", "300", 1, "1")}})
	if e != nil {
		t.Fatal(e)
	}
	d = DiscoverHistogramWindows(r)
	if len(d.Windows) != 0 || d.Exclusions[0].Reason != "aggregate-unavailable" {
		t.Fatal("offered overflowing population", d)
	}
}
func TestDiscoverBoundsOutputAndLongRun(t *testing.T) {
	s := supportStream("one")
	r := &CounterResult{FromNS: "0", ToNS: "20000", Series: []CounterSeries{{Stream: s, Temporality: "delta", State: "ok"}}}
	for i := 0; i < 10000; i++ {
		r.Series[0].Intervals = append(r.Series[0].Intervals, CounterInterval{FromNS: strconv.Itoa(i * 2), ToNS: strconv.Itoa(i*2 + 2), State: "ok", Delta: "1"})
	}
	d := DiscoverCounterWindows(r)
	if len(d.Windows) != 1 || d.Windows[0].Members[0].Summary.Delta != "10000" {
		t.Fatal("long contiguous support")
	}
	for i := range r.Series[0].Intervals {
		if i%2 == 1 {
			r.Series[0].Intervals[i].State = "unknown"
		}
	}
	d = DiscoverCounterWindows(r)
	if len(d.Windows) != 32 || d.WindowCount != 5000 || len(d.Exclusions) != 64 || d.ExclusionCount != 5000 {
		t.Fatal("silent output truncation")
	}
}

func TestDiscoverEmptyHistogramRemainsNoData(t *testing.T) {
	s := supportStream("one")
	s.Unit = "s"
	r, e := EvaluateHistograms(HistogramRequest{FromNS: "100", ToNS: "200", MaxGapNS: "1000", Points: []HistogramPoint{{Stream: s, Temporality: "delta", StartNS: "100", EndNS: "200", Bounds: []float64{1}, Count: "0", BucketCounts: []string{"0", "0"}}}})
	if e != nil {
		t.Fatal(e)
	}
	d := DiscoverHistogramWindows(r)
	if len(d.Windows) != 0 || d.ExclusionCount != 1 || d.Exclusions[0].Reason != "empty-distribution" {
		t.Fatal(d)
	}
}
