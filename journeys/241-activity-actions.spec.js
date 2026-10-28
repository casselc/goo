const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const {execFile}=require('node:child_process'),{promisify}=require('node:util');
const fs=require('node:fs/promises'),os=require('node:os'),path=require('node:path');
const {startRuntime}=require('./191-generic-runtime');
const {source,verifyWithGo,startBombadil}=require('./support/cross-app');

test('241 compact Activity actions stage qualified service and namespace scopes without reads',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);
 const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-actions-241-')),dir=path.resolve('artifacts/activity-actions',randomBytes(4).toString('hex'));
 let runtime;
 try{
  const binary=process.env.GOO_GENERIC_BINARY||path.join(owned,'goo');
  if(!process.env.GOO_GENERIC_BINARY)await promisify(execFile)(process.env.GOO_TEST_GO||'go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});
  runtime=await startRuntime(binary,dir);
  const environment='inspect-'+randomBytes(4).toString('hex'),attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}})),at=BigInt(Date.now()-1000)*1000000n;
  const resources=service=>({attributes:attrs({'service.name':service,'service.namespace':'shop','deployment.environment.name':environment})});
  const logs=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:['inspect-api','other-api'].map(service=>({resource:resources(service),scopeLogs:[{logRecords:Array.from({length:service==='inspect-api'?12:1},(_,i)=>({timeUnixNano:String(at+BigInt(i)),body:{stringValue:'inspect log '+i+(i%2?' long ordinary telemetry '.repeat(20):'')},severityText:'INFO',attributes:attrs({region:'west'})}))}]}))}});
  expect(logs.ok()).toBe(true);
  const spans=await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:['inspect-api','other-api'].map(service=>({resource:resources(service),scopeSpans:[{spans:Array.from({length:service==='inspect-api'?5:1},(_,i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'inspect span '+i,kind:2,startTimeUnixNano:String(at+BigInt(i)),endTimeUnixNano:String(at+1000000n+BigInt(i)),attributes:attrs({region:'west'})}))}]}))}});
  expect(spans.ok()).toBe(true);
  for(const [table,count] of [['otel_logs',13],['otel_traces',6]])await expect.poll(async()=>Number((await source(request,`SELECT count() AS n FROM default.${table} WHERE ResourceAttributes['deployment.environment.name']={env:String} FORMAT JSONEachRow`,{param_env:environment}))[0].n),{timeout:45000}).toBe(count);
  const where={field:{kind:'column',key:'service'},op:'eq',value:'inspect-api'};
  const save=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Inspect service',selector:{signal:'logs',environment,where},columns:['utc','service','summary']}});expect(save.status()).toBe(201);const view=await save.json();
  const streams=[],events=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/tail')streams.push(r.postDataJSON())});
  const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),rows=tail.locator('tbody tr'),details=shell.locator(':scope>.ui-inspector');
  const event=async action=>events.push({action,reads:streams.length,retained:Number(await tail.getAttribute('data-retained')),gap:await tail.getAttribute('data-gap')==='true',coherent:!errors.length});
  await page.goto(runtime.base+'/#activity-view='+view.id);await shell.getByRole('button',{name:'Start stream',exact:true}).click();await expect(rows).toHaveCount(12);await event('connect');
  await shell.getByRole('button',{name:'Pause stream',exact:true}).click();await event('pause');
  const geometry=[];
  // Identical supplied records and paused scope: compare prior presentation against this cut.
  const prior={};for(const file of ['tail.mjs','attributes.mjs','stream-entry.mjs','product.css'])prior[file]=(await promisify(execFile)('git',['show','2776bea:ui/'+file])).stdout;
   const baseline=await page.context().newPage();
   await baseline.route('**/*',route=>{const file=new URL(route.request().url()).pathname.slice(1);return prior[file]!==undefined?route.fulfill({body:prior[file],contentType:file.endsWith('.css')?'text/css':'text/javascript'}):route.continue()});
   await baseline.goto(runtime.base+'/#activity-view='+view.id);await baseline.locator('.stream-entry').getByRole('button',{name:'Start stream',exact:true}).click();await expect(baseline.locator('.live-tail tbody tr')).toHaveCount(12);await baseline.locator('.stream-entry').getByRole('button',{name:'Pause stream',exact:true}).click();
  const beforeGeometry=new Map();for(const width of [320,390,700,1440]){await baseline.setViewportSize({width,height:844});
   const before=await baseline.locator('.live-tail tbody tr').evaluateAll(rs=>({heights:rs.slice(0,2).map(r=>r.getBoundingClientRect().height),visible:rs.filter(r=>{const b=r.getBoundingClientRect(),v=r.closest('goo-table').getBoundingClientRect();return b.top>=v.top&&b.bottom<=v.bottom}).length}));beforeGeometry.set(width,before);}
  await baseline.close();
  for(const width of [320,390,700,1440]){
   await page.setViewportSize({width,height:844});
   const measure=()=>rows.evaluateAll(rs=>({heights:rs.slice(0,2).map(r=>r.getBoundingClientRect().height),visible:rs.filter(r=>{const b=r.getBoundingClientRect(),v=r.closest('goo-table').getBoundingClientRect();return b.top>=v.top&&b.bottom<=v.bottom}).length}));
   const after=await measure();
   const before=beforeGeometry.get(width);geometry.push({width,before,after});expect(after.heights[0]).toBeLessThanOrEqual(before.heights[0]);expect(after.heights[1]).toBeLessThanOrEqual(before.heights[1]);expect(after.visible).toBeGreaterThanOrEqual(before.visible);
   expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
   const opener=rows.first().locator('[data-tail-inspect]');await opener.focus();await opener.press('Enter');await event('pause');
   await details.getByRole('searchbox').fill('region');const include=details.getByRole('button',{name:'Include attribute:region from record',exact:true});await include.focus();await include.press('Enter');expect(streams).toHaveLength(1);await event('hold');
   const fieldMenu=details.getByText('Field actions',{exact:true});await fieldMenu.click();await fieldMenu.press('Escape');await expect(details).toBeVisible();await expect(fieldMenu).toBeFocused();expect(streams).toHaveLength(1);await details.getByRole('button',{name:'Close selected item details'}).click();await expect(opener).toBeFocused();
   if(width===390||width===1440)await page.screenshot({path:'docs/journeys/screenshots/241-actions-'+width+'.png'});
  }
  await rows.first().evaluate(row=>{Object.defineProperty(document,'hidden',{configurable:true,value:true});try{for(const b of row.querySelectorAll('button'))b.click()}finally{delete document.hidden}});await expect(details).toBeHidden();await expect(tail).toHaveAttribute('data-selected','0');expect(streams).toHaveLength(1);await event('hold');
  const menu=rows.first().getByLabel('Filter service',{exact:true});await menu.focus();await menu.press('Enter');await event('hold');expect(streams).toHaveLength(1);
  const service=rows.first().getByRole('button',{name:'Include column:service from row',exact:true});await service.focus();await service.press('Enter');await event('hold');
  await menu.click();await rows.first().getByRole('button',{name:'Include resource:service.namespace from row',exact:true}).click();await event('hold');expect(streams).toHaveLength(1);
  await shell.getByRole('button',{name:'Apply scope',exact:true}).click();await expect(rows).toHaveCount(12);await event('connect');
  expect(streams.at(-1).where.all).toContainEqual({field:{kind:'resource',key:'service.namespace'},op:'eq',value:'shop'});
  expect(streams.at(-1).where.all).toContainEqual({field:{kind:'column',key:'service'},op:'eq',value:'inspect-api'});
  // Detached callbacks cannot pause or select into the replacement stream.
  await rows.first().evaluate(row=>{window.oldRow=row;row.remove()});await page.evaluate(()=>{for(const b of window.oldRow.querySelectorAll('button'))b.click()});await expect(details).toBeHidden();await expect(tail).toHaveAttribute('data-selected','0');expect(streams).toHaveLength(2);await event('hold');
  await shell.getByLabel('Live signal',{exact:true}).selectOption('spans');await expect(rows).toHaveCount(5);await event('connect');await rows.first().evaluate(row=>{window.oldSpan=row;row.remove()});await page.evaluate(()=>{for(const b of window.oldSpan.querySelectorAll('button'))b.click()});await expect(shell).toBeVisible();await expect(details).toBeHidden();await expect(tail).toHaveAttribute('data-selected','0');expect(streams).toHaveLength(3);await event('hold');
  const touch=await page.context().browser().newContext({viewport:{width:390,height:844},isMobile:true,hasTouch:true});
  try{const mobile=await touch.newPage(),requests=[];mobile.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/tail')requests.push(r.postDataJSON())});await mobile.goto(runtime.base+'/#activity-view='+view.id);const app=mobile.locator('.stream-entry');await app.getByRole('button',{name:'Start stream',exact:true}).tap();await expect(app.locator('tbody tr')).toHaveCount(12);const first=app.locator('tbody tr').first(),scope=first.getByLabel('Filter service',{exact:true});await scope.tap();expect((await scope.boundingBox()).height).toBeGreaterThanOrEqual(40);await first.getByRole('button',{name:'Include resource:service.namespace from row',exact:true}).tap();expect(requests).toHaveLength(1);await expect(app.getByRole('button',{name:'Apply scope',exact:true})).toBeVisible();await first.locator('[data-tail-inspect]').tap();await app.getByRole('searchbox').fill('region');await app.getByRole('button',{name:'Include attribute:region from record',exact:true}).tap();await app.getByRole('button',{name:'Close selected item details',exact:true}).tap();expect(requests).toHaveLength(1);expect(await mobile.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true)}finally{await touch.close()}
  expect(errors).toEqual([]);const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({streams,events,errors,unexpected:[],geometry}));await verifyWithGo(dir,'TestQuintActivityNavigationLedger',{GOO_ACTIVITY_NAVIGATION_LEDGER:ledger});
  if(process.env.GOO_ACTIONS_BOMBADIL==='1'){const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/#activity-view='+view.id,specification:'verification/bombadil/inspection-return.ts',seconds:20});try{await actor.finish()}finally{await actor.stop()}}
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
