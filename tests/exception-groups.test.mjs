import {test} from 'node:test';
import assert from 'node:assert/strict';
import {question,stageQuestion,admit} from '../ui/exception-groups.mjs';
import {decode} from '../ui/exception-fingerprint-identity.mjs';
import {validate} from '../ui/predicate-validation.mjs';
import {questionLink,readQuestionLink} from '../ui/question-protocol.mjs';
const id='ef1:c:'+'a'.repeat(64),q={signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:'prod'}},fromMs:1000,toMs:61000,limit:100,offset:0,where:{any:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'attribute',key:'team'},op:'eq',value:'payments'}]}};
const receipt={readAt:'2026-11-13T00:00:00Z',basis:'a'.repeat(64)};
function result(){return{version:'goo/exception-groups/v1',question:structuredClone(q),count:'151',state:'observed',capturedAt:receipt.readAt,basis:'b'.repeat(64),reads:{count:{...receipt},groups:{...receipt}},warnings:[],groups:[{fingerprint:id,state:'classified',reason:'',count:'150',firstNS:'1000000000',lastNS:'60999999999',inputs:{type:'CheckoutError',stacktrace:'checkout\n  reserve:42'},affected:[{service:'api',namespace:null,version:''},{service:'api',namespace:'',version:null}],affectedCount:'2',affectedLimited:false},{fingerprint:'ef1:u:missing-stack',state:'unclassified',reason:'missing-stack',count:'1',firstNS:'5000000000',lastNS:'5000000000',inputs:null,affected:[{service:'api',namespace:null,version:null}],affectedCount:'1',affectedLimited:false}]};}

test('exception question/stage preserves full selection and fixed clocks without measurement directives',()=>{
 const parent={...structuredClone(q),groupBy:[{kind:'column',key:'service'}],bucketMs:1000},before=structuredClone(parent),request=question(parent),staged=stageQuestion(parent,id);
 assert.deepEqual(request,q);assert.deepEqual(staged.where,{all:[q.where,{field:{kind:'column',key:'exception_fingerprint_v1'},op:'eq',value:id}]});assert.deepEqual(parent,before);assert.equal(staged.offset,0);assert.equal(staged.limit,100);
 for(const change of [{offset:100},{reportedEdge:{}},{numeric:{}}])assert.throws(()=>question({...parent,...change}));
 for(const signal of ['logs','spans']){const record={...staged,signal};validate(record.where,signal);assert.deepEqual(readQuestionLink(questionLink(record)).question,record);}
 assert.throws(()=>stageQuestion(q,'ef1:c:'+'A'.repeat(64)));assert.throws(()=>validate({field:{kind:'column',key:'exception_fingerprint_v1'},op:'contains',value:id},'logs'));
});

test('exception admission keeps full totals, exact bounds, literal inputs and qualified missing/empty contexts',()=>{
 const supplied=result(),before=structuredClone(supplied),accepted=admit(q,supplied);assert.deepEqual(accepted,supplied);assert.notEqual(accepted,supplied);accepted.groups[0].inputs.stacktrace='changed';assert.deepEqual(supplied,before);
 const large=result();large.count='9007199254740994';large.groups[0].count='9007199254740993';assert.equal(admit(q,large).count,large.count);
 const changed=result();changed.count='152';changed.state='changed_between_reads';assert.equal(admit(q,changed).state,'changed_between_reads');
 const bounded=result();bounded.groups[0].affected=Array.from({length:32},(_,i)=>({service:'service-'+i,namespace:null,version:''}));bounded.groups[0].affectedCount='33';bounded.groups[0].affectedLimited=true;assert.equal(admit(q,bounded).groups[0].affectedLimited,true);
 assert.notEqual(decode('ef1:u:both-empty').label,decode('ef1:u:both-missing').label);
});

test('exception admission refuses truncated, substituted, malformed and incoherent population facts',()=>{
 const mutants=[x=>x.count='152',x=>x.groups.push(structuredClone(x.groups[0])),x=>x.groups[0].firstNS='999999999',x=>x.groups[0].lastNS='61000000000',x=>x.groups[0].inputs.type='',x=>x.groups[0].inputs.stacktrace='a'.repeat(8193),x=>x.groups[1].inputs={type:'made-up',stacktrace:'sample'},x=>x.groups[0].affectedLimited=true,x=>x.groups[0].affectedCount='3',x=>x.groups[0].affected[1]=structuredClone(x.groups[0].affected[0]),x=>x.reads.groups.basis='x',x=>x.question.toMs++,x=>x.groups[0].rows=[],x=>x.groups[0].count='01',x=>x.groups[1].fingerprint='ef1:u:invented',x=>x.groups[0].inputs.type='\ud800'];
 for(const mutate of mutants){const value=result();mutate(value);assert.throws(()=>admit(q,value));}
});
