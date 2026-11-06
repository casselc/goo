import {refusal} from './source-refusal.mjs';
/* Exact supplied grouping partitions. No source reads or upstream completeness claim. */
const MAX=18446744073709551615n;
const uint=value=>typeof value==='string'&&/^(0|[1-9][0-9]*)$/.test(value)&&value.length<=20&&BigInt(value)<=MAX;
export const valueKey=value=>value===null?'0':'1'+value;
export function countedState(value,field){
 const data=value||{state:'unloaded'};
 if(!data||typeof data!=='object'||Array.isArray(data)||Object.keys(data).some(key=>!['state','field','items','total','readAt','basis','reads','label','error','refusal'].includes(key))||!['unloaded','loading','error','ready'].includes(data.state))throw Error('Supply a facet state.');
 if(data.refusal!==undefined){if(data.state!=='error')throw Error('Source refusal requires an unavailable facet state.');refusal(data.refusal);}
 if(data.state==='ready'){
  if(data.field?.kind!==field.kind||data.field?.key!==field.key||!Array.isArray(data.items)||data.items.length>1000||!uint(data.total)||!Number.isFinite(Date.parse(data.readAt))||typeof data.basis!=='string'||!data.basis)throw Error('Supply matching bounded counted facet values.');
  let sum=0n;const seen=new Set();
  for(const item of data.items){if(!(item.value===null||typeof item.value==='string'&&item.value.isWellFormed())||!uint(item.count)||BigInt(item.count)===0n||seen.has(item.value))throw Error('Supply unique positive categorical counts.');seen.add(item.value);sum+=BigInt(item.count);}
  if(sum!==BigInt(data.total))throw Error('Supplied facet counts must partition their supplied total.');
  for(const name of ['count','groups']){const receipt=data.reads?.[name];if(!receipt||!Number.isFinite(Date.parse(receipt.readAt))||typeof receipt.basis!=='string'||!/^[a-f0-9]{64}$/.test(receipt.basis))throw Error('Supply both independent read receipts.');}
 }
 return structuredClone(data);
}
export function projection(selected,baseline,field){
 const sides=[countedState(selected,field),countedState(baseline,field)],maps=sides.map(side=>new Map(side.state==='ready'?side.items.map(item=>[item.value,item.count]):[])),values=new Set(maps.flatMap(map=>[...map.keys()]));
 const cell=(index,value)=>sides[index].state!=='ready'?{state:'unknown',count:null}:maps[index].has(value)?{state:'observed',count:maps[index].get(value)}:{state:'absent',count:'0'};
 const rows=[...values].map(value=>({value,selected:cell(0,value),baseline:cell(1,value)}));
 const high=row=>{const a=BigInt(row.selected.count||'0'),b=BigInt(row.baseline.count||'0');return a>b?a:b};
 rows.sort((a,b)=>{const x=high(a),y=high(b),ak=valueKey(a.value),bk=valueKey(b.value);return x===y?(ak<bk?-1:ak>bk?1:0):x>y?-1:1});
 return {rows,max:String(rows.reduce((max,row)=>high(row)>max?high(row):max,0n)),complete:sides.every(side=>side.state==='ready')};
}
