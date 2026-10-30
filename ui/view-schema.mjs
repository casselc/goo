import {categoryText} from './view-values.mjs';
export {categoryText};
/* Structural, IO-free dataset schema. Metadata consistency is not authority. Exact values are never replaced by plotting coordinates. */
const own=(o,k)=>Object.prototype.hasOwnProperty.call(o,k);
const record=o=>o!==null&&typeof o==='object'&&!Array.isArray(o);
const text=v=>typeof v==='string'&&v.length>0&&v.length<=1024;
const member=(v,a)=>a.includes(v);
const timestamp=v=>typeof v==='string'&&/^(?:0|[1-9]\d{0,15})$/.test(v)&&Number(v)<=8640000000000000;
const positiveMS=v=>timestamp(v)&&v!=='0';
const uint64=v=>typeof v==='string'&&/^(?:0|[1-9]\d{0,19})$/.test(v)&&BigInt(v)<=18446744073709551615n;
const numeric=v=>typeof v==='number'?Number.isFinite(v)&&(!Number.isInteger(v)||Number.isSafeInteger(v)):typeof v==='string'&&v.length<=1024&&/^-?(?:0|[1-9]\d*)(?:\.\d+)?$/.test(v);
const numericCoordinate=value=>{if(!numeric(value))return null;const n=Number(value);return Number.isFinite(n)&&(n!==0||typeof value==='number'||!/[1-9]/.test(value))?n:null};
const numericType=f=>f?.type==='number'||f?.type==='rational';
const gcd=(a,b)=>{a=a<0n?-a:a;while(b){const next=a%b;a=b;b=next}return a};
function rationalParts(value){
 if(typeof value!=='string'||value.length>1024)return null;
 const match=/^(-?(?:0|[1-9]\d*))(?:\/([1-9]\d*))?$/.exec(value);
 if(!match||match[1]==='-0')return null;
 const numerator=BigInt(match[1]),denominator=match[2]?BigInt(match[2]):1n;
 if(match[2]&&(denominator===1n||gcd(numerator,denominator)!==1n))return null;
 return {numerator,denominator};
}
export const isCanonicalRational=value=>rationalParts(value)!==null;
// Disposable coordinates are approximate. Exact strings remain authoritative.
export function rationalCoordinate(value){
 const parts=rationalParts(value);if(!parts)return null;
 const {numerator,denominator}=parts;if(numerator===0n)return 0;
 const directN=Number(numerator),directD=Number(denominator);
 if(Number.isFinite(directN)&&Number.isFinite(directD)){const coordinate=directN/directD;return Number.isFinite(coordinate)&&coordinate!==0?coordinate:null}
 const n=(numerator<0n?-numerator:numerator).toString(),d=denominator.toString();
 const mantissa=Number('0.'+n.slice(0,16))/Number('0.'+d.slice(0,16)),coordinate=Number(`${numerator<0n?'-':''}${mantissa}e${n.length-d.length}`);
 return Number.isFinite(coordinate)&&coordinate!==0?coordinate:null;
}
function typed(value,field){return field.type==='number'?numeric(value):field.type==='rational'?rationalParts(value)!==null:field.type==='timestamp'?timestamp(value):typeof value==='string'&&value.length<=1024}
function aggregationValid(field){
 const a=field.aggregation;if(!record(a)||!numericType(field)||!text(field.unit))return false;
 if(a.family==='span-status')return Object.keys(a).length===3&&a.kind==='count'&&a.method==='exact'&&field.type==='number'&&field.unit==='records';
 if(a.family==='ratio')return Object.keys(a).length===5&&a.kind==='fraction'&&a.method==='exact-rational'&&field.type==='rational'&&field.unit==='1'&&text(a.numerator)&&text(a.denominator)&&a.numerator!==a.denominator;
 if(Object.keys(a).some(key=>!['family','kind','method','quantile'].includes(key))||!member(a.family,['numeric','duration'])||!member(a.kind,['mean','sum','min','max','quantile']))return false;
 if(a.family==='duration'&&field.unit!=='ns')return false;
 if(a.kind==='mean')return a.method==='exact-rational'&&field.type==='rational'&&a.quantile===undefined;
 if(a.kind==='quantile')return a.family==='duration'&&a.method==='quantilesExactHigh'&&member(a.quantile,['0.5','0.95','0.99']);
 return a.method==='exact'&&a.quantile===undefined;
}
// This is the public pilot registry, not a vendor-renderer registry.
export const kinds=Object.freeze({table:Object.freeze({category:'any',numericValue:false}),bar:Object.freeze({shape:'ranked',category:'any',numericValue:true}),line:Object.freeze({shape:'ranked',category:'number',numericValue:true}),'time-bar':Object.freeze({shape:'time-series',category:'timestamp',numericValue:true})});
const result=reasons=>({ok:reasons.length===0,reasons});
const enumField=(o,k,a,r)=>{if(!member(o?.[k],a))r.push(`Invalid ${k}`)};
export function validateSnapshot(s){
 const reasons=[];
 if(!record(s))return result(['A dataset snapshot is required']);
 if(s.version!=='goo.dataset/1')reasons.push('Unsupported dataset version');
 if(s.origin!==undefined&&!member(s.origin,['source-read','browser-held']))reasons.push('Invalid dataset origin');
 if(s.origin==='browser-held'&&(s.completeness!==undefined||s.quality?.extent==='complete'))reasons.push('Browser-held snapshots cannot attest complete source coverage');
 for(const key of ['sourceBasis','capturedAt'])if(s[key]!==undefined&&!text(s[key]))reasons.push('Invalid optional dataset provenance '+key);
 for(const k of ['id','revision'])if(!text(s[k]))reasons.push(`Missing dataset ${k}`);
 if(!member(s.shape,['ranked','time-series','summary']))reasons.push('Unsupported dataset shape');
 if(!record(s.scope)||!text(s.scope.id)||!(text(s.scope.generation)||Number.isSafeInteger(s.scope.generation)&&s.scope.generation>=0)||!member(s.scope.signal,['spans','logs'])||!record(s.scope.selector))reasons.push('A checked scope identity, generation, signal and selector are required');
 const fields=new Map();
 if(!Array.isArray(s.fields)||!s.fields.length||s.fields.length>64)reasons.push('Dataset requires 1–64 fields');
 else for(const f of s.fields){if(!record(f)||!text(f.key)||!text(f.label)||!member(f.type,['string','number','timestamp','rational'])||fields.has(f.key)||f.unit!==undefined&&!text(f.unit)||f.type==='timestamp'&&f.unit!=='ms'||f.nullable!==undefined&&typeof f.nullable!=='boolean')reasons.push('Invalid or duplicate field');else {fields.set(f.key,f);if(f.aggregation!==undefined&&!aggregationValid(f))reasons.push('Invalid aggregation method, family, type or unit')}}
 if(s.dimensions!==undefined&&(!Array.isArray(s.dimensions)||!s.dimensions.length||s.shape!=='ranked'||new Set(s.dimensions).size!==s.dimensions.length||s.dimensions.some(key=>typeof key!=='string'||!fields.has(key)||fields.get(key)?.aggregation!==undefined)))reasons.push('Ranked dimensions must be distinct declared field keys');
 if(!Array.isArray(s.rows)||s.rows.length>10000)reasons.push('Dataset requires at most 10000 supplied rows');
 else {const keys=new Set();for(const row of s.rows){if(!record(row)||!text(row.key)||keys.has(row.key)||!record(row.values)){reasons.push('Invalid or duplicate stable row key');continue}keys.add(row.key);for(const [key,value] of Object.entries(row.values)){const field=fields.get(key);if(!field){reasons.push(`Undeclared field ${key}`);continue}if(value===null){if(field.nullable!==true)reasons.push('Null values require an explicitly nullable field');continue}if(!typed(value,field))reasons.push(`Invalid ${field.type} value for ${key}`)}}}
 if(s.shape==='ranked'&&Array.isArray(s.rows))for(const row of s.rows){if(fields.has('count')&&!uint64(row?.values?.count))reasons.push('Every ranked row requires its declared exact captured count');if(Array.isArray(s.dimensions)&&s.dimensions.some(key=>!record(row?.values)||!own(row.values,key)))reasons.push('Every declared group dimension requires a supplied cell value')}
 if(s.partition!==undefined)validatePartition(s,fields,reasons);
 if(Array.isArray(s.rows)){validateCells(s,fields,reasons);validateCountRelationships(s,fields,reasons)}
 if(!record(s.quality))reasons.push('Dataset quality is required');
 else {enumField(s.quality,'extent',['complete','partial','unknown'],reasons);enumField(s.quality,'basis',['captured-query'],reasons);enumField(s.quality,'value',['recorded','derived','estimated','unknown'],reasons);enumField(s.quality,'sampling',['unknown','none','sampled'],reasons);enumField(s.quality,'freshness',['fixed','live'],reasons)}
 if(s.completeness!==undefined){const c=s.completeness;if(!record(c)||c.basis!=='captured-query'||!text(c.revision)||!text(c.scopeId)||!Array.isArray(c.aspects)||c.aspects.some(a=>!member(a,['groups','intervals','summary']))||new Set(c.aspects).size!==c.aspects.length)reasons.push('Invalid completeness attestation')}
 if(s.shape==='time-series')validateIntervals(s,fields,reasons);
 if(s.shape==='summary')validateSummary(s,fields,reasons);
 if(s.quality?.extent==='complete'&&!attested(s))reasons.push(`Complete extent requires matching captured-query ${aspect(s)} attestation`);
 return result([...new Set(reasons)]);
}
function aspect(s){return s.shape==='time-series'?'intervals':s.shape==='summary'?'summary':'groups'}
function attested(s){const c=s.completeness;return c?.basis===s.quality?.basis&&c?.revision===s.revision&&c?.scopeId===s.scope?.id&&Array.isArray(c?.aspects)&&c.aspects.includes(aspect(s))}
export function eligibleKinds(snapshot){
 if(!validateSnapshot(snapshot).ok)return [];
 if(snapshot.dimensions?.length>1)return ['table'];
 const fields=snapshot.fields,numericFields=fields.filter(numericType);
 return Object.entries(kinds).filter(([,kind])=>(!kind.shape||kind.shape===snapshot.shape)&&(!kind.numericValue||kind.shape==='time-series'||numericFields.length>0&&(kind.category==='number'?numericFields.length>=2&&fields.some(field=>field.type==='number'&&(!snapshot.dimensions||snapshot.dimensions.includes(field.key))):fields.length>=2))).map(([name])=>name);
}
export function validateDisplay(snapshot,display){
 const base=validateSnapshot(snapshot);if(!base.ok)return base;
 const reasons=[];
 if(!record(display)||typeof display.kind!=='string'||!own(kinds,display.kind))return result(['Unsupported display kind']);
 if(Object.keys(display).some(key=>!['kind','title','category','value','start','end','requiresComplete','scale','query','order'].includes(key)))reasons.push('Unknown display option');
 for(const key of ['title','category','value','start','end'])if(display[key]!==undefined&&!text(display[key]))reasons.push('Display options require bounded strings');
 if(display.scale!==undefined&&!member(display.scale,['max','population']))reasons.push('Choose max or population scale');
 if(display.query!==undefined&&(typeof display.query!=='string'||display.query.length>256))reasons.push('Find text must contain at most 256 characters');
 if(display.order!==undefined&&!member(display.order,['source','label','value-desc','value-asc']))reasons.push('Choose a supported display order');
 if((display.scale!==undefined||display.query!==undefined||display.order!==undefined)&&(snapshot.shape!=='ranked'||!member(display.kind,['bar','table'])))reasons.push('Find, ordering and scale options require a ranked table or bar view');
 if(display.scale==='population'){
  if(display.value!=='count')reasons.push('Population shares are available only for captured record counts');
  if(!snapshot.partition)reasons.push('Captured-population shares are unavailable for this result');
  else if(snapshot.partition.total==='0')reasons.push('No captured population is available for shares');
 }
 if(display.requiresComplete!==undefined&&typeof display.requiresComplete!=='boolean')reasons.push('requiresComplete must be boolean');
 if(snapshot.dimensions?.length>1&&display.kind!=='table')reasons.push('Multiple group dimensions require the exact table until a series mapping is declared');
 const kind=kinds[display.kind];
 if(kind.shape&&kind.shape!==snapshot.shape)reasons.push('Display kind does not support this dataset shape');
 if(display.kind==='time-bar'){
  const value=snapshot.fields.find(field=>field.key===(display.value??'count'));
  if((display.start??'fromMs')!=='fromMs'||(display.end??'toMs')!=='toMs'||!value||!numericType(value)||value.key!=='count'&&!value.aggregation)reasons.push('Time bars require supplied interval bounds and a declared captured measure');
 }else if(kind.numericValue){
  const category=snapshot.fields.find(f=>f.key===display.category),value=snapshot.fields.find(f=>f.key===display.value);
  if(!category||!value||!numericType(value)||category.key===value.key)reasons.push('Display requires distinct declared category and numeric value fields');
  if(kind.category==='number'&&(category?.type!=='number'||snapshot.dimensions&&!snapshot.dimensions.includes(category?.key)))reasons.push('Line requires an ordered numeric category dimension');
 }
 if(kind.numericValue){const valueKey=display.value??(display.kind==='time-bar'?'count':undefined);if(valueKey)for(const row of snapshot.rows){const value=measureValue(snapshot,row,valueKey);if(member(value.state,['recorded','derived','estimated'])&&value.coordinate===null)reasons.push('A known measure is outside the finite plotting range; use exact values')}}
 if(display.requiresComplete){if(snapshot.quality.extent!=='complete'||!attested(snapshot))reasons.push(`Complete captured-query ${aspect(snapshot)} are required for this scope and revision`)}
 return result(reasons);
}
function context(snapshot){const v=validateSnapshot(snapshot);if(!v.ok)throw TypeError(v.reasons.join('; '));return {datasetId:snapshot.id,revision:snapshot.revision,scopeId:snapshot.scope.id,generation:snapshot.scope.generation}}
export function scopeIntent(snapshot,field,value){
 const c=context(snapshot),definition=snapshot.fields.find(f=>f.key===field);
 if(snapshot.shape==='summary'||!definition||definition.aggregation||snapshot.dimensions&&!snapshot.dimensions.includes(field)||value===undefined||value!==null&&!typed(value,definition))throw TypeError('Scope filter requires a declared field and typed value');
 if(!snapshot.rows.some(row=>own(row.values,field)&&row.values[field]===value))throw TypeError('Scope filter value is not in the supplied dataset');
 return {type:'scope-filter',...c,field,value,op:value===null?'missing':'eq'};
}
export function selectionIntent(snapshot,key){const c=context(snapshot);if(snapshot.shape==='summary')throw TypeError('Population summaries do not identify record or subgroup membership');if(!snapshot.rows.some(row=>row.key===key))throw TypeError('Selection key is not in this dataset revision');return {type:'select',...c,key}}

// Complete means exhaustive captured aggregate intervals, never full collection.
function validateIntervals(s,fields,reasons){
 const interval=s.interval;
 if(!record(interval)||!timestamp(interval.fromMs)||!timestamp(interval.toMs)||!positiveMS(interval.bucketMs)||BigInt(interval.toMs)<=BigInt(interval.fromMs)){reasons.push('A bounded half-open interval and positive millisecond bucket are required');return}
 for(const key of ['fromMs','toMs'])if(fields.get(key)?.type!=='timestamp'||fields.get(key)?.unit!=='ms')reasons.push('Time-series bounds require timestamp fields in milliseconds');
 if(fields.get('count')?.type!=='number'||fields.get('count')?.unit!=='records')reasons.push('Time-series counts require a numeric records field');
 if(!Array.isArray(s.rows))return;
 const from=BigInt(interval.fromMs),to=BigInt(interval.toMs),bucket=BigInt(interval.bucketMs);
 let previous=from,covered=true;
 for(const row of s.rows){
  const values=row?.values;
  if(!record(values)||!timestamp(values.fromMs)||!timestamp(values.toMs)){reasons.push('Every supplied interval requires exact timestamp bounds');covered=false;continue}
  const start=BigInt(values.fromMs),end=BigInt(values.toMs),expectedEnd=start+bucket<to?start+bucket:to;
  if(start<from||end>to||end<=start||start<previous||(start-from)%bucket!==0n||end!==expectedEnd)reasons.push('Intervals must be ordered, nonoverlapping and match the bounded bucket grid');
  if(start!==previous)covered=false;
  previous=end;
  if(!member(row.state,['observed','no_data','unknown']))reasons.push('Every interval requires an observed, no_data or unknown state');
  if(row.state==='observed'&&(!uint64(values.count)||values.count==='0'))reasons.push('Observed intervals require a positive exact record count');
  if(row.state==='no_data'&&values.count!=='0')reasons.push('No-data intervals require an exact zero captured count');
  if(row.state==='unknown'){covered=false;if(values.count!==null&&values.count!==undefined)reasons.push('Unknown intervals cannot supply a captured count')}
 }
 if(previous!==to)covered=false;
 if(s.quality?.extent==='complete'&&!covered)reasons.push('Complete interval coverage cannot contain gaps or unknown buckets');
}
export function intervalIntent(snapshot,key){
 const c=context(snapshot),row=snapshot.rows.find(row=>row.key===key);
 if(snapshot.shape!=='time-series'||!row||row.state==='unknown')throw TypeError('Window selection requires a known supplied interval in this dataset revision');
 return {type:'scope-window',...c,key,fromMs:row.values.fromMs,toMs:row.values.toMs};
}

export function groupIntent(snapshot,key){
 const c=context(snapshot),row=snapshot.rows.find(row=>row.key===key);
 if(snapshot.shape!=='ranked'||!snapshot.dimensions?.length||!row)throw TypeError('Group selection requires declared dimensions and a supplied row in this dataset revision');
 const values=snapshot.dimensions.map(field=>{if(!own(row.values,field))throw TypeError('Group dimension value was not supplied');const value=row.values[field];return {field,value,op:value===null?'missing':'eq'}});
 return {type:'scope-group',...c,key,values};
}

function validateCells(snapshot,fields,reasons){
 const measures=[...fields.values()].filter(field=>field.aggregation);
 for(const row of snapshot.rows){
  if(!record(row)||!record(row.values))continue;
  if(row.cells!==undefined&&!record(row.cells)){reasons.push('Measure cells must be keyed by declared aggregations');continue}
  for(const key of Object.keys(row.cells??{}))if(!fields.get(key)?.aggregation)reasons.push('Undeclared measure cell');
  for(const field of measures){
   const cell=row.cells?.[field.key],value=row.values[field.key];
   if(!record(cell)||!member(cell.state,['derived','estimated','no_data','no_valid_values','unknown'])){reasons.push('Every aggregation requires an explicit measure cell state');continue}
   if(!own(row.values,field.key)){reasons.push('Every aggregation requires a supplied exact value or explicit null');continue}
   if(field.aggregation.family==='span-status'||field.aggregation.family==='ratio'){
    if(cell.support!==undefined)reasons.push('Count relationships do not use numeric parse support');
    continue;
   }
   const known=cell.state==='derived'||cell.state==='estimated';
   if(known?value===null||!typed(value,field):value!==null)reasons.push('Measure value and state disagree; unavailable measures require null');
   if(field.aggregation.family==='duration'&&value!==null){
    if(field.type==='rational'?rationalParts(value)?.numerator<0n:typed(value,field)&&Number(value)<0)reasons.push('Duration measures cannot be negative');
    if(field.aggregation.kind!=='mean'&&(typeof value==='string'?!/^(?:0|[1-9]\d*)$/.test(value):!Number.isSafeInteger(value)))reasons.push('Exact duration sums, extrema and empirical quantiles require integer nanoseconds');
    if(member(field.aggregation.kind,['min','max','quantile'])&&(typeof value==='string'?!uint64(value):value<0||!Number.isSafeInteger(value)))reasons.push('Duration extrema and empirical quantiles require uint64 nanoseconds');
   }
   const support=cell.support,count=row.values.count;
   if(cell.state==='unknown'&&(count===null||count===undefined)){if(support!==undefined)reasons.push('Unknown count cannot invent measure support');continue}
   const keys=['total','valid','missing','empty','invalid'];
   if(!record(support)||keys.some(key=>!uint64(support[key]))||!uint64(count)){reasons.push('Known measure support requires exact captured record counts');continue}
   const n=Object.fromEntries(keys.map(key=>[key,BigInt(support[key])]));
   if(n.total!==n.valid+n.missing+n.empty+n.invalid||support.total!==count)reasons.push('Measure support must partition the exact captured row count');
   if(known&&n.valid===0n||cell.state==='no_data'&&n.total!==0n||cell.state==='no_valid_values'&&(n.total===0n||n.valid!==0n))reasons.push('Measure support and availability state disagree');
  }
 }
}
// Callers validate the supplied snapshot once; this pure projection performs no
// source reads and never replaces the exact value with its plotting coordinate.
export function measureValue(snapshot,row,fieldKey){
 const field=snapshot.fields?.find(field=>field.key===fieldKey),cell=row?.cells?.[fieldKey],raw=row?.values?.[fieldKey];
 if(!field||!numericType(field))return {state:'unknown',exact:null,coordinate:null,unit:field?.unit??'',aggregation:field?.aggregation??null,support:null,reason:'No numeric measure is declared'};
 const state=field.aggregation?(cell?.state??'unknown'):row.state==='unknown'||raw===null||raw===undefined?'unknown':row.state==='no_data'?'no_data':snapshot.quality?.value??'unknown';
 const exact=raw===undefined?null:raw,known=member(state,['recorded','derived','estimated'])||!field.aggregation&&state==='no_data';
 const coordinate=!known||raw===null||raw===undefined?null:field.type==='rational'?rationalCoordinate(raw):numericCoordinate(raw);
 return {state,exact,coordinate,unit:field.unit??'',aggregation:field.aggregation??null,support:cell?.support??null,...(known&&coordinate===null?{reason:'Exact value is outside the finite plotting range'}:{})};
}

function validateSummary(snapshot,fields,reasons){
 if(!member(snapshot.origin,['source-read','browser-held']))reasons.push('Population summaries require an explicit source-read or browser-held origin');
 if(snapshot.dimensions!==undefined||snapshot.interval!==undefined)reasons.push('Population summaries cannot declare subgroup dimensions or interval geometry');
 if(!Array.isArray(snapshot.rows)||snapshot.rows.length!==1||snapshot.rows[0]?.key!=='population'||!uint64(snapshot.rows[0]?.values?.count))reasons.push('Population summaries require exactly one population row with an exact captured count');
 if(fields.get('count')?.type!=='number'||fields.get('count')?.unit!=='records'||fields.get('count')?.aggregation)reasons.push('Population summaries require a count field in captured records');
 for(const field of fields.values())if(field.key!=='count'&&(!field.aggregation||field.type!=='rational'&&field.aggregation.family!=='span-status'))reasons.push('Population summaries contain only captured count and declared exact rational aggregate measures');
}

const shapeCapabilities=Object.freeze({ranked:Object.freeze({selectRows:true,filterDimensions:true,selectWindows:false}), 'time-series':Object.freeze({selectRows:true,filterDimensions:false,selectWindows:true}), summary:Object.freeze({selectRows:false,filterDimensions:false,selectWindows:false})});
const none=Object.freeze({selectRows:false,filterDimensions:false,selectWindows:false,completeAspects:Object.freeze([])});
export function capabilities(snapshot){
 if(!validateSnapshot(snapshot).ok)return none;
 const shape=shapeCapabilities[snapshot.shape];
 return Object.freeze({...shape,filterDimensions:shape.filterDimensions&&!!snapshot.dimensions?.length,completeAspects:Object.freeze(snapshot.quality.extent==='complete'&&attested(snapshot)?[aspect(snapshot)]:[])});
}
export const supportsSelection=snapshot=>capabilities(snapshot).selectRows;

function validatePartition(snapshot,fields,reasons){
 const p=snapshot.partition,count=fields.get('count');
 if(snapshot.shape!=='ranked'||!Array.isArray(snapshot.dimensions)||!snapshot.dimensions.length||!record(p)||Object.keys(p).length!==3||!['kind','measure','total'].every(key=>own(p,key))||p.kind!=='disjoint-record-groups'||p.measure!=='count'||!uint64(p.total)||count?.type!=='number'||count.unit!=='records'||count.aggregation){reasons.push('Invalid disjoint captured-record partition');return}
 if(!Array.isArray(snapshot.rows)||snapshot.rows.some(row=>!uint64(row?.values?.count))){reasons.push('Partition rows require exact captured counts');return}
 const sum=snapshot.rows.reduce((total,row)=>total+BigInt(row.values.count),0n),total=BigInt(p.total);
 if(sum>18446744073709551615n)reasons.push('Partition sum exceeds the captured-count range');
 if(sum>total||snapshot.quality?.extent==='complete'&&sum!==total)reasons.push('Partition total does not match its captured group extent');
}
function comparable(value){
 if(value===null||value===undefined)return null;
 const text=String(value),fraction=rationalParts(text);if(fraction)return fraction;
 const match=/^(-?)(\d+)(?:\.(\d+))?(?:[eE]([+-]?\d+))?$/.exec(text);if(!match)return null;
 const shift=Number(match[4]??0)-(match[3]?.length??0);if(!Number.isSafeInteger(shift)||Math.abs(shift)>2048)return null;
 const n=BigInt(match[2]+(match[3]??''))*(match[1]?-1n:1n);
 return shift<0?{numerator:n,denominator:10n**BigInt(-shift)}:{numerator:n*10n**BigInt(shift),denominator:1n};
}
// Display-only Find/order. These rows still belong to the original snapshot:
// no scope, denominator, guarantee, dataset identity or source totals change.
export function displayRows(snapshot,display){
 const dimensions=snapshot.dimensions?.length?snapshot.dimensions:display.category?[display.category]:[],query=(display.query??'').toLowerCase(),order=display.order??'source',valueKey=display.value??'count';
 const labels=row=>dimensions.map(key=>categoryText(row.values[key])).join(' / ').toLowerCase();
 const byValue=order==='value-asc'||order==='value-desc';
 const rows=snapshot.rows.map((row,index)=>{const value=byValue?measureValue(snapshot,row,valueKey):null,known=value&&member(value.state,['recorded','derived','estimated','no_data'])&&value.exact!==null;return {row,index,label:labels(row),number:known?comparable(value.exact):null}}).filter(item=>!query||item.label.includes(query));
 if(order!=='source')rows.sort((a,b)=>{
  let compared=0;
  if(order==='label')compared=a.label<b.label?-1:a.label>b.label?1:0;
  else {const x=a.number,y=b.number;
   if(!x||!y)return x?-1:y?1:a.index-b.index;
   const delta=x.numerator*y.denominator-y.numerator*x.denominator;compared=delta<0n?-1:delta>0n?1:0;if(order==='value-desc')compared=-compared;
  }
  return compared||a.index-b.index;
 });
 return Object.freeze(rows.map(item=>item.row));
}

const statusKeys=['status.error','status.ok','status.unset','status.unknown'];
function countField(field){return field?.type==='number'&&field.unit==='records'&&(!field.aggregation||field.aggregation.family==='span-status'&&field.aggregation.kind==='count'&&field.aggregation.method==='exact')}
function validateCountRelationships(snapshot,fields,reasons){
 const statuses=[...fields.values()].filter(field=>field.aggregation?.family==='span-status'),fractions=[...fields.values()].filter(field=>field.aggregation?.family==='ratio');
 if(statuses.length&&(fields.get('count')?.type!=='number'||fields.get('count')?.unit!=='records'||fields.get('count')?.aggregation||snapshot.scope?.signal!=='spans'||statuses.length!==statusKeys.length||!statusKeys.every(key=>fields.get(key)?.aggregation?.family==='span-status')))reasons.push('Span-status counts require all four exact span classes');
 for(const field of fractions)if(!countField(fields.get(field.aggregation.numerator))||!countField(fields.get(field.aggregation.denominator)))reasons.push('Fractions must reference declared exact captured-count fields');
 for(const row of snapshot.rows){
  if(!record(row?.values))continue;
  if(statuses.length){
   const count=row.values.count,known=uint64(count);let sum=0n,valid=true;
   for(const field of statuses){const value=row.values[field.key],cell=row.cells?.[field.key];
    if(!known){if(value!==null||cell?.state!=='unknown')reasons.push('Unknown captured counts require unknown null span classes')}
    else if(!uint64(value)||cell?.state!=='derived'){reasons.push('Known span classes require exact derived captured counts');valid=false}
    else sum+=BigInt(value);
   }
   if(known&&valid&&sum!==BigInt(count))reasons.push('Span-status classes must partition the captured record count');
  }
  for(const field of fractions){
   const a=field.aggregation,n=row.values[a.numerator],d=row.values[a.denominator],value=row.values[field.key],cell=row.cells?.[field.key];
   if(n===null||n===undefined||d===null||d===undefined){if(value!==null||cell?.state!=='unknown')reasons.push('Unknown fraction operands require a null unknown value');continue}
   if(!uint64(n)||!uint64(d)){reasons.push('Fraction operands must be exact captured counts');continue}
   const numerator=BigInt(n),denominator=BigInt(d);
   if(numerator>denominator){reasons.push('Subset fractions require numerator at most denominator');continue}
   if(denominator===0n){if(value!==null||cell?.state!=='no_data')reasons.push('A zero denominator requires a null no-data fraction');continue}
   const parts=rationalParts(value);
   if(cell?.state!=='derived'||!parts||parts.numerator<0n||parts.numerator>parts.denominator||parts.numerator*denominator!==numerator*parts.denominator)reasons.push('Fraction must equal its exact same-row numerator divided by denominator');
  }
 }
}
// Reference facts for percent displays; no synthetic valid/missing/invalid support.
export function fractionFacts(snapshot,row,fieldKey){
 const field=snapshot.fields.find(field=>field.key===fieldKey),a=field?.aggregation;
 if(a?.family!=='ratio'||a.kind!=='fraction')return null;
 const fact=key=>{const f=snapshot.fields.find(field=>field.key===key);return {field:key,label:f?.label??key,value:row.values[key]??null,state:row.cells?.[key]?.state??(row.values[key]===null||row.values[key]===undefined?'unknown':'recorded')}};
 const numerator=fact(a.numerator),denominator=fact(a.denominator),family=snapshot.fields.find(field=>field.key===a.numerator)?.aggregation?.family;
 return {numerator,denominator,partition:family==='span-status'?snapshot.fields.filter(field=>field.aggregation?.family==='span-status').map(field=>fact(field.key)):[]};
}
