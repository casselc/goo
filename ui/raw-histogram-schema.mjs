/* Closed supplied-data profile for one raw explicit histogram occurrence. */
import {object} from './intent-shape.mjs';
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import {canonical} from './raw-observation-identity.mjs';
const same=(a,b)=>canonical(a)===canonical(b);
export function state(row){return row.recordingState==='no-recorded-value'?'not-recorded':row.recordingState!=='recorded'?'unknown':row.bucketsState==='valid'?'recorded':'invalid'}
export function values(row){return {endNS:row.endNS,occurrence:row.occurrenceKey,count:row.countState==='valid'?row.count:null}}
export function rawHistogramFields(){return [{key:'endNS',label:'Observed UTC',type:'string',unit:'ns'},{key:'occurrence',label:'Held occurrence',type:'string'},{key:'count',label:'Source count',type:'number',nullable:true,unit:'observations'}]}
export function validateRawHistogram(s){
 const reasons=[];
 try{
  object(s.rawHistogram,['request','readAt','limit']);const request=checkedRequest(s.rawHistogram.request);
  if(request.instrument.kind!=='histogram'||s.rawGauge!==undefined)throw Error('Raw histogram datasets require one histogram occurrence.');
  if(!same(s.scope.selector,request)||s.scope.signal!=='metrics')throw Error('Raw histogram selector differs from its instrument request.');
  if(!['source-read','browser-held'].includes(s.origin)||s.capturedAt!==s.rawHistogram.readAt||s.sourceBasis!==undefined||s.completeness!==undefined||s.quality?.extent!=='unknown'||s.quality?.value!=='unknown'||s.quality?.sampling!=='unknown'||s.quality?.freshness!=='fixed')throw Error('Raw observation coverage and value certainty must remain explicit.');
  if(!Array.isArray(s.rows)||s.rows.length!==1||!same(s.fields,rawHistogramFields())||s.dimensions!==undefined||s.interval!==undefined||s.partition!==undefined)throw Error('Raw histogram datasets require one supplied occurrence and its declared fields.');
  checkedResult({version:'goo/raw-histogram-observations/v1',question:request.question,instrument:request.instrument,readAt:s.rawHistogram.readAt,limit:s.rawHistogram.limit,rows:s.rows.map(row=>row.observation)},request);
  const row=s.rows[0];object(row,['key','values','observation','state']);if(row.state!==state(row.observation)||row.key!==row.observation.occurrenceKey||!same(row.values,values(row.observation)))throw Error('Raw histogram row differs from its admitted source occurrence.');
 }catch(error){reasons.push(error.message||'Invalid raw histogram observation dataset.');}
 return reasons;
}
