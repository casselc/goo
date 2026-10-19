package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type savedComparisonMachine struct {
	store     workbench.Store
	a, b      *exploration.SavedCapture
	parent    string
	draft     workbench.DocumentInput
	documents map[string]workbench.Document
	receipts  map[string]workbench.ComparisonReceipt
	ids       []string
}

func (m *savedComparisonMachine) RuleCaptureComparison(tc hegel.TestCase) {
	a, b := m.a, m.b
	if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
		a, b = b, a
	}
	expected, e := exploration.CompareEvidence(a, b)
	if e != nil {
		panic(e)
	}
	r, e := m.store.CreateComparison(a.ID, b.ID, expected, func(id string) (*exploration.SavedCapture, error) {
		if id == m.a.ID {
			return m.a, nil
		}
		return m.b, nil
	})
	if e != nil {
		panic(e)
	}
	m.receipts[r.ID] = *r
	m.draft.Cards[0].Reference = r.ID
}
func (m *savedComparisonMachine) RuleFinding(tc hegel.TestCase) {
	m.draft.Cards[0].Text = fmt.Sprintf("Finding 雪 %d", hegel.Draw(tc, hegel.Integers(0, 100000)))
}
func (m *savedComparisonMachine) RuleSave(tc hegel.TestCase) {
	if m.draft.Cards[0].Reference == "" {
		return
	}
	d, e := m.store.CreateDocument(m.draft, m.parent, nil)
	if e != nil {
		panic(e)
	}
	m.documents[d.ID] = *d
	m.ids = append(m.ids, d.ID)
	m.parent = d.ID
	// Detach the mutable draft from the recorded oracle.
	raw, _ := json.Marshal(m.draft)
	m.draft = workbench.DocumentInput{}
	json.Unmarshal(raw, &m.draft)
}
func (m *savedComparisonMachine) RuleFork(tc hegel.TestCase) {
	if len(m.ids) == 0 {
		return
	}
	id := m.ids[hegel.Draw(tc, hegel.Integers(0, len(m.ids)-1))]
	raw, _ := json.Marshal(m.documents[id].DocumentInput)
	m.draft = workbench.DocumentInput{}
	json.Unmarshal(raw, &m.draft)
	m.parent = id
}
func (m *savedComparisonMachine) InvariantHistory(tc hegel.TestCase) {
	for id, want := range m.receipts {
		got, e := m.store.LoadComparison(id)
		if e != nil || !reflect.DeepEqual(*got, want) {
			panic("comparison snapshot changed")
		}
	}
	for id, want := range m.documents {
		got, e := m.store.LoadDocument(id)
		if e != nil || !reflect.DeepEqual(*got, want) {
			panic("finding revision changed")
		}
	}
}
func TestHegelSavedComparisonStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		makeCapture := func(id, count string) *exploration.SavedCapture {
			q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
			return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{ID: id, Version: 1, Title: "Stored logs", CapturedAt: "2026-10-19T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q, Count: count, State: "observed", Basis: strings.Repeat("c", 64)}}
		}
		m := &savedComparisonMachine{store: workbench.Store{Dir: t.TempDir()}, a: makeCapture(strings.Repeat("a", 64), "2"), b: makeCapture(strings.Repeat("b", 64), "4"), draft: workbench.DocumentInput{Title: "Saved comparison", Cards: []workbench.Card{{Kind: "comparison"}}}, documents: map[string]workbench.Document{}, receipts: map[string]workbench.ComparisonReceipt{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(15), hegel.WithAlwaysCheckInvariants("InvariantHistory"))
	}, options(60)...)
}
func TestQuintSavedComparisonLive(t *testing.T) {
	file := os.Getenv("GOO_SAVED_PAIR_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action           string
				Reads, Revisions int
			}
			Original, Revised, Recovered workbench.Document
			Receipt, Orphan              workbench.ComparisonReceipt
			Before, After                int
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("bad ledger")
		}
		if mutation == 0 {
			l.Events[2].Reads++
		}
		if mutation == 1 {
			l.Revised.Cards[len(l.Revised.Cards)-1].Reference = strings.Repeat("f", 64)
		}
		if mutation == 2 {
			l.Receipt.Result.State = "incompatible"
		}
		if mutation == 3 {
			l.Recovered.Cards[len(l.Recovered.Cards)-1].Text = "lost finding"
		}
		if mutation == 5 {
			l.Receipt.Result.Baseline.Receipt.ID = l.Receipt.Result.Selection.Receipt.ID
		}
		if mutation == 4 {
			l.Events[1].Revisions++
		}
		oc := l.Original.Cards[len(l.Original.Cards)-1]
		rc := l.Revised.Cards[len(l.Revised.Cards)-1]
		cc := l.Recovered.Cards[len(l.Recovered.Cards)-1]
		coherent := l.Before == l.After && oc.Reference == l.Receipt.ID && rc.Reference == l.Receipt.ID && cc.Reference == l.Orphan.ID && oc.Text == "Latency increased; investigate the dependency." && rc.Text == "Latency increased; causality remains unverified." && cc.Text == rc.Text && l.Revised.Parent == l.Original.ID && l.Receipt.Result.State == "comparable" && l.Receipt.Result.CountDifference == "2" && l.Receipt.Result.Baseline.Receipt.ID == l.Original.Cards[0].Evidence[0] && l.Receipt.Result.Selection.Receipt.ID == l.Original.Cards[0].Evidence[1] && l.Orphan.Result.CountDifference == "2" && l.Orphan.Result.Baseline.Receipt.ID == l.Receipt.Result.Baseline.Receipt.ID && l.Orphan.Result.Selection.Receipt.ID == l.Receipt.Result.Selection.Receipt.ID
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.revisions", event.Revisions, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/saved_comparison.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
