import * as Fields from './field-presentation.mjs';
import * as Environment from './record-environment.mjs';
/* Supplied predicate display only: no draft ownership, validation or IO. */
export function present(element,where,{prefix='Applied',empty='All records in scope'}={}){
 const text=where?Fields.selectorText(where):empty;
 element.textContent=(prefix?prefix+' · ':'')+text;
 element.title=where?'Exact predicate · '+Fields.referenceText(where):empty;
}

export function scope(element,selector){
 const signal=selector.signal==='both'?'Logs + spans':selector.signal==='spans'?'Spans':'Logs';
 element.textContent='Applied · '+Environment.questionLabel(selector)+' · '+signal;
 element.title=JSON.stringify(selector);
}
