import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';

const source=readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
function statement(start){const index=source.indexOf(start);assert(index>=0);return source.slice(index,source.indexOf('\n',index))}
function fixture(){
 const controls=new Map(),calls={pages:0,tails:0,blocked:null},q={signal:'logs',environment:'prod',fromMs:1000,toMs:61000,offset:0,limit:20};
 const el=id=>{if(!controls.has(id))controls.set(id,{disabled:false});return controls.get(id)};
 const state={el,tailButton:el('exploreTail'),filterError:null,numericError:null,phase:'applied',question:q,lastResult:{question:q,hasMore:true},rootControlsCurrent:()=>true,showStatus(){},run(){calls.pages++}};
 state.session={snapshot:()=>({phase:state.phase})};state.numericControls={error:()=>state.numericError};state.tail={start(){calls.tails++},setReadBlocked(reason){calls.blocked=reason}};
 vm.runInNewContext(statement(' function syncFilterValidity()')+'\n'+statement('tailButton.onclick=')+'\n'+statement("  el('exploreNext').onclick=")+'\nthis.sync=syncFilterValidity;',state);
 return {state,calls,el};
}

test('held fixed rows cannot grant Next or Tail authority to invalid CEL or numeric drafts',()=>{
 for(const kind of ['filterError','numericError']){
  const f=fixture(),oldNext=f.el('exploreNext').onclick,oldTail=f.state.tailButton.onclick;
  f.state[kind]='Check the expression before applying.';f.state.sync();
  assert.equal(f.el('exploreNext').disabled,true);assert.equal(f.state.tailButton.disabled,true);assert.equal(f.calls.blocked,f.state[kind]);
  oldNext();oldTail();assert.equal(f.calls.pages,0);assert.equal(f.calls.tails,0);
  f.state[kind]=null;f.state.phase='draft';f.state.sync();oldNext();oldTail();
  assert.equal(f.el('exploreNext').disabled,true);assert.equal(f.state.tailButton.disabled,true);assert.equal(f.calls.pages,0);assert.equal(f.calls.tails,0);
  f.state.phase='applied';f.state.sync();assert.equal(f.el('exploreNext').disabled,false);assert.equal(f.state.tailButton.disabled,false);assert.equal(f.calls.blocked,'');
  oldNext();oldTail();assert.equal(f.calls.pages,1);assert.equal(f.calls.tails,1);
 }
});

test('next-page native and direct authority also require a reported continuation within the page budget',()=>{
 for(const edit of [f=>f.state.lastResult.hasMore=false,f=>f.state.question.offset=1000]){
  const f=fixture();edit(f);f.state.sync();assert.equal(f.el('exploreNext').disabled,true);f.el('exploreNext').onclick();assert.equal(f.calls.pages,0);
 }
});
