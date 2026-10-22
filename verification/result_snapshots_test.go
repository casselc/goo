package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"hegel.dev/go/hegel"
	"testing"
)

type resultPageMachine struct {
	bank     exploration.ResultSnapshots
	ids      []string
	expected map[string]string
	sequence int
}

func (m *resultPageMachine) RuleRead(tc hegel.TestCase) {
	m.sequence++
	row := fmt.Sprintf(`{"n":9007199254740993,"occurrence":%d,"timestamp_ns":"1000000001"}`, m.sequence)
	r := &exploration.Result{Question: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1, ToMS: 2, Limit: 100, Offset: (m.sequence % 11) * 100}, Rows: []json.RawMessage{json.RawMessage(row), json.RawMessage(row)}, Basis: fmt.Sprint(m.sequence)}
	id, e := m.bank.Freeze(r)
	if e != nil {
		panic(e)
	}
	m.ids = append(m.ids, id)
	m.expected[id] = row
	r.Rows[0][2] = 'z'
	r.Question.Environment = "foreign"
}
func (m *resultPageMachine) RuleSelectOldest(tc hegel.TestCase) {
	if len(m.ids) == 0 {
		return
	}
	c, e := m.bank.Load(m.ids[0])
	if e != nil {
		panic(e)
	}
	a, e := retention.SelectResult(c, []int{1, 0})
	if e != nil {
		panic(e)
	}
	b, e := retention.SelectResult(c, []int{0, 1})
	if e != nil || a.ID != b.ID || len(a.Members) != 2 || a.Members[0].Index == a.Members[1].Index {
		panic("unstable page selection")
	}
	c.Explore.Rows[0][2] = 'x'
	if a.Validate() != nil {
		panic("selection aliased page")
	}
}
func (m *resultPageMachine) RuleRejectForeign(tc hegel.TestCase) {
	if _, e := m.bank.Load("not-a-issued-token"); e == nil {
		panic("forged page accepted")
	}
}
func (m *resultPageMachine) InvariantCustody(tc hegel.TestCase) {
	for _, id := range m.ids {
		c, e := m.bank.Load(id)
		if e != nil || c.Question.Explore.Environment != "prod" || string(c.Explore.Rows[0]) != m.expected[id] {
			panic("a new read or caller mutation changed old observations")
		}
	}
}
func TestHegelResultSnapshotStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &resultPageMachine{expected: map[string]string{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(18), hegel.WithAlwaysCheckInvariants("InvariantCustody"))
	}, options(60)...)
}
