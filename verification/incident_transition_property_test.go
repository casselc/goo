package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

const incidentTransitionsJS = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';import {partitionQueries,selectionCommandCurrent} from '../ui/population-workspace.mjs';const {view,seed}=JSON.parse(fs.readFileSync(0,'utf8')),q=view.queries[0],before=JSON.stringify(view),field=C.facet(view.composition).field,value=seed%3===0?null:seed%3===1?'':'blue';const parts=partitionQueries([q],[{signal:q.signal,field,value,mode:seed%2?'include':'exclude'}]),selected=parts.selected[0],rest=parts.rest[0],out=C.includeComparisonWithoutFacet(view.composition,selected,rest);assert.equal(C.facet(out),undefined);assert.deepEqual(C.records(out),C.records(view.composition));assert.deepEqual(C.aggregate(out),C.aggregate(view.composition));assert.deepEqual(C.comparison(out).baselineQuery,rest);assert.equal(JSON.stringify(view),before);for(const bad of [{...rest,toMs:rest.toMs+1},{...rest,offset:1},{...rest,groupBy:[{kind:'column',key:'service'}]}])assert.throws(()=>C.includeComparisonWithoutFacet(view.composition,selected,bad));assert.throws(()=>C.includeComparisonWithoutFacet(view.composition,{...selected,spanStatus:true},rest));assert.equal(JSON.stringify(view),before);
const owner={},expected={owner,generation:seed,draftRevision:seed+1,selectionKey:JSON.stringify([field,value]),questionKey:JSON.stringify(q)},ready={...expected,active:true,visible:true,connected:true,child:false,busy:false,countPending:false,sourceDraft:false,queryDraft:false,recordPending:false,pagePending:false};assert(selectionCommandCurrent(expected,ready));for(const key of ['generation','draftRevision','selectionKey','questionKey'])assert(!selectionCommandCurrent(expected,{...ready,[key]:String(ready[key])+'changed'}));assert(!selectionCommandCurrent(expected,{...ready,owner:{}}));for(const key of ['active','visible','connected'])assert(!selectionCommandCurrent(expected,{...ready,[key]:false}));for(const key of ['child','busy','countPending','sourceDraft','queryDraft','recordPending','pagePending'])assert(!selectionCommandCurrent(expected,{...ready,[key]:true}));assert(selectionCommandCurrent(expected,ready));const renewed={...expected,owner:{}};assert(!selectionCommandCurrent(expected,{...ready,owner:renewed.owner}));assert(selectionCommandCurrent(renewed,{...ready,owner:renewed.owner}));process.stdout.write(JSON.stringify({version:1,title:'Explicit comparison',queries:[selected],composition:out}));`

func TestHegelIncidentPromotionAndSelectionLeaseGuards(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		q := authoredTemplateQuery(seed, "ordinary", "incident-source", true)
		q.Offset = 0
		q.SpanStatus = false
		scope := map[string]any{"kind": "record-query", "signal": q.Signal}
		field := exploration.Field{Kind: "attribute", Key: "cohort"}
		if seed%3 == 0 {
			field = exploration.Field{Kind: "resource", Key: "cloud.region"}
		}
		raw, _ := json.Marshal(map[string]any{"version": 1, "title": "Incident count", "queries": []exploration.Query{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"facet": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "categorical-facet", "version": 1}, "field": field, "display": map[string]any{"basis": "applied", "title": "Incident field"}}}, map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}, map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "record-list", "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary"}, "fieldColumns": []exploration.Field{{Kind: "attribute", Key: "zone"}, {Kind: "resource", Key: "zone"}}}}}}}})
		var source workbench.ExplorationView
		if e := json.Unmarshal(raw, &source); e != nil {
			tc.Fatal(e)
			return
		}
		if e := source.Validate(); e != nil {
			tc.Fatal(e)
			return
		}
		before, _ := json.Marshal(source)
		payload, _ := json.Marshal(map[string]any{"view": source, "seed": seed})
		cmd := exec.Command("node", "--input-type=module", "-e", incidentTransitionsJS)
		cmd.Stdin = bytes.NewReader(payload)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(out))
			return
		}
		var promoted workbench.ExplorationView
		if e = json.Unmarshal(out, &promoted); e != nil {
			tc.Fatal(e, string(out))
			return
		}
		if e = promoted.Validate(); e != nil {
			tc.Fatal(e)
			return
		}
		after, _ := json.Marshal(source)
		if !bytes.Equal(before, after) {
			tc.Fatal("source changed")
			return
		}
		var baseline *exploration.Query
		for _, c := range promoted.Composition.Children {
			if c.Facet != nil {
				tc.Fatal("facet survived explicit removal")
				return
			}
			if c.Comparison != nil {
				baseline = c.Comparison.BaselineQuery
			}
		}
		if baseline == nil || !reflect.DeepEqual(promoted.Queries[0].GroupBy, q.GroupBy) || !promoted.Queries[0].SameEnvironment(q) || promoted.Queries[0].FromMS != q.FromMS || promoted.Queries[0].ToMS != q.ToMS {
			tc.Fatal("query context changed")
			return
		}
		encoded, e := workbench.EncodeViewArtifact(workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "incident", Name: "comparison", Revision: "v1", View: promoted})
		if e != nil {
			tc.Fatal(e)
			return
		}
		decoded, e := workbench.DecodeViewArtifact(encoded.YAML)
		if e != nil || !reflect.DeepEqual(decoded.Artifact.View, promoted) {
			tc.Fatal("Go/portable promotion changed", e)
			return
		}
	}, options(50)...)
}
