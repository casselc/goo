package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"os/exec"
	"testing"
)

func TestHegelPairedWindowJSOutputAdmittedByGo(t *testing.T) {
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
		children := []any{records}
		q["groupBy"] = []any{map[string]any{"kind": "attribute", "key": "cohort"}}
		baselineRaw, _ := json.Marshal(q)
		var baseline map[string]any
		json.Unmarshal(baselineRaw, &baseline)
		baseline["where"] = map[string]any{"field": map[string]any{"kind": "attribute", "key": "cohort"}, "op": "eq", "value": fmt.Sprint(seed)}
		baseline["limit"] = 1 + seed%100
		comparison := map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "query-count-comparison", "version": 1}, "baselineQuery": baseline, "display": map[string]any{"value": "count"}}
		children = []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": scope, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}, records, map[string]any{"comparison": comparison}}
		viewSpec := map[string]any{"version": 1, "title": fmt.Sprintf("Fixed pair %d", seed), "queries": []any{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}}

		raw, _ := json.Marshal(viewSpec)
		var view workbench.ExplorationView
		if e := json.Unmarshal(raw, &view); e != nil {
			tc.Fatal(e)
			return
		}
		transitionInput, _ := json.Marshal(map[string]any{"view": view, "range": map[string]any{"fromMs": int64(seed)*1000 + 1000, "toMs": int64(seed)*1000 + 59000}})
		script := `import fs from 'node:fs';import assert from 'node:assert/strict';import * as C from '../ui/view-composition.mjs';const {view,range}=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(view),q=view.queries[0],old=C.comparison(view.composition);const result=C.rebaseFixedComparison(view.composition,q,range),next=C.comparison(result.composition);const strip=q=>{const c=structuredClone(q);delete c.fromMs;delete c.toMs;c.offset=0;return c};assert.deepEqual(strip(result.query),strip(q));assert.deepEqual(strip(next.baselineQuery),strip(old.baselineQuery));assert.equal(result.query.fromMs,next.baselineQuery.fromMs);assert.equal(result.query.toMs,next.baselineQuery.toMs);assert.equal(result.query.offset,0);assert.equal(next.baselineQuery.offset,0);const presentation=c=>{const v=structuredClone(c);delete C.comparison(v).baselineQuery;return v};assert.deepEqual(presentation(result.composition),presentation(view.composition));for(const bad of [{fromMs:0,toMs:0},{fromMs:-1,toMs:10},{fromMs:1,toMs:7200002},{fromMs:0,toMs:4102444800001},{fromMs:0.5,toMs:10}])assert.throws(()=>C.rebaseFixedComparison(view.composition,q,bad));assert.equal(JSON.stringify(view),before);process.stdout.write(JSON.stringify({...view,queries:[result.query],composition:result.composition}));`
		cmd := exec.Command("node", "--input-type=module", "-e", script)
		cmd.Stdin = bytes.NewReader(transitionInput)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(out))
			return
		}
		if e = json.Unmarshal(out, &view); e != nil {
			tc.Fatal(e, string(out))
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
		payloadRaw, _ := json.Marshal(encoded.Artifact)
		var payload map[string]any
		json.Unmarshal(payloadRaw, &payload)
		if e := runNativeHistogramJS(payload, `const fs=await import('node:fs');const a=JSON.parse(fs.readFileSync(0,'utf8'));const C=await import('../ui/view-composition.mjs');const q=a.view.queries[0];C.validate(a.view.composition,q);for(const change of [b=>b.baselineCapture='a'.repeat(64),b=>b.baselineQuery.offset=1,b=>b.baselineQuery.fromMs--]){const bad=structuredClone(a.view.composition);change(bad.children.find(c=>c.comparison).comparison);let refused=false;try{C.validate(bad,q)}catch{refused=true}if(!refused)throw Error('invalid fixed pair accepted')}const out=C.validate(a.view.composition,q);out.children.find(c=>c.comparison).comparison.baselineQuery.where.value='edited';if(a.view.composition.children.find(c=>c.comparison).comparison.baselineQuery.where.value==='edited')throw Error('aliased baseline');`); e != nil {
			tc.Fatal(e)
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
