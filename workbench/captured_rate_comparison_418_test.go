package workbench

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestCapturedRateComparisonDeclaration418RoundTrip(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		view := compositionView()
		view.Queries[0].Signal = signal
		// Source aggregate and record child declarations retain their existing scope.
		for i := range view.Composition.Children {
			child := &view.Composition.Children[i]
			if child.Aggregate != nil {
				child.Aggregate.Scope.Signal = signal
			}
			if child.Records != nil {
				child.Records.Scope.Signal = signal
			}
		}
		q := view.Queries[0]
		d := countComparison(q)
		d.Display = CountComparisonViewDisplay{Value: exploration.RecordRateKey, View: "summary"}
		view.Composition.Children = append(view.Composition.Children, ViewCompositionChild{Comparison: d})
		if err := view.Validate(); err != nil {
			t.Fatal(signal, err)
		}
		if !slices.Contains(SummaryViewFields(q), exploration.RecordRateKey) {
			t.Fatal("supplied population summary rate unavailable", signal)
		}
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
			t.Fatal("stored rate declaration changed")
		}
		artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "captured-rate", Revision: "v1", View: *loaded}
		// Captured operands remain store-local references. Selecting a rate does
		// not make that ID portable or turn held evidence into configuration.
		if _, err := EncodeViewArtifact(artifact); err == nil || !strings.Contains(err.Error(), "store-local comparison capture") {
			t.Fatal("captured rate escaped portable-reference refusal", err)
		}
		raw, _ := json.Marshal(d)
		var read CountComparisonViewDeclaration
		if json.Unmarshal(raw, &read) != nil || read.Validate(q) != nil {
			t.Fatal("rate wire refused")
		}
		d.Display.View = ""
		if d.Validate(q) != nil {
			t.Fatal("default summary convention changed")
		}
	}
}

func TestCapturedRateComparisonDeclaration418RefusesFreshAndGroups(t *testing.T) {
	q := compositionView().Queries[0]
	for _, mutate := range []func(*CountComparisonViewDeclaration){
		func(d *CountComparisonViewDeclaration) { d.Display.View = "groups" },
		func(d *CountComparisonViewDeclaration) { d.Display.Order = "change" },
		func(d *CountComparisonViewDeclaration) { d.Display.Order = "label" },
		func(d *CountComparisonViewDeclaration) {
			d.Kind.Name = "query-count-comparison"
			d.BaselineCapture = ""
			baseline := q
			d.BaselineQuery = &baseline
		},
	} {
		d := countComparison(q)
		d.Display.Value = exploration.RecordRateKey
		mutate(d)
		if d.Validate(q) == nil {
			t.Fatal("unsupported rate declaration accepted", d)
		}
	}
	// Count's fresh-query window contract is unchanged by the added captured value.
	d := countComparison(q)
	d.Kind.Name = "query-count-comparison"
	d.BaselineCapture = ""
	baseline := q
	baseline.ToMS++
	d.BaselineQuery = &baseline
	if d.Validate(q) == nil {
		t.Fatal("fresh unequal windows were broadened")
	}
}
