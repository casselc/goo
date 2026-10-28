const {test,expect}=require('@playwright/test');
const {startBombadil}=require('./support/cross-app');
const fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');
const base='http://127.0.0.1:8088';
test.use({actionTimeout:15000,navigationTimeout:45000});
const deferredPaths=['/api/prebuilts','/api/package','/api/operation-packages','/api/example-questions','/api/investigation-workflows'];
function observe(page){const calls=[],errors=[];page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))calls.push({method:r.method(),path:new URL(r.url()).pathname})});page.on('pageerror',e=>errors.push(e.message));return {calls,errors,count:path=>calls.filter(c=>c.path===path).length,source:()=>calls.filter(c=>c.method==='POST'&&['/api/tail','/api/explore','/api/query','/api/measurements'].includes(c.path)).length}}

test('219 Activity defers workspace catalogs, first open loads once, explicit refresh stays fresh',async({page})=>{
 const observed=observe(page);
 await page.goto(base+'/');await expect(page.locator('.stream-entry .live-tail')).toHaveAttribute('data-state',/unknown|partial/);
 await expect(page.locator('#measureLibrary option')).not.toHaveCount(1);
 for(const path of deferredPaths)expect(observed.count(path),path).toBe(0);
 expect(observed.count('/api/measurement-packages')).toBe(1);
 await expect(page.locator('#exampleQuestionCards')).toHaveCount(0);await expect(page.locator('#measurementCards')).toHaveCount(0);
 const sourceBefore=observed.source();
 await page.locator('[data-workspace=definitions]').click();
 await expect(page.locator('#definitionCatalog .definition-card')).not.toHaveCount(0);
 await expect(page.locator('#prebuilts details')).not.toHaveCount(0);
 await expect(page.locator('#package')).toContainText('{');
 for(const path of ['/api/prebuilts','/api/package','/api/operation-packages'])expect(observed.count(path),path).toBe(1);
 await page.locator('[data-workspace=services]').click();await page.locator('[data-workspace=definitions]').click();
 for(const path of ['/api/prebuilts','/api/package','/api/operation-packages'])expect(observed.count(path),path).toBe(1);
 await page.locator('[data-workspace=examples]').click();
 await expect(page.locator('#exampleQuestionCards .definition-card')).not.toHaveCount(0);
 await expect(page.locator('#measurementCards .definition-card')).not.toHaveCount(0);
 await expect(page.locator('#guidedPackageWorkflow option')).not.toHaveCount(1);
 expect(observed.count('/api/example-questions')).toBe(1);expect(observed.count('/api/investigation-workflows')).toBe(1);
 const cards=await page.locator('#exampleQuestionCards .definition-card').count();
 await page.locator('#guidedPackageWorkflow').selectOption('red');await expect(page.locator('#guidedPackageSteps')).toContainText('Shared fixed window');
 await page.locator('[data-workspace=services]').click();await page.locator('[data-workspace=examples]').click();
 await expect(page.locator('#exampleQuestionCards .definition-card')).toHaveCount(cards);await expect(page.locator('#guidedPackageWorkflow')).toHaveValue('red');
 expect(observed.count('/api/example-questions')).toBe(1);expect(observed.count('/api/investigation-workflows')).toBe(1);
 const preview=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/telemetry-packages/plan');
 await page.locator('#guidedPackageSteps [data-workflow-package=checkout-spans]').click();expect((await preview).status()).toBe(200);
 expect(observed.source()).toBe(sourceBefore);
 await page.locator('[data-workspace=definitions]').click();const before=observed.count('/api/measurement-packages');
 const refresh=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/measurement-packages');await page.locator('#measurementDefinitionRefresh').click();expect((await refresh).status()).toBe(200);
 expect(observed.count('/api/measurement-packages')).toBe(before+1);expect(observed.source()).toBe(sourceBefore);expect(observed.errors).toEqual([]);
});

for(const route of ['examples','definitions','operations'])test('219 direct '+route+' entry activates only its required deferred catalogs',async({page})=>{
 const observed=observe(page);await page.goto(base+'/#'+route);
 if(route==='examples'){
  await expect(page.locator('#exampleQuestionCards .definition-card')).not.toHaveCount(0);await expect(page.locator('#measurementCards .definition-card')).not.toHaveCount(0);
  expect(observed.count('/api/example-questions')).toBe(1);expect(observed.count('/api/operation-packages')).toBe(0);expect(observed.count('/api/prebuilts')).toBe(0);
 }else{
  await expect(page.locator('#operationPackage option')).not.toHaveCount(0);await expect(page.locator('#operationPackage')).not.toHaveValue('');
  expect(observed.count('/api/operation-packages')).toBe(1);expect(observed.count('/api/example-questions')).toBe(0);
  if(route==='definitions'){await expect(page.locator('#prebuilts details')).not.toHaveCount(0);expect(observed.count('/api/prebuilts')).toBe(1)}else expect(observed.count('/api/prebuilts')).toBe(0);
 }
 expect(observed.source()).toBe(0);expect(observed.errors).toEqual([]);
});

test('219 Bombadil revisits deferred catalog owners without duplicate loads',async()=>{
 test.skip(process.env.GOO_STARTUP_BOMBADIL!=='1','Opt-in bounded navigation campaign');
 const dir=await fs.mkdtemp(path.join(os.tmpdir(),'goo-startup-catalog-'));
 const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/',specification:'verification/bombadil/workspace.ts',seconds:20});
 let passed=false;try{await actor.finish();passed=true}finally{await actor.stop();if(passed)await fs.rm(dir,{recursive:true,force:true})}
});
