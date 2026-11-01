package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const addBreakdownJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';const {q,c,options,expected}=JSON.parse(fs.readFileSync(0,'utf8'));const before=structuredClone({q,c,options}),next=C.addBreakdown(c,q,options);assert.deepEqual(next,expected);assert.deepEqual({q,c,options},before);assert.equal(next.query.offset,0);assert.deepEqual(next.composition.children[1],c.children[0]);if(q.groupBy?.length&&(q.groupBy.length!==1||JSON.stringify(q.groupBy[0])!==JSON.stringify(options.field)))assert.throws(()=>C.addBreakdown(c,q,{...options,replaceGrouping:false}));for(const patch of [{kind:'line'},{replaceGrouping:1},{field:{kind:'attribute',key:''}},{field:{kind:'column',key:'not_a_column'}}])assert.throws(()=>C.addBreakdown(c,q,{...options,...patch}));assert.throws(()=>C.addBreakdown(next.composition,next.query,options));next.query.groupBy[0].key='edited';next.composition.children[1].records.display.columns.reverse();assert.deepEqual({q,c,options},before);`

func TestHegelAddBreakdownPreservesRecordsAndRequiresGroupingConsent(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		signal := "logs"
		if seed%2 == 0 {
			signal = "spans"
		}
		q := map[string]any{"signal": signal, "fromMs": int64(seed) * 1000, "toMs": int64(seed)*1000 + 60000, "limit": 100, "offset": 100, "where": map[string]any{"field": map[string]any{"kind": "column", "key": "service"}, "op": "eq", "value": "source-service"}}
		if seed%3 == 0 {
			q["environment"] = "prod"
		} else {
			q["environmentSelector"] = map[string]any{"version": "goo/environment-selector/v1", "policy": "effective", "scope": map[string]any{"mode": "named", "value": "prod"}}
		}
		field := map[string]any{"kind": "attribute", "key": "cohort"}
		groups := []any{}
		switch seed % 4 {
		case 1:
			groups = []any{field}
		case 2:
			groups = []any{map[string]any{"kind": "resource", "key": "zone"}}
		case 3:
			groups = []any{field, map[string]any{"kind": "resource", "key": "zone"}}
		}
		if len(groups) > 0 {
			q["groupBy"] = groups
		}
		kind := []string{"record-table", "record-list"}[seed%2]
		if signal == "spans" && seed%3 == 0 {
			kind = "record-span-waterfall"
		}
		record := map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": kind, "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary", "service"}, "fieldColumns": []any{map[string]any{"kind": "attribute", "key": "zone"}, map[string]any{"kind": "resource", "key": "zone"}}}}}
		c := map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{record}}
		chart := []string{"bar", "table"}[seed%2]
		options := map[string]any{"field": field, "kind": chart, "replaceGrouping": seed%4 >= 2}
		raw, _ := json.Marshal(q)
		query := map[string]any{}
		_ = json.Unmarshal(raw, &query)
		query["groupBy"] = []any{field}
		query["offset"] = 0
		aggregate := map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": chart, "version": 1}, "display": map[string]any{"category": "category", "value": "count", "scale": "population", "order": "value-desc", "query": "", "requiresComplete": false}}}
		composition := map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{aggregate, record}}
		expected := map[string]any{"query": query, "composition": composition}
		encoded, _ := json.Marshal(map[string]any{"version": 1, "title": fmt.Sprint("Grouped records ", seed), "queries": []any{query}, "composition": composition})
		var view workbench.ExplorationView
		if e := json.Unmarshal(encoded, &view); e != nil {
			tc.Fatal(e)
			return
		}
		if e := view.Validate(); e != nil {
			tc.Fatal(e)
			return
		}
		if e := runNativeHistogramJS(map[string]any{"q": q, "c": c, "options": options, "expected": expected}, addBreakdownJS); e != nil {
			tc.Fatal(e)
		}
	}, options(50)...)
}
