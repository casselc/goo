const {test,expect}=require('@playwright/test');
const {startBombadil}=require('./support/cross-app');
const path=require('node:path'),os=require('node:os'),fs=require('node:fs/promises');
const base='http://127.0.0.1:8088';
const sourcePaths=new Set(['/api/operations','/api/counter-rates','/api/histograms','/api/http-package']);

async function holdSource(page,endpoint){
 let release,entered,finished;
 const gate=new Promise(r=>release=r),started=new Promise(r=>entered=r),done=new Promise(r=>finished=r);
 const pattern='**'+endpoint+'?*';let held;
 await page.route(pattern,async route=>{
  const response=await route.fetch();held={status:response.status(),id:route.request().headers()['x-goo-query-id']};
  entered();await gate;
  try{await route.fulfill({response})}catch{}finally{finished()}
 });
 return {started,done,release,held:()=>held,dispose:()=>page.unroute(pattern)};
}
const evidence=page=>page.locator('#operationView,#counterRateView,#histogramView,#httpPackageView').evaluateAll(nodes=>nodes.map(n=>n.innerHTML));

test('217 legacy owners stop every delayed source phase on leave and can explicitly read again',async({page})=>{
 const reads=[],cancels=[],errors=[];
 page.on('request',r=>{const p=new URL(r.url()).pathname;if(sourcePaths.has(p))reads.push(p);if(/^\/api\/queries\/[^/]+\/cancel$/.test(p))cancels.push(p)});
 page.on('pageerror',e=>errors.push(e.message));
 await page.goto(base+'/#operations');await page.locator('#minutes').fill('5');
 for(const endpoint of ['/api/operations','/api/counter-rates','/api/histograms','/api/http-package']){
  const http=endpoint==='/api/http-package';
  await page.locator('[data-workspace='+(http?'definitions':'operations')+']').click();
  const hold=await holdSource(page,endpoint);
  try{
   await page.locator(http?'#runHttpPackage':'#runOperations').click();await hold.started;
   expect(hold.held().status).toBe(200);expect(hold.held().id).toMatch(/^[a-f0-9-]{36}$/);
   const before=await evidence(page);
   await page.locator('[data-workspace=services]').click();
   await expect.poll(()=>cancels.includes('/api/queries/'+hold.held().id+'/cancel')).toBe(true);
   const count=reads.length;hold.release();await hold.done;await page.waitForTimeout(350);
   expect(reads).toHaveLength(count);expect(await evidence(page)).toEqual(before);
   await expect(page.locator(http?'#httpPackageStatus':'#operationStatus')).toContainText('Read stopped.');
  }finally{hold.release();await hold.dispose()}
 }
 await page.locator('[data-workspace=operations]').click();
 const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/histograms'&&r.request().method()==='GET');
 await page.locator('#runOperations').click();expect((await response).status()).toBe(200);
 await expect(page.locator('#runOperations')).toBeEnabled();await expect(page.locator('#operationMode')).toContainText('Captured operation query');
 expect(errors).toEqual([]);
});

test('217 scope changes and a hidden document abort pending legacy reads without automatic restart',async({page})=>{
 const reads=[],cancels=[];
 page.on('request',r=>{const p=new URL(r.url()).pathname;if(sourcePaths.has(p))reads.push(p);if(p.endsWith('/cancel'))cancels.push(p)});
 await page.goto(base+'/#operations');await page.locator('#minutes').fill('5');
 for(const reason of ['scope','hidden']){
  const hold=await holdSource(page,'/api/operations');
  try{
   await page.locator('#runOperations').click();await hold.started;const before=await evidence(page);
   if(reason==='scope')await page.locator('#minutes').fill('4');
   else await page.evaluate(()=>{Object.defineProperty(document,'visibilityState',{configurable:true,value:'hidden'});document.dispatchEvent(new Event('visibilitychange'))});
   await expect.poll(()=>cancels.includes('/api/queries/'+hold.held().id+'/cancel')).toBe(true);
   const count=reads.length;hold.release();await hold.done;await page.waitForTimeout(350);
   expect(reads).toHaveLength(count);expect(await evidence(page)).toEqual(before);await expect(page.locator('#runOperations')).toBeEnabled();
   if(reason==='hidden')await page.evaluate(()=>{Object.defineProperty(document,'visibilityState',{configurable:true,value:'visible'});document.dispatchEvent(new Event('visibilitychange'))});
   await page.waitForTimeout(100);expect(reads).toHaveLength(count);
  }finally{hold.release();await hold.dispose()}
 }
});

test('217 Bombadil explores legacy source ownership without saves',async()=>{
 test.skip(process.env.GOO_LEGACY_OWNER_BOMBADIL!=='1','Opt-in bounded browser campaign');
 const dir=await fs.mkdtemp(path.join(os.tmpdir(),'goo-legacy-owner-'));
 const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#operations',specification:'verification/bombadil/legacy-query-ownership.ts',seconds:20});
 let passed=false;try{await actor.finish();passed=true}finally{await actor.stop();if(passed)await fs.rm(dir,{recursive:true,force:true})}
});
