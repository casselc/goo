package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"reflect"
	"strings"
	"testing"
)

type followupMachine struct {
	capture  exploration.SavedCapture
	original string
	row      int
}

func (m *followupMachine) RuleSelect(tc hegel.TestCase) { m.row = hegel.Draw(tc, hegel.Integers(0, 2)) }
func (m *followupMachine) RulePlan(tc hegel.TestCase) {
	p, e := m.capture.Followups(m.row)
	if e != nil {
		panic(e)
	}
	q := m.capture.Question.Explore
	for _, a := range p.Actions {
		if a.Query != nil {
			next := a.Query
			if next.Environment != q.Environment || next.FromMS != q.FromMS || next.ToMS != q.ToMS || next.Offset != 0 || next.Limit != 100 {
				panic("scope/time drift")
			}
			if _, e := exploration.Compile(*next); e != nil {
				panic(e)
			}
			if strings.HasPrefix(a.Relation, "Same matching") && !reflect.DeepEqual(next.Where, q.Where) {
				panic("same population changed")
			}
			if a.ID == "trace" || a.ID == "trace-logs" {
				if next.Where.Value != p.Observations[m.row].TraceID {
					panic("invented correlation")
				}
			}
			next.Environment = "mutated returned plan"
			if next.Where != nil && next.Where.Field != nil {
				next.Where.Value = "mutated"
			}
		}
		if a.Discovery != nil {
			d := a.Discovery
			if d.Environment != q.Environment || d.FromMS != q.FromMS || d.ToMS != q.ToMS || d.Service != p.Observations[m.row].Service {
				panic("discovery scope drift")
			}
		}
	}
}
func (m *followupMachine) InvariantImmutable(tc hegel.TestCase) {
	b, _ := json.Marshal(m.capture)
	if string(b) != m.original {
		panic("stored evidence changed")
	}
}
func TestHegelFollowupStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		tenant := fmt.Sprint(hegel.Draw(tc, hegel.Integers(0, 10000)))
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Offset: 0, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: tenant}}
		c := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{ID: strings.Repeat("c", 64), Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q}}
		for i := 0; i < 3; i++ {
			trace := strings.Repeat(fmt.Sprint(i+1), 32)
			if i == 2 {
				trace = strings.Repeat("0", 32)
			}
			b, _ := json.Marshal(map[string]any{"ServiceName": "edge", "TraceId": trace, "SpanId": strings.Repeat("b", 16), "Attributes": map[string]string{"tenant": tenant, "goo.operation.outcome": "ok"}})
			c.Explore.Rows = append(c.Explore.Rows, b)
		}
		raw, _ := json.Marshal(c)
		m := &followupMachine{capture: c, original: string(raw)}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25), hegel.WithAlwaysCheckInvariants("InvariantImmutable"))
	}, options(100)...)
}
