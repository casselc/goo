package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func countComparison(q exploration.Query) *CapturedCountViewDeclaration {
	return &CapturedCountViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: q.Signal}, Kind: CapturedCountViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: strings.Repeat("a", 64), Display: CapturedCountViewDisplay{Value: "count"}}
}
func TestCapturedCountCompositionStoresOnlyReferenceAndReusesExactRoot(t *testing.T) {
	v := compositionView()
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: countComparison(v.Queries[0])})
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.LoadExplorationView(saved.ID)
	if err != nil || reopened.Composition.Children[2].Comparison.BaselineCapture != strings.Repeat("a", 64) {
		t.Fatal("capture reference not retained", err)
	}
	q := v.Queries[0]
	q.FromMS = 3000
	q.ToMS = 5000
	q.Offset = 100
	q.Limit = 2
	reused, err := ReuseViewComposition(*v.Composition, v.Queries[0], q)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(reused, *v.Composition) {
		t.Fatal("reuse changed definition")
	}
	reused.Children[2].Comparison.Display.Title = "New title"
	if v.Composition.Children[2].Comparison.Display.Title != "" {
		t.Fatal("reused declaration aliases source")
	}
	q.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "other"}
	if _, err := ReuseViewComposition(*v.Composition, v.Queries[0], q); err == nil {
		t.Fatal("comparison reference reused across changed root predicates")
	}
	// Timeline and comparison coexist without inventing a fifth role.
	v.Queries[0].BucketMS = 1000
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "count"}}})
	if err := v.Validate(); err != nil {
		t.Fatal("four unique roles refused", err)
	}
}
func TestCapturedCountCompositionRejectsMalformedRolesAndEligibility(t *testing.T) {
	base := compositionView()
	base.Composition.Children = append(base.Composition.Children, ViewCompositionChild{Comparison: countComparison(base.Queries[0])})
	changes := map[string]func(*ExplorationView){
		"duplicate comparison": func(v *ExplorationView) {
			v.Composition.Children = append(v.Composition.Children, v.Composition.Children[2])
		},
		"double role":   func(v *ExplorationView) { v.Composition.Children[2].Records = v.Composition.Children[1].Records },
		"bad ID":        func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineCapture = "latest" },
		"other measure": func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.Value = "duration.p95" },
		"status parent": func(v *ExplorationView) { v.Queries[0].SpanStatus = true },
		"wrong signal":  func(v *ExplorationView) { v.Composition.Children[2].Comparison.Scope.Signal = "logs" },
		"wrong kind":    func(v *ExplorationView) { v.Composition.Children[2].Comparison.Kind.Name = "compare-anything" },
		"too long title": func(v *ExplorationView) {
			v.Composition.Children[2].Comparison.Display.Title = strings.Repeat("a", 121)
		},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			wrong := cloneCompositionView(t, base)
			change(&wrong)
			if wrong.Validate() == nil {
				t.Fatal("invalid comparison declaration accepted")
			}
		})
	}
	raw, _ := json.Marshal(base)
	for _, extra := range []string{`"query":{}`, `"capture":{}`, `"completeness":{}`} {
		wrong := strings.Replace(string(raw), `"baselineCapture":`, extra+`,"baselineCapture":`, 1)
		var v ExplorationView
		if json.Unmarshal([]byte(wrong), &v) == nil {
			t.Fatal("runtime/source field accepted", extra)
		}
	}
	var v ExplorationView
	if json.Unmarshal([]byte(strings.Replace(string(raw), `"display":{"value":"count"}`, `"display":{"value":"count","title":null}`, 1)), &v) == nil {
		t.Fatal("null optional title accepted")
	}
	if json.Unmarshal([]byte(strings.Replace(string(raw), `"display":{"value":"count"}`, `"display":{"value":"count","fields":["count"]}`, 1)), &v) == nil {
		t.Fatal("second display vocabulary accepted")
	}
}

func TestCapturedCountGroupsPresentationIsBoundedAndReusable(t *testing.T) {
	v := compositionView()
	declaration := countComparison(v.Queries[0])
	declaration.Display.View = "groups"
	declaration.Display.Order = "label"
	v.Composition.Children = append(v.Composition.Children, ViewCompositionChild{Comparison: declaration})
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	reused, err := ReuseViewComposition(*v.Composition, v.Queries[0], v.Queries[0])
	if err != nil || reused.Children[2].Comparison.Display.View != "groups" || reused.Children[2].Comparison.Display.Order != "label" {
		t.Fatal("presentation lost on reuse", err)
	}
	for _, bad := range []CapturedCountViewDisplay{{Value: "count", View: "histogram"}, {Value: "count", Order: "rate"}} {
		declaration.Display = bad
		if declaration.Validate(v.Queries[0]) == nil {
			t.Fatal("unknown group display accepted")
		}
	}
	for _, raw := range []string{`{"value":"count","view":null}`, `{"value":"count","order":null}`, `{"value":"count","find":"hidden-query"}`} {
		var d CapturedCountViewDisplay
		if json.Unmarshal([]byte(raw), &d) == nil {
			t.Fatal("unknown or null display option accepted", raw)
		}
	}
	declaration.Display = CapturedCountViewDisplay{Value: "count", View: "", Order: ""}
	if err := declaration.Validate(v.Queries[0]); err != nil {
		t.Fatal("optional default convention changed", err)
	}
}
