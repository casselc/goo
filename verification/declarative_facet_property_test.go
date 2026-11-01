package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelDeclarativeFacetGoJSParityAndProjection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		signal := "logs"
		if seed%2 == 0 {
			signal = "spans"
		}
		q := map[string]any{"signal": signal, "fromMs": int64(seed) * 1000, "toMs": int64(seed)*1000 + 60000, "limit": 100, "offset": 0, "where": map[string]any{"field": map[string]any{"kind": "attribute", "key": "tenant"}, "op": "eq", "value": fmt.Sprintf("null: #tenant-%d [9007199254740993] \"quoted\"", seed)}}
		if seed%3 == 0 {
			q["environment"] = "prod"
		} else {
			policy := "effective"
			if seed%2 == 0 {
				policy = "current-only"
				q["environment"] = ""
			}
			q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": map[string]any{"mode": "named", "value": "prod"}}
		}
		recordKind := "record-list"
		if signal == "spans" && seed%3 == 0 {
			recordKind = "record-span-waterfall"
		}
		scope := map[string]any{"kind": "record-query", "signal": signal}
		records := map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": recordKind, "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary", "service"}, "fieldColumns": []any{map[string]any{"kind": "attribute", "key": "zone"}, map[string]any{"kind": "resource", "key": "zone"}}}}}
		facetField := map[string]any{"kind": "attribute", "key": "cohort"}
		if seed%3 == 0 {
			facetField = map[string]any{"kind": "resource", "key": "cloud.region"}
		} else if seed%3 == 1 {
			facetField = map[string]any{"kind": "column", "key": "service"}
		}
		facet := map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "categorical-facet", "version": 1}, "field": facetField, "display": map[string]any{"basis": "applied", "title": fmt.Sprintf("Team facet %d", seed)}}
		children := []any{records, map[string]any{"facet": facet}}
		if seed%5 == 0 {
			q["groupBy"] = []any{map[string]any{"kind": "column", "key": "service"}}
			children = append([]any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}}, children...)
		}
		if signal == "spans" {
			if seed%4 == 0 {
				q["summary"] = "span-duration"
			} else {
				q["spanStatus"] = true
			}
		}
		if seed%7 == 0 {
			q["bucketMs"] = 1000
		}
		viewSpec := map[string]any{"version": 1, "title": fmt.Sprintf("Service facet %d", seed), "queries": []any{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}}

		raw, _ := json.Marshal(viewSpec)
		var view workbench.ExplorationView
		if e := json.Unmarshal(raw, &view); e != nil {
			tc.Fatal(e)
			return
		}
		artifact := workbench.ViewArtifact{Version: workbench.ViewArtifactVersion, Namespace: "null", Name: fmt.Sprintf("service-%d", seed), Revision: fmt.Sprint(seed), View: view}
		before, _ := json.Marshal(artifact)
		encoded, e := workbench.EncodeViewArtifact(artifact)
		if e != nil {
			tc.Fatal(e)
			return
		}
		decoded, e := workbench.DecodeViewArtifact(encoded.YAML)
		if e != nil {
			tc.Fatal(e)
			return
		}
		after, _ := json.Marshal(decoded.Artifact)
		if string(before) != string(after) || encoded.YAML != decoded.YAML {
			tc.Fatal("portable AST/canonical YAML changed", encoded.YAML)
			return
		}
		again, e := workbench.EncodeViewArtifact(decoded.Artifact)
		if e != nil || again.YAML != encoded.YAML {
			tc.Fatal("canonical encoding unstable", e)
			return
		}
		payload, _ := json.Marshal(decoded.Artifact.View)
		script := `import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';import * as F from '../ui/facet-declaration.mjs';import * as V from '../ui/filter-values.mjs';import * as T from '../ui/view-template-handoff.mjs';import * as CC from '../ui/captured-count-comparison.mjs';const view=JSON.parse(fs.readFileSync(0,'utf8')),q=view.queries[0],before=JSON.stringify(view),comp=C.validate(view.composition,q),f=C.facet(comp);assert.deepEqual(F.validate(f,q),f);assert.equal(C.recordOnly(comp),!comp.children.some(c=>c.aggregate));const projected=V.query(q,f.field);assert.equal(projected.limit,1);assert.equal(projected.offset,0);assert.deepEqual(projected.groupBy,[f.field]);assert.deepEqual(projected.where,q.where);assert.equal(projected.summary,undefined);assert.equal(projected.bucketMs,undefined);assert.equal(projected.spanStatus,undefined);for(const mutate of [d=>d.display.basis='self-excluded',d=>d.field={kind:'column',key:'duration_ns'},d=>d.values=['invented'],d=>d.scope.signal=q.signal==='spans'?'logs':'spans']){const bad=structuredClone(f);mutate(bad);assert.throws(()=>F.validate(bad,q))}assert.throws(()=>T.templateView({...view,template:{version:'goo.template/1',profile:'ordinary',group:'destination'}}));const reused=C.reuse(comp,q,{...q,where:{field:{kind:'attribute',key:'destination'},op:'eq',value:'other'}});assert.deepEqual(C.facet(reused),f);C.facet(reused).field.key='changed';assert.equal(JSON.stringify(view),before);const pq={...structuredClone(q),groupBy:[{kind:'column',key:'service'}]};for(const key of ['summary','bucketMs','spanStatus'])delete pq[key];const aggregate={version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'bar',version:1},display:{category:'category',value:'count'}};const invalidPair={...view,queries:[pq],composition:{...comp,children:[{aggregate},...comp.children.filter(c=>c.records||c.facet),{comparison:CC.createQuery(pq,pq)}]}};assert.throws(()=>C.validate(invalidPair.composition,pq));process.stdout.write(JSON.stringify({projected,invalidPair}));`
		cmd := exec.Command("node", "--input-type=module", "-e", script)
		cmd.Stdin = bytes.NewReader(payload)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(out))
			return
		}
		var returned struct {
			Projected   exploration.Query         `json:"projected"`
			InvalidPair workbench.ExplorationView `json:"invalidPair"`
		}
		if e = json.Unmarshal(out, &returned); e != nil {
			tc.Fatal(e, string(out))
			return
		}
		if e = exploration.ValidateCountedValuesQuery(returned.Projected); e != nil {
			tc.Fatal(e)
			return
		}
		if returned.InvalidPair.Validate() == nil {
			tc.Fatal("Go accepted facet/comparison pair refused by JS")
			return
		}
		// Both encode and decode return detached data, never aliases into author state.
		encoded.Artifact.View.Queries[0].Where.Value = "edited encode"
		decoded.Artifact.View.Queries[0].Where.Value = "edited decode"
		after, _ = json.Marshal(artifact)
		if string(before) != string(after) {
			tc.Fatal("codec mutated author state")
			return
		}
		for _, bad := range []string{encoded.YAML + "\n---\n{}\n", encoded.YAML + "\nunknown: true\n", "version: goo.view-artifact/1\nversion: goo.view-artifact/1\n"} {
			if _, e = workbench.DecodeViewArtifact(bad); e == nil {
				tc.Fatal("malformed artifact accepted")
				return
			}
		}
	}, options(50)...)
}
