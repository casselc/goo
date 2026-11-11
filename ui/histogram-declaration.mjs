/* Display intent for supplied numeric bins; query and range authority stay outside. */
import {validateRecordQuestion} from './question-protocol.mjs';
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function closed(value,keys){if(!object(value)||Object.keys(value).some(key=>!keys.includes(key)))throw Error('Unknown or invalid histogram declaration fields.');}
export function eligible(q){try{validateRecordQuestion(q);return ['logs','spans'].includes(q.signal)&&(q.offset??0)===0&&(q.groupBy?.length??0)===0&&!q.groupTimeSeries&&!q.reportedEdge&&!!q.numeric&&q.numericHistogram?.binning==='powers-of-two'}catch{return false}}
export function validate(value,q){
 if(!eligible(q))throw Error('Choose an ungrouped page-zero numeric histogram question.');
 closed(value,['version','scope','kind','display']);closed(value.scope,['kind','signal']);closed(value.kind,['name','version']);closed(value.display,['category','value','title','requiresComplete']);
 if(value.version!=='goo.view/1'||value.scope.kind!=='record-query'||value.scope.signal!==q.signal||value.kind.version!==1||!['histogram','table'].includes(value.kind.name))throw Error('Choose a pinned histogram or table for this record signal.');
 const d=value.display;if(d.category!=='from'||d.value!=='count')throw Error('Histogram displays use bin starts and captured counts.');
 if(d.title!==undefined&&(typeof d.title!=='string'||new TextEncoder().encode(d.title).length>120||/\p{Cc}/u.test(d.title)||!d.title.isWellFormed()||d.title!==''&&!d.title.trim()))throw Error('Use a bounded valid histogram title.');
 if(d.requiresComplete!==undefined&&typeof d.requiresComplete!=='boolean')throw Error('Complete-only must be a boolean.');
 const next=structuredClone(value);if(next.display.requiresComplete===false)delete next.display.requiresComplete;if(next.display.title==='')delete next.display.title;return next;
}
export function create(q){return validate({version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'histogram',version:1},display:{category:'from',value:'count'}},q)}
export function presentation(value){return {kind:value.kind.name,title:value.display.title||'Numeric value distribution',requiresComplete:value.display.requiresComplete??false}}
export function update(value,props,q){return validate({...value,kind:{name:props.kind??value.kind.name,version:1},display:{...value.display,...Object.fromEntries(['title','requiresComplete'].filter(key=>Object.hasOwn(props,key)).map(key=>[key,props[key]]))}},q)}
export function compatible(source,destination){
 if(!eligible(source)||!eligible(destination)||source.signal!==destination.signal)throw Error('Histogram reuse requires the same record signal and numeric histogram profile.');
 const a=source.numeric,b=destination.numeric;
 if(a.field.kind!==b.field.kind||a.field.key!==b.field.key||a.type!==b.type||(a.scale??0)!==(b.scale??0)||a.unit!==b.unit||a.parse!==b.parse||source.numericHistogram.binning!==destination.numericHistogram.binning)throw Error('Histogram reuse requires the same qualified numeric field, type, scale, unit, parse policy and bin specification.');return true;
}
export function reuse(value,source,destination){const next=validate(value,source);compatible(source,destination);return validate(next,destination)}
