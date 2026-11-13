import assert from 'node:assert/strict';
import fs from 'node:fs';
import * as Selection from '../../ui/counted-value-selection.mjs';
import * as Values from '../../ui/filter-values.mjs';
import * as Entry from '../../ui/activity-facet-entry.mjs';
const cases=JSON.parse(fs.readFileSync(0,'utf8')),clone=structuredClone;
let calls=0;globalThis.fetch=()=>{calls++;throw Error('selection attempted source acquisition')};
const key=f=>JSON.stringify([f.kind,f.key]);
// An independent truth evaluator, with source-map absence distinct from ''.
function matches(p,row){if(!p)return true;if(p.all)return p.all.every(v=>matches(v,row));if(p.any)return p.any.some(v=>matches(v,row));const value=row.get(key(p.field));if(p.op==='exists')return value!==undefined;if(p.op==='missing')return value===undefined;if(p.op==='eq')return value!==undefined&&value===(p.value??'');throw Error('oracle unsupported operator');}
for(const s of cases){
 const original=JSON.stringify(s),q=Values.query(s.Scope,s.Field),context={selector:clone(s.Scope),readWindow:{fromMs:q.fromMs,toMs:q.toMs},active:true,clean:true,phase:'paused'},entry=Entry.create(context,s.Field,{fromMs:0,toMs:60000});assert.deepEqual(entry.question,q);
 const groups=s.Values.map((value,index)=>({g0:value,count:index===0?s.Count:'1'})),response={version:'goo/counted-values/v1',question:q,state:'observed',count:(BigInt(s.Count)+BigInt(s.Values.length-1)).toString(),groups,capturedAt:'2026-11-13T00:02:03Z',basis:'counted-facts',reads:{count:{readAt:'2026-11-13T00:02:02Z',basis:'a'.repeat(64)},groups:{readAt:'2026-11-13T00:02:03Z',basis:'b'.repeat(64)}}};
 const data=Values.result(q,response);assert.equal(data.total,response.count);assert.equal(data.completeness,undefined);const dataBefore=JSON.stringify(data);
 const selected=new Set(),base=s.Scope.where,rows=[];
 for(const value of [...s.Values,'outside'])for(const tenant of [s.Scope.where.all[0].value,'foreign'])for(const namespace of ['',undefined]){
  const row=new Map([[key({kind:'attribute',key:'tenant'}),tenant],[key({kind:'resource',key:'service.namespace'}),namespace],[key({kind:s.Field.kind==='resource'?'attribute':'resource',key:s.Field.key}),'lookalike']]);
  if(value!==null)row.set(key(s.Field),value);rows.push({value,row});
 }
 let last=null;
 for(const step of s.Steps){
  if(step.Clear)selected.clear();else if(step.Selected)selected.add(step.Index);else selected.delete(step.Index);
  const values=[...selected].map(i=>data.items.find(item=>item.value===s.Values[i]).value),before=JSON.stringify(base);
  if(!values.length){assert.throws(()=>Selection.predicate(base,s.Field,values,s.Signal));assert.equal(JSON.stringify(base),before);continue;}
  const predicate=Selection.predicate(base,s.Field,values,s.Signal);
  for(const {value,row} of rows){const wanted=matches(base,row)&&values.some(v=>v===value);assert.equal(matches(predicate,row),wanted,'OR must remain within the exact field and AND with the retained scope');}
  assert.equal(JSON.stringify(base),before);last=predicate;
  assert.equal(JSON.stringify(data),dataBefore,'local selection must not rewrite the applied count or snapshot');assert.equal(predicate.total,undefined);assert.equal(predicate.count,undefined);
 }
 if(last){const held=clone(last);assert.equal(Entry.current(entry,{...context,clean:false}),false);assert.equal(Entry.current(entry,{...context,readWindow:{...context.readWindow,toMs:q.toMs+1}}),false);assert.deepEqual(last,held);}
 for(const values of [[],[null,null],['',''],[1],[false],[{}],['x'.repeat(1025)],['\ud800'],Array.from({length:33},(_,i)=>String(i))]){const before=JSON.stringify(base);assert.throws(()=>Selection.predicate(base,s.Field,values,s.Signal));assert.equal(JSON.stringify(base),before);}
 for(const column of ['timestamp_ns','duration_ns','severity_number'])assert(Selection.valueReason({kind:'column',key:column},'1',s.Signal));
 assert.equal(Selection.valueReason(s.Field,null,s.Signal),'');assert.equal(Selection.valueReason(s.Field,'',s.Signal),'');assert.equal(Selection.valueReason(s.Field,'Missing key',s.Signal),'');
 // Both size and depth refusals must occur before caller replacement.
 const leaf={field:{kind:'attribute',key:'budget'},op:'exists'},dense={all:Array.from({length:31},()=>clone(leaf))},deep={any:[{all:[{any:[{all:[clone(leaf)]}]}]}]};
 for(const where of [dense,deep]){const before=JSON.stringify(where);assert.throws(()=>Selection.predicate(where,s.Field,[null,''],s.Signal));assert.equal(JSON.stringify(where),before);}
 assert.equal(JSON.stringify(s),original);assert.equal(calls,0);
}
// Exercise the real source-owner authority passed to the renderer. A scope
// change refuses immediately, and replacement/ABA cannot revive its old lease.
for(const s of cases.slice(0,3))for(const mode of ['scope','replace','aba']){
 let now=clone(s.Scope),loader,ready,authority,acquisitions=0;
 globalThis.fetch=async(url,options)=>{if(url.startsWith('/api/queries/'))return {body:{cancel:async()=>{}}};assert.equal(url,'/api/explore/counted-values');acquisitions++;const question=JSON.parse(options.body);return {ok:true,status:200,json:async()=>({version:'goo/counted-values/v1',question,state:'observed',count:'3',groups:[{g0:null,count:'1'},{g0:'',count:'1'},{g0:'canary',count:'1'}],capturedAt:'2026-11-13T00:02:03Z',basis:'held-counts',reads:{count:{readAt:'2026-11-13T00:02:02Z',basis:'a'.repeat(64)},groups:{readAt:'2026-11-13T00:02:03Z',basis:'b'.repeat(64)}}})}};
 const editor={setValueLoader:value=>loader=value,setCountedValues:(value,lease)=>{if(value?.state==='ready'){ready=value;authority=lease}}},owner=Values.mount(editor,{current:()=>now});
 await loader(s.Field);assert.equal(acquisitions,1);assert(authority.current());assert.deepEqual(ready.question,Values.query(s.Scope,s.Field));const old=authority;
 if(mode==='scope'){now={...now,toMs:now.toMs+1};assert.equal(old.current(),false,'live scope withdrawal must refuse before a sync notification');owner.sync();}
 if(mode==='replace'){await loader({kind:s.Field.kind==='resource'?'attribute':'resource',key:'other'});assert(authority.current());assert.equal(acquisitions,2);}
 if(mode==='aba'){now={...now,toMs:now.toMs+1};owner.sync();now=clone(s.Scope);assert.equal(old.current(),false,'observed ABA must not revive the retired read generation');}
 assert.equal(old.current(),false);owner.dispose();assert.equal(loader,null);
}
process.stdout.write(JSON.stringify(cases.length));
