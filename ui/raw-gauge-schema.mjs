/* Closed supplied-data profile for raw gauge occurrences. No IO or evaluation. */
import {object} from './intent-shape.mjs';
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import {binary64Rational} from './measurement-scope.mjs';
const canonical=value=>JSON.stringify(value,(_,entry)=>entry&&typeof entry==='object'&&!Array.isArray(entry)?Object.fromEntries(Object.entries(entry).sort(([a],[b])=>a.localeCompare(b))):entry);
const same=(a,b)=>canonical(a)===canonical(b);
export function recording(row){return row.flagsState==='valid'&&row.flags===0?'recorded':row.flagsState==='valid'&&(row.flags&1)!==0?'no-recorded-value':'unknown'}
export function state(row){const reported=recording(row);return reported==='no-recorded-value'?'not-recorded':reported==='unknown'?'unknown':row.valueState==='finite'?'recorded':'invalid'}
export function values(row){return {endNS:row.endNS,occurrence:row.occurrenceKey,value:row.valueState==='finite'?binary64Rational(row.value):null}}
export function validateRawGauge(snapshot){
 const reasons=[];
 try{
  const s=snapshot;object(s.rawGauge,['request','readAt','limit']);
  const request=checkedRequest(s.rawGauge.request);if(request.instrument.kind!=='gauge')throw Error('Raw gauge datasets require a gauge instrument.');
  if(!same(s.scope.selector,request)||s.scope.signal!=='metrics')throw Error('Raw gauge selector differs from its instrument request.');
  if(!['source-read','browser-held'].includes(s.origin)||s.capturedAt!==s.rawGauge.readAt||s.sourceBasis!==undefined||s.completeness!==undefined||s.quality?.extent!=='unknown'||s.quality?.value!=='unknown'||s.quality?.sampling!=='unknown'||s.quality?.freshness!=='fixed')throw Error('Raw observation coverage and value certainty must remain explicit.');
  if(!Array.isArray(s.rows)||s.rows.length>10000)throw Error('Raw gauge rows exceed the held-bank budget.');
  const unit=request.instrument.unit,expectedFields=[{key:'endNS',label:'Observed UTC',type:'string',unit:'ns'},{key:'occurrence',label:'Held occurrence',type:'string'},{key:'value',label:'Parsed binary64 value',type:'rational',nullable:true,...(unit?{unit}:{})}];
  if(!same(s.fields,expectedFields)||s.dimensions!==undefined||s.interval!==undefined||s.partition!==undefined)throw Error('Raw gauge fields and geometry require their declared observation profile.');
  const bank={version:'goo/raw-gauge-observations/v1',question:request.question,instrument:request.instrument,readAt:s.rawGauge.readAt,limit:s.rawGauge.limit,rows:s.rows.map(row=>row.observation)};
  checkedResult(bank,request);
  for(const row of s.rows){object(row,['key','values','observation','state']);if(row.state!==state(row.observation)||row.key!==row.observation.occurrenceKey||!same(row.values,values(row.observation)))throw Error('Raw gauge row differs from its admitted source occurrence.');}
 }catch(error){reasons.push(error.message||'Invalid raw gauge observation dataset.');}
 return reasons;
}
export function rawGaugeFields(unit){return [{key:'endNS',label:'Observed UTC',type:'string',unit:'ns'},{key:'occurrence',label:'Held occurrence',type:'string'},{key:'value',label:'Parsed binary64 value',type:'rational',nullable:true,...(unit?{unit}:{})}]}
