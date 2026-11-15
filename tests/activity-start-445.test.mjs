import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
import * as PairedPlan from '../ui/paired-activity-plan.mjs';
import * as RecordEnvironment from '../ui/record-environment.mjs';
const source=readFileSync(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8'),begin=source.indexOf('  function startPair(){'),end=source.indexOf('  function start(){',begin),start=source.slice(begin,end);

test('paired parent commits only after local start acceptance and always retires its action lease',()=>{
 assert(begin>=0&&end>begin);
 for(const answer of [false,true,'throw']){
  // Normalize the VM's data-only selector into the validator's realm.
  const previous={signal:'both',environment:'held'},where={field:{kind:'resource',key:'service.namespace'},op:'eq',value:'workers'},calls=[],state={filterError:null,draft:{},apply:{hidden:true},applied:{},appliedSelector:previous,pairStarting:false,scopeStaged:true,returnAnchor:{held:true},selectedOpener:{},PairedPlan:{create:value=>PairedPlan.create(JSON.parse(JSON.stringify(value)))},RecordEnvironment,environmentDraft:{snapshot:()=>({environment:'candidate'})},FilterEditor:{normalize:structuredClone},filters:{where:()=>where,apply(){calls.push('commit')}},gate:{explicit(){calls.push('explicit')}},renderFilters(){calls.push('render')},present(){calls.push('present')}};
  state.paired={sync(){calls.push(state.pairStarting?'lease':'retired')},start(selector){assert.equal(state.pairStarting,true);assert.equal(state.appliedSelector,previous);assert(!calls.includes('commit'));assert.equal(selector.where.value,'workers');if(answer==='throw')throw Error('Withdrawn');return answer}};
  vm.runInNewContext(start+'\nresult=startPair();',state);
  assert.equal(state.pairStarting,false);assert.equal(calls.at(-1),'retired');
  if(answer===true){assert.equal(state.result,true);assert.notEqual(state.appliedSelector,previous);assert.equal(state.appliedSelector.environment,'candidate');assert(calls.indexOf('commit')>calls.indexOf('lease'));assert.equal(state.apply.hidden,true)}
  else{assert.equal(state.result,false);assert.equal(state.appliedSelector,previous);assert(!calls.includes('commit'));assert(!calls.includes('explicit'));assert.equal(state.scopeStaged,true)}
  assert.deepEqual(where,{field:{kind:'resource',key:'service.namespace'},op:'eq',value:'workers'});
 }
});
