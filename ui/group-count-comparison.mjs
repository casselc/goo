/* Closed supplied group-count comparison profile. No reads, actions or source guarantees. */
import {validateSnapshot} from './view-schema.mjs';
import {compatible} from './captured-count-comparison.mjs';
const record=x=>x!==null&&typeof x==='object'&&!Array.isArray(x),encoder=new TextEncoder(),MAX=18446744073709551615n;
const own=(x,k)=>Object.hasOwn(x,k),admitted=new WeakSet();
function closed(value,keys,required=keys){
 if(!record(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))throw Error('Use plain group-comparison objects.');
 for(const key of Reflect.ownKeys(value)){if(typeof key!=='string'||!keys.includes(key))throw Error('Unknown group-comparison field.');const d=Object.getOwnPropertyDescriptor(value,key);if(!d?.enumerable||!own(d,'value'))throw Error('Group-comparison fields must be data properties.');}
 if(required.some(key=>!own(value,key)))throw Error('Missing group-comparison field.');
}
function text(value,max){return typeof value==='string'&&value.isWellFormed()&&encoder.encode(value).length<=max;}
function reason(value,required=false){if(value===undefined&&!required)return;if(!text(value,1024)||!value.trim())throw Error('Unavailable comparisons require a bounded reason.');}
function uint(value){if(typeof value!=='string'||!/^(0|[1-9][0-9]*)$/.test(value)||value.length>20||BigInt(value)>MAX)throw Error('Group counts require canonical uint64 strings.');return BigInt(value);}
function signed(value){if(typeof value!=='string'||!/^(-?[1-9][0-9]*|0)$/.test(value)||value.length>21||BigInt(value)>MAX||BigInt(value)<-MAX)throw Error('Group changes require bounded canonical signed integers.');return BigInt(value);}
export const categoryKey=value=>value===null?'0':'1'+value;
export const categoryLabel=value=>value===null?'Missing value':value===''?'Empty string':JSON.stringify(value);
export function categoryPreview(value){const label=categoryLabel(value),points=Array.from(label);return {label,preview:points.length>80?points.slice(0,80).join('')+'…':label,long:points.length>80};}
function sideCell(cell,accounting){
 closed(cell,['presence','state','value','reason'],['presence','state','value']);reason(cell.reason);
 if(accounting==='diagnostic'){
  if(!['present','unknown'].includes(cell.presence)||cell.state!=='unavailable'||cell.value!==null)throw Error('Diagnostic group counts must remain unavailable, not aligned zeros.');reason(cell.reason,true);return 0n;
 }
 if(!['present','absent'].includes(cell.presence)||cell.state!=='derived')throw Error('Coherent cells require reported membership or aligned absence.');
 const n=uint(cell.value);if(cell.presence==='present'?n===0n:n!==0n)throw Error('Reported groups are positive; aligned absent groups are zero.');return n;
}
// Descriptor walk protects generic element properties before snapshot validation.
function plainSummary(input){
 const path=new Set();let nodes=0,bytes=0;
 function copy(value,depth){
  if(++nodes>250000||depth>32)throw Error('Group summary exceeds the supplied-data budget.');
  if(value===null||typeof value==='boolean')return value;
  if(typeof value==='number'){if(!Number.isFinite(value))throw Error('Use finite JSON values.');return value}
  if(typeof value==='string'){if(!value.isWellFormed()||(bytes+=encoder.encode(value).length)>8*1024*1024)throw Error('Group summary exceeds the supplied-data budget.');return value}
  if(typeof value!=='object'||path.has(value))throw Error('Use acyclic plain summary data.');
  const array=Array.isArray(value),proto=Object.getPrototypeOf(value);if(array?proto!==Array.prototype:proto!==Object.prototype&&proto!==null)throw Error('Use plain summary data.');
  const descriptors=Object.getOwnPropertyDescriptors(value),keys=Reflect.ownKeys(descriptors);if(array&&keys.length!==value.length+1)throw Error('Use dense summary arrays.');
  path.add(value);const result=array?[]:{};
  for(const key of keys){if(array&&key==='length')continue;const d=descriptors[key];if(typeof key!=='string'||!d.enumerable||!own(d,'value')||array&&(!/^(0|[1-9][0-9]*)$/.test(key)||Number(key)>=value.length))throw Error('Summary fields must be data properties.');bytes+=encoder.encode(key).length;if(bytes>8*1024*1024)throw Error('Group summary exceeds the supplied-data budget.');Object.defineProperty(result,key,{value:copy(d.value,depth+1),enumerable:true,writable:true,configurable:true});}
  path.delete(value);return result;
 }
 return copy(input,0);
}
function check(value){
 closed(value,['version','baseline','selection','field','state','reason','baselineAccounting','selectionAccounting','rows'],['version','baseline','selection','field','state','baselineAccounting','selectionAccounting','rows']);
 if(value.version!=='goo.group-count-comparison/1'||!['observed','no_data','diagnostic'].includes(value.state))throw Error('A supported group-count comparison is required.');reason(value.reason,value.state==='diagnostic');
 closed(value.field,['kind','key']);
 const snapshots={baseline:plainSummary(value.baseline),selection:plainSummary(value.selection)};
 for(const side of ['baseline','selection']){
  const snapshot=snapshots[side],result=validateSnapshot(snapshot);if(!result.ok)throw Error(side+': '+result.reasons.join('; '));
  if(snapshot.shape!=='summary'||snapshot.origin!=='browser-held'||snapshot.quality.extent!=='unknown'||snapshot.quality.sampling!=='unknown'||snapshot.completeness!==undefined)throw Error('Group comparison operands are conservative held summaries.');
  if(!['coherent','diagnostic'].includes(value[side+'Accounting']))throw Error('Specify supplied partition accounting.');
  const group=snapshot.scope.selector.groupBy;if(group?.length!==1||group[0].kind!==value.field.kind||group[0].key!==value.field.key)throw Error('Comparison grouping differs from an operand.');
 }
 compatible(snapshots.baseline.scope.selector,snapshots.selection.scope.selector);
 const diagnostic=value.baselineAccounting==='diagnostic'||value.selectionAccounting==='diagnostic',a=uint(snapshots.baseline.rows[0].values.count),b=uint(snapshots.selection.rows[0].values.count);
 if(value.state!==(diagnostic?'diagnostic':a===0n||b===0n?'no_data':'observed'))throw Error('Comparison state disagrees with operand accounting.');
 if(!Array.isArray(value.rows)||Object.getPrototypeOf(value.rows)!==Array.prototype||value.rows.length>2000||Reflect.ownKeys(value.rows).length!==value.rows.length+1)throw Error('Use at most 2000 dense aligned rows.');
 let sumA=0n,sumB=0n,sumChange=0n,presentA=0,presentB=0;const keys=new Set(),identities=new Set();
 for(let i=0;i<value.rows.length;i++){
  const descriptor=Object.getOwnPropertyDescriptor(value.rows,String(i));if(!descriptor||!descriptor.enumerable||!own(descriptor,'value'))throw Error('Group rows must be ordinary data.');
  const row=descriptor.value;closed(row,['key','value','baseline','selection','change']);
  if(row.value!==null&&!text(row.value,4096)||!text(row.key,4097)||row.key!==categoryKey(row.value)||keys.has(row.key)||identities.has(categoryKey(row.value)))throw Error('Group category/key identity must be unique and exact.');keys.add(row.key);identities.add(categoryKey(row.value));
  if(row.value===null&&value.field.kind==='column'&&value.field.key!=='exception_type')throw Error('Fixed grouping columns cannot be missing.');
  const na=sideCell(row.baseline,value.baselineAccounting),nb=sideCell(row.selection,value.selectionAccounting);sumA+=na;sumB+=nb;if(row.baseline.presence==='present')presentA++;if(row.selection.presence==='present')presentB++;if(presentA>1000||presentB>1000)throw Error('Each supplied partition has at most 1000 reported groups.');
  if(row.baseline.presence!=='present'&&row.selection.presence!=='present')throw Error('Every union row must be reported by an operand.');
  closed(row.change,['state','value','unit','reason'],['state','value','unit']);reason(row.change.reason);
  if(row.change.unit!=='records')throw Error('Group change unit must be records.');
  if(diagnostic){if(row.change.state!=='unavailable'||row.change.value!==null)throw Error('Diagnostic differences are unavailable.');reason(row.change.reason,true)}
  else{if(row.change.state!=='derived'||signed(row.change.value)!==nb-na)throw Error('Group change differs from exact selection minus baseline.');sumChange+=BigInt(row.change.value)}
 }
 if(value.baselineAccounting==='coherent'&&sumA!==a||value.selectionAccounting==='coherent'&&sumB!==b||!diagnostic&&sumChange!==b-a)throw Error('Aligned supplied groups do not account for their captured totals.');
 return {...value,...snapshots,field:{...value.field},rows:value.rows.map(row=>({...row,baseline:{...row.baseline},selection:{...row.selection},change:{...row.change}}))};
}
export function validateGroupComparison(value){try{if(!admitted.has(value))check(value);return {ok:true,reasons:[]}}catch(error){return {ok:false,reasons:[error.message||'Group comparison unavailable.']}}}
function freeze(value){if(value&&typeof value==='object'){for(const item of Object.values(value))freeze(item);Object.freeze(value)}return value;}
export function asGroupComparison(value){if(admitted.has(value))return value;const result=freeze(check(value));admitted.add(result);return result;}
export function projection(comparison,{query='',order='change',page=0,pageSize=50}={}){
 const valid=validateGroupComparison(comparison);if(!valid.ok)throw TypeError(valid.reasons.join('; '));
 if(typeof query!=='string'||query.length>4096||!['change','label'].includes(order)||!Number.isInteger(page)||page<0||pageSize!==50)throw TypeError('Use bounded local group presentation options.');
 const diagnostic=comparison.state==='diagnostic';let maximum=0n;
 if(!diagnostic)for(const row of comparison.rows)for(const side of ['baseline','selection']){const n=BigInt(row[side].value);if(n>maximum)maximum=n}
 const tie=(a,b)=>a.key<b.key?-1:a.key>b.key?1:0,abs=n=>n<0n?-n:n;
 const filtered=comparison.rows.filter(row=>categoryLabel(row.value).toLocaleLowerCase().includes(query.toLocaleLowerCase())).sort((a,b)=>{
  if(order==='change'){const av=a.change.state==='derived'?abs(BigInt(a.change.value)):null,bv=b.change.state==='derived'?abs(BigInt(b.change.value)):null;if(av===null&&bv!==null)return 1;if(av!==null&&bv===null)return -1;if(av!==null&&bv!==null&&av!==bv)return av>bv?-1:1}return tie(a,b)
 });
 const current=Math.min(page,Math.max(0,Math.ceil(filtered.length/pageSize)-1));return {rows:filtered.slice(current*pageSize,(current+1)*pageSize),total:comparison.rows.length,matched:filtered.length,page:current,pageSize,domainMax:diagnostic?null:String(maximum),diagnostic};
}
