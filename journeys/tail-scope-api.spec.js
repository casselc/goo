'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const {startRuntime}=require('./191-generic-runtime');
const F=require('./support/activity-facets');
const sourcePaths=new Set(['/api/tail','/api/explore','/api/explore/page','/api/record-aggregates','/api/explore/counted-values','/api/discovery','/api/inventory','/api/metric-discovery','/api/measurements']);

test('423 Tail scope API is current, detached metadata while full snapshots retain independent rows',async({page,request})=>{
 test.setTimeout(120000);
 const dir=path.resolve('artifacts/tail-scope-api423');await fs.mkdir(dir,{recursive:true});
 const report={version:'goo.tailScopeAPI423/1',errors:[],calls:[]};let runtime;
 page.on('pageerror',error=>report.errors.push(error.message));
 page.on('request',r=>{const p=new URL(r.url()).pathname;if(r.method()==='POST'&&sourcePaths.has(p))report.calls.push({path:p})});
 try{
  runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir);const fixture=F.fixture('logs',12);report.raw=await F.ingest(request,fixture);
  await page.goto(runtime.base+'/#questions');
  const before=report.calls.length;
  report.staged=await page.evaluate(async selector=>{
   const {mount}=await import('/tail.mjs'),host=document.createElement('section');document.body.append(host);
   const owner=mount(host,{compact:true});window.__scopeAPI423={owner,host};
   const original=structuredClone(selector);owner.stage(selector);
   const first=owner.scope(),second=owner.scope(),full=owner.snapshot(),{rows,...metadata}=full;
   const out={scope:structuredClone(first),metadata,rows:rows.length,distinct:first!==second&&first.selector!==second.selector,hasRows:Object.hasOwn(first,'rows')};
   first.selector.environmentSelector.scope.value='forged';first.selector.where={field:{kind:'column',key:'service'},op:'missing'};first.phase='forged';
   selector.where={field:{kind:'column',key:'service'},op:'missing'};
   out.detached=JSON.stringify(owner.scope())===JSON.stringify(metadata);
   owner.reset();out.reset=owner.scope();owner.stage(original);out.restaged=owner.scope();
   return out;
  },fixture.selector);
  expect(report.calls.length).toBe(before);expect(report.staged.hasRows).toBe(false);expect(report.staged.distinct).toBe(true);expect(report.staged.detached).toBe(true);expect(report.staged.rows).toBe(0);expect(report.staged.scope.phase).toBe('paused');expect(report.staged.scope.staged).toBe(true);expect(report.staged.reset.phase).toBe('stopped');expect(report.staged.reset.selector).toBe(null);expect(report.staged.restaged.phase).toBe('paused');
  await page.evaluate(selector=>window.__scopeAPI423.owner.start(selector),fixture.selector);
  await expect.poll(()=>page.evaluate(()=>window.__scopeAPI423.owner.snapshot().rows.length),{timeout:60000}).toBe(12);
  await page.evaluate(()=>window.__scopeAPI423.owner.pause());
  expect(report.calls.length).toBe(before+1);expect(report.calls.at(-1).path).toBe('/api/tail');
  report.observed=await page.evaluate(()=>{
   const owner=window.__scopeAPI423.owner,scope=owner.scope(),full=owner.snapshot(),{rows,...metadata}=full;
   const original=JSON.stringify(metadata),ids=rows.map(row=>row.record.Attributes['work.id']);
   const sourceClone=window.structuredClone;let itemCopies=0;
   window.structuredClone=function(value,...rest){if(value?.record&&value?.readWindow)itemCopies++;return sourceClone.call(window,value,...rest)};
   try{owner.scope();}finally{window.structuredClone=sourceClone}
   const scopeCopies=itemCopies;window.structuredClone=function(value,...rest){if(value?.record&&value?.readWindow)itemCopies++;return sourceClone.call(window,value,...rest)};
   try{owner.snapshot();}finally{window.structuredClone=sourceClone}
   scope.readWindow.fromMs=-1;scope.selector.environmentSelector.scope.value='forged';rows[0].record.Attributes['work.id']='forged';
   const next=owner.snapshot(),detached=JSON.stringify(owner.scope())===original&&JSON.stringify(next.rows.map(row=>row.record.Attributes['work.id']))===JSON.stringify(ids);
   return {metadata,ids,hasRows:Object.hasOwn(owner.scope(),'rows'),scopeCopies,snapshotCopies:itemCopies-scopeCopies,detached};
  });
  expect(report.observed.metadata.phase).toBe('paused');expect(report.observed.metadata.staged).toBe(false);expect(report.observed.metadata.readWindow.toMs-report.observed.metadata.readWindow.fromMs).toBe(60000);expect(report.observed.hasRows).toBe(false);expect(report.observed.scopeCopies).toBe(0);expect(report.observed.snapshotCopies).toBe(12);expect(report.observed.detached).toBe(true);expect(report.observed.ids.sort()).toEqual(fixture.rows.filter(r=>r.environment===fixture.environment&&r.bank===fixture.token).map(r=>r.id).sort());
  await page.evaluate(()=>{window.__scopeAPI423.owner.dispose();window.__scopeAPI423.host.remove();delete window.__scopeAPI423});expect(report.errors).toEqual([]);report.status='passed';
 }finally{await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));await runtime?.stop()}
});
