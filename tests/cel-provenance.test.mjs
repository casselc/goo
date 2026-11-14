import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as P from '../ui/cel-provenance.mjs';
import {recordCELContract,AuthoringSlot} from '../ui/cel-authoring.mjs';
import * as Artifact from '../ui/view-artifact.mjs';
const clone=structuredClone;
function fixture(){
 const base={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,where:{field:{kind:'attribute',key:'tenant'},op:'missing'}},predicate={field:{kind:'column',key:'body'},op:'contains',value:'slow'},q={...clone(base),where:{all:[clone(base.where),predicate]}};
 const binding={...recordCELContract,signal:'logs',baseIdentity:'a'.repeat(64),slot:'condition',epoch:'mount',revision:'2'},metadata={version:'goo.cel-authoring/1',representation:'cel',source:'// exact source\r\n log.body.contains("slow") ',base:clone(base),binding,parentRevision:'1',loweringDigest:'b'.repeat(64)};
 return {base,predicate,q,metadata};
}
test('provenance retains exact source and rejects omitted or substituted owner fields',()=>{
 const {q,metadata}=fixture();assert.deepEqual(P.condition(metadata,q),metadata);
 for(const edit of [x=>x.source='',x=>x.binding.library='old',x=>x.binding.revision=2,x=>x.loweringDigest='bad',x=>x.base.toMs++,x=>x.base.where.op='exists',x=>x.rows=[],x=>x.binding.extra=true,x=>x.parentRevision='']){const changed=clone(metadata);edit(changed);assert.throws(()=>P.condition(changed,q));}
 assert.throws(()=>P.authoring({spans:metadata},[q]));assert.throws(()=>P.authoring({},[q]));const detached=P.authoring({logs:metadata},[q]);detached.logs.source='changed';assert.notEqual(metadata.source,'changed');
});
test('portable response normalizes typed base omissions without changing source',()=>{
 const {q,metadata}=fixture(),view={version:1,title:'Authored logs',queries:[q],authoring:{logs:metadata},layout:'records',columns:['utc','summary']},labels={namespace:'team',name:'slow',revision:'v1'};
 const expected=Artifact.artifact(view,labels),wire=clone(expected);delete wire.view.queries[0].offset;delete wire.view.authoring.logs.base.offset;
 const response={artifact:wire,yaml:'version: goo.view-artifact/1\n'};assert.equal(Artifact.response(response,expected).artifact.view.authoring.logs.source,metadata.source);
 const modified=clone(response);modified.artifact.view.authoring.logs.source=metadata.source.replace('\r\n','\n');assert.throws(()=>Artifact.response(modified,expected),/definition changed/);
});
test('checked owner refuses missing lowering identity rather than publishing partial provenance',()=>{
 const {base,q,predicate}=fixture(),owner=new AuthoringSlot({base,baseIdentity:'a'.repeat(64),parentRevision:'1',slot:'condition',epoch:'mount'});owner.edit('log.body.contains("slow")');const check=owner.beginCheck(),receipt={version:'goo.cel-plan/1',celLibrary:recordCELContract.library,expression:check.request.expression,binding:check.request.binding,predicate,question:q};
 assert.equal(owner.accept(check.token,receipt),false);assert.equal(owner.stage(check.token),null);assert.equal(owner.accept(check.token,{...receipt,loweringDigest:'b'.repeat(64)}),true);
});
