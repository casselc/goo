package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func serviceTemplateJS356(t interface{ Fatal(...any) }, input any) []byte {
	raw, _ := json.Marshal(input)
	cmd := exec.Command("node", "service_templates_356_fixture.mjs")
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.Output()
	if err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			t.Fatal("actual service template helper", err, string(failure.Stderr))
		} else {
			t.Fatal(err)
		}
		return nil
	}
	return out
}
func serviceTemplateCase356(t interface{ Fatal(...any) }, seed, mode, kind int) {
	profile := []string{"ordinary", "otel.db-client", "otel.http-server"}[seed%3]
	sourceQ, target := authoredTemplateQuery(seed, profile, "specimen-service", false), authoredTemplateQuery(seed, profile, "destination-service", true)
	target.Offset = 0
	target.EnvironmentSelector = nil
	target.Environment = ""
	if mode == 0 {
		target.Environment = "destination-prod"
	} else {
		selector := identity.EnvironmentSelector{Version: "goo/environment-selector/v1", Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "destination-prod"}}
		if mode == 1 {
			selector.Policy = "current-only"
		}
		if mode == 3 {
			selector.Scope = identity.EnvironmentScope{Mode: "empty"}
		}
		if mode == 4 {
			selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		target.EnvironmentSelector = &selector
	}
	if kind == 2 {
		sourceQ.SpanStatus = false
		target.SpanStatus = false
	}
	if kind == 1 {
		sourceQ.BucketMS = target.BucketMS
	}
	scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: target.Signal}
	composition := &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: []workbench.ViewCompositionChild{
		{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.ViewKind{Name: []string{"bar", "table"}[seed%2], Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count", Scale: "population", Title: fmt.Sprintf("Team display %d", seed), RequiresComplete: seed%2 == 0}}},
		{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: "record-table", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "service", "summary"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "release"}}}}},
	}}
	source := workbench.ExplorationView{Version: 1, Title: "Team reusable view", Queries: []exploration.Query{sourceQ}, Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: profile}, Composition: composition}
	baseline := ""
	if kind == 1 {
		composition.Children = append(composition.Children, workbench.ViewCompositionChild{Timeline: &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.TimelineViewKind{Name: "time-bar", Version: 1}, Display: workbench.TimelineViewDisplay{Value: "count", Title: "Destination intervals", RequiresComplete: true}}})
	}
	if kind == 2 {
		source.Template.Baseline = "destination"
		baseline = strings.Repeat("d", 64)
		composition.Children = append(composition.Children, workbench.ViewCompositionChild{Comparison: &workbench.CountComparisonViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.CountComparisonViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: strings.Repeat("a", 64), Display: workbench.CountComparisonViewDisplay{Value: "count", View: "summary", Order: "change", Title: "Explicit destination baseline"}}})
	}
	input := map[string]any{"source": source, "target": target}
	if baseline != "" {
		input["baseline"] = baseline
	}
	before := continuationJSON352(input)
	actual := serviceTemplateJS356(t, input)
	if actual == nil {
		return
	}
	var request workbench.ViewTemplateExpansionRequest
	if err := json.Unmarshal(actual, &request); err != nil {
		t.Fatal("actual UI request Go admission", err)
		return
	}
	if request.Query == nil || !reflect.DeepEqual(*request.Query, target) || continuationJSON352(request.Source) != continuationJSON352(source) {
		t.Fatal("helper changed destination whole question or selected presentation")
		return
	}
	result, err := workbench.ExpandViewTemplate(request)
	if err != nil {
		t.Fatal("actual UI expansion", err)
		return
	}
	if !reflect.DeepEqual(result.View.Queries, []exploration.Query{target}) {
		t.Fatal("specimen changed complete destination query")
		return
	}
	expected := composition
	if kind == 2 {
		raw, _ := json.Marshal(composition)
		var copy workbench.ViewComposition
		json.Unmarshal(raw, &copy)
		copy.Children[2].Comparison.BaselineCapture = baseline
		expected = &copy
	}
	if !reflect.DeepEqual(result.View.Composition, expected) {
		t.Fatal("presentation or explicit destination binding drift")
		return
	}
	if kind == 1 {
		bad := target
		bad.BucketMS = 0
		_, err = workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &bad})
		if err == nil {
			t.Fatal("specimen interval silently supplied")
			return
		}
	}
	if kind == 2 {
		_, err = workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &target})
		if err == nil {
			t.Fatal("specimen baseline silently supplied")
			return
		}
	}
	input["expanded"] = result
	raw := serviceTemplateJS356(t, input)
	var outcome struct {
		Complete bool
		Refusals int
	}
	if json.Unmarshal(raw, &outcome) != nil || !outcome.Complete || outcome.Refusals < 20 {
		t.Fatal("incomplete actual helper stage/refusal proof", string(raw))
		return
	}
	delete(input, "expanded")
	if continuationJSON352(input) != before {
		t.Fatal("expansion mutated caller source/scope")
		return
	}
	result.View.Queries[0].Where.All[0].Value = "mutated result"
	result.View.Composition.Children[0].Aggregate.Display.Title = "mutated result"
	if continuationJSON352(input) != before {
		t.Fatal("expansion output aliases caller inputs")
	}
}
func TestHegelServiceTemplates356(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		serviceTemplateCase356(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 4)), hegel.Draw(tc, hegel.Integers(0, 2)))
	}, options(40)...)
}
