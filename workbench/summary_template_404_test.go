package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func summaryTemplate404(signal string) ExplorationView {
	v := summaryView400(signal)
	v.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "none", Profile: "ordinary-summary"}
	v.Queries[0].Limit = 3
	if signal == "spans" {
		v.Queries[0].Summary = "span-duration"
		v.Queries[0].SpanStatus = true
		v.Composition.Children[0].Summary.Display.Fields = []string{"status.error_fraction", "duration.p95", "count", "duration.mean"}
	}
	return v
}
func TestSummaryTemplate404StoreAndExactDestination(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, reverse := range []bool{false, true} {
			source := summaryTemplate404(signal)
			if reverse {
				source.Composition.Children[0], source.Composition.Children[1] = source.Composition.Children[1], source.Composition.Children[0]
			}
			store := Store{Dir: t.TempDir()}
			saved, err := store.CreateExplorationView(source)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || !reflect.DeepEqual(loaded, saved) {
				t.Fatal("stored template changed", err)
			}
			q := source.Queries[0]
			q.Environment = ""
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}}
			q.FromMS = 80000
			q.ToMS = 140000
			q.Limit = 7
			q.Where = &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "雪"}, {Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "missing"}}}}}
			beforeSource, _ := json.Marshal(saved)
			beforeQuery, _ := json.Marshal(q)
			out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &q})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(out.View.Queries[0], q) || !reflect.DeepEqual(out.View.Composition, saved.Composition) || !reflect.DeepEqual(out.View.Template, saved.Template) || out.View.ID != "" {
				t.Fatal("template transferred or changed destination/presentation")
			}
			out.View.Queries[0].Where.All[0].Value = "changed"
			out.View.Composition.Children[0].Hidden = true
			out.View.Template.Profile = "changed"
			afterSource, _ := json.Marshal(saved)
			afterQuery, _ := json.Marshal(q)
			if string(beforeSource) != string(afterSource) || string(beforeQuery) != string(afterQuery) {
				t.Fatal("expansion aliases source or destination")
			}
		}
	}
}
func TestSummaryTemplate404RefusesUnsupportedProfilesAndCarriers(t *testing.T) {
	for name, mutate := range map[string]func(*ExplorationView){
		"group": func(v *ExplorationView) { v.Template.Group = "destination" }, "category profile": func(v *ExplorationView) { v.Template.Profile = "ordinary" }, "baseline": func(v *ExplorationView) { v.Template.Baseline = "destination" },
		"grouped": func(v *ExplorationView) { v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} }, "numeric even count": func(v *ExplorationView) { v.Queries[0].Numeric = summaryNumeric400() },
		"continuation root": func(v *ExplorationView) { v.Queries[0].Offset = 1 }, "timeline companion": func(v *ExplorationView) {
			v.Queries[0].BucketMS = 1000
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "count"}}})
		},
		"records absent": func(v *ExplorationView) { v.Composition.Children = v.Composition.Children[:1] }, "unsupported duration": func(v *ExplorationView) { v.Queries[0].Summary = "" }, "unsupported status": func(v *ExplorationView) { v.Queries[0].SpanStatus = false },
	} {
		t.Run(name, func(t *testing.T) {
			v := summaryTemplate404("spans")
			mutate(&v)
			if v.Validate() == nil {
				t.Fatal("unsupported template admitted")
			}
		})
	}
	source := summaryTemplate404("spans")
	for name, mutate := range map[string]func(*exploration.Query){"continuation": func(q *exploration.Query) { q.Offset = 1 }, "missing duration": func(q *exploration.Query) { q.Summary = "" }, "missing status": func(q *exploration.Query) { q.SpanStatus = false }, "numeric": func(q *exploration.Query) { q.Numeric = summaryNumeric400() }, "grouping": func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}} }} {
		t.Run(name, func(t *testing.T) {
			q := source.Queries[0]
			mutate(&q)
			before, _ := json.Marshal(q)
			original, _ := json.Marshal(source)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &q}); err == nil {
				t.Fatal("unsupported destination admitted")
			}
			after, _ := json.Marshal(q)
			unchanged, _ := json.Marshal(source)
			if string(before) != string(after) || string(original) != string(unchanged) {
				t.Fatal("refusal changed caller configuration")
			}
		})
	}
}
func TestAuthoredTemplates404NeverRewriteContinuationPosition(t *testing.T) {
	variants := map[string]func() ExplorationView{
		"summary":    func() ExplorationView { return summaryTemplate404("spans") },
		"category":   func() ExplorationView { return authoredDBTemplate(t) },
		"timeline":   func() ExplorationView { return authoredTimelineTemplate(t) },
		"comparison": func() ExplorationView { return authoredComparisonTemplate(t, false) },
		"facet": func() ExplorationView {
			v := facetComposition(true)
			v.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
			return v
		},
		"list": func() ExplorationView {
			v := authoredDBTemplate(t)
			v.Composition.Children[1].Records.Kind.Name = "record-list"
			return v
		},
		"waterfall": func() ExplorationView {
			v := authoredDBTemplate(t)
			v.Composition.Children[1].Records.Kind.Name = "record-span-waterfall"
			return v
		},
	}
	for name, create := range variants {
		t.Run(name, func(t *testing.T) {
			source := create()
			if err := source.Validate(); err != nil {
				t.Fatal("invalid root fixture", err)
			}
			q := cloneAuthored(t, source).Queries[0]
			q.Offset = 100
			before, _ := json.Marshal(q)
			original, _ := json.Marshal(source)
			var baseline *string
			for _, child := range source.Composition.Children {
				if child.Comparison != nil {
					id := strings.Repeat("b", 64)
					baseline = &id
				}
			}
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &q, BaselineCapture: baseline}); err == nil {
				t.Fatal("continuation silently rewritten")
			}
			after, _ := json.Marshal(q)
			unchanged, _ := json.Marshal(source)
			if string(before) != string(after) || string(original) != string(unchanged) {
				t.Fatal("refusal mutated caller configuration")
			}
		})
	}
}
