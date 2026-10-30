/* Validate supplied descriptive differences. This module never executes queries,
   chooses populations, merges provenance or manufactures relative changes. */
import {validateSnapshot,isCanonicalRational} from './view-schema.mjs';
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const own=(value,key)=>Object.prototype.hasOwnProperty.call(value,key);
const text=value=>typeof value==='string'&&value.length>0&&value.length<=1024;
const gcd=(a,b)=>{a=a<0n?-a:a;while(b){const rest=a%b;a=b;b=rest}return a};
function difference(baseline,selection){
 const [an,ad='1']=baseline.split('/'),[bn,bd='1']=selection.split('/');
 const numerator=BigInt(bn)*BigInt(ad)-BigInt(an)*BigInt(bd),denominator=BigInt(ad)*BigInt(bd),divisor=gcd(numerator,denominator);
 return denominator/divisor===1n?String(numerator/divisor):`${numerator/divisor}/${denominator/divisor}`;
}
const equal=(a,b)=>{if(a===b)return true;if(!record(a)||!record(b))return false;const keys=Object.keys(a);return keys.length===Object.keys(b).length&&keys.every(key=>own(b,key)&&equal(a[key],b[key]))};
function binding(snapshot){
 const b=snapshot.scope.selector.numeric;
 if(!record(b)||!record(b.field)||!['attribute','resource'].includes(b.field.kind)||!text(b.field.key)||!['int64','uint64','decimal64'].includes(b.type)||b.parse!=='canonical-decimal'||!text(b.unit))return null;
 const scale=b.scale??0;
 if(!Number.isSafeInteger(scale)||(b.type==='decimal64'?scale<1||scale>9:scale!==0))return null;
 return {field:{kind:b.field.kind,key:b.field.key},type:b.type,parse:b.parse,unit:b.unit,scale};
}
export function validateComparison(comparison){
 const reasons=[];
 if(!record(comparison)||comparison.version!=='goo.comparison/1'||!['observed','no_data','diagnostic','incompatible'].includes(comparison.state))return {ok:false,reasons:['A supported comparison envelope and state are required']};
 if(Object.keys(comparison).some(key=>!['version','state','baseline','selection','changes','reason'].includes(key)))reasons.push('Unknown comparison option');
 if(comparison.reason!==undefined&&!text(comparison.reason))reasons.push('Comparison reason requires a bounded string');
 for(const key of ['scope','basis','capturedAt','quality'])if(own(comparison,key))reasons.push('Comparison provenance belongs to each independent snapshot');
 for(const role of ['baseline','selection']){const result=validateSnapshot(comparison[role]);if(!result.ok)reasons.push(...result.reasons.map(reason=>role+': '+reason));if(comparison[role]?.shape!=='summary')reasons.push(role+': comparison requires a population summary')}
 if(reasons.length)return {ok:false,reasons};
 if([comparison.baseline,comparison.selection].some(snapshot=>snapshot.fields.some(field=>['span-status','ratio'].includes(field.aggregation?.family))))return {ok:false,reasons:['This comparison view does not support status-count or fraction measures.']};
 const a=comparison.baseline,b=comparison.selection,ar=a.rows[0],br=b.rows[0],fields=new Map([...a.fields,...b.fields].map(field=>[field.key,field]));
 if(!record(comparison.changes))return {ok:false,reasons:['Comparison changes must cover the declared field union']};
 const changeKeys=Object.keys(comparison.changes);
 if(changeKeys.length!==fields.size||changeKeys.some(key=>!fields.has(key)))reasons.push('Comparison changes must cover the declared field union');
 if(comparison.state==='observed'&&(ar.values.count==='0'||br.values.count==='0'))reasons.push('An empty captured population requires the no_data comparison state');
 if(comparison.state==='no_data'&&ar.values.count!=='0'&&br.values.count!=='0')reasons.push('The no_data comparison state requires an empty captured population');
 const blocked=['diagnostic','incompatible'].includes(comparison.state),numericA=binding(a),numericB=binding(b);
 for(const [key,field] of fields){
  const change=comparison.changes[key],af=a.fields.find(field=>field.key===key),bf=b.fields.find(field=>field.key===key);
  if(!record(change)||Object.keys(change).some(key=>!['state','value','unit','reason','aggregation'].includes(key))||!['derived','unavailable'].includes(change.state)||change.unit!==field.unit||change.reason!==undefined&&!text(change.reason)){reasons.push('Invalid comparison change metadata for '+key);continue}
  const same=af&&bf&&af.type===bf.type&&af.unit===bf.unit&&equal(af.aggregation??null,bf.aggregation??null);
  const bound=field.aggregation?.family!=='numeric'||numericA&&numericB&&equal(numericA,numericB)&&numericA.unit===af?.unit&&numericB.unit===bf?.unit;
  const known=key==='count'||ar.cells?.[key]?.state==='derived'&&br.cells?.[key]?.state==='derived';
  const usable=!blocked&&same&&bound&&known&&a.scope.signal===b.scope.signal;
  if(change.aggregation!==undefined&&!equal(change.aggregation,field.aggregation))reasons.push('Comparison aggregation metadata disagrees with its field');
  if(!usable){if(change.state!=='unavailable'||change.value!==null||!text(change.reason))reasons.push('Unavailable comparison values require null and an explicit reason for '+key);continue}
  const av=ar.values[key],bv=br.values[key];
  if(typeof av!=='string'||typeof bv!=='string'||!isCanonicalRational(av)||!isCanonicalRational(bv)){reasons.push('Comparison operands require exact canonical rational strings');continue}
  if(change.state!=='derived'||!isCanonicalRational(change.value)||change.value!==difference(av,bv))reasons.push('Comparison change must equal exact selection minus baseline for '+key);
 }
 return {ok:reasons.length===0,reasons:[...new Set(reasons)]};
}
