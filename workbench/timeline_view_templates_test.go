package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func authoredTimelineTemplate(t *testing.T) ExplorationView {
	t.Helper()
	source := authoredDBTemplate(t)
	source.Queries[0].BucketMS = 10000
	source.Composition.Children = append(source.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{
		Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"},
		Kind:    TimelineViewKind{Name: "time-bar", Version: 1},
		Display: TimelineViewDisplay{Value: "count", Title: "Service activity over time", RequiresComplete: true},
	}})
	return source
}

func TestAuthoredTimelineTemplatePreservesDestinationAndPresentation(t *testing.T) {
	for _, measure := range []string{"count", "duration.mean", "duration.p95"} {
		for _, kind := range []string{"time-bar", "table"} {
			for _, hidden := range []bool{false, true} {
				t.Run(measure+"/"+kind+map[bool]string{false: "/visible", true: "/hidden"}[hidden], func(t *testing.T) {
					source := authoredTimelineTemplate(t)
					timeline := source.Composition.Children[2]
					timeline.Hidden, timeline.Timeline.Kind.Name, timeline.Timeline.Display.Value = hidden, kind, measure
					// A template's order and hidden roles are presentation, not query instructions.
					source.Composition.Children = []ViewCompositionChild{timeline, source.Composition.Children[1], source.Composition.Children[0]}
					source.Composition.Children[1].Records.Kind.Name = "record-list"
					store := Store{Dir: t.TempDir()}
					saved, err := store.CreateExplorationView(source)
					if err != nil {
						t.Fatal(err)
					}
					loaded, err := store.LoadExplorationView(saved.ID)
					if err != nil || !reflect.DeepEqual(loaded.Composition, source.Composition) {
						t.Fatal("timeline changed on reload", err)
					}
					destination := cloneAuthored(t, source).Queries[0]
					destination.Environment, destination.FromMS, destination.ToMS, destination.Offset = "staging", 70000, 135000, 100
					destination.Where.All[0].Value = "other-db-service"
					destination.GroupBy = []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}
					beforeSource, _ := json.Marshal(loaded)
					beforeDestination, _ := json.Marshal(destination)
					out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: loaded, Query: &destination})
					if err != nil {
						t.Fatal(err)
					}
					expected := destination
					expected.Offset = 0
					if !reflect.DeepEqual(out.View.Queries[0], expected) || !reflect.DeepEqual(out.View.Composition, source.Composition) || out.Template.ID != saved.ID || out.View.ID != "" {
						t.Fatal("template imported scope or changed presentation", out)
					}
					out.View.Composition.Children[0].Timeline.Display.Title = "mutated"
					out.View.Queries[0].Where.All[0].Value = "mutated"
					afterSource, _ := json.Marshal(loaded)
					afterDestination, _ := json.Marshal(destination)
					if string(beforeSource) != string(afterSource) || string(beforeDestination) != string(afterDestination) {
						t.Fatal("expansion aliased input")
					}
				})
			}
		}
	}
}

func TestAuthoredTimelineTemplateRejectsIncompatibleAcquisitionEvenWhenHidden(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		for name, mutate := range map[string]func(*exploration.Query){
			"missing width":             func(q *exploration.Query) { q.BucketMS = 0 },
			"different supported width": func(q *exploration.Query) { q.BucketMS = 60000 },
			"unsupported width":         func(q *exploration.Query) { q.BucketMS = 5000 },
			"over 240 intervals":        func(q *exploration.Query) { q.ToMS = q.FromMS + 241*q.BucketMS },
			"missing duration summary":  func(q *exploration.Query) { q.Summary = "" },
			"split intervals":           func(q *exploration.Query) { q.GroupTimeSeries = true },
		} {
			t.Run(name+map[bool]string{false: "/visible", true: "/hidden"}[hidden], func(t *testing.T) {
				source := authoredTimelineTemplate(t)
				source.Composition.Children[2].Hidden = hidden
				source.Composition.Children[2].Timeline.Display.Value = "duration.p95"
				destination := cloneAuthored(t, source).Queries[0]
				mutate(&destination)
				before, _ := json.Marshal(destination)
				if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &destination}); err == nil {
					t.Fatal("incompatible destination accepted")
				}
				after, _ := json.Marshal(destination)
				if string(before) != string(after) {
					t.Fatal("refusal changed destination")
				}
			})
		}
	}
	for name, mutate := range map[string]func(*ExplorationView){
		"missing source width":     func(v *ExplorationView) { v.Queries[0].BucketMS = 0 },
		"unsupported source width": func(v *ExplorationView) { v.Queries[0].BucketMS = 5000 },
		"unsupported measure":      func(v *ExplorationView) { v.Composition.Children[2].Timeline.Display.Value = "rate" },
		"wrong timeline scope":     func(v *ExplorationView) { v.Composition.Children[2].Timeline.Scope.Signal = "logs" },
		"captured comparison": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: &CountComparisonViewDeclaration{}})
		},
	} {
		t.Run(name, func(t *testing.T) {
			source := authoredTimelineTemplate(t)
			destination := source.Queries[0]
			mutate(&source)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &destination}); err == nil {
				t.Fatal("invalid source accepted")
			}
		})
	}
}
