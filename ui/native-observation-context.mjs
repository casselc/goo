/* Supplied ordinary gauge result selection, not a metric evaluator or source authority. */
import {validateGaugeMeasurementQuestion,sameNativeStream} from './measurement-scope.mjs';
import * as Environment from './record-environment.mjs';
export const radiusNS='60000000000';
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value),own=(value,key)=>Object.hasOwn(value,key);
const fail=()=>{throw Error('Choose a unique retained gauge observation in the current service measurement.');};
const text=(value,max)=>typeof value==='string'&&value.isWellFormed()&&!value.includes('\0')&&new TextEncoder().encode(value).length<=max;
function timestamp(value){if(typeof value!=='string'||!/^(0|[1-9]\d{0,18})$/.test(value)||BigInt(value)>=4102444800000000000n)fail();return BigInt(value)}
export function plan(result,stream,signal,atNS){
 if(!object(result)||!['logs','spans'].includes(signal)||result.question?.definition?.kind!=='gauge')fail();
 const q=result.question,d=q.definition,b=d.source,bounds=validateGaugeMeasurementQuestion(q),lo=BigInt(bounds.fromNS),hi=BigInt(bounds.toNS),at=timestamp(atNS);
 if(at<lo||at>=hi||b.resource!=null||b.exactStream&&!sameNativeStream(b.exactStream,stream))fail();
 const selected=structuredClone(q);selected.definition.source.exactStream=structuredClone(stream);validateGaugeMeasurementQuestion(selected);
 const family=result.gauges;if(!object(family)||family.fromNS!==bounds.fromNS||family.toNS!==bounds.toNS||!Array.isArray(family.series)||family.series.length>10000)fail();
 const matching=family.series.filter(s=>object(s)&&sameNativeStream(s.stream,stream));if(matching.length!==1||!['observed','stale','unknown'].includes(matching[0].state)||!Array.isArray(matching[0].samples)||matching[0].samples.length>10000)fail();
 const samples=matching[0].samples.filter(p=>object(p)&&p.atNS===atNS);if(samples.length!==1)fail();const sample=samples[0];
 if(!['observed','stale','unknown'].includes(sample.state)||sample.state==='unknown'&&sample.value!=null||sample.state!=='unknown'&&(typeof sample.value!=='number'||!Number.isFinite(sample.value)))fail();
 const r=stream.resource;if(!object(r)||!text(r['service.instance.id'],1024)||!r['service.instance.id']||own(r,'service.namespace')&&!text(r['service.namespace'],1024))fail();
 const radius=BigInt(radiusNS),from=at-radius>lo?at-radius:lo,to=at+radius<hi?at+radius:hi;if(to<=from)fail();const fromNS=String(from),toNS=String(to),leaf=(kind,key,op,value)=>({field:{kind,key},op,...(value!==undefined&&value!==''?{value}:{})});
 return {question:{signal,...Environment.questionFields(d),fromMs:Number(from/1000000n),toMs:Number((to+999999n)/1000000n),where:{all:[leaf('column','service','eq',r['service.name']),own(r,'service.namespace')?leaf('resource','service.namespace','eq',r['service.namespace']):leaf('resource','service.namespace','missing'),leaf('resource','service.instance.id','eq',r['service.instance.id']),leaf('column','timestamp_ns','gte',fromNS),leaf('column','timestamp_ns','lte',String(to-1n))]},limit:100,offset:0},fromNS,toNS,atNS,radiusNS,grain:'observation-context',relation:'observation-context'};
}
