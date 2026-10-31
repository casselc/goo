/* Bounded, IO-free composition intent. The host owns one query and its supplied aspects. */
import * as Aggregate from './view-declaration.mjs';
import * as Timeline from './timeline-declaration.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function closed(value,keys){if(!object(value)||Object.keys(value).some(key=>!keys.includes(key)))throw Error('Unknown or invalid composition fields.');}
export function recordDisplay(value,q){
 validateRecordQuestion(q);closed(value,['version','scope','kind','display']);closed(value.scope,['kind','signal']);closed(value.kind,['name','version']);closed(value.display,['columns','fieldColumns']);
 if(value.version!=='goo.view/1'||value.scope.kind!=='record-query'||value.scope.signal!==q.signal||value.kind.name!=='record-table'||value.kind.version!==1)throw Error('Choose a pinned record table for this query.');
 const {columns,fieldColumns=[]}=value.display;
 if(!Array.isArray(columns)||columns.length<2||columns.length>4||new Set(columns).size!==columns.length||!columns.includes('utc')||!columns.includes('summary')||columns.some(key=>!['utc','service','summary','context'].includes(key)))throw Error('Keep UTC and summary with two to four unique record columns.');
 if(!Array.isArray(fieldColumns)||fieldColumns.length>8||new Set(fieldColumns.map(f=>JSON.stringify([f?.kind,f?.key]))).size!==fieldColumns.length)throw Error('Choose at most eight unique qualified record fields.');
 for(const field of fieldColumns){closed(field,['kind','key']);if(typeof field.key!=='string'||!field.key.isWellFormed())throw Error('Choose a valid qualified record field.');validatePredicate({field,op:'exists'},q.signal)}
 return structuredClone(value);
}
export function validate(value,q){
 validateRecordQuestion(q);closed(value,['version','kind','children']);closed(value.kind,['name','version']);
 if(value.version!=='goo.composition/1'||value.kind.name!=='stack'||value.kind.version!==1||!Array.isArray(value.children)||![2,3].includes(value.children.length))throw Error('Use a pinned stack with aggregate, records and an optional timeline.');
 const roles=new Set();let visible=false;
 for(const child of value.children){closed(child,['aggregate','records','timeline','hidden']);const names=['aggregate','records','timeline'].filter(name=>Object.hasOwn(child,name));if(names.length!==1||child.hidden!==undefined&&typeof child.hidden!=='boolean')throw Error('Choose exactly one typed role per composition child.');const name=names[0];if(roles.has(name))throw Error('Use each composition role once.');roles.add(name);visible ||= child.hidden!==true;if(name==='aggregate')Aggregate.validate(child.aggregate,q);else if(name==='timeline')Timeline.validate(child.timeline,q);else recordDisplay(child.records,q)}
 if(!roles.has('aggregate')||!roles.has('records'))throw Error('Keep aggregate and records children.');
 if(!visible)throw Error('Keep at least one composition child visible.');
 return structuredClone(value);
}
export const aggregate=value=>value?.children?.find(child=>child.aggregate)?.aggregate;
export const records=value=>value?.children?.find(child=>child.records)?.records;
export const timeline=value=>value?.children?.find(child=>child.timeline)?.timeline;
export const role=child=>child.aggregate?'aggregate':child.timeline?'timeline':'records';
export const visible=(value,name)=>!!value.children.find(child=>role(child)===name&&!child.hidden);
export function create(q,declaration,options={columns:['utc','service','summary','context']}){return validate({version:'goo.composition/1',kind:{name:'stack',version:1},children:[{aggregate:declaration},{records:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'record-table',version:1},display:structuredClone(options)}}]},q)}
export function setAggregate(value,next,q){return validate({...value,children:value.children.map(child=>child.aggregate?{...child,aggregate:next}:child)},q)}
export function setRecords(value,next,q){return validate({...value,children:value.children.map(child=>child.records?{...child,records:{...child.records,display:next}}:child)},q)}
export function setHidden(value,name,hidden,q){if(!['aggregate','records','timeline'].includes(name))throw Error('Choose an existing composition role.');return validate({...value,children:value.children.map(child=>role(child)===name?{...child,hidden}:child)},q)}
export function reverse(value,q){return validate({...value,children:[...value.children].reverse()},q)}
// Compatibility concerns configuration only; current inputs and paging stay with
// the host. Browser metadata checks do not replace authoritative Go compilation.
export function reuse(value,source,destination){
 const next=validate(value,source);
 next.children=next.children.map(child=>child.aggregate?{...child,aggregate:Aggregate.reuse(child.aggregate,source,destination)}:child.timeline?{...child,timeline:Timeline.reuse(child.timeline,source,destination)}:{...child,records:recordDisplay(child.records,destination)});
 return validate(next,destination);
}
export function declaration(view,signal){const found=aggregate(view?.composition);return found?.scope.signal===signal?found:view?.displays?.[signal]}
export function declarations(view){const found=aggregate(view?.composition);return found?{[found.scope.signal]:found}:view?.displays||{}}
export function view(value){
 if(!Object.hasOwn(value,'composition'))return structuredClone(value);
 if(!Array.isArray(value.queries)||value.queries.length!==1||['layout','columns','fieldColumns','displays','breakdowns','tail'].some(key=>Object.hasOwn(value,key)))throw Error('Composition owns child presentation and requires one query; omit flat view options.');
 return {...structuredClone(value),composition:validate(value.composition,value.queries[0])};
}

export function setTimeline(value,next,q){return validate({...value,children:timeline(value)?value.children.map(child=>child.timeline?{...child,timeline:next}:child):[{timeline:next},...value.children]},q)}
