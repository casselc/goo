package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"strconv"
	"testing"
)

type comparisonMachine struct{ fleetMachine }

func (m *comparisonMachine) InvariantBoundedComparison(tc hegel.TestCase) {
	req := metrics.FleetRequest{FromNS: fmt.Sprint(m.from), ToNS: fmt.Sprint(m.end), Members: []string{"one", "two"}, Policy: metrics.RatioPolicy{Outcomes: ratioOutcomes, Errors: []string{"error", "panic"}}, Comparison: &metrics.ComparisonPolicy{Mode: "bounded-skew", MaxSkewNS: "1"}}
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
	count := 0
	for _, row := range r.Comparison.Series[0].Intervals {
		if row.State == "comparable" {
			count++
			if m.unsupported || m.missing {
				panic("compared unsupported windows")
			}
			a, _ := strconv.ParseInt(row.StartSpreadNS, 10, 64)
			b, _ := strconv.ParseInt(row.EndSpreadNS, 10, 64)
			if a > 1 || b > 1 {
				panic("tolerance exceeded")
			}
			for _, member := range row.Members {
				if member.State != "ok" && member.State != "no_data" {
					panic("invented member counts")
				}
			}
		}
	}
	if !m.unsupported && !m.missing && count != 1 {
		panic("lost valid bounded comparison")
	}
	if m.skew || m.missing || m.unsupported {
		for _, i := range r.Series[0].Intervals {
			if i.State != "unknown" || i.Percent != nil {
				panic("comparison invented exact fleet percent")
			}
		}
	}
}
func TestHegelComparisonStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &comparisonMachine{fleetMachine: fleetMachine{start: [2]int64{1, 1}, from: 10, end: 12}}
		for member := 0; member < 2; member++ {
			for _, o := range ratioOutcomes {
				m.points[member] = append(m.points[member], ratioPoint(1, 10, 0, o), ratioPoint(1, 12, 0, o))
			}
		}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantBoundedComparison"))
	}, options(100)...)
}
func comparisonObserved(t *testing.T, mode string) *metrics.FleetResult {
	q := metrics.FleetRequest{FromNS: "10", ToNS: "30", Members: []string{"one", "two"}, Policy: metrics.RatioPolicy{Outcomes: ratioOutcomes, Errors: []string{"error", "panic"}}, Comparison: &metrics.ComparisonPolicy{Mode: "bounded-skew", MaxSkewNS: "2"}}
	for n, instance := range []string{"one", "two"} {
		offset := int64(0)
		if n == 1 && mode != "aligned" {
			offset = 1
		}
		if n == 1 && mode == "outside" {
			offset = 3
		}
		points := []metrics.CounterPoint{}
		for k, o := range ratioOutcomes {
			value := []int64{8, 1, 1}[k]
			if n == 1 {
				value = []int64{90, 8, 2}[k]
			}
			p := ratioPoint(1, 10+offset, 0, o)
			p.Stream.Resource["service.instance.id"] = instance
			next := ratioPoint(1, 20+offset, value, o)
			next.Stream.Resource["service.instance.id"] = instance
			points = append(points, p, next)
		}
		r, e := metrics.EvaluateCounters(metrics.CounterRequest{FromNS: "10", ToNS: "30", MaxGapNS: "10", Points: points})
		if e != nil {
			t.Fatal(e)
		}
		if n == 1 && (mode == "loss" || mode == "reset") {
			reason := "excessive-gap"
			if mode == "reset" {
				reason = "counter-epoch-change"
			}
			for k := range r.Series {
				r.Series[k].Intervals[0].State = "unknown"
				r.Series[k].Intervals[0].Delta = ""
				r.Series[k].Intervals[0].PerSecond = nil
				r.Series[k].Intervals[0].Reason = reason
			}
		}
		q.Inputs = append(q.Inputs, metrics.FleetInput{Instance: instance, Counter: r})
	}
	r, e := metrics.EvaluateFleet(q)
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestQuintObservedBoundedComparison(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("comparison"), qobs.WithProcess("fixture"))
	for _, mode := range []string{"aligned", "skew", "outside", "loss", "reset", "recovered"} {
		r := comparisonObserved(t, mode)
		comparable, exact := false, false
		spread := int64(-1)
		for _, i := range r.Comparison.Series[0].Intervals {
			if i.State == "comparable" {
				comparable = true
				fmt.Sscan(i.StartSpreadNS, &spread)
			}
		}
		for _, i := range r.Series[0].Intervals {
			if i.State == "ok" {
				exact = true
			}
		}
		rec.Record(context.Background(), mode, nil, "obs.comparable", comparable, "obs.exact", exact, "obs.spread", spread)
	}
	checker, e := validate.NewChecker("model/comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	if e = artifact("comparison-validation.txt", []byte(report.Summary())); e != nil {
		t.Fatal(e)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, e := os.ReadFile(actor.ITF)
			if e != nil {
				t.Fatal(e)
			}
			if e = artifact("comparison-observed.itf.json", data); e != nil {
				t.Fatal(e)
			}
		}
	}
	for _, index := range []int{1, 3, 4} {
		original := map[string]any{}
		for k, v := range steps[index].Obs {
			original[k] = v
		}
		steps[index].Obs["comparable"] = true
		steps[index].Obs["exact"] = true
		steps[index].Obs["spread"] = int64(0)
		bad, e := checker.Check(context.Background(), steps)
		if e != nil {
			t.Fatal(e)
		}
		if bad.OK() {
			t.Fatal("invented common population accepted")
		}
		if e = artifact(fmt.Sprintf("comparison-negative-%d.txt", index), []byte(bad.Summary())); e != nil {
			t.Fatal(e)
		}
		for k, v := range original {
			steps[index].Obs[k] = v
		}
	}
}
