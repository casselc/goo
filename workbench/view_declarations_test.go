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

func TestAuthoredViewMeasuresFollowCheckedQuery(t *testing.T) {
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "db.response.returned_rows"}, Type: "int64", Unit: "{row}", Parse: "canonical-decimal"}
	for _, signal := range []string{"spans", "logs"} {
		for _, measure := range []string{"numeric.mean", "numeric.sum", "numeric.min", "numeric.max"} {
			t.Run(signal+"/"+measure, func(t *testing.T) {
				v := authoredView()
				v.Queries[0].Signal = signal
				v.Queries[0].Numeric = &binding
				d := v.Displays["spans"]
				d.Scope.Signal, d.Display.Value = signal, measure
				v.Displays = map[string]ViewDeclaration{signal: d}
				assertMeasureRevision(t, v, signal, measure)
				v.Queries[0].Numeric = nil
				if v.Validate() == nil {
					t.Fatal("numeric display accepted without a numeric binding")
				}
			})
		}
	}
	for _, measure := range []string{"duration.mean", "duration.min", "duration.max", "duration.p50", "duration.p95", "duration.p99"} {
		t.Run(measure, func(t *testing.T) {
			v := authoredView()
			v.Queries[0].Summary = "span-duration"
			d := v.Displays["spans"]
			d.Display.Value = measure
			v.Displays["spans"] = d
			assertMeasureRevision(t, v, "spans", measure)
			v.Queries[0].Summary = ""
			if v.Validate() == nil {
				t.Fatal("duration display accepted without its checked summary")
			}
		})
	}
	v := authoredView()
	v.Queries[0].SpanStatus = true
	d := v.Displays["spans"]
	d.Display.Value = "status.error_fraction"
	v.Displays["spans"] = d
	if v.Validate() == nil {
		t.Fatal("out-of-profile status measure accepted")
	}
}

func assertMeasureRevision(t *testing.T, v ExplorationView, signal, measure string) {
	t.Helper()
	store := Store{Dir: t.TempDir()}
	d := v.Displays[signal]
	for _, scale := range []string{"", "max"} {
		d.Display.Scale = scale
		v.Displays[signal] = d
		for _, kind := range []string{"bar", "table"} {
			d.Kind.Name = kind
			v.Displays[signal] = d
			saved, err := store.CreateExplorationView(v)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || loaded.Displays[signal].Display.Value != measure || loaded.Displays[signal].Kind.Name != kind || loaded.Queries[0].Numeric != nil && loaded.Queries[0].Numeric.Unit != "{row}" {
				t.Fatal("measure configuration lost in immutable revision", loaded, err)
			}
		}
	}
	d.Display.Scale = "population"
	v.Displays[signal] = d
	if v.Validate() == nil {
		t.Fatal("non-count population share accepted")
	}
}
