const {test,expect}=require('@playwright/test');
const s=require('./192-native-context-support');
const {measurementStep,stepLink}=require('../ui/workspace.mjs');
test('229 compact native actions retain exact scope and explicit read boundaries',async({page,request})=>{
 const facts=await s.seed(request),q=s.question(facts),calls=[],errors=[];
 page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/'))calls.push({path:new URL(r.url()).pathname,body:r.postDataJSON()})});page.on('pageerror',e=>errors.push(e.message));
 await page.goto(s.base+'/'+stepLink(measurementStep(q)));await expect(page.locator('#measureRun')).toBeEnabled();
 const menu=page.locator('#measureReuse'),summary=menu.locator(':scope > summary'),scope=await page.locator('#measureScope').innerHTML();
 await expect(menu).not.toHaveAttribute('open','');await expect(page.locator('#measureShare')).toBeHidden();
 for(const width of [320,390,700]){
  await page.setViewportSize({width,height:900});await page.locator('#measurementWorkbench').scrollIntoViewIfNeeded();const before=calls.length;
  await summary.focus();await summary.press('Enter');await expect(page.locator('#measureShare')).toBeVisible();await page.locator('#measureShare').focus();await page.keyboard.press('Escape');await expect(summary).toBeFocused();await expect(page.locator('#measureShare')).toBeHidden();
  expect(await page.locator('#measureScope').innerHTML()).toBe(scope);expect(calls).toHaveLength(before);expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  await page.screenshot({path:`docs/journeys/screenshots/229-native-actions-${width}.png`});
 }
 await page.setViewportSize({width:700,height:1000});await summary.click();await page.locator('#measureShare').click();await expect(page.locator('#measureQuestionLink')).toBeVisible();expect(await page.locator('#measureQuestionLink').inputValue()).toContain(q.fromNS);await expect(page.locator('#measureShare')).toBeHidden();
 let pending=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/measurements');await page.locator('#measureRun').click();const native=await s.json(await pending);expect(native.question).toEqual(q);await expect(page.locator('#measureCapture')).toBeEnabled();
 await page.setViewportSize({width:390,height:900});await summary.click();await page.locator('#measureActions').scrollIntoViewIfNeeded();await page.screenshot({path:'docs/journeys/screenshots/229-native-reuse-open.png'});await summary.press('Escape');await page.setViewportSize({width:700,height:1000});
 await page.locator('#measureResult').getByLabel('Find a stream',{exact:true}).fill('inventory');const stream=page.locator('#measureResult .native-stream-list > details');await expect(stream).toHaveCount(1);if(!await stream.evaluate(e=>e.open))await stream.locator(':scope > summary').click();await stream.locator('goo-value-timeline button[data-key="0"]').click();
 const link=stream.locator('[data-native-record-context=spans]'),held=await page.locator('#measureResult').innerHTML(),before=calls.length;await link.evaluate(e=>window.actionReturnOpener=e);await link.click();await expect(page.locator('#exploreRun')).toBeEnabled();expect(calls).toHaveLength(before);
 pending=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/explore');await page.locator('#exploreRun').click();const records=await s.json(await pending);expect(records.count).toBe('115');await page.getByRole('button',{name:'Back to measurement',exact:true}).click();expect(calls).toHaveLength(before+1);expect(await page.locator('#measureResult').innerHTML()).toBe(held);expect(await page.evaluate(()=>document.activeElement===window.actionReturnOpener)).toBe(true);expect(await page.locator('#measureScope').innerHTML()).toBe(scope);
 await page.locator('#measureCapture').click();await expect(page.locator('#captureDialog')).toBeVisible();expect(calls).toHaveLength(before+1);await page.keyboard.press('Escape');
 // Changing the input invalidates reusable references; revealing actions cannot revive them.
 await page.locator('#measureEditor > summary').click();await page.locator('#measureYAML').fill('{}');await summary.click();for(const id of ['measureShare','measureAddStep','measureStore','measureExport','measureRun','measureCapture'])await expect(page.locator('#'+id)).toBeDisabled();expect(calls).toHaveLength(before+1);expect(errors).toEqual([]);
});
