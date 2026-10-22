const {test,expect}=require('@playwright/test');
const {questionLink}=require('../ui/workspace');
const base='http://127.0.0.1:8088';
test('109 mount fixed authoring explicitly and retain its draft without a read',async({page})=>{
 test.setTimeout(90000);const reads=[],errors=[];page.on('request',r=>{if(r.method()==='POST'&&/\/api\/(explore|population\/plan|captures|retained-evidence)$/.test(r.url()))reads.push(r.url())});page.on('pageerror',e=>errors.push(e.message));
 await page.goto(base+'/');await expect(page.locator('.stream-entry .live-tail')).toHaveAttribute('data-state',/unknown|partial/);await expect(page.locator('#exploreRun')).toHaveCount(0);
 await page.locator('.entry-advanced > summary').click();await expect(page.locator('#exploreRun')).toHaveCount(1);await expect(page.locator('.stream-entry')).toBeHidden();await page.locator('#exploreEnvironment').fill('draft-kept');const editor=await page.locator('#exploreRun').elementHandle();await page.getByRole('button',{name:'Return to paused activity',exact:true}).click();await expect(page.locator('#exploreRun')).toHaveCount(0);await expect(page.locator('.stream-entry .live-tail')).toHaveAttribute('data-state','paused');
 await page.locator('.entry-advanced > summary').click();await expect(page.locator('#exploreEnvironment')).toHaveValue('draft-kept');expect(await editor.evaluate(e=>e===document.getElementById('exploreRun'))).toBe(true);await expect(page.locator('#exploreRun')).toHaveCount(1);expect(reads).toEqual([]);await page.screenshot({path:'docs/journeys/screenshots/109a-explicit-fixed-authoring.png'});
 const now=Date.now(),q={signal:'logs',environment:'local',fromMs:now-60000,toMs:now,limit:100,offset:0};await page.goto(base+'/'+questionLink(q));await expect(page.locator('#exploreRun')).toHaveCount(1);await expect(page.locator('#exploration')).toHaveAttribute('data-question-phase','draft');await expect(page.locator('.stream-entry')).toBeHidden();expect(reads).toEqual([]);expect(errors).toEqual([]);
});
