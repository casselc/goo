package workbench

import (
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestDurationTimelinePersistsQuerySupportedMeasures(t *testing.T) {
	for _, measure := range []string{"count", "duration.mean", "duration.p95"} {
		for _, kind := range []string{"time-bar", "table"} {
			t.Run(measure+"/"+kind, func(t *testing.T) {
				v := timelineComposition()
				v.Queries[0].Summary = "span-duration"
				timeline := v.Composition.Children[2].Timeline
				timeline.Kind.Name, timeline.Display.Value = kind, measure
				before := cloneCompositionView(t, v)
				store := Store{Dir: t.TempDir()}
				saved, err := store.CreateExplorationView(v)
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := store.LoadExplorationView(saved.ID)
				if err != nil || !reflect.DeepEqual(loaded.Queries, before.Queries) || !reflect.DeepEqual(loaded.Composition, before.Composition) {
					t.Fatal("duration timeline altered definition", err)
				}
				// Declaration support is query metadata, not an availability/completeness claim.
				if measure != "count" {
					found := false
					for _, field := range exploration.ViewMeasureFields(v.Queries[0]) {
						if field.Key == measure {
							found = field.Type == "rational" && field.Unit == "ns" && field.Nullable && field.Aggregation != nil && field.Aggregation.Family == "duration"
							if measure == "duration.p95" {
								found = found && field.Aggregation.Method == "quantilesExactHigh" && field.Aggregation.Quantile == "0.95"
							}
						}
					}
					if !found {
						t.Fatal("missing canonical duration metadata")
					}
				}
			})
		}
	}
}

func TestDurationTimelineRejectsUnsupportedQueryOrMeasure(t *testing.T) {
	v := timelineComposition()
	q := v.Queries[0]
	timeline := *v.Composition.Children[2].Timeline
	for _, measure := range []string{"duration.mean", "duration.p95"} {
		timeline.Display.Value = measure
		if err := timeline.Validate(q); err == nil {
			t.Fatal("accepted missing span summary", measure)
		}
		q.Summary = "span-duration"
		if err := timeline.Validate(q); err != nil {
			t.Fatal(err)
		}
		logs := q
		logs.Signal, logs.Summary = "logs", ""
		logTimeline := timeline
		logTimeline.Scope.Signal = "logs"
		if err := logTimeline.Validate(logs); err == nil {
			t.Fatal("accepted log duration", measure)
		}
		logTimeline.Display.Value = "count"
		if err := logTimeline.Validate(logs); err != nil {
			t.Fatal("count support regressed", err)
		}
		q.Summary = ""
	}
	q.Summary = "span-duration"
	for _, measure := range []string{"duration.p50", "duration.p99", "duration.min", "duration.max", "numeric.mean", "rate"} {
		timeline.Display.Value = measure
		if err := timeline.Validate(q); err == nil {
			t.Fatal("accepted outside finite timeline profile", measure)
		}
	}
}

func TestDurationTimelineReuseRequiresDestinationSummaryEvenWhenHidden(t *testing.T) {
	for _, measure := range []string{"duration.mean", "duration.p95"} {
		v := timelineComposition()
		v.Queries[0].Summary = "span-duration"
		v.Composition.Children[2].Timeline.Display.Value = measure
		v.Composition.Children[2].Hidden = true
		source := v.Queries[0]
		destination := source
		destination.FromMS, destination.ToMS = 5000, 7500
		before := destination
		reused, err := ReuseViewComposition(*v.Composition, source, destination)
		if err != nil || !reflect.DeepEqual(destination, before) {
			t.Fatal(err)
		}
		reused.Children[2].Timeline.Display.Value = "count"
		if v.Composition.Children[2].Timeline.Display.Value != measure {
			t.Fatal("reused timeline aliases source")
		}
		destination.Summary = ""
		if _, err := ReuseViewComposition(*v.Composition, source, destination); err == nil {
			t.Fatal("hidden timeline adopted unsupported query")
		}
		destination.Summary, destination.BucketMS = "span-duration", 10000
		if _, err := ReuseViewComposition(*v.Composition, source, destination); err == nil {
			t.Fatal("duration timeline changed width implicitly")
		}
	}
}
