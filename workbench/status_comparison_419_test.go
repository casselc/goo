package workbench

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFreshRecordedStatusComparison419PortableRoundTrip(t *testing.T) {
	for _, value := range []string{"count", "duration.mean", "status.error", "status.error_fraction"} {
		t.Run(value, func(t *testing.T) {
			v := fixedQueryComparisonView()
			q := &v.Queries[0]
			q.SpanStatus = true
			d := v.Composition.Children[2].Comparison
			d.BaselineQuery.SpanStatus = true
			d.Display = CountComparisonViewDisplay{Value: value, View: "summary"}
			if value == "duration.mean" {
				q.Summary = "span-duration"
				d.BaselineQuery.Summary = "span-duration"
			}
			if err := v.Validate(); err != nil {
				t.Fatal(err)
			}
			raw, _ := json.Marshal(d)
			var decoded CountComparisonViewDeclaration
			if err := json.Unmarshal(raw, &decoded); err != nil || decoded.Validate(*q) != nil {
				t.Fatal("wire rejected status summary", err)
			}
			a := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "status-versions", Revision: "v1", View: v}
			encoded, err := EncodeViewArtifact(a)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := DecodeViewArtifact(encoded.YAML)
			if err != nil || !reflect.DeepEqual(loaded.Artifact, a) {
				t.Fatal("portable status intent changed", err)
			}
		})
	}
}
func TestFreshRecordedStatusComparison419RejectsUnsupportedPresentation(t *testing.T) {
	for name, mutate := range map[string]func(*ExplorationView){
		"groups": func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.View = "groups" },
		"order":  func(v *ExplorationView) { v.Composition.Children[2].Comparison.Display.Order = "change" },
		"missing directive": func(v *ExplorationView) {
			v.Queries[0].SpanStatus = false
			v.Composition.Children[2].Comparison.BaselineQuery.SpanStatus = false
		},
		"mismatched directive": func(v *ExplorationView) { v.Composition.Children[2].Comparison.BaselineQuery.SpanStatus = false },
		"captured": func(v *ExplorationView) {
			d := v.Composition.Children[2].Comparison
			d.Kind.Name = "captured-count-comparison"
			d.BaselineQuery = nil
			d.BaselineCapture = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		},
	} {
		t.Run(name, func(t *testing.T) {
			v := fixedQueryComparisonView()
			v.Queries[0].SpanStatus = true
			d := v.Composition.Children[2].Comparison
			d.BaselineQuery.SpanStatus = true
			d.Display = CountComparisonViewDisplay{Value: "status.error_fraction", View: "summary"}
			mutate(&v)
			if v.Validate() == nil {
				t.Fatal("invalid status declaration admitted")
			}
		})
	}
}
