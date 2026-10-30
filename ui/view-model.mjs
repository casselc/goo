/* Public, IO-free dataset contracts. Exact values are never replaced by plotting coordinates. */
const own=(o,k)=>Object.prototype.hasOwnProperty.call(o,k);
const record=o=>o!==null&&typeof o==='object'&&!Array.isArray(o);
const text=v=>typeof v==='string'&&v.length>0&&v.length<=1024;
const member=(v,a)=>a.includes(v);
const timestamp=v=>typeof v==='string'&&/^(?:0|[1-9]\d{0,15})$/.test(v)&&Number(v)<=8640000000000000;
const positiveMS=v=>timestamp(v)&&v!=='0';
const uint64=v=>typeof v==='string'&&/^(?:0|[1-9]\d{0,19})$/.test(v)&&BigInt(v)<=18446744073709551615n;
const numeric=v=>typeof v==='number'?Number.isFinite(v)&&(!Number.isInteger(v)||Number.isSafeInteger(v)):typeof v==='string'&&v.length<=1024&&/^-?(?:0|[1-9]\d*)(?:\.\d+)?$/.test(v)&&Number.isFinite(Number(v))&&(Number(v)!==0||!/[1-9]/.test(v));
// This is the public pilot registry, not a vendor-renderer registry.
export const kinds=Object.freeze({table:Object.freeze({category:'any',numericValue:false}),bar:Object.freeze({shape:'ranked',category:'any',numericValue:true}),line:Object.freeze({shape:'ranked',category:'number',numericValue:true}),'time-bar':Object.freeze({shape:'time-series',category:'timestamp',numericValue:true})});
const result=reasons=>({ok:reasons.length===0,reasons});
const enumField=(o,k,a,r)=>{if(!member(o?.[k],a))r.push(`Invalid ${k}`)};
export function validateSnapshot(s){
 const reasons=[];
 if(!record(s))return result(['A dataset snapshot is required']);
 if(s.version!=='goo.dataset/1')reasons.push('Unsupported dataset version');
 for(const k of ['id','revision'])if(!text(s[k]))reasons.push(`Missing dataset ${k}`);
 if(!member(s.shape,['ranked','time-series']))reasons.push('Unsupported dataset shape');
 if(!record(s.scope)||!text(s.scope.id)||!(text(s.scope.generation)||Number.isSafeInteger(s.scope.generation)&&s.scope.generation>=0)||!member(s.scope.signal,['spans','logs'])||!record(s.scope.selector))reasons.push('A checked scope identity, generation, signal and selector are required');
 const fields=new Map();
 if(!Array.isArray(s.fields)||!s.fields.length||s.fields.length>64)reasons.push('Dataset requires 1–64 fields');
 else for(const f of s.fields){if(!record(f)||!text(f.key)||!text(f.label)||!member(f.type,['string','number','timestamp'])||fields.has(f.key)||f.unit!==undefined&&!text(f.unit)||f.type==='timestamp'&&f.unit!=='ms')reasons.push('Invalid or duplicate field');else fields.set(f.key,f)}
 if(s.dimensions!==undefined&&(!Array.isArray(s.dimensions)||!s.dimensions.length||s.shape!=='ranked'||new Set(s.dimensions).size!==s.dimensions.length||s.dimensions.some(key=>typeof key!=='string'||!fields.has(key))))reasons.push('Ranked dimensions must be distinct declared field keys');
 if(!Array.isArray(s.rows)||s.rows.length>10000)reasons.push('Dataset requires at most 10000 supplied rows');
 else {const keys=new Set();for(const row of s.rows){if(!record(row)||!text(row.key)||keys.has(row.key)||!record(row.values)){reasons.push('Invalid or duplicate stable row key');continue}keys.add(row.key);for(const [key,value] of Object.entries(row.values)){const field=fields.get(key);if(!field){reasons.push(`Undeclared field ${key}`);continue}if(value===null)continue;if(field.type==='number'?!numeric(value):field.type==='timestamp'?!timestamp(value):typeof value!=='string'||value.length>1024)reasons.push(`Invalid ${field.type} value for ${key}`)}}}
 if(!record(s.quality))reasons.push('Dataset quality is required');
 else {enumField(s.quality,'extent',['complete','partial','unknown'],reasons);enumField(s.quality,'basis',['captured-query'],reasons);enumField(s.quality,'value',['recorded','derived','estimated','unknown'],reasons);enumField(s.quality,'sampling',['unknown','none','sampled'],reasons);enumField(s.quality,'freshness',['fixed','live'],reasons)}
 if(s.completeness!==undefined){const c=s.completeness;if(!record(c)||c.basis!=='captured-query'||!text(c.revision)||!text(c.scopeId)||!Array.isArray(c.aspects)||c.aspects.some(a=>!member(a,['groups','intervals']))||new Set(c.aspects).size!==c.aspects.length)reasons.push('Invalid completeness attestation')}
 if(s.shape==='time-series')validateIntervals(s,fields,reasons);
 if(s.quality?.extent==='complete'&&!attested(s))reasons.push(`Complete extent requires matching captured-query ${aspect(s)} attestation`);
 return result([...new Set(reasons)]);
}
function aspect(s){return s.shape==='time-series'?'intervals':'groups'}
function attested(s){const c=s.completeness;return c?.basis===s.quality?.basis&&c?.revision===s.revision&&c?.scopeId===s.scope?.id&&Array.isArray(c?.aspects)&&c.aspects.includes(aspect(s))}
export function eligibleKinds(snapshot){
 if(!validateSnapshot(snapshot).ok)return [];
 if(snapshot.dimensions?.length>1)return ['table'];
 const fields=snapshot.fields,numericFields=fields.filter(f=>f.type==='number');
 return Object.entries(kinds).filter(([,kind])=>(!kind.shape||kind.shape===snapshot.shape)&&(!kind.numericValue||kind.shape==='time-series'||numericFields.length>0&&(kind.category==='number'?numericFields.length>=2:fields.length>=2))).map(([name])=>name);
}
export function validateDisplay(snapshot,display){
 const base=validateSnapshot(snapshot);if(!base.ok)return base;
 const reasons=[];
 if(!record(display)||!own(kinds,display.kind))return result(['Unsupported display kind']);
 if(display.requiresComplete!==undefined&&typeof display.requiresComplete!=='boolean')reasons.push('requiresComplete must be boolean');
 if(snapshot.dimensions?.length>1&&display.kind!=='table')reasons.push('Multiple group dimensions require the exact table until a series mapping is declared');
 const kind=kinds[display.kind];
 if(kind.shape&&kind.shape!==snapshot.shape)reasons.push('Display kind does not support this dataset shape');
 if(display.kind==='time-bar'){
  if((display.start??'fromMs')!=='fromMs'||(display.end??'toMs')!=='toMs'||(display.value??'count')!=='count')reasons.push('Time bars require supplied interval bounds and captured count');
 }else if(kind.numericValue){
  const category=snapshot.fields.find(f=>f.key===display.category),value=snapshot.fields.find(f=>f.key===display.value);
  if(!category||!value||value.type!=='number'||category.key===value.key)reasons.push('Display requires distinct declared category and numeric value fields');
  if(kind.category==='number'&&category?.type!=='number')reasons.push('Line requires an ordered numeric category');
 }
 if(display.requiresComplete){if(snapshot.quality.extent!=='complete'||!attested(snapshot))reasons.push(`Complete captured-query ${aspect(snapshot)} are required for this scope and revision`)}
 return result(reasons);
}
function context(snapshot){const v=validateSnapshot(snapshot);if(!v.ok)throw TypeError(v.reasons.join('; '));return {datasetId:snapshot.id,revision:snapshot.revision,scopeId:snapshot.scope.id,generation:snapshot.scope.generation}}
export function scopeIntent(snapshot,field,value){
 const c=context(snapshot),definition=snapshot.fields.find(f=>f.key===field);
 if(!definition||value===undefined||value!==null&&(definition.type==='number'?!numeric(value):definition.type==='timestamp'?!timestamp(value):typeof value!=='string'||value.length>1024))throw TypeError('Scope filter requires a declared field and typed value');
 if(!snapshot.rows.some(row=>own(row.values,field)&&row.values[field]===value))throw TypeError('Scope filter value is not in the supplied dataset');
 return {type:'scope-filter',...c,field,value,op:value===null?'missing':'eq'};
}
export function selectionIntent(snapshot,key){const c=context(snapshot);if(!snapshot.rows.some(row=>row.key===key))throw TypeError('Selection key is not in this dataset revision');return {type:'select',...c,key}}

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
