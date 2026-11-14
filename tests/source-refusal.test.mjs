import {renderRefusal as render} from '../ui/source-refusal-view.mjs';
import test from 'node:test';
import assert from 'node:assert/strict';
import {SourceReadError,reasons,refusal,responseError,failure} from '../ui/source-refusal.mjs';
class Element{
 constructor(tag){this.tagName=tag;this.children=[];this.dataset={};this.open=false;this.focused=false;this.attributes=new Map();}
 setAttribute(key,value){this.attributes.set(key,String(value));}getAttribute(key){return this.attributes.get(key)??null;}
 append(...children){this.children.push(...children)}replaceChildren(...children){this.children=children;this.textContent='';}focus(){this.focused=true;}
}
test('source wire categories retain canonical reasons and bounded typed details without request-error classification',()=>{
 for(const kind of Object.keys(reasons)){
  const data={error:reasons[kind],refusal:{kind,detail:'A bounded source diagnostic.'}},before=structuredClone(data),error=responseError({ok:false,status:502,data});
  assert(error instanceof SourceReadError);assert.equal(error.message,reasons[kind]);assert.deepEqual(failure(error),data);assert.deepEqual(data,before);error.refusal.detail='changed';assert.equal(data.refusal.detail,before.refusal.detail);
 }
 for(const status of [400,409,429]){const error=responseError({status,data:{error:'Validation mentions timeout but is not a source refusal.'}});assert.equal(error instanceof SourceReadError,false);assert.equal(error.message,'Validation mentions timeout but is not a source refusal.');const shaped=responseError({status,data:{error:reasons.budget,refusal:{kind:'budget',detail:'A diagnostic.'}}});assert.equal(shaped instanceof SourceReadError,false);}
});
test('unrecognized or malformed envelopes use an explicit generic source refusal and never promote raw transport text',()=>{
 for(const change of [{kind:'future',detail:'x'},{kind:'timeout',detail:'x'.repeat(257)},{kind:'timeout',detail:''},{kind:'timeout',detail:'bad\ncontrol'},{kind:'timeout',detail:'x',retry:'automatic'},null]){
  const error=responseError({status:502,data:{error:'GET https://credentials@example.invalid/?sql=private',refusal:change}});assert(error instanceof SourceReadError);assert.equal(error.refusal.kind,'unavailable');assert.equal(error.message,reasons.unavailable);assert.doesNotMatch(error.message,/https|GET|private/);
 }
 const unknown=responseError({status:502,data:{error:'raw upstream URL'}});assert.equal(unknown.message,reasons.unavailable);assert.throws(()=>refusal({kind:'budget',detail:'🙂'.repeat(65)}));assert.doesNotThrow(()=>refusal({kind:'budget',detail:'🙂'.repeat(64)}));
});
test('passive refusal rendering shows a terse reason, collapsed literal detail and performs no IO or focus changes',()=>{
 const previousDocument=globalThis.document,previousFetch=globalThis.fetch;let reads=0;globalThis.document={createElement:tag=>new Element(tag)};globalThis.fetch=()=>{reads++;throw Error('Passive refusal cannot fetch');};
 try{
  const host=new Element('div'),detail='<script>literal diagnostic</script>',error=new SourceReadError({kind:'malformed',detail}),shown=render(host,error,{prefix:'Previous values unavailable · '});
  assert.equal(shown.dataset.sourceRefusalKind,'malformed');assert.equal(shown.state,'unavailable');assert.equal(shown.children[0].textContent,'Previous values unavailable · '+reasons.malformed);const details=shown.children[1];assert.equal(details.tagName,'details');assert.equal(details.open,false);assert.equal(details.children[0].textContent,'Technical details');assert.equal(details.children[1].textContent,detail);assert.equal(reads,0);assert.equal(shown.focused,false);const announced=new Element('div');announced.setAttribute('role','status');const grouped=render(announced,error);assert.equal(grouped.getAttribute('role'),'group');assert.equal(grouped.getAttribute('announce'),'off');
  const ordinary=new Element('div');assert.equal(render(ordinary,{error:'Choose a valid field.'}),null);assert.equal(ordinary.textContent,'Choose a valid field.');
 }finally{globalThis.document=previousDocument;globalThis.fetch=previousFetch;}
});

test('shared labels remain canonical by default while a host can supply a presentation-only fixed label',async()=>{
 const previous=globalThis.document;globalThis.document={createElement:tag=>new Element(tag)};
 try{
  for(const kind of Object.keys(reasons)){const error=new SourceReadError({kind,detail:'A bounded source diagnostic.'}),host=new Element('div');assert.equal(render(host,error).children[0].textContent,reasons[kind]);}
  const {gap,malformed,sourceError,title}=await import('../ui/tail-recovery.mjs');
  for(const [reason,detail,label]of [['source_missing','The requested record source is missing.','Source table missing.'],['source_access','The source refused access to the requested records.','Access denied.'],['source_schema','The record source does not match the supported exporter schema.','Source schema incompatible.']]){
   const outcome=gap({reason,retry:'explicit',refusal:{kind:'unavailable',detail}}),shown=render(new Element('div'),sourceError(outcome),{label:title(outcome)});assert.equal(shown.children[0].textContent,label);assert.equal(shown.children[1].children[1].textContent,detail);assert.equal(shown.dataset.sourceRefusalKind,'unavailable');
  }
  const unknown=gap({reason:'raw SQL password',retry:'explicit',refusal:{kind:'unavailable',detail:'raw SQL password'}});assert.deepEqual(unknown,malformed());assert.equal(title({reason:'raw SQL password'}),'Source response rejected.');const shown=render(new Element('div'),sourceError(unknown),{label:title(unknown)});assert.equal(shown.children[0].textContent,'Source response rejected.');assert.doesNotMatch(shown.children[1].children[1].textContent,/SQL|password/);
 }finally{globalThis.document=previous;}
});
