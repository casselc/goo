package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/investigations"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"testing"
)

type investigationMachine struct {
	store     investigations.Store
	q         investigations.Question
	captures  []*investigations.Capture
	originals map[string]string
	value     int
}

func (m *investigationMachine) save(parent string) {
	c, e := m.store.Save(m.q, json.RawMessage(`{"filter":{"outcome":"error"}}`), investigations.Evidence{Operations: json.RawMessage(fmt.Sprintf(`{"basis":"%d"}`, m.value))}, parent)
	if e != nil {
		panic(e)
	}
	m.captures = append(m.captures, c)
	b, _ := json.Marshal(c)
	m.originals[c.ID] = string(b)
}
func (m *investigationMachine) RuleRefresh(tc hegel.TestCase) {
	m.value++
	m.save(m.captures[len(m.captures)-1].ID)
}
func (m *investigationMachine) RuleReopen(tc hegel.TestCase) {
	for _, c := range m.captures {
		if _, e := (investigations.Store{Dir: m.store.Dir}).Load(c.ID); e != nil {
			panic(e)
		}
	}
}
func (m *investigationMachine) reject() {
	q := m.q
	q.ToMS++
	if _, e := m.store.Save(q, nil, investigations.Evidence{Operations: json.RawMessage(`{}`)}, m.captures[0].ID); e == nil {
		panic("changed question accepted")
	}
}
func (m *investigationMachine) RuleRejectChangedQuestion(tc hegel.TestCase) { m.reject() }
func (m *investigationMachine) InvariantRetainEveryReceipt(tc hegel.TestCase) {
	for _, c := range m.captures {
		got, e := m.store.Load(c.ID)
		if e != nil {
			panic(e)
		}
		b, _ := json.Marshal(got)
		if string(b) != m.originals[c.ID] || got.Question != m.q {
			panic("receipt changed")
		}
	}
}
func newInvestigation(t *testing.T) *investigationMachine {
	return &investigationMachine{store: investigations.Store{Dir: t.TempDir()}, q: investigations.Question{Package: "checkout", Revision: "v1", FromMS: 60000, ToMS: 120000}, originals: map[string]string{}}
}
func TestHegelInvestigationStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := newInvestigation(t)
		m.save("")
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantRetainEveryReceipt"))
	}, options(100)...)
}
func TestQuintObservedInvestigation(t *testing.T) {
	m := newInvestigation(t)
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("investigation"), qobs.WithProcess("fixture"))
	for _, action := range []string{"save", "reopen", "refresh", "reject", "reopen"} {
		switch action {
		case "save":
			m.save("")
		case "refresh":
			m.value++
			m.save(m.captures[0].ID)
		case "reject":
			m.reject()
		}
		original, e := m.store.Load(m.captures[0].ID)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.Marshal(original)
		rec.Record(context.Background(), action, nil, "obs.revisions", len(m.captures), "obs.original", string(b) == m.originals[original.ID], "obs.fixed", m.captures[len(m.captures)-1].Question == m.q)
	}
	checker, e := validate.NewChecker("model/investigations.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	steps := sink.Steps()
	report, e := checker.Check(context.Background(), steps)
	if e != nil || !report.OK() {
		t.Fatalf("%v %v", e, report)
	}
	artifact("investigations-validation.txt", []byte(report.Summary()))
	for _, key := range []string{"original", "fixed"} {
		steps[2].Obs[key] = false
		bad, e := checker.Check(context.Background(), steps)
		if e != nil || bad.OK() {
			t.Fatalf("invalid %s accepted: %v", key, e)
		}
		artifact("investigations-negative-"+key+".txt", []byte(bad.Summary()))
		steps[2].Obs[key] = true
	}
}
