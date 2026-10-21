package workbench

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"os"
	"path/filepath"
	"testing"
)

func TestExplorationViewImmutableConfiguration(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	p, e := exploration.PlanPopulation(exploration.PopulationRequest{Question: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}})
	if e != nil {
		t.Fatal(e)
	}
	v := ExplorationView{Version: 1, Title: "Across services", Queries: p.Queries, Layout: "charts-records", Columns: []string{"utc", "service", "summary", "context"}}
	a, e := s.CreateExplorationView(v)
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.CreateExplorationView(v)
	if e != nil || a.ID != b.ID {
		t.Fatal("unstable view", e)
	}
	v.Columns = []string{"utc", "summary"}
	c, e := s.CreateExplorationView(v)
	if e != nil || c.ID == a.ID {
		t.Fatal("layout not addressed", e)
	}
	old, e := s.LoadExplorationView(a.ID)
	if e != nil || len(old.Columns) != 4 {
		t.Fatal("old revision changed", e)
	}
	v.Queries[1].Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "severity"}, Op: "eq", Value: "ERROR"}
	if _, e = s.CreateExplorationView(v); e != nil {
		t.Fatal(e)
	}
	v.Queries[1].Environment = "foreign"
	if _, e = s.CreateExplorationView(v); e == nil {
		t.Fatal("mixed scopes")
	}
	list, e := s.ExplorationViews("")
	if e != nil || len(list.Items) != 3 {
		t.Fatal(list, e)
	}
	raw, _ := json.Marshal(a)
	var corrupt ExplorationView
	json.Unmarshal(raw, &corrupt)
	corrupt.Title = "tampered"
	raw, _ = json.Marshal(corrupt)
	os.WriteFile(filepath.Join(s.Dir, "exploration-views", a.ID+".json"), raw, 0600)
	if _, e = s.LoadExplorationView(a.ID); e == nil {
		t.Fatal("tampered view")
	}
}
