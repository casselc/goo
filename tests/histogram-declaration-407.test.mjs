import test from 'node:test';
import assert from 'node:assert/strict';
import * as Histogram from '../ui/histogram-declaration.mjs';
import * as Composition from '../ui/view-composition.mjs';
import * as Templates from '../ui/view-template-handoff.mjs';
const question=signal=>({signal,environment:'source',fromMs:1000,toMs:6000,limit:2,offset:0,numeric:{field:{kind:'attribute',key:'amount'},type:'decimal64',scale:2,unit:'USD',parse:'canonical-decimal'},numericHistogram:{binning:'powers-of-two'}});
test('histogram declarations reuse only presentation while preserving destination binding and full scope',()=>{
 for(const signal of ['logs','spans']){const q=question(signal),declaration=Histogram.create(q),composition=Composition.createHistogram(q,declaration),source={version:1,title:'Value distribution',queries:[q],composition,template:Templates.templateFor('ordinary-histogram')};
  const destination={...structuredClone(q),environment:'destination',fromMs:11000,toMs:19000,limit:7,where:{field:{kind:'attribute',key:'team'},op:'eq',value:'destination'}},before=JSON.stringify(source),reused=Composition.reuse(composition,q,destination);
  assert.deepEqual(reused,composition);assert.deepEqual(Templates.authoredRequest(source,destination),{source,query:destination});assert.equal(JSON.stringify(source),before);
  for(const mutate of [x=>x.numeric.field.kind='resource',x=>x.numeric.field.key='other',x=>x.numeric.type='int64',x=>x.numeric.scale=3,x=>x.numeric.unit='usd',x=>x.numeric.parse='lenient',x=>x.numericHistogram.binning='linear',x=>x.offset=1,x=>x.groupBy=[{kind:'column',key:'service'}],x=>delete x.numericHistogram]){const bad=structuredClone(destination);mutate(bad);assert.throws(()=>Templates.authoredRequest(source,bad));assert.throws(()=>Histogram.reuse(declaration,q,bad));}
 }
});
test('histogram defaults canonicalize before complete template transport comparison',()=>{
 const q=question('logs'),declaration={...Histogram.create(q),display:{category:'from',value:'count',requiresComplete:false,title:''}},source={version:1,title:'Value distribution',queries:[q],composition:Composition.createHistogram(q,declaration),template:Templates.templateFor('ordinary-histogram')};
 assert.deepEqual(Composition.histogram(source.composition).display,{category:'from',value:'count'});
 source.composition.children[0].histogram.display.requiresComplete=false;const input=Templates.authoredRequest(source,q);assert.equal(Object.hasOwn(input.source.composition.children[0].histogram.display,'requiresComplete'),false);
 const response={template:{id:'a'.repeat(64),version:1,title:source.title},view:input.source};assert.deepEqual(Templates.authoredResponse(response,source,q).composition,input.source.composition);
 const complete=Histogram.update(declaration,{kind:'table',requiresComplete:true},q);assert.equal(complete.kind.name,'table');assert.equal(complete.display.requiresComplete,true);assert.deepEqual(Histogram.presentation(complete),{kind:'table',title:'Numeric value distribution',requiresComplete:true});
});
test('histogram role remains finite and visible without normalization or executable extras',()=>{
 const q=question('logs'),d=Histogram.create(q),c=Composition.createHistogram(q,d);
 for(const field of ['normalization','bins','query','scale'])assert.throws(()=>Histogram.validate({...d,display:{...d.display,[field]:'density'}},q));
 for(const role of ['histogram','records'])assert.throws(()=>Composition.setHidden(c,role,true,q));
 assert.throws(()=>Composition.validate({...c,children:[...c.children,{summary:{}}]},q));
 assert.throws(()=>Templates.binding({...Templates.templateFor('ordinary-histogram'),baseline:'destination'}));
 assert.throws(()=>Templates.templateView({version:1,title:'Wrong profile',queries:[q],composition:c,template:Templates.templateFor('ordinary-summary')}));
 assert.equal(Composition.role(c.children[0]),'histogram');assert.equal(Composition.recordOnly(c),false);
});
