/* Bounded, IO-free composition intent. The host owns one query and its supplied aspects. */
import * as CapturedCount from './captured-count-comparison.mjs';
import * as Aggregate from './view-declaration.mjs';
import * as Timeline from './timeline-declaration.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
const object=x=>x!==null&&typeof x==='object'&&!Array.isArray(x);
function closed(value,keys){if(!object(value)||Object.keys(value).some(key=>!keys.includes(key)))throw Error('Unknown or invalid composition fields.');}
export function recordDisplay(value,q){
 validateRecordQuestion(q);closed(value,['version','scope','kind','display']);closed(value.scope,['kind','signal']);closed(value.kind,['name','version']);closed(value.display,['columns','fieldColumns']);
 if(value.version!=='goo.view/1'||value.scope.kind!=='record-query'||value.scope.signal!==q.signal||!['record-table','record-list','record-span-waterfall'].includes(value.kind.name)||value.kind.version!==1)throw Error('Choose a pinned record table, list or span waterfall for this query.');
 if(value.kind.name==='record-span-waterfall'&&q.signal!=='spans')throw Error('Span waterfall requires a spans query.');
 const {columns,fieldColumns=[]}=value.display;
 if(!Array.isArray(columns)||columns.length<2||columns.length>4||new Set(columns).size!==columns.length||!columns.includes('utc')||!columns.includes('summary')||columns.some(key=>!['utc','service','summary','context'].includes(key)))throw Error('Keep UTC and summary with two to four unique record columns.');
 if(!Array.isArray(fieldColumns)||fieldColumns.length>8||new Set(fieldColumns.map(f=>JSON.stringify([f?.kind,f?.key]))).size!==fieldColumns.length)throw Error('Choose at most eight unique qualified record fields.');
 for(const field of fieldColumns){closed(field,['kind','key']);if(typeof field.key!=='string'||!field.key.isWellFormed())throw Error('Choose a valid qualified record field.');validatePredicate({field,op:'exists'},q.signal)}
 return structuredClone(value);
}
export function validate(value,q){
 validateRecordQuestion(q);closed(value,['version','kind','children']);closed(value.kind,['name','version']);
 if(value.version!=='goo.composition/1'||value.kind.name!=='stack'||value.kind.version!==1||!Array.isArray(value.children)||![1,2,3,4].includes(value.children.length))throw Error('Use a pinned records-only stack or aggregate and records with optional timeline/comparison.');
 const roles=new Set();let visible=false;
 for(const child of value.children){closed(child,['aggregate','records','timeline','comparison','hidden']);const names=['aggregate','records','timeline','comparison'].filter(name=>Object.hasOwn(child,name));if(names.length!==1||child.hidden!==undefined&&typeof child.hidden!=='boolean')throw Error('Choose exactly one typed role per composition child.');const name=names[0];if(roles.has(name))throw Error('Use each composition role once.');roles.add(name);visible ||= child.hidden!==true;if(name==='aggregate')Aggregate.validate(child.aggregate,q);else if(name==='timeline')Timeline.validate(child.timeline,q);else if(name==='comparison')CapturedCount.validate(child.comparison,q);else recordDisplay(child.records,q)}
 if(!roles.has('records'))throw Error('Keep a records child.');
 if(!roles.has('aggregate')){if(roles.size!==1||!['logs','spans'].includes(q.signal)||q.numeric||q.numericHistogram||q.reportedEdge||q.groupTimeSeries)throw Error('Records-only compositions require one ordinary logs or spans records child without specialized query directives.');if(q.where)validatePredicate(q.where,q.signal);}
 if(!visible)throw Error('Keep at least one composition child visible.');
 return structuredClone(value);
}
export const recordOnly=value=>Array.isArray(value?.children)&&value.children.length===1&&!!value.children[0]?.records;
export const aggregate=value=>value?.children?.find(child=>child.aggregate)?.aggregate;
export const records=value=>value?.children?.find(child=>child.records)?.records;
export const timeline=value=>value?.children?.find(child=>child.timeline)?.timeline;
export const comparison=value=>value?.children?.find(child=>child.comparison)?.comparison;
export const role=child=>child.aggregate?'aggregate':child.timeline?'timeline':child.comparison?'comparison':'records';
export const visible=(value,name)=>!!value.children.find(child=>role(child)===name&&!child.hidden);
export function create(q,declaration,options={columns:['utc','service','summary','context']}){return validate({version:'goo.composition/1',kind:{name:'stack',version:1},children:[{aggregate:declaration},{records:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'record-table',version:1},display:structuredClone(options)}}]},q)}
export function createRecords(q,options={columns:['utc','service','summary','context']}){return validate({version:'goo.composition/1',kind:{name:'stack',version:1},children:[{records:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'record-table',version:1},display:structuredClone(options)}}]},q)}
export function setAggregate(value,next,q){return validate({...value,children:value.children.map(child=>child.aggregate?{...child,aggregate:next}:child)},q)}
export function setRecords(value,next,q){return validate({...value,children:value.children.map(child=>child.records?{...child,records:{...child.records,display:next}}:child)},q)}
export function setRecordKind(value,name,q){return validate({...value,children:value.children.map(child=>child.records?{...child,records:{...child.records,kind:{name,version:1}}}:child)},q)}
export function setHidden(value,name,hidden,q){if(!['aggregate','records','timeline','comparison'].includes(name))throw Error('Choose an existing composition role.');return validate({...value,children:value.children.map(child=>role(child)===name?{...child,hidden}:child)},q)}
export function reverse(value,q){return validate({...value,children:[...value.children].reverse()},q)}
// Compatibility concerns configuration only; current inputs and paging stay with
// the host. Browser metadata checks do not replace authoritative Go compilation.
export function reuse(value,source,destination){
 const next=validate(value,source);if(CapturedCount.queryVariant(comparison(next)))throw Error('Fixed comparison views must open their original questions; presentation Apply is unavailable.');if(comparison(next))CapturedCount.compatible(source,destination);
 next.children=next.children.map(child=>child.aggregate?{...child,aggregate:Aggregate.reuse(child.aggregate,source,destination)}:child.timeline?{...child,timeline:Timeline.reuse(child.timeline,source,destination)}:child.comparison?{...child,comparison:CapturedCount.validate({...child.comparison,scope:{kind:'record-query',signal:destination.signal}},destination)}:{...child,records:recordDisplay(child.records,destination)});
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

export function setComparison(value,next,q){return validate({...value,children:next?(comparison(value)?value.children.map(child=>child.comparison?{...child,comparison:next}:child):[{comparison:next},...value.children]):value.children.filter(child=>!child.comparison)},q)}

// Explicitly promote a raw records view; query acquisition remains with the host.
export function addBreakdown(value,q,{field,kind='bar',replaceGrouping=false}={}){
 const current=validate(value,q);if(!recordOnly(current))throw Error('Add a breakdown to a records-only view.');
 validatePredicate({field,op:'exists'},q.signal);
 if(!['bar','table'].includes(kind))throw Error('Choose Bars or Table.');if(typeof replaceGrouping!=='boolean')throw Error('Confirm grouping replacement explicitly.');
 const same=q.groupBy?.length===1&&q.groupBy[0].kind===field.kind&&q.groupBy[0].key===field.key;
 if(q.groupBy?.length&&!same&&!replaceGrouping)throw Error('Confirm replacement of the existing grouping.');
 const query={...structuredClone(q),groupBy:[structuredClone(field)],offset:0};
 const declaration=Aggregate.create(query);declaration.kind.name=kind;
 return {query,composition:validate({...current,children:[{aggregate:Aggregate.validate(declaration,query)},...current.children]},query)};
}

// Move both fixed count operands together; the host stages acquisition explicitly.
export function rebaseFixedComparison(value,q,range){
 const current=validate(value,q),pair=comparison(current);if(!CapturedCount.queryVariant(pair))throw Error('Choose a fixed-query count comparison.');
 const {fromMs,toMs}=range||{};if(![fromMs,toMs].every(Number.isSafeInteger)||fromMs<0||toMs<=fromMs||toMs>4102444800000||toMs-fromMs>7200000)throw Error('Choose a valid UTC comparison window of at most 120 minutes, before 2100.');
 const rebase=query=>{const next={...structuredClone(query),fromMs,toMs,offset:0};if(next.bucketMs&&Math.ceil((toMs-fromMs)/next.bucketMs)>240)throw Error('Choose a comparison window with at most 240 existing time buckets.');return next},query=rebase(q),next={...pair,baselineQuery:rebase(pair.baselineQuery)};
 return {query,composition:validate({...current,children:current.children.map(child=>child.comparison?{...child,comparison:next}:child)},query)};
}
