/* Reusable display intent only. Queries belong to the parent exploration view. */
import {validateRecordQuestion} from './question-protocol.mjs';
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function closed(x,keys){if(!object(x)||Object.keys(x).some(k=>!keys.includes(k)))throw Error('Unknown or invalid display declaration fields.');}
export function availableMeasures(q){return ['count',...(q?.numeric?['numeric.mean','numeric.sum','numeric.min','numeric.max']:[]),...(q?.summary==='span-duration'?['duration.mean','duration.min','duration.max','duration.p50','duration.p95','duration.p99']:[])];}
export function eligible(q){return !!q&&q.groupBy?.length===1&&!q.numericHistogram&&!q.groupTimeSeries&&!(q.groupBy[0].kind==='column'&&['body_pattern_v1','duration_bucket'].includes(q.groupBy[0].key));}
export function validate(value,q){
 validateRecordQuestion(q);closed(value,['version','scope','kind','display']);closed(value.scope,['kind','signal']);closed(value.kind,['name','version']);closed(value.display,['category','value','title','scale','order','query','requiresComplete']);
 if(value.version!=='goo.view/1'||value.scope.kind!=='record-query'||value.scope.signal!==q.signal||value.kind.version!==1||!['bar','table'].includes(value.kind.name))throw Error('Choose a pinned bar or table display for this signal.');
 if(!eligible(q))throw Error('Reusable displays currently require an ordinary single-field breakdown.');
 const d=value.display;
 if(d.category!=='category'||!availableMeasures(q).includes(d.value))throw Error('Choose a measure supported by this query.');
 if(d.scale!==undefined&&!['','max','population'].includes(d.scale))throw Error('Choose max or population scale.');
 if(d.value!=='count'&&d.scale==='population')throw Error('Population share requires captured-record count.');
 if(d.order!==undefined&&!['','source','label','value-desc','value-asc'].includes(d.order))throw Error('Choose a supported display order.');
 for(const [k,max]of [['query',256],['title',120]])if(d[k]!==undefined&&(typeof d[k]!=='string'||new TextEncoder().encode(d[k]).length>max||/\p{Cc}/u.test(d[k])||!d[k].isWellFormed()||k==='title'&&d[k]!==''&&!d[k].trim()))throw Error('Display text must be bounded valid text.');
 if(d.requiresComplete!==undefined&&typeof d.requiresComplete!=='boolean')throw Error('Complete-only must be a boolean.');
 return structuredClone(value);
}
export function create(q){return validate({version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'bar',version:1},display:{category:'category',value:'count',scale:'population',order:'value-desc',query:'',requiresComplete:false}},q);}
export function presentation(value){return {...value.display,kind:value.kind.name,title:value.display.title||'',requiresComplete:value.display.requiresComplete??false,scale:value.display.scale||'max',order:value.display.order||'source',query:value.display.query||''};}
export function update(value,props,q){return validate({...value,kind:{name:props.kind,version:1},display:{...value.display,...Object.fromEntries(['value','scale','query','order','requiresComplete'].map(k=>[k,props[k]]))}},q);}

export function withQuery(value,query,q){return validate({...value,display:{...value.display,query}},q);}
