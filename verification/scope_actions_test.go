package verification

import (
	"hegel.dev/go/hegel"
	"testing"
)

const scopeActionScheduleScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {createSourceAdapter,importSnapshot,selectRows} from '../ui/snapshots.mjs';
import {createActionContext} from '../ui/scope-actions.mjs';
import {groupIntent,scopeIntent} from '../ui/view-model.mjs';
const {snapshot:raw,question,seed=1,length=20}=JSON.parse(fs.readFileSync(0,'utf8'));
let unexpectedReads=0;globalThis.fetch=()=>{unexpectedReads++;throw Error('Resolver cannot read sources')};
const source=createSourceAdapter({id:'state-'+seed}).adoptSnapshot(raw,{question});
let epoch=1,active=true,canStage=true,serial=0,current=null;const bindings=[];
const context=createActionContext({epoch:()=>epoch,active:()=>active,canStage:()=>canStage});
const bind=()=>{const binding=context.bind({role:'current',slot:'groups',epoch,snapshot:source,question,coherent:true});serial++;const saved={binding,epoch,serial,intent:groupIntent(source,source.rows[seed%source.rows.length].key)};bindings.push(saved);current=saved;return saved};
const attempt=saved=>{const before=JSON.stringify(source),result=context.resolve(saved.binding,saved.intent),expected=active&&canStage&&saved.epoch===epoch&&saved.serial===serial;assert.equal(result.ok,expected,JSON.stringify({epoch,active,canStage,saved:saved.epoch,serial,owned:saved.serial,result}));if(result.ok){assert(Object.isFrozen(result.question));assert.equal(result.question.environment,question.environment);assert.equal(result.question.fromMs,question.fromMs);assert.equal(result.question.toMs,question.toMs);assert.deepEqual(result.question.numeric,question.numeric);assert.equal(result.question.offset,0);assert(JSON.stringify(result.question.where).includes(source.rows[seed%source.rows.length].values.category));}assert.equal(JSON.stringify(source),before);return result};
const first=bind();assert(attempt(first).ok);
// Rebinding the IDENTICAL admitted object, id, revision and selector fences the
// old execution. Content equality cannot be used as execution ownership.
const second=bind();assert(!attempt(first).ok);assert(attempt(second).ok);epoch++;const third=bind();assert(!attempt(second).ok);assert(attempt(third).ok);
let rng=(seed+1)>>>0;const draw=n=>{rng=(Math.imul(rng,1664525)+1013904223)>>>0;return rng%n};
for(let i=0;i<length;i++){
 switch(draw(7)){
 case 0:bind();break;
 case 1:epoch++;current=null;break;
 case 2:context.revoke();serial++;current=null;break;
 case 3:active=!active;break;
 case 4:canStage=!canStage;break;
 case 5:attempt(bindings[draw(bindings.length)]);break;
 case 6:if(current)attempt(current);break;
 }
 for(const saved of bindings)attempt(saved);
}
epoch++;active=true;canStage=true;const final=bind();assert(attempt(final).ok);
const baseline=context.bind({role:'baseline',slot:'groups',epoch,snapshot:source,question,coherent:true});assert(!context.resolve(baseline,groupIntent(source,source.rows[0].key)).ok);assert(attempt(final).ok);
const foreign=createActionContext({epoch:()=>epoch,active:()=>true,canStage:()=>true});assert(!foreign.resolve(final.binding,final.intent).ok);assert(!context.resolve({},final.intent).ok);assert(!context.resolve(structuredClone(final.binding),final.intent).ok);
const held=importSnapshot(source);const passive=context.bind({role:'current',slot:'groups',epoch,snapshot:held,question,coherent:true});assert(!context.resolve(passive,groupIntent(held,held.rows[0].key)).ok);
assert.throws(()=>context.bind({role:'current',slot:'groups',epoch,snapshot:structuredClone(source),question,coherent:true}));
assert.equal(unexpectedReads,0);process.stdout.write('ok');
`

func TestHegelScopeActionEpochSchedules(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		length := hegel.Draw(tc, hegel.Integers(10, 35))
		raw, q, err := snapshotSourceFixture(seed)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed, "length": length}, scopeActionScheduleScript); err != nil {
			tc.Fatal(err)
		}
	}, options(70)...)
}

const scopeActionMembershipScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {createSourceAdapter,importSnapshot,selectRows} from '../ui/snapshots.mjs';import {createActionContext} from '../ui/scope-actions.mjs';import {groupIntent,scopeIntent} from '../ui/view-model.mjs';
const {snapshot:raw,question,seed=0}=JSON.parse(fs.readFileSync(0,'utf8'));let reads=0;globalThis.fetch=()=>{reads++;throw Error('Local actions cannot fetch')};
const source=createSourceAdapter({id:'membership-'+seed}).adoptSnapshot(raw,{question});let epoch=1;
const context=createActionContext({epoch:()=>epoch,active:()=>true,canStage:()=>true});const root=context.bind({role:'current',slot:'groups',epoch,snapshot:source,question,coherent:true});
const selected=source.rows[seed%source.rows.length],other=source.rows[(seed+1)%source.rows.length];const intent=groupIntent(source,selected.key),before=JSON.stringify(source);
assert(context.resolve(root,intent).ok);assert(context.resolve(root,structuredClone(intent)).ok); // Data intents are claims, not authentication nonces.
let getterCalls=0;const accessor=structuredClone(intent);Object.defineProperty(accessor,'key',{enumerable:true,get(){getterCalls++;return selected.key}});assert(!context.resolve(root,accessor).ok);assert.equal(getterCalls,0);const cyclic=structuredClone(intent);cyclic.bad=cyclic;assert(!context.resolve(root,cyclic).ok);
for(const forged of [ {...intent,key:'foreign'}, {...intent,revision:intent.revision+'x'}, {...intent,scopeId:'foreign'}, {...intent,generation:'foreign'}, {...intent,extra:true}, {...intent,values:[{field:'category',op:'eq',value:'not-present'}]}, {...intent,values:[{field:'count',op:'eq',value:selected.values.count}]} ])assert(!context.resolve(root,forged).ok);
const facet=scopeIntent(source,'category',selected.values.category);assert(context.resolve(root,facet).ok);assert(!context.resolve(root,{...facet,value:'foreign'}).ok);assert(!context.resolve(root,{...facet,field:'count',value:selected.values.count}).ok);
const subset=selectRows(source,[selected.key]),child=context.project(root,subset);assert(context.resolve(child,groupIntent(subset,selected.key)).ok);assert(!context.resolve(child,groupIntent(source,other.key)).ok);assert(!context.resolve(child,{...groupIntent(subset,selected.key),key:other.key}).ok);
assert.throws(()=>context.project(root,importSnapshot(subset)));assert.throws(()=>context.project(root,structuredClone(subset)));
const sibling=context.project(root,selectRows(source,[other.key]));assert(context.resolve(sibling,groupIntent(sibling.snapshot,other.key)).ok);assert(context.resolve(child,groupIntent(subset,selected.key)).ok);assert.throws(()=>context.project(child,selectRows(subset,[])));
const replacement=context.bind({role:'current',slot:'groups',epoch,snapshot:source,question,coherent:true});assert(!context.resolve(root,intent).ok);assert(!context.resolve(child,groupIntent(subset,selected.key)).ok);assert(context.resolve(replacement,intent).ok);
const detached=context.bind({role:'current',slot:'groups',epoch,snapshot:subset,question,coherent:true});assert(!context.resolve(detached,groupIntent(subset,selected.key)).ok);
const incoherent=context.bind({role:'current',slot:'groups',epoch,snapshot:source,question,coherent:false});assert(!context.resolve(incoherent,intent).ok);
epoch++;assert(!context.resolve(incoherent,intent).ok);assert.throws(()=>context.bind({role:'current',slot:'groups',epoch:epoch-1,snapshot:source,question,coherent:true}));
assert.equal(JSON.stringify(source),before);assert.equal(reads,0);process.stdout.write('ok');
`

func TestHegelScopeActionMembership(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 10000))
		raw, q, err := snapshotSourceFixture(seed)
		if err != nil {
			tc.Fatal(err)
		}
		if _, err = runSnapshotJS(raw, q, map[string]any{"seed": seed}, scopeActionMembershipScript); err != nil {
			tc.Fatal(err)
		}
	}, options(30)...)
}
