const {test}=require('node:test'),assert=require('node:assert/strict');
const {plan}=require('../ui/native-record-context');
const fixtures=require('./fixtures/native-record-context.json');
for(const c of fixtures)test('native record context shared Go fixture: '+c.name,()=>{
 const input=structuredClone(c),run=()=>plan(c.parent,c.stream,c.signal,c.fromNS,c.toNS);
 if(c.refused)assert.throws(run);else assert.deepEqual(run(),c.expected);
 assert.deepEqual(c,input,'pure mapper must not mutate caller data');
});
const valid=fixtures.find(c=>!c.refused);
for(const [name,change]of Object.entries({
 'wrong metric':c=>c.stream.name='other', 'wrong scope':c=>c.stream.scope='other',
 'wrong schema':c=>c.stream.schemaURL='other','wrong environment':c=>c.stream.resource['deployment.environment.name']='other',
 'missing instance':c=>delete c.stream.resource['service.instance.id'],'wrong service':c=>c.stream.resource['service.name']='other',
 'invalid parent version':c=>c.parent.definition.apiVersion='other','invalid parent kind':c=>c.parent.definition.kind='lifecycle',
 'duplicate services':c=>c.parent.definition.source.services=['api','api'], 'bad lookback':c=>c.parent.definition.lookbackMinutes=121,
 'noncanonical bounds':c=>c.fromNS='0'+c.fromNS,'unsafe ms':c=>c.parent.toMs=Number.MAX_SAFE_INTEGER+1,
 'outside parent':c=>c.fromNS=String(BigInt(c.parent.fromNS)-1n),'empty interval':c=>c.toNS=c.fromNS,
 'wrong native selector':c=>c.parent.definition.source.attributes={'metric.only':'other'},
 'pinned attr drift':c=>{c.parent.definition.source.exactStream=structuredClone(c.stream);c.stream.attributes['metric.only']='changed'},
 'pinned resource schema drift':c=>{c.parent.definition.source.exactStream=structuredClone(c.stream);c.stream.resourceSchemaURL='changed'},
 'namespace missing differs empty':c=>{c.parent.definition.source.namespace={mode:'missing'};c.stream.resource['service.namespace']=''},
 'namespace value differs missing':c=>{c.parent.definition.source.namespace={mode:'value',value:''};delete c.stream.resource['service.namespace']},
 'oversized identity':c=>c.stream.attributes.large='x'.repeat(2049),
}))test('native record context rejects '+name,()=>{const c=structuredClone(valid);change(c);assert.throws(()=>plan(c.parent,c.stream,c.signal,c.fromNS,c.toNS))});
test('context projects only shared identity, never metric dimensions or scope',()=>{
 const c=structuredClone(valid),result=plan(c.parent,c.stream,c.signal,c.fromNS,c.toNS),all=result.question.where.all;
 assert.deepEqual(all.map(p=>p.field.key),['service','service.namespace','service.instance.id','timestamp_ns','timestamp_ns']);
 assert.equal(all.at(-1).value,String(BigInt(c.toNS)-1n));
});
