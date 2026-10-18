package verification

import (
	"context"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
)

type explorationMachine struct {
	q          exploration.Query
	field      exploration.Field
	values     map[exploration.Field]string
	predicates []exploration.Predicate
	any        bool
}

func (m *explorationMachine) RuleAddFilter(tc hegel.TestCase) {
	if len(m.predicates) >= 12 {
		m.predicates = nil
	}
	values := []string{"", "a", "x'] OR 1=1 --", "🔥"}
	v := values[hegel.Draw(tc, hegel.Integers(0, 3))]
	m.predicates = append(m.predicates, exploration.Predicate{Field: &m.field, Op: "eq", Value: v})
}
func (m *explorationMachine) RuleMissing(tc hegel.TestCase)      { delete(m.values, m.field) }
func (m *explorationMachine) RulePresentEmpty(tc hegel.TestCase) { m.values[m.field] = "" }
func (m *explorationMachine) RulePresentValue(tc hegel.TestCase) { m.values[m.field] = "a" }
func (m *explorationMachine) RuleAnyOrAll(tc hegel.TestCase)     { m.any = !m.any }
func (m *explorationMachine) RulePage(tc hegel.TestCase) {
	m.q.Offset = hegel.Draw(tc, hegel.Integers(0, 10)) * 100
	m.q.Limit = hegel.Draw(tc, hegel.Integers(1, 100))
}
func (m *explorationMachine) InvariantScopeAndPopulation(tc hegel.TestCase) {
	q := m.q
	if len(m.predicates) > 0 {
		q.Where = &exploration.Predicate{All: m.predicates}
		if m.any {
			q.Where = &exploration.Predicate{Any: m.predicates}
		}
	}
	plan, e := exploration.Compile(q)
	if e != nil {
		panic(e)
	}
	q.Offset = 0
	q.Limit = 100
	other, e := exploration.Compile(q)
	if e != nil || plan.Count != other.Count || plan.Groups != other.Groups {
		panic("pagination changes population")
	}
	if !strings.Contains(plan.Rows, "environment:String") || strings.Contains(plan.Rows, "x'] OR 1=1 --") {
		panic("scope or parameter safety violated")
	}
	v, present := m.values[m.field]
	expected := len(m.predicates) == 0 || !m.any
	if len(m.predicates) > 0 {
		for _, p := range m.predicates {
			match := present && v == p.Value
			if m.any {
				expected = expected || match
			} else {
				expected = expected && match
			}
		}
	}
	if exploration.Match(q.Where, m.values) != expected {
		panic("predicate composition or missing-value semantics changed")
	}
	trace, e := exploration.TraceQuestion(q.Environment, q.FromMS, q.ToMS, "0123456789abcdef0123456789abcdef", "", "spans", 0)
	if e != nil || len(trace.Where.All) != 1 || trace.Where.All[0].Field.Key != "trace_id" {
		panic("context inherited query restrictions")
	}
}
func TestHegelExplorationStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &explorationMachine{q: exploration.Query{Signal: "spans", Environment: "local", FromMS: 1, ToMS: 60001, Limit: 100}, field: exploration.Field{Kind: "attribute", Key: "tenant'[]"}, values: map[exploration.Field]string{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantScopeAndPopulation"))
	}, options(100)...)
}
func TestQuintObservedExploration(t *testing.T) {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("exploration"), qobs.WithProcess("fixture"))
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 10, ToMS: 60010, Limit: 100}
	q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "one"}
	q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	initial, _ := exploration.Compile(q)
	for _, action := range []string{"search", "page", "trace", "logs"} {
		scope, counts, exact := true, true, true
		switch action {
		case "page":
			q.Offset = 100
			p, e := exploration.Compile(q)
			if e != nil {
				t.Fatal(e)
			}
			counts = p.Count == initial.Count
		case "trace", "logs":
			signal := "spans"
			span := ""
			if action == "logs" {
				signal = "logs"
				span = "0123456789abcdef"
			}
			expanded, e := exploration.TraceQuestion(q.Environment, q.FromMS, q.ToMS, "0123456789abcdef0123456789abcdef", span, signal, 0)
			if e != nil {
				t.Fatal(e)
			}
			scope = expanded.Environment == q.Environment && expanded.FromMS == q.FromMS && expanded.ToMS == q.ToMS
			exact = len(expanded.Where.All) == 1 || action == "logs" && len(expanded.Where.All) == 2
			for _, p := range expanded.Where.All {
				exact = exact && (p.Field.Key == "trace_id" || action == "logs" && p.Field.Key == "span_id")
			}
		}
		rec.Record(context.Background(), action, nil, "obs.scope", scope, "obs.population", counts, "obs.exact", exact)
	}
	checker, e := validate.NewChecker("model/exploration.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	artifact("exploration-validation.txt", []byte(report.Summary()))
	for i, key := range []string{"population", "scope", "exact"} {
		index := i + 1
		steps[index].Obs[key] = false
		bad, e := checker.Check(context.Background(), steps)
		if e != nil || bad.OK() {
			t.Fatalf("negative %s accepted: %v", key, e)
		}
		artifact(fmt.Sprintf("exploration-negative-%s.txt", key), []byte(bad.Summary()))
		steps[index].Obs[key] = true
	}
}
