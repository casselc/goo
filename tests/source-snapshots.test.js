const {test}=require('node:test'),assert=require('node:assert/strict');
const {adoptExploration,adoptStoredExplore,importComparisonResult}=require('../ui/source-snapshots.mjs');
const {isSnapshot,hasGuarantee,getSnapshotInfo}=require('../ui/snapshots.mjs');
const {validateDisplay}=require('../ui/view-model.mjs');
const clone=value=>structuredClone(value);
const question=()=>({signal:'spans',environment:'test',fromMs:0,toMs:20000,bucketMs:10000,groupBy:[{kind:'column',key:'service'}],limit:100,offset:0});
function response(q=question()){
  const basis='a'.repeat(64),capturedAt='2026-10-30T05:00:00.123456789Z',scope={id:'scope-'+q.fromMs,generation:basis,signal:q.signal,selector:clone(q)};
  const base=(shape,aspect)=>({version:'goo.dataset/1',id:scope.id+':'+aspect,revision:basis,origin:'source-read',capturedAt,shape,scope:clone(scope),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:basis,scopeId:scope.id,aspects:[aspect]}});
  const count={key:'count',label:'Captured records',type:'number',unit:'records'};
  const viewDataset={...base('ranked','groups'),dimensions:['category'],fields:[{key:'category',label:'Service',type:'string'},count],rows:[{key:'api',values:{category:'api',count:'300'}}]};
  const timeDataset={...base('time-series','intervals'),fields:[{key:'fromMs',label:'From',type:'timestamp',unit:'ms'},{key:'toMs',label:'To',type:'timestamp',unit:'ms'},count],interval:{fromMs:String(q.fromMs),toMs:String(q.toMs),bucketMs:String(q.bucketMs)},rows:[{key:'first',state:'observed',values:{fromMs:String(q.fromMs),toMs:String(q.fromMs+10000),count:'150'}},{key:'last',state:'observed',values:{fromMs:String(q.fromMs+10000),toMs:String(q.toMs),count:'150'}}]};
  const summaryDataset={...base('summary','summary'),fields:[count],rows:[{key:'population',values:{count:'300'}}]};
  return {question:clone(q),basis,capturedAt,count:'300',state:'observed',rows:[],groups:[],viewDataset,timeDataset,summaryDataset};
}

test('only checked exploration slots gain source handles without mutating the wire response',()=>{
  const q=question(),data=response(q),before=JSON.stringify(data);data.unrelated=clone(data.viewDataset);
  const admitted=adoptExploration(data,q);assert.notEqual(admitted,data);
  for(const [slot,aspect] of [['viewDataset','groups'],['timeDataset','intervals'],['summaryDataset','summary']]){
    assert(isSnapshot(admitted[slot]));assert(hasGuarantee(admitted[slot],aspect));assert.equal(getSnapshotInfo(admitted[slot]).sourceId,'goo.explore');assert(Object.isFrozen(admitted[slot].rows[0].values));
    assert.equal(adoptExploration(admitted,q)[slot],admitted[slot]);
  }
  assert(!isSnapshot(admitted.unrelated));delete data.unrelated;assert.equal(JSON.stringify(data),before);
});

test('returned question, slot scope, shape, basis and clock cannot drift from the captured request',()=>{
  const q=question();
  for(const mutate of [data=>data.question.environment='other',data=>data.question.toMs++,data=>data.viewDataset.scope.selector.environment='other',data=>data.viewDataset.revision='foreign',data=>data.viewDataset.scope.generation='foreign',data=>data.summaryDataset.capturedAt='2026-10-30T06:00:00Z',data=>data.viewDataset=clone(data.summaryDataset)]){
    const data=response(q);mutate(data);assert.throws(()=>adoptExploration(data,q));
  }
  const requested={...q,environmentSelector:{version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'named',value:'test'}}};delete requested.environment;
  const data=response(requested);data.question.environmentSelector.policy='effective';assert.throws(()=>adoptExploration(data,requested),/scope/);
  const withoutSlots={question:{...q,environment:'other'}};assert.throws(()=>adoptExploration(withoutSlots,q),/scope/);
});

test('verified stored pages check named-environment receipts as well as selectors and ignore only paging',()=>{
  const q={...question(),limit:20},data=response(q),page={receipt:{id:'b'.repeat(64),capturedAt:'2026-10-30T05:01:00Z',question:{kind:'explore',explore:clone(q)}},explore:data,retainedRecords:300};
  page.explore.question={...q,limit:100,offset:100};
  const admitted=adoptStoredExplore(page);assert(hasGuarantee(admitted.explore.viewDataset,'groups'));assert.equal(getSnapshotInfo(admitted.explore.viewDataset).sourceId,'goo.verified-saved-explore');
  assert.equal(admitted.explore.summaryDataset.capturedAt,data.capturedAt);assert.notEqual(admitted.explore.summaryDataset.capturedAt,page.receipt.capturedAt);assert.equal(admitted.explore.viewDataset.scope.selector.limit,20);assert.equal(admitted.explore.question.offset,100);
  for(const mutate of [p=>p.explore.question.environment='other',p=>p.explore.question.toMs++,p=>p.explore.question.where={field:{kind:'column',key:'service'},op:'eq',value:'foreign'},p=>p.explore.viewDataset.scope.selector.environment='other',p=>p.receipt.question.kind='measurement']){const changed=clone(page);mutate(changed);assert.throws(()=>adoptStoredExplore(changed));}
});

test('comparison transport is conservative held data even when JSON claims source completeness',()=>{
  const a=question(),b={...a,fromMs:20000,toMs:40000},baseline=response(a).summaryDataset,selection=response(b).summaryDataset;
  const data={state:'ok',viewComparison:{version:'goo.comparison/1',state:'observed',baseline,selection,changes:{count:{state:'derived',value:'0',unit:'records'}}}},request={baseline:{question:a},selection:{question:b}};
  const admitted=importComparisonResult(data,request);assert.equal(admitted.viewComparison.changes.count.value,'0');
  for(const role of ['baseline','selection']){const snapshot=admitted.viewComparison[role];assert.equal(snapshot.origin,'browser-held');assert(!hasGuarantee(snapshot,'summary'));assert.equal(validateDisplay(snapshot,{kind:'table',requiresComplete:true}).ok,false);assert.equal(getSnapshotInfo(snapshot).entry,'held');}
  assert.equal(data.viewComparison.baseline.origin,'source-read');
  const branded={...data,viewComparison:{...data.viewComparison,baseline:adoptExploration(response(a),a).summaryDataset,selection:adoptExploration(response(b),b).summaryDataset}},reimported=importComparisonResult(branded,request);
  assert(!hasGuarantee(reimported.viewComparison.baseline,'summary'));assert(!hasGuarantee(reimported.viewComparison.selection,'summary'));assert.equal(getSnapshotInfo(reimported.viewComparison.baseline).entry,'held');
  const foreign=clone(data);foreign.viewComparison.selection.scope.selector.environment='other';assert.throws(()=>importComparisonResult(foreign,request),/scope/);
});
