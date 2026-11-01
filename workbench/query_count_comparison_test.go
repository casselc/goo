package workbench

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func fixedQueryComparisonView() ExplorationView {
	v := compositionView()
	baseline := v.Queries[0]
	baseline.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "cohort"}, Op: "eq", Value: "baseline"}
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: &CountComparisonViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: CountComparisonViewKind{Name: "query-count-comparison", Version: 1}, BaselineQuery: &baseline, Display: CountComparisonViewDisplay{Value: "count", View: "groups"}}})
	return v
}
func TestFixedQueryCountComparisonSaveAndPortableRoundTrip(t *testing.T) {
	view := fixedQueryComparisonView()
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
		t.Fatal("fixed pair changed on save/load")
	}
	a := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "cohort-counts", Revision: "v1", View: *loaded}
	encoded, err := EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(encoded.YAML)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, a) {
		t.Fatal("portable fixed baseline changed")
	}
	decoded.Artifact.View.Composition.Children[2].Comparison.BaselineQuery.Where.Value = "mutated"
	if !reflect.DeepEqual(view, before) {
		t.Fatal("fixed query baseline aliases source")
	}
	if _, err = ReuseViewComposition(*view.Composition, view.Queries[0], view.Queries[0]); err == nil {
		t.Fatal("fixed pair reused even into identical destination")
	}
	view.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary", Baseline: "destination"}
	if err = view.Validate(); err == nil {
		t.Fatal("fixed query pair became a reusable template")
	}
}
func TestFixedQueryCountComparisonRejectsContextDriftAndHiddenInvalid(t *testing.T) {
	changes := map[string]func(*ExplorationView){
		"window":      func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.ToMS++ },
		"offset":      func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Offset = 1 },
		"environment": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Environment = "other" },
		"group": func(v *ExplorationView) {
			v.Composition.Children[2].Comparison.BaselineQuery.GroupBy = []exploration.Field{{Kind: "attribute", Key: "service"}}
		},
		"profile":       func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Summary = "span-duration" },
		"bad predicate": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Where.Op = "invented" },
		"null query":    func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery = nil },
		"mixed operand": func(v *ExplorationView) {
			v.Composition.Children[2].Comparison.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			for _, hidden := range []bool{false, true} {
				v := fixedQueryComparisonView()
				change(&v)
				v.Composition.Children[2].Hidden = hidden
				if err := v.Validate(); err == nil {
					t.Fatal("invalid fixed pair admitted", hidden)
				}
			}
		})
	}
	v := fixedQueryComparisonView()
	v.Composition.Children[2].Comparison.BaselineQuery.Limit = 100
	if err := v.Validate(); err != nil {
		t.Fatal("independent valid delivery limit refused", err)
	}
}
func TestCountComparisonWireOperandDiscriminator(t *testing.T) {
	v := fixedQueryComparisonView()
	data, _ := json.Marshal(v.Composition.Children[2].Comparison)
	var base map[string]json.RawMessage
	json.Unmarshal(data, &base)
	cases := map[string]func(map[string]json.RawMessage){
		"mixed empty capture": func(m map[string]json.RawMessage) { m["baselineCapture"] = json.RawMessage(`""`) },
		"mixed null capture":  func(m map[string]json.RawMessage) { m["baselineCapture"] = json.RawMessage(`null`) },
		"null query":          func(m map[string]json.RawMessage) { m["baselineQuery"] = json.RawMessage(`null`) },
		"missing query":       func(m map[string]json.RawMessage) { delete(m, "baselineQuery") },
		"unknown":             func(m map[string]json.RawMessage) { m["rows"] = json.RawMessage(`[]`) },
		"captured with query": func(m map[string]json.RawMessage) {
			m["kind"] = json.RawMessage(`{"name":"captured-count-comparison","version":1}`)
			m["baselineCapture"] = json.RawMessage(`"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"`)
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			m := map[string]json.RawMessage{}
			for k, value := range base {
				m[k] = value
			}
			change(m)
			raw, _ := json.Marshal(m)
			var out CountComparisonViewDeclaration
			if err := json.Unmarshal(raw, &out); err == nil {
				t.Fatal("mixed/unknown operand wire admitted")
			}
		})
	}
}
