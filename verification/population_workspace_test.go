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
func (m *viewMachine) RuleBreakdown(tc hegel.TestCase) {
	q := &m.draft.Queries[0]
	q.GroupBy = []exploration.Field{{Kind: "resource", Key: "region"}}
	baseline := *q
	baseline.FromMS = 0
	baseline.ToMS = 1000
	m.draft.Breakdowns = map[string]workbench.ExplorationBreakdown{q.Signal: {Metric: "p95", ComparisonMetric: "p95", Baseline: &baseline}}
}
func (m *viewMachine) RuleClearBaseline(tc hegel.TestCase) {
	m.draft.Breakdowns = nil
}
func (m *viewMachine) RuleTail(tc hegel.TestCase) {
	if m.draft.Tail != nil {
		m.draft.Tail = nil
	} else {
		m.draft.Tail = &workbench.ExplorationTail{Signal: m.draft.Queries[0].Signal}
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

func TestQuintPopulationBreakdownLedger(t *testing.T) {
	name := os.Getenv("GOO_BREAKDOWN_LEDGER")
	if name == "" {
		t.Skip("requires browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			View                                                          workbench.ExplorationView
			BeforeSave, AfterSave, BeforeReopen, AfterReopen, ReplayReads int
			Events                                                        []struct {
				Action string
				Reads  int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterSave++
		case 1:
			l.AfterReopen++
		case 2:
			l.View.Breakdowns["spans"].Baseline.Environment = "foreign"
		case 3:
			l.ReplayReads = 2
		case 4:
			c := l.View.Breakdowns["spans"]
			c.Metric = "bogus"
			l.View.Breakdowns["spans"] = c
		}
		coherent := l.View.Validate() == nil && l.BeforeSave == l.AfterSave && l.BeforeReopen == l.AfterReopen && l.ReplayReads == 3
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
		t.Fatal("observed breakdown rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

func TestQuintPopulationLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_POPULATION_LIVE_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			View                            workbench.ExplorationView
			BeforeSave, AfterOpen, Retained int
			Gap, Baseline                   string
			Events                          []struct {
				Action string
				Reads  int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterOpen++
		case 1:
			l.View.Tail.Signal = "metrics"
		case 2:
			l.Retained = 101
		case 3:
			l.Gap = "false"
		case 4:
			l.View.Tail.Baseline = "foreign"
		}
		coherent := l.View.Validate() == nil && l.BeforeSave == l.AfterOpen && l.Retained == 100 && l.Gap == "true" && l.View.Tail.Baseline == l.Baseline
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
		t.Fatal("live ledger rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

// Replay actual entry/read/local-edit initiations; reject widened or hidden reads.
func TestQuintRecentEntryLedger(t *testing.T) {
	name := os.Getenv("GOO_RECENT_LEDGER")
	if name == "" {
		t.Skip("requires observed browser ledger")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, Filtered                                                         []exploration.Query
			View                                                                      workbench.ExplorationView
			BeforePreview, AfterPreview, BeforeSave, BeforeReopen, AfterReopen, Tails int
			Events                                                                    []struct {
				Action string
				Reads  int
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.AfterPreview++
		case 1:
			l.AfterReopen++
		case 2:
			l.Initial[1].Environment = "foreign"
		case 3:
			l.Filtered[0].Where = nil
		case 4:
			l.Tails++
		}
		coherent := len(l.Initial) == 2 && len(l.Filtered) == 2 && l.View.Validate() == nil && l.BeforePreview == l.AfterPreview && l.BeforeSave == l.BeforeReopen && l.BeforeReopen == l.AfterReopen && l.Tails == 0
		if len(l.Initial) == 2 && len(l.Filtered) == 2 {
			a, b := l.Initial[0], l.Initial[1]
			coherent = coherent && a.Signal == "spans" && b.Signal == "logs" && a.Environment == b.Environment && a.FromMS == b.FromMS && a.ToMS == b.ToMS && a.ToMS-a.FromMS == 300000 && l.Filtered[0].Where != nil && reflect.DeepEqual(l.Filtered[1].Where, b.Where)
			for _, q := range append(l.Initial, l.Filtered...) {
				_, e := exploration.Compile(q)
				coherent = coherent && e == nil
			}
		}
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
		t.Fatal("recent entry ledger rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}
