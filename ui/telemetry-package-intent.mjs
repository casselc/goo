/* Definition/question admission only. Go remains the compiler and source executor. */
import * as Environment from './record-environment.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {identity} from './question-session.mjs';
const own=(value,key)=>Object.hasOwn(value,key),object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
function sourceFields(source){
 if(!object(source)||Object.keys(source).some(key=>!['adapter','signal','environment','environmentSelector'].includes(key))||source.adapter!=='clickhouse.otel/v1'||!['logs','spans'].includes(source.signal)||own(source,'environment')===own(source,'environmentSelector')||source.environment===null||source.environmentSelector===null)throw Error('Choose exactly one package environment name or typed selector.');
 return Environment.questionFields(source);
}
export function sourceLabel(source){sourceFields(source);return Environment.questionLabel(source);}
export function packageQuestion(definition,{fromMs,toMs}){
 const q={signal:definition.source?.signal,...sourceFields(definition.source),fromMs,toMs,limit:100,offset:0};
 for(const key of ['numericHistogram','reportedEdge','numeric','groupTimeSeries','spanStatus','logSeverity','bucketMs','summary','where','groupBy'])if(definition[key]!==undefined&&definition[key]!==null)q[key]=structuredClone(definition[key]);
 validateRecordQuestion(q);return q;
}
function canonical(value){return JSON.stringify(value,(_key,item)=>object(item)?Object.fromEntries(Object.keys(item).sort().map(key=>[key,item[key]])):item);}
export function packagePlan(value,{question,revision,definition:knownDefinition,relative=false}={}){
 if(!object(value)||!object(value.definition)||!object(value.question)||typeof value.revision!=='string'||!/^[a-f0-9]{64}$/.test(value.revision))throw Error('A checked package revision and fixed question are required.');
 const plan=structuredClone(value),definition=plan.definition;if(definition.apiVersion!=='goo/telemetry-package/v1alpha1'||definition.measure!=='captured-record-count'||!Number.isInteger(definition.lookbackMinutes)||definition.lookbackMinutes<1||definition.lookbackMinutes>120)throw Error('Unsupported reusable record definition.');
 if(knownDefinition!==undefined&&canonical(knownDefinition)!==canonical(definition))throw Error('Returned definition differs from the selected immutable revision.');
 if(!Array.isArray(definition.views)||definition.views.length<1||definition.views.length>3||new Set(definition.views).size!==definition.views.length||definition.views.some(view=>!['records','distribution','trace-context'].includes(view))||definition.views.includes('distribution')&&definition.groupBy?.length!==1)throw Error('Unsupported package views.');
 if(plan.handoff!==undefined&&(!object(plan.handoff)||plan.handoff.fields!==undefined&&(!Array.isArray(plan.handoff.fields)||plan.handoff.fields.some(field=>!object(field)||Object.keys(field).some(key=>!['kind','key'].includes(key))||!['column','attribute','resource'].includes(field.kind)||typeof field.key!=='string'||!field.key||!field.key.isWellFormed()))||plan.handoff.notes!==undefined&&(!Array.isArray(plan.handoff.notes)||plan.handoff.notes.some(note=>typeof note!=='string'))))throw Error('Unsupported package handoff.');
 if(relative&&plan.question.toMs-plan.question.fromMs!==definition.lookbackMinutes*60000)throw Error('Relative package plan changed its declared lookback.');
 const expected=packageQuestion(definition,plan.question);validateRecordQuestion(plan.question);if(identity(expected)!==identity(plan.question))throw Error('Package question differs from its definition scope or profile.');
 if(revision!==undefined&&plan.revision!==revision)throw Error('Returned package revision differs from the selected definition.');
 if(question!==undefined){validateRecordQuestion(question);const original={...structuredClone(question),limit:100,offset:0};if(identity(original)!==identity(plan.question))throw Error('Package plan changed the original fixed question.');}
 return plan;
}
