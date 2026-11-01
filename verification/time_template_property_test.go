package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const timeTemplatePublicJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';import * as T from '../ui/view-template-handoff.mjs';
const {source,target,expanded}=JSON.parse(fs.readFileSync(0,'utf8'));const before=structuredClone({source,target});const expected=structuredClone(target);expected.offset=0;
assert.deepEqual(T.authoredRequest(source,target),{source,query:target});assert.deepEqual(T.authoredResponse(expanded,source,target),expanded.view);C.view(expanded.view);assert.deepEqual(expanded.view.queries,[expected]);assert.deepEqual(expanded.view.composition,source.composition);assert.deepEqual(C.timeline(expanded.view.composition),C.timeline(source.composition));
for(const width of [undefined,target.bucketMs===1000?10000:1000,5000]){const bad=structuredClone(target);if(width===undefined)delete bad.bucketMs;else bad.bucketMs=width;assert.throws(()=>T.authoredRequest(source,bad));}
const drift=structuredClone(expanded);C.timeline(drift.view.composition).display.title='lost original';assert.throws(()=>T.authoredResponse(drift,source,target));
const changed=structuredClone(target);changed.fromMs++;assert.throws(()=>T.authoredResponse(expanded,source,changed));assert.deepEqual({source,target},before);
expanded.view.queries[0].where.all[0].value='changed output';C.timeline(expanded.view.composition).display.title='changed output';assert.deepEqual({source,target},before);`

func timeTemplateCase(t interface{ Fatal(...any) }, seed, mutation int) {
	profile := []string{"ordinary", "otel.db-client", "otel.http-server"}[seed%3]
	if mutation >= 0 {
		profile = "otel.db-client"
	}
	sourceQ, targetQ := authoredTemplateQuery(seed, profile, "source-service", false), authoredTemplateQuery(seed, profile, "target-service", true)
	width := []int64{1000, 10000, 60000, 300000}[seed%4]
	sourceQ.BucketMS = width
	targetQ.BucketMS = width
	measure := "count"
	if sourceQ.Signal == "spans" {
		measure = []string{"count", "duration.mean", "duration.p95"}[seed%3]
	}
	scope := workbench.ViewDeclarationScope{Kind: "record-query", Signal: sourceQ.Signal}
	timeline := workbench.ViewCompositionChild{Hidden: seed%3 == 0, Timeline: &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.TimelineViewKind{Name: []string{"time-bar", "table"}[seed%2], Version: 1}, Display: workbench.TimelineViewDisplay{Value: measure, Title: fmt.Sprintf("Service time Δ %d", seed), RequiresComplete: seed%2 == 0}}}
	aggregate := workbench.ViewCompositionChild{Hidden: seed%4 == 0, Aggregate: &workbench.ViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.ViewKind{Name: "bar", Version: 1}, Display: workbench.ViewDisplay{Category: "category", Value: "count", Scale: "population", Title: "All groups"}}}
	records := workbench.ViewCompositionChild{Records: &workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: scope, Kind: workbench.RecordViewKind{Name: []string{"record-table", "record-list"}[seed%2], Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "summary"}, FieldColumns: []exploration.Field{{Kind: []string{"attribute", "resource"}[seed%2], Key: "db.namespace"}}}}}
	children := []workbench.ViewCompositionChild{timeline, aggregate, records}
	if seed%2 == 1 {
		children = []workbench.ViewCompositionChild{records, aggregate, timeline}
	}
	source := workbench.ExplorationView{Version: 1, Title: "Team service time", Queries: []exploration.Query{sourceQ}, Template: &workbench.ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: profile}, Composition: &workbench.ViewComposition{Version: "goo.composition/1", Kind: workbench.CompositionKind{Name: "stack", Version: 1}, Children: children}}
	switch mutation {
	case 0:
		targetQ.BucketMS = 0
	case 1:
		targetQ.BucketMS = 1000
		if width == 1000 {
			targetQ.BucketMS = 10000
		}
	case 2:
		targetQ.BucketMS = 5000
	case 3:
		targetQ.ToMS = targetQ.FromMS + 241*width
	case 4:
		timeline.Timeline.Display.Value = "duration.p95"
		targetQ.Summary = ""
	case 5:
		targetQ.GroupTimeSeries = true
	case 6:
		source.Composition.Children = append(source.Composition.Children, workbench.ViewCompositionChild{Comparison: &workbench.CountComparisonViewDeclaration{}})
	case 7:
		timeline.Timeline.Kind.Version = 2
	}
	input := workbench.ViewTemplateExpansionRequest{Source: &source, Query: &targetQ}
	before, _ := json.Marshal(input)
	out, err := workbench.ExpandViewTemplate(input)
	if (err == nil) != (mutation < 0) {
		t.Fatal("time template verdict", seed, mutation, err)
		return
	}
	after, _ := json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("expansion mutated input")
		return
	}
	if err != nil {
		return
	}
	expected := targetQ
	expected.Offset = 0
	if !reflect.DeepEqual(out.View.Queries, []exploration.Query{expected}) || !reflect.DeepEqual(out.View.Composition, source.Composition) {
		t.Fatal("scope or presentation drift")
		return
	}
	if err = runNativeHistogramJS(map[string]any{"source": source, "target": targetQ, "expanded": out}, timeTemplatePublicJS); err != nil {
		t.Fatal(err)
		return
	}
	out.View.Queries[0].Where.All[0].Value = "changed"
	for i := range out.View.Composition.Children {
		if out.View.Composition.Children[i].Timeline != nil {
			out.View.Composition.Children[i].Timeline.Display.Title = "changed"
		}
	}
	after, _ = json.Marshal(input)
	if string(before) != string(after) {
		t.Fatal("output aliases input")
	}
}
func TestHegelTimeTemplateTargetBinding(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { timeTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelTimeTemplateAcquisitionRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		timeTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 7)))
	}, options(50)...)
}
