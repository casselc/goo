// Drive the production request and retained-side response planners. Go validates
// their full outputs independently; this fixture does not reproduce either AST.
import {freshPairExplanationRequest} from '../../ui/population-workspace.mjs';
import {filterRequest,expectedSideQuestion} from '../../ui/explanation.mjs';
import * as CapturedCount from '../../ui/captured-count-comparison.mjs';
let raw='';for await(const chunk of process.stdin)raw+=chunk;
const input=JSON.parse(raw),before=structuredClone(input);
const request=freshPairExplanationRequest(input.selected,input.comparison,input.fields);
input.selected.where.all[0].value='caller-mutated-selected';
input.comparison.baselineQuery.where.all[0].value='caller-mutated-baseline';
input.fields[0].key='caller-mutated-field';
const filters=[];
for(const side of ['baseline','selection'])for(const mode of ['include','exclude'])for(const value of [null,'',before.literal]){
 const field=request.fields[0],wire=filterRequest(request,field,value,side,mode),question=expectedSideQuestion(request,field,value,side,mode);
 filters.push({wire,question});
}
const rejects=[];
for(const mutate of [
 x=>x.comparison=CapturedCount.create(x.selected,'a'.repeat(64)),
 x=>x.comparison.kind.name='arbitrary-explanation',
 x=>{const q=x.comparison.baselineQuery,selector=q.environmentSelector;if(selector?.scope.mode==='named'){selector.policy=selector.policy==='effective'?'current-only':'effective';}else if(!selector){q.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:q.environment}};delete q.environment;}else{selector.scope={mode:'named',value:'prod'};selector.policy='current-only';}},
 x=>x.comparison.baselineQuery.toMs++,
 x=>x.fields.push(structuredClone(x.fields[0])),
 x=>x.fields=[],
]){
 const bad=structuredClone(before);mutate(bad);try{freshPairExplanationRequest(bad.selected,bad.comparison,bad.fields);rejects.push(false)}catch{rejects.push(true)}
}
const output=structuredClone(request);output.baseline.where.all[0].value='mutated-return';
process.stdout.write(JSON.stringify({request,filters,rejects,afterReturnMutation:freshPairExplanationRequest(before.selected,before.comparison,before.fields)})+'\n');
