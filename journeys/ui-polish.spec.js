const {openFilters}=require('./support/explore');
const { test, expect } = require('@playwright/test');
const { randomBytes } = require('node:crypto');
const path = require('node:path');
const fs = require('node:fs/promises');
const {startBombadil}=require('./support/cross-app');
const GooWorkspace=require('../ui/workspace');
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
  await page.goto(goo+'/#questions');await page.locator('#exploreEnvironment').fill(environment);await run();
  await page.locator('#exploreRows [data-inspect-record]').first().click();const details=page.locator('#exploreRows .ui-inspector');await expect(page.locator('#exploreRows .record-attributes')).toHaveCount(0);
  const viewer=details.locator('.attribute-viewer'),search=viewer.getByRole('searchbox',{name:'Search attribute keys and values'});
  await expect(viewer).toBeVisible();await expect(details.locator('goo-attribute-list')).toHaveCount(1);await expect(viewer.locator('goo-field-value')).not.toHaveCount(0);await expect(details.locator('.raw-record')).not.toHaveAttribute('open','');
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
  await expect(page.locator('#distributionContent goo-facet-list > [data-facet-item]')).toHaveCount(3);
  const facetReads=reads;await page.locator('#facetSearch').fill('no matching facet');await expect(page.locator('#distributionContent goo-result-state[state=empty]')).toContainText('totals and scale are unchanged');expect(reads).toBe(facetReads);await page.locator('#facetSearch').fill('empty');await expect(page.locator('#distributionContent .distribution-row:visible')).toHaveCount(1);await expect(page.locator('#facetStatus')).toContainText('1 of 3');await expect(page.locator('#exploreCount')).toContainText('6');
  const facetResponse=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');
  await page.getByRole('button',{name:'Filter and run attribute:geo.region equals Present empty string',exact:true}).click();const empty=await(await facetResponse).json();expect(empty.count).toBe('2');
  await expect(page.locator('#exploreFilters')).toContainText('= ""');
  await require('./support/explore').removeFilter(page,field('attribute','geo.region'),'eq');await run('#exploreRun');await page.locator('#facetSearch').fill('');
  await page.screenshot({path:path.join(frames,'29b-facets-and-explicit-empty-values.png'),fullPage:true});

  const {addFilter:add,removeFilter,matchFilters}=require('./support/explore');
  const missingDraft=await add(page,field('attribute','geo.region'),'missing');await expect(missingDraft.editor.getByLabel('Condition '+missingDraft.id+' value',{exact:true})).toBeDisabled();const missing=await run('#exploreRun');expect(missing.count).toBe('2');
  await removeFilter(page,field('attribute','geo.region'),'missing');
  await add(page,field('attribute','geo.region'),'eq','east');await add(page,field('column','service'),'eq','edge');const and=await run();expect(and.count).toBe('2');
  await matchFilters(page,'any');const or=await run();expect(or.count).toBe('3');await matchFilters(page,'all');await run();
  const numericDraft=await add(page,field('column','service'),'gte','1');await expect(page.locator('#exploreRun')).toBeDisabled();await numericDraft.editor.getByLabel('Condition '+numericDraft.id+' field',{exact:true}).fill('duration_ns');await expect(page.locator('#exploreRun')).toBeEnabled();await numericDraft.editor.getByLabel('Condition '+numericDraft.id+' operator',{exact:true}).selectOption('contains');await expect(page.locator('#exploreRun')).toBeDisabled();await numericDraft.editor.getByRole('button',{name:'Remove condition '+numericDraft.id,exact:true}).click();await run();
  await page.getByRole('tab',{name:'Records',exact:true}).click();const opener=page.locator('#exploreRows').getByRole('button',{name:'Expand trace '+trace,exact:true}).first();await opener.click();
  await expect(page.locator('#exploreContextSummary')).toContainText('6 captured spans');await expect(page.locator('#exploreContextStatus')).not.toBeVisible();
  const spanButton=page.locator('#exploreTrace .waterfall-row').first();await spanButton.click();await expect(spanButton).toHaveAttribute('aria-pressed','true');await expect(page.locator('#exploreSpanDetails goo-attribute-list')).toBeVisible();await expect(page.locator('#exploreSpanDetails goo-field-value')).not.toHaveCount(0);
  await expect(page.locator('#exploreLogStatus')).toContainText('exact trace and span ID');await expect(page.locator('#exploreLogs [data-inspect-record]')).toHaveCount(1);await page.locator('#exploreLogs [data-inspect-record]').click();const linked=page.locator('#exploreLogs .ui-inspector');await expect(linked.getByRole('button',{name:/^(Filter by|Break down by)/})).toHaveCount(0);
  await linked.getByRole('button',{name:'Close selected item details',exact:true}).click();await page.locator('#exploreSpanDetails').getByRole('searchbox',{name:'Search attribute keys and values'}).fill('SpanName');await page.locator('#exploreContext').evaluate(e=>e.scrollTop=0);
  await page.screenshot({path:path.join(frames,'29c-selected-trace-and-structured-context.png'),fullPage:true});
  await page.keyboard.press('Escape');await expect(page.locator('#exploreContext')).toBeHidden();await expect(opener).toBeFocused();

  await page.setViewportSize({width:390,height:844});await require('./support/explore').clearFilters(page);await require('./support/explore').chooseGrouping(page,null);await run();
  await page.locator('#exploreRows [data-inspect-record]').first().click();const mobile=page.locator('#exploreRows .ui-inspector');await mobile.getByRole('searchbox',{name:'Search attribute keys and values'}).fill('test.multiline');
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=window.innerWidth)).toBe(true);
  await mobile.scrollIntoViewIfNeeded();await page.screenshot({path:path.join(frames,'29d-mobile-attribute-inspection.png'),fullPage:true});
  if(process.env.GOO_INSPECTION_BOMBADIL==='1'){const dir=path.resolve('artifacts/typed-inspection',environment);await fs.mkdir(dir,{recursive:true});const question={signal:'spans',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0};const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:goo+'/'+GooWorkspace.stepLink(GooWorkspace.recordStep(question)),specification:'verification/bombadil/inspection.ts',seconds:30});try{await actor.finish();const records=(await fs.readFile(path.join(dir,'bombadil/trace.jsonl'),'utf8')).trim().split('\n').map(JSON.parse),states=records.flatMap(r=>r.snapshots||[]).map(s=>s.value).filter(s=>s?.version==='goo.inspection.v2');expect(states.some(s=>s.waterfallStage==='complete')).toBe(true)}finally{await actor.stop()}}

});
