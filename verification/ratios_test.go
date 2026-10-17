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

var ratioOutcomes = []string{"ok", "error", "panic"}

func ratioPoint(start, end, value int64, outcome string) metrics.CounterPoint {
	p := cp(start, end, value)
	p.Stream.Attributes = map[string]string{"goo.operation.name": "checkout", "goo.operation.role": "server", "goo.operation.outcome": outcome}
	return p
}
func evaluateRatioPoints(points []metrics.CounterPoint, from, to int64, missing bool) (*metrics.RatioResult, error) {
	r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: fmt.Sprint(from), ToNS: fmt.Sprint(to), MaxGapNS: "2", Points: points})
	if e != nil {
		return nil, e
	}
	if missing {
		s := []metrics.CounterSeries{}
		for _, v := range r.Series {
			if v.Stream.Attributes["goo.operation.outcome"] != "panic" {
				s = append(s, v)
			}
		}
		r.Series = s
	}
	return metrics.EvaluateErrorRatios(*r, metrics.RatioPolicy{Outcomes: ratioOutcomes, Errors: []string{"error", "panic"}})
}

type ratioMachine struct {
	points           []metrics.CounterPoint
	start, end, from int64
	values           [3]int64
	calls, errors    int64
	expected         string
	missing          bool
}

func (m *ratioMachine) advance(tc hegel.TestCase, gap, reset bool) {
	m.from = m.end
	m.end++
	m.expected = "ok"
	m.calls = 0
	m.errors = 0
	if gap {
		m.end += 4
		m.expected = "unknown"
	}
	if reset {
		m.start = m.from + 1
		m.end = m.from + 2
		m.values = [3]int64{}
		m.expected = "unknown"
	}
	for i, o := range ratioOutcomes {
		n := int64(hegel.Draw(tc, hegel.Integers(0, 5)))
		m.values[i] += n
		m.calls += n
		if i > 0 {
			m.errors += n
		}
		m.points = append(m.points, ratioPoint(m.start, m.end, m.values[i], o))
	}
	if m.expected == "ok" && m.calls == 0 {
		m.expected = "no_data"
	}
}
func (m *ratioMachine) RuleCalls(tc hegel.TestCase)          { m.advance(tc, false, false) }
func (m *ratioMachine) RuleRestart(tc hegel.TestCase)        { m.advance(tc, false, true) }
func (m *ratioMachine) RuleGap(tc hegel.TestCase)            { m.advance(tc, true, false) }
func (m *ratioMachine) RuleMissingOutcome(tc hegel.TestCase) { m.missing = !m.missing }
func (m *ratioMachine) InvariantNoInventedRatio(tc hegel.TestCase) {
	r, e := evaluateRatioPoints(m.points, m.from, m.end, m.missing)
	if e != nil {
		panic(fmt.Sprintf("%v", e))
	}
	expected := m.expected
	if m.missing {
		expected = "unknown"
	}
	i := r.Series[0].Intervals[0]
	if i.State != expected {
		panic(fmt.Sprintf("state %s want %s", i.State, expected))
	}
	if expected == "unknown" {
		if i.Percent != nil || i.Calls != "" {
			panic(fmt.Sprintf("invented ratio"))
		}
	} else {
		if i.Calls != fmt.Sprint(m.calls) || i.Errors != fmt.Sprint(m.errors) {
			panic(fmt.Sprintf("lost population"))
		}
		if expected == "no_data" {
			if i.Percent != nil {
				panic(fmt.Sprintf("zero divided"))
			}
		} else if i.Percent == nil || *i.Percent < 0 || *i.Percent > 100 {
			panic(fmt.Sprintf("invalid ratio"))
		}
	}
}
func TestHegelErrorRatioStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &ratioMachine{start: 1, end: 11, from: 10, expected: "no_data"}
		for _, o := range ratioOutcomes {
			m.points = append(m.points, ratioPoint(1, 10, 0, o), ratioPoint(1, 11, 0, o))
		}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantNoInventedRatio"))
	}, options(100)...)
}
func TestQuintObservedErrorRatios(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("ratio"), qobs.WithProcess("fixture"))
	points := []metrics.CounterPoint{}
	for _, o := range ratioOutcomes {
		points = append(points, ratioPoint(1, 10, 0, o))
	}
	record := func(action string, from, to int64, missing bool) {
		r, e := evaluateRatioPoints(points, from, to, missing)
		if e != nil {
			t.Fatal(e)
		}
		i := r.Series[0].Intervals[0]
		calls, errors := int64(-1), int64(-1)
		if i.State != "unknown" {
			fmt.Sscan(i.Calls, &calls)
			fmt.Sscan(i.Errors, &errors)
		}
		rec.Record(context.Background(), action, nil, "obs.state", i.State, "obs.calls", calls, "obs.errors", errors)
	}
	for n, o := range ratioOutcomes {
		v := int64(0)
		if n == 0 {
			v = 8
		}
		if n == 1 {
			v = 2
		}
		points = append(points, ratioPoint(1, 11, v, o))
	}
	record("traffic", 10, 11, false)
	record("missing", 10, 11, true)
	for n, o := range ratioOutcomes {
		v := int64(0)
		if n == 0 {
			v = 8
		}
		if n == 1 {
			v = 2
		}
		points = append(points, ratioPoint(1, 12, v, o))
	}
	record("idle", 11, 12, false)
	for _, o := range ratioOutcomes {
		points = append(points, ratioPoint(13, 14, 0, o))
	}
	record("reset", 12, 14, false)
	for n, o := range ratioOutcomes {
		points = append(points, ratioPoint(13, 18, int64(n), o))
	}
	record("gap", 14, 18, false)
	for n, o := range ratioOutcomes {
		points = append(points, ratioPoint(13, 16, int64(n), o))
	}
	record("repair", 14, 16, false)
	checker, e := validate.NewChecker("model/ratio.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	if e = artifact("ratio-validation.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, e := os.ReadFile(actor.ITF)
			if e != nil {
				t.Fatal(e)
			}
			if e = artifact("ratio-observed.itf.json", data); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, index := range []int{1, 2, 3, 4} {
		state, calls, errors := steps[index].Obs["state"], steps[index].Obs["calls"], steps[index].Obs["errors"]
		steps[index].Obs["state"] = "ok"
		steps[index].Obs["calls"] = int64(10)
		steps[index].Obs["errors"] = int64(2)
		bad, e := checker.Check(context.Background(), steps)
		if e != nil {
			t.Fatal(e)
		}
		if bad.OK() {
			t.Fatalf("accepted invented ratio at %d", index)
		}
		if e = artifact(fmt.Sprintf("ratio-negative-%d.txt", index), []byte(bad.Summary())); e != nil {
			t.Fatal(e)
		}
		steps[index].Obs["state"], steps[index].Obs["calls"], steps[index].Obs["errors"] = state, calls, errors
	}
}
