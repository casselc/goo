package workbench

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func facetComposition(aggregate bool) ExplorationView {
	v := compositionView()
	if !aggregate {
		v.Composition.Children = v.Composition.Children[1:]
	}
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Facet: &FacetViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: FacetViewKind{Name: "categorical-facet", Version: 1}, Field: exploration.Field{Kind: "attribute", Key: "http.route"}, Display: FacetViewDisplay{Basis: "applied", Title: "Route"}}})
	return v
}
func TestFacetCompositionSaveArtifactAndReuse310(t *testing.T) {
	for _, aggregate := range []bool{false, true} {
		view := facetComposition(aggregate)
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
		if !reflect.DeepEqual(*loaded, view) {
			t.Fatal("facet save changed root scope")
		}
		encoded, err := EncodeViewArtifact(ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "routes", Revision: "v1", View: *loaded})
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeViewArtifact(encoded.YAML)
		if err != nil || !reflect.DeepEqual(decoded.Artifact.View, view) {
			t.Fatal("portable facet changed", err)
		}
		destination := view.Queries[0]
		destination.Environment = "staging"
		destination.FromMS = 61000
		destination.ToMS = 121000
		destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: "payments"}
		reused, err := ReuseViewComposition(*view.Composition, view.Queries[0], destination)
		if err != nil {
			t.Fatal(err)
		}
		facet := reused.Children[len(reused.Children)-1].Facet
		if !reflect.DeepEqual(*facet, *view.Composition.Children[len(view.Composition.Children)-1].Facet) {
			t.Fatal("reuse changed explicit facet intent")
		}
		facet.Field.Key = "mutated"
		facet.Display.Title = "mutated"
		if !reflect.DeepEqual(view, before) {
			t.Fatal("facet reuse aliases source")
		}
	}
}
func TestFacetCompositionProfilesVisibilityAndTemplate310(t *testing.T) {
	v := timelineComposition()
	v.Queries[0].Summary = "span-duration"
	v.Queries[0].SpanStatus = true
	v.Composition.Children = append(v.Composition.Children, facetComposition(true).Composition.Children[2])
	if err := v.Validate(); err != nil {
		t.Fatal("existing summary/status/time projections blocked count facet", err)
	}
	for name, change := range map[string]func(*ExplorationView){
		"numeric field": func(v *ExplorationView) {
			v.Composition.Children[2].Facet.Field = exploration.Field{Kind: "column", Key: "duration_ns"}
		},
		"unknown field": func(v *ExplorationView) {
			v.Composition.Children[2].Facet.Field = exploration.Field{Kind: "column", Key: "invented"}
		},
		"wrong signal": func(v *ExplorationView) { v.Composition.Children[2].Facet.Scope.Signal = "logs" },
		"own basis":    func(v *ExplorationView) { v.Composition.Children[2].Facet.Display.Basis = "without-own-filter" },
		"numeric root": func(v *ExplorationView) {
			v.Queries[0].Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "1", Parse: "canonical-decimal"}
		},
		"split root": func(v *ExplorationView) { v.Queries[0].BucketMS = 1000; v.Queries[0].GroupTimeSeries = true },
		"duplicate facet": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[2])
		},
		"facet only visible": func(v *ExplorationView) {
			v.Composition.Children[0].Hidden = true
			v.Composition.Children[1].Hidden = true
		},
	} {
		t.Run(name, func(t *testing.T) {
			for _, hidden := range []bool{false, true} {
				v := facetComposition(true)
				change(&v)
				v.Composition.Children[2].Hidden = hidden
				if err := v.Validate(); err == nil {
					t.Fatal("invalid facet admitted", hidden)
				}
			}
		})
	}
	raw := facetComposition(false)
	raw.Composition.Children[0].Hidden = true
	if err := raw.Validate(); err == nil {
		t.Fatal("facet cannot replace visible records")
	}
	for _, capture := range []bool{false, true} {
		v := fixedQueryComparisonView()
		comparison := v.Composition.Children[2].Comparison
		if capture {
			comparison.Kind.Name = "captured-count-comparison"
			comparison.BaselineQuery = nil
			comparison.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}
		v.Composition.Children = append(v.Composition.Children, facetComposition(true).Composition.Children[2])
		err := v.Validate()
		if capture && err == nil {
			t.Fatal("facet accompanied an immutable captured comparison")
		}
		if !capture && err != nil {
			t.Fatal("facet refused a fresh query comparison", err)
		}
	}
}
func TestFacetWireRejectsPrivateStateAndMixedRoles310(t *testing.T) {
	base := facetComposition(true)
	raw, _ := json.Marshal(base)
	var original map[string]any
	json.Unmarshal(raw, &original)
	for name, change := range map[string]func(map[string]any){
		"null facet":      func(c map[string]any) { c["facet"] = nil },
		"mixed role":      func(c map[string]any) { c["records"] = nil },
		"private query":   func(c map[string]any) { c["facet"].(map[string]any)["query"] = map[string]any{} },
		"selected values": func(c map[string]any) { c["facet"].(map[string]any)["values"] = []any{"x"} },
		"null title":      func(c map[string]any) { c["facet"].(map[string]any)["display"].(map[string]any)["title"] = nil },
	} {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			json.Unmarshal(raw, &value)
			child := value["composition"].(map[string]any)["children"].([]any)[2].(map[string]any)
			change(child)
			data, _ := json.Marshal(value)
			var view ExplorationView
			err := json.Unmarshal(data, &view)
			if err == nil {
				err = view.Validate()
			}
			if err == nil {
				t.Fatal("private/mixed facet wire admitted")
			}
		})
	}
}
func TestFacetCodeManagedExample310(t *testing.T) {
	view := facetComposition(false)
	view.Title = "Service log exploration"
	q := &view.Queries[0]
	q.Signal = "logs"
	q.FromMS = 1767225600000
	q.ToMS = 1767225900000
	q.Limit = 100
	q.GroupBy = nil
	q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "http.request.method"}, Op: "eq", Value: "GET"}
	view.Composition.Children[0].Records.Scope.Signal = "logs"
	facet := view.Composition.Children[1].Facet
	facet.Scope.Signal = "logs"
	facet.Field = exploration.Field{Kind: "column", Key: "service"}
	facet.Display.Title = "Service"
	view.Composition.Children[0], view.Composition.Children[1] = view.Composition.Children[1], view.Composition.Children[0]
	a := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "service-log-facet", Revision: "v1", View: view}
	encoded, err := EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	const file = "../config/views/service-log-facet.yaml"
	if os.Getenv("GOO_UPDATE_FACET_EXAMPLE") == "1" {
		if err = os.WriteFile(file, []byte(encoded.YAML), 0644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(string(data))
	if err != nil || !reflect.DeepEqual(decoded.Artifact, a) || decoded.YAML != encoded.YAML {
		t.Fatal("checked facet example drift", err)
	}
}
