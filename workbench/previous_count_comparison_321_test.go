package workbench

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func previousCountView321() ExplorationView {
	view := fixedQueryComparisonView()
	view.Queries[0].FromMS = 4000
	view.Queries[0].ToMS = 6000
	baseline := view.Composition.Children[2].Comparison.BaselineQuery
	baseline.FromMS = 2000
	baseline.ToMS = 4000
	return view
}
func TestPreviousCountViewSaveArtifactAndPlan321(t *testing.T) {
	selectors := []*identity.EnvironmentSelector{nil, {Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}, {Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}, {Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}}}
	for _, selector := range selectors {
		view := previousCountView321()
		if selector != nil {
			view.Queries[0].Environment = ""
			view.Queries[0].EnvironmentSelector = selector
			baseline := view.Composition.Children[2].Comparison.BaselineQuery
			baseline.Environment = ""
			baseline.EnvironmentSelector = selector
		}
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
			t.Fatal("adjacent pair changed on save/reopen")
		}
		artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "previous-counts", Revision: "v1", View: *loaded}
		encoded, err := EncodeViewArtifact(artifact)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeViewArtifact(encoded.YAML)
		if err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) {
			t.Fatal("adjacent pair changed on YAML roundtrip", err)
		}
		for _, question := range []exploration.Query{loaded.Queries[0], *loaded.Composition.Children[2].Comparison.BaselineQuery} {
			plan, err := exploration.Compile(question)
			if err != nil || plan.Parameters["from"] != strconv.FormatInt(question.FromMS, 10) || plan.Parameters["to"] != strconv.FormatInt(question.ToMS, 10) {
				t.Fatal("source plan lost the authored fixed window", err)
			}
		}
		decoded.Artifact.View.Composition.Children[2].Comparison.BaselineQuery.Where.Value = "changed"
		if !reflect.DeepEqual(view, before) {
			t.Fatal("loaded baseline aliases caller")
		}
		if _, err := ReuseViewComposition(*view.Composition, view.Queries[0], view.Queries[0]); err == nil {
			t.Fatal("previous pair became presentation reuse")
		}
	}
	// The first possible complete predecessor begins exactly at the epoch.
	boundary := previousCountView321()
	boundary.Queries[0].FromMS = 2000
	boundary.Queries[0].ToMS = 4000
	baseline := boundary.Composition.Children[2].Comparison.BaselineQuery
	baseline.FromMS = 0
	baseline.ToMS = 2000
	if err := boundary.Validate(); err != nil {
		t.Fatal("epoch boundary pair refused", err)
	}
}
func TestPreviousCountViewRefusesUnrelatedWindowsAndDrift321(t *testing.T) {
	mutations := map[string]func(*ExplorationView){
		"gap": func(v *ExplorationView) {
			b := v.Composition.Children[2].Comparison.BaselineQuery
			b.FromMS = 1000
			b.ToMS = 3000
		},
		"overlap": func(v *ExplorationView) {
			b := v.Composition.Children[2].Comparison.BaselineQuery
			b.FromMS = 3000
			b.ToMS = 5000
		},
		"future": func(v *ExplorationView) {
			b := v.Composition.Children[2].Comparison.BaselineQuery
			b.FromMS = 6000
			b.ToMS = 8000
		},
		"unequal": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.FromMS = 3000 },
		"underflow": func(v *ExplorationView) {
			v.Queries[0].FromMS = 1000
			v.Queries[0].ToMS = 3000
			b := v.Composition.Children[2].Comparison.BaselineQuery
			b.FromMS = -1000
			b.ToMS = 1000
		},
		"baseline-page": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Offset = 1 },
		"selected-page": func(v *ExplorationView) { v.Queries[0].Offset = 1 },
		"environment":   func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Environment = "other" },
		"profile":       func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Summary = "span-duration" },
		"bad-predicate": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.Where.Op = "unknown" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			for _, hidden := range []bool{false, true} {
				view := previousCountView321()
				view.Composition.Children[2].Hidden = hidden
				mutate(&view)
				if view.Validate() == nil {
					t.Fatal("invalid adjacent pair admitted", hidden)
				}
				if _, err := (Store{Dir: t.TempDir()}).CreateExplorationView(view); err == nil {
					t.Fatal("invalid pair persisted")
				}
			}
		})
	}
}
