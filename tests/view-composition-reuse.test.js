const {test}=require('node:test'),assert=require('node:assert/strict');
const C=require('../ui/view-composition.mjs'),D=require('../ui/view-declaration.mjs');
const q=()=>({signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:1,offset:0,groupBy:[{kind:'column',key:'service'}]});
const binding=()=>({field:{kind:'attribute',key:'work.delta'},type:'int64',unit:'items',parse:'canonical-decimal'});
const make=source=>C.create(source,D.create(source),{columns:['utc','summary'],fieldColumns:[{kind:'attribute',key:'team.unreported'}]});

test('whole reuse detaches ordered and hidden child intent without changing destination scope or paging',()=>{
 const source=q(),composition=C.setHidden(C.reverse(make(source),source),'aggregate',true,source);
 C.aggregate(composition).display.requiresComplete=true;
 const destination={...q(),environment:'',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},fromMs:3000,toMs:5000,limit:3,offset:2,where:{field:{kind:'attribute',key:'team'},op:'eq',value:'other'}};
 const before=structuredClone({source,destination,composition}),next=C.reuse(composition,source,destination);
 assert.deepEqual(next,composition);assert.deepEqual(next.children.map(C.role),['records','aggregate']);assert.equal(C.visible(next,'aggregate'),false);
 C.records(next).display.columns[0]='context';C.records(next).display.fieldColumns[0].key='changed';C.records(next).kind.name='changed';C.aggregate(next).display.title='Changed';next.children.reverse();
 assert.deepEqual({source,destination,composition},before);assert.deepEqual(Object.keys(next).sort(),['children','kind','version']);
});

test('whole reuse requires the selected aggregate quantity but not unused bindings or observed record attributes',()=>{
 const source={...q(),numeric:binding(),summary:'span-duration'};
 for(const value of D.availableMeasures(source)){
  const composition=make(source);C.aggregate(composition).display.value=value;C.aggregate(composition).display.scale='max';
  const destination={...structuredClone(source),environment:'staging',fromMs:3000,toMs:4000,limit:7};destination.numeric.scale=0;
  assert.deepEqual(C.reuse(composition,source,destination),composition,'omitted scale equals explicit zero');
  destination.numeric.unit='bytes';
  if(value.startsWith('numeric.'))assert.throws(()=>C.reuse(composition,source,destination),/same field, type, scale, unit and parse/);
  else assert.deepEqual(C.reuse(composition,source,destination),composition);
 }
});

test('whole reuse rejects unsupported scopes, fields, queries and runtime payloads before returning intent',()=>{
 const source={...q(),numeric:binding(),summary:'span-duration'},composition=make(source);C.aggregate(composition).display.value='numeric.mean';C.aggregate(composition).display.scale='max';
 for(const change of [x=>{x.signal='logs';delete x.summary;delete x.numeric},x=>x.groupBy=[{kind:'attribute',key:'service'}],x=>x.groupBy=[],x=>x.toMs=x.fromMs,x=>delete x.numeric,x=>x.numeric.field.key='other',x=>x.numeric.unit='bytes']){const destination=structuredClone(source);change(destination);assert.throws(()=>C.reuse(composition,source,destination));}
 assert.throws(()=>C.reuse(composition,{...source,toMs:source.fromMs},source));
 for(const change of [x=>x.children.forEach(c=>c.hidden=true),x=>C.records(x).display.fieldColumns=[{kind:'column',key:'body'}],x=>C.records(x).scope.signal='logs',x=>x.rows=[],x=>x.children[1].resultToken='runtime',x=>C.records(x).display.completeness='complete']){const bad=structuredClone(composition);change(bad);assert.throws(()=>C.reuse(bad,source,source));}
});
