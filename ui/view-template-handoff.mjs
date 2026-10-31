import * as Composition from './view-composition.mjs';
/* Checked configuration handoff only. No source data, witnesses or navigation. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {identity as questionIdentity} from './question-session.mjs';
import * as ViewDeclaration from './view-declaration.mjs';
const presets={
 'http-server':{id:'otel.http-server-routes',group:'http.route',label:'HTTP route view'},
 'db-client':{id:'otel.db-client-operations',group:'db.operation.name',label:'database client view'}
};
export const profiles=Object.freeze({'ordinary':'Ordinary records','otel.db-client':'OTel database client spans','otel.http-server':'OTel HTTP server spans'});
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
const stable=x=>Array.isArray(x)?x.map(stable):object(x)?Object.fromEntries(Object.keys(x).sort().filter(k=>x[k]!==undefined).map(k=>[k,stable(x[k])])):x;
const same=(a,b)=>JSON.stringify(stable(a))===JSON.stringify(stable(b));
function closed(value,keys){if(!object(value)||Object.keys(value).some(k=>!keys.includes(k)))throw Error('Invalid view-template configuration.');}
export function title(value){if(typeof value!=='string'||!value.trim()||new TextEncoder().encode(value).length>120||/\p{Cc}/u.test(value)||!value.isWellFormed())throw Error('Enter a bounded valid view title.');return value;}
function conjuncts(p){if(!p||p.any?.length)return [];return p.all?.length?p.all.flatMap(conjuncts):[p];}
export function binding(value){closed(value,['version','group','profile']);if(value.version!=='goo.template/1'||value.group!=='destination'||typeof value.profile!=='string'||!Object.hasOwn(profiles,value.profile))throw Error('Choose a supported template profile and destination grouping.');return structuredClone(value);}
export function requirements(profile){if(profile==='otel.db-client')return 'Exact service, Client kind and nonempty current db.system.name.';if(profile==='otel.http-server')return 'Exact service, Server kind and nonempty current http.request.method.';return 'Same signal, one ordinary grouping and requested measure support.';}
export function templateView(value){
 closed(value,['version','id','title','queries','composition','template']);if(value.version!==1||value.id!==undefined&&(typeof value.id!=='string'||!/^[a-f0-9]{64}$/.test(value.id)))throw Error('Choose a checked view/template revision.');
 if(!object(value)||!value.template||!value.composition||value.queries?.length!==1||['layout','columns','fieldColumns','displays','breakdowns','tail'].some(key=>Object.hasOwn(value,key)))throw Error('Reusable templates require one aggregate-and-records composition.');
 binding(value.template);title(value.title);const q=value.queries[0];validateRecordQuestion(q);Composition.validate(value.composition,q);
 if((q.offset??0)!==0||q.numeric||q.numericHistogram||q.reportedEdge||q.groupTimeSeries||q.groupBy?.length!==1||value.composition.children.length!==2||!Composition.aggregate(value.composition)||!Composition.records(value.composition))throw Error('Reusable templates support ordinary count/duration aggregate and records only. Remove timeline, comparison or specialized measures first.');
 if(Composition.aggregate(value.composition).display.query)throw Error('Clear display Find before making a reusable template.');
 const profile=value.template.profile;if(profile!=='ordinary'){
  if(q.signal!=='spans')throw Error('This data profile requires spans.');const leaves=conjuncts(q.where),has=(kind,key,op,value)=>leaves.some(p=>p.field?.kind===kind&&p.field.key===key&&p.op===op&&(p.value??'')===value);
  if(!leaves.some(p=>p.field?.kind==='column'&&p.field.key==='service'&&p.op==='eq'&&typeof p.value==='string'&&p.value!==''))throw Error('The data profile requires an exact nonempty service selector.');
  const db=profile==='otel.db-client';if(!has('column','kind','eq',db?'Client':'Server')||!has('attribute',db?'db.system.name':'http.request.method','ne',''))throw Error('The query must require '+(db?'Client kind and nonempty current db.system.name.':'Server kind and nonempty current http.request.method.'));
 }
 return structuredClone(value);
}
export function templateFor(profile='ordinary'){return binding({version:'goo.template/1',group:'destination',profile});}
export function authoredRequest(source,question){templateView(source);validateRecordQuestion(question);if(source.queries[0].signal!==question.signal)throw Error('Template and destination must use the same signal.');const destination={...structuredClone(source),queries:[{...structuredClone(question),offset:0}]};templateView(destination);return {source:structuredClone(source),query:structuredClone(question)};}
export function eligible(widget){const q=widget?.query,preset=presets[widget?.family];return !!preset&&widget?.state==='supported'&&widget.key==='dimension'&&q?.signal==='spans'&&q.groupBy?.length===1&&q.groupBy[0].kind==='attribute'&&q.groupBy[0].key===preset.group&&!widget.measurement&&!q.numericHistogram&&!q.groupTimeSeries;}
export function customizeLabel(widget){return presets[widget?.family]?.label||'view';}
export function request(widget){if(!eligible(widget))throw Error('Customize requires a supported HTTP route or database operation widget.');validateRecordQuestion(widget.query);return {template:presets[widget.family].id,version:1,widget:structuredClone(widget)};}
export function stagedView(value,question){
 validateRecordQuestion(question);closed(value,['version','title','queries','layout','columns','displays','composition','template']);
 if(value.version!==1||!Array.isArray(value.queries)||value.queries.length!==1)throw Error('Template changed the current widget question.');
 validateRecordQuestion(value.queries[0]);if(questionIdentity(value.queries[0])!==questionIdentity(question))throw Error('Template changed the current widget question.');
 title(value.title);if(Object.hasOwn(value,'composition')){const view=Composition.view(value);if(view.template)templateView(view);return view;}if(value.template)throw Error('Template requires a composition.');if(!['charts-records','records'].includes(value.layout))throw Error('Invalid template layout.');
 if(!Array.isArray(value.columns)||value.columns.length<2||value.columns.length>4||new Set(value.columns).size!==value.columns.length||!value.columns.includes('utc')||!value.columns.includes('summary')||value.columns.some(k=>!['utc','service','summary','context'].includes(k)))throw Error('Invalid template record columns.');
 closed(value.displays,[question.signal]);if(!value.displays[question.signal])throw Error('Template requires a display for its question.');ViewDeclaration.validate(value.displays[question.signal],question);return structuredClone(value);
}
export function response(value,question,expected='otel.http-server-routes'){closed(value,['template','view']);if(!object(value.template)||value.template.id!==expected||value.template.version!==1)throw Error('Unsupported view-template response.');title(value.template.title);return stagedView(value.view,question);}
export function authoredResponse(value,source,question){closed(value,['template','view']);if(!object(value.template)||value.template.version!==1||value.template.title!==source.title||typeof value.template.id!=='string'||!/^[a-f0-9]{64}$/.test(value.template.id)||source.id&&value.template.id!==source.id)throw Error('Template response differs from the selected definition.');const target={...structuredClone(question),offset:0},view=stagedView(value.view,target);if(view.title!==source.title||!same(view.template,source.template)||!same(view.composition,source.composition))throw Error('Template response changed the selected presentation.');templateView(view);return view;}
export async function expand(widget,{post,current=()=>true}={}){const input=request(widget);if(!current())return null;const result=await post(input);if(!current())return null;return response(result,input.widget.query,input.template);}
export async function expandAuthored(source,question,{post,current=()=>true}={}){const input=authoredRequest(source,question);if(!current())return null;const result=await post(input);if(!current())return null;return authoredResponse(result,input.source,input.query);}
