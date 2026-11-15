import {object} from './intent-shape.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import * as RecordEnvironment from './record-environment.mjs';

/* One shared selector, two independently owned tails. No clocks, reads or state. */
export function sharedWhere(value){
 let where;for(const signal of ['logs','spans'])try{const accepted=validatePredicate(value,signal);if(signal==='logs')where=accepted}catch(error){throw Error('Shared filter is unavailable for '+signal+': '+error.message)}
 return where;
}
export function create(selector){
 object(selector,['signal','environment','environmentSelector','where'],'Choose an ordinary shared Activity selector.');
 if(selector.signal!==undefined&&!['logs','spans'].includes(selector.signal))throw Error('Choose an ordinary log or span selector before pairing.');
 RecordEnvironment.requireNarrow(selector);
 const environment=RecordEnvironment.questionFields(selector),where=sharedWhere(selector.where),result={};
 for(const signal of ['logs','spans']){
  result[signal]={signal,...structuredClone(environment),...(where===undefined?{}:{where:structuredClone(where)})};
 }
 return result;
}
export function reason(selector){try{create(selector);return ''}catch(error){return error.message}}
