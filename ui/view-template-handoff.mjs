import {recordAggregates} from './record-aggregate-protocol.mjs';
import * as Composition from './view-composition.mjs';
/* Checked configuration handoff only. No source data, witnesses or navigation. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {identity as questionIdentity} from './question-session.mjs';
import * as ViewDeclaration from './view-declaration.mjs';
import * as Timeline from './timeline-declaration.mjs';
import * as CapturedCount from './captured-count-comparison.mjs';
const presets={
 'http-server':{id:'otel.http-server-routes',group:'http.route',label:'HTTP route view'},
 'db-client':{id:'otel.db-client-operations',group:'db.operation.name',label:'database client view'},
 'rpc-server-current':{id:'otel.rpc-server-current-methods',group:'rpc.method',label:'current RPC method view'}
};
export const profiles=Object.freeze({'ordinary':'Ordinary records','ordinary-summary':'Ungrouped records and summary','otel.db-client':'OTel database client spans','otel.http-server':'OTel HTTP server spans','otel.rpc-server-current':'OTel current RPC server spans'});
const profileRules={'otel.db-client':{kind:'Client',attribute:'db.system.name'},'otel.http-server':{kind:'Server',attribute:'http.request.method'},'otel.rpc-server-current':{kind:'Server',attribute:'rpc.system.name'}};
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
const stable=x=>Array.isArray(x)?x.map(stable):object(x)?Object.fromEntries(Object.keys(x).sort().filter(k=>x[k]!==undefined).map(k=>[k,stable(x[k])])):x;
const same=(a,b)=>JSON.stringify(stable(a))===JSON.stringify(stable(b));
function closed(value,keys){if(!object(value)||Object.keys(value).some(k=>!keys.includes(k)))throw Error('Invalid view-template configuration.');}
export function title(value){if(typeof value!=='string'||!value.trim()||new TextEncoder().encode(value).length>120||/\p{Cc}/u.test(value)||!value.isWellFormed())throw Error('Enter a bounded valid view title.');return value;}
function conjuncts(p){if(!p||p.any?.length)return [];return p.all?.length?p.all.flatMap(conjuncts):[p];}
export function binding(value){closed(value,['version','group','profile','baseline']);if(Object.hasOwn(value,'baseline')&&value.baseline!=='destination')throw Error('Choose the explicit destination baseline binding.');if(value.version!=='goo.template/1'||value.group!==(value.profile==='ordinary-summary'?'none':'destination')||typeof value.profile!=='string'||!Object.hasOwn(profiles,value.profile))throw Error('Choose a supported template profile and destination binding.');if(value.profile==='ordinary-summary'&&Object.hasOwn(value,'baseline'))throw Error('Population summary templates have no baseline binding.');return structuredClone(value);}
export function intervalLabel(width){if(!Number.isInteger(width)||width<=0)return 'Not configured';const unit=width<60000?'second':'minute',value=width<60000?width/1000:width/60000;return value+' '+unit+(value===1?'':'s');}
export function requirements(profile){if(profile==='ordinary-summary')return 'Same signal, ungrouped page-zero records and supplied count, duration or recorded-status facts; no numeric binding.';const rule=Object.hasOwn(profileRules,profile)?profileRules[profile]:null;return rule?'Exact service, '+rule.kind+' kind and nonempty current '+rule.attribute+'.'+(profile==='otel.rpc-server-current'?' rpc.method is conditional; missing and empty labels remain populations.':''):'Same signal, one ordinary grouping and requested measure support.';}
export function templateView(value){
 closed(value,['version','id','title','queries','composition','template']);if(value.version!==1||value.id!==undefined&&(typeof value.id!=='string'||!/^[a-f0-9]{64}$/.test(value.id)))throw Error('Choose a checked view/template revision.');
 if(!object(value)||!value.template||!value.composition||value.queries?.length!==1||['layout','columns','fieldColumns','displays','breakdowns','tail'].some(key=>Object.hasOwn(value,key)))throw Error('Reusable templates require one checked record composition.');
 binding(value.template);title(value.title);const q=value.queries[0];validateRecordQuestion(q);Composition.validate(value.composition,q);
 if(value.template.profile==='ordinary-summary'){if((q.offset??0)!==0||(q.groupBy?.length??0)!==0||q.numeric||q.numericHistogram||q.reportedEdge||q.groupTimeSeries||value.composition.children.length!==2||!Composition.summary(value.composition)||!Composition.records(value.composition))throw Error('Population summary templates require ungrouped page-zero ordinary records, exactly summary and records, and no numeric binding.');return structuredClone(value);}
 if(Composition.summary(value.composition)||ViewDeclaration.logPattern(q)||(q.offset??0)!==0||q.numeric||q.numericHistogram||q.reportedEdge||q.groupTimeSeries||q.groupBy?.length!==1||![2,3,4].includes(value.composition.children.length)||!Composition.aggregate(value.composition)||!Composition.records(value.composition))throw Error('Reusable templates support ordinary count/duration aggregate, records and optional fixed timeline/count comparison. Remove specialized measures first.');
 const comparison=Composition.comparison(value.composition);if(CapturedCount.queryVariant(comparison))throw Error('Fixed query comparisons cannot be category templates.');if(!!comparison!==Object.hasOwn(value.template,'baseline'))throw Error('Comparison templates require an explicit destination baseline binding, only when a comparison is present.');if(comparison&&!CapturedCount.eligible(q))throw Error('Comparison templates require the existing ordinary count profile.');
 if(Composition.aggregate(value.composition).display.query)throw Error('Clear display Find before making a reusable template.');
 const profile=value.template.profile;if(profile!=='ordinary'){
  if(q.signal!=='spans')throw Error('This data profile requires spans.');const leaves=conjuncts(q.where),has=(kind,key,op,value)=>leaves.some(p=>p.field?.kind===kind&&p.field.key===key&&p.op===op&&(p.value??'')===value);
  if(!leaves.some(p=>p.field?.kind==='column'&&p.field.key==='service'&&p.op==='eq'&&typeof p.value==='string'&&p.value!==''))throw Error('The data profile requires an exact nonempty service selector.');
  const rule=profileRules[profile];if(!has('column','kind','eq',rule.kind)||!has('attribute',rule.attribute,'ne',''))throw Error('The query must require '+rule.kind+' kind and nonempty current '+rule.attribute+'.');
 }
 return structuredClone(value);
}
export function templateFor(profile='ordinary',{baseline=false}={}){if(typeof baseline!=='boolean')throw Error('Choose explicitly whether the template binds a destination baseline.');return binding({version:'goo.template/1',group:profile==='ordinary-summary'?'none':'destination',profile,...(baseline?{baseline:'destination'}:{})});}
// Explicit composition edits retain the profile and derive only the baseline slot.
export function bindingForComposition(value,composition){const next=binding(value);delete next.baseline;if(Composition.comparison(composition))next.baseline='destination';return binding(next);}
function boundComposition(source,baselineCapture){const comparison=Composition.comparison(source.composition),next=structuredClone(source.composition);if(comparison){if(typeof baselineCapture!=='string'||!/^[a-f0-9]{64}$/.test(baselineCapture))throw Error('Choose an explicit destination baseline capture before inspecting this template.');Composition.comparison(next).baselineCapture=baselineCapture;}else if(baselineCapture!==undefined)throw Error('A destination baseline is only allowed for comparison templates.');return next;}
export function authoredRequest(source,question,baselineCapture){templateView(source);validateRecordQuestion(question);if(source.queries[0].signal!==question.signal)throw Error('Template and destination must use the same signal.');const timeline=Composition.timeline(source.composition);if(timeline){if(source.queries[0].bucketMs!==question.bucketMs)throw Error('Template timeline requires '+intervalLabel(source.queries[0].bucketMs)+' intervals; configure the destination interval width and Run explicitly.');Timeline.reuse(timeline,source.queries[0],question);}const composition=boundComposition(source,baselineCapture),destination={...structuredClone(source),queries:[structuredClone(question)],composition};templateView(destination);return {source:structuredClone(source),query:structuredClone(question),...(Composition.comparison(source.composition)?{baselineCapture}:{})};}
// Service evidence supplies the entire destination; templates supply presentation only.
export function templateQuestion(question){
 validateRecordQuestion(question);if(ViewDeclaration.logPattern(question)||(question.offset??0)!==0||(question.groupBy?.length??0)>1||question.numeric||question.numericHistogram||question.reportedEdge||question.groupTimeSeries)throw Error('Saved templates require ordinary page-zero records with at most one grouping and no specialized measures.');
 return structuredClone(question);
}
export function templateDestination(widget,data){
 const q=widget?.query;if(widget?.state!=='supported'||widget.measurement||!q||!['observed','no_data'].includes(data?.state))throw Error('Choose a supported ordinary record widget with a completed read.');
 const destination=templateQuestion(q);recordAggregates(q,data);return destination;
}
export function eligible(widget){const q=widget?.query,preset=presets[widget?.family];return !!preset&&widget?.state==='supported'&&widget.key==='dimension'&&q?.signal==='spans'&&q.groupBy?.length===1&&q.groupBy[0].kind==='attribute'&&q.groupBy[0].key===preset.group&&!widget.measurement&&!q.numericHistogram&&!q.groupTimeSeries;}
export function customizeLabel(widget){return presets[widget?.family]?.label||'view';}
export function request(widget){if(!eligible(widget))throw Error('Customize requires a supported HTTP route, current RPC method or database operation widget.');validateRecordQuestion(widget.query);return {template:presets[widget.family].id,version:1,widget:structuredClone(widget)};}
export function stagedView(value,question){
 validateRecordQuestion(question);closed(value,['version','title','queries','layout','columns','displays','composition','template']);
 if(value.version!==1||!Array.isArray(value.queries)||value.queries.length!==1)throw Error('Template changed the current widget question.');
 validateRecordQuestion(value.queries[0]);if(questionIdentity(value.queries[0])!==questionIdentity(question))throw Error('Template changed the current widget question.');
 title(value.title);if(Object.hasOwn(value,'composition')){const view=Composition.view(value);if(view.template)templateView(view);return view;}if(value.template)throw Error('Template requires a composition.');if(!['charts-records','records'].includes(value.layout))throw Error('Invalid template layout.');
 if(!Array.isArray(value.columns)||value.columns.length<2||value.columns.length>4||new Set(value.columns).size!==value.columns.length||!value.columns.includes('utc')||!value.columns.includes('summary')||value.columns.some(k=>!['utc','service','summary','context'].includes(k)))throw Error('Invalid template record columns.');
 closed(value.displays,[question.signal]);if(!value.displays[question.signal])throw Error('Template requires a display for its question.');ViewDeclaration.validate(value.displays[question.signal],question);return structuredClone(value);
}
export function response(value,question,expected='otel.http-server-routes'){closed(value,['template','view']);if(!object(value.template)||value.template.id!==expected||value.template.version!==1)throw Error('Unsupported view-template response.');title(value.template.title);return stagedView(value.view,question);}
export function authoredResponse(value,source,question,baselineCapture){authoredRequest(source,question,baselineCapture);closed(value,['template','view']);if(!object(value.template)||value.template.version!==1||value.template.title!==source.title||typeof value.template.id!=='string'||!/^[a-f0-9]{64}$/.test(value.template.id)||source.id&&value.template.id!==source.id)throw Error('Template response differs from the selected definition.');const target=structuredClone(question),view=stagedView(value.view,target);if(view.title!==source.title||!same(view.template,source.template)||!same(view.composition,boundComposition(source,baselineCapture)))throw Error('Template response changed the selected presentation.');templateView(view);return view;}
export async function expand(widget,{post,current=()=>true}={}){const input=request(widget);if(!current())return null;const result=await post(input);if(!current())return null;return response(result,input.widget.query,input.template);}
export async function expandAuthored(source,question,{baselineCapture,post,current=()=>true}={}){const input=authoredRequest(source,question,baselineCapture);if(!current())return null;const result=await post(input);if(!current())return null;return authoredResponse(result,input.source,input.query,input.baselineCapture);}
