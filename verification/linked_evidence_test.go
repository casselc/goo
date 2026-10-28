package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestHegelLinkedEvidenceDifference(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7199999))
		input, _ := json.Marshal([]int{from, width})
		script := `const assert=require('node:assert/strict'),w=require('../ui/workspace.mjs');const [from,width]=JSON.parse(process.argv[1]);for(const step of [w.recordStep({signal:'spans',environment:'prod',fromMs:from,toMs:from+width,limit:100,offset:100,where:{field:{kind:'attribute',key:'region'},op:'eq',value:''}}),w.measurementStep({definition:{environment:'prod',source:{metric:'latency',unit:'s'}},fromMs:from,toMs:from+width})]){const c=w.captureQuestion(step);assert.deepEqual(w.evidenceDifference(step,c),[]);const q=c.explore||c.measurement;q.toMs++;assert.deepEqual(w.evidenceDifference(step,c),['Fixed window differs']);assert.equal(step.scope.toMs,from+width);if(q.definition)q.definition.source.unit='ms';else q.where.value='雪';assert(w.evidenceDifference(step,c).includes('Question definition differs'));}`
		if out, e := exec.Command("node", "-e", script, string(input)).CombinedOutput(); e != nil {
			tc.Fatalf("actual JavaScript: %s %v", out, e)
		}
	}, options(100)...)
}

type linkedMachine struct {
	store   workbench.Store
	draft   workbench.DocumentInput
	parent  string
	history map[string]workbench.Document
	known   map[string]bool
	ids     []string
}

func (m *linkedMachine) RuleLink(tc hegel.TestCase) {
	c := &m.draft.Cards[0]
	id := fmt.Sprintf("%064x", hegel.Draw(tc, hegel.Integers(1, 100000)))
	for _, old := range c.Evidence {
		if id == old {
			return
		}
	}
	if len(c.Evidence) >= 6 {
		return
	}
	m.known[id] = true
	c.Evidence = append(c.Evidence, id)
}
func (m *linkedMachine) RuleSave(tc hegel.TestCase) {
	d, e := m.store.CreateDocument(m.draft, m.parent, func(id string) error {
		if !m.known[id] {
			return fmt.Errorf("unavailable")
		}
		return nil
	})
	if e != nil {
		panic(e)
	}
	if !reflect.DeepEqual(d.DocumentInput, m.draft) || d.Parent != m.parent {
		panic("changed authored history")
	}
	m.history[d.ID] = *d
	m.ids = append(m.ids, d.ID)
	m.parent = d.ID
}
func (m *linkedMachine) RuleFork(tc hegel.TestCase) {
	if len(m.ids) == 0 {
		return
	}
	id := m.ids[hegel.Draw(tc, hegel.Integers(0, len(m.ids)-1))]
	raw, _ := json.Marshal(m.history[id].DocumentInput)
	m.draft = workbench.DocumentInput{}
	json.Unmarshal(raw, &m.draft)
	m.parent = id
}
func (m *linkedMachine) InvariantHistory(tc hegel.TestCase) {
	for id, want := range m.history {
		got, e := m.store.LoadDocument(id)
		if e != nil || !reflect.DeepEqual(*got, want) {
			panic("linked history changed")
		}
	}
}
func TestHegelLinkedEvidenceStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
		step := workbench.Step{Version: 1, Kind: "record", Scope: workbench.StepScope{Environment: "prod", FromMS: 1000, ToMS: 2000}, View: "Records", Record: &q}
		m := &linkedMachine{store: workbench.Store{Dir: t.TempDir()}, draft: workbench.DocumentInput{Title: "Linked history", Cards: []workbench.Card{{Kind: "question", Key: strings.Repeat("a", 64), Step: &step}}}, history: map[string]workbench.Document{}, known: map[string]bool{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(15), hegel.WithAlwaysCheckInvariants("InvariantHistory"))
	}, options(60)...)
}

func TestQuintLinkedEvidenceLive(t *testing.T) {
	name := os.Getenv("GOO_LINKED_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action        string
				Reads, Linked int
			}
			Original, Refined, Recovered workbench.Document
			First, Second                struct {
				Capture exploration.CapturePage
				Doc     workbench.Document
			}
			Drift, Orphan                                      exploration.CapturePage
			BeforeRetry, AfterRetry, BeforeReopen, AfterReopen int
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("bad ledger")
		}
		if mutation == 0 {
			l.Recovered.Cards[0].Evidence = l.Recovered.Cards[0].Evidence[1:]
		}
		if mutation == 1 {
			l.Recovered.Cards[0].Evidence[3] = strings.Repeat("f", 64)
		}
		if mutation == 2 {
			l.AfterReopen++
		}
		if mutation == 3 {
			l.Events[6].Reads++
		}
		if mutation == 4 {
			l.Events[5].Linked++
		}
		if mutation == 5 {
			l.Drift.Receipt.Question.Explore = l.Original.Cards[0].Step.Record
		}
		if mutation == 6 {
			l.Recovered.Cards[0].Key = strings.Repeat("b", 64)
		}
		key := l.Original.Cards[0].Key
		want := []string{l.First.Capture.Receipt.ID, l.Second.Capture.Receipt.ID, l.Drift.Receipt.ID, l.Orphan.Receipt.ID}
		a, _ := json.Marshal(l.Original.Cards[0].Step.Record.Where)
		b, _ := json.Marshal(l.Drift.Receipt.Question.Explore.Where)
		coherent := l.Original.DocumentInput.Validate() == nil && l.Recovered.DocumentInput.Validate() == nil && l.Recovered.Cards[0].Key == key && reflect.DeepEqual(l.Recovered.Cards[0].Evidence, want) && reflect.DeepEqual(l.First.Doc.Cards[0].Evidence, want[:1]) && reflect.DeepEqual(l.Second.Doc.Cards[0].Evidence, want[:2]) && reflect.DeepEqual(l.Refined.Cards[0].Evidence, want[:3]) && l.First.Doc.Parent == l.Original.ID && l.Second.Doc.Parent == l.First.Doc.ID && l.Recovered.Parent == l.Refined.ID && l.BeforeRetry == l.AfterRetry && l.BeforeReopen == l.AfterReopen && string(a) != string(b)
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.linked", event.Linked, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/linked_evidence.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed history rejected")
	}
	for i := 0; i < 7; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
