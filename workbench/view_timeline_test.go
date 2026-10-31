package workbench

import (
	"encoding/json"
	"reflect"
	"testing"
)

func timelineComposition() ExplorationView {
	v := compositionView()
	v.Queries[0].BucketMS = 1000
	v.Queries[0].ToMS = 3500 // final bucket is deliberately clipped.
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{
		Version: "goo.view/1", Scope: v.Composition.Children[0].Aggregate.Scope,
		Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "count", RequiresComplete: true},
	}})
	return v
}

func TestTimelineCompositionPersistsWithoutChangingQuery(t *testing.T) {
	v := timelineComposition()
	before := cloneCompositionView(t, v).Queries
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil || !reflect.DeepEqual(loaded.Queries, before) || loaded.Composition.Children[2].Timeline.Display.Value != "count" {
		t.Fatal("timeline changed checked query", err)
	}
	for _, kind := range []string{"time-bar", "table"} {
		v.Composition.Children[2].Timeline.Kind.Name = kind
		if err := v.Validate(); err != nil {
			t.Fatal(kind, err)
		}
	}
}

func TestTimelineCompositionRejectsUnsupportedProfiles(t *testing.T) {
	for name, mutate := range map[string]func(*ExplorationView){
		"missing width":    func(v *ExplorationView) { v.Queries[0].BucketMS = 0 },
		"bad width":        func(v *ExplorationView) { v.Queries[0].BucketMS = 500 },
		"too many buckets": func(v *ExplorationView) { v.Queries[0].ToMS = 242000 },
		"split series":     func(v *ExplorationView) { v.Queries[0].GroupTimeSeries = true },
		"rate":             func(v *ExplorationView) { v.Composition.Children[2].Timeline.Display.Value = "rate" },
		"line":             func(v *ExplorationView) { v.Composition.Children[2].Timeline.Kind.Name = "line" },
		"duplicate": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[2])
		},
		"two roles":  func(v *ExplorationView) { v.Composition.Children[2].Aggregate = v.Composition.Children[0].Aggregate },
		"no records": func(v *ExplorationView) { v.Composition.Children = v.Composition.Children[1:] },
	} {
		t.Run(name, func(t *testing.T) {
			v := timelineComposition()
			mutate(&v)
			if err := v.Validate(); err == nil {
				t.Fatal("accepted unsupported timeline")
			}
		})
	}
	for _, wire := range []string{
		`{"timeline":null}`, `{"timeline":{},"aggregate":{}}`, `{"timeline":{"display":{"value":"count","scale":"population"}}}`,
	} {
		var child ViewCompositionChild
		if err := json.Unmarshal([]byte(wire), &child); err == nil {
			v := timelineComposition()
			v.Composition.Children[2] = child
			if err := v.Validate(); err == nil {
				t.Fatal("accepted invalid timeline wire", wire)
			}
		}
	}
}

func TestTimelineReuseRequiresSameExplicitWidthEvenWhenHidden(t *testing.T) {
	v := timelineComposition()
	v.Composition.Children[2].Hidden = true
	source := v.Queries[0]
	destination := source
	destination.FromMS = 2000
	destination.ToMS = 5500
	before := destination
	reused, err := ReuseViewComposition(*v.Composition, source, destination)
	if err != nil || !reflect.DeepEqual(destination, before) {
		t.Fatal(err)
	}
	reused.Children[2].Timeline.Display.Title = "Detached"
	if v.Composition.Children[2].Timeline.Display.Title != "" {
		t.Fatal("reuse aliased timeline")
	}
	for _, width := range []int64{0, 10000} {
		destination.BucketMS = width
		if _, err := ReuseViewComposition(*v.Composition, source, destination); err == nil {
			t.Fatal("accepted incompatible width", width)
		}
	}
}
