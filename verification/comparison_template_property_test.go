package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const comparisonTemplatePublicJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';import * as T from '../ui/view-template-handoff.mjs';
const {source,target,baseline,expanded}=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone({source,target});const expected=structuredClone(source.composition);C.comparison(expected).baselineCapture=baseline;
assert.deepEqual(T.authoredRequest(source,target,baseline),{source,query:target,baselineCapture:baseline});assert.deepEqual(T.authoredResponse(expanded,source,target,baseline),expanded.view);assert.deepEqual(expanded.view.composition,expected);assert.deepEqual(expanded.view.queries,[{...target,offset:0}]);assert.equal(C.comparison(expanded.view.composition).baselineCapture,baseline);assert.equal(C.comparison(source.composition).baselineCapture,'a'.repeat(64));
for(const id of [undefined,null,'','A'.repeat(64),'bad'])assert.throws(()=>T.authoredRequest(source,target,id));
const drift=structuredClone(expanded);C.comparison(drift.view.composition).baselineCapture=baseline==='c'.repeat(64)?'d'.repeat(64):'c'.repeat(64);assert.throws(()=>T.authoredResponse(drift,source,target,baseline));C.comparison(drift.view.composition).baselineCapture=baseline;C.comparison(drift.view.composition).display.title='wrong presentation';assert.throws(()=>T.authoredResponse(drift,source,target,baseline));
const changed=structuredClone(target);changed.where.all[0].value='changed-current-scope';assert.throws(()=>T.authoredResponse(expanded,source,changed,baseline));assert.deepEqual({source,target},before);
expanded.view.queries[0].where.all[0].value='output mutation';C.comparison(expanded.view.composition).baselineCapture='f'.repeat(64);assert.deepEqual({source,target},before);`

func comparisonTemplateCase(t interface{ Fatal(...any) }, seed, mutation int) {
	profile := []string{"ordinary", "otel.db-client", "otel.http-server"}[seed%3]
	sourceQ, targetQ := authoredTemplateQuery(seed, profile, "source-service", false), authoredTemplateQuery(seed, profile, "target-service", true)
	sourceQ.SpanStatus = false
	targetQ.SpanStatus = false
	scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: sourceQ.Signal}
	comparison := workbench.ViewCompositionChild{Hidden: seed%4 == 0, Comparison: &workbench.CapturedCountViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.CapturedCountViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: strings.Repeat("a", 64), Display: workbench.CapturedCountViewDisplay{Value: "count", View: []string{"summary", "groups"}[seed%2], Order: []string{"change", "label"}[seed%2], Title: fmt.Sprintf("Recorded changes Δ %d", seed)}}}
	aggregate := workbench.ViewCompositionChild{Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.ViewKind{Name: []string{"bar", "table"}[seed%2], Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count", Scale: "population", RequiresComplete: seed%2 == 0}}}
	records := workbench.ViewCompositionChild{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: []string{"record-table", "record-list"}[seed%2], Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}}}}
	children := []workbench.ViewCompositionChild{comparison, aggregate, records}
	if seed%2 == 0 {
		sourceQ.BucketMS = 10000
		targetQ.BucketMS = 10000
		children = append(children, workbench.ViewCompositionChild{Hidden: seed%3 == 0, Timeline: &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.TimelineViewKind{Name: "time-bar", Version: 1}, Display: workbench.TimelineViewDisplay{Value: "count", RequiresComplete: true}}})
	}
	source := workbench.ExplorationView{Version: 1, Title: "Reusable recorded changes", Queries: []exploration.Query{sourceQ}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: children}, Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: profile, Baseline: "destination"}}
	baseline := fmt.Sprintf("%064x", seed+1)
	if seed%7 == 0 {
		baseline = strings.Repeat("a", 64)
	}
	input := workbench.ViewTemplateExpansionRequest{Source: &source, Query: &targetQ, BaselineCapture: &baseline}
	switch mutation {
	case 0:
		input.BaselineCapture = nil
	case 1:
		baseline = ""
	case 2:
		baseline = strings.Repeat("A", 64)
	case 3:
		source.Template.Baseline = ""
	case 4:
		source.Template.Baseline = "source"
	case 5:
		targetQ.SpanStatus = true
		if targetQ.Signal == "logs" {
			targetQ.Signal = "spans"
		}
	case 6:
		source.Queries[0].BucketMS = 10000
		targetQ.BucketMS = 1000
		source.Composition.Children = append([]workbench.ViewCompositionChild{comparison, aggregate, records}, workbench.ViewCompositionChild{Hidden: true, Timeline: &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.TimelineViewKind{Name: "table", Version: 1}, Display: workbench.TimelineViewDisplay{Value: "count"}}})
	case 7:
		source.Composition.Children[0].Comparison.Display.Value = "duration.mean"
	case 8:
		source.Composition.Children = []workbench.ViewCompositionChild{aggregate, records}
		source.Template.Baseline = ""
	case 9:
		source.Composition.Children[0].Comparison.BaselineCapture = "invalid-specimen"
	}
	before, _ := json.Marshal(input)
	out, err := workbench.ExpandViewTemplate(input)
	if (err == nil) != (mutation < 0) {
		t.Fatal("comparison template verdict", seed, mutation, err)
		return
	}
	after, _ := json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("input mutation")
		return
	}
	if err != nil {
		var parameter any
		if input.BaselineCapture != nil {
			parameter = *input.BaselineCapture
		}
		refusedJS := `import assert from 'node:assert/strict';import fs from 'node:fs';import * as T from '../ui/view-template-handoff.mjs';const {source,target,baseline}=JSON.parse(fs.readFileSync(0,'utf8'));assert.throws(()=>T.authoredRequest(source,target,baseline));`
		if e := runNativeHistogramJS(map[string]any{"source": source, "target": targetQ, "baseline": parameter}, refusedJS); e != nil {
			t.Fatal(e)
		}
		return
	}
	expected := targetQ
	expected.Offset = 0
	raw, _ := json.Marshal(source.Composition)
	var expectedComposition workbench.ViewComposition
	_ = json.Unmarshal(raw, &expectedComposition)
	for i := range expectedComposition.Children {
		if expectedComposition.Children[i].Comparison != nil {
			expectedComposition.Children[i].Comparison.BaselineCapture = baseline
		}
	}
	if !reflect.DeepEqual(out.View.Queries, []exploration.Query{expected}) || !reflect.DeepEqual(out.View.Composition, &expectedComposition) {
		t.Fatal("scope/presentation or baseline substitution drift")
		return
	}
	if err = runNativeHistogramJS(map[string]any{"source": source, "target": targetQ, "baseline": baseline, "expanded": out}, comparisonTemplatePublicJS); err != nil {
		t.Fatal(err)
		return
	}
}
func TestHegelComparisonTemplateDestinationBinding(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { comparisonTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelComparisonTemplateRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		comparisonTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 9)))
	}, options(50)...)
}
