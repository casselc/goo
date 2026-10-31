package workbench

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"strings"
	"testing"
)

func authoredView() ExplorationView {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
	return ExplorationView{Version: 1, Title: "Services", Queries: []exploration.Query{q}, Layout: "charts-records", Columns: []string{"utc", "summary"}, Displays: map[string]ViewDeclaration{"spans": {Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: ViewKind{Name: "bar", Version: 1}, Display: ViewDisplay{Category: "category", Value: "count"}}}}
}

func TestAuthoredViewConfigurationRoundTrip(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	v := authoredView()
	first, err := s.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.LoadExplorationView(first.ID)
	if err != nil || again.Displays["spans"].Kind.Name != "bar" {
		t.Fatal(again, err)
	}
	d := v.Displays["spans"]
	d.Kind.Name = "table"
	d.Display.RequiresComplete = true
	d.Display.Query = "worker"
	v.Displays["spans"] = d
	next, err := s.CreateExplorationView(v)
	if err != nil || next.ID == first.ID {
		t.Fatal("display configuration must affect identity", err)
	}
	old, err := s.LoadExplorationView(first.ID)
	if err != nil || old.Displays["spans"].Display.RequiresComplete {
		t.Fatal("old configuration changed", err)
	}
}

func TestAuthoredViewRejectsUnsupportedDeclarations(t *testing.T) {
	tests := map[string]func(*ExplorationView){
		"overlapping breakdown": func(v *ExplorationView) {
			v.Breakdowns = map[string]ExplorationBreakdown{"spans": {Metric: "count", ComparisonMetric: "count"}}
		},
		"version": func(v *ExplorationView) { d := v.Displays["spans"]; d.Version = "goo.view/2"; v.Displays["spans"] = d },
		"scope":   func(v *ExplorationView) { d := v.Displays["spans"]; d.Scope.Kind = "native"; v.Displays["spans"] = d },
		"signal":  func(v *ExplorationView) { d := v.Displays["spans"]; d.Scope.Signal = "logs"; v.Displays["spans"] = d },
		"kind":    func(v *ExplorationView) { d := v.Displays["spans"]; d.Kind.Version = 2; v.Displays["spans"] = d },
		"measure": func(v *ExplorationView) { d := v.Displays["spans"]; d.Display.Value = "p95"; v.Displays["spans"] = d },
		"query bytes": func(v *ExplorationView) {
			d := v.Displays["spans"]
			d.Display.Query = strings.Repeat("é", 129)
			v.Displays["spans"] = d
		},
		"title bytes": func(v *ExplorationView) {
			d := v.Displays["spans"]
			d.Display.Title = strings.Repeat("é", 61)
			v.Displays["spans"] = d
		},
		"control":        func(v *ExplorationView) { d := v.Displays["spans"]; d.Display.Query = "a\nb"; v.Displays["spans"] = d },
		"native profile": func(v *ExplorationView) { v.Queries[0].GroupBy[0].Key = "duration_bucket" },
		"no grouping":    func(v *ExplorationView) { v.Queries[0].GroupBy = nil },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			v := authoredView()
			mutate(&v)
			if v.Validate() == nil {
				t.Fatal("unsupported declaration accepted")
			}
		})
	}
}

func TestExplorationViewTypedEnvironmentParity(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			scope := identity.EnvironmentScope{Mode: mode}
			if mode == "named" {
				scope.Value = "prod"
			}
			selection, err := identity.EffectiveEnvironment(scope)
			if err != nil {
				t.Fatal(err)
			}
			wire, _ := selection.Selector()
			v := authoredView()
			v.Displays = nil
			v.Queries[0].Environment = ""
			v.Queries[0].EnvironmentSelector = &wire
			baseline := v.Queries[0]
			baseline.FromMS = 0
			baseline.ToMS = 1000
			v.Breakdowns = map[string]ExplorationBreakdown{"spans": {Metric: "count", ComparisonMetric: "count", Baseline: &baseline}}
			if err := v.Validate(); err != nil {
				t.Fatal(err)
			}
			mismatch := baseline
			mismatch.EnvironmentSelector = nil
			mismatch.Environment = "prod"
			v.Breakdowns["spans"] = ExplorationBreakdown{Metric: "count", ComparisonMetric: "count", Baseline: &mismatch}
			if v.Validate() == nil {
				t.Fatal("different policy/scope baseline accepted")
			}
		})
	}
}
