package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func summaryView400(signal string) ExplorationView {
	v := recordOnlyView302(signal, "record-table")
	v.Title = "Chosen population facts"
	q := v.Queries[0]
	d := SummaryViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: SummaryViewKind{Name: "measure-summary", Version: 1}, Display: SummaryViewDisplay{Fields: []string{"count"}}}
	v.Composition.Children = append([]ViewCompositionChild{{Summary: &d}}, v.Composition.Children...)
	return v
}
func summaryNumeric400() *exploration.NumericBinding {
	return &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.value"}, Type: "int64", Unit: "{item}", Parse: "canonical-decimal"}
}
func TestSummary400SourcesStoreAndGroupingIndependentReuse(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, groups := range [][]exploration.Field{nil, {{Kind: "resource", Key: "service.name"}}, {{Kind: "resource", Key: "service.name"}, {Kind: "attribute", Key: "work.group"}}} {
			v := summaryView400(signal)
			v.Queries[0].GroupBy = groups
			v.Queries[0].Limit = 3
			v.Queries[0].Numeric = summaryNumeric400()
			v.Composition.Children[0].Summary.Display.Fields = []string{"numeric.max", "count", "numeric.mean"}
			if signal == "spans" {
				v.Queries[0].Summary = "span-duration"
				v.Queries[0].SpanStatus = true
				v.Composition.Children[0].Summary.Display.Fields = []string{"duration.p95", "count", "status.error_fraction", "numeric.mean"}
			}
			store := Store{Dir: t.TempDir()}
			saved, err := store.CreateExplorationView(v)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil || !reflect.DeepEqual(saved, loaded) {
				t.Fatal("summary changed on persistence", err)
			}
			destination := v.Queries[0]
			destination.FromMS = 80000
			destination.ToMS = 140000
			destination.Environment = "other"
			destination.GroupBy = nil
			destination.Offset = 17
			destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "雪"}
			beforeSource, _ := json.Marshal(v)
			beforeDestination, _ := json.Marshal(destination)
			reused, err := ReuseViewComposition(*v.Composition, v.Queries[0], destination)
			if err != nil || !reflect.DeepEqual(reused, *v.Composition) {
				t.Fatal("summary reuse changed intent or required grouping", err)
			}
			reused.Children[0].Summary.Display.Fields[0] = "changed"
			afterSource, _ := json.Marshal(v)
			afterDestination, _ := json.Marshal(destination)
			if string(beforeSource) != string(afterSource) || string(beforeDestination) != string(afterDestination) {
				t.Fatal("reuse aliases source or destination")
			}
		}
	}
}
func TestSummary400CompositionVisibilityAndCompanions(t *testing.T) {
	v := summaryView400("spans")
	v.Composition.Children[1].Hidden = true
	if err := v.Validate(); err != nil {
		t.Fatal("visible summary should not require visible records", err)
	}
	v.Composition.Children[0].Hidden = true
	if v.Validate() == nil {
		t.Fatal("all-hidden summary composition accepted")
	}
	v = summaryView400("spans")
	v.Queries[0].BucketMS = 1000
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "count"}}})
	if err := v.Validate(); err != nil {
		t.Fatal("summary plus unsplit timeline rejected", err)
	}
	v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
	aggregate := ViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: ViewKind{Name: "bar", Version: 1}, Display: ViewDisplay{Category: "category", Value: "count"}}
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Aggregate: &aggregate})
	if err := v.Validate(); err != nil {
		t.Fatal("summary plus aggregate/timeline rejected", err)
	}
}
func TestSummary400RefusalsAndQuantityCompatibility(t *testing.T) {
	base := summaryView400("spans")
	base.Queries[0].Numeric = summaryNumeric400()
	base.Composition.Children[0].Summary.Display.Fields = []string{"numeric.mean"}
	for name, mutate := range map[string]func(*ExplorationView){
		"empty": func(v *ExplorationView) { v.Composition.Children[0].Summary.Display.Fields = nil },
		"duplicate": func(v *ExplorationView) {
			v.Composition.Children[0].Summary.Display.Fields = []string{"count", "count"}
		},
		"too many": func(v *ExplorationView) {
			v.Composition.Children[0].Summary.Display.Fields = []string{"count", "numeric.mean", "numeric.sum", "numeric.min", "numeric.max", "duration.mean", "duration.p95"}
		},
		"unknown":              func(v *ExplorationView) { v.Composition.Children[0].Summary.Display.Fields = []string{"made.up"} },
		"unsupported duration": func(v *ExplorationView) { v.Composition.Children[0].Summary.Display.Fields = []string{"duration.mean"} },
		"numeric histogram": func(v *ExplorationView) {
			v.Queries[0].NumericHistogram = &exploration.NumericHistogramSpec{Binning: "powers-of-two"}
		},
		"grouped": func(v *ExplorationView) {
			v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
			v.Queries[0].BucketMS = 1000
			v.Queries[0].GroupTimeSeries = true
		},
		"pattern": func(v *ExplorationView) {
			v.Queries[0].Signal = "logs"
			v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}
		},
		"duration bucket": func(v *ExplorationView) {
			v.Queries[0].GroupBy = []exploration.Field{{Kind: "column", Key: "duration_bucket"}}
		},
		"missing records": func(v *ExplorationView) { v.Composition.Children = v.Composition.Children[:1] },
		"duplicate summary": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[0])
		},
		"template": func(v *ExplorationView) { v.Template = &ViewTemplateBinding{} },
	} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(base)
			var bad ExplorationView
			if err := json.Unmarshal(raw, &bad); err != nil {
				t.Fatal(err)
			}
			mutate(&bad)
			if bad.Validate() == nil {
				t.Fatal("invalid summary accepted")
			}
		})
	}
	q := base.Queries[0]
	destination := q
	b := *q.Numeric
	destination.Numeric = &b
	destination.Numeric.Unit = "other"
	if _, err := ReuseSummaryViewDeclaration(*base.Composition.Children[0].Summary, q, destination); err == nil {
		t.Fatal("numeric quantity changed")
	}
	count := *base.Composition.Children[0].Summary
	count.Display.Fields = []string{"count"}
	if _, err := ReuseSummaryViewDeclaration(count, q, destination); err != nil {
		t.Fatal("count should not inherit numeric binding compatibility", err)
	}
	destination = q
	destination.Signal = "logs"
	if _, err := ReuseSummaryViewDeclaration(count, q, destination); err == nil {
		t.Fatal("summary signal changed")
	}
}
func TestSummary400ClosedWireAndText(t *testing.T) {
	v := summaryView400("spans")
	raw, _ := json.Marshal(v.Composition.Children[0].Summary)
	for _, mutate := range []func(string) string{
		func(s string) string { return strings.Replace(s, `"fields":["count"]`, `"fields":null`, 1) },
		func(s string) string { return strings.Replace(s, `"fields":["count"]`, `"fields":["count",null]`, 1) },
		func(s string) string {
			return strings.Replace(s, `"fields":["count"]`, `"fields":["count"],"title":null`, 1)
		},
		func(s string) string { return strings.Replace(s, `"fields":["count"]`, `"Fields":["count"]`, 1) },
		func(s string) string { return strings.Replace(s, `"signal":"spans"`, `"Signal":"spans"`, 1) },
		func(s string) string {
			return strings.Replace(s, `"name":"measure-summary"`, `"name":"measure-summary","extra":true`, 1)
		},
		func(s string) string {
			return strings.Replace(s, `"fields":["count"]`, `"fields":["count"],"title":"\ud800"`, 1)
		},
		func(s string) string {
			return strings.Replace(s, `"fields":["count"]`, `"fields":["count"],"title":"\n"`, 1)
		},
	} {
		var bad SummaryViewDeclaration
		err := json.Unmarshal([]byte(mutate(string(raw))), &bad)
		if err == nil && bad.Validate(v.Queries[0]) == nil {
			t.Fatal("malformed summary wire accepted")
		}
	}
	var child ViewCompositionChild
	if err := json.Unmarshal([]byte(`{"summary":null}`), &child); err == nil {
		v.Composition.Children[0] = child
		if v.Validate() == nil {
			t.Fatal("null summary role accepted")
		}
	}
}
