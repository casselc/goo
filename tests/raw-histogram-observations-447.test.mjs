import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {checkedRequest,checkedResult,mount} from '../ui/raw-observations.mjs';
import {facts,intervals,summary} from '../ui/raw-histogram-observations.mjs';
import {Element} from './support/raw-observation-dom.mjs';
const presence=()=>['sum','min','max'].map(field=>({field,code:'presence-unavailable'}));
function request(){return {question:{environment:'local',service:'worker',fromMs:1000,toMs:3000},instrument:{kind:'histogram',scope:'',version:'',schemaURL:'',metric:'work.latency',unit:'widgets',temporality:1}}}
function response(q=request()){
 const stream={resourceSchemaURL:'',resource:{'service.name':'worker','deployment.environment.name':'local'},attributes:{queue:'jobs'},scope:'',version:'',schemaURL:'',name:q.instrument.metric,unit:q.instrument.unit};
 const rawRow={resource:structuredClone(stream.resource),attributes:structuredClone(stream.attributes),resource_schema_url:'',scope:'',version:'',schema_url:'',name:q.instrument.metric,unit:q.instrument.unit,start_ns:'1000000000',end_ns:'2000000000',temporality:q.instrument.temporality,flags:0,count:'7',bounds:['0','1'],bucket_counts:['1','2','4'],sum:'4.25',min:'0',max:'0'};
 const statistic=(raw,value)=>({rawValue:raw,value,valueState:'finite',presenceState:'unavailable'});
 return {version:'goo/raw-histogram-observations/v1',question:structuredClone(q.question),instrument:structuredClone(q.instrument),readAt:'2026-11-15T12:00:00.123456789Z',limit:10000,rows:[{occurrenceKey:'1',stream,startNS:'1000000000',startState:'valid',endNS:'2000000000',flags:0,flagsState:'valid',temporality:q.instrument.temporality,recordingState:'recorded',count:'7',countState:'valid',bounds:[0,1],boundsState:'valid',bucketCounts:['1','2','4'],bucketCountsState:'valid',bucketsState:'valid',sum:statistic('4.25',4.25),min:statistic('0',0),max:statistic('0',0),diagnostics:[...(q.instrument.temporality===0?[{field:'temporality',code:'unspecified'}]:![1,2].includes(q.instrument.temporality)?[{field:'temporality',code:'unrecognized'}]:[]),...presence()],rawRow}]};
}
test('447 explicit histogram request preserves exact identity, unit and unknown temporality without evaluator readiness',()=>{
 for(const tempo of [0,1,2,-7,2147483647]){const q=request();q.instrument.temporality=tempo;assert.deepEqual(checkedRequest(q),q);assert.deepEqual(checkedResult(response(q),q),response(q))}
 for(const change of [q=>q.instrument.kind='exponential-histogram',q=>q.instrument.monotonic=false,q=>delete q.instrument.temporality,q=>q.instrument.temporality=null,q=>q.instrument.temporality=1.5,q=>q.instrument.temporality=2147483648,q=>q.instrument.scope=undefined,q=>{delete q.question.environment;q.question.environmentSelector={version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'all'}}}]){const q=request();change(q);assert.throws(()=>checkedRequest(q))}
});
test('447 raw histogram result preserves duplicates, exact counts and source statistic presence uncertainty',()=>{
 const q=request(),r=response();r.rows.push({...structuredClone(r.rows[0]),occurrenceKey:'2'});const out=checkedResult(r,q);assert.equal(out.rows.length,2);out.rows[0].bucketCounts[0]='changed';assert.equal(r.rows[0].bucketCounts[0],'1');
 const large=response(),row=large.rows[0];row.rawRow.count=row.count='18446744073709551615';row.rawRow.bucket_counts=row.bucketCounts=['18446744073709551615','0','0'];assert.equal(checkedResult(large,q).rows[0].count,'18446744073709551615');assert.match(summary(row),/≈/);assert.equal(row.min.presenceState,'unavailable');
 assert.deepEqual(intervals(r.rows[0]),[{interval:'(−∞, 0]',count:'1'},{interval:'(0, 1]',count:'2'},{interval:'(1, +∞)',count:'4'}]);
});
test('447 count-only bucket absence and an all-values bucket remain different legal observations',()=>{
 const q=request(),r=response(),row=r.rows[0];row.rawRow.bounds=row.bounds=[];row.rawRow.bucket_counts=row.bucketCounts=[];row.bucketsState='absent';row.diagnostics=[{field:'buckets',code:'absent'},...presence()];assert.equal(checkedResult(r,q).rows[0].count,'7');assert.match(summary(row),/Buckets absent/);assert.deepEqual(intervals(row),[]);
 row.rawRow.bucket_counts=row.bucketCounts=['7'];row.bucketsState='valid';row.diagnostics=presence();assert.equal(checkedResult(r,q).rows[0].bucketsState,'valid');assert.deepEqual(intervals(row),[{interval:'(−∞, +∞)',count:'7'}]);
});
test('447 invalid vectors stay diagnostic while forged facts and wrong source identity refuse whole reads',()=>{
 const q=request(),r=response(),row=r.rows[0];row.rawRow.count=row.count='8';row.bucketsState='invalid';row.diagnostics=[{field:'buckets',code:'count-mismatch'},...presence()];assert.equal(checkedResult(r,q).rows[0].bucketsState,'invalid');assert.match(summary(row),/Buckets invalid/);
 for(const change of [r=>r.rows[0].count='8',r=>r.rows[0].bucketCounts[0]='999',r=>r.rows[0].bounds=[1,2],r=>r.rows[0].recordingState='unknown',r=>r.rows[0].sum.presenceState='present',r=>r.rows[0].rawRow.temporality=2,r=>r.rows[0].stream.name='other',r=>r.rows[0].endNS='3000000000',r=>r.rows.push(structuredClone(r.rows[0])),r=>r.rows[0].value=0,r=>r.rows[0].diagnostics=[],r=>r.rows[0].rawRow.bounds=Array(10001).fill('1')]){const next=response();change(next);assert.throws(()=>checkedResult(next,q))}
 const mixed=response(),m=mixed.rows[0];m.rawRow.count=null;m.count=null;m.countState='absent';m.rawRow.bounds=['NaN'];m.bounds=null;m.boundsState='invalid';m.bucketsState='invalid';m.diagnostics=[{field:'count',code:'absent'},{field:'bounds',code:'invalid'},{field:'buckets',code:'invalid'},...presence()];assert.equal(checkedResult(mixed,q).rows[0].bucketsState,'invalid');
});
test('447 no-recorded flag suppresses normalized histogram values while retaining original source fields',()=>{
 const q=request(),r=response(),row=r.rows[0];row.flags=row.rawRow.flags=1;row.recordingState='no-recorded-value';row.count=row.bounds=row.bucketCounts=null;row.countState=row.boundsState=row.bucketCountsState=row.bucketsState='absent';for(const key of ['sum','min','max']){row[key].value=null;row[key].valueState='absent'}row.diagnostics=[{field:'count',code:'no-recorded-value'},{field:'buckets',code:'no-recorded-value'},...['sum','min','max'].flatMap(field=>[{field,code:'no-recorded-value'},{field,code:'presence-unavailable'}])];
 assert.equal(checkedResult(r,q).rows[0].rawRow.count,'7');assert.equal(summary(row),'Not recorded');assert.deepEqual(intervals(row),[]);row.sum.value=4.25;assert.throws(()=>checkedResult(r,q));
});
test('447 unknown or absent flags cannot become a recorded distribution, but structural evidence remains inspectable',()=>{
 const q=request(),r=response(),row=r.rows[0];row.flags=row.rawRow.flags=2;row.recordingState='unknown';row.diagnostics=[{field:'flags',code:'unknown-bits'},...presence()];assert.equal(checkedResult(r,q).rows[0].bucketsState,'valid');assert.match(summary(row),/recording unknown/);
 delete row.rawRow.flags;row.flags=null;row.flagsState='absent';row.diagnostics=[{field:'flags',code:'absent'},...presence()];assert.equal(checkedResult(r,q).rows[0].recordingState,'unknown');
});
test('447 exact bucket arithmetic and invalid precedence are independent of JavaScript numeric rounding',()=>{
 const raw=response().rows[0].rawRow;raw.count='9007199254740993';raw.bucket_counts=['9007199254740992','1','0'];assert.equal(facts(raw,{flags:0,flagsState:'valid'}).result.bucketsState,'valid');raw.bucket_counts=['9007199254740992','0','0'];assert.equal(facts(raw,{flags:0,flagsState:'valid'}).diagnostics[0].code,'count-mismatch');raw.count=null;raw.bounds=['Infinity'];assert.equal(facts(raw,{flags:0,flagsState:'valid'}).result.bucketsState,'invalid');
});

test('447 histogram inspection uses exact source bucket table and presence-qualified statistics without new reads',async()=>{
 globalThis.Node=Element;const body=new Element('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};const host=new Element('main'),opener=new Element('button');body.append(host,opener);const calls=[];let active=true;
 const owner=mount(host,{active:()=>active,post:async(path,input)=>{calls.push({path,input});return {ok:true,data:response(input)}}});
 await owner.read(request(),opener);assert.equal(calls[0].path,'/api/histogram-observations');assert.match(host.textContent,/Raw histogram observations/);assert.match(host.textContent,/7 count · 3 buckets/);assert.doesNotMatch(host.textContent,/7 widgets/);
 host.querySelectorAll('button').find(b=>b.textContent==='Inspect').onclick();assert.match(host.textContent,/Histogram observation 1/);const buckets=host.querySelectorAll('goo-table').find(t=>t.getAttribute('aria-label')==='Source histogram buckets');assert.ok(buckets);assert.equal(buckets.querySelector('table').getAttribute('aria-label'),'Source histogram buckets');assert.match(buckets.textContent,/\(−∞, 0\]/);assert.match(buckets.textContent,/\(1, \+∞\)/);assert.match(buckets.textContent,/Source count/);const stats=host.querySelectorAll('goo-table').find(t=>t.getAttribute('aria-label')==='Source histogram statistics');assert.ok(stats);assert.equal(stats.querySelector('table').getAttribute('aria-label'),'Source histogram statistics');assert.match(stats.textContent,/Not preserved by exporter/);assert.match(stats.textContent,/presence unavailable, including source zero/);assert.equal(calls.length,1);
 const held=host.textContent;active=false;owner.suspend();active=true;assert.equal(host.textContent,held);assert.equal(calls.length,1);
});
test('447 selected bucket paging survives held Find and navigation, with no source work',async()=>{
 globalThis.Node=Element;const body=new Element('body');body.root=true;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};const host=new Element('main'),opener=new Element('button');body.append(host,opener);const data=response(),row=data.rows[0];row.count=row.rawRow.count='151';row.bounds=Array.from({length:150},(_,i)=>i);row.rawRow.bounds=row.bounds.map(String);row.bucketCounts=Array(151).fill('1');row.rawRow.bucket_counts=[...row.bucketCounts];let reads=0,active=true;const owner=mount(host,{active:()=>active,post:async()=>{reads++;return {ok:true,data}}});await owner.read(request(),opener);host.querySelectorAll('button').find(b=>b.textContent==='Inspect').onclick();const more=host.querySelectorAll('button').find(b=>b.textContent==='More buckets');assert.ok(more);const find=host.querySelector('input');find.value='jobs';find.oninput();more.onclick();assert.match(host.textContent,/101–151 of 151 buckets/);assert.equal(more.disabled,true);assert.ok(host.querySelectorAll('span').some(e=>e.getAttribute('role')==='status'&&e.textContent==='101–151 of 151 buckets'));const previous=host.querySelectorAll('button').find(b=>b.textContent==='Previous buckets');active=false;owner.suspend();previous.onclick();assert.match(host.textContent,/101–151 of 151 buckets/);active=true;previous.onclick();assert.match(host.textContent,/1–100 of 151 buckets/);assert.equal(reads,1);
});

test('447 UI admits the independently generated Go HTTP success fixtures',()=>{
 const fixture=JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url),'utf8')),successes=fixture.cases.filter(c=>c.status===200);assert.ok(successes.length>0);for(const item of successes)assert.doesNotThrow(()=>checkedResult(JSON.parse(item.response),JSON.parse(item.body)),item.name);
});
