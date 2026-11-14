/* Ordinary record configuration admission. Execution stays authoritative in Go. */
import {object,text,array} from './intent-shape.mjs';
import {validate as validatePredicate,validateField} from './predicate-validation.mjs';
import {validateNumericPredicate} from './numeric-predicate.mjs';
import * as Environment from './record-environment.mjs';
const allowed=['environmentSelector','signal','environment','fromMs','toMs','limit','offset','where','groupBy','numeric','numericHistogram','summary','spanStatus','logSeverity','bucketMs','groupTimeSeries','reportedEdge'];
const present=(q,k)=>q[k]!==undefined;
function namespace(n){object(n,['mode','value'],'Invalid reported namespace.');if(!['missing','value'].includes(n.mode)||present(n,'value')&&!text(n.value,128)||n.mode==='missing'&&n.value!==undefined&&n.value!=='')throw Error('Invalid reported namespace.');return [n.mode,n.value??''];}
function edge(e,signal){
 object(e,['parent','child','discoveryNamespace'],'Invalid reported edge.');
 if(signal!=='spans')throw Error('Reported edges require spans.');
 const sides=[e.parent,e.child].map(s=>{object(s,['service','namespace'],'Invalid edge service.');if(!text(s.service,128))throw Error('Invalid edge service.');return [s.service,...namespace(s.namespace)];});
 if(JSON.stringify(sides[0])===JSON.stringify(sides[1]))throw Error('Reported edges require distinct service identities.');
 if(present(e,'discoveryNamespace')){const n=JSON.stringify(namespace(e.discoveryNamespace));if(sides.some(s=>JSON.stringify(s.slice(1))!==n))throw Error('Edge identities differ from discovery namespace.');}
}
export function validateRecordQuestion(q){
 object(q,allowed,'Unsupported record question fields.');
 if(!['logs','spans'].includes(q.signal)||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000||q.toMs>4102444800000)throw Error('Require a fixed window of at most 120 minutes before 2100.');
 if(present(q,'environment')&&typeof q.environment!=='string'||q.environmentSelector===null)throw Error('Invalid record environment.');
 Environment.selection(q);
 if(!Number.isInteger(q.limit)||q.limit<1||q.limit>100||present(q,'offset')&&(!Number.isInteger(q.offset)||q.offset<0||q.offset>1000))throw Error('Page limit 1..100 and offset 0..1000 required.');
 if(present(q,'where'))validatePredicate(q.where,q.signal);
 if(present(q,'groupBy')){array(q.groupBy);if(q.groupBy.length>2)throw Error('Use at most two grouping fields.');const seen=new Set();for(const f of q.groupBy){validateField(f,q.signal);const key=JSON.stringify([f.kind,f.key]);if(seen.has(key))throw Error('Duplicate grouping field.');seen.add(key);}}
 if(present(q,'numeric'))validateNumericPredicate({numeric:q.numeric,op:'valid'});
 if(present(q,'numericHistogram')){object(q.numericHistogram,['binning'],'Invalid numeric histogram.');if(!q.numeric||q.numericHistogram.binning!=='powers-of-two')throw Error('Numeric distribution requires a numeric binding and powers-of-two bins.');}
 if(present(q,'summary')&&(!['','span-duration'].includes(q.summary)||q.summary&&q.signal!=='spans')||present(q,'spanStatus')&&(typeof q.spanStatus!=='boolean'||q.spanStatus&&q.signal!=='spans')||present(q,'groupTimeSeries')&&typeof q.groupTimeSeries!=='boolean')throw Error('Invalid record presentation.');
 if(present(q,'logSeverity')&&(typeof q.logSeverity!=='boolean'||q.logSeverity&&(q.signal!=='logs'||q.numeric||q.numericHistogram||q.groupTimeSeries)))throw Error('Numeric severity requires ordinary log count projections.');
 if(present(q,'bucketMs')&&![0,1000,10000,60000,300000].includes(q.bucketMs))throw Error('Choose a supported time bucket.');
 if(q.bucketMs&&Math.ceil((q.toMs-q.fromMs)/q.bucketMs)>240)throw Error('Use at most 240 time intervals; increase bucket width or narrow window.');
 if(q.groupTimeSeries&&(!q.bucketMs||!q.groupBy?.length))throw Error('Split chart requires a time bucket and breakdown.');
 if(present(q,'reportedEdge'))edge(q.reportedEdge,q.signal);
 return q;
}
