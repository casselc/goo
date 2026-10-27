package metrics

import (
	"encoding/json"
	"testing"
)

func testStream() Stream {
	return Stream{Resource: map[string]string{"service.name": "app", "service.instance.id": "one"}, Scope: "goo", Version: "1", SchemaURL: "goo://v1", Name: "calls", Unit: "{invocation}", Attributes: map[string]string{"outcome": "ok"}}
}
func point(start, end, value string) CounterPoint {
	return CounterPoint{Stream: testStream(), Temporality: "cumulative", Monotonic: true, StartNS: start, EndNS: end, Value: value}
}
func TestCounterIntervals(t *testing.T) {
	cases := []struct {
		name   string
		points []CounterPoint
		state  string
		reason string
		delta  string
	}{
		{"steady", []CounterPoint{point("0", "10", "5"), point("0", "20", "8")}, "ok", "", "3"},
		{"zero", []CounterPoint{point("0", "10", "5"), point("0", "20", "5")}, "ok", "", "0"},
		{"reset", []CounterPoint{point("0", "10", "5"), point("12", "20", "2")}, "unknown", "counter-epoch-change", ""},
		{"decrease", []CounterPoint{point("0", "10", "5"), point("0", "20", "2")}, "unknown", "counter-decrease", ""},
		{"baseline", []CounterPoint{point("0", "10", "5")}, "no_data", "", ""},
		{"exactuint64", []CounterPoint{point("0", "10", "18446744073709551613"), point("0", "20", "18446744073709551615")}, "ok", "", "2"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, e := EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "20", MaxGapNS: "10", Points: c.points})
			if e != nil {
				t.Fatal(e)
			}
			s := r.Series[0]
			if s.State != c.state {
				t.Fatalf("%+v", s)
			}
			if len(s.Intervals) > 0 && (s.Intervals[0].Delta != c.delta || s.Intervals[0].Reason != c.reason) {
				t.Fatalf("%+v", s.Intervals)
			}
		})
	}
}
func TestCounterGapsLatePointsAndIdentity(t *testing.T) {
	request := CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{point("0", "10", "0"), point("0", "30", "4")}}
	old, e := EvaluateCounters(request)
	if e != nil {
		t.Fatal(e)
	}
	snapshot, _ := json.Marshal(old)
	if old.Series[0].State != "unknown" {
		t.Fatal("bridged gap")
	}
	request.Points = append(request.Points, point("0", "20", "2"))
	fresh, e := EvaluateCounters(request)
	if e != nil {
		t.Fatal(e)
	}
	if fresh.Series[0].State != "ok" || fresh.Basis == old.Basis {
		t.Fatal("late point did not revise basis")
	}
	again, _ := json.Marshal(old)
	if string(snapshot) != string(again) {
		t.Fatal("mutated previous evidence")
	}
	request.Points[0], request.Points[2] = request.Points[2], request.Points[0]
	request.Points = append(request.Points, request.Points[0])
	copy, e := EvaluateCounters(request)
	if e != nil || copy.Basis != fresh.Basis {
		t.Fatal("order/copies changed basis", e)
	}
	other := point("0", "20", "100")
	other.Stream.Resource["service.instance.id"] = "two"
	request.Points = append(request.Points, other)
	separate, e := EvaluateCounters(request)
	if e != nil || len(separate.Series) != 2 {
		t.Fatal("mixed instances", e)
	}
}
func TestCounterDeltaAndInvalidEvidence(t *testing.T) {
	a, b := point("10", "20", "3"), point("20", "30", "4")
	a.Temporality = "delta"
	b.Temporality = "delta"
	r, e := EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{a, b}})
	if e != nil || r.Series[0].State != "ok" {
		t.Fatal(r, e)
	}
	a.NoRecordedValue = true
	r, e = EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{a, b}})
	if e != nil || r.Series[0].State != "unknown" || r.Series[0].Intervals[0].PerSecond != nil {
		t.Fatal("missing point became zero", e)
	}
	a.NoRecordedValue = false
	b.StartNS = "19"
	if _, e = EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{a, b}}); e == nil {
		t.Fatal("overlap accepted")
	}
	a = point("0", "10", "1")
	b = point("0", "10", "2")
	if _, e = EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{a, b}}); e == nil {
		t.Fatal("conflicting copy accepted")
	}
	a = point("0", "10", "1")
	b = point("0", "20", "2")
	b.Temporality = "delta"
	if _, e = EvaluateCounters(CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: []CounterPoint{a, b}}); e == nil {
		t.Fatal("mixed temporality accepted")
	}
	a = point("0", "10", "1")
	b = point("0", "20", "2")
	r, e = EvaluateCounters(CounterRequest{FromNS: "11", ToNS: "19", MaxGapNS: "10", Points: []CounterPoint{a, b}})
	if e != nil || r.Series[0].Intervals[0].Reason != "boundary-crossing" {
		t.Fatal("interpolated boundary")
	}
}
