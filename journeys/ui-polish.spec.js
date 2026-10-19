const {openFilters}=require('./support/explore');
const { test, expect } = require('@playwright/test');
const { randomBytes } = require('node:crypto');
const path = require('node:path');
const fs = require('node:fs/promises');
const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const field = (kind,key) => ({kind,key});
test.use({actionTimeout:15000});

test('29 inspect structured attributes, refine facets and navigate with keyboard and mobile', async ({page,request}) => {
  const environment='polish-'+randomBytes(6).toString('hex'),trace=randomBytes(16).toString('hex'),now=Date.now();
  const multiline='  first line\nsecond\tline 雪  ',long='long-value '+('雪 <tag> '.repeat(70)),unsafe='<img src=x onerror="globalThis.polishInjected=true">';
  const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
  const spans=Array.from({length:6},(_,i)=>({traceId:trace,spanId:String(i+1).padStart(16,'0'),parentSpanId:i?'0000000000000001':undefined,
    name:'polish captured span '+i,kind:2,startTimeUnixNano:String(BigInt(now-2000+i)*1000000n),endTimeUnixNano:String(BigInt(now-2000+i)*1000000n+128n),
    attributes:attrs({'same.key':'span-value','test.counter':'9007199254740993','test.multiline':multiline,'test.long':long,'test.unsafe':unsafe,...(i<4?{'geo.region':i<2?'east':''}:{})})}));
  const resources=[0,1].map(i=>({resource:{attributes:attrs({'service.name':i?'worker':'edge','deployment.environment.name':environment,'same.key':'resource-value'})},scopeSpans:[{scope:{name:'goo.documentation.polish'},spans:spans.slice(i*3,i*3+3)}]}));
  expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:resources}})).ok()).toBe(true);
  expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource:resources[0].resource,scopeLogs:[{scope:{name:'goo.documentation.polish'},logRecords:[{timeUnixNano:String(BigInt(now-1500)*1000000n),traceId:trace,spanId:'0000000000000001',severityText:'INFO',body:{stringValue:'polish captured log'},attributes:attrs({'same.key':'log-value','test.multiline':multiline})}]}]}]}})).ok()).toBe(true);
  await expect.poll(async()=>{const r=await request.post(goo+'/api/explore',{data:{signal:'spans',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0}});return r.ok()&&(await r.json()).count},{timeout:45000,intervals:[1000]}).toBe('6');
  const frames=path.resolve('docs/journeys/screenshots');await fs.mkdir(frames,{recursive:true});
  let reads=0;page.on('request',r=>{if(r.url().endsWith('/api/explore'))reads++});
  const run=async(selector='#exploreRun')=>{const response=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await page.locator(selector).click();const r=await response;expect(r.ok()).toBe(true);return r.json()};
  await page.goto(goo+'/#explore');await page.locator('#exploreEnvironment').fill(environment);await run();
  const details=page.locator('#exploreRows .record-attributes').first();await details.locator(':scope > summary').click();
  const viewer=details.locator('.attribute-viewer'),search=viewer.getByRole('searchbox',{name:'Search attribute keys and values'});
  await expect(viewer).toBeVisible();await expect(details.locator(':scope > .raw-record')).not.toHaveAttribute('open','');
  const before=reads;await search.fill('same.key');await expect(viewer.locator('.attribute-row')).toHaveCount(2);
  await expect(viewer.getByRole('group',{name:'Resource attributes: same.key'})).toContainText('resource-value');
  await expect(viewer.getByRole('group',{name:'Span attributes: same.key'})).toContainText('span-value');
  await viewer.getByLabel('Attribute scope').selectOption('Resource attributes');await expect(viewer.locator('.attribute-row')).toHaveCount(1);
  await viewer.getByLabel('Attribute scope').selectOption('All scopes');await search.fill('test.unsafe');
  await expect(viewer.locator('.attribute-value-text')).toHaveText(unsafe);expect(await page.evaluate(()=>globalThis.polishInjected)).toBeUndefined();await expect(viewer.locator('img')).toHaveCount(0);
  await search.fill('test.multiline');expect(await viewer.locator('.attribute-value-text').textContent()).toBe(multiline);
  await page.context().grantPermissions(['clipboard-read','clipboard-write'],{origin:goo});
  await viewer.getByRole('button',{name:'Copy value for test.multiline',exact:true}).click();await expect(viewer.getByRole('status')).toContainText('Copied value');expect(await page.evaluate(()=>navigator.clipboard.readText())).toBe(multiline);
  await search.fill('test.long');await viewer.locator('.attribute-long > summary').click();expect(await viewer.locator('.attribute-value-text').textContent()).toBe(long);
  await viewer.getByRole('button',{name:'Copy value for test.long',exact:true}).click();expect(await page.evaluate(()=>navigator.clipboard.readText())).toBe(long);
  await search.fill('test.counter');await expect(viewer.locator('.attribute-value-text')).toHaveText('9007199254740993');await expect(viewer.locator('.attribute-type')).toHaveText('string');
  await search.fill('');expect(reads).toBe(before);
  await search.fill('same.key');await page.screenshot({path:path.join(frames,'29a-structured-attribute-inspection.png'),fullPage:true});

  await search.fill('geo.region');await viewer.getByRole('button',{name:'Break down and run by attribute:geo.region',exact:true}).click();
  await expect(page.locator('#distributionContent .distribution-row')).toHaveCount(3);
  await page.locator('#facetSearch').fill('empty');await expect(page.locator('#distributionContent .distribution-row:visible')).toHaveCount(1);await expect(page.locator('#facetStatus')).toContainText('1 of 3');await expect(page.locator('#exploreCount')).toContainText('6');
  const facetResponse=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');
  await page.getByRole('button',{name:'Filter and run attribute:geo.region equals Present empty string',exact:true}).click();const empty=await(await facetResponse).json();expect(empty.count).toBe('2');
  await expect(page.locator('#exploreFilters')).toContainText('Empty string');
  await page.getByRole('button',{name:'Remove filter attribute:geo.region eq',exact:true}).click();await expect(page.locator('#filterComposer > summary')).toBeFocused();await run('#exploreRun');await page.locator('#facetSearch').fill('');
  await page.screenshot({path:path.join(frames,'29b-facets-and-explicit-empty-values.png'),fullPage:true});

  const add=async(f,op,value)=>{await openFilters(page);await page.locator('#exploreFieldSearch').fill('');await openFilters(page);await page.locator('#exploreField').selectOption(JSON.stringify(f));await openFilters(page);await page.locator('#exploreOp').selectOption(op);if(!['missing','exists'].includes(op)){await openFilters(page);await page.locator('#exploreValue').fill(value)};await page.locator('#exploreAdd').click()};
  await openFilters(page);await page.locator('#exploreFieldSearch').fill('geo.region');expect(await page.locator('#exploreField option').count()).toBe(1);await openFilters(page);await page.locator('#exploreOp').selectOption('missing');await expect(page.locator('#exploreValue')).toBeDisabled();
  await page.locator('#exploreAdd').click();const missing=await run('#exploreRun');expect(missing.count).toBe('2');
  await page.getByRole('button',{name:'Remove filter attribute:geo.region missing',exact:true}).click();
  await add(field('attribute','geo.region'),'eq','east');await add(field('column','service'),'eq','edge');const and=await run();expect(and.count).toBe('2');
  await page.locator('#exploreJoin').selectOption('any');const or=await run();expect(or.count).toBe('3');
  await page.locator('#exploreJoin').selectOption('all');await run();
  await openFilters(page);await page.locator('#exploreFieldSearch').fill('');await openFilters(page);await page.locator('#exploreField').selectOption(JSON.stringify(field('column','service')));await expect(page.locator('#exploreOp option[value=gte]')).toHaveJSProperty('disabled',true);
  await openFilters(page);await page.locator('#exploreField').selectOption(JSON.stringify(field('column','duration_ns')));await expect(page.locator('#exploreOp option[value=gte]')).toHaveJSProperty('disabled',false);await expect(page.locator('#exploreOp option[value=contains]')).toHaveJSProperty('disabled',true);
  await page.getByRole('tab',{name:'Records',exact:true}).click();const opener=page.locator('#exploreRows').getByRole('button',{name:'Expand trace '+trace,exact:true}).first();await opener.click();
  await expect(page.locator('#exploreContextSummary')).toContainText('6 captured spans');await expect(page.locator('#exploreContextStatus')).not.toBeVisible();
  const spanButton=page.locator('#exploreTrace .waterfall-row').first();await spanButton.click();await expect(spanButton).toHaveAttribute('aria-pressed','true');await expect(page.locator('#exploreSpanDetails .attribute-viewer')).toBeVisible();
  await expect(page.locator('#exploreLogStatus')).toContainText('exact trace and span ID');await expect(page.locator('#exploreLogs .record-attributes')).toHaveCount(1);const linked=page.locator('#exploreLogs .record-attributes').first();await linked.locator(':scope > summary').click();await expect(linked.getByRole('button',{name:/^(Filter by|Break down by)/})).toHaveCount(0);
  await linked.locator(':scope > summary').click();await page.locator('#exploreSpanDetails').getByRole('searchbox',{name:'Search attribute keys and values'}).fill('SpanName');await page.locator('#exploreContext').evaluate(e=>e.scrollTop=0);
  await page.screenshot({path:path.join(frames,'29c-selected-trace-and-structured-context.png'),fullPage:true});
  await page.keyboard.press('Escape');await expect(page.locator('#exploreContext')).toBeHidden();await expect(opener).toBeFocused();

  await page.setViewportSize({width:390,height:844});await page.locator('#exploreClear').click();await page.locator('#exploreGroup').selectOption('');await run();
  const mobile=page.locator('#exploreRows .record-attributes').first();await mobile.locator(':scope > summary').click();await mobile.getByRole('searchbox',{name:'Search attribute keys and values'}).fill('test.multiline');
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBe(true);
  await mobile.scrollIntoViewIfNeeded();await page.screenshot({path:path.join(frames,'29d-mobile-attribute-inspection.png'),fullPage:true});
});
