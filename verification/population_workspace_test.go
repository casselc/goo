package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"testing"
)

func TestHegelPopulationPortableSelectors(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := int64(hegel.Draw(tc, hegel.Integers(1000, 1000000)))
		width := int64(hegel.Draw(tc, hegel.Integers(1, 100000)))
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: from, ToMS: from + width, Limit: 100, Where: &exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "region"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "region"}, Op: "eq", Value: ""}}}}
		p, e := exploration.PlanPopulation(exploration.PopulationRequest{Question: q})
		if e != nil || len(p.Queries) != 2 {
			tc.Fatal(p, e)
		}
		for _, a := range p.Queries {
			if !reflect.DeepEqual(a.Where, q.Where) || a.Environment != q.Environment || a.FromMS != from || a.ToMS != from+width || a.Validate() != nil {
				tc.Fatal("widened filter")
			}
		}
	}, options(100)...)
}

type viewMachine struct {
	store   workbench.Store
	draft   workbench.ExplorationView
	history map[string]workbench.ExplorationView
}

func (m *viewMachine) RuleLayout(tc hegel.TestCase) {
	if m.draft.Layout == "records" {
		m.draft.Layout = "charts-records"
	} else {
		m.draft.Layout = "records"
	}
}
func (m *viewMachine) RuleColumns(tc hegel.TestCase) {
	if len(m.draft.Columns) == 2 {
		m.draft.Columns = []string{"utc", "service", "summary", "context"}
	} else {
		m.draft.Columns = []string{"utc", "summary"}
	}
}
func (m *viewMachine) RuleNarrow(tc hegel.TestCase) {
	q := m.draft.Queries[0]
	if q.ToMS-q.FromMS < 3 {
		return
	}
	from := q.FromMS + 1
	for i := range m.draft.Queries {
		m.draft.Queries[i].FromMS = from
	}
}
func (m *viewMachine) RuleSave(tc hegel.TestCase) {
	v, e := m.store.CreateExplorationView(m.draft)
	if e != nil {
		panic(e)
	}
	m.history[v.ID] = *v
}
func (m *viewMachine) InvariantPublishedViews(tc hegel.TestCase) {
	for id, want := range m.history {
		got, e := m.store.LoadExplorationView(id)
		if e != nil || !reflect.DeepEqual(*got, want) {
			panic("published view changed")
		}
	}
}
func TestHegelPopulationViewStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		p, e := exploration.PlanPopulation(exploration.PopulationRequest{Question: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}})
		if e != nil {
			tc.Fatal(e)
		}
		m := &viewMachine{store: workbench.Store{Dir: t.TempDir()}, draft: workbench.ExplorationView{Version: 1, Title: "Generic view", Queries: p.Queries, Layout: "charts-records", Columns: []string{"utc", "summary"}}, history: map[string]workbench.ExplorationView{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(15), hegel.WithAlwaysCheckInvariants("InvariantPublishedViews"))
	}, options(60)...)
}
func TestQuintPopulationWorkspaceLedger(t *testing.T) {
	name := os.Getenv("GOO_POPULATION_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Q                                                      exploration.Query
			Selected, Independent                                  []exploration.Query
			View                                                   workbench.ExplorationView
			BeforePreview, AfterPreview, BeforeReopen, AfterReopen int
			Events                                                 []struct {
				Action string
				Reads  int
			}
		}
		if json.Unmarshal(raw, &l) != nil {
			t.Fatal("bad ledger")
		}
		switch mutation {
		case 0:
			l.Selected[1].Environment = "foreign"
		case 1:
			l.View.Queries[1].Where = nil
		case 2:
			l.AfterPreview++
		case 3:
			l.AfterReopen++
		case 4:
			l.View.Columns = []string{"utc"}
		}
		coherent := l.View.Validate() == nil && l.BeforePreview == l.AfterPreview && l.BeforeReopen == l.AfterReopen && len(l.Selected) == 2 && len(l.Independent) == 2 && reflect.DeepEqual(l.Selected[0].Where, l.Q.Where) && reflect.DeepEqual(l.Selected[1].Where, l.Q.Where) && reflect.DeepEqual(l.View.Queries, l.Independent) && l.Selected[0].Environment == l.Selected[1].Environment && l.Selected[0].FromMS == l.Selected[1].FromMS && l.Selected[0].ToMS == l.Selected[1].ToMS && l.Selected[0].FromMS >= l.Q.FromMS && l.Selected[0].ToMS <= l.Q.ToMS
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.coherent", coherent)
		}
		c, e := validate.NewChecker("model/question_links.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed population rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
