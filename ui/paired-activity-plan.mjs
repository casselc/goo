import {object} from './intent-shape.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import * as RecordEnvironment from './record-environment.mjs';

/* One shared selector, two independently owned tails. No clocks, reads or state. */
export function sharedWhere(value){
 let where;for(const signal of ['logs','spans'])try{const accepted=validatePredicate(value,signal);if(signal==='logs')where=accepted}catch(error){throw Error('Shared filter is unavailable for '+signal+': '+error.message)}
 return where;
}
export function create(selector,conditions={}){
 object(conditions,['logs','spans'],'Choose log and span lane conditions.');
 object(selector,['signal','environment','environmentSelector','where'],'Choose an ordinary shared Activity selector.');
 if(selector.signal!==undefined&&!['logs','spans'].includes(selector.signal))throw Error('Choose an ordinary log or span selector before pairing.');
 RecordEnvironment.requireNarrow(selector);
 const environment=RecordEnvironment.questionFields(selector),where=sharedWhere(selector.where),result={};
 for(const signal of ['logs','spans']){
  const local=validatePredicate(conditions[signal],signal),combined=where===undefined?local:local===undefined?where:{all:[where,local]},admitted=validatePredicate(combined,signal);
  result[signal]={signal,...structuredClone(environment),...(admitted===undefined?{}:{where:admitted})};
 }
 return result;
}
export function laneSelector(selector,signal,conditions={}){
 if(!['logs','spans'].includes(signal))throw Error('Choose logs or spans.');
 return create(selector,conditions)[signal];
}
export function reason(selector,conditions){try{create(selector,conditions);return ''}catch(error){return error.message}}
