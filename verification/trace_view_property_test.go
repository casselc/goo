package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const traceViewJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import {traceView} from '../ui/record-results.mjs';const {origin,id,expected}=JSON.parse(fs.readFileSync(0,'utf8'));const before=structuredClone(origin),view=traceView(origin,id);assert.deepEqual(view,expected);assert.deepEqual(origin,before);assert.equal(view.queries[0].offset,0);assert.equal(view.composition.children.length,1);assert.equal(view.composition.children[0].records.kind.name,'record-span-waterfall');view.queries[0].where.all[0].value='edited';assert.deepEqual(origin,before);for(const bad of ['0'.repeat(32),'a'.repeat(31),'g'.repeat(32),'A'.repeat(32)])assert.throws(()=>traceView(origin,bad));`

func TestHegelTraceViewPreservesScopeWithoutInheritingSelectors(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		id := fmt.Sprintf("%032x", seed)
		from := int64(seed) * 1000
		origin := map[string]any{"signal": "logs", "fromMs": from, "toMs": from + 60000, "limit": 10, "offset": 100 + (seed%5)*100, "where": map[string]any{"field": map[string]any{"kind": "column", "key": "service"}, "op": "eq", "value": "origin-service"}, "groupBy": []any{map[string]any{"kind": "attribute", "key": "cohort"}}, "bucketMs": 10000}
		q := map[string]any{"signal": "spans", "fromMs": from, "toMs": from + 60000, "limit": 100, "offset": 0, "where": map[string]any{"all": []any{map[string]any{"field": map[string]any{"kind": "column", "key": "trace_id"}, "op": "eq", "value": id}}}}
		if seed%3 == 0 {
			origin["environment"] = "prod"
			q["environment"] = "prod"
		} else {
			policy := "effective"
			if seed%2 == 0 {
				policy = "current-only"
			}
			selector := map[string]any{"version": "goo/environment-selector/v1", "policy": policy, "scope": map[string]any{"mode": "named", "value": "prod"}}
			origin["environmentSelector"] = selector
			q["environmentSelector"] = selector
		}
		expected := map[string]any{"version": 1, "title": "Trace " + id, "queries": []any{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": "spans"}, "kind": map[string]any{"name": "record-span-waterfall", "version": 1}, "display": map[string]any{"columns": []string{"utc", "service", "summary", "context"}}}}}}}
		raw, _ := json.Marshal(expected)
		var admitted workbench.ExplorationView
		if e := json.Unmarshal(raw, &admitted); e != nil {
			tc.Fatal(e)
			return
		}
		if e := admitted.Validate(); e != nil {
			tc.Fatal(e)
			return
		}
		if e := runNativeHistogramJS(map[string]any{"origin": origin, "id": id, "expected": expected}, traceViewJS); e != nil {
			tc.Fatal(e)
		}
	}, options(50)...)
}
