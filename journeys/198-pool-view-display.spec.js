const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { base, json, observe, loaded } = require('./198-pool-view-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const target = page => page.locator('.pool-connection-row[data-pool=orders][data-instance=worker-a]');
async function fit(page) { await target(page).scrollIntoViewIfNeeded(); await page.evaluate(() => { const e = document.querySelector('.pool-connection-row[data-pool=orders][data-instance=worker-a]'); scrollBy(0, e.getBoundingClientRect().top - 90); }); }
async function geometry(page) {
 expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
 const result = await target(page).evaluate(e => {
  const r = e.getBoundingClientRect(), view = e.querySelector('.pool-aligned-view'), b = view.getBoundingClientRect(), table=view.querySelector('goo-table'), tb=table.getBoundingClientRect();
  return { rootRem:parseFloat(getComputedStyle(document.documentElement).fontSize),row: { x:r.left,y:r.top,right:r.right,bottom:r.bottom,scrollWidth:e.scrollWidth,clientWidth:e.clientWidth }, view:{left:b.left,right:b.right},table:{top:tb.top,bottom:tb.bottom,scrollHeight:table.scrollHeight,clientHeight:table.clientHeight},
   checks:[...e.querySelectorAll('input[type=checkbox]')].map(c=>{const r=c.getBoundingClientRect();return {width:r.width,height:r.height}}),
   cells:[...view.querySelectorAll('th[scope=row],td')].map(c=>{const r=c.getBoundingClientRect();return {text:c.textContent,left:r.left,right:r.right,top:r.top,bottom:r.bottom,scrollWidth:c.scrollWidth,clientWidth:c.clientWidth}}),
   times:[...view.querySelectorAll('td time')].map(t=>({text:t.textContent,title:t.title,dateTime:t.dateTime})),
   tracks:[...view.querySelectorAll('.comparison-bar-track')].map(t=>t.getBoundingClientRect().width),
   maxima:[...view.querySelectorAll('.comparison-bar-cell')].map(t=>t.dataset.scaleMax),
   order:[...view.querySelectorAll('th[scope=row]')].map(t=>t.textContent),provenanceOpen:view.querySelector(':scope > details').open };
 });
 const issues=[];
 if(result.row.scrollWidth>result.row.clientWidth+1)issues.push('row horizontal overflow');if(result.table.scrollHeight>result.table.clientHeight+1)issues.push('pool table vertically crops primary signals');if(result.provenanceOpen)issues.push('provenance unexpectedly open');
 for(const [i,c]of result.checks.entries())if(Math.abs(c.width-result.rootRem)>=1||Math.abs(c.height-result.rootRem)>=1||c.width>18||c.height>18)issues.push('checkbox '+i+' is not a compact root-rem square');
 for(const [i,c]of result.cells.entries())if(c.left<result.view.left-1||c.right>result.view.right+1||c.scrollWidth>c.clientWidth+1||c.top<result.table.top-1||c.bottom>result.table.bottom+1)issues.push('cell '+i+' clipped: '+c.text);
 if(JSON.stringify(result.order)!==JSON.stringify(['used','idle','reserved','Capacity','Acquisition wait · p95']))issues.push('state priority differs');
 if(result.tracks.length!==4)issues.push('count track count differs');for(const t of result.tracks)if(t<=50||Math.abs(t-result.tracks[0])>=1)issues.push('count track widths differ');
 if(JSON.stringify(result.maxima)!==JSON.stringify(['12','12','12','12']))issues.push('count scales differ');
 if(result.times.length<6)issues.push('observation/interval times missing');for(const t of result.times)if(!t.text.includes('UTC')||!t.title.includes('2026-10-')||!/[0-9]{18,}/.test(t.title))issues.push('exact UTC/ns tooltip missing');
 result.issues=issues;
 return result;
}
test('198 display replay uses accepted real fixture and original fixed scope',async({page,request})=>{
 test.skip(!process.env.GOO_POOL_VIEW_LEDGER,'Explicit retained accepted ledger required');test.setTimeout(120000);
 const ledgerPath=path.resolve(process.env.GOO_POOL_VIEW_LEDGER), archived=JSON.parse(await fs.readFile(ledgerPath,'utf8')),observation=observe(page);
 await page.goto(base+'/#services');await page.locator('#minutes').fill('2');
 let pending=page.waitForResponse(r=>r.url().endsWith('/api/discovery')&&r.request().method()==='POST');
 // Existing application API used by fixed inventory reopening; no supplied rows.
 await page.evaluate(scope=>liveInventory.openQuestion(scope),archived.initial.request.scope);const catalog=await json(await pending);
 pending=page.waitForResponse(r=>r.url().endsWith('/api/dashboards/plan'));await page.getByRole('button',{name:'Open service inventory / pool-api',exact:true}).click();const plan=await json(await pending);await loaded(page,'http-server');await observation.settle();expect(observation.source()).toBe(9);
 expect(plan.scope).toEqual(archived.initial.plan.scope);
 const count=observation.responses.find(r=>r.endpoint==='/api/measurements'&&r.body?.question.definition.source.metric==='db.client.connection.count').body;
 expect(count.question).toEqual(archived.initial.countResult.question);expect(count.stocks).toEqual(archived.initial.countResult.stocks);
 pending=page.waitForResponse(r=>r.url().endsWith('/api/pool-signals/plan'));await target(page).locator('.pool-signals > summary').click();const planner=await json(await pending);
 pending=page.waitForResponse(r=>r.url().endsWith('/api/pool-signals/view'));await target(page).locator('[data-pool-load]').click();const viewResponse=await pending,composeRequest=viewResponse.request().postDataJSON(),view=await json(viewResponse);await expect(target(page).locator('.pool-aligned-view')).toBeVisible();
 expect(observation.source()).toBe(11);expect(composeRequest.capacity.stocks).toEqual(archived.primary.request.capacity.stocks);expect(composeRequest.wait.histograms).toEqual(archived.primary.request.wait.histograms);
 expect(view.states).toEqual(archived.primary.response.states);expect(view.capacity.latest).toEqual(archived.primary.response.capacity.latest);expect(view.wait.latest).toEqual(archived.primary.response.wait.latest);expect(view.utilization).toEqual(archived.primary.response.utilization);
 await expect(target(page).locator('.pool-utilization')).toContainText('30%');
 const beforeFrames=observation.source();await page.setViewportSize({width:1440,height:1800});await fit(page);const desktop=await geometry(page);
 await page.setViewportSize({width:390,height:1800});await fit(page);const mobile=await geometry(page);
 await fs.writeFile(path.resolve('artifacts/pool-view/display-geometry-all.json'),JSON.stringify({desktop,mobile},null,2));
 expect([...desktop.issues,...mobile.issues]).toEqual([]);
 await target(page).screenshot({path:'docs/journeys/screenshots/198b-mobile-aligned-pool.png'});
 await page.setViewportSize({width:1440,height:1800});await fit(page);await target(page).screenshot({path:'docs/journeys/screenshots/198a-aligned-pool-evidence.png'});await page.screenshot({path:'docs/journeys/screenshots/198c-pool-view-in-context.png'});
 expect(observation.source()).toBe(beforeFrames);expect(observation.errors).toEqual([]);
 const id=archived.saved.receipt.id,reopened=await json(await request.get(base+'/api/captures/'+id));expect(reopened).toEqual(archived.saved);
 await fs.writeFile(path.join(path.dirname(ledgerPath),'display-replay.json'),JSON.stringify({sourceLedger:ledgerPath,source:'display-only2f8405e',newIngestion:false,newCapture:false,scope:archived.initial.request.scope,catalog,plan,count,planner,composeRequest,view,desktop,mobile,sourceReads:observation.source(),sourceFreeViewport:true,requests:observation.requests,responses:observation.responses,existingCaptureId:id},null,2));
});
