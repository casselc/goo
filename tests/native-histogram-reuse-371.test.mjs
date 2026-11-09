import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createRequire} from 'node:module';
import * as Native from '../ui/native-view-declaration.mjs';
import {buildNativeHistogramView} from '../ui/native-histogram-view.mjs';
import {createSourceAdapter,hasGuarantee} from '../ui/snapshots.mjs';
import {validateDisplay} from '../ui/view-model.mjs';
const require=createRequire(import.meta.url),{result}=require('./fixtures/native-histogram.cjs');
const clone=x=>structuredClone(x);
function source(kind='histogram'){
 const question=result().question;
 return {version:1,id:'a'.repeat(64),title:'Reusable interval bins',question,declaration:Native.create(question,kind)};
}
function destination({policy='effective',mode='named',namespace='team-雪',exact=true,unit='s'}={}){
 const q=result().question,d=q.definition,s=clone(result().histograms.series[0].stream);
 delete d.environment;d.environmentSelector={version:'goo/environment-selector/v1',policy,scope:{mode,...(mode==='named'?{value:'destination'}:{})}};
 Object.assign(d,{name:'destination-duration',title:'Destination duration',maxGapMs:20000,lookbackMinutes:2});
 Object.assign(d.source,{services:['destination-worker'],scope:'otel.destination',version:'2',schemaURL:'https://opentelemetry.io/schemas/1.37.0',metric:'destination.duration',unit,attributes:{'rpc.method':'雪/quoted "method"'}});
 d.source.namespace=namespace===null?{mode:'missing'}:{mode:'value',value:namespace};
 q.fromMs=20000;q.toMs=40000;q.fromNS='20000000123';q.toNS='39999999877';
 Object.assign(s,{scope:d.source.scope,version:d.source.version,schemaURL:d.source.schemaURL,name:d.source.metric,unit,attributes:clone(d.source.attributes)});
 s.resource={'service.name':'destination-worker','service.instance.id':'instance-destination',...(namespace===null?{}:{'service.namespace':namespace})};
 if(mode!=='unreported')s.resource[policy==='legacy-only'?'deployment.environment':'deployment.environment.name']=mode==='named'?'destination':'';
 if(exact)d.source.exactStream=s;
 return q;
}

test('native reuse transfers only display intent and preserves every destination question field',()=>{
 const previousFetch=globalThis.fetch;globalThis.fetch=()=>{throw Error('Configuration attempted telemetry I/O')};
 try{for(const kind of ['histogram','table'])for(const [policy,mode]of [['current-only','named'],['effective','named'],['effective','empty'],['effective','unreported']])for(const namespace of [null,'','team-雪']){
  const saved=source(kind),q=destination({policy,mode,namespace}),before=clone({saved,q}),next=Native.reuse(saved,q);
  assert.deepEqual(next,{version:1,title:saved.title,question:q,declaration:saved.declaration});
  assert.equal(Object.hasOwn(next,'id'),false);assert.deepEqual({saved,q},before);
  next.question.definition.source.services[0]='caller mutation';next.declaration.kind.name='caller mutation';
  assert.deepEqual({saved,q},before);
 }}finally{globalThis.fetch=previousFetch}
});

test('declarations reject record aliases, inferred interval identity, captured values and unsupported shapes',()=>{
 const saved=source(),q=destination();
 for(const change of [d=>d.version='goo.view/2',d=>d.scope.kind='record-query',d=>d.scope.signal='spans',d=>d.selection.kind='time-bucket',d=>d.kind.name='bar',d=>d.kind.version=2,d=>d.display.category='category',d=>d.display.value='count',d=>d.display.requiresComplete=false,d=>delete d.display.requiresComplete]){const bad=clone(saved.declaration);change(bad);assert.throws(()=>Native.declaration(bad,q))}
 for(const [part,key]of [['root','basis'],['root','readAt'],['root','rows'],['root','intervalIndex'],['selection','stream'],['selection','fromNS'],['selection','toNS'],['selection','index'],['display','bins'],['display','temporality'],['display','unit']]){const bad=clone(saved.declaration);(part==='root'?bad:bad[part])[key]='donor';assert.throws(()=>Native.declaration(bad,q))}
 for(const key of ['basis','readAt','result','snapshot','stream','fromNS','intervalIndex'])assert.throws(()=>Native.view({...saved,[key]:'donor'},{stored:true}));
 for(const change of [d=>d.definition.kind='counter',d=>d.definition.grain='consumer-group',d=>d.definition.histogramBucketMs=1000,d=>d.definition.source.resource={kind:'host',id:'host-a'},d=>d.definition.source.unit='By',d=>d.fromNS='19999999999']){const bad=clone(q);change(bad);assert.throws(()=>Native.reuse(saved,bad))}
 assert.throws(()=>Native.reuse(saved,destination({unit:'ms'})),/unit/);
});

test('closed response admission rejects changed display and destination fields while accepting native wire defaults',()=>{
 const saved=source('table'),q=destination({namespace:''}),expected=Native.reuse(saved,q),wire=clone(expected);
 delete wire.question.definition.source.namespace.value;wire.question.definition.environment='';wire.question.definition.source.exactStream.resourceSchemaURL='';
 assert.deepEqual(Native.response(wire,expected).declaration,expected.declaration);
 assert.equal(Native.questionIdentity(wire.question),Native.questionIdentity(q));
 for(const change of [v=>v.title='other',v=>v.question.fromNS='20000000124',v=>v.question.definition.source.attributes['rpc.method']='other',v=>v.question.definition.source.exactStream.resource['service.instance.id']='other',v=>v.question.definition.environmentSelector.policy='current-only',v=>v.declaration.kind.name='histogram',v=>v.id='b'.repeat(64),v=>v.result={basis:'donor'}]){const bad=clone(expected);change(bad);assert.throws(()=>Native.response(bad,expected))}
 const bad=clone(q);bad.definition.source.exactStream.resource['service.namespace']='other';assert.throws(()=>Native.question(bad));
});

test('stored view identity is separate from the measurement question and cannot be silently exchanged',()=>{
 const histogram=source('histogram'),table=source('table');
 assert.equal(Native.questionIdentity(histogram.question),Native.questionIdentity(table.question));
 assert.notDeepEqual(histogram.declaration,table.declaration);
 assert.throws(()=>Native.view(histogram));assert.doesNotThrow(()=>Native.view(histogram,{stored:true}));
 const wrong={...histogram,id:'b'.repeat(64)};assert.throws(()=>Native.response(wrong,histogram,{stored:true}));
 for(const id of ['',undefined,'../other','x'.repeat(64)])assert.throws(()=>Native.view({...histogram,id},{stored:true}));
});

test('complete-bin displays use the selected evaluated interval, preserving observed temporality and uncertainty',()=>{
 for(const temporality of ['delta','cumulative']){
  const r=result();r.histograms.series[0].temporality=temporality;const before=clone(r),raw=buildNativeHistogramView(r,0,1);
  assert(raw);const snapshot=createSourceAdapter({id:'native-reuse371'}).adoptSnapshot(raw,{question:raw.scope.selector});
  assert.equal(snapshot.histogram.temporality,temporality);assert.equal(snapshot.histogram.unit,'s');assert.equal(snapshot.histogram.total,'4');
  assert.equal(snapshot.quality.sampling,'unknown');assert.equal(hasGuarantee(snapshot,'bins'),true);
  for(const kind of ['histogram','table'])assert.equal(validateDisplay(snapshot,{kind,requiresComplete:true}).ok,true);
  assert.equal(buildNativeHistogramView(r,0,0),null,'unknown interval cannot inherit the known interval');
  assert.deepEqual(r,before);
 }
 const r=result();r.histograms.series[0].intervals[1].state='unknown';assert.equal(buildNativeHistogramView(r,0,1),null);
});
