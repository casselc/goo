/* Raw gauge supplied-data projection and disposable plot geometry. */
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import {recording,state,values,rawGaugeFields} from './raw-gauge-schema.mjs';
import {validateSnapshot,validateDisplay} from './view-schema.mjs';
const canonical=value=>JSON.stringify(value,(_,entry)=>entry&&typeof entry==='object'&&!Array.isArray(entry)?Object.fromEntries(Object.entries(entry).sort(([a],[b])=>a.localeCompare(b))):entry);
// Local snapshot identity only: this is not a cryptographic source receipt.
function identity(value){let a=2166136261,b=2246822519;for(const byte of new TextEncoder().encode(canonical(value))){a=Math.imul(a^byte,16777619);b=Math.imul(b^byte,3266489917)}return (a>>>0).toString(16).padStart(8,'0')+(b>>>0).toString(16).padStart(8,'0')}
export function dataset(data){
 const request=checkedRequest({question:data?.question,instrument:data?.instrument}),bank=checkedResult(data,request);
 if(bank.instrument.kind!=='gauge')throw Error('Points require raw gauge observations.');
 const id='raw-gauge:'+identity(request),revision='held:'+identity(bank);
 const snapshot={version:'goo.dataset/1',id,revision,origin:'source-read',capturedAt:bank.readAt,shape:'observations',scope:{id,generation:revision,signal:'metrics',selector:request},rawGauge:{request,readAt:bank.readAt,limit:bank.limit},fields:rawGaugeFields(bank.instrument.unit),rows:bank.rows.map(row=>({key:row.occurrenceKey,state:state(row),values:values(row),observation:row})),quality:{extent:'unknown',basis:'captured-query',value:'unknown',sampling:'unknown',freshness:'fixed'}};
 const admitted=validateSnapshot(snapshot);if(!admitted.ok)throw Error(admitted.reasons.join('; '));return structuredClone(snapshot);
}
function reason(row){const state=recording(row);if(state==='no-recorded-value')return 'not-recorded';if(state==='unknown')return 'recording-unknown';if(row.valueState!=='finite'||typeof row.value!=='number'||!Number.isFinite(row.value))return 'invalid-value';return null}
export function model(snapshot){
 const admitted=validateDisplay(snapshot,{kind:'points'});
 if(!admitted.ok)return {state:'unavailable',reasons:admitted.reasons,points:[],excluded:[],counts:{loaded:0,eligible:0,plotted:0,unplotted:0,notShownEligible:0,undisplayed:0,unknownRecording:0,noRecordedValue:0,invalidValue:0}};
 const request=snapshot.rawGauge.request,fromNS=String(BigInt(request.question.fromMs)*1000000n),toNS=String(BigInt(request.question.toMs)*1000000n),eligible=[],excluded=[];
 const counts={loaded:snapshot.rows.length,eligible:0,plotted:0,unplotted:0,notShownEligible:0,undisplayed:0,unknownRecording:0,noRecordedValue:0,invalidValue:0};
 for(const entry of snapshot.rows){const row=entry.observation,why=reason(row),item={key:entry.key,endNS:row.endNS,rawValue:structuredClone(row.rawValue??null)};if(why){excluded.push({...item,reason:why});counts[why==='not-recorded'?'noRecordedValue':why==='recording-unknown'?'unknownRecording':'invalidValue']++}else eligible.push({...item,value:row.value})}
 counts.eligible=eligible.length;
 let min=null,max=null;for(const row of eligible){if(min===null||row.value<min)min=row.value;if(max===null||row.value>max)max=row.value}
 const range=max===null?0:max-min,normalize=value=>min===max?.5:Number.isFinite(range)?(value-min)/range:(value/2-min/2)/(max/2-min/2),groups=new Map();
 for(const row of eligible){const x=Number(BigInt(row.endNS)-BigInt(fromNS))/Number(BigInt(toNS)-BigInt(fromNS)),y=normalize(row.value),key=x+':'+y;let group=groups.get(key);if(!group){group={key:'point:'+row.key,occurrenceKeys:[],rawValues:[],members:[],endNS:row.endNS,value:row.value,x,y,multiplicity:0};groups.set(key,group)}group.occurrenceKeys.push(row.key);group.rawValues.push(row.rawValue);group.members.push({key:row.key,endNS:row.endNS,rawValue:row.rawValue});group.multiplicity++}
 const all=[...groups.values()],points=all.slice(0,1000);
 for(const group of points)counts.plotted+=group.multiplicity;
 for(const group of all.slice(1000))for(const member of group.members)excluded.push({...member,reason:'visual-budget'});
 counts.notShownEligible=counts.undisplayed=counts.eligible-counts.plotted;counts.unplotted=counts.loaded-counts.plotted;counts.renderedLocations=points.length;
 return {state:snapshot.rows.length?'observed':'no_data',points,excluded,counts,coordinateGroups:all.length,fromNS,toNS,unit:request.instrument.unit,readAt:snapshot.rawGauge.readAt,atLimit:snapshot.rows.length===snapshot.rawGauge.limit,domain:{min,max}};
}
