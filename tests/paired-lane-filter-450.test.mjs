import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {validate} from '../ui/predicate-validation.mjs';
let editors=[];
mock.module('../ui/filter-editor.mjs',{exports:{normalize:value=>structuredClone(value),mount:(host,options)=>{
 let value,cel=false;const calls={reset:0,retire:0,dispose:0};
 const editor={options,calls,edit(next,error=null){value=structuredClone(next);options.onChange(next,error)},cel(value){cel=value},snapshot:()=>structuredClone(value),hasUnappliedCELDraft:()=>cel,hasCELDraft:()=>cel,enableCEL(config){editor.config=config},syncCEL(){},resetCEL(){calls.reset++;cel=false},retireCEL(){calls.retire++;cel=false},dispose(){calls.dispose++},set(next){value=structuredClone(next)},setSuggestions(){}};
 editors.push(editor);return editor;
}}});
const {mount}=await import('../ui/paired-lane-filter.mjs');
const leaf=value=>({field:{kind:'column',key:'severity_number'},op:'gte',value});
function setup(onApply=()=>true){
 globalThis.Node=Element;const body=new Element('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};
 const host=new Element('main');body.append(host);let active=true;const applied=[];
 const owner=mount(host,{signal:'logs',current:()=>active,base:()=>({signal:'logs',environment:'prod',fromMs:1,toMs:2}),validate:value=>validate(value,'logs'),onApply:(...args)=>{applied.push(args);return onApply(...args)}});
 const details=host.querySelector('details');details.open=true;
 const buttons=host.querySelectorAll('button'),editor=editors.at(-1);
 return {host,owner,editor,apply:buttons[0],discard:buttons[1],applied,deactivate:()=>active=false,activate:()=>active=true};
}
test('450 lane draft is local until explicit Apply and applied state is detached',()=>{
 const h=setup();h.editor.edit(leaf('17'));assert.equal(h.applied.length,0);assert.equal(h.owner.dirty(),true);assert.equal(h.apply.disabled,false);
 h.apply.focus();assert.equal(h.apply.onclick(),true);assert.equal(document.activeElement,h.host.querySelector('summary'));h.host.querySelector('details').open=false;assert.equal(h.host.querySelector('details').contains(h.host.querySelector('small')),false);assert.match(h.host.querySelector('small').textContent,/^Applied/);assert.equal(h.applied.length,1);assert.deepEqual(h.owner.where(),leaf('17'));assert.equal(h.owner.dirty(),false);
 h.applied[0][0].value='20';assert.deepEqual(h.owner.where(),leaf('17'));
});
test('450 invalid and unapplied CEL drafts cannot widen an applied condition; Discard preserves it',()=>{
 const h=setup();h.owner.reset(leaf('17'));h.editor.edit(undefined,'Invalid condition');assert.equal(h.apply.disabled,true);assert.equal(h.apply.onclick(),false);assert.deepEqual(h.owner.where(),leaf('17'));assert.equal(h.applied.length,0);
 h.discard.onclick();assert.deepEqual(h.editor.snapshot(),leaf('17'));assert.equal(h.owner.dirty(),false);
 h.editor.cel(true);h.owner.sync();assert.equal(h.apply.disabled,true);assert.equal(h.apply.onclick(),false);assert.deepEqual(h.owner.where(),leaf('17'));
 h.discard.onclick();assert.equal(h.editor.hasCELDraft(),false);assert.deepEqual(h.owner.where(),leaf('17'));
});
test('450 hidden, retired owner, disabled and closed disclosure Apply callbacks refuse',()=>{
 for(const withdraw of [h=>h.host.hidden=true,h=>h.deactivate(),h=>h.apply.disabled=true,h=>h.host.querySelector('details').open=false]){
  const h=setup();h.editor.edit(leaf('17'));const callback=h.apply.onclick;withdraw(h);assert.equal(callback(),false);assert.equal(h.applied.length,0);assert.equal(h.owner.where(),undefined);
 }
 const h=setup();h.editor.edit(leaf('17'));const callback=h.apply.onclick;h.owner.dispose();assert.equal(callback(),false);assert.equal(h.applied.length,0);assert.equal(h.editor.calls.dispose,1);
});
test('450 synchronous source withdrawal or draft edit prevents adoption and retires captured Apply lease',()=>{
 for(const withdraw of [h=>h.deactivate(),h=>h.editor.edit(leaf('19')),h=>h.owner.retire()]){
  let h;h=setup((value,opener,capability)=>{assert.equal(capability.current(),true);withdraw(h);assert.equal(capability.current(),false);return true});
  h.editor.edit(leaf('17'));assert.equal(h.apply.onclick(),false);assert.equal(h.owner.where(),undefined);assert.equal(h.applied.length,1);
 }
});
test('450 scope retirement invalidates CEL work and reset adopts only explicit applied condition',()=>{
 const h=setup();h.editor.edit(leaf('17'));h.owner.retire();assert.equal(h.editor.calls.retire,1);
 h.deactivate();assert.equal(h.editor.config.current(),false);assert.equal(h.editor.options.canMutate(),false);
 h.owner.reset(leaf('20'));assert.equal(h.owner.dirty(),false);assert.deepEqual(h.owner.where(),leaf('20'));assert.equal(h.editor.calls.reset,1);
});
