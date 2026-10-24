package workbench

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

func TestExplorationViewBreakdownBaseline(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, GroupBy: []exploration.Field{{Kind: "resource", Key: "region"}}}
	v := ExplorationView{Version: 1, Title: "Region differences", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}}
	old, e := s.CreateExplorationView(v)
	if e != nil {
		t.Fatal(e)
	}
	baseline := q
	baseline.FromMS = 0
	baseline.ToMS = 1000
	v.Breakdowns = map[string]ExplorationBreakdown{"spans": {Metric: "p95", ComparisonMetric: "p95", Baseline: &baseline}}
	saved, e := s.CreateExplorationView(v)
	if e != nil {
		t.Fatal(e)
	}
	if saved.ID == old.ID {
		t.Fatal("comparison config not addressed")
	}
	baseline.Environment = "foreign"
	if _, e = s.CreateExplorationView(v); e == nil {
		t.Fatal("foreign baseline accepted")
	}
	again, e := s.LoadExplorationView(old.ID)
	if e != nil || len(again.Breakdowns) != 0 {
		t.Fatal("old artifact changed", e)
	}
	baseline.Environment = "prod"
	baseline.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	if _, e = s.CreateExplorationView(v); e == nil {
		t.Fatal("incompatible grouping accepted")
	}
}

func TestExplorationTailConfiguration(t *testing.T) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
	s := Store{Dir: t.TempDir()}
	v := ExplorationView{Version: 1, Title: "Watch", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"utc", "summary"}}
	old, e := s.CreateExplorationView(v)
	if e != nil {
		t.Fatal(e)
	}
	v.Tail = &ExplorationTail{Signal: "spans", Baseline: strings.Repeat("a", 64)}
	saved, e := s.CreateExplorationView(v)
	if e != nil || saved.ID == old.ID {
		t.Fatal(saved, e)
	}
	v.Tail.Signal = "logs"
	if _, e = s.CreateExplorationView(v); e == nil {
		t.Fatal("unconfigured tail accepted")
	}
	loaded, e := s.LoadExplorationView(saved.ID)
	if e != nil || loaded.Tail.Signal != "spans" {
		t.Fatal("tail config aliased", e)
	}
}

func TestExplorationViewFieldColumns(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 2}
	base := ExplorationView{Version: 1, Title: "Fields", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"summary", "utc"}}
	s := Store{Dir: t.TempDir()}
	old, err := s.CreateExplorationView(base)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(old)
	if strings.Contains(string(raw), "fieldColumns") {
		t.Fatal("legacy wire changed")
	}
	v := base
	v.FieldColumns = []exploration.Field{{Kind: "attribute", Key: "region"}, {Kind: "resource", Key: "region"}}
	saved, err := s.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadExplorationView(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.ID == old.ID || len(loaded.FieldColumns) != 2 || loaded.FieldColumns[0] != v.FieldColumns[0] || loaded.Queries[0].Limit != 2 {
		t.Fatal("layout or question lost")
	}
	v.FieldColumns = []exploration.Field{{Kind: "resource", Key: "region"}, {Kind: "attribute", Key: "region"}}
	reordered, err := s.CreateExplorationView(v)
	if err != nil || reordered.ID == saved.ID {
		t.Fatal("ordered layout identity lost")
	}
	for _, fields := range [][]exploration.Field{{{Kind: "column", Key: "duration_ns"}}, {{Kind: "attribute", Key: "x"}, {Kind: "attribute", Key: "x"}}, {{Kind: "attribute", Key: ""}}, {{Kind: "resource", Key: strings.Repeat("é", 65)}}, make([]exploration.Field, 9)} {
		v.FieldColumns = fields
		if v.Validate() == nil {
			t.Fatalf("accepted invalid fields: %#v", fields)
		}
	}
	v = base
	v.Queries = append(v.Queries, exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 2})
	v.FieldColumns = []exploration.Field{{Kind: "column", Key: "duration_ns"}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestExplorationViewCountComparisonRanking(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "attribute", Key: "region"}}}
	v := ExplorationView{Version: 1, Title: "Count changes", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Breakdowns: map[string]ExplorationBreakdown{"logs": {Metric: "count", ComparisonMetric: "count"}}}
	s := Store{Dir: t.TempDir()}
	saved, err := s.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := s.LoadExplorationView(saved.ID)
	if err != nil || loaded.Breakdowns["logs"].ComparisonMetric != "count" || loaded.Queries[0].Limit != 1 {
		t.Fatal("ranking or bounded page changed", err)
	}
}

func TestExplorationViewNumericComparisonReuse(t *testing.T) {
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Numeric: &binding}
	baseline := q
	baseline.FromMS = 1
	baseline.ToMS = 1000
	v := ExplorationView{Version: 1, Title: "Numeric changes", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Breakdowns: map[string]ExplorationBreakdown{"logs": {Metric: "numeric-mean", ComparisonMetric: "numeric-mean", Baseline: &baseline}}}
	store := Store{Dir: t.TempDir()}
	saved, e := store.CreateExplorationView(v)
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := store.LoadExplorationView(saved.ID)
	if e != nil || loaded.Breakdowns["logs"].ComparisonMetric != "numeric-mean" || !reflect.DeepEqual(loaded.Breakdowns["logs"].Baseline.Numeric, &binding) {
		t.Fatal("numeric intent lost", e)
	}
	for _, mutate := range []func(*exploration.NumericBinding){func(b *exploration.NumericBinding) { b.Unit = "bytes" }, func(b *exploration.NumericBinding) { b.Field.Kind = "resource" }, func(b *exploration.NumericBinding) { b.Field.Key = "other" }, func(b *exploration.NumericBinding) { b.Type = "uint64" }} {
		changed := binding
		mutate(&changed)
		baseline.Numeric = &changed
		if v.Validate() == nil {
			t.Fatal("accepted mismatched numeric baseline")
		}
	}
	baseline.Numeric = nil
	if v.Validate() == nil {
		t.Fatal("accepted missing numeric baseline")
	}
	baseline.Numeric = &binding
	v.Queries[0].Numeric = nil
	if v.Validate() == nil {
		t.Fatal("accepted numeric ranking without binding")
	}
}

func TestDistributionViewRetainsIntentAndHistogramBinding(t *testing.T) {
	b := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "uint64", Unit: "items", Parse: "canonical-decimal"}
	for _, group := range [][]exploration.Field{nil, {{Kind: "column", Key: "service"}}} {
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, Numeric: &b, NumericHistogram: &exploration.NumericHistogramSpec{Binning: "powers-of-two"}, GroupBy: group}
		baseline := q
		baseline.FromMS = 1
		baseline.ToMS = 1000
		v := ExplorationView{Version: 1, Title: "Distribution", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Breakdowns: map[string]ExplorationBreakdown{"spans": {Metric: "count", ComparisonMetric: "count", Distribution: true, Baseline: &baseline}}}
		store := Store{Dir: t.TempDir()}
		saved, e := store.CreateExplorationView(v)
		if e != nil {
			t.Fatal(e)
		}
		loaded, e := store.LoadExplorationView(saved.ID)
		if e != nil || !loaded.Breakdowns["spans"].Distribution || !reflect.DeepEqual(loaded.Breakdowns["spans"].Baseline, &baseline) {
			t.Fatal("distribution view meaning changed", e)
		}
		baseline.NumericHistogram = nil
		if v.Validate() == nil {
			t.Fatal("orphan distribution baseline accepted")
		}
		baseline.NumericHistogram = q.NumericHistogram
		v.Queries[0].NumericHistogram = nil
		if v.Validate() == nil {
			t.Fatal("orphan distribution request accepted")
		}
	}
}
