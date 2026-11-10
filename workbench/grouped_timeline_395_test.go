package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func groupedTimeline395(signal string) ExplorationView {
	q := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 4500, Limit: 3, GroupBy: []exploration.Field{{Kind: "attribute", Key: "route"}}, BucketMS: 1000, GroupTimeSeries: true}
	scope := ViewDeclarationScope{Kind: "record-query", Signal: signal}
	return ExplorationView{Version: 1, Title: "Grouped count history", Queries: []exploration.Query{q}, Composition: &ViewComposition{Version: "goo.composition/1", Kind: CompositionKind{Name: "stack", Version: 1}, Children: []ViewCompositionChild{
		{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: TimelineViewKind{Name: "small-multiples", Version: 1}, Display: TimelineViewDisplay{Profile: "grouped-count-v1", Value: "count", Query: "", Order: "source"}}},
		{Records: &RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: RecordViewKind{Name: "record-table", Version: 1}, Display: RecordViewDisplay{Columns: []string{"utc", "summary"}}}},
	}}}
}
func TestGroupedTimeline395StoreAndDestinationReuse(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, kind := range []string{"small-multiples", "table", "heatmap"} {
			v := groupedTimeline395(signal)
			d := v.Composition.Children[0].Timeline
			d.Kind.Name = kind
			d.Display.Query = "read"
			d.Display.Order = "label"
			d.Display.RequiresComplete = true
			store := Store{Dir: t.TempDir()}
			saved, e := store.CreateExplorationView(v)
			if e != nil {
				t.Fatal(e)
			}
			loaded, e := store.LoadExplorationView(saved.ID)
			if e != nil || !reflect.DeepEqual(saved, loaded) {
				t.Fatal("immutable grouped view changed", e)
			}
			q := v.Queries[0]
			destination := q
			destination.Environment = "destination"
			destination.FromMS += 10000
			destination.ToMS += 10000
			destination.Limit = 1
			destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "destination-worker"}
			before, _ := json.Marshal(destination)
			reused, e := ReuseViewComposition(*v.Composition, q, destination)
			if e != nil || !reflect.DeepEqual(reused, *v.Composition) {
				t.Fatal("compatible grouped composition reuse refused", e)
			}
			after, _ := json.Marshal(destination)
			if string(before) != string(after) {
				t.Fatal("display reuse mutated destination")
			}
			for _, change := range []func(*exploration.Query){func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "resource", Key: "route"}} }, func(q *exploration.Query) { q.BucketMS = 2000 }, func(q *exploration.Query) { q.GroupTimeSeries = false }} {
				raw, _ := json.Marshal(destination)
				var bad exploration.Query
				json.Unmarshal(raw, &bad)
				change(&bad)
				if _, e := ReuseViewComposition(*v.Composition, q, bad); e == nil {
					t.Fatal("incompatible grouped identity or width accepted")
				}
			}
		}
	}
}
func TestGroupedTimeline395ClosedProfileRefusals(t *testing.T) {
	for _, change := range []func(*ExplorationView){
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Profile = "" },
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Value = "duration.mean" },
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Kind.Name = "time-bar" },
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Order = "value-desc" },
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Query = "a\nb" },
		func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Query = strings.Repeat("é", 129) },
		func(v *ExplorationView) {
			v.Queries[0].GroupBy = append(v.Queries[0].GroupBy, exploration.Field{Kind: "attribute", Key: "extra"})
		},
		func(v *ExplorationView) {
			v.Queries[0].GroupBy[0] = exploration.Field{Kind: "column", Key: "body_pattern_v1"}
		},
		func(v *ExplorationView) {
			v.Queries[0].Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "size"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		},
		func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Aggregate: &ViewDeclaration{}})
		},
	} {
		v := groupedTimeline395("logs")
		change(&v)
		if v.Validate() == nil {
			t.Fatal("unsupported grouped profile accepted")
		}
	}
	v := groupedTimeline395("spans")
	v.Composition.Children[0].Timeline.Kind.Name = "heatmap"
	v.Queries[0].ToMS = v.Queries[0].FromMS + 33000
	if v.Validate() == nil {
		t.Fatal("heatmap above32intervals accepted")
	}
	v.Composition.Children[0].Timeline.Kind.Name = "table"
	if e := v.Validate(); e != nil {
		t.Fatal("bounded table incorrectly acquired heatmap limit", e)
	}
	// The existing unsplit profile retains its own contract.
	plain := timelineComposition()
	if e := plain.Validate(); e != nil {
		t.Fatal("ordinary timeline regression", e)
	}
}

func TestGroupedTimeline395LocalControlWireTypes(t *testing.T) {
	for _, field := range []string{"query", "order"} {
		for _, value := range []string{"null", "1", "true", "[]", "{}"} {
			var display TimelineViewDisplay
			wire := []byte(`{"value":"count","` + field + `":` + value + `}`)
			if json.Unmarshal(wire, &display) == nil {
				t.Fatalf("%s accepted non-string %s", field, value)
			}
		}
	}
	for _, wire := range []string{`{"value":"count"}`, `{"value":"count","query":"","order":""}`, `{"value":"count","profile":"grouped-count-v1","query":"雪","order":"label"}`} {
		var display TimelineViewDisplay
		if err := json.Unmarshal([]byte(wire), &display); err != nil {
			t.Fatal("valid optional strings refused", err)
		}
	}
	var display TimelineViewDisplay
	if json.Unmarshal([]byte(`{"value":"count","unexpected":true}`), &display) == nil {
		t.Fatal("unknown display field accepted")
	}
}
