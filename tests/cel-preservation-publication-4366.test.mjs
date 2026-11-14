import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import {identity} from '../ui/question-session.mjs';
const source=readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
function fixture(){
 const line=source.split('\n').find(line=>line.includes('getAuthoring:q=>')),start=line.indexOf('getAuthoring:')+13,end=line.indexOf(',canRecall:',start),question={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,limit:100},state={draft:true,unapplied:false,candidate:null},scope={question,QuestionSession:{identity},filterEditor:{hasCELDraft:()=>state.draft,hasUnappliedCELDraft:()=>state.unapplied,celProvenance:()=>state.candidate}};
 vm.runInNewContext('this.get='+line.slice(start,end),scope);return {state,question,get:scope.get};
}
test('a freshly checked condition cannot silently strip source from an older held question',()=>{
 const f=fixture();assert.throws(()=>f.get(f.question),/Run the checked question or discard CEL source/);
 const entry={source:'log.body == "new"'};f.state.candidate=entry;assert.equal(f.get(f.question),entry);
 f.state.unapplied=true;assert.throws(()=>f.get(f.question),/Check or discard/);
});
test('only absent or deliberately discarded ownership returns null; unrelated questions remain independent',()=>{
 const f=fixture();f.state.draft=false;assert.equal(f.get(f.question),null);
 f.state.draft=true;assert.equal(f.get({...f.question,environment:'other'}),undefined);
});
