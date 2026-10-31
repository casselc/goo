const {test}=require('node:test'),assert=require('node:assert/strict'),D=require('../ui/view-declaration.mjs'),{definition}=require('../ui/population-workspace.mjs');
const q={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}]};
test('reusable declarations detach configuration, pin kinds and reference the owning checked query',()=>{
 const d=D.create(q),v=definition('Services',[q,{...q,signal:'logs'}],'records',['utc'],undefined,undefined,{spans:d});
 assert.equal(v.displays.spans.kind.version,1);assert.equal(v.displays.spans.scope.signal,'spans');assert.equal(v.displays.spans.display.value,'count');
 d.kind.name='table';assert.equal(v.displays.spans.kind.name,'bar');assert.equal(v.displays.spans.snapshot,undefined);
 const next=D.update(v.displays.spans,{kind:'table',value:'count',scale:'max',order:'label',query:'api',requiresComplete:true},q);
 assert.equal(D.presentation(D.create(q)).title,'');assert.equal(D.presentation({...D.create(q),display:{category:'category',value:'count'}}).requiresComplete,false);assert.equal(next.kind.name,'table');assert.equal(next.display.query,'api');assert.equal(D.presentation(next).requiresComplete,true);assert.deepEqual(D.validate(JSON.parse(JSON.stringify(next)),q),next);
});
test('unsupported kinds, fields, owners, source payloads and group profiles fail explicitly',()=>{
 for(const mutate of [d=>d.kind.version=2,d=>d.kind.name='line',d=>d.scope.signal='logs',d=>d.display.value='p95',d=>d.display.category='service',d=>d.display.scale='value',d=>d.snapshot={},d=>d.display.query='x'.repeat(257),d=>d.display.requiresComplete='yes']){const d=D.create(q);mutate(d);assert.throws(()=>D.validate(d,q));}
 for(const changed of [{...q,groupBy:[]},{...q,groupTimeSeries:true},{...q,groupBy:[q.groupBy[0],{kind:'column',key:'status'}]},{...q,groupBy:[{kind:'column',key:'duration_bucket'}]}])assert.throws(()=>D.validate(D.create(q),changed));
 assert.throws(()=>definition('Overlapping',[q],'records',['utc'],{spans:{metric:'count'}},undefined,{spans:D.create(q)}));
 assert.throws(()=>definition('Wrong owner',[{...q,signal:'logs'}],'records',['utc'],undefined,undefined,{spans:D.create(q)}));
});

test('Find drafts reject controls and byte overflow before replacing applied intent',()=>{const applied=D.create(q),before=structuredClone(applied);for(const query of ['api\u0001','api\t','😀'.repeat(65)])assert.throws(()=>D.withQuery(applied,query,q));assert.deepEqual(applied,before);const next=D.withQuery(applied,'api',q);assert.equal(next.display.query,'api');assert.deepEqual(applied,before)});

const binding={field:{kind:'attribute',key:'amount'},type:'int64',unit:'USD',parse:'canonical-decimal'};
test('query-supported numeric and duration measures preserve declared units, intent and local display settings',()=>{
 const supported={...q,numeric:binding,summary:'span-duration'},expected=['count','numeric.mean','numeric.sum','numeric.min','numeric.max','duration.mean','duration.min','duration.max','duration.p50','duration.p95','duration.p99'];
 assert.deepEqual(D.availableMeasures(supported),expected);assert.deepEqual(D.availableMeasures(q),['count']);assert.deepEqual(D.availableMeasures({...q,spanStatus:true}),['count']);assert.deepEqual(D.availableMeasures({...q,signal:'logs',numeric:binding}),expected.slice(0,5));
 for(const value of expected){const d=D.create(supported);d.display.value=value;d.display.scale=value==='count'?'population':'max';const before=structuredClone(d),accepted=D.validate(d,supported),props=D.presentation(accepted);assert.equal(props.value,value);const next=D.update(accepted,{...props,query:'worker',order:'label',kind:'table'},supported);assert.equal(next.display.value,value);assert.equal(next.display.query,'worker');assert.deepEqual(d,before);assert.equal(next.kind.name,'table');assert.equal(supported.numeric.unit,'USD');}
 const numeric=D.create(supported);numeric.display.value='numeric.mean';assert.throws(()=>D.validate(numeric,supported),/Population share/);numeric.display.scale='max';assert.throws(()=>D.validate(numeric,q),/supported/);assert.throws(()=>D.validate(numeric,{...supported,numeric:undefined}),/supported/);
 const duration=D.create(supported);duration.display.value='duration.p95';delete duration.display.scale;assert.equal(D.presentation(D.validate(duration,supported)).scale,'max');assert.throws(()=>D.validate(duration,{...supported,summary:undefined}),/supported/);
 for(const value of ['status.error','status.error_fraction','numeric.p95','duration.sum','amount']){const d=D.create(supported);d.display.value=value;d.display.scale='max';assert.throws(()=>D.validate(d,supported));}
});
