package metrics

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
)

func decimalPoint(end, value string) CounterPoint {
	p := point("0", end, value)
	p.Stream.Unit = "s"
	return p
}
func TestDecimalCountersExactIntervalsAndLegacyOptIn(t *testing.T) {
	q := CounterRequest{ValueDomain: DecimalCounterDomain, FromNS: "1000000000", ToNS: "4000000000", MaxGapNS: "4000000000", Points: []CounterPoint{decimalPoint("1000000000", "10000000000000000.1"), decimalPoint("4000000000", "10000000000000000.2")}}
	before, _ := json.Marshal(q)
	r, e := EvaluateCounters(q)
	if e != nil {
		t.Fatal(e)
	}
	got := r.Series[0].Intervals[0]
	if got.Delta != "0.1" || got.ExactPerSecond != "1/30" || got.PerSecond == nil || math.Abs(*got.PerSecond-1.0/30) > 1e-17 {
		t.Fatal(got)
	}
	after, _ := json.Marshal(q)
	if string(before) != string(after) {
		t.Fatal("mutated inputs")
	}
	reverse := q
	reverse.Points = []CounterPoint{q.Points[1], q.Points[0], q.Points[0]}
	rr, e := EvaluateCounters(reverse)
	if e != nil || !reflect.DeepEqual(r, rr) {
		t.Fatal("order/duplicate", rr, e)
	}
	q.ValueDomain = ""
	if _, e := EvaluateCounters(q); e == nil {
		t.Fatal("legacy integer domain relaxed")
	}
}
func TestDecimalCounterQualityAndEpochs(t *testing.T) {
	points := []CounterPoint{decimalPoint("1000000000", "1"), decimalPoint("2000000000", "1.25"), decimalPoint("3000000000", "1.1"), decimalPoint("4000000000", "2"), decimalPoint("5000000000", "0.1"), decimalPoint("6000000000", "0.2"), decimalPoint("9000000000", "0.5"), decimalPoint("10000000000", "0.6"), decimalPoint("11000000000", "0.7")}
	for i := 4; i < len(points); i++ {
		points[i].StartNS = "4000000000"
	}
	points[7].NoRecordedValue = true
	points[7].Value = "not a value"
	r, e := EvaluateCounters(CounterRequest{ValueDomain: DecimalCounterDomain, FromNS: "1000000000", ToNS: "11000000000", MaxGapNS: "2000000000", Points: points})
	if e != nil {
		t.Fatal(e)
	}
	want := []string{"", "counter-decrease", "invalid-counter-epoch", "counter-epoch-change", "", "excessive-gap", "no-recorded-value", "no-recorded-value"}
	for i, p := range r.Series[0].Intervals {
		if p.Reason != want[i] {
			t.Fatal(i, p, want[i])
		}
		if p.Reason != "" && (p.Delta != "" || p.PerSecond != nil || p.ExactPerSecond != "") {
			t.Fatal("unknown received arithmetic", p)
		}
	}
}
func TestDecimalCounterBoundaryDeltaAndRefusals(t *testing.T) {
	p := decimalPoint("3000000000", "0.75")
	p.Temporality = "delta"
	p.StartNS = "1000000000"
	q := CounterRequest{ValueDomain: DecimalCounterDomain, FromNS: "1000000000", ToNS: "3000000000", MaxGapNS: "3000000000", Points: []CounterPoint{p}}
	r, e := EvaluateCounters(q)
	if e != nil || r.Series[0].Intervals[0].ExactPerSecond != "3/8" {
		t.Fatal(r, e)
	}
	q.FromNS = "1500000000"
	r, e = EvaluateCounters(q)
	if e != nil || r.Series[0].Intervals[0].Reason != "boundary-crossing" {
		t.Fatal(r, e)
	}
	for _, v := range []string{"-1", "1.0", "01", "1e-1", "NaN", "1/2", strings.Repeat("1", 513)} {
		q.Points[0].Value = v
		if _, e := EvaluateCounters(q); e == nil {
			t.Fatal("accepted", v)
		}
	}
	q.Points[0].Value = "0." + strings.Repeat("0", 399) + "1"
	r, e = EvaluateCounters(q)
	if e != nil {
		t.Fatal(e)
	}
	q.FromNS = "1000000000"
	r, e = EvaluateCounters(q)
	if e != nil || r.Series[0].Intervals[0].Reason != "unrepresentable-rate" {
		t.Fatal(r, e)
	}
	a := decimalPoint("1000000000", "0.1")
	b := a
	b.Value = "0.10000000000000001"
	q.Points = []CounterPoint{a, b}
	if _, e := EvaluateCounters(q); e == nil {
		t.Fatal("conflicting raw decimals accepted")
	}
}
func TestDecimalHostAuthorityAndCountIsolation(t *testing.T) {
	a, b := decimalPoint("1000000000", "1"), decimalPoint("2000000000", "1.5")
	for _, p := range []*CounterPoint{&a, &b} {
		p.Stream.Resource = map[string]string{"host.id": "host"}
	}
	q := CounterRequest{ValueDomain: DecimalCounterDomain, FromNS: "1000000000", ToNS: "2000000000", MaxGapNS: "1000000000", Points: []CounterPoint{a, b}}
	if _, e := EvaluateCounters(q); e == nil {
		t.Fatal("service authority relaxed")
	}
	r, e := EvaluateHostCounters(q, "host")
	if e != nil {
		t.Fatal(e)
	}
	if _, e := EvaluateHostCounters(q, "other"); e == nil {
		t.Fatal("host authority broadened")
	}
	if _, e := EvaluateErrorRatios(*r, RatioPolicy{Outcomes: []string{"ok", "error"}, Errors: []string{"error"}}); e == nil {
		t.Fatal("decimal became count ratio")
	}
	if got := DiscoverCounterWindows(r); got.WindowCount != 0 || got.ExclusionCount == 0 {
		t.Fatal(got)
	}
	if got := CompareCounterWindows(r, r); len(got) != 1 || got[0].State == "observed" {
		t.Fatal(got)
	}
}
