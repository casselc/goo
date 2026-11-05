package workbench

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func pairedFacetView337(signal, mode string, previous bool) ExplorationView {
	view := fixedQueryComparisonView()
	selected := &view.Queries[0]
	baseline := view.Composition.Children[2].Comparison.BaselineQuery
	selected.Signal, baseline.Signal = signal, signal
	selected.FromMS, selected.ToMS = 60000, 120000
	baseline.FromMS, baseline.ToMS = selected.FromMS, selected.ToMS
	if previous {
		baseline.FromMS, baseline.ToMS = 0, selected.FromMS
	}
	selected.Environment, baseline.Environment = "", ""
	selected.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
	if mode == "named" {
		selected.EnvironmentSelector.Scope.Value = "prod-'雪"
		selected.EnvironmentSelector.Policy = "current-only"
	}
	baselineSelector := *selected.EnvironmentSelector
	baseline.EnvironmentSelector = &baselineSelector
	selected.Where = &exploration.Predicate{Any: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"},
		{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "selected-'雪"},
	}}
	baseline.Where = &exploration.Predicate{All: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""},
		{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "baseline-'雪"},
	}}
	view.Composition.Children[0].Aggregate.Scope.Signal = signal
	view.Composition.Children[1].Records.Scope.Signal = signal
	view.Composition.Children[2].Comparison.Scope.Signal = signal
	facet := *facetComposition(true).Composition.Children[2].Facet
	facet.Scope.Signal = signal
	view.Composition.Children = append(view.Composition.Children, ViewCompositionChild{Facet: &facet})
	return view
}

func TestFixedCountFacetDefinitionRoundTrips337(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, mode := range []string{"named", "empty", "unreported"} {
			t.Run(signal+"/"+mode, func(t *testing.T) {
				for _, previous := range []bool{false, true} {
					view := pairedFacetView337(signal, mode, previous)
					before := cloneCompositionView(t, view)
					store := Store{Dir: t.TempDir()}
					saved, err := store.CreateExplorationView(view)
					if err != nil {
						t.Fatal(err)
					}
					loaded, err := store.LoadExplorationView(saved.ID)
					if err != nil {
						t.Fatal(err)
					}
					loaded.ID = ""
					if !reflect.DeepEqual(*loaded, before) {
						t.Fatal("saved facet changed an operand definition")
					}
					encoded, err := EncodeViewArtifact(ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "paired-routes", Revision: "v1", View: *loaded})
					if err != nil {
						t.Fatal(err)
					}
					decoded, err := DecodeViewArtifact(encoded.YAML)
					if err != nil || !reflect.DeepEqual(decoded.Artifact.View, before) {
						t.Fatal("YAML changed paired facet or scope", err)
					}
					for _, q := range []exploration.Query{loaded.Queries[0], *loaded.Composition.Children[2].Comparison.BaselineQuery} {
						plan, err := exploration.Compile(q)
						if err != nil || plan.Parameters["from"] != strconv.FormatInt(q.FromMS, 10) || plan.Parameters["to"] != strconv.FormatInt(q.ToMS, 10) {
							t.Fatal("operand plan lost its fixed window", err)
						}
					}
					if !reflect.DeepEqual(view, before) {
						t.Fatal("validation or persistence mutated caller's operands")
					}
				}
			})
		}
	}
}

func TestFixedCountFacetRefusesUnadmittedOperands337(t *testing.T) {
	changes := map[string]func(*ExplorationView){
		"captured baseline": func(v *ExplorationView) {
			comparison := v.Composition.Children[2].Comparison
			comparison.Kind.Name = "captured-count-comparison"
			comparison.BaselineQuery = nil
			comparison.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"baseline predicate budget": func(v *ExplorationView) {
			leaves := make([]exploration.Predicate, 32)
			for i := range leaves {
				leaves[i] = exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "x"}
			}
			v.Composition.Children[2].Comparison.BaselineQuery.Where = &exploration.Predicate{All: leaves}
		},
		"baseline policy": func(v *ExplorationView) {
			v.Composition.Children[2].Comparison.BaselineQuery.EnvironmentSelector.Scope.Value = "other"
		},
		"baseline window": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.ToMS++ },
		"noncategorical facet": func(v *ExplorationView) {
			v.Composition.Children[3].Facet.Field = exploration.Field{Kind: "column", Key: "duration_ns"}
		},
		"facet scope": func(v *ExplorationView) { v.Composition.Children[3].Facet.Scope.Signal = "logs" },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			for _, hidden := range []bool{false, true} {
				view := pairedFacetView337("spans", "named", true)
				change(&view)
				view.Composition.Children[3].Hidden = hidden
				view.Composition.Children[2].Hidden = hidden
				before, _ := json.Marshal(view)
				if err := view.Validate(); err == nil {
					t.Fatal("invalid operand/facet admitted", hidden)
				}
				after, _ := json.Marshal(view)
				if string(after) != string(before) {
					t.Fatal("refused paired definition mutated caller")
				}
			}
		})
	}
}

func TestFixedCountFacetRetainsCheckedTimeline337(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		t.Run(signal, func(t *testing.T) {
			view := pairedFacetView337(signal, "named", true)
			view.Queries[0].BucketMS = 1000
			view.Composition.Children[2].Comparison.BaselineQuery.BucketMS = 1000
			timeline := *timelineComposition().Composition.Children[2].Timeline
			timeline.Scope.Signal = signal
			view.Composition.Children = append(view.Composition.Children, ViewCompositionChild{Timeline: &timeline})
			if err := view.Validate(); err != nil {
				t.Fatal("five distinct checked roles refused", err)
			}
			before := cloneCompositionView(t, view)
			encoded, err := EncodeViewArtifact(ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "paired-timeline-routes", Revision: "v1", View: view})
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeViewArtifact(encoded.YAML)
			if err != nil || !reflect.DeepEqual(decoded.Artifact.View, before) {
				t.Fatal("five-role YAML changed checked intent", err)
			}
			changes := map[string]func(*ExplorationView){
				"sixth duplicate child": func(v *ExplorationView) {
					v.Composition.Children = append(v.Composition.Children, v.Composition.Children[3])
				},
				"five-child duplicate role": func(v *ExplorationView) { v.Composition.Children[4] = v.Composition.Children[3] },
				"missing timeline width": func(v *ExplorationView) {
					v.Queries[0].BucketMS = 0
					v.Composition.Children[2].Comparison.BaselineQuery.BucketMS = 0
				},
				"baseline width drift":  func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.BucketMS = 2000 },
				"timeline signal drift": func(v *ExplorationView) { v.Composition.Children[4].Timeline.Scope.Signal = "metrics" },
			}
			for name, change := range changes {
				t.Run(name, func(t *testing.T) {
					for _, hidden := range []bool{false, true} {
						candidate := cloneCompositionView(t, view)
						change(&candidate)
						candidate.Composition.Children[4].Hidden = hidden
						if err := candidate.Validate(); err == nil {
							t.Fatal("invalid timeline or duplicate role admitted", hidden)
						}
					}
				})
			}
			if !reflect.DeepEqual(view, before) {
				t.Fatal("timeline admission mutated original operands")
			}
		})
	}
}
