const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const {execFile}=require('node:child_process'),{promisify}=require('node:util');
const fs=require('node:fs/promises'),os=require('node:os'),path=require('node:path');
const {startRuntime}=require('./191-generic-runtime');
const {source,verifyWithGo,startBombadil}=require('./support/cross-app');

test('240 explicit Activity Inspect reveals named details and returns to its row without a source read',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);
 const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-inspect-240-')),dir=path.resolve('artifacts/activity-inspect',randomBytes(4).toString('hex'));
 let runtime;
 try{
  const binary=process.env.GOO_GENERIC_BINARY||path.join(owned,'goo');
  if(!process.env.GOO_GENERIC_BINARY)await promisify(execFile)(process.env.GOO_TEST_GO||'go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});
  runtime=await startRuntime(binary,dir);
  const environment='inspect-'+randomBytes(4).toString('hex'),attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}})),at=BigInt(Date.now()-1000)*1000000n;
  const resources=service=>({attributes:attrs({'service.name':service,'deployment.environment.name':environment})});
  const logs=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:['inspect-api','other-api'].map(service=>({resource:resources(service),scopeLogs:[{logRecords:Array.from({length:service==='inspect-api'?12:1},(_,i)=>({timeUnixNano:String(at+BigInt(i)),body:{stringValue:'inspect log '+i},severityText:'INFO',attributes:attrs({region:'west'})}))}]}))}});
  expect(logs.ok()).toBe(true);
  const spans=await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:['inspect-api','other-api'].map(service=>({resource:resources(service),scopeSpans:[{spans:Array.from({length:service==='inspect-api'?5:1},(_,i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'inspect span '+i,kind:2,startTimeUnixNano:String(at+BigInt(i)),endTimeUnixNano:String(at+1000000n+BigInt(i)),attributes:attrs({region:'west'})}))}]}))}});
  expect(spans.ok()).toBe(true);
  for(const [table,count] of [['otel_logs',13],['otel_traces',6]])await expect.poll(async()=>Number((await source(request,`SELECT count() AS n FROM default.${table} WHERE ResourceAttributes['deployment.environment.name']={env:String} FORMAT JSONEachRow`,{param_env:environment}))[0].n),{timeout:45000}).toBe(count);
  const where={field:{kind:'column',key:'service'},op:'eq',value:'inspect-api'};
  const save=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Inspect service',selector:{signal:'logs',environment,where},columns:['utc','service','summary']}});expect(save.status()).toBe(201);const view=await save.json();
  const streams=[],errors=[],unexpected=[],events=[];
  page.on('pageerror',error=>errors.push(error.message));page.on('request',req=>{if(req.method()!=='POST')return;const route=new URL(req.url()).pathname;if(route==='/api/tail')streams.push(req.postDataJSON());else if(!route.startsWith('/api/queries/'))unexpected.push(route)});
  const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),details=shell.locator(':scope>.ui-inspector'),heading=details.locator('.ui-inspector-heading>h3'),close=details.getByRole('button',{name:'Close selected item details',exact:true});
  const event=async action=>events.push({action,reads:streams.length,retained:Number(await tail.getAttribute('data-retained')),gap:await tail.getAttribute('data-gap')==='true',coherent:!errors.length&&!unexpected.length});
  async function visible(locator){const box=await locator.boundingBox(),size=page.viewportSize();expect(box).not.toBeNull();expect(box.y).toBeGreaterThanOrEqual(0);expect(box.y+box.height).toBeLessThanOrEqual(size.height);expect(box.x).toBeGreaterThanOrEqual(0);expect(box.x+box.width).toBeLessThanOrEqual(size.width)}
  await page.goto(runtime.base+'/#activity-view='+view.id);await shell.getByRole('button',{name:'Start stream',exact:true}).click();await expect(tail.locator('tbody tr')).toHaveCount(12);await event('connect');
  await shell.getByLabel('Live signal',{exact:true}).selectOption('spans');await expect(tail.locator('tbody tr')).toHaveCount(5);await event('connect');
  expect(streams).toEqual([{environment,where},{signal:'spans',environment,where}]);
  for(const width of [1440,320,390,700]){
   await page.setViewportSize({width,height:844});
   for(const keyboard of [false,true]){
    const opener=tail.locator('tbody tr').filter({hasText:'inspect span 2'}).locator('[data-tail-inspect]'),before=streams.length;
    if(keyboard){await opener.focus();await page.keyboard.press('Enter')}else await opener.click();
    await expect(tail).toHaveAttribute('data-state','paused');await expect(heading).toBeFocused();await expect(heading).toHaveText('Observed span');await expect(details).toHaveAccessibleName('Observed span');
    await expect(opener).toHaveAttribute('aria-expanded','true');expect(await opener.getAttribute('aria-controls')).toBe(await details.getAttribute('id'));
    await visible(heading);await visible(close);await event('pause');expect(streams).toHaveLength(before);
    if(width===390&&!keyboard)await page.screenshot({path:'docs/journeys/screenshots/240-inspect-visible-mobile.png'});
    if(width===1440&&!keyboard)await page.screenshot({path:'docs/journeys/screenshots/240-inspect-visible-desktop.png'});
    if(width===390&&!keyboard){const search=details.getByRole('searchbox',{name:'Search attribute keys and values',exact:true});await search.fill('region');await search.press('Escape');await expect(details).toBeVisible();await expect(search).toBeFocused();const scope=details.getByLabel('Attribute scope',{exact:true});await scope.focus();await scope.press('Escape');await expect(details).toBeVisible();await expect(scope).toBeFocused();expect(streams).toHaveLength(before)}
    if(keyboard){await page.keyboard.press('Tab');await expect(close).toBeFocused();await page.keyboard.press('Escape')}else await close.click();
    await expect(details).toBeHidden();await expect(opener).toBeFocused();await expect(opener).toHaveAttribute('aria-expanded','false');await visible(opener);await event('hold');expect(streams).toHaveLength(before);
   }
  }

  // A fixed child stages without a read and returns to the action that opened it.
  const childOpener=tail.locator('tbody tr').filter({hasText:'inspect span 2'}).locator('[data-tail-inspect]');await childOpener.click();await event('pause');
  const fixed=details.getByRole('button',{name:'Analyze this observation window',exact:true}),childReads=streams.length;
  await fixed.click();await expect(shell).toBeHidden();expect(streams).toHaveLength(childReads);await event('hold');
  await page.getByRole('button',{name:'Return to paused activity',exact:true}).click();await expect(fixed).toBeFocused();await visible(fixed);await expect(details).toBeVisible();await event('hold');
  await close.click();await expect(childOpener).toBeFocused();await visible(childOpener);await event('hold');

  // A DOM replacement can withdraw the paused opener. Close must use a live fallback.
  const withdrawn=tail.locator('tbody tr').filter({hasText:'inspect span 3'}).locator('[data-tail-inspect]');await withdrawn.click();await event('pause');
  await withdrawn.evaluate(button=>{window.withdrawnInspect=button;button.closest('tr').remove()});await close.click();
  await expect(shell.getByLabel('Live environment',{exact:true})).toBeFocused();await visible(shell.getByLabel('Live environment',{exact:true}));await event('hold');
  expect(await page.evaluate(()=>window.withdrawnInspect.getAttribute('aria-expanded'))).toBe('false');

  // Cache updates in a passive inspector must not acquire focus automatically.
  await page.evaluate(async()=>{
   const UI=await import('/components.mjs'),panel=document.createElement('section'),content=document.createElement('p');document.body.append(panel);
   const inspector=UI.inspector(panel),before=document.activeElement;content.textContent='Synthetic passive contract';
   inspector.show({key:'passive',title:'Passive update',content});inspector.show({key:'passive',title:'Updated supplied data'});
   if(document.activeElement!==before)throw Error('Passive inspector stole focus');inspector.clear();panel.remove();
  });await event('hold');

  // Withdrawn/cached callbacks cannot open old details or stage source work in a new signal.
  const old=tail.locator('tbody tr').filter({hasText:'inspect span 2'}).locator('[data-tail-inspect]');await old.click();await event('pause');
  await details.evaluate(panel=>{window.oldInspect=panel.ownerDocument.querySelector('[data-tail-inspect][aria-expanded=true]');window.oldFixed=[...panel.querySelectorAll('button')].find(b=>b.textContent==='Analyze this observation window');window.oldFacet=panel.querySelector('[aria-label="Include attribute:region from record"]')});
  await shell.getByLabel('Live signal',{exact:true}).selectOption('logs');await expect(tail.locator('tbody tr')).toHaveCount(12);await expect(details).toBeHidden();await event('connect');
  expect(await page.evaluate(()=>window.oldInspect.getAttribute('aria-expanded'))).toBe('false');const before=streams.length;
  await page.evaluate(()=>{window.oldInspect.click();window.oldFacet.click();window.oldFixed.click();window.withdrawnInspect.click()});
  await expect(details).toBeHidden();await expect(tail).toHaveAttribute('data-state',/unknown|partial/);await expect(shell.locator('.entry-filters')).toHaveAttribute('data-phase','applied');expect(streams).toHaveLength(before);await event('hold');

  // Resume is explicit, invalidates details and cannot retain a hidden focused child.
  const logOpener=tail.locator('tbody tr').filter({hasText:'inspect log 2'}).locator('[data-tail-inspect]');await logOpener.click();await expect(heading).toBeFocused();await event('pause');
  await shell.getByRole('button',{name:'Resume with recent logs',exact:true}).evaluate(button=>button.click());
  await expect(details).toBeHidden();await expect(tail).toHaveAttribute('data-gap','true');await expect(tail).toHaveAttribute('data-state',/unknown|partial/);await expect(shell.getByLabel('Live environment',{exact:true})).toBeFocused();await event('resume');
  expect(streams).toHaveLength(before+1);expect(errors).toEqual([]);expect(unexpected).toEqual([]);
  const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({streams,events,errors,unexpected,firstRowY:await tail.locator('tbody tr').first().evaluate(row=>row.getBoundingClientRect().top)}));
  await verifyWithGo(dir,'TestQuintStreamEntryLedger',{GOO_STREAM_ENTRY_LEDGER:ledger});
  if(process.env.GOO_INSPECT_BOMBADIL==='1'){
   const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/#activity-view='+view.id,specification:'verification/bombadil/inspection-return.ts',seconds:20});try{await actor.finish()}finally{await actor.stop()}
  }
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
