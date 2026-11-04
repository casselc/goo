package verification

import (
	"hegel.dev/go/hegel"
	"testing"
)

// A bounded state sequence checks the projection independently of tail custody.
// DOM focus/scroll/source-effect checks live in the maintained browser journey.
func TestHegelHeldFindState324(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		steps := hegel.Draw(tc, hegel.Integers(15, 30))
		commands := []map[string]any{}
		for i := 0; i < steps; i++ {
			commands = append(commands, map[string]any{"kind": hegel.Draw(tc, hegel.SampledFrom([]string{"append", "append", "find", "layout", "select", "signal", "reset"})), "value": hegel.Draw(tc, hegel.Integers(0, 11))})
		}
		if e := runNativeHistogramJS(map[string]any{"seed": seed, "commands": commands}, heldFind324JS); e != nil {
			tc.Fatal(e)
			return
		}
	}, options(30)...)
}

const heldFind324JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {TailBuffer,TailSelection} from '../ui/tail.mjs';import * as R from '../ui/record-presentation.mjs';
const {seed,commands}=JSON.parse(fs.readFileSync(0,'utf8')),buffer=new TailBuffer(4,65536),selection=new TailSelection(),now=1000000,expiry=new Date(now+60000).toISOString();let signal='logs',query='',chosen=false,ref=[],seen=new Set(),selected=new Map(),observed=0,evicted=0;
const queries=['','MESSAGE','operation','svc-','ERROR','needle','Null','unused-token','Present empty string','secret-proof','alpha beta','No-such-value'];
const item=i=>{const values=[null,'','null','needle-'+seed],r={timestamp_ns:String(1790000000000000000n+BigInt(i)*1000000n+123456n),ServiceName:'svc-'+(i%3),SeverityText:i%2?'ERROR':'INFO',StatusCode:i%2?'Error':'Ok',TraceId:'b'.repeat(32),SpanId:'c'.repeat(16),ParentSpanId:'d'.repeat(16),duration_ns:'1000000',ResourceAttributes:{'service.namespace':'ns-'+(i%2),'hidden.secret':'secret-proof'},Attributes:{hidden:'secret-proof'}};if(signal==='logs')r.Body=i===0?'alpha':'message-'+i;else r.SpanName=i===0?'alpha':'operation-'+i;if(i%5!==0)r.Attributes.chosen=values[i%4];return {key:'a'.repeat(64)+':'+(i+1),record:r,connection:'e'.repeat(64),membership:{token:'f'.repeat(64),expiresAt:expiry},readWindow:{fromMs:1790000000000,toMs:1790000060000}}};
const selector=()=>({signal,environment:'prod'}),fields=()=>chosen?[{kind:'attribute',key:'chosen'}]:[];
function oracle(r){const values=[signal==='logs'?r.Body:r.SpanName,r.ServiceName,signal==='logs'?r.SeverityText:r.StatusCode];if(chosen&&Object.hasOwn(r.Attributes,'chosen')){const v=r.Attributes.chosen;if(v!==null&&v!=='')values.push(v)}return values.some(v=>typeof v==='string'&&v.toLowerCase().includes(query.toLowerCase()))}
function check(){const before=JSON.stringify(buffer.rows),actual=buffer.rows.filter(x=>R.matchesLoaded(x.record,signal,fields(),query)).map(x=>x.key),expected=ref.filter(x=>oracle(x.record)).map(x=>x.key);assert.deepEqual(actual,expected);assert.equal(JSON.stringify(buffer.rows),before);assert.deepEqual(buffer.rows.map(x=>x.key),ref.map(x=>x.key));assert.equal(buffer.observed,observed);assert.equal(buffer.evicted,evicted);assert(buffer.rows.length<=4);assert(buffer.size<=65536);assert.deepEqual([...selection.items.keys()],[...selected.keys()]);for(const [key,value]of selected)assert.deepEqual(selection.items.get(key),value);if(selected.size){const intent=selection.keepIntent('request',now);assert.deepEqual(intent.refs.map(r=>r.key),[...selected.keys()]);assert.deepEqual(intent.selector,selector())}}
// Deterministic prefix ensures eviction, hidden selection, and literal distinctions
// even when shrinking the following generated action sequence.
const actions=[{kind:'append',value:1},{kind:'select',value:0},{kind:'find',value:9},{kind:'append',value:2},{kind:'append',value:3},{kind:'append',value:4},{kind:'append',value:5},...commands];
for(const c of actions){if(c.kind==='append'){const next=item(c.value),before=JSON.stringify(next);buffer.append([next]);if(!seen.has(next.key)){seen.add(next.key);observed++;ref.push({...structuredClone(next),size:new TextEncoder().encode(JSON.stringify(next)).length});if(ref.length>4){ref.shift();evicted++}}assert.equal(JSON.stringify(next),before)}else if(c.kind==='find'){query=queries[c.value%queries.length]}else if(c.kind==='layout'){chosen=!!(c.value%2)}else if(c.kind==='select'){const visible=ref.filter(x=>oracle(x.record));if(visible.length){const x=visible[c.value%visible.length];selection.toggle(buffer.rows.find(row=>row.key===x.key),selector(),now);if(selected.has(x.key))selected.delete(x.key);else selected.set(x.key,structuredClone(x))}}else{if(c.kind==='signal')signal=signal==='logs'?'spans':'logs';buffer.reset();selection.clear();ref=[];seen=new Set();selected=new Map();observed=evicted=0;query=''}check()}
const raw=item(7),before=JSON.stringify(raw);assert.equal(R.matchesLoaded(raw.record,signal,fields(),'secret-proof'),false);assert.equal(R.matchesLoaded({Body:'alpha',ServiceName:'beta',ResourceAttributes:{},Attributes:{}},'logs',[],'alpha beta'),false);assert.equal(JSON.stringify(raw),before);
const millis=Number(BigInt(raw.record.timestamp_ns)/1000000n),compactUTC=new Date(millis).toISOString().split('T')[1];assert(R.matchesLoaded(raw.record,signal,[],compactUTC),'copied compact UTC timestamp must match independently of nine-digit nanoseconds');assert.equal(R.findQuery('ERROR'),'error');assert.equal(R.findQuery(' X '),' x ');assert(R.matchesLoaded({Body:'.* [x]'},'logs',[],'.*'));assert(!R.matchesLoaded({Body:'alpha'},'logs',[],'.*'));assert(R.matchesLoaded(raw.record,signal,[],'b'.repeat(32)));assert(!R.matchesLoaded(raw.record,signal,[],raw.record.timestamp_ns));assert(R.matchesLoaded(raw.record,'spans',[],'ms'));assert(!R.matchesLoaded(raw.record,'spans',[{kind:'column',key:'duration_ns'}],'1000000'));assert(R.findValues(raw.record,signal,[]).some(v=>v.includes(new Date(Number(BigInt(raw.record.timestamp_ns)/1000000n)).toISOString().slice(0,19))));for(const bad of ['x'.repeat(257),'\ud800',null,{}])assert.throws(()=>R.findQuery(bad));
const missing={Body:'alpha',Attributes:{}},empty={Body:'alpha',Attributes:{chosen:''}},nil={Body:'alpha',Attributes:{chosen:null}},literal={Body:'alpha',Attributes:{chosen:'null'}},f=[{kind:'attribute',key:'chosen'}];assert(!R.findValues(missing,'logs',f).includes('Missing'));assert(!R.matchesLoaded(empty,'logs',f,'Present empty string'));assert(!R.matchesLoaded(nil,'logs',f,'null'));assert(R.matchesLoaded(literal,'logs',f,'null'));assert(!R.matchesLoaded(literal,'logs',f,'"null"'));assert(R.matchesLoaded({Body:'alpha',Attributes:{chosen:'Missing'}},'logs',f,'missing'));
`
