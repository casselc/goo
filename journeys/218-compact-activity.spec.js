const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const {source}=require('./support/cross-app');
const base='http://127.0.0.1:8088';
test('218 compact ordinary OTel log and span rows retain UTC, actions and full inspection',async({page,request})=>{
 const environment='density-'+randomBytes(4).toString('hex'),trace=randomBytes(16).toString('hex');
 const at=BigInt(Date.now()-1000)*1000000n+123n,attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 const resource={attributes:attrs({'service.name':'catalog-api','service.namespace':'shop','deployment.environment.name':environment})};
 const long='Full diagnostic payload: '+'preserved '.repeat(180);
 const logRecords=Array.from({length:8},(_,i)=>({timeUnixNano:String(at+BigInt(i)*1000000n),traceId:trace,spanId:'0000000000000002',severityText:i===6?'WARN':'INFO',body:{stringValue:i===7?long:'Catalog request completed '+i}}));
 const spans=[{traceId:trace,spanId:'0000000000000001',name:'GET /catalog',kind:2,startTimeUnixNano:String(at),endTimeUnixNano:String(at+15000000n),status:{code:1}},{traceId:trace,spanId:'0000000000000002',parentSpanId:'0000000000000001',name:'cache.lookup',kind:3,startTimeUnixNano:String(at+1000000n),endTimeUnixNano:String(at+4000000n),status:{code:1}}];
 expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource,scopeLogs:[{logRecords}]}]}})).ok()).toBe(true);
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource,scopeSpans:[{spans}]}]}})).ok()).toBe(true);
 for(const table of ['otel_logs','otel_traces'])await expect.poll(async()=>Number((await source(request,"SELECT count() AS n FROM default."+table+" WHERE ResourceAttributes['deployment.environment.name']={env:String} FORMAT JSONEachRow",{param_env:environment}))[0].n),{timeout:45000}).toBeGreaterThan(0);
 const reads=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&/\/api\/(tail|explore|measurements)$/.test(new URL(r.url()).pathname))reads.push(r.url())});
 await page.goto(base+'/');const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail');
 await shell.getByLabel('Live environment').fill(environment);await shell.getByRole('button',{name:'Apply scope',exact:true}).click();
 for(const signal of ['logs','spans']){
  if(signal==='spans')await shell.getByLabel('Live signal').selectOption('spans');
  await expect(tail.locator('tbody>tr')).toHaveCount(signal==='logs'?8:2);
  await shell.getByRole('button',{name:'Pause stream',exact:true}).click();
  const row=tail.locator('tbody>tr').filter({hasText:signal==='logs'?'Catalog request completed 0':'GET /catalog'}).first();
  await page.setViewportSize({width:1440,height:1000});expect((await row.boundingBox()).height).toBeLessThan(48);
  const exact=await row.locator('time').getAttribute('datetime');expect(exact).toMatch(/Z$/);await expect(row.locator('time')).toHaveAttribute('title',/ns since Unix epoch/);
  await tail.evaluate(e=>{e.querySelector('goo-table').scrollTop=0});
  await page.evaluate(()=>scrollTo(0,0));await page.screenshot({path:'docs/journeys/screenshots/218a-compact-'+signal+'-desktop.png',fullPage:false});
  for(const width of [320,390,700,701]){
   await page.setViewportSize({width,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
   if(width<=700)for(const button of await row.getByRole('button').all())expect((await button.boundingBox()).height).toBeGreaterThanOrEqual(28);
  }
  const before=reads.length,inspect=row.locator('[data-tail-inspect]');await inspect.focus();await page.keyboard.press('Enter');
  const detail=shell.locator(':scope>.ui-inspector');await expect(detail).toBeVisible();await expect(detail.locator('.attribute-row[data-key=timestamp_ns] time')).toHaveAttribute('datetime',exact);
  await detail.getByRole('button',{name:'Close selected item details',exact:true}).click();await expect(inspect).toBeFocused();expect(reads).toHaveLength(before);
  if(signal==='logs'){
   const longRow=tail.locator('tbody>tr').filter({hasText:'Full diagnostic payload:'});const label=await longRow.locator('.record-label').boundingBox();expect(label.width).toBeGreaterThan(15);expect(label.height).toBeLessThan(24);
   await longRow.locator('[data-tail-inspect]').click();
   const disclosure=detail.locator('goo-record-summary details');await disclosure.locator(':scope>summary').click();await expect(disclosure.locator('div')).toHaveText(long);
   await detail.getByRole('button',{name:'Close selected item details',exact:true}).click();expect(reads).toHaveLength(before);
  }
  await page.setViewportSize({width:390,height:844});await tail.evaluate(e=>{e.querySelector('goo-table').scrollTop=0});
  await page.evaluate(()=>scrollTo(0,0));await page.screenshot({path:'docs/journeys/screenshots/218b-compact-'+signal+'-mobile.png',fullPage:false});
 }
 expect(errors).toEqual([]);
});
