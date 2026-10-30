/* Independent captured histogram operands. No merged scope, custody or source IO. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {validateSnapshot} from './view-schema.mjs';
import {sameQuestionEnvironment} from './record-environment.mjs';
const record=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
const binding=b=>JSON.stringify([b.field.kind,b.field.key,b.type,b.scale??0,b.unit,b.parse]);
export function validateHistogramComparison(value){
 const reasons=[];
 if(!record(value)||value.version!=='goo.histogram-comparison/1'||!['observed','no_data','diagnostic'].includes(value.state))return {ok:false,reasons:['A supported histogram comparison and state are required']};
 if(Object.keys(value).some(key=>!['version','state','reason','baseline','selection'].includes(key)))reasons.push('Unknown histogram comparison option');
 if(value.state==='diagnostic'?value.reason!=='source-reads-disagree':value.reason!==undefined)reasons.push('Diagnostic comparisons require the source-reads-disagree reason');
 for(const side of ['baseline','selection']){const checked=validateSnapshot(value[side]);if(!checked.ok)reasons.push(...checked.reasons.map(reason=>side+': '+reason));if(value[side]?.shape!=='histogram')reasons.push(side+': a full fixed numeric histogram is required')}
 if(reasons.length)return {ok:false,reasons};
 const a=value.baseline,b=value.selection,aq=a.scope.selector,bq=b.scope.selector;
 for(const side of [a,b]){try{const q=side.scope.selector;validateRecordQuestion(q);if(q.signal!==side.scope.signal||q.numericHistogram?.binning!=='powers-of-two'||!q.numeric||binding(q.numeric)!==binding(side.histogram.binding))reasons.push('Histogram operand scope must match its numeric binding and requested bins')}catch{reasons.push('Histogram operand requires a checked numeric histogram question')}}
 if(binding(a.histogram.binding)!==binding(b.histogram.binding))reasons.push('Histogram comparisons require identical numeric bindings');
 if(a.scope.signal!==b.scope.signal||!sameQuestionEnvironment(aq,bq))reasons.push('Histogram comparison signals and environments must match');
 if(a.rows.length!==b.rows.length||a.rows.some((row,i)=>row.key!==b.rows[i].key||row.values.from!==b.rows[i].values.from||row.values.to!==b.rows[i].values.to))reasons.push('Histogram comparison bins must align exactly');
 const empty=a.histogram.support.total==='0'||b.histogram.support.total==='0';
 if(value.state==='observed'&&empty||value.state==='no_data'&&!empty)reasons.push('Histogram comparison state must distinguish empty captured populations');
 return {ok:reasons.length===0,reasons:[...new Set(reasons)]};
}
