package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"strconv"
	"testing"
)

func counterStream() metrics.Stream {
	return metrics.Stream{Resource: map[string]string{"service.name": "test", "service.instance.id": "one"}, Scope: "goo", Version: "1", SchemaURL: "goo://v1", Name: "calls", Unit: "{invocation}", Attributes: map[string]string{}}
}
func cp(start, end, value int64) metrics.CounterPoint {
	return metrics.CounterPoint{Stream: counterStream(), Temporality: "cumulative", Monotonic: true, StartNS: fmt.Sprint(start), EndNS: fmt.Sprint(end), Value: fmt.Sprint(value)}
}

type counterMachine struct {
	points                  []metrics.CounterPoint
	start, end, value, from int64
	tainted                 bool
	expected                string
}

func (m *counterMachine) add(end, value, start int64) {
	m.from = m.end
	m.expected = "ok"
	if start != m.start {
		m.expected = "unknown"
		m.tainted = false
	} else if value < m.value {
		m.expected = "unknown"
		m.tainted = true
	} else if m.tainted || end-m.end > 2 {
		m.expected = "unknown"
	}
	m.start, m.end, m.value = start, end, value
	m.points = append(m.points, cp(start, end, value))
}
func (m *counterMachine) RuleIncrease(tc hegel.TestCase) {
	m.add(m.end+1, m.value+int64(hegel.Draw(tc, hegel.Integers(0, 8))), m.start)
}
func (m *counterMachine) RuleGap(tc hegel.TestCase)     { m.add(m.end+5, m.value+1, m.start) }
func (m *counterMachine) RuleRestart(tc hegel.TestCase) { m.add(m.end+2, 0, m.end+1) }
func (m *counterMachine) RuleDecrease(tc hegel.TestCase) {
	if m.value > 0 {
		m.add(m.end+1, m.value-1, m.start)
	}
}
func (m *counterMachine) RuleDuplicate(tc hegel.TestCase) {
	m.points = append(m.points, m.points[len(m.points)-1])
}
func (m *counterMachine) InvariantNoInventedRate(tc hegel.TestCase) {
	result, err := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: fmt.Sprint(m.from), ToNS: fmt.Sprint(m.end), MaxGapNS: "2", Points: m.points})
	if err != nil {
		panic(err)
	}
	if result.Series[0].State != m.expected {
		panic(fmt.Sprintf("state %s expected %s", result.Series[0].State, m.expected))
	}
	for _, i := range result.Series[0].Intervals {
		if i.State != "ok" && i.PerSecond != nil {
			panic("unknown interval acquired rate")
		}
	}
}
func TestHegelCounterStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &counterMachine{points: []metrics.CounterPoint{cp(1, 10, 0)}, start: 1, end: 10, from: 9, expected: "no_data"}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantNoInventedRate"))
	}, options(100)...)
}
func TestQuintObservedCounterEpochs(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("counter"), qobs.WithProcess("fixture"))
	points := []metrics.CounterPoint{cp(1, 10, 5)}
	for _, step := range []struct {
		action string
		point  metrics.CounterPoint
		from   int64
	}{{"increase", cp(1, 20, 8), 10}, {"restart", cp(21, 30, 2), 20}, {"increase", cp(21, 40, 4), 30}, {"gap", cp(21, 70, 9), 40}} {
		points = append(points, step.point)
		r, err := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: fmt.Sprint(step.from), ToNS: step.point.EndNS, MaxGapNS: "10", Points: points})
		if err != nil {
			t.Fatal(err)
		}
		known := r.Series[0].State == "ok"
		delta := int64(-1)
		if known {
			delta, _ = strconv.ParseInt(r.Series[0].Intervals[0].Delta, 10, 64)
		}
		rec.Record(context.Background(), step.action, nil, "obs.known", known, "obs.delta", delta)
	}
	points = append(points, cp(21, 50, 6), cp(21, 60, 8))
	repaired, err := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: "40", ToNS: "70", MaxGapNS: "10", Points: points})
	if err != nil {
		t.Fatal(err)
	}
	delta := int64(0)
	for _, interval := range repaired.Series[0].Intervals {
		d, _ := strconv.ParseInt(interval.Delta, 10, 64)
		delta += d
	}
	rec.Record(context.Background(), "late_repair", nil, "obs.known", repaired.Series[0].State == "ok", "obs.delta", delta)
	checker, err := validate.NewChecker("model/counter.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	steps := sink.Steps()
	report, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatal(report.Summary())
	}
	if err = artifact("counter-validation.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, e := os.ReadFile(actor.ITF)
			if e != nil {
				t.Fatal(e)
			}
			artifact("counter-observed.itf.json", data)
		}
	}
	data, _ := json.MarshalIndent(steps, "", "  ")
	artifact("counter-observed.json", data)
	steps[1].Obs["known"] = true
	steps[1].Obs["delta"] = int64(2)
	bad, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if bad.OK() {
		t.Fatal("accepted rate crossing restart")
	}
	artifact("counter-negative.txt", []byte(bad.Summary()))
}
