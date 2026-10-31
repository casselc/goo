/* Pure declaration, compatibility and supplied-cohort helpers. No source IO. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {eligible as aggregateEligible} from './view-declaration.mjs';
import {sameQuestionEnvironment} from './record-environment.mjs';
import {identity} from './question-session.mjs';
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function closed(x,keys){if(!object(x)||Object.keys(x).some(k=>!keys.includes(k)))throw Error('Unknown captured-count comparison fields.');}
export function eligible(q){return aggregateEligible(q)&&!q.numeric&&!q.numericHistogram&&!q.reportedEdge&&!q.spanStatus&&!q.groupTimeSeries;}
export function validate(value,q){
 validateRecordQuestion(q);closed(value,['version','scope','kind','baselineCapture','display']);closed(value.scope,['kind','signal']);closed(value.kind,['name','version']);closed(value.display,['value','title','view','order']);
 if(!eligible(q))throw Error('Captured count comparison requires one ordinary grouping without status, numeric or reported-edge measures.');
 if(value.version!=='goo.view/1'||value.scope.kind!=='record-query'||value.scope.signal!==q.signal||value.kind.name!=='captured-count-comparison'||value.kind.version!==1||value.display.value!=='count'||typeof value.baselineCapture!=='string'||!/^[a-f0-9]{64}$/.test(value.baselineCapture))throw Error('Choose an exact capture and a pinned captured-count comparison.');
 if(value.display.view!==undefined&&!['','summary','groups'].includes(value.display.view)||value.display.order!==undefined&&!['','change','label'].includes(value.display.order))throw Error('Choose summary/groups and change/label order.');
 const title=value.display.title;if(title!==undefined&&(typeof title!=='string'||!title.isWellFormed()||new TextEncoder().encode(title).length>120||/\p{Cc}/u.test(title)||title!==''&&!title.trim()))throw Error('Comparison title must be bounded valid text.');
 return structuredClone(value);
}
export function create(q,id){return validate({version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'captured-count-comparison',version:1},baselineCapture:id,display:{value:'count'}},q)}
export function compatible(a,b){
 for(const q of [a,b]){validateRecordQuestion(q);if(!eligible(q))throw Error('Baseline requires an ordinary single-group count query.');}
 if(!sameQuestionEnvironment(a,b))throw Error('Baseline environment policy differs from the current population.');
 const root=q=>{const copy=structuredClone(q);delete copy.fromMs;delete copy.toMs;delete copy.environment;delete copy.environmentSelector;copy.offset=0;copy.limit=100;return identity(copy)};
 if(root(a)!==root(b))throw Error('Baseline query definition differs from the current population. Windows and delivery pages may differ; filters and measures must match.');
 return true;
}
export function cohort(data){if(!object(data))throw Error('An applied whole-population result is required.');const out={};for(const key of ['question','basis','capturedAt','count','state','groups','duration'])if(data[key]!==undefined)out[key]=structuredClone(data[key]);return out;}
export function request(baseline,selection){compatible(baseline.question,selection.question);return {baseline:cohort(baseline),selection:cohort(selection)}}
