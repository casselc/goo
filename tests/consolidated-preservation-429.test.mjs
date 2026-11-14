import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';
import {captureDestination} from '../ui/capture-destination.mjs';
import {nativeControlReady} from '../ui/dom-visibility.mjs';
const source=await readFile(new URL('../ui/captures.mjs',import.meta.url),'utf8');
const exploration=await readFile(new URL('../ui/exploration.mjs',import.meta.url),'utf8');

test('429 destination presentation preserves explicit eligible links and clears hidden or missing choices',()=>{
 const questions=[{key:'exact',label:'Investigated question',step:{kind:'question'}}],before=structuredClone(questions);
 for(const target of ['', 'saved-revision']){const state=captureDestination({target,questions,questionKey:'exact'});assert.equal(state.showQuestion,false);assert.equal(state.questionKey,'');assert.equal(state.selected,undefined);assert.equal(!!state.note,!!target);}
 const linked=captureDestination({target:'current',questions,questionKey:'exact'});assert.equal(linked.showQuestion,true);assert.equal(linked.selected,questions[0]);assert.equal(linked.questionKey,'exact');assert.match(linked.note,/save its revision/);
 assert.equal(captureDestination({target:'current',questions,questionKey:'retired'}).questionKey,'');assert.equal(captureDestination({target:'current'}).showQuestion,false);assert.deepEqual(questions,before);
});

test('429 actual destination rendering removes irrelevant native fields and clears stale links before confirmation',()=>{
 const controls=new Map(),el=id=>{if(!controls.has(id))controls.set(id,{value:'',textContent:'',hidden:false});return controls.get(id)},questionLabel={},questionTarget={value:'exact'},questionScope={},questionContext={questions:[{key:'exact',label:'Question',step:{version:1}}]},draft={kind:'comparison'};let comparisons=0;
 const body=source.match(/function questionDestination\(\)\{[^\n]+\}/)[0];
 const update=new Function('captureDestination','el','questionLabel','questionTarget','questionScope','questionContext','Workspace','draft','return ('+body+');')(captureDestination,el,questionLabel,questionTarget,questionScope,questionContext,{evidenceDifference(){comparisons++;return[]}},draft);
 el('captureDraftInvestigation').value='current';update();assert.equal(questionLabel.hidden,false);assert.equal(questionTarget.disabled,false);assert.equal(questionScope.hidden,false);assert.equal(comparisons,1);
 el('captureDraftInvestigation').value='';update();assert.equal(questionLabel.hidden,true);assert.equal(questionTarget.disabled,true);assert.equal(questionTarget.value,'');assert.equal(questionScope.hidden,true);assert.equal(el('captureDraftDestinationNote').hidden,true);assert.equal(comparisons,1);
 el('captureDraftInvestigation').value='saved-revision';update();assert.equal(questionLabel.hidden,true);assert.match(el('captureDraftDestinationNote').textContent,/new revision/);
 el('captureDraftInvestigation').value='current';update();assert.equal(questionLabel.hidden,false);assert.equal(questionTarget.value,'');assert.equal(questionScope.hidden,true,'returning to current does not silently restore the previous question association');
});

test('429 actual quick-filter callbacks refuse closed, disabled, hidden and detached controls without editing or reading',()=>{
 globalThis.document={hidden:false};const details={open:true,parentElement:null,closest(){return null},querySelector(){return {contains:()=>false}}};
 const controls=new Map();function control(){return {isConnected:true,disabled:false,hidden:false,parentElement:details,getClientRects:()=>[1],getAttribute:()=>null,matches(selector){return selector===':disabled'&&this.disabled},closest(selector){if(selector.includes('[hidden]')&&this.hidden)return this;if(selector==='details:not([open])'&&!details.open)return details;return null}}}
 const el=id=>{if(!controls.has(id))controls.set(id,control());return controls.get(id)};let edits=0,marks=0,reads=0,owner=true;
 const bindings=exploration.split('\n').find(line=>line.includes("for(const id of ['discoveryService','discoveryMessage'")&&line.includes('.oninput='));
 new Function('el','rootEditorCurrent','nativeControlReady','change','markFilters',bindings)(el,()=>owner,nativeControlReady,()=>edits++,()=>marks++);
 const b=control(),binding=exploration.match(/b\.onclick=\(\)=>\{if\(rootEditorCurrent\(\)&&nativeControlReady\(b\)\)breakdown\(\{kind,key\}\)\}/)[0];
 new Function('b','rootEditorCurrent','nativeControlReady','breakdown','kind','key',binding)(b,()=>owner,nativeControlReady,()=>reads++,'column','service');
 const inputs=['discoveryService','discoveryMessage','discoverySeverity','discoveryDuration'].map(el),error=el('discoveryErrors');
 const invoke=()=>{for(const input of inputs)input.oninput();error.onchange();b.onclick()};
 invoke();assert.equal(edits,5);assert.equal(marks,5);assert.equal(reads,1);
 for(const state of ['closed','disabled','hidden','detached','departed']){details.open=state!=='closed';owner=state!=='departed';for(const node of [...inputs,error,b]){node.disabled=state==='disabled';node.hidden=state==='hidden';node.isConnected=state!=='detached'}invoke();assert.equal(edits,5,state);assert.equal(reads,1,state);}
 delete globalThis.document;
});
