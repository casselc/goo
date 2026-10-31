const {test}=require('node:test'),assert=require('node:assert/strict');
const {fromExploration}=require('../ui/record-page.mjs');
const {adoptExploration,adoptRecordPage,adoptStoredExplore,getRecordPage}=require('../ui/source-snapshots.mjs');
const {isSnapshot,hasGuarantee}=require('../ui/snapshots.mjs');
const query=()=>({signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:2,offset:4,groupBy:[{kind:'column',key:'service'}]});
const response=()=>({question:query(),basis:'a'.repeat(64),capturedAt:'2026-10-31T05:00:00.123456789Z',state:'observed',count:'500',hasMore:true,resultToken:'owner-only-token',rows:[{timestamp_ns:'1793422800123456789',SpanId:'0000000000000001',Attributes:{empty:'',missing:null,nested:['value',{number:1}]},ResourceAttributes:{'service.namespace':'team'}}]});

test('record pages preserve rich supplied data and ns clocks without borrowing aggregate completeness',()=>{
 const raw=response(),before=structuredClone(raw),page=fromExploration(raw);
 assert.deepEqual(page.rows,raw.rows);assert.equal(page.rows[0].timestamp_ns,'1793422800123456789');assert.equal(page.capturedAt,raw.capturedAt);
 assert.deepEqual(page.page,{offset:4,limit:2,returned:1,hasMore:true});assert.deepEqual(page.quality,{extent:'bounded-page',upstream:'unknown',consistency:'independent-reads',state:'observed'});
 assert(!isSnapshot(page));assert(!hasGuarantee(page,'groups'));assert(!Object.hasOwn(page,'count'));assert(!Object.hasOwn(page,'completeness'));assert(!Object.hasOwn(page,'resultToken'));
 raw.rows[0].Attributes.nested[1].number=2;raw.question.environment='other';assert.equal(page.rows[0].Attributes.nested[1].number,1);assert.equal(page.question.environment,'prod');
 assert(Object.isFrozen(page.rows[0].Attributes.nested[1]));assert(Object.isFrozen(page.question.groupBy[0]));assert(Object.isFrozen(page.page));assert.deepEqual(before.rows[0].Attributes.nested[1],{number:1});
});

test('known source admission binds the immutable page to its exact response wrapper',()=>{
 const raw=response(),admitted=adoptExploration(raw,raw.question),page=getRecordPage(admitted);
 assert(page);assert.equal(admitted.rows,page.rows);assert.equal(admitted.question,page.question);assert.equal(admitted.resultToken,'owner-only-token');assert.equal(getRecordPage({...admitted}),null);
 assert.equal(getRecordPage({recordPage:page}),null);assert.equal(getRecordPage({...raw,recordPage:fromExploration(raw)}),null);
 const other=response();other.question.environment='elsewhere';const admittedOther=adoptExploration(other,other.question);admittedOther.recordPage=page;assert.equal(getRecordPage(admittedOther),null);
 admitted.basis='b'.repeat(64);assert.equal(getRecordPage(admitted),null);
});

test('missing rows remain unavailable and omitted continuation stays unknown',()=>{
 const raw=response();delete raw.rows;assert.equal(getRecordPage(adoptExploration(raw,raw.question)),null);
 const empty=response();empty.rows=[];empty.state='no_data';delete empty.hasMore;const page=fromExploration(empty);assert.equal(page.page.returned,0);assert.equal(page.page.hasMore,null);assert.equal(page.quality.extent,'bounded-page');
 const changed=response();changed.state='changed_between_reads';assert.equal(fromExploration(changed).quality.state,'changed_between_reads');
});

test('stored delivery pages retain their actual bounds and source clock rather than the receipt clock',()=>{
 const raw=response(),original={...raw.question,offset:0};
 const stored=adoptStoredExplore({receipt:{id:'b'.repeat(64),capturedAt:'2026-10-31T06:00:00Z',question:{kind:'explore',explore:original}},explore:raw,retainedRecords:20});
 const page=getRecordPage(stored.explore);assert(page);assert.equal(page.question.offset,4);assert.equal(page.page.offset,4);assert.equal(page.page.limit,2);assert.equal(page.capturedAt,raw.capturedAt);assert.notEqual(page.capturedAt,stored.receipt.capturedAt);assert.equal(page.quality.upstream,'unknown');
});

test('record page validation refuses malformed bounds, source metadata and non-JSON mutable payloads',()=>{
 for(const mutate of [x=>x.rows.push({},{}),x=>x.rows=[null],x=>x.rows=[new Date()],x=>x.rows=[{value:new Map()}],x=>x.rows=[{value:NaN}],x=>x.rows=[{value:undefined}],x=>x.rows=[{value:1n}],x=>x.rows=[{value:'\ud800'}],x=>x.rows=[{get value(){throw Error('getter executed')}}],x=>{const row={};row.loop=row;x.rows=[row]},x=>{x.rows=[];x.rows.length=1},x=>x.hasMore='yes',x=>x.basis='forged',x=>x.capturedAt='yesterday',x=>x.capturedAt='2026-02-30T00:00:00Z',x=>x.state='complete',x=>x.question.limit=0]){const raw=response();mutate(raw);assert.throws(()=>fromExploration(raw));}
 const raw=response();Object.defineProperty(raw,'rows',{get(){throw Error('top getter executed')}});assert.throws(()=>fromExploration(raw),/data properties/);
});


test('count-and-row source admission keeps page retention without aggregate authority',()=>{
 const raw={...response(),resultToken:'c'.repeat(64),warnings:['Independent count and row reads.']},before=structuredClone(raw),admitted=adoptRecordPage(raw,raw.question),page=getRecordPage(admitted);
 assert(page);assert.equal(admitted.count,'500');assert.equal(admitted.resultToken,raw.resultToken);assert.equal(page.question.offset,4);assert.equal(page.page.returned,1);
 assert(!isSnapshot(page));assert(!hasGuarantee(page,'summary'));assert(!hasGuarantee(page,'groups'));assert(!Object.hasOwn(page,'count'));assert(!Object.hasOwn(page,'resultToken'));
 assert(!Object.hasOwn(admitted,'summaryDataset'));assert(!Object.hasOwn(admitted,'viewDataset'));assert.equal(getRecordPage({...admitted}),null);
 raw.rows[0].Attributes.nested[1].number=99;raw.question.offset=6;raw.warnings.push('new');assert.deepEqual(admitted.rows,before.rows);assert.equal(admitted.question.offset,4);assert.deepEqual(admitted.warnings,before.warnings);
 const unavailable={...before};delete unavailable.resultToken;assert(getRecordPage(adoptRecordPage(unavailable,unavailable.question)));
});

test('page-only admission rejects scope drift, full-result slots and invented count/token metadata',()=>{
 const valid=()=>({...response(),resultToken:'d'.repeat(64)});
 for(const mutate of [x=>x.question.offset=0,x=>x.question.environment='other',x=>x.question.toMs++,x=>x.count='1.5',x=>x.count='-1',x=>x.count='01',x=>x.count='18446744073709551616',x=>x.count=500,x=>x.resultToken='foreign-token',x=>x.warnings=[{}],x=>delete x.hasMore,x=>x.rows=null]){const raw=valid(),q=structuredClone(raw.question);mutate(raw);assert.throws(()=>adoptRecordPage(raw,q))}
 for(const slot of ['summaryDataset','viewDataset','timeDataset','histogramDataset','groups','fields','duration','numeric','timeSeries','plan','completeness']){const raw=valid();raw[slot]=null;assert.throws(()=>adoptRecordPage(raw,raw.question),/unsupported/)}
});

test('an empty late record page does not borrow held whole-population state or totals',()=>{
 const held=response(),late={...response(),basis:'b'.repeat(64),capturedAt:'2026-10-31T06:00:00Z',resultToken:'e'.repeat(64),count:'0',rows:[],hasMore:false,state:'no_data'};
 const admitted=adoptRecordPage(late,late.question),page=getRecordPage(admitted);
 assert.equal(page.page.returned,0);assert.equal(page.quality.extent,'bounded-page');assert.equal(page.quality.state,'no_data');assert.equal(admitted.count,'0');assert.equal(held.count,'500');assert.equal(held.state,'observed');assert.notEqual(page.basis,held.basis);assert(!hasGuarantee(page,'summary'));
 const changed={...late,rows:held.rows,state:'changed_between_reads'};assert.equal(getRecordPage(adoptRecordPage(changed,changed.question)).quality.state,'changed_between_reads');
});
