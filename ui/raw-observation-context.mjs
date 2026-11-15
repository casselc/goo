/* A held source point may stage ordinary nearby records, never an evaluation. */
import {checkedResult} from './raw-observation-protocol.mjs';
import * as Environment from './record-environment.mjs';
import {object,text} from './intent-shape.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';

export const radiusNS='60000000000';
const has=(value,key)=>Object.hasOwn(value,key);
const fail=()=>{throw Error('Choose one held raw observation in the current fixed read.');};
const leaf=(kind,key,op,value)=>({field:{kind,key},op,...(value===undefined?{}:{value})});

// Membership is the read-local occurrence key, not value/time/stream equality:
// identical exports are separate occurrences and need not imply producer identity.
export function plan(data,occurrenceKey,signal){
 if(!['logs','spans'].includes(signal)||!text(occurrenceKey,128,{empty:false}))fail();
 object(data,['version','question','instrument','readAt','rows','limit']);
 const admitted=checkedResult(data,{question:data.question,instrument:data.instrument});
 const selected=admitted.rows.filter(row=>row.occurrenceKey===occurrenceKey);
 if(selected.length!==1)fail();
 const row=selected[0],q=admitted.question,resource=row.stream.resource;
 const at=BigInt(row.endNS),lo=BigInt(q.fromMs)*1000000n,hi=BigInt(q.toMs)*1000000n,radius=BigInt(radiusNS);
 const from=at-radius>lo?at-radius:lo,to=at+radius<hi?at+radius:hi;
 if(to<=from)fail();
 const predicates=[leaf('column','service','eq',q.service),has(resource,'service.namespace')?leaf('resource','service.namespace','eq',resource['service.namespace']):leaf('resource','service.namespace','missing')];
 const scope=has(resource,'service.instance.id')?'reported-instance':'service';
 if(scope==='reported-instance')predicates.push(leaf('resource','service.instance.id','eq',resource['service.instance.id']));
 predicates.push(leaf('column','timestamp_ns','gte',String(from)),leaf('column','timestamp_ns','lte',String(to-1n)));
 const question={signal,...Environment.questionFields(q),fromMs:Number(from/1000000n),toMs:Number((to+999999n)/1000000n),where:{all:predicates},limit:100,offset:0};
 validateRecordQuestion(question);
 return {question,fromNS:String(from),toNS:String(to),atNS:row.endNS,radiusNS,grain:'raw-observation-context',relation:'raw-observation-context',scope,origin:{version:admitted.version,readAt:admitted.readAt,question:structuredClone(q),instrument:structuredClone(admitted.instrument),occurrenceKey:row.occurrenceKey,stream:structuredClone(row.stream),endNS:row.endNS}};
}
