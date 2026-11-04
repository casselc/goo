package verification

import (
	"hegel.dev/go/hegel"
	"testing"
)

// Batch selection is a local atomic union; no publication or source fixture is
// needed to verify its custody references and all-or-nothing class mutation.
func TestHegelLoadedMatchSelectionAtomic325(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		steps := hegel.Draw(tc, hegel.Integers(10, 25))
		commands := []map[string]any{}
		for i := 0; i < steps; i++ {
			commands = append(commands, map[string]any{"kind": hegel.Draw(tc, hegel.SampledFrom([]string{"add", "add", "duplicate", "conflict", "overcap", "clear"})), "start": hegel.Draw(tc, hegel.Integers(0, 24)), "width": hegel.Draw(tc, hegel.Integers(1, 5))})
		}
		if e := runNativeHistogramJS(map[string]any{"seed": seed, "commands": commands}, loadedMatchSelection325JS); e != nil {
			tc.Fatal(e)
			return
		}
	}, options(30)...)
}

const loadedMatchSelection325JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import {TailSelection} from '../ui/tail.mjs';
const {seed,commands}=JSON.parse(fs.readFileSync(0,'utf8')),now=1000000,selector={signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod-'+seed}}},connection='b'.repeat(64);
const item=n=>({key:'a'.repeat(64)+':'+(n+1),connection,observedAt:new Date(now).toISOString(),limited:false,late:false,record:{timestamp_ns:String(1790000000000000000n+BigInt(n)),Body:'match-'+n,ResourceAttributes:{'deployment.environment.name':'prod-'+seed}},readWindow:{fromMs:1790000000000,toMs:1790000060000},membership:{token:(n+1).toString(16).padStart(64,'0'),expiresAt:new Date(now+60000).toISOString()}});
const snapshot=s=>JSON.stringify({items:[...s.items],selector:s.selector,connection:s.connection});
function unchanged(s,before,map,scope,entries){assert.equal(snapshot(s),before);assert.equal(s.items,map);assert.equal(s.selector,scope);for(const [key,value]of entries)assert.equal(s.items.get(key),value)}
function refuse(s,batch,q=selector,at=now){const before=snapshot(s),map=s.items,scope=s.selector,entries=[...s.items];assert.throws(()=>s.planMany(batch,q,at));unchanged(s,before,map,scope,entries);assert.throws(()=>s.addMany(batch,q,at));unchanged(s,before,map,scope,entries)}
function add(s,batch,expected){const before=JSON.stringify(batch),old=[...s.items],oldSelector=s.selector,oldMap=s.items,oldState=snapshot(s);const added=new Set(batch.filter(x=>!expected.has(x.key)).map(x=>x.key)).size,total=expected.size+added,plan=s.planMany(batch,selector,now);assert.equal(plan.added,added);assert.equal(plan.total,total);if(plan.items.size)plan.items.values().next().value.record.Body='edited plan';plan.selector.environmentSelector.scope.value='edited plan';unchanged(s,oldState,oldMap,oldSelector,old);const outcome=s.addMany(batch,selector,now);assert.deepEqual(outcome,{added,total});assert.equal(JSON.stringify(batch),before);for(const [key,value]of old)assert.equal(s.items.get(key),value);if(old.length)assert.equal(s.selector,oldSelector);for(const x of batch)if(!expected.has(x.key))expected.set(x.key,structuredClone(x));assert.deepEqual([...s.items.keys()],[...expected.keys()]);for(const [key,value]of expected)assert.deepEqual(s.items.get(key),value);const intent=s.keepIntent('request',now);assert.deepEqual(intent.refs,[...expected.values()].map(x=>({key:x.key,token:x.membership.token})));assert.deepEqual(intent.selector,selector)}
const s=new TailSelection();let expected=new Map();add(s,[item(0)],expected);add(s,[item(1),item(2)],expected); // old hidden/evicted key0 is absent from current matches
for(const c of commands){if(c.kind==='clear'){s.clear();expected=new Map();continue}let batch=Array.from({length:c.width},(_,i)=>item((c.start+i)%25));if(c.kind==='duplicate')batch=[batch[0],structuredClone(batch[0])];if(c.kind==='conflict'){const other=structuredClone(batch[0]);other.record.Body='changed occurrence';refuse(s,[batch[0],other]);continue}if(c.kind==='overcap'){refuse(s,Array.from({length:21},(_,i)=>item(i)));continue}const keys=new Set([...expected.keys(),...batch.map(x=>x.key)]);if(keys.size>20)refuse(s,batch);else add(s,batch,expected)}
// Exact cap, original reference identity and no prefix selection on refusal.
const cap=new TailSelection(),capExpected=new Map();add(cap,Array.from({length:20},(_,i)=>item(i)),capExpected);refuse(cap,[item(20)]);add(cap,[item(1),item(1)],capExpected);
const stable=new TailSelection(),stableExpected=new Map();add(stable,[item(0)],stableExpected);const proof=stable.keepIntent('stable',now),originalItem=stable.items.get(item(0).key);for(const bad of [x=>x.membership.expiresAt=new Date(now).toISOString(),x=>x.membership.token='bad',x=>x.connection='c'.repeat(64),x=>x.key='bad',x=>x.extra=()=>{}]){const x=item(1);bad(x);refuse(stable,[item(2),x]);assert.deepEqual(stable.keepIntent('stable',now),proof);assert.equal(stable.items.get(item(0).key),originalItem)}
refuse(stable,[item(1)],{...selector,environmentSelector:{...selector.environmentSelector,policy:'current-only'}});const fresh=item(1);fresh.membership.expiresAt=new Date(now+120000).toISOString();refuse(stable,[fresh],selector,now+60000);for(const change of [x=>x.membership.token='d'.repeat(64),x=>x.observedAt=new Date(now+1).toISOString(),x=>x.limited=true,x=>x.late=true]){const changed=item(0);change(changed);refuse(stable,[changed]);}const derived=item(0);derived.size=999;add(stable,[derived],stableExpected);refuse(stable,[]);for(const signal of [false,0,null,{}])refuse(stable,[item(1)],{...selector,signal});const corruptOwner=new TailSelection();corruptOwner.addMany([item(0)],selector,now);corruptOwner.connection=null;refuse(corruptOwner,[item(1)]);
const detached=new TailSelection(),input=item(9),inputSelector=structuredClone(selector);detached.addMany([input],inputSelector,now);input.record.Body='edited';input.membership.token='d'.repeat(64);inputSelector.environmentSelector.scope.value='edited';assert.equal(detached.items.get(item(9).key).record.Body,'match-9');assert.equal(detached.items.get(item(9).key).membership.token,item(9).membership.token);assert.equal(detached.selector.environmentSelector.scope.value,selector.environmentSelector.scope.value);
`
