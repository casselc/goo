/* Prepare one immutable held bank; project copies without rescanning other rows. */
import {checkedRequest,checkedResult} from './raw-observation-protocol.mjs';
import * as Gauge from './raw-gauge-schema.mjs';
import * as Histogram from './raw-histogram-schema.mjs';
import {identity} from './raw-observation-identity.mjs';
import {validateSnapshot} from './view-schema.mjs';
function freeze(value){if(value&&typeof value==='object'&&!Object.isFrozen(value)){for(const child of Object.values(value))freeze(child);Object.freeze(value)}return value}
const quality=()=>({extent:'unknown',basis:'captured-query',value:'unknown',sampling:'unknown',freshness:'fixed'});
function checked(snapshot){const admitted=validateSnapshot(snapshot);if(!admitted.ok)throw Error(admitted.reasons.join('; '));return snapshot}
export function prepare(data,expectedRequest={question:data?.question,instrument:data?.instrument}){
 const request=checkedRequest(expectedRequest),bank=freeze(checkedResult(data,request)),rows=new Map(bank.rows.map(row=>[row.occurrenceKey,row]));freeze(request);
 let gauge=null;
 return Object.freeze({
  data:bank,
  gauge(){
   if(bank.instrument.kind!=='gauge')throw Error('Points require raw gauge observations.');
   if(!gauge){
    const id='raw-gauge:'+identity(request),revision='held:'+identity(bank);
    gauge=freeze(checked({version:'goo.dataset/1',id,revision,origin:'source-read',capturedAt:bank.readAt,shape:'observations',scope:{id,generation:revision,signal:'metrics',selector:request},rawGauge:{request,readAt:bank.readAt,limit:bank.limit},fields:Gauge.rawGaugeFields(bank.instrument.unit),rows:bank.rows.map(row=>({key:row.occurrenceKey,state:Gauge.state(row),values:Gauge.values(row),observation:row})),quality:quality()}));
   }
   return structuredClone(gauge);
  },
  histogram(key){
   if(bank.instrument.kind!=='histogram')throw Error('Bucket displays require raw histogram observations.');
   const row=typeof key==='string'?rows.get(key):null;if(!row)throw Error('Choose one occurrence from this held histogram bank.');
   const id='raw-histogram:'+identity({request,key}),revision='held:'+identity({request,readAt:bank.readAt,row});
   return structuredClone(checked({version:'goo.dataset/1',id,revision,origin:'source-read',capturedAt:bank.readAt,shape:'observations',scope:{id,generation:revision,signal:'metrics',selector:request},rawHistogram:{request,readAt:bank.readAt,limit:bank.limit},fields:Histogram.rawHistogramFields(),rows:[{key,state:Histogram.state(row),values:Histogram.values(row),observation:row}],quality:quality()}));
  }
 });
}
