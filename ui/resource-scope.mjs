import {RequestOwner} from './request-owner.mjs';
/* Shared bounded resource scope and request ownership; no entity identity inference. */
import * as FilterEditor from './predicate-validation.mjs';
const global=globalThis;
function scope(environment,minutes,at=Date.now()){if(typeof environment!=='string'||!environment||new TextEncoder().encode(environment).length>128||/[\x00-\x1f\x7f]/.test(environment)||!Number.isFinite(minutes)||minutes<=0||minutes>120||!Number.isSafeInteger(at))throw Error('Choose an environment and lookback of at most 120 minutes.');return {environment,fromMs:at-Math.round(minutes*60000),toMs:at}}
function matchingRecords(q,resource,signal,{preserveMissing=[]}={}){const predicates=Object.entries(resource).sort(([a],[b])=>a<b?-1:a>b?1:0).map(([key,value])=>({field:{kind:'resource',key},op:'eq',value}));for(const key of [...new Set(preserveMissing)].sort())if(!Object.hasOwn(resource,key))predicates.push({field:{kind:'resource',key},op:'missing'});const where={all:predicates};(FilterEditor).validate(where,signal);return {signal,environment:q.environment,fromMs:q.fromMs,toMs:q.toMs,where,limit:100,offset:0}}

function resourceLabel(source){return source?.resource?.kind==='host'?'Host '+source.resource.id:source?.resource?.kind==='kubernetes-pod'?'Pod '+source.resource.id:source?.resource?.kind==='kubernetes-container'?'Container '+source.resource.id:null}

export {resourceLabel,scope,matchingRecords,RequestOwner};
