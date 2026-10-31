package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const authoredTemplatePublicJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';import * as T from '../ui/view-template-handoff.mjs';
const {source,target,expanded}=JSON.parse(fs.readFileSync(0,'utf8'));const expected=structuredClone(target);expected.offset=0;
assert.deepEqual(expanded.view.queries,[expected]);assert.deepEqual(T.authoredRequest(source,target),{source,query:target});assert.deepEqual(T.authoredResponse(expanded,source,target),expanded.view);assert.deepEqual(T.stagedView(expanded.view,expected),expanded.view);C.view(expanded.view);
const wire=structuredClone(target);function explicitEmpty(p){if(!p)return;if(p.op==='eq'&&!Object.hasOwn(p,'value'))p.value='';for(const child of [...(p.all||[]),...(p.any||[])])explicitEmpty(child)}explicitEmpty(wire.where);wire.groupTimeSeries=false;if(wire.signal==='spans'&&!Object.hasOwn(wire,'spanStatus'))wire.spanStatus=false;assert.deepEqual(T.authoredRequest(source,wire),{source,query:wire});assert.deepEqual(T.authoredResponse(expanded,source,wire),expanded.view);
const changed=structuredClone(wire);changed.where.all[0].value='different-current-service';assert.throws(()=>T.authoredResponse(expanded,source,changed));const qualified=structuredClone(wire);qualified.groupBy[0].kind=qualified.groupBy[0].kind==='attribute'?'resource':'attribute';assert.throws(()=>T.authoredResponse(expanded,source,qualified));
assert.deepEqual(expanded.view.template,source.template);assert.equal(C.aggregate(expanded.view.composition).display.query||'','');
assert.notDeepEqual(source.queries[0].groupBy,expected.groupBy);assert.notEqual(source.queries[0].where.all[0].value,expected.where.all[0].value);
const before=structuredClone(source);expanded.view.queries[0].where.all[0].value='changed-output';expanded.view.composition.children[0].aggregate.display.title='changed-output';assert.deepEqual(source,before);`

func authoredTemplateQuery(seed int, profile, service string, target bool) exploration.Query {
	signal := "spans"
	if profile == "ordinary" && seed%2 == 0 {
		signal = "logs"
	}
	q := exploration.Query{Signal: signal, Environment: "prod", FromMS: int64(1001 + seed), ToMS: int64(61001 + seed), Limit: seed%20 + 1, Offset: 0, BucketMS: 1000, GroupBy: []exploration.Field{{Kind: "attribute", Key: "db.operation.name"}}}
	clauses := []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: service}}
	if profile != "ordinary" {
		kind, key := "Client", "db.system.name"
		if profile == "otel.http-server" {
			kind, key = "Server", "http.request.method"
		}
		clauses = append(clauses, exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "kind"}, Op: "eq", Value: kind}, exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: key}, Op: "ne"})
	}
	clauses = append(clauses, exploration.Predicate{Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "missing"}, {Field: &exploration.Field{Kind: "attribute", Key: "release"}, Op: "eq", Value: ""}}})
	q.Where = &exploration.Predicate{All: clauses}
	if signal == "spans" {
		q.Summary = "span-duration"
		q.SpanStatus = seed%2 == 0
	}
	if target {
		q.FromMS += 100000
		q.ToMS += 100000
		q.Limit = seed%40 + 30
		q.Offset = (seed % 3) * 100
		q.BucketMS = 10000
		q.GroupBy = []exploration.Field{{Kind: []string{"attribute", "resource"}[seed%2], Key: "db.namespace"}}
		scope := identity.EnvironmentScope{Mode: []string{"named", "empty", "unreported"}[seed%3]}
		if scope.Mode == "named" {
			scope.Value = "target"
		}
		selected, _ := identity.EffectiveEnvironment(scope)
		wire, _ := selected.Selector()
		q.Environment = ""
		q.EnvironmentSelector = &wire
		q.Where.All = append(q.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: fmt.Sprintf("team-%d", seed)})
	}
	return q
}
func authoredTemplateCase(t interface{ Fatal(...any) }, seed, mutation int) {
	profile := []string{"ordinary", "otel.db-client", "otel.http-server"}[seed%3]
	if mutation >= 0 {
		profile = "otel.db-client"
	}
	sourceQ, targetQ := authoredTemplateQuery(seed, profile, "source-service", false), authoredTemplateQuery(seed, profile, "target-service", true)
	value := "count"
	scale := "population"
	if sourceQ.Signal == "spans" && (seed%2 == 0 || mutation == 7) {
		value, scale = "duration.p95", "max"
	}
	raw, _ := json.Marshal(map[string]any{"version": 1, "title": "Team database observations", "queries": []exploration.Query{sourceQ}, "template": map[string]any{"version": "goo.template/1", "group": "destination", "profile": profile}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": sourceQ.Signal}, "kind": map[string]any{"name": []string{"bar", "table"}[seed%2], "version": 1}, "display": map[string]any{"category": "category", "value": value, "scale": scale, "order": "label", "query": "", "requiresComplete": seed%2 == 0}}}, map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": sourceQ.Signal}, "kind": map[string]any{"name": "record-table", "version": 1}, "display": map[string]any{"columns": []string{"utc", "service", "summary"}, "fieldColumns": []exploration.Field{{Kind: "attribute", Key: "db.namespace"}}}}}}}})
	var source workbench.ExplorationView
	if e := json.Unmarshal(raw, &source); e != nil {
		t.Fatal(e)
		return
	}
	switch mutation {
	case 0:
		source.Template.Group = "source"
	case 1:
		targetQ.Where = &exploration.Predicate{Any: []exploration.Predicate{*targetQ.Where, {Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "other"}}}
	case 2:
		targetQ.GroupBy = []exploration.Field{{Kind: "column", Key: "duration_bucket"}}
	case 3:
		targetQ.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "bytes"}, Type: "uint64", Parse: "canonical-decimal", Unit: "By"}
	case 4:
		source.Composition.Children[0].Aggregate.Display.Query = "source-specific"
	case 5:
		source.ID = strings.Repeat("a", 64)
	case 6:
		source.Composition.Children = append(source.Composition.Children, workbench.ViewCompositionChild{})
	case 7:
		targetQ.Summary = ""
	}
	in := workbench.ViewTemplateExpansionRequest{Source: &source, Query: &targetQ}
	before, _ := json.Marshal(in)
	out, e := workbench.ExpandViewTemplate(in)
	if (e == nil) != (mutation < 0) {
		t.Fatal("authored template verdict", seed, mutation, e)
		return
	}
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("expansion mutated source/target")
		return
	}
	if e != nil {
		return
	}
	expected := targetQ
	expected.Offset = 0
	if len(out.View.Queries) != 1 || !reflect.DeepEqual(out.View.Queries[0], expected) {
		t.Fatal("expansion changed target query")
		return
	}
	if e = runNativeHistogramJS(map[string]any{"source": source, "target": targetQ, "expanded": out}, authoredTemplatePublicJS); e != nil {
		t.Fatal(e)
		return
	}
	out.View.Queries[0].Where.All[0].Value = "mutated"
	out.View.Composition.Children[0].Aggregate.Display.Title = "mutated"
	after, _ = json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("output aliases source/target")
	}
}
func TestHegelAuthoredTemplateTargetBinding(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { authoredTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), -1) }, options(50)...)
}
func TestHegelAuthoredTemplateRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		authoredTemplateCase(tc, hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 7)))
	}, options(50)...)
}
