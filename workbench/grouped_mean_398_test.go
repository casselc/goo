package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func groupedMean398() ExplorationView {
	v := groupedTimeline395("spans")
	v.Title = "Mean span duration by route"
	v.Queries[0].Summary = "span-duration"
	v.Composition.Children[0].Timeline.Display.Profile = "grouped-duration-mean-v1"
	v.Composition.Children[0].Timeline.Display.Value = "duration.mean"
	return v
}
func TestGroupedMean398StoreAndReuse(t *testing.T) {
	for _, kind := range []string{"small-multiples", "table", "heatmap"} {
		v := groupedMean398()
		v.Composition.Children[0].Timeline.Kind.Name = kind
		store := Store{Dir: t.TempDir()}
		saved, e := store.CreateExplorationView(v)
		if e != nil {
			t.Fatal(e)
		}
		loaded, e := store.LoadExplorationView(saved.ID)
		if e != nil || !reflect.DeepEqual(saved, loaded) {
			t.Fatal("saved mean changed", e)
		}
		q := v.Queries[0]
		destination := q
		destination.Environment = "other"
		destination.FromMS += 10000
		destination.ToMS += 10000
		destination.Limit = 7
		destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "other"}
		before, _ := json.Marshal(destination)
		reused, e := ReuseViewComposition(*v.Composition, q, destination)
		after, _ := json.Marshal(destination)
		if e != nil || !reflect.DeepEqual(reused, *v.Composition) || string(before) != string(after) {
			t.Fatal("mean reuse changed destination", e)
		}
		for _, change := range []func(*exploration.Query){func(q *exploration.Query) { q.Summary = "" }, func(q *exploration.Query) { q.Signal = "logs" }, func(q *exploration.Query) { q.GroupBy = []exploration.Field{{Kind: "resource", Key: "route"}} }, func(q *exploration.Query) { q.BucketMS = 10000 }} {
			bad := destination
			change(&bad)
			if _, e := ReuseViewComposition(*v.Composition, q, bad); e == nil {
				t.Fatal("incompatible grouped mean accepted")
			}
		}
	}
}
func TestGroupedMean398BoundedProfile(t *testing.T) {
	for _, change := range []func(*ExplorationView){func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Value = "duration.p95" }, func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Value = "count" }, func(v *ExplorationView) { v.Composition.Children[0].Timeline.Display.Profile = "grouped-count-v1" }, func(v *ExplorationView) { v.Queries[0].SpanStatus = true }, func(v *ExplorationView) {
		v.Queries[0].GroupBy = append(v.Queries[0].GroupBy, exploration.Field{Kind: "attribute", Key: "extra"})
	}, func(v *ExplorationView) {
		v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Aggregate: &ViewDeclaration{}})
	}} {
		v := groupedMean398()
		change(&v)
		if v.Validate() == nil {
			t.Fatal("unsupported grouped mean accepted")
		}
	}
	v := groupedMean398()
	v.Composition.Children[0].Timeline.Kind.Name = "heatmap"
	v.Queries[0].ToMS = v.Queries[0].FromMS + 33000
	if v.Validate() == nil {
		t.Fatal("heatmap above32 admitted")
	}
	v.Composition.Children[0].Timeline.Kind.Name = "table"
	if e := v.Validate(); e != nil {
		t.Fatal("table acquired heatmap budget", e)
	}
	for _, raw := range []string{`{"value":"duration.mean","title":null}`, `{"value":"duration.mean","requiresComplete":null}`, `{"value":"duration.mean","title":"\ud800"}`} {
		var d TimelineViewDisplay
		if json.Unmarshal([]byte(raw), &d) == nil {
			t.Fatal("invalid timeline wire normalized", raw)
		}
	}
}
