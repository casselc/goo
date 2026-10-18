package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type notesContextMachine struct {
	store       workbench.Store
	draft       workbench.DocumentInput
	parent      string
	revisions   map[string]workbench.Document
	ids         []string
	markers     []workbench.Deployment
	from        int64
	environment string
}

func (m *notesContextMachine) RuleNote(tc hegel.TestCase) {
	if len(m.draft.Cards) < 6 {
		m.draft.Cards = append(m.draft.Cards, workbench.Card{Kind: "note", Role: "finding", Text: fmt.Sprintf("%d 雪\n<literal>", hegel.Draw(tc, hegel.Integers(0, 100000)))})
	}
}
func (m *notesContextMachine) RuleOrder(tc hegel.TestCase) {
	if len(m.draft.Cards) > 1 {
		a := hegel.Draw(tc, hegel.Integers(0, len(m.draft.Cards)-1))
		b := hegel.Draw(tc, hegel.Integers(0, len(m.draft.Cards)-1))
		m.draft.Cards[a], m.draft.Cards[b] = m.draft.Cards[b], m.draft.Cards[a]
	}
}
func (m *notesContextMachine) RuleSave(tc hegel.TestCase) {
	d, e := m.store.CreateDocument(m.draft, m.parent, nil)
	if e != nil {
		panic(e)
	}
	if d.Parent != m.parent || !reflect.DeepEqual(d.Cards, m.draft.Cards) {
		panic("changed ordered draft")
	}
	m.revisions[d.ID] = *d
	m.ids = append(m.ids, d.ID)
	m.parent = d.ID
}
func (m *notesContextMachine) RuleFork(tc hegel.TestCase) {
	if len(m.ids) == 0 {
		return
	}
	id := m.ids[hegel.Draw(tc, hegel.Integers(0, len(m.ids)-1))]
	d := m.revisions[id]
	m.draft = workbench.DocumentInput{Title: d.Title, Cards: append([]workbench.Card(nil), d.Cards...)}
	m.parent = id
}
func (m *notesContextMachine) RuleDeploy(tc hegel.TestCase) {
	mkr, e := m.store.CreateDeployment(workbench.DeploymentInput{Environment: m.environment, Title: "Release", Revision: "v2", OccurredAtMS: int64(hegel.Draw(tc, hegel.Integers(0, 300)))})
	if e != nil {
		panic(e)
	}
	m.markers = append(m.markers, *mkr)
}
func (m *notesContextMachine) RuleScope(tc hegel.TestCase) {
	if m.environment == "prod" {
		m.environment = "staging"
	} else {
		m.environment = "prod"
	}
	m.from = int64(hegel.Draw(tc, hegel.Integers(0, 200)))
}
func (m *notesContextMachine) InvariantHistoryAndContext(tc hegel.TestCase) {
	for id, want := range m.revisions {
		got, e := m.store.LoadDocument(id)
		if e != nil || !reflect.DeepEqual(*got, want) {
			panic("narrative history was overwritten")
		}
	}
	got, e := m.store.Deployments(workbench.DeploymentScope{Environment: m.environment, FromMS: m.from, ToMS: m.from + 100})
	if e != nil {
		panic(e)
	}
	want := []workbench.Deployment{}
	for _, marker := range m.markers {
		if marker.Environment == m.environment && marker.OccurredAtMS >= m.from && marker.OccurredAtMS < m.from+100 {
			want = append(want, marker)
		}
	}
	sort.Slice(want, func(i, j int) bool {
		if want[i].OccurredAtMS == want[j].OccurredAtMS {
			return want[i].ID < want[j].ID
		}
		return want[i].OccurredAtMS < want[j].OccurredAtMS
	})
	if !reflect.DeepEqual(got.Items, want) {
		panic("context scope changed event/custody meaning")
	}
}
func TestHegelContextNotesStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &notesContextMachine{store: workbench.Store{Dir: t.TempDir()}, draft: workbench.DocumentInput{Title: "Investigation", Cards: []workbench.Card{{Kind: "note", Role: "hypothesis", Text: "Consider alternatives"}}}, revisions: map[string]workbench.Document{}, environment: "prod"}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(15), hegel.WithAlwaysCheckInvariants("InvariantHistoryAndContext"))
	}, options(60)...)
}

type contextLedger struct {
	Report struct {
		ID      string `json:"id"`
		Outcome string `json:"outcome"`
	} `json:"report"`
	Question         exploration.Query           `json:"question"`
	Result           exploration.Result          `json:"result"`
	Marker           workbench.Deployment        `json:"marker"`
	Context          workbench.DeploymentContext `json:"context"`
	Excluded         workbench.Deployment        `json:"excluded"`
	Other            workbench.Deployment        `json:"other"`
	Capture          exploration.CapturePage     `json:"capture"`
	First            workbench.Document          `json:"first"`
	Successor        workbench.Document          `json:"successor"`
	Branch           workbench.Document          `json:"branch"`
	Reopened         workbench.Document          `json:"reopened"`
	ReadsBeforeNotes int                         `json:"readsBeforeNotes"`
	SourceReads      int                         `json:"sourceReads"`
}

func TestQuintContextNotesLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_CONTEXT_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l contextLedger
	if json.Unmarshal(data, &l) != nil {
		t.Fatal("ledger")
	}
	check := func(l contextLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("context-notes"))
		observedReport := false
		for _, raw := range l.Result.Rows {
			var row struct {
				Events []map[string]string `json:"event_attributes"`
			}
			if json.Unmarshal(raw, &row) == nil {
				for _, a := range row.Events {
					if a["goo.event.name"] == "report.ready" && a["goo.subject.id"] == l.Report.ID && a["goo.subject.namespace"] == "report" && a["goo.subject.tenant"] == "demo" && a["goo.subject.outcome"] == "ready" {
						observedReport = true
					}
				}
			}
		}
		coherent := observedReport && l.Report.ID != "" && l.Report.Outcome == "ready" && l.Result.Count != "0" && l.Context.Scope.Environment == l.Question.Environment && l.Context.Scope.FromMS == l.Question.FromMS && l.Context.Scope.ToMS == l.Question.ToMS
		found := false
		for _, m := range l.Context.Items {
			coherent = coherent && m.Environment == l.Question.Environment && m.OccurredAtMS >= l.Question.FromMS && m.OccurredAtMS < l.Question.ToMS && m.ID != l.Excluded.ID && m.ID != l.Other.ID
			if m.ID == l.Marker.ID {
				found = true
			}
		}
		rec.Record(context.Background(), "observe", nil, "obs.coherent", coherent && found)
		first := l.First
		coherent = first.Parent == "" && first.Root == "" && len(first.Cards) == 3
		if len(first.Cards) == 3 {
			coherent = coherent && first.Cards[0].Kind == "capture" && first.Cards[0].Reference == l.Capture.Receipt.ID && first.Cards[1].Kind == "deployment" && first.Cards[1].Reference == l.Marker.ID && first.Cards[2].Role == "hypothesis"
		}
		rec.Record(context.Background(), "draft", nil, "obs.coherent", coherent)
		child := func(d workbench.Document, role string) bool {
			return d.Parent == first.ID && d.Root == first.ID && d.ID != first.ID && len(d.Cards) == 4 && reflect.DeepEqual(d.Cards[:3], first.Cards) && d.Cards[3].Role == role
		}
		rec.Record(context.Background(), "revise", nil, "obs.coherent", child(l.Successor, "finding"))
		rec.Record(context.Background(), "branch", nil, "obs.coherent", child(l.Branch, "next-step") && l.Branch.ID != l.Successor.ID)
		rec.Record(context.Background(), "reopen", nil, "obs.coherent", reflect.DeepEqual(l.Reopened, first) && l.SourceReads == l.ReadsBeforeNotes)
		checker, e := validate.NewChecker("model/context_notes.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK(), report.Summary()
	}
	ok, summary := check(l)
	artifact("context-notes-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 6; i++ {
		var bad contextLedger
		json.Unmarshal(data, &bad)
		switch i {
		case 0:
			for k := range bad.Context.Items {
				if bad.Context.Items[k].ID == bad.Marker.ID {
					bad.Context.Items[k].OccurredAtMS = bad.Question.ToMS
				}
			}
		case 1:
			bad.Successor.Parent = strings.Repeat("f", 64)
		case 2:
			bad.Reopened.Cards[0].Reference = strings.Repeat("f", 64)
		case 3:
			bad.SourceReads++
		case 4:
			bad.Branch.Root = bad.Successor.ID
		case 5:
			bad.Report.ID = "different_report"
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("context-notes-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative accepted", i)
		}
	}
}
