const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const {startBombadil}=require('./support/cross-app');
const fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');
const base='http://127.0.0.1:8088';
test.use({actionTimeout:15000,navigationTimeout:45000});
test('220 Saved opens with compact stored rows, exact UTC, metadata paging and source-free evidence',async({page,request})=>{
 let response=await request.get(base+'/api/captures/search?order=newest');expect(response.ok()).toBe(true);let catalog=await response.json();
 if(!catalog.items.length){
  const toMs=Date.now();const created=await request.post(base+'/api/captures',{data:{title:'Saved catalog fixture',question:{kind:'explore',explore:{signal:'logs',environment:'saved-catalog-'+randomBytes(4).toString('hex'),fromMs:toMs-60000,toMs,limit:10,offset:0}}}});expect(created.status()).toBe(201);
  catalog=await(await request.get(base+'/api/captures/search?order=newest')).json();
 }
 let selected,original;
 for(const receipt of catalog.items){const r=await request.get(base+'/api/captures/'+receipt.id);if(r.ok()){selected=receipt;original=await r.json();break}}
 expect(selected,'a valid stored receipt is available').toBeTruthy();
 const source=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{const path=new URL(r.url()).pathname;if(['/api/tail','/api/explore','/api/measurements','/api/inventory','/api/discovery','/api/query','/api/operations','/api/counter-rates','/api/histograms','/api/http-package'].includes(path))source.push(path)});
 await page.goto(base+'/#investigations');const rows=page.locator('#captureCatalog tbody>tr');await expect(rows).toHaveCount(catalog.items.length);
 expect(await rows.evaluateAll(nodes=>nodes.map(n=>n.dataset.captureId))).toEqual(catalog.items.map(r=>r.id));
 expect((await rows.first().boundingBox()).y).toBeLessThan(400);
 const row=page.locator('#captureCatalog tr[data-capture-id="'+selected.id+'"]');const time=row.locator('time');
 await expect(time).toHaveAttribute('datetime',selected.capturedAt);await expect(time).toHaveAttribute('title','Captured at '+selected.capturedAt);await expect(time).toContainText('UTC');
 await time.focus();await expect(time).toBeFocused();
 await page.screenshot({path:'docs/journeys/screenshots/220a-saved-catalog-desktop.png',fullPage:false});
 const open=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/captures/'+selected.id);
 await row.getByRole('button').focus();await page.keyboard.press('Enter');expect((await open).status()).toBe(200);
 await expect(page.locator('#captureSavedTitle')).toHaveText(selected.title);expect(source).toEqual([]);
 const receipt=await page.locator('#captureGeneralReceipt').textContent();
 await page.getByRole('button',{name:'Browse evidence',exact:true}).click();
 if(catalog.hasMore){
  const next=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/captures/search'&&new URL(r.url()).searchParams.has('cursor'));
  await page.locator('#captureCatalogNext').click();const second=await(await next).json();
  await expect(rows).toHaveCount(second.items.length);expect(await rows.evaluateAll(nodes=>nodes.map(n=>n.dataset.captureId))).toEqual(second.items.map(r=>r.id));
  const previous=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/captures/search'&&!new URL(r.url()).searchParams.has('cursor'));
  await page.locator('#captureCatalogPrevious').click();await previous;await expect(rows).toHaveCount(catalog.items.length);
  expect(await rows.evaluateAll(nodes=>nodes.map(n=>n.dataset.captureId))).toEqual(catalog.items.map(r=>r.id));
 }
 expect(await page.locator('#captureGeneralReceipt').textContent()).toBe(receipt);expect(source).toEqual([]);
 expect((await(await request.get(base+'/api/captures/'+selected.id)).json()).receipt).toEqual(original.receipt);
 await page.goto(base+'/#investigations');await expect(rows).toHaveCount(catalog.items.length);
 for(const width of [320,390,700,701]){await page.setViewportSize({width,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true)}
 await page.setViewportSize({width:390,height:844});await page.locator('.saved-workspace').screenshot({path:'docs/journeys/screenshots/220b-saved-catalog-mobile.png'});
 await page.locator('#captureIDTools>summary').click();await expect(page.locator('#captureGeneralID')).toBeVisible();expect(source).toEqual([]);expect(errors).toEqual([]);
});

test('220 Bombadil browses stored rows and recovery tools without source reads',async()=>{
 test.skip(process.env.GOO_SAVED_CATALOG_BOMBADIL!=='1','Opt-in bounded stored-page campaign');
 const dir=await fs.mkdtemp(path.join(os.tmpdir(),'goo-saved-catalog-'));
 const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#investigations',specification:'verification/bombadil/saved-catalog.ts',seconds:20,width:390,height:844});
 let passed=false;try{await actor.finish();passed=true}finally{await actor.stop();if(passed)await fs.rm(dir,{recursive:true,force:true})}
});
