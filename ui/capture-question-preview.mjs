/* Supplied questions only. Preview neither reads telemetry nor grants capture authority. */
import {node} from './components.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {selectorText} from './question-session.mjs';
import * as Environment from './record-environment.mjs';
import {element} from './evidence-scope.mjs';

export function comparisonScopes(question){
 if(question?.kind!=='comparison'||!question.comparison)throw Error('Choose a comparison question.');
 return [['Baseline',question.comparison.baseline],['Selected',question.comparison.selection]].map(([label,q])=>{
  validateRecordQuestion(q);const policy=Environment.selection(q);
  return {label,selector:q.where?selectorText(q.where):'All matching records',data:{mode:'fixed',state:'unknown',facts:[
   {label:'Population',value:q.signal+' · '+Environment.questionLabel(q)},
   {label:'Window · UTC',value:'['+BigInt(q.fromMs)*1000000n+', '+BigInt(q.toMs)*1000000n+')',format:'unix-ns-range'}
  ],details:[{label:'Environment policy',value:policy.policy},{label:'Grouping',value:(q.groupBy||[]).map(field=>field.kind+':'+field.key).join(', ')||'None'},{label:'Question',value:'Fixed bounds · independent new read on confirmation'}],notice:''}};
 });
}
export function renderComparison(host,question){
 host.replaceChildren();host.hidden=question?.kind!=='comparison';if(host.hidden)return;
 for(const scope of comparisonScopes(question)){const section=node('section');const selector=node('details'),text=node('p',scope.selector);text.className='capture-selector';selector.append(node('summary','Selector'),text);section.append(node('h3',scope.label),element(scope.data),selector);host.append(section)}
}
