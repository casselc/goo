/* Closed tail transport outcomes; no source reads or browser planning. */
import {SourceReadError,refusal} from './source-refusal.mjs';
const kinds=Object.freeze({source_missing:'unavailable',source_access:'unavailable',source_schema:'unavailable',source_unavailable:'unavailable',source_timeout:'timeout',source_malformed:'malformed',source_budget:'budget',identity_budget:'budget',source_cancelled:'cancelled'});
const details=Object.freeze({source_missing:'The requested record source is missing.',source_access:'The source refused access to the requested records.',source_schema:'The record source does not match the supported exporter schema.',source_unavailable:'The source request could not be completed.',source_timeout:'The source read exceeded its time limit.',source_cancelled:'The source read was cancelled before completion.',source_budget:'The source response exceeded a bounded read or grouping limit. Try a narrower scope.',identity_budget:'The source response exceeded a bounded read or grouping limit. Try a narrower scope.',source_malformed:'The source returned invalid or unsupported record data. No partial result was admitted.'});
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function malformed(){return {reason:'source_malformed',retry:'explicit',refusal:{kind:'malformed',detail:'The tail response did not match the supported protocol.'}}}
function gap(value){
 if(!object(value)||Object.keys(value).some(k=>!['reason','retry','refusal'].includes(k)))return malformed();
 if(value.reason==='connection_ended')return value.retry==='automatic'&&value.refusal===undefined?{reason:value.reason,retry:value.retry}:malformed();
 const kind=kinds[value.reason];if(!kind)return malformed();
 let checked;try{checked=refusal(value.refusal)}catch{return malformed()}
 const expected=value.reason==='source_timeout'?'automatic':'explicit';
 if(checked.kind!==kind||checked.detail!==details[value.reason]||!['automatic','explicit'].includes(value.retry)||value.reason!=='source_unavailable'&&value.retry!==expected)return malformed();
 return {reason:value.reason,retry:value.retry,refusal:checked};
}
function sourceError(value){return new SourceReadError(value.refusal||malformed().refusal)}
export {gap,malformed,sourceError};
function http(status,value){
 const expected={400:['Invalid tail selector.','invalid_selector','explicit'],429:['Live stream capacity reached.','capacity','automatic'],500:['Live streaming unavailable.','server_unavailable','automatic']}[status];
 if(!expected||!object(value)||Object.keys(value).some(k=>!['error','reason','retry'].includes(k))||value.error!==expected[0]||value.reason!==expected[1]||value.retry!==expected[2])return malformed();
 return {reason:value.reason,retry:value.retry,...(value.retry==='explicit'?{refusal:{kind:'malformed',detail:'The applied tail selector was rejected.'}}:{})};
}
export {http};

const titles=Object.freeze({source_missing:'Source table missing.',source_access:'Access denied.',source_schema:'Source schema incompatible.',source_unavailable:'Source unavailable.',source_timeout:'Source read timed out.',source_cancelled:'Source read cancelled.',source_budget:'Source read exceeded a limit.',identity_budget:'Observation identity limit reached.',source_malformed:'Source response rejected.',invalid_selector:'Stream selector rejected.'});
function title(value){return titles[value?.reason]||titles.source_malformed}
export {title};
