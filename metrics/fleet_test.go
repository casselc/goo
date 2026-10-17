package metrics

import (
	"testing"
)

func fleetFixture() FleetRequest {
	one := ratioFixture("8", "1", "1")
	two := ratioFixture("90", "8", "2")
	for i := range two.Series {
		two.Series[i].Stream.Resource["service.instance.id"] = "two"
	}
	return FleetRequest{FromNS: "10", ToNS: "20", Members: []string{"one", "two"}, Policy: ratioPolicy(), Inputs: []FleetInput{{"one", &one}, {"two", &two}}}
}
func TestFleetExactPopulationsAndCoverage(t *testing.T) {
	r, e := EvaluateFleet(fleetFixture())
	if e != nil {
		t.Fatal(e)
	}
	i := r.Series[0].Intervals[0]
	if i.Calls != "110" || i.Errors != "12" || i.Fraction != "12/110" || i.Supported != 2 || i.Percent == nil || *i.Percent < 10 || *i.Percent > 11 {
		t.Fatal(i)
	}
	// 20% and 10% must not become a 15% fleet rate.
	for _, mode := range []string{"missing", "unavailable", "empty", "reset", "unaligned", "resource"} {
		t.Run(mode, func(t *testing.T) {
			q := fleetFixture()
			switch mode {
			case "missing":
				q.Inputs = q.Inputs[:1]
			case "unavailable":
				q.Inputs[1].Counter = nil
			case "empty":
				q.Inputs[1].Counter.Series = nil
			case "reset":
				q.Inputs[1].Counter.Series[0].Intervals[0] = CounterInterval{FromNS: "10", ToNS: "20", State: "unknown", Reason: "reset"}
			case "unaligned":
				for n := range q.Inputs[1].Counter.Series {
					q.Inputs[1].Counter.Series[n].Intervals[0].FromNS = "11"
				}
			case "resource":
				for n := range q.Inputs[1].Counter.Series {
					q.Inputs[1].Counter.Series[n].Stream.Resource["region"] = "other"
				}
			}
			r, e := EvaluateFleet(q)
			if e != nil {
				t.Fatal(e)
			}
			for _, s := range r.Series {
				for _, i := range s.Intervals {
					if i.State != "unknown" || i.Percent != nil || i.Calls != "" || i.Errors != "" || i.Supported == i.Expected {
						t.Fatalf("invented fleet total %+v", i)
					}
				}
			}
		})
	}
}
func TestFleetZeroExactOverflowAndBasis(t *testing.T) {
	q := fleetFixture()
	for _, input := range q.Inputs {
		for n := range input.Counter.Series {
			input.Counter.Series[n].Intervals[0].Delta = "0"
		}
	}
	r, e := EvaluateFleet(q)
	if e != nil {
		t.Fatal(e)
	}
	i := r.Series[0].Intervals[0]
	if i.State != "no_data" || i.Percent != nil || i.Calls != "0" || i.Supported != 2 {
		t.Fatal(i)
	}
	q = fleetFixture()
	for _, input := range q.Inputs {
		for n := range input.Counter.Series {
			input.Counter.Series[n].Intervals[0].Delta = "18446744073709551615"
		}
	}
	r, e = EvaluateFleet(q)
	if e != nil || r.Series[0].Intervals[0].Calls != "110680464442257309690" {
		t.Fatal(r, e)
	}
	q = fleetFixture()
	a, e := EvaluateFleet(q)
	if e != nil {
		t.Fatal(e)
	}
	q.Members[0], q.Members[1] = q.Members[1], q.Members[0]
	q.Inputs[0], q.Inputs[1] = q.Inputs[1], q.Inputs[0]
	b, e := EvaluateFleet(q)
	if e != nil || a.Basis != b.Basis {
		t.Fatal("order dependence", e)
	}
	q.Inputs[0].Counter.Series[0].Intervals[0].Delta = "91"
	c, e := EvaluateFleet(q)
	if e != nil || c.Basis == a.Basis {
		t.Fatal("missing evidence basis", e)
	}
	if a.Series[0].Intervals[0].Calls != "110" {
		t.Fatal("prior result changed")
	}
}
func TestFleetRejectsInvalidIdentityAndMembership(t *testing.T) {
	for _, mode := range []string{"duplicate-member", "selected-only", "undeclared", "duplicate-input", "identity", "window"} {
		t.Run(mode, func(t *testing.T) {
			q := fleetFixture()
			switch mode {
			case "duplicate-member":
				q.Members[1] = "one"
			case "selected-only":
				q.Members = q.Members[:1]
			case "undeclared":
				q.Inputs[1].Instance = "unknown"
			case "duplicate-input":
				q.Inputs = append(q.Inputs, q.Inputs[0])
			case "identity":
				q.Inputs[1].Counter.Series[0].Stream.Resource["service.instance.id"] = "one"
			case "window":
				q.Inputs[1].Counter.FromNS = "11"
			}
			if _, e := EvaluateFleet(q); e == nil {
				t.Fatal("accepted invalid fleet")
			}
		})
	}
}
