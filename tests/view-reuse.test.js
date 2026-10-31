const {test}=require('node:test'),assert=require('node:assert/strict'),D=require('../ui/view-declaration.mjs');
const q={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}]};
const binding={field:{kind:'attribute',key:'work.delta'},type:'int64',unit:'items',parse:'canonical-decimal'};
test('reuse detaches display intent and keeps destination environment, predicates and window',()=>{
 const source=structuredClone(q),d=D.create(source);d.display.title='Requests';d.display.query='api';d.display.requiresComplete=true;
 const destination={...structuredClone(q),environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},fromMs:3000,toMs:5000,limit:1,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'}};
 const before=structuredClone({source,destination,d}),next=D.reuse(d,source,destination);
 assert.deepEqual(next,d);next.display.title='Independent draft';assert.deepEqual({source,destination,d},before);assert.deepEqual(Object.keys(next).sort(),['display','kind','scope','version']);
});
test('reuse requires selected quantity binding while count and duration ignore unrelated numeric data',()=>{
 const source={...q,numeric:structuredClone(binding),summary:'span-duration'};
 for(const value of D.availableMeasures(source)){
  const d=D.create(source);d.display.value=value;d.display.scale='max';
  const destination={...structuredClone(source),environment:'staging',fromMs:3000,toMs:4000};destination.numeric.scale=0;
  assert.deepEqual(D.reuse(d,source,destination),d,'omitted integer scale equals explicit zero');
  destination.numeric.unit='bytes';
  if(value.startsWith('numeric.'))assert.throws(()=>D.reuse(d,source,destination),/same field, type, scale, unit and parse/);
  else assert.deepEqual(D.reuse(d,source,destination),d);
 }
});
test('reuse refuses incompatible signal, qualified grouping, measure and invalid queries',()=>{
 const source={...q,numeric:structuredClone(binding),summary:'span-duration'},d=D.create(source);d.display.value='numeric.mean';d.display.scale='max';
 for(const change of [x=>x.signal='logs',x=>x.groupBy=[{kind:'attribute',key:'service'}],x=>x.groupBy=[{kind:'column',key:'name'}],x=>x.groupBy=[],x=>x.toMs=x.fromMs,x=>delete x.numeric,x=>x.numeric.field.kind='resource',x=>x.numeric.field.key='other',x=>x.numeric.type='uint64',x=>x.numeric.unit='bytes',x=>x.numeric.parse='lenient',x=>{x.numeric.type='decimal64';x.numeric.scale=2}]){const destination=structuredClone(source);change(destination);assert.throws(()=>D.reuse(d,source,destination));}
 const duration={...structuredClone(d),display:{...d.display,value:'duration.p95'}},destination={...source};delete destination.summary;assert.throws(()=>D.reuse(duration,source,destination));
 const invalid={...source,toMs:source.fromMs};assert.throws(()=>D.reuse(d,invalid,source));
});

test('reuse refuses span-only summary metadata on logs even for a count display',()=>{
 const logs={...q,signal:'logs'},d=D.create(logs);
 for(const extra of [{summary:'span-duration'},{spanStatus:true}]){assert.throws(()=>D.reuse(d,{...logs,...extra},logs));assert.throws(()=>D.reuse(d,logs,{...logs,...extra}));}
});
