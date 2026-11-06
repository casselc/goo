package workbench

import (
	"encoding/json"
	"reflect"
	"testing"
)

func meanComparisonView341() ExplorationView {
	v := pairedFacetView337("spans", "named", true)
	v.Queries[0].Summary = "span-duration"
	d := v.Composition.Children[2].Comparison
	d.BaselineQuery.Summary = "span-duration"
	d.Display.Value, d.Display.View, d.Display.Order = "duration.mean", "summary", ""
	return v
}

func TestMeanComparisonDefinitionRoundTrip341(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			v := pairedFacetView337("spans", mode, true)
			v.Queries[0].Summary = "span-duration"
			d := v.Composition.Children[2].Comparison
			d.BaselineQuery.Summary = "span-duration"
			d.Display.Value, d.Display.View, d.Display.Order = "duration.mean", "summary", ""
			before := cloneCompositionView(t, v)
			store := Store{Dir: t.TempDir()}
			saved, err := store.CreateExplorationView(v)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadExplorationView(saved.ID)
			if err != nil {
				t.Fatal(err)
			}
			loaded.ID = ""
			if !reflect.DeepEqual(*loaded, before) {
				t.Fatal("save changed mean definition, categorical facet or whole operands")
			}
			artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "previous-mean", Revision: "v1", View: *loaded}
			encoded, err := EncodeViewArtifact(artifact)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeViewArtifact(encoded.YAML)
			if err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
				t.Fatal("portable mean changed", err)
			}
			raw, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			var out CountComparisonViewDeclaration
			if err = json.Unmarshal(raw, &out); err != nil || out.Validate(v.Queries[0]) != nil {
				t.Fatal("wire mean refused", err)
			}
			d.Display.View = ""
			if err = v.Validate(); err != nil {
				t.Fatal("default summary refused", err)
			}
		})
	}
}

func TestMeanComparisonAdmission341(t *testing.T) {
	mutations := map[string]func(*ExplorationView){
		"captured": func(v *ExplorationView) {
			d := v.Composition.Children[2].Comparison
			d.Kind.Name = "captured-count-comparison"
			d.BaselineQuery = nil
			d.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
		"groups":      func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.View = "groups" },
		"group order": func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.Order = "change" },
		"label order": func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.Order = "label" },
		"missing duration": func(v *ExplorationView) {
			v.Queries[0].Summary = ""
			v.Composition.Children[2].Comparison.BaselineQuery.Summary = ""
		},
		"baseline profile": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Summary = "" },
		"percentile":       func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.Value = "duration.p95" },
		"arbitrary kind":   func(v *ExplorationView) { v.Composition.Children[2].Comparison.Kind.Name = "query-mean-comparison" },
		"baseline policy": func(v *ExplorationView) {
			v.Composition.Children[2].Comparison.BaselineQuery.EnvironmentSelector.Policy = "effective"
		},
		"window": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.ToMS++ },
		"offset": func(v *ExplorationView) { v.Queries[0].Offset = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			for _, hidden := range []bool{false, true} {
				v := meanComparisonView341()
				mutate(&v)
				v.Composition.Children[2].Hidden = hidden
				if err := v.Validate(); err == nil {
					t.Fatal("invalid mean admitted", hidden)
				}
			}
		})
	}
}
