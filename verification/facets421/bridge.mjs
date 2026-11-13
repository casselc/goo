import assert from 'node:assert/strict';
import fs from 'node:fs';
import * as Values from '../../ui/filter-values.mjs';
import * as Entry from '../../ui/activity-facet-entry.mjs';
const cases=JSON.parse(fs.readFileSync(0,'utf8')),clone=structuredClone;
const context=s=>({selector:clone(s.Scope),readWindow:{fromMs:s.Scope.fromMs,toMs:s.Scope.toMs},phase:'paused',rows:[],active:true,clean:true});
let calls=0;globalThis.fetch=()=>{calls++;throw Error('pure facet code attempted a source read')};
for(const s of cases){
 const before=JSON.stringify(s),q=Values.query(s.Scope,s.Field);assert.deepEqual(q,s.Expected);
 const data=Values.result(q,s.Response);assert.equal(data.total,s.Response.count);assert.deepEqual(data.field,s.Field);assert.equal(data.reads.count.basis,s.Response.reads.count.basis);assert.equal(data.reads.groups.basis,s.Response.reads.groups.basis);
 if(s.Response.state==='observed'){assert.equal(data.items.reduce((n,x)=>n+BigInt(x.count),0n).toString(),s.Response.count);assert(data.items.some(x=>x.value===null));assert(data.items.some(x=>x.value===''));assert(data.items.some(x=>x.value==='Missing key'));}
 else assert.deepEqual(data.items,[]);
 const c=context(s),intent=Entry.create(c,s.Field,s.Window);assert.deepEqual(intent.question,q);assert.deepEqual(intent.observationWindow,s.Window);assert.equal(Entry.current(intent,c),true);assert.equal(typeof intent.label,'string');assert(intent.label.includes(new Date(q.fromMs).toISOString()));assert(intent.label.includes(new Date(q.toMs).toISOString()));assert(intent.label.includes('inspected record window differs'));assert.equal(data.completeness,undefined);
 for(const mutate of [c=>c.active=false,c=>c.clean=false,c=>c.phase='loading',c=>c.phase='stopped',c=>c.readWindow.toMs++,c=>c.selector.where={field:{kind:'column',key:'service'},op:'eq',value:'other'},c=>c.selector.environmentSelector={version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'named',value:'other'}}]){const changed=clone(c);mutate(changed);assert.equal(Entry.current(intent,changed),false);}
 for(const key of ['duration_ns','timestamp_ns','severity_number'])assert.throws(()=>Values.query(s.Scope,{kind:'column',key}));
 for(const mutate of [r=>r.question.toMs++,r=>r.question.groupBy[0].kind=s.Field.kind==='resource'?'attribute':'resource',r=>r.count='1',r=>r.state='changed_between_reads',r=>r.reads.count.basis='bad',r=>{r.count='2';r.state='observed';r.groups=[{g0:'dup',count:'1'},{g0:'dup',count:'1'}]},r=>{r.count='0';r.state='no_data';r.groups=[{g0:'a',count:'0'}]},r=>r.version='goo/explore/v1',r=>{r.count='18446744073709551616';r.groups=[{g0:'a',count:r.count}]},r=>r.count='00']){const bad=clone(s.Response);mutate(bad);assert.throws(()=>Values.result(q,bad));}
 assert.equal(JSON.stringify(s),before);assert.equal(calls,0);
}
// The source owner must ignore a response after its fixed scope/selector changes.
for(const s of cases.slice(0,3))for(const mode of ['scope','entry']){
 let now=clone(s.Scope),loader,complete,signal,sourceCalls=0;const states=[];
 globalThis.fetch=(url,options)=>{if(url.startsWith('/api/queries/'))return Promise.resolve({body:{cancel:async()=>{}}});assert.equal(url,'/api/explore/counted-values');sourceCalls++;signal=options.signal;return new Promise(resolve=>complete=()=>resolve({ok:true,status:200,json:async()=>clone(s.Response)}))};
 const editor={setValueLoader:f=>loader=f,setCountedValues:value=>states.push(value)},owner=Values.mount(editor,{current:()=>now});
 const pending=loader(s.Field);assert.equal(sourceCalls,1);if(mode==='scope'){now={...now,toMs:now.toMs+1};owner.sync();}else owner.invalidate();assert.equal(signal.aborted,true);complete();await pending;assert(!states.some(value=>value?.state==='ready'));owner.dispose();assert.equal(loader,null);
}
process.stdout.write(JSON.stringify(cases.length));
