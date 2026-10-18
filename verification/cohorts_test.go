package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"strconv"
	"testing"
)

func cohort(n int, state string) exploration.Cohort {
	q := exploration.Query{Signal: "spans", Environment: "local", FromMS: 1, ToMS: 60001, Limit: 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}}
	c := exploration.Cohort{Question: q, Basis: "fixture", Count: strconv.Itoa(n), State: state, Groups: []exploration.ValueCount{}}
	if n > 0 {
		v := "east"
		c.Groups = []exploration.ValueCount{{Value: &v, Count: strconv.Itoa(n)}}
	}
	return c
}

type cohortMachine struct {
	n              int
	changed, drift bool
}

func (m *cohortMachine) RulePopulation(tc hegel.TestCase) {
	m.n = hegel.Draw(tc, hegel.Integers(0, 1000000))
}
func (m *cohortMachine) RuleSourceRevision(tc hegel.TestCase) { m.changed = !m.changed }
func (m *cohortMachine) RuleScope(tc hegel.TestCase)          { m.drift = !m.drift }
func (m *cohortMachine) InvariantNoInventedShares(tc hegel.TestCase) {
	a, b := cohort(12, "observed"), cohort(m.n, "observed")
	if m.n == 0 {
		b.State = "no_data"
	}
	if m.changed {
		b.State = "changed_between_reads"
	}
	if m.drift {
		b.Question.ToMS++
	}
	r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: a, Selection: b})
	if m.drift {
		if e == nil {
			panic("scope drift accepted")
		}
		return
	}
	if e != nil {
		panic(e)
	}
	expected := "ok"
	if m.changed {
		expected = "unknown"
	} else if m.n == 0 {
		expected = "no_data"
	}
	if r.State != expected {
		panic("incorrect comparison state")
	}
	if expected != "ok" {
		if len(r.Rows) != 0 {
			panic("invented normalized shares")
		}
		return
	}
	if len(r.Rows) != 1 || r.Rows[0].DifferencePP != "0.00" || r.Rows[0].SelectionCount != strconv.Itoa(m.n) {
		panic("same category changed its share")
	}
}
func TestHegelCohortStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &cohortMachine{n: 1}, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantNoInventedShares"))
	}, options(100)...)
}
func TestQuintObservedCohorts(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("cohorts"), qobs.WithProcess("fixture"))
	for _, action := range []string{"pin", "selection", "empty", "changed", "scopeChange"} {
		b := cohort(4, "observed")
		if action == "empty" {
			b = cohort(0, "no_data")
		}
		if action == "changed" {
			b.State = "changed_between_reads"
		}
		if action == "scopeChange" {
			b.Question.ToMS++
		}
		r, e := exploration.CompareCohorts(exploration.ComparisonRequest{Baseline: cohort(12, "observed"), Selection: b})
		supported := e == nil
		if action == "scopeChange" {
			supported = e != nil
		}
		honest := supported
		if e == nil {
			switch action {
			case "empty":
				honest = r.State == "no_data" && len(r.Rows) == 0
			case "changed":
				honest = r.State == "unknown" && len(r.Rows) == 0
			default:
				honest = r.State == "ok" && len(r.Rows) == 1 && r.Rows[0].SelectionPercent == "100.00"
			}
		}
		rec.Record(context.Background(), action, nil, "obs.scope", supported, "obs.honest", honest)
	}
	checker, e := validate.NewChecker("model/cohorts.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	artifact("cohorts-validation.txt", []byte(report.Summary()))
	for _, control := range []struct {
		i   int
		key string
	}{{2, "honest"}, {3, "honest"}, {4, "scope"}} {
		steps[control.i].Obs[control.key] = false
		bad, e := checker.Check(context.Background(), steps)
		if e != nil || bad.OK() {
			t.Fatalf("negative accepted %v", e)
		}
		artifact(fmt.Sprintf("cohorts-negative-%d.txt", control.i), []byte(bad.Summary()))
		steps[control.i].Obs[control.key] = true
	}
}
