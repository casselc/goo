package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const recordOnlyJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';const {q,target,c,accepted,reused}=JSON.parse(fs.readFileSync(0,'utf8'));const before=JSON.stringify({q,target,c});let next;try{next=C.validate(c,q)}catch{}assert.equal(!!next,accepted);if(!accepted)process.exit(0);assert(C.recordOnly(next));const actual=C.reuse(c,q,target);for(const value of [actual,reused])if(value.children[0].records.display.fieldColumns?.length===0)delete value.children[0].records.display.fieldColumns;assert.deepEqual(actual,reused);assert.equal(JSON.stringify({q,target,c}),before);assert.throws(()=>C.reuse(c,q,{...target,signal:target.signal==='logs'?'spans':'logs'}));assert.throws(()=>C.setHidden(c,'records',true,q));`

func recordOnlyProperty(tc *hegel.T, seed, bad int) {
	signal := "logs"
	if seed%2 == 1 {
		signal = "spans"
	}
	q := exploration.Query{Signal: signal, Environment: "source", FromMS: 1000, ToMS: 61000, Limit: 100, Offset: 0}
	for i := 0; i < seed%3; i++ {
		q.GroupBy = append(q.GroupBy, exploration.Field{Kind: "attribute", Key: fmt.Sprint("group", i)})
	}
	kind := []string{"record-table", "record-list"}[seed%2]
	if signal == "spans" && seed%3 == 0 {
		kind = "record-span-waterfall"
	}
	fields := []any{}
	for i := 0; i < seed%9; i++ {
		fields = append(fields, map[string]any{"kind": "attribute", "key": fmt.Sprint("field", i)})
	}
	d := map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": kind, "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary"}, "fieldColumns": fields}}
	child := map[string]any{"records": d}
	c := map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{child}}
	switch bad {
	case 0:
		child["hidden"] = true
	case 1:
		c["children"] = []any{child, child}
	case 2:
		c["children"] = []any{}
	case 3:
		d["kind"] = map[string]any{"name": "record-list", "version": 2}
	case 4:
		d["scope"] = map[string]any{"kind": "record-query", "signal": "metrics"}
	case 5:
		d["display"] = map[string]any{"columns": []string{"summary"}}
	case 6:
		q.BucketMS = 10000
		c["children"] = []any{child, map[string]any{"timeline": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": signal}, "kind": map[string]any{"name": "time-bar", "version": 1}, "display": map[string]any{"value": "count"}}}}
	case 7:
		d["display"] = map[string]any{"columns": []string{"utc", "summary"}, "fieldColumns": []any{map[string]any{"kind": "attribute", "key": "same"}, map[string]any{"kind": "attribute", "key": "same"}}}
	}
	raw, _ := json.Marshal(c)
	var native workbench.ViewComposition
	if e := json.Unmarshal(raw, &native); e != nil {
		tc.Fatal(e)
		return
	}
	accepted := bad < 0
	if e := native.Validate(q); (e == nil) != accepted {
		tc.Fatal("Go admission mismatch", bad, e)
		return
	}
	target := q
	target.Environment = "target"
	target.FromMS += 100000
	target.ToMS += 100000
	target.GroupBy = []exploration.Field{{Kind: "resource", Key: "destination"}}
	target.Offset = 100
	var reused workbench.ViewComposition
	if accepted {
		before, _ := json.Marshal(native)
		var e error
		reused, e = workbench.ReuseViewComposition(native, q, target)
		if e != nil {
			tc.Fatal(e)
			return
		}
		after, _ := json.Marshal(native)
		if string(before) != string(after) {
			tc.Fatal("reuse mutated source")
			return
		}
	}
	if e := runNativeHistogramJS(map[string]any{"q": q, "target": target, "c": c, "accepted": accepted, "reused": reused}, recordOnlyJS); e != nil {
		tc.Fatal(e)
	}
}
func TestHegelRecordOnlyCompositionReuse(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { recordOnlyProperty(tc, hegel.Draw(tc, hegel.Integers(0, 100000)), -1) }, options(50)...)
}
func TestHegelRecordOnlyCompositionRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		recordOnlyProperty(tc, hegel.Draw(tc, hegel.Integers(0, 100000)), hegel.Draw(tc, hegel.Integers(0, 7)))
	}, options(50)...)
}
