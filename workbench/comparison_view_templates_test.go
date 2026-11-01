package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func authoredComparisonTemplate(t *testing.T, timeline bool) ExplorationView {
	t.Helper()
	source := authoredDBTemplate(t)
	source.Queries[0].SpanStatus = false
	source.Template.Baseline = "destination"
	source.Composition.Children = append(source.Composition.Children, ViewCompositionChild{Comparison: countComparison(source.Queries[0])})
	source.Composition.Children[2].Comparison.Display = CapturedCountViewDisplay{Value: "count", View: "groups", Order: "label", Title: "Change in activity"}
	if timeline {
		source.Queries[0].BucketMS = 10000
		source.Composition.Children = append(source.Composition.Children, ViewCompositionChild{Timeline: &TimelineViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: TimelineViewKind{Name: "time-bar", Version: 1}, Display: TimelineViewDisplay{Value: "duration.p95", Title: "Timing", RequiresComplete: true}}})
	}
	return source
}

func TestComparisonTemplateRequiresExplicitDestinationBaselineAndDetaches(t *testing.T) {
	for _, timeline := range []bool{false, true} {
		for _, hidden := range []bool{false, true} {
			source := authoredComparisonTemplate(t, timeline)
			source.Composition.Children[2].Hidden = hidden
			// The source ordering, visibility and data-bearing record display remain exact.
			source.Composition.Children[0], source.Composition.Children[2] = source.Composition.Children[2], source.Composition.Children[0]
			source.Composition.Children[1].Records.Kind.Name = "record-list"
			saved, err := (Store{Dir: t.TempDir()}).CreateExplorationView(source)
			if err != nil {
				t.Fatal(err)
			}
			target := cloneAuthored(t, source).Queries[0]
			target.Environment = "other-environment"
			target.FromMS = 70000
			target.ToMS = 135000
			target.Offset = 100
			target.GroupBy = []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}
			target.Where.All[0].Value = "other-service"
			baseline := strings.Repeat("b", 64)
			beforeSource, _ := json.Marshal(saved)
			beforeTarget, _ := json.Marshal(target)
			out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &target, BaselineCapture: &baseline})
			if err != nil {
				t.Fatal(err)
			}
			expected := cloneAuthored(t, source)
			expected.Composition.Children[0].Comparison.BaselineCapture = baseline
			target.Offset = 0
			if !reflect.DeepEqual(out.View.Queries[0], target) || !reflect.DeepEqual(out.View.Composition, expected.Composition) || out.Template.ID != saved.ID || out.View.ID != "" {
				t.Fatal("template changed scope or presentation", out)
			}
			// Saving pins the chosen destination ID; it grants no prepared evidence.
			store := Store{Dir: t.TempDir()}
			published, err := store.CreateExplorationView(out.View)
			if err != nil {
				t.Fatal(err)
			}
			reopened, err := store.LoadExplorationView(published.ID)
			if err != nil || reopened.Composition.Children[0].Comparison.BaselineCapture != baseline {
				t.Fatal("destination reference changed", err)
			}
			out.View.Composition.Children[0].Comparison.Display.Title = "mutated"
			out.View.Queries[0].Where.All[0].Value = "mutated"
			afterSource, _ := json.Marshal(saved)
			target.Offset = 100
			afterTarget, _ := json.Marshal(target)
			if string(beforeSource) != string(afterSource) || string(beforeTarget) != string(afterTarget) {
				t.Fatal("expansion aliased caller state")
			}
			// Each expansion supplies a reference afresh, even after one successful call.
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &target}); err == nil {
				t.Fatal("source reference inherited after earlier expansion")
			}
			own := source.Composition.Children[0].Comparison.BaselineCapture
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &target, BaselineCapture: &own}); err != nil {
				t.Fatal("explicit source-equal choice must still be allowed", err)
			}
		}
	}
}

func TestComparisonTemplateRejectsBindingAndDestinationConfusion(t *testing.T) {
	for name, mutate := range map[string]func(*ExplorationView){
		"missing binding": func(v *ExplorationView) { v.Template.Baseline = "" },
		"source binding":  func(v *ExplorationView) { v.Template.Baseline = "source" },
		"status profile":  func(v *ExplorationView) { v.Queries[0].SpanStatus = true },
		"source digest":   func(v *ExplorationView) { v.ID = strings.Repeat("d", 64) },
	} {
		t.Run(name, func(t *testing.T) {
			source := authoredComparisonTemplate(t, false)
			mutate(&source)
			q := source.Queries[0]
			id := strings.Repeat("b", 64)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &q, BaselineCapture: &id}); err == nil {
				t.Fatal("invalid source accepted")
			}
		})
	}
	source := authoredComparisonTemplate(t, true)
	for name, mutate := range map[string]func(*exploration.Query){"status": func(q *exploration.Query) { q.SpanStatus = true }, "width": func(q *exploration.Query) { q.BucketMS = 60000 }, "summary": func(q *exploration.Query) { q.Summary = "" }, "split": func(q *exploration.Query) { q.GroupTimeSeries = true }} {
		t.Run(name, func(t *testing.T) {
			q := cloneAuthored(t, source).Queries[0]
			mutate(&q)
			id := strings.Repeat("b", 64)
			if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &q, BaselineCapture: &id}); err == nil {
				t.Fatal("unsupported target accepted")
			}
		})
	}
	q := source.Queries[0]
	for _, id := range []string{"", strings.Repeat("b", 63), strings.Repeat("B", 64), "latest"} {
		if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &q, BaselineCapture: &id}); err == nil {
			t.Fatal("malformed reference accepted", id)
		}
	}
	plain := authoredDBTemplate(t)
	q = plain.Queries[0]
	id := strings.Repeat("b", 64)
	if _, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &plain, Query: &q, BaselineCapture: &id}); err == nil {
		t.Fatal("unused baseline accepted")
	}
	plain.Template.Baseline = "destination"
	if plain.Validate() == nil {
		t.Fatal("binding without a comparison accepted")
	}
	shipped := dbTemplateInput(t)
	shipped.BaselineCapture = &id
	if _, err := ExpandViewTemplate(shipped); err == nil {
		t.Fatal("shipped baseline input accepted")
	}
}

func TestComparisonTemplateWireRejectsNullEmptyAndExtraneousBindings(t *testing.T) {
	source := authoredComparisonTemplate(t, false)
	q := source.Queries[0]
	id := strings.Repeat("b", 64)
	raw, _ := json.Marshal(ViewTemplateExpansionRequest{Source: &source, Query: &q, BaselineCapture: &id})
	var in ViewTemplateExpansionRequest
	if err := json.Unmarshal(raw, &in); err != nil {
		t.Fatal(err)
	}
	for _, replacement := range []string{`null`, `""`, `"latest"`, `12`, `{}`} {
		bad := strings.Replace(string(raw), `"baselineCapture":"`+id+`"`, `"baselineCapture":`+replacement, 1)
		if json.Unmarshal([]byte(bad), &in) == nil {
			t.Fatal("bad baseline wire accepted", replacement)
		}
	}
	for _, value := range []string{`null`, `""`, `"source"`, `12`} {
		bad := strings.Replace(string(raw), `"baseline":"destination"`, `"baseline":`+value, 1)
		if json.Unmarshal([]byte(bad), &in) == nil {
			t.Fatal("bad binding wire accepted", value)
		}
	}
	for _, bad := range []string{strings.Replace(string(raw), `"baseline":"destination"`, `"Baseline":"destination"`, 1), string(raw[:len(raw)-1]) + `,"widget":null}`, string(raw[:len(raw)-1]) + `,"template":""}`} {
		if json.Unmarshal([]byte(bad), &in) == nil {
			t.Fatal("mixed arm accepted")
		}
	}
	plain := authoredDBTemplate(t)
	base, _ := json.Marshal(plain)
	for _, value := range []string{`null`, `""`, `"source"`} {
		bad := strings.Replace(string(base), `"group":"destination"`, `"group":"destination","baseline":`+value, 1)
		var v ExplorationView
		if json.Unmarshal([]byte(bad), &v) == nil {
			t.Fatal("ordinary binding presence accepted", value)
		}
	}
}
