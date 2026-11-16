'use strict';
const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const fs=require('node:fs/promises'),path=require('node:path');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {questionLink}=require('../ui/workspace.mjs');
const {revealControl}=require('./support/explore');
const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));

function bank(){
 const token=randomBytes(5).toString('hex'),environment='polish459-'+token,at=Date.now()-10000;
 const records=Array.from({length:30},(_,i)=>({id:randomBytes(8).toString('hex'),traceId:randomBytes(16).toString('hex'),service:['edge-api','queue-worker','catalog-db'][i%3],at:at+i*10,durationNS:String((i%5+1)*1250000),status:i%7===0?2:1,severity:i%7===0?'ERROR':'INFO',message:`${i%7===0?'Retrying upstream request':'Processed work item'} ${i+1} · region eu-west · request payload validated; delivery accepted; diagnostic context remains available in the inspector.`,attributes:{'http.route':['/search','/jobs','/catalog'][i%3],'work.region':'eu-west','work.queue':'interactive'}}));
 return{environment,records,question:{signal:'spans',environment,fromMs:at-1000,toMs:at+1000,limit:100,offset:0}};
}
async function emit(request,b){
 for(const signal of['logs','spans']){
  const rows=b.records.map(r=>{
   const resource={attributes:attrs({'service.name':r.service,'service.namespace':'platform','deployment.environment.name':b.environment})},start=BigInt(r.at)*1000000n;
   return signal==='logs'?{resource,scopeLogs:[{scope:{name:'ordinary.operator'},logRecords:[{timeUnixNano:String(start),spanId:r.id,traceId:r.traceId,severityText:r.severity,severityNumber:r.status===2?17:9,body:{stringValue:r.message},attributes:attrs(r.attributes)}]}]}:{resource,scopeSpans:[{scope:{name:'ordinary.operator'},spans:[{startTimeUnixNano:String(start),endTimeUnixNano:String(start+BigInt(r.durationNS)),spanId:r.id,traceId:r.traceId,name:r.attributes['http.route'],kind:r.service==='catalog-db'?3:2,status:{code:r.status},attributes:attrs(r.attributes)}]}]};
  });
  expect((await request.post('http://127.0.0.1:4318/v1/'+(signal==='logs'?'logs':'traces'),{data:signal==='logs'?{resourceLogs:rows}:{resourceSpans:rows}})).ok()).toBe(true);
 }
}

test('459 shared polish: activity, inspect, fixed records, services and saved evidence',async({page,request})=>{
 test.setTimeout(240000);page.setDefaultTimeout(20000);
 const pass=process.env.GOO_POLISH_PASS;if(!['before','after'].includes(pass))throw Error('Set GOO_POLISH_PASS=before or after');
 const binary=process.env.GOO_GENERIC_BINARY;if(!binary)throw Error('Reviewed binary required');
 const dir=path.resolve('artifacts/ui-polish459');await fs.mkdir(dir,{recursive:true});
 const bankPath=path.join(dir,'bank.json'),fixed=pass==='before'?bank():JSON.parse(await fs.readFile(bankPath,'utf8'));
 if(pass==='before')await fs.writeFile(bankPath,JSON.stringify(fixed,null,2));
 const previous=pass==='after'?JSON.parse(await fs.readFile(path.join(dir,'before.json'),'utf8')):null;
 const report={pass,status:'running',prose:{},fixedQuestion:fixed.question,frames:[],local:[],errors:[]};let runtime,proxy;
 const calls=[];page.on('pageerror',e=>report.errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/'))calls.push(new URL(r.url()).pathname)});
 const effects=()=>({posts:calls.length,sql:proxy.calls.length});
 const settle=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
 const response=p=>page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===p);
 const read=async(button,p)=>{const pending=response(p);await button.click();const r=await pending;expect(r.ok()).toBe(true);return r.json()};
 async function frame(scene,rows){
  report.prose[scene]=await page.locator('p').evaluateAll(nodes=>nodes.filter(e=>e.getClientRects().length&&e.innerText.trim()).map(e=>({text:e.innerText,owner:e.closest('[id]')?.id||'',className:e.className})));
  for(const width of(pass==='after'?[1440,390,320]:[1440,390])){
   await page.setViewportSize({width,height:width===390?844:1000});await page.evaluate(()=>scrollTo(0,0));await settle();
   const measured=await rows.evaluateAll(nodes=>{const rects=nodes.map(n=>n.getBoundingClientRect());return{firstRow:rects[0]?.top,rowHeight:rects[0]?.height,visible:rects.filter(r=>r.top>=0&&r.bottom<=innerHeight).length,pageOverflow:document.documentElement.scrollWidth>innerWidth+1,cells:nodes[0]?[...nodes[0].children].map(e=>({label:e.dataset.label,text:e.innerText,height:e.getBoundingClientRect().height,width:e.getBoundingClientRect().width})):[]}});
   report.frames.push({scene,width,...measured});
   expect(measured.pageOverflow,scene+' overflow').toBe(false);if(scene==='services'){expect(measured.cells).toHaveLength(6);expect(measured.cells.every(cell=>cell.height>0&&cell.width>0)).toBe(true);expect(measured.cells[3].text).toBe('Not observed');if(previous&&width===390){const old=previous.frames.find(frame=>frame.scene===scene&&frame.width===width);expect(measured.rowHeight).toBeLessThan(old.rowHeight)}}
   const image=path.join(dir,`${pass}-${scene}-${width}.png`);await page.screenshot({path:image});if(pass==='after')await fs.copyFile(image,'docs/journeys/screenshots/459-'+scene+'-'+width+'.png');
  }
 }
 async function skipRoute(scene){
  if(pass!=='after')return;
  const before=effects(),hash=await page.evaluate(()=>location.hash);await page.evaluate(()=>{document.body.tabIndex=-1;document.body.focus();document.body.removeAttribute('tabindex');scrollTo(0,0)});await page.keyboard.press('Tab');
  const skip=page.getByRole('link',{name:'Skip to main content',exact:true});await expect(skip).toBeFocused();await expect(skip).toBeInViewport();await page.keyboard.press('Enter');await expect(page.locator('#mainContent')).toBeFocused();expect(await page.evaluate(()=>location.hash)).toBe(hash);await settle();expect(effects()).toEqual(before);report.local.push({scene:'skip-'+scene,before,after:effects(),hash});
 }
 async function fieldControls(inspector,scene){
  const before=effects(),scope=inspector.getByRole('combobox',{name:'Attribute scope',exact:true}),search=inspector.getByRole('searchbox',{name:'Search attribute keys and values',exact:true});
  await scope.selectOption('Resource attributes');await search.fill('service.name');const row=inspector.locator('.attribute-row[data-key="service.name"]');await expect(row).toHaveCount(1);await expect(row).toContainText(/edge-api|queue-worker|catalog-db/);
  if(pass==='after')for(const width of [390,320]){
   await page.setViewportSize({width,height:1000});const menu=row.getByLabel('Field actions for resource:service.name',{exact:true});await menu.scrollIntoViewIfNeeded();await menu.focus();await menu.press('Enter');await expect(menu.locator('..')).toHaveAttribute('open','');
   const action=menu.locator('..').locator('.ui-actions-body button').first();await action.scrollIntoViewIfNeeded();expect(await action.evaluate(e=>{const r=e.getBoundingClientRect();return e.contains(document.elementFromPoint(r.left+r.width/2,r.top+r.height/2))})).toBe(true);expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
   if(scene==='fixed'&&width===320){const image=path.join(dir,pass+'-field-actions-320.png');await page.screenshot({path:image});await fs.copyFile(image,'docs/journeys/screenshots/459-field-actions-320.png')}await menu.focus();await menu.press('Escape');await expect(menu).toBeFocused();await expect(inspector).toBeVisible();await expect(menu.locator('..')).not.toHaveAttribute('open','');
  }
  await scope.selectOption('All scopes');await search.fill('');await settle();expect(effects()).toEqual(before);report.local.push({scene:scene+'-attribute-controls',before,after:effects()});
 }
 async function inspectFrames(inspector,scene){
  for(const width of [1440,390,320]){
   await page.setViewportSize({width,height:1000});await inspector.locator('.ui-inspector-heading').evaluate(e=>e.scrollIntoView({block:'start'}));
   await inspector.locator('.ui-inspector-body').evaluate(e=>e.scrollTop=0);await settle();
   const geometry=await inspector.locator('.attribute-viewer').evaluate(root=>{
    const box=root.getBoundingClientRect(),tools=root.querySelector('.attribute-tools').getBoundingClientRect(),first=root.querySelector('.attribute-row').getBoundingClientRect();
    return {toolbarHeight:tools.height,firstFieldFromViewer:first.top-box.top,overflow:document.documentElement.scrollWidth>innerWidth,search:root.querySelector('input').getBoundingClientRect().height,scope:root.querySelector('select').getBoundingClientRect().height};
   });report.frames.push({scene,width,...geometry});expect(geometry.overflow).toBe(false);
   if(previous){const old=previous.frames.find(frame=>frame.scene===scene&&frame.width===width);if(width<=390)expect(geometry.firstFieldFromViewer).toBeLessThan(old.firstFieldFromViewer);}
   const image=path.join(dir,`${pass}-${scene}-${width}.png`);await page.screenshot({path:image});if(pass==='after')await fs.copyFile(image,'docs/journeys/screenshots/459-'+scene+'-'+width+'.png');
  }
 }
 function assertRows(result){
  expect(result.count).toBe(String(fixed.records.length));
  const actual=result.rows.map(r=>({id:r.SpanId,traceId:r.TraceId,time:r.timestamp_ns,duration:r.duration_ns,status:r.StatusCode,service:r.ServiceName})).sort((a,b)=>a.id.localeCompare(b.id));
  expect(actual.map(r=>r.id)).toEqual(fixed.records.map(r=>r.id).sort());
  for(const row of actual){const expected=fixed.records.find(r=>r.id===row.id);expect(row.time).toBe(String(BigInt(expected.at)*1000000n));expect(row.duration).toBe(expected.durationNS);expect(row.service).toBe(expected.service);expect(row.status).toBe(expected.status===2?'Error':'Ok')}
  return actual;
 }
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(binary,dir,{clickHouseURL:proxy.url});
  if(pass==='before')await emit(request,fixed);
  // Activity is a real server-clock recent window. Use a separate fresh bank
  // each pass; only fixed Records/Services/Saved are compared across binaries.
  const activity=bank();report.activityEnvironment=activity.environment;await emit(request,activity);
  const staged=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Platform operations',selector:{signal:'logs',environment:activity.environment},columns:['utc','service','summary']}});expect(staged.status()).toBe(201);const view=await staged.json();
  await page.goto(runtime.base+'/#activity-view='+view.id);const shell=page.getByRole('region',{name:'Live activity',exact:true});await shell.getByRole('button',{name:'Start stream',exact:true}).click();const liveRows=shell.locator('.live-tail tbody > tr');await expect(liveRows).toHaveCount(activity.records.length);await shell.getByRole('button',{name:'Pause stream',exact:true}).click();await frame('activity',liveRows);
  const opener=liveRows.first().getByRole('button',{name:'Inspect',exact:true});const beforeInspect=effects();await opener.click();const inspector=shell.locator(':scope > .ui-inspector');await expect(inspector).toBeVisible();const message=await liveRows.first().locator('goo-record-summary').evaluate(e=>e.record.Body);await expect(inspector.locator('goo-record-summary')).toContainText(message);
  if(pass==='after'){
   const before=effects(),help=inspector.locator('.ui-context-help > summary');await help.focus();await help.press('Enter');await expect(help.locator('..')).toHaveAttribute('open','');await expect(help).toBeFocused();await expect(help.locator('..')).toContainText('Stages a fixed question for a separate source read; held rows are not saved evidence.');await help.press('Space');await expect(help.locator('..')).not.toHaveAttribute('open','');await expect(help).toBeFocused();expect(effects()).toEqual(before);report.local.push({scene:'context-help',before,after:effects()});
  }

  const exactTime=await liveRows.first().locator('time').getAttribute('title');expect(exactTime).toContain('ns');report.inspectTimestamp=exactTime;
  const search=inspector.getByRole('searchbox',{name:'Search attribute keys and values'});await search.fill('work.region');await expect(inspector).toContainText('eu-west');await search.fill('');await inspectFrames(inspector,'activity-inspect');await fieldControls(inspector,'activity');await search.focus();await page.screenshot({path:path.join(dir,`${pass}-activity-inspect-390.png`)});
  if(pass==='after')await search.press('Escape');else await inspector.getByRole('button',{name:'Close selected item details',exact:true}).click();await expect(inspector).toBeHidden();await expect(opener).toBeFocused();await settle();expect(effects()).toEqual(beforeInspect);report.local.push({scene:'inspect-search-close',before:beforeInspect,after:effects()});
  await opener.click();const stageBefore=effects();await inspector.getByRole('button',{name:'Analyze this observation window',exact:true}).click();await expect(page.locator('#exploreRun')).toBeEnabled();expect(effects()).toEqual(stageBefore);await read(page.locator('#exploreRun'),'/api/explore');
  // Fixed comparison uses the identical question and source bank in both passes.
  await page.goto(runtime.base+'/'+questionLink(fixed.question));let result;await expect.poll(async()=>{const r=await request.post(runtime.base+'/api/explore',{data:fixed.question});return (await r.json()).count},{timeout:45000}).toBe(String(fixed.records.length));result=await read(page.locator('#exploreRun'),'/api/explore');report.fixedRows=assertRows(result);if(previous)expect(report.fixedRows).toEqual(previous.fixedRows);
  const records=page.locator('#exploreRows .record-row');await expect(records).toHaveCount(fixed.records.length);await frame('fixed',records);
  await page.setViewportSize({width:1440,height:1000});await page.evaluate(()=>scrollTo(0,0));
  const fixedOpener=records.first().getByRole('button',{name:'Inspect record',exact:true});const localBefore=effects();await fixedOpener.click();const fixedInspector=page.locator('#exploreRows').getByRole('region',{name:'Selected item details',exact:true});await expect(fixedInspector).toBeVisible();await fixedInspector.getByRole('searchbox',{name:'Search attribute keys and values'}).fill('work.region');await expect(fixedInspector).toContainText('eu-west');await fixedInspector.getByRole('searchbox',{name:'Search attribute keys and values'}).fill('');
  await fixedInspector.getByRole('searchbox',{name:'Search attribute keys and values'}).evaluate(e=>e.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,isComposing:true})));await expect(fixedInspector).toBeVisible();
  await fixedInspector.getByRole('combobox',{name:'Attribute scope',exact:true}).focus();await page.keyboard.press('Escape');await expect(fixedInspector).toBeVisible();
  await fixedInspector.getByRole('searchbox',{name:'Search attribute keys and values'}).focus();await settle();await page.screenshot({path:path.join(dir,`${pass}-fixed-inspect-1440.png`)});await fixedInspector.getByRole('button',{name:'Close selected item details',exact:true}).click();await expect(fixedOpener).toBeFocused();expect(effects()).toEqual(localBefore);report.local.push({scene:'fixed-inspect',before:localBefore,after:effects()});await fixedOpener.click();await expect(fixedInspector).toBeVisible();
  await inspectFrames(fixedInspector,'fixed-inspect');await fieldControls(fixedInspector,'fixed');
  const servicesNav=page.getByRole('button',{name:/^Services(?: & resources)?$/});if(!await servicesNav.isVisible())await page.getByRole('button',{name:'Menu',exact:true}).click();await servicesNav.click();await skipRoute('services');const catalog=page.locator('#serviceCatalog'),inventory=page.locator('#serviceInventory');await inventory.getByLabel('Environment',{exact:true}).filter({visible:true}).fill(fixed.environment);await revealControl(page.locator('#inventoryWindow'));await inventory.getByRole('combobox',{name:'Window',exact:true}).selectOption('fixed');await inventory.getByLabel('From UTC',{exact:true}).fill(new Date(fixed.question.fromMs).toISOString());await inventory.getByLabel('To UTC (exclusive)',{exact:true}).fill(new Date(fixed.question.toMs).toISOString());await read(inventory.getByRole('button',{name:'Discover services',exact:true}),'/api/discovery');for(const service of['edge-api','queue-worker','catalog-db'])await expect(catalog).toContainText(service);await frame('services',catalog.locator('tbody > tr'));
  // Capture rereads the fixed question explicitly; subsequent reopening is stored evidence.
  const captured=await request.post(runtime.base+'/api/captures',{data:{title:'Platform fixed evidence',question:{kind:'explore',explore:fixed.question}}});expect(captured.status()).toBe(201);const saved=await captured.json();expect(assertRows(saved.explore)).toEqual(report.fixedRows);report.captureQuestion=saved.explore.question;
  const reopenBefore=effects();await page.goto(runtime.base+'/#capture='+saved.receipt.id);await expect(page.locator('#captureSavedTitle')).toHaveText('Platform fixed evidence');await expect(page.locator('#savedexploreRows .record-row')).toHaveCount(fixed.records.length);await frame('saved',page.locator('#savedexploreRows .record-row'));await page.reload();await expect(page.locator('#captureSavedTitle')).toHaveText('Platform fixed evidence');await skipRoute('saved');await settle();expect(effects()).toEqual(reopenBefore);report.local.push({scene:'saved-reopen',before:reopenBefore,after:effects()});
  if(pass==='after'){
   const before=effects(),touch=await page.context().browser().newContext({hasTouch:true,isMobile:true,viewport:{width:390,height:1000},storageState:await page.context().storageState()});try{const mobile=await touch.newPage();mobile.setDefaultTimeout(15000);await mobile.goto(runtime.base+'/#capture='+saved.receipt.id);await expect(mobile.locator('#savedexploreRows .record-row')).toHaveCount(fixed.records.length);await mobile.locator('#savedexploreRows .record-row').first().getByRole('button',{name:'Inspect captured occurrence 1',exact:true}).click();const inspector=mobile.locator('#savedexploreRows .ui-inspector');const sizes=await inspector.locator('.attribute-tools').evaluate(e=>({coarse:matchMedia('(pointer:coarse)').matches,heights:[...e.querySelectorAll('input,select')].map(n=>n.getBoundingClientRect().height)}));expect(sizes.coarse).toBe(true);expect(sizes.heights.every(h=>h>=44)).toBe(true);const actions=await inspector.locator('.attribute-actions').first().evaluate(e=>[...e.querySelectorAll('button,summary')].filter(n=>n.getClientRects().length).map(n=>n.getBoundingClientRect().height));expect(actions.length).toBeGreaterThan(0);expect(actions.every(h=>h>=44)).toBe(true);report.coarse={...sizes,actions};expect(effects()).toEqual(before)}finally{await touch.close()}
  }
  if(pass==='after'){
   const before=effects(),context=await page.context().browser().newContext({javaScriptEnabled:false,viewport:{width:390,height:1000}});try{const native=await context.newPage();native.setDefaultTimeout(15000);await native.goto(runtime.base+'/');await native.keyboard.press('Tab');await expect(native.getByRole('link',{name:'Skip to main content',exact:true})).toBeFocused();await native.keyboard.press('Enter');await expect(native.locator('#mainContent')).toBeFocused();expect(new URL(native.url()).hash).toBe('#mainContent');expect(effects()).toEqual(before);report.nativeSkip=true;}finally{await context.close()}
  }
  expect(report.captureQuestion).toEqual(fixed.question);expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;await page.screenshot({path:path.join(dir,pass+'-failure.png')}).catch(()=>{});throw error}finally{report.effects=effects();await runtime?.stop();await proxy?.stop();report.cleanup=runtime?.receipt.cleanup||false;report.runtimeSHA256=runtime?.receipt.binarySHA256;await fs.writeFile(path.join(dir,pass+'.json'),JSON.stringify(report,null,2))}
});
