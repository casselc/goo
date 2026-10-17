package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"testing"
)

func fleetCounter(points []metrics.CounterPoint, from, to int64, instance string) (*metrics.CounterResult, error) {
	copied := append([]metrics.CounterPoint{}, points...)
	for n := range copied {
		p := copied[n]
		s := p.Stream
		r := map[string]string{}
		for k, v := range s.Resource {
			r[k] = v
		}
		r["service.instance.id"] = instance
		s.Resource = r
		copied[n].Stream = s
	}
	return metrics.EvaluateCounters(metrics.CounterRequest{FromNS: fmt.Sprint(from), ToNS: fmt.Sprint(to), MaxGapNS: "2", Points: copied})
}

type fleetMachine struct {
	points                     [2][]metrics.CounterPoint
	start                      [2]int64
	values                     [2][3]int64
	deltas                     [2][3]int64
	from, end                  int64
	unsupported, missing, skew bool
}

func (m *fleetMachine) advance(tc hegel.TestCase, gap, reset bool) {
	m.from = m.end
	m.end += 2
	m.unsupported = gap || reset
	if gap {
		m.end += 4
	}
	for member := 0; member < 2; member++ {
		if reset && member == 1 {
			m.start[member] = m.from + 1
			m.values[member] = [3]int64{}
		}
		for n, o := range ratioOutcomes {
			v := int64(hegel.Draw(tc, hegel.Integers(0, 10)))
			m.deltas[member][n] = v
			m.values[member][n] += v
			m.points[member] = append(m.points[member], ratioPoint(m.start[member], m.end, m.values[member][n], o))
		}
	}
}
func (m *fleetMachine) RuleCalls(tc hegel.TestCase)             { m.advance(tc, false, false) }
func (m *fleetMachine) RuleOneMemberRestarts(tc hegel.TestCase) { m.advance(tc, false, true) }
func (m *fleetMachine) RuleGap(tc hegel.TestCase)               { m.advance(tc, true, false) }
func (m *fleetMachine) RuleMissingMember(tc hegel.TestCase)     { m.missing = !m.missing }
func (m *fleetMachine) RuleClockSkew(tc hegel.TestCase)         { m.skew = !m.skew }
func (m *fleetMachine) InvariantNoPartialFleet(tc hegel.TestCase) {
	req := metrics.FleetRequest{FromNS: fmt.Sprint(m.from), ToNS: fmt.Sprint(m.end), Members: []string{"one", "two"}, Policy: metrics.RatioPolicy{Outcomes: ratioOutcomes, Errors: []string{"error", "panic"}}}
	for n, instance := range []string{"one", "two"} {
		if n == 1 && m.missing {
			continue
		}
		r, e := fleetCounter(m.points[n], m.from, m.end, instance)
		if e != nil {
			panic(e)
		}
		if n == 1 && m.skew {
			for k := range r.Series {
				for j := range r.Series[k].Intervals {
					r.Series[k].Intervals[j].FromNS = fmt.Sprint(m.from + 1)
				}
			}
		}
		req.Inputs = append(req.Inputs, metrics.FleetInput{Instance: instance, Counter: r})
	}
	r, e := metrics.EvaluateFleet(req)
	if e != nil {
		panic(e)
	}
	unsupported := m.unsupported || m.missing || m.skew
	var calls, errors int64
	for member := 0; member < 2; member++ {
		for n, d := range m.deltas[member] {
			calls += d
			if n > 0 {
				errors += d
			}
		}
	}
	for _, i := range r.Series[0].Intervals {
		if unsupported {
			if i.State != "unknown" || i.Calls != "" || i.Percent != nil || i.Supported == 2 {
				panic("invented a partial fleet total")
			}
		} else {
			if i.Supported != 2 || i.Calls != fmt.Sprint(calls) || i.Errors != fmt.Sprint(errors) {
				panic("lost fleet population")
			}
			if calls == 0 {
				if i.State != "no_data" || i.Percent != nil {
					panic("divided idle fleet")
				}
			} else if i.State != "ok" || i.Percent == nil || *i.Percent < 0 || *i.Percent > 100 {
				panic("invalid fleet ratio")
			}
		}
	}
}
func TestHegelFleetStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &fleetMachine{start: [2]int64{1, 1}, from: 10, end: 12}
		for member := 0; member < 2; member++ {
			for _, o := range ratioOutcomes {
				m.points[member] = append(m.points[member], ratioPoint(1, 10, 0, o), ratioPoint(1, 12, 0, o))
			}
		}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantNoPartialFleet"))
	}, options(100)...)
}
func fleetObservedFixture(t *testing.T, mode string) *metrics.FleetResult {
	req := metrics.FleetRequest{FromNS: "10", ToNS: "12", Members: []string{"one", "two"}, Policy: metrics.RatioPolicy{Outcomes: ratioOutcomes, Errors: []string{"error", "panic"}}}
	for n, instance := range []string{"one", "two"} {
		if mode == "missing" && n == 1 {
			continue
		}
		points := []metrics.CounterPoint{}
		for k, o := range ratioOutcomes {
			v := []int64{8, 1, 1}[k]
			if n == 1 {
				v = []int64{90, 8, 2}[k]
			}
			if mode == "idle" {
				v = 0
			}
			start := int64(1)
			if mode == "reset" && n == 1 {
				start = 11
			}
			points = append(points, ratioPoint(1, 10, 0, o), ratioPoint(start, 12, v, o))
		}
		r, e := fleetCounter(points, 10, 12, instance)
		if e != nil {
			t.Fatal(e)
		}
		if mode == "skew" && n == 1 {
			for k := range r.Series {
				r.Series[k].Intervals[0].FromNS = "11"
			}
		}
		req.Inputs = append(req.Inputs, metrics.FleetInput{Instance: instance, Counter: r})
	}
	result, e := metrics.EvaluateFleet(req)
	if e != nil {
		t.Fatal(e)
	}
	return result
}
func TestQuintObservedFleetCoverage(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("fleet"), qobs.WithProcess("fixture"))
	for _, mode := range []string{"aligned", "missing", "skew", "reset", "idle", "aligned"} {
		r := fleetObservedFixture(t, mode)
		i := r.Series[0].Intervals[0]
		calls, errors := int64(-1), int64(-1)
		if i.State != "unknown" {
			fmt.Sscan(i.Calls, &calls)
			fmt.Sscan(i.Errors, &errors)
		}
		rec.Record(context.Background(), mode, nil, "obs.state", i.State, "obs.calls", calls, "obs.errors", errors, "obs.supported", int64(i.Supported))
	}
	checker, e := validate.NewChecker("model/fleet.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	if e = artifact("fleet-validation.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, e := os.ReadFile(actor.ITF)
			if e != nil {
				t.Fatal(e)
			}
			if e = artifact("fleet-observed.itf.json", data); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, index := range []int{1, 2, 3, 4} {
		original := map[string]any{}
		for k, v := range steps[index].Obs {
			original[k] = v
		}
		steps[index].Obs["state"] = "ok"
		steps[index].Obs["calls"] = int64(110)
		steps[index].Obs["errors"] = int64(12)
		steps[index].Obs["supported"] = int64(2)
		bad, e := checker.Check(context.Background(), steps)
		if e != nil {
			t.Fatal(e)
		}
		if bad.OK() {
			t.Fatal("accepted unsupported fleet total")
		}
		if e = artifact(fmt.Sprintf("fleet-negative-%d.txt", index), []byte(bad.Summary())); e != nil {
			t.Fatal(e)
		}
		for k, v := range original {
			steps[index].Obs[k] = v
		}
	}
}
