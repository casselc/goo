package metrics

import "testing"

func comparisonFixture() FleetRequest {
	q := fleetFixture()
	q.ToNS = "30"
	q.Comparison = &ComparisonPolicy{Mode: "bounded-skew", MaxSkewNS: "2"}
	for _, input := range q.Inputs {
		input.Counter.ToNS = "30"
	}
	for n := range q.Inputs[1].Counter.Series {
		q.Inputs[1].Counter.Series[n].Intervals[0].FromNS = "11"
		q.Inputs[1].Counter.Series[n].Intervals[0].ToNS = "21"
	}
	return q
}
func TestFleetComparisonRetainsActualWindows(t *testing.T) {
	q := comparisonFixture()
	r, e := EvaluateFleet(q)
	if e != nil {
		t.Fatal(e)
	}
	rows := r.Comparison.Series[0].Intervals
	var known *ComparisonInterval
	for i := range rows {
		if rows[i].State == "comparable" {
			known = &rows[i]
			break
		}
	}
	if known == nil || known.StartSpreadNS != "1" || known.EndSpreadNS != "1" {
		t.Fatal(rows)
	}
	if known.Members[0].Calls != "10" || known.Members[1].Calls != "100" || known.Members[1].FromNS != "11" || known.Members[1].ToNS != "21" {
		t.Fatal("windows or counts shifted", known)
	}
	for _, i := range r.Series[0].Intervals {
		if i.State != "unknown" || i.Calls != "" || i.Percent != nil {
			t.Fatal("comparison invented exact fleet total")
		}
	}
	old := r.Comparison.Basis
	q.Members[0], q.Members[1] = q.Members[1], q.Members[0]
	q.Inputs[0], q.Inputs[1] = q.Inputs[1], q.Inputs[0]
	r, e = EvaluateFleet(q)
	if e != nil || r.Comparison.Basis != old {
		t.Fatal("order dependence", e)
	}
	q.Comparison.MaxSkewNS = "1"
	r, e = EvaluateFleet(q)
	if e != nil || r.Comparison.Basis == old {
		t.Fatal("policy missing from basis", e)
	}
}
func TestFleetComparisonMissingResetAndTolerance(t *testing.T) {
	for _, mode := range []string{"missing", "gap", "skew", "policy"} {
		t.Run(mode, func(t *testing.T) {
			q := comparisonFixture()
			switch mode {
			case "missing":
				q.Inputs = q.Inputs[:1]
			case "gap":
				q.Inputs[1].Counter.Series[0].Intervals[0] = CounterInterval{FromNS: "11", ToNS: "21", State: "unknown", Reason: "gap"}
			case "skew":
				for n := range q.Inputs[1].Counter.Series {
					q.Inputs[1].Counter.Series[n].Intervals[0].FromNS = "13"
				}
			case "policy":
				q.Comparison.Mode = "interpolate"
			}
			r, e := EvaluateFleet(q)
			if mode == "policy" {
				if e == nil {
					t.Fatal("invalid policy")
				}
				return
			}
			if e != nil {
				t.Fatal(e)
			}
			for _, i := range r.Comparison.Series[0].Intervals {
				if i.State == "comparable" {
					t.Fatal("accepted unsupported comparison", i)
				}
			}
		})
	}
}
func TestFleetComparisonDoesNotReuseSupportedIntervals(t *testing.T) {
	q := comparisonFixture()
	for n := range q.Inputs[0].Counter.Series {
		q.Inputs[0].Counter.Series[n].Intervals = []CounterInterval{{FromNS: "10", ToNS: "12", State: "ok", Delta: "1"}, {FromNS: "12", ToNS: "14", State: "ok", Delta: "1"}}
	}
	for n := range q.Inputs[1].Counter.Series {
		q.Inputs[1].Counter.Series[n].Intervals = []CounterInterval{{FromNS: "11", ToNS: "13", State: "ok", Delta: "1"}}
	}
	r, e := EvaluateFleet(q)
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, i := range r.Comparison.Series[0].Intervals {
		if i.State == "comparable" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("reused member interval", count)
	}
}
