/* Raw gauge supplied-data projection and disposable plot geometry. */
import {recording} from './raw-gauge-schema.mjs';
import {validateDisplay} from './view-schema.mjs';
import {prepare} from './raw-observation-dataset.mjs';
export function dataset(data){return prepare(data).gauge()}
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
