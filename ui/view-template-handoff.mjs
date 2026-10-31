/* Checked configuration handoff only. No source data, witnesses or navigation. */
import {validateRecordQuestion} from './question-protocol.mjs';
import * as ViewDeclaration from './view-declaration.mjs';
const id='otel.http-server-routes',version=1;
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
const stable=x=>Array.isArray(x)?x.map(stable):object(x)?Object.fromEntries(Object.keys(x).sort().filter(k=>x[k]!==undefined).map(k=>[k,stable(x[k])])):x;
function closed(value,keys){if(!object(value)||Object.keys(value).some(k=>!keys.includes(k)))throw Error('Invalid view-template configuration.');}
export function title(value){if(typeof value!=='string'||!value.trim()||new TextEncoder().encode(value).length>120||/\p{Cc}/u.test(value)||!value.isWellFormed())throw Error('Enter a bounded valid view title.');return value;}
export function eligible(widget){const q=widget?.query;return widget?.state==='supported'&&widget.family==='http-server'&&widget.key==='dimension'&&q?.signal==='spans'&&q.groupBy?.length===1&&q.groupBy[0].kind==='attribute'&&q.groupBy[0].key==='http.route'&&!widget.measurement&&!q.numericHistogram&&!q.groupTimeSeries;}
export function request(widget){if(!eligible(widget))throw Error('Customize requires a supported HTTP server route widget.');validateRecordQuestion(widget.query);return {template:id,version,widget:structuredClone(widget)};}
export function stagedView(value,question){
 validateRecordQuestion(question);closed(value,['version','title','queries','layout','columns','displays']);
 if(value.version!==1||!['charts-records','records'].includes(value.layout)||!Array.isArray(value.queries)||value.queries.length!==1||JSON.stringify(stable(value.queries[0]))!==JSON.stringify(stable(question)))throw Error('Template changed the current widget question.');
 title(value.title);validateRecordQuestion(value.queries[0]);
 if(!Array.isArray(value.columns)||value.columns.length<2||value.columns.length>4||new Set(value.columns).size!==value.columns.length||!value.columns.includes('utc')||!value.columns.includes('summary')||value.columns.some(k=>!['utc','service','summary','context'].includes(k)))throw Error('Invalid template record columns.');
 closed(value.displays,[question.signal]);if(!value.displays[question.signal])throw Error('Template requires a display for its question.');ViewDeclaration.validate(value.displays[question.signal],question);
 return structuredClone(value);
}
export function response(value,question){closed(value,['template','view']);if(!object(value.template)||value.template.id!==id||value.template.version!==version)throw Error('Unsupported view-template response.');title(value.template.title);return stagedView(value.view,question);}
export async function expand(widget,{post,current=()=>true}={}){const input=request(widget);if(!current())return null;const result=await post(input);if(!current())return null;return response(result,input.widget.query);}
