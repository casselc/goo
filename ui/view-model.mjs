/* Public, IO-free ranked-dataset pilot. Exact values are never replaced by plotting coordinates. */
const own=(o,k)=>Object.prototype.hasOwnProperty.call(o,k);
const record=o=>o!==null&&typeof o==='object'&&!Array.isArray(o);
const text=v=>typeof v==='string'&&v.length>0&&v.length<=1024;
const member=(v,a)=>a.includes(v);
const numeric=v=>typeof v==='number'?Number.isFinite(v)&&(!Number.isInteger(v)||Number.isSafeInteger(v)):typeof v==='string'&&v.length<=1024&&/^-?(?:0|[1-9]\d*)(?:\.\d+)?$/.test(v)&&Number.isFinite(Number(v))&&(Number(v)!==0||!/[1-9]/.test(v));
// This is the public pilot registry, not a vendor-renderer registry.
export const kinds=Object.freeze({table:Object.freeze({category:'any',numericValue:false}),bar:Object.freeze({category:'any',numericValue:true}),line:Object.freeze({category:'number',numericValue:true})});
const result=reasons=>({ok:reasons.length===0,reasons});
const enumField=(o,k,a,r)=>{if(!member(o?.[k],a))r.push(`Invalid ${k}`)};
export function validateSnapshot(s){
 const reasons=[];
 if(!record(s))return result(['A dataset snapshot is required']);
 if(s.version!=='goo.dataset/1')reasons.push('Unsupported dataset version');
 for(const k of ['id','revision'])if(!text(s[k]))reasons.push(`Missing dataset ${k}`);
 if(s.shape!=='ranked')reasons.push('Unsupported dataset shape');
 if(!record(s.scope)||!text(s.scope.id)||!(text(s.scope.generation)||Number.isSafeInteger(s.scope.generation)&&s.scope.generation>=0)||!member(s.scope.signal,['spans','logs'])||!record(s.scope.selector))reasons.push('A checked scope identity, generation, signal and selector are required');
 const fields=new Map();
 if(!Array.isArray(s.fields)||!s.fields.length||s.fields.length>64)reasons.push('Dataset requires 1–64 fields');
 else for(const f of s.fields){if(!record(f)||!text(f.key)||!text(f.label)||!member(f.type,['string','number'])||fields.has(f.key)||f.unit!==undefined&&!text(f.unit))reasons.push('Invalid or duplicate field');else fields.set(f.key,f)}
 if(!Array.isArray(s.rows)||s.rows.length>10000)reasons.push('Dataset requires at most 10000 supplied rows');
 else {const keys=new Set();for(const row of s.rows){if(!record(row)||!text(row.key)||keys.has(row.key)||!record(row.values)){reasons.push('Invalid or duplicate stable row key');continue}keys.add(row.key);for(const [key,value] of Object.entries(row.values)){const field=fields.get(key);if(!field){reasons.push(`Undeclared field ${key}`);continue}if(value===null)continue;if(field.type==='number'?!numeric(value):typeof value!=='string'||value.length>1024)reasons.push(`Invalid ${field.type} value for ${key}`)}}}
 if(!record(s.quality))reasons.push('Dataset quality is required');
 else {enumField(s.quality,'extent',['complete','partial','unknown'],reasons);enumField(s.quality,'basis',['captured-query'],reasons);enumField(s.quality,'value',['recorded','derived','estimated','unknown'],reasons);enumField(s.quality,'sampling',['unknown','none','sampled'],reasons);enumField(s.quality,'freshness',['fixed','live'],reasons)}
 if(s.completeness!==undefined){const c=s.completeness;if(!record(c)||c.basis!=='captured-query'||!text(c.revision)||!text(c.scopeId)||!Array.isArray(c.aspects)||c.aspects.some(a=>a!=='groups')||new Set(c.aspects).size!==c.aspects.length)reasons.push('Invalid completeness attestation')}
 if(s.quality?.extent==='complete'&&!attested(s))reasons.push('Complete extent requires matching captured-query groups attestation');
 return result([...new Set(reasons)]);
}
function attested(s){const c=s.completeness;return c?.basis===s.quality?.basis&&c?.revision===s.revision&&c?.scopeId===s.scope?.id&&Array.isArray(c?.aspects)&&c.aspects.includes('groups')}
export function eligibleKinds(snapshot){
 if(!validateSnapshot(snapshot).ok)return [];
 const fields=snapshot.fields,numericFields=fields.filter(f=>f.type==='number');
 return Object.entries(kinds).filter(([,kind])=>!kind.numericValue||numericFields.length>0&&(kind.category==='number'?numericFields.length>=2:fields.length>=2)).map(([name])=>name);
}
export function validateDisplay(snapshot,display){
 const base=validateSnapshot(snapshot);if(!base.ok)return base;
 const reasons=[];
 if(!record(display)||!own(kinds,display.kind))return result(['Unsupported display kind']);
 if(display.requiresComplete!==undefined&&typeof display.requiresComplete!=='boolean')reasons.push('requiresComplete must be boolean');
 const kind=kinds[display.kind];
 if(kind.numericValue){
  const category=snapshot.fields.find(f=>f.key===display.category),value=snapshot.fields.find(f=>f.key===display.value);
  if(!category||!value||value.type!=='number'||category.key===value.key)reasons.push('Display requires distinct declared category and numeric value fields');
  if(kind.category==='number'&&category?.type!=='number')reasons.push('Line requires an ordered numeric category');
 }
 if(display.requiresComplete){if(snapshot.quality.extent!=='complete'||!attested(snapshot))reasons.push('Complete captured-query groups are required for this scope and revision')}
 return result(reasons);
}
function context(snapshot){const v=validateSnapshot(snapshot);if(!v.ok)throw TypeError(v.reasons.join('; '));return {datasetId:snapshot.id,revision:snapshot.revision,scopeId:snapshot.scope.id,generation:snapshot.scope.generation}}
export function scopeIntent(snapshot,field,value){
 const c=context(snapshot),definition=snapshot.fields.find(f=>f.key===field);
 if(!definition||value===undefined||value!==null&&(definition.type==='number'?!numeric(value):typeof value!=='string'||value.length>1024))throw TypeError('Scope filter requires a declared field and typed value');
 if(!snapshot.rows.some(row=>own(row.values,field)&&row.values[field]===value))throw TypeError('Scope filter value is not in the supplied dataset');
 return {type:'scope-filter',...c,field,value,op:value===null?'missing':'eq'};
}
export function selectionIntent(snapshot,key){const c=context(snapshot);if(!snapshot.rows.some(row=>row.key===key))throw TypeError('Selection key is not in this dataset revision');return {type:'select',...c,key}}
