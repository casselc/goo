// Real product CSS and native dialog behavior; long content is a layout fixture.
const {chromium}=require('playwright');
const fs=require('node:fs/promises'),path=require('node:path'),assert=require('node:assert/strict');
(async()=>{let browser;try{
 browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
 const page=await browser.newPage({hasTouch:true}),css=await fs.readFile(path.resolve(__dirname,'../ui/product.css'));
 await page.route('http://trace-dialog.invalid/**',async route=>{if(new URL(route.request().url()).pathname.endsWith('.css')){await route.fulfill({contentType:'text/css',body:css});return}await route.fulfill({contentType:'text/html',body:'<!doctype html><html lang="en"><meta name="viewport" content="width=device-width"><title>Trace dialog layout</title><link rel="stylesheet" href="/product.css">'+['exploreContext','savedexploreContext'].map(id=>`<dialog id="${id}" style="left:auto;max-height:none"><div class="inspector-heading"><button type="button">Close trace</button></div><div style="height:1800px">Expanded span attributes and captured trace details</div><button type="button" data-inspect-record>Inspect linked log</button></dialog>`).join('')})});
 for(const viewport of [{width:1440,height:900},{width:390,height:844}])for(const id of ['exploreContext','savedexploreContext']){
  await page.setViewportSize(viewport);await page.goto('http://trace-dialog.invalid/');await page.locator('#'+id).evaluate(dialog=>dialog.showModal());
  const before=await page.locator('#'+id).evaluate(dialog=>{const r=dialog.getBoundingClientRect();return {top:r.top,bottom:r.bottom,client:dialog.clientHeight,scroll:dialog.scrollHeight,height:innerHeight}});
  assert(before.top>=0&&before.bottom<=before.height+.5,JSON.stringify(before));assert(before.scroll>before.client);
  const inspect=page.locator('#'+id+' [data-inspect-record]');await inspect.evaluate(button=>button.onclick=()=>button.dataset.clicked='true');await inspect.scrollIntoViewIfNeeded();await inspect.click();assert.equal(await inspect.getAttribute('data-clicked'),'true');
  const after=await inspect.evaluate(button=>{const r=button.getBoundingClientRect();return {top:r.top,bottom:r.bottom,height:innerHeight,scroll:button.closest('dialog').scrollTop}});assert(after.scroll>0);assert(after.top>=0&&after.bottom<=after.height+.5,JSON.stringify(after));
  await page.locator('#'+id).evaluate(dialog=>dialog.close());
 }
 console.log('Trace dialog: desktop/mobile live and Saved sheets stay inside viewport, long content scrolls, linked Inspect is reachable by ordinary pointer click');
}finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
