package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"reflect"
	"testing"
)

type recipeMachine struct{ q exploration.Query }

func (m *recipeMachine) RuleRecipe(tc hegel.TestCase) {
	ids := []string{"latency", "errors", "logs"}
	id := ids[hegel.Draw(tc, hegel.Integers(0, 2))]
	before, _ := json.Marshal(m.q)
	p, e := exploration.PlanRecipe(id, m.q)
	mismatch := (id == "logs") != (m.q.Signal == "logs")
	if mismatch {
		if e == nil {
			panic("signal translated")
		}
		return
	}
	if e != nil {
		panic(e)
	}
	after, _ := json.Marshal(m.q)
	if string(before) != string(after) || !reflect.DeepEqual(m.q.Where, p.Question.Where) || m.q.FromMS != p.Question.FromMS || m.q.ToMS != p.Question.ToMS || m.q.Environment != p.Question.Environment || p.Question.Offset != 0 {
		panic("population changed")
	}
	m.q = p.Question
}
func (m *recipeMachine) RuleScope(tc hegel.TestCase) {
	m.q.Environment = []string{"prod", "staging", "雪"}[hegel.Draw(tc, hegel.Integers(0, 2))]
	m.q.FromMS = int64(hegel.Draw(tc, hegel.Integers(0, 10000)))
	m.q.ToMS = m.q.FromMS + 60000
}
func (m *recipeMachine) RuleSignal(tc hegel.TestCase) {
	m.q.Signal = []string{"spans", "logs"}[hegel.Draw(tc, hegel.Integers(0, 1))]
	m.q.Summary = ""
	m.q.SpanStatus = false
	m.q.GroupBy = nil
}
func (m *recipeMachine) RulePredicate(tc hegel.TestCase) {
	f := exploration.Field{Kind: "attribute", Key: "tenant.id"}
	m.q.Where = &exploration.Predicate{Any: []exploration.Predicate{{Field: &f, Op: "eq", Value: []string{"a", "雪", "' OR 1=1"}[hegel.Draw(tc, hegel.Integers(0, 2))]}, {Field: &f, Op: "missing"}}}
}
func (m *recipeMachine) InvariantValid(tc hegel.TestCase) {
	if e := m.q.Validate(); e != nil {
		panic(e)
	}
}
func TestHegelRecipeStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &recipeMachine{q: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 0, ToMS: 60000, Limit: 100}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantValid"))
	}, options(100)...)
}
