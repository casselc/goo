import fs from 'node:fs';
import assert from 'node:assert/strict';
import {validateRecordQuestion} from '../../ui/record-intent.mjs';

// This bridge executes the product's pure admission function. Neither a browser
// nor source access is available; accepting intent must never acquire evidence.
globalThis.fetch=()=>{throw Error('Record-intent admission acquired a source')};
const cases=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(cases);
const results=cases.map(({name,query})=>{try{validateRecordQuestion(query);return{name,valid:true}}catch(error){return{name,valid:false,error:error.message}}});
assert.deepEqual(cases,before,'Admission mutated caller-owned intent');
const base={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0};
// These values exist only as JavaScript objects; the strict wire validator also
// refuses null and malformed Unicode. Configuration admission must reject them
// before link/editor/history code can rewrite or silently drop supplied intent.
for(const mutate of [
 q=>q.where=null,q=>q.groupBy=null,q=>q.numeric=null,q=>q.numericHistogram=null,q=>q.reportedEdge=null,
 q=>q.environmentSelector=null,q=>q.offset=null,q=>q.limit=null,
 q=>q.where={all:new Array(1)},
 q=>{const list=[{field:{kind:'column',key:'service'},op:'eq',value:'api'}];list.extra='do-not-ignore';q.where={all:list}},
 q=>{delete q.environment;q.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'},[Symbol('hidden')]:true}},
 q=>{delete q.environment;q.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:Object.assign(Object.create({poison:true}),{mode:'empty'})}},
 q=>q.where={field:{kind:'attribute',key:'\ud800'},op:'eq',value:'ok'},
 q=>q.where={field:{kind:'attribute',key:'tag'},op:'eq',value:'\udfff'},
 q=>q.where={all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'}],op:''},
 q=>q.where={field:{kind:'column',key:'service'},op:'eq',value:'api',[Symbol('hidden')]:true},
 q=>q.where=Object.assign(Object.create({poison:'inherited'}),{field:{kind:'column',key:'service'},op:'eq',value:'api'}),
 q=>{const cycle={all:[]};cycle.all.push(cycle);q.where=cycle;}
]){
 const q=structuredClone(base);mutate(q);assert.throws(()=>validateRecordQuestion(q),'Unsafe noncanonical configuration was accepted');
}
let getterCalls=0;
const accessor=structuredClone(base);Object.defineProperty(accessor,'where',{enumerable:true,get(){getterCalls++;return {field:{kind:'column',key:'service'},op:'eq',value:'api'}}});
assert.throws(()=>validateRecordQuestion(accessor),'Accessor configuration was accepted');
assert.equal(getterCalls,0,'Admission executed a caller accessor');
const environmentAccessor=structuredClone(base);delete environmentAccessor.environment;
environmentAccessor.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}};
Object.defineProperty(environmentAccessor.environmentSelector.scope,'value',{enumerable:true,get(){getterCalls++;return ''}});
assert.throws(()=>validateRecordQuestion(environmentAccessor),'Nested environment accessor was accepted');
assert.equal(getterCalls,0,'Admission executed an environment accessor');
process.stdout.write(JSON.stringify(results));
