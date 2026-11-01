package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func recordOnlyView302(signal, kind string) ExplorationView {
	q := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	return ExplorationView{Version: 1, Title: "Raw records", Queries: []exploration.Query{q}, Composition: &ViewComposition{Version: "goo.composition/1", Kind: CompositionKind{Name: "stack", Version: 1}, Children: []ViewCompositionChild{{Records: &RecordViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: signal}, Kind: RecordViewKind{Name: kind, Version: 1}, Display: RecordViewDisplay{Columns: []string{"utc", "service", "summary", "context"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "zone"}, {Kind: "resource", Key: "zone"}}}}}}}}
}

func TestRecordsOnlySaveAndReusePreserveUnaggregatedQuestions302(t *testing.T) {
	for _, tc := range []struct{ signal, kind string }{{"logs", "record-table"}, {"logs", "record-list"}, {"spans", "record-table"}, {"spans", "record-list"}, {"spans", "record-span-waterfall"}} {
		t.Run(tc.signal+"/"+tc.kind, func(t *testing.T) {
			for _, grouping := range [][]exploration.Field{nil, {{Kind: "column", Key: "service"}}, {{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}} {
				if tc.signal == "logs" && len(grouping) > 1 {
					grouping = []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "severity"}}
				}
				source := recordOnlyView302(tc.signal, tc.kind)
				source.Queries[0].GroupBy = grouping
				if tc.signal == "spans" {
					source.Queries[0].Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: strings.Repeat("a", 32)}
					source.Queries[0].Summary = "span-duration"
				}
				store := Store{Dir: t.TempDir()}
				saved, err := store.CreateExplorationView(source)
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := store.LoadExplorationView(saved.ID)
				if err != nil || !reflect.DeepEqual(loaded.Queries, source.Queries) || !reflect.DeepEqual(loaded.Composition, source.Composition) {
					t.Fatal("raw view changed on persistence", err)
				}
				target := source.Queries[0]
				target.Environment = "staging"
				target.FromMS = 80000
				target.ToMS = 140000
				target.Offset = 100
				target.GroupBy = nil
				beforeSource, _ := json.Marshal(source)
				beforeTarget, _ := json.Marshal(target)
				reused, err := ReuseViewComposition(*source.Composition, source.Queries[0], target)
				if err != nil || !reflect.DeepEqual(reused, *source.Composition) {
					t.Fatal("raw reuse imposed grouping or changed display", err)
				}
				reused.Children[0].Records.Display.Columns[0] = "mutated"
				reused.Children[0].Records.Display.FieldColumns[0].Key = "mutated"
				afterSource, _ := json.Marshal(source)
				afterTarget, _ := json.Marshal(target)
				if string(beforeSource) != string(afterSource) || string(beforeTarget) != string(afterTarget) {
					t.Fatal("reuse aliased input")
				}
			}
		})
	}
}

func TestRecordsOnlyAdmissionAndCategoryTemplateBoundary302(t *testing.T) {
	for name, mutate := range map[string]func(*ExplorationView){
		"hidden": func(v *ExplorationView) { v.Composition.Children[0].Hidden = true },
		"metrics": func(v *ExplorationView) {
			v.Queries[0].Signal = "metrics"
			v.Composition.Children[0].Records.Scope.Signal = "metrics"
		},
		"timeline without aggregate": func(v *ExplorationView) {
			v.Queries[0].BucketMS = 10000
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "count"}}})
		},
		"comparison without aggregate": func(v *ExplorationView) {
			v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: &CountComparisonViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: CountComparisonViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: strings.Repeat("a", 64), Display: CountComparisonViewDisplay{Value: "count"}}})
		},
		"duplicate records": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[0])
		},
		"unsupported version": func(v *ExplorationView) { v.Composition.Children[0].Records.Kind.Version = 2 },
		"waterfall logs": func(v *ExplorationView) {
			v.Queries[0].Signal = "logs"
			v.Composition.Children[0].Records.Scope.Signal = "logs"
			v.Composition.Children[0].Records.Kind.Name = "record-span-waterfall"
		},
		"flat owner": func(v *ExplorationView) { v.Layout = "records" },
		"category template": func(v *ExplorationView) {
			v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
			v.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := recordOnlyView302("spans", "record-list")
			mutate(&v)
			if v.Validate() == nil {
				t.Fatal("invalid raw composition admitted")
			}
			if _, err := (Store{Dir: t.TempDir()}).CreateExplorationView(v); err == nil {
				t.Fatal("invalid raw composition published")
			}
		})
	}
	source := recordOnlyView302("logs", "record-list")
	for _, bad := range []exploration.Query{{Signal: "logs", Environment: "prod", FromMS: 2000, ToMS: 1000, Limit: 100}, {Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "not_a_column"}, Op: "eq", Value: "x"}}} {
		if source.Composition.Validate(bad) == nil {
			t.Fatal("direct raw composition bypassed query compilation")
		}
		if _, err := ReuseViewComposition(*source.Composition, source.Queries[0], bad); err == nil {
			t.Fatal("reuse accepted invalid destination")
		}
	}
}

func TestRecordsOnlyRawProfileDoesNotStripSpecializedDirectives302(t *testing.T) {
	for name, mutate := range map[string]func(*exploration.Query){
		"numeric": func(q *exploration.Query) {
			q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		},
		"numeric histogram": func(q *exploration.Query) {
			q.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
			q.NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
		},
		"reported edge": func(q *exploration.Query) {
			q.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: "api", Namespace: &identity.Namespace{Mode: "missing"}}, Child: exploration.EdgeService{Service: "worker", Namespace: &identity.Namespace{Mode: "missing"}}}
		},
		"grouped time": func(q *exploration.Query) {
			q.BucketMS = 1000
			q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
			q.GroupTimeSeries = true
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := recordOnlyView302("spans", "record-table")
			mutate(&v.Queries[0])
			if _, err := exploration.Compile(v.Queries[0]); err != nil {
				t.Fatal("fixture not otherwise valid", err)
			}
			before, _ := json.Marshal(v)
			if v.Validate() == nil {
				t.Fatal("specialized raw query accepted")
			}
			after, _ := json.Marshal(v)
			if string(before) != string(after) {
				t.Fatal("refusal stripped query directives")
			}
		})
	}
}
