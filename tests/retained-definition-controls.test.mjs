import test from 'node:test';
import assert from 'node:assert/strict';
import {retainedDefinitionControls,definitionControlState,restoreDefinitionControlState} from '../ui/retained-definition-controls.mjs';
import {campaign} from '../verification/fixtures/parent-return-342.mjs';
test('production local handlers rotate across repeated Back, hidden, detached and same-text replacement schedules',()=>{
 for(const seed of [0,1,42,0xffffffff])for(const rounds of [1,2,11])for(const duplicates of [1,4]){
  const result=campaign({rounds,duplicates,seed});assert.equal(result.legitimateEffects,1+rounds*duplicates);assert.equal(result.retiredEffects,0);assert.equal(result.failedRenewals,rounds*4);assert.equal(result.reads,0);assert.equal(result.draftPreserved,true);assert.equal(result.receiptPreserved,true);
 }
});
test('old bound methods preserve this/args only while their immutable local owner remains current',()=>{
 let generation=1,eligible=true;const lease=retainedDefinitionControls(()=>({generation,eligible})),calls=[],action=function(value){calls.push([this,value]);return value;},host={};
 const old=lease.bind(action);assert.equal(old.call(host,'first'),'first');generation++;assert.equal(old.call(host,'retired'),undefined);assert.equal(lease.renew(),true);const fresh=lease.bind(action);old.call(host,'still retired');assert.equal(fresh.call(host,'current'),'current');assert.deepEqual(calls,[[host,'first'],[host,'current']]);eligible=false;assert.equal(lease.renew(),false);fresh('hidden');assert.equal(calls.length,2);lease.retire();eligible=true;fresh('reattached');assert.equal(calls.length,2);
});
test('remount state retains raw text, checkbox, exact closed disclosure and caret without copying owner disabled state',()=>{
 const field=(tag,value,extra={})=>({tagName:tag,type:tag==='INPUT'?'text':undefined,value,getAttribute:()=>null,...extra});
 const old=[field('INPUT','not yet applied',{checked:false,selectionStart:2,selectionEnd:7,selectionDirection:'backward'}),field('DETAILS',undefined,{open:false})],next=[field('INPUT','initial',{disabled:true,setSelectionRange(start,end,direction){this.range=[start,end,direction]}}),field('DETAILS',undefined,{open:true})];
 const state=definitionControlState({querySelectorAll:()=>old});restoreDefinitionControlState({querySelectorAll:()=>next},state);assert.equal(next[0].value,'not yet applied');assert.equal(next[0].checked,false);assert.equal(next[0].disabled,true);assert.deepEqual(next[0].range,[2,7,'backward']);assert.equal(next[1].open,false);
});
