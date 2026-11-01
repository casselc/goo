package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"testing"
)

const suppliedWaterfallJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import * as C from '../ui/view-composition.mjs';import * as P from '../ui/record-page.mjs';import * as W from '../ui/waterfall.mjs';import * as R from '../ui/record-span-waterfall.mjs';import {createSelection} from '../ui/record-interactions.mjs';
const {q,d,accepted,seed,expanded}=JSON.parse(fs.readFileSync(0,'utf8'));let verdict=true;try{C.recordDisplay(d,q)}catch{verdict=false}assert.equal(verdict,accepted);if(!accepted)process.exit(0);
assert.equal(expanded.view.composition.children.find(c=>c.records).records.kind.name,'record-span-waterfall');
const trace=n=>String(n).repeat(32),id=n=>n.toString(16).padStart(16,'0'),row=(traceID,span,parent,start,duration)=>({TraceId:traceID,SpanId:span,ParentSpanId:parent,timestamp_ns:String(start),duration_ns:String(duration),ServiceName:'service',SpanName:'span',Attributes:{zone:'',nil:null},ResourceAttributes:{zone:'resource'}}),origin=1700000000000000000n+BigInt(seed)*1000000000n;
const rows=[row(trace(1),id(2),id(1),origin+1000000n,2000000n),row(trace(1),id(1),'',origin,5000000n),row(trace(2),id(3),'',origin+60000000000n,500000000n),row(trace(3),id(4),'',origin,1n),row(trace(3),id(4),'',origin+1n,1n),row(trace(3),id(5),id(4),origin+2n,1n),row('bad',id(6),'',origin,1n),row('bad',id(7),id(6),origin+1n,1n),row(trace(4),id(8),id(9),origin,1n),row(trace(4),id(9),id(8),origin+1n,1n)];
for(let i=0;i<seed%30;i++)rows.push(row(trace(5),id(i+100),'',origin+BigInt(i)*100n,BigInt(i+1)));
const before=structuredClone(rows),a=W.layout(rows,{origin:'trace-relative'}),byIndex=new Map(a.spans.map(s=>[s.index,s]));assert.deepEqual(rows,before);assert.equal(a.spans.length,rows.length);assert.deepEqual([...byIndex.keys()].sort((a,b)=>a-b),rows.map((_,i)=>i));assert(a.spans.findIndex(s=>s.index===1)<a.spans.findIndex(s=>s.index===0),'parent appears first but retains its source index');
assert.equal(byIndex.get(1).left,0);assert.equal(byIndex.get(2).left,0);assert.equal(byIndex.get(2).width/byIndex.get(1).width,100);assert(byIndex.get(5).ambiguousParent);assert(byIndex.get(5).parentIndex==null);assert.notEqual(byIndex.get(6).groupKey,byIndex.get(7).groupKey);assert(byIndex.get(7).parentIndex==null);assert(byIndex.get(8).cycle||byIndex.get(9).cycle);
for(const s of a.spans){assert(s.left>=0&&s.width>=0&&s.left+s.width<=100.00001);if(s.parentIndex!=null){const parent=rows[s.parentIndex];assert.equal(parent.TraceId,s.row.TraceId);assert.equal(parent.SpanId,s.row.ParentSpanId);assert.equal(rows.filter(r=>r.TraceId===parent.TraceId&&r.SpanId===parent.SpanId).length,1)}}
const moved=rows.map(r=>({...r,timestamp_ns:String(BigInt(r.timestamp_ns)+1000000000000n)})),b=W.layout(moved,{origin:'trace-relative'});for(const s of b.spans){assert.equal(s.width,byIndex.get(s.index).width);assert.equal(s.left,byIndex.get(s.index).left)}
const zero=W.layout([row(trace(6),id(1),'',origin,0n)],{origin:'trace-relative'});assert.equal(zero.spans[0].width,0);assert.equal(zero.spans[0].timingValid,true);
const malformed=[...rows,{...rows[0],timestamp_ns:'bad',duration_ns:'bad'}],m=W.layout(malformed,{origin:'trace-relative'});assert.equal(m.spans.length,malformed.length);assert.equal(m.spans.find(s=>s.index===rows.length).timingValid,false);
const raw={question:q,basis:'a'.repeat(64),capturedAt:'2026-11-01T00:00:00Z',state:'observed',rows,hasMore:true},page=P.fromExploration(raw),model=R.waterfallModel(page,d.display);assert.equal(model.page,page);assert.equal(page.page.returned,rows.length);assert.equal(page.quality.extent,'bounded-page');assert(!Object.hasOwn(model,'completeness'));assert(Object.isFrozen(page.rows));
const selected=createSelection(rows.length),visualFirst=a.spans[0].index;assert(selected.set(visualFirst,true));assert.deepEqual(selected.indexes(),[visualFirst]);assert.deepEqual(raw.rows,rows);
`

func waterfallProperty(tc *hegel.T, seed, bad int) {
	q := exploration.Query{Signal: "spans", Environment: "source", FromMS: 1000, ToMS: 61000, Limit: 100, Offset: 0, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}}
	d := workbench.RecordViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.RecordViewKind{Name: "record-span-waterfall", Version: 1}, Display: workbench.RecordViewDisplay{Columns: []string{"utc", "service", "summary", "context"}, FieldColumns: []exploration.Field{{Kind: "attribute", Key: "zone"}, {Kind: "resource", Key: "zone"}, {Kind: "attribute", Key: "nil"}}}}
	switch bad {
	case 0:
		q.Signal = "logs"
		d.Scope.Signal = "logs"
	case 1:
		d.Kind.Version = 2
	case 2:
		d.Display.Columns = []string{"summary"}
	case 3:
		d.Scope.Signal = "logs"
	case 4:
		d.Display.FieldColumns = append(d.Display.FieldColumns, d.Display.FieldColumns[0])
	case 5:
		d.Kind.Name = "trace-complete"
	}
	accepted := bad < 0
	err := d.Validate(q)
	if (err == nil) != accepted {
		tc.Fatal("Go declaration verdict", seed, bad, err)
		return
	}
	var expanded workbench.ViewTemplateExpansion
	if accepted {
		raw, _ := json.Marshal(map[string]any{"version": 1, "title": "Reusable supplied timing", "queries": []exploration.Query{q}, "template": map[string]any{"version": "goo.template/1", "group": "destination", "profile": "ordinary"}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": []any{map[string]any{"aggregate": map[string]any{"version": "goo.view/1", "scope": map[string]any{"kind": "record-query", "signal": "spans"}, "kind": map[string]any{"name": "bar", "version": 1}, "display": map[string]any{"category": "category", "value": "count"}}}, map[string]any{"records": d}}}})
		var source workbench.ExplorationView
		if err = json.Unmarshal(raw, &source); err != nil {
			tc.Fatal(err)
			return
		}
		target := q
		target.Environment = "destination"
		target.GroupBy = []exploration.Field{{Kind: "resource", Key: "other"}}
		input := workbench.ViewTemplateExpansionRequest{Source: &source, Query: &target}
		before, _ := json.Marshal(input)
		expanded, err = workbench.ExpandViewTemplate(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			tc.Fatal("expansion mutated supplied declaration")
			return
		}
	}
	if err := runNativeHistogramJS(map[string]any{"q": q, "d": d, "accepted": accepted, "seed": seed, "expanded": expanded}, suppliedWaterfallJS); err != nil {
		tc.Fatal(fmt.Sprint(err))
	}
}
func TestHegelSuppliedWaterfallGeometryAndDeclarations(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) { waterfallProperty(tc, hegel.Draw(tc, hegel.Integers(0, 100000)), -1) }, options(50)...)
}
func TestHegelSuppliedWaterfallAdmissionRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		waterfallProperty(tc, hegel.Draw(tc, hegel.Integers(0, 100000)), hegel.Draw(tc, hegel.Integers(0, 5)))
	}, options(50)...)
}
