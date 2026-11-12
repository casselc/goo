/* Pure fixed-scope service triage. No acquisition, DOM, timer or publication owner. */
import * as Environment from './record-environment.mjs';
import * as ServiceIdentity from './service-identity.mjs';
import * as Aggregate from './record-aggregate-protocol.mjs';
import * as Values from './distribution-values.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {identity} from './question-session.mjs';
const clone=value=>structuredClone(value),own=(value,key)=>Object.hasOwn(value,key);
const string=value=>typeof value==='string'&&value.isWellFormed()&&!value.includes('\0');
function namespace(value){
 if(value===undefined||value===null)return {mode:'missing'};
 if(typeof value!=='object'||Array.isArray(value)||Object.keys(value).some(k=>!['mode','value'].includes(k))||!['missing','value'].includes(value.mode)||own(value,'value')&&!string(value.value)||value.mode==='missing'&&own(value,'value')&&value.value!=='')throw Error('Invalid service namespace.');
 return value.mode==='missing'?{mode:'missing'}:{mode:'value',...(value.value?{value:value.value}:{})};
}
export function plan(scope){
 if(!scope||typeof scope!=='object'||Array.isArray(scope)||Object.keys(scope).some(k=>!['environment','environmentSelector','namespace','fromMs','toMs'].includes(k)))throw Error('Use the accepted fixed catalog scope.');
 if(!Number.isSafeInteger(scope.fromMs)||!Number.isSafeInteger(scope.toMs)||scope.fromMs<0||scope.toMs<=scope.fromMs||scope.toMs-scope.fromMs>7200000||scope.toMs>4102444800000)throw Error('Use a fixed catalog window of at most 120 minutes before 2100.');
 const where=[{field:{kind:'column',key:'kind'},op:'eq',value:'Server'}];
 if(scope.namespace!=null){const ns=namespace(scope.namespace);if(ns.mode==='value'&&new TextEncoder().encode(ns.value||'').length>128)throw Error('Choose an exact namespace up to 128 UTF8 bytes.');where.push(ServiceIdentity.predicate(ns));}
 const current={signal:'spans',...Environment.questionFields(scope),fromMs:scope.fromMs,toMs:scope.toMs,limit:100,offset:0,groupBy:[{kind:'column',key:'service'},{kind:'resource',key:'service.namespace'}],summary:'span-duration',spanStatus:true,where:{all:where}};
 validateRecordQuestion(current);
 const width=current.toMs-current.fromMs;
 if(current.fromMs<width)throw Error('The previous window would begin before the Unix epoch.');
 const baseline={...clone(current),fromMs:current.fromMs-width,toMs:current.fromMs};
 validateRecordQuestion(baseline);
 return {scope:clone(scope),current,baseline};
}
function checkedPlan(value){const expected=plan(value?.scope);if(identity(value?.current)!==identity(expected.current)||identity(value?.baseline)!==identity(expected.baseline))throw Error('Triage questions differ from the fixed catalog scope.');return expected;}
function snapshot(query,value,scope){
 Aggregate.recordAggregates(query,value);
 if(!['observed','no_data'].includes(value.state))throw Error('Independent reads changed; service triage is unavailable.');
 for(const group of value.groups){if(!string(group.g0))throw Error('Invalid service identity.');if(scope.namespace!=null&&!ServiceIdentity.same(namespace(scope.namespace),ServiceIdentity.namespace(group.g1)))throw Error('Returned service namespace differs from the scope.');}
 return clone(value);
}
export function admit(value,current,baseline){const p=checkedPlan(value);return {plan:p,current:snapshot(p.current,current,p.scope),baseline:snapshot(p.baseline,baseline,p.scope)};}
const fractionDifference=(a,b,numerator)=>!a||!b?null:{n:(BigInt(numerator(a))*BigInt(b.count)-BigInt(numerator(b))*BigInt(a.count)).toString(),d:(BigInt(a.count)*BigInt(b.count)).toString()};
export function rows(value,catalog=[]){
 const accepted=admit(value?.plan,value?.current,value?.baseline),items=new Map();
 if(!Array.isArray(catalog))throw Error('Invalid service catalog.');
 const add=(service,ns,qualified=true)=>{if(!string(service))throw Error('Invalid service identity.');ns=qualified?namespace(ns):null;const key=ServiceIdentity.key(service,ns);if(!items.has(key))items.set(key,{key,service,namespace:ns,qualified,current:null,baseline:null});return items.get(key);};
 for(const row of catalog)add(row.service,row.namespace,row.namespace!=null);
 for(const side of ['current','baseline'])for(const group of accepted[side].groups)add(group.g0,ServiceIdentity.namespace(group.g1))[side]=clone(group);
 return [...items.values()].map(row=>{const a=row.current,b=row.baseline;return {...row,countDelta:a||b?(BigInt(a?.count||'0')-BigInt(b?.count||'0')).toString():null,errorDelta:fractionDifference(a,b,g=>g.spanStatus.error),durationDelta:fractionDifference(a,b,g=>g.duration.sumNs),p95Delta:a&&b?(BigInt(a.duration.p95Ns)-BigInt(b.duration.p95Ns)).toString():null};});
}
export function sortRows(items,metric='count'){
 if(!['name','count','error-rate','mean','p95','count-change'].includes(metric))throw Error('Unsupported service triage ranking.');
 const magnitude=row=>metric==='count-change'?(row.countDelta===null?null:{n:BigInt(row.countDelta),d:1n}):row.current?Values.groupMetric(row.current,metric):null;
 return [...items].sort((a,b)=>{if(metric==='name'){const x=ServiceIdentity.name(a.service,a.namespace),y=ServiceIdentity.name(b.service,b.namespace);return x<y?-1:x>y?1:a.key<b.key?-1:a.key>b.key?1:0}const x=magnitude(a),y=magnitude(b);if(!x||!y)return x?-1:y?1:a.key<b.key?-1:a.key>b.key?1:0;const delta=x.n*y.d-y.n*x.d;return delta>0n?-1:delta<0n?1:a.key<b.key?-1:a.key>b.key?1:0;});
}
export function question(value,row){
 const p=checkedPlan(value);if(!row||row.qualified===false||row.namespace==null||!string(row.service))throw Error('Choose an observed service.');const ns=namespace(row.namespace),q=clone(p.current),service=ServiceIdentity.serviceWhere(row.service,ns);q.where=ServiceIdentity.and(q.where,...(service.all||[service]));validateRecordQuestion(q);return q;
}
