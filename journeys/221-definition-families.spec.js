const {test,expect}=require('@playwright/test');
const base='http://127.0.0.1:8088';
test.use({actionTimeout:15000,navigationTimeout:45000});
test('221 families preserve drafts, keyboard selection and workspace return without source reads',async({page,request})=>{
 const reads=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&/\/api\/(tail|explore|measurements|query)$/.test(new URL(r.url()).pathname))reads.push(r.url())});
 await page.goto(base+'/#definitions');const now=Date.now(),draft=await(await request.post(base+'/api/telemetry-packages/draft',{data:{question:{signal:'logs',environment:'local',fromMs:now-60000,toMs:now,limit:100,offset:0},name:'family-draft',title:'Family draft',lookbackMinutes:1}})).json();await page.locator('#packageFile').setInputFiles({name:'family.yaml',mimeType:'application/yaml',buffer:Buffer.from(draft.yaml)});await expect(page.locator('#packageYAML')).toBeVisible();
 await page.locator('#packageYAML').fill('# private unsaved draft\n');await page.locator('#packageLibrarySearch').fill('retained-record-filter');
 await page.locator('#definitionsRecords').focus();await page.keyboard.press('ArrowRight');await expect(page.locator('#definitionsMeasurements')).toBeFocused();
 await page.locator('#measurementDefinitionSearch').fill('retained-native-filter');await expect(page.locator('#packageYAML')).toBeHidden();
 await page.keyboard.press('Tab');await page.locator('#definitionsMeasurements').focus();await page.keyboard.press('End');await expect(page.locator('#definitionsContracts')).toBeFocused();await expect(page.locator('#httpPackage')).toBeVisible();
 await page.locator('[data-workspace=services]').click();await page.locator('[data-workspace=definitions]').click();await expect(page.locator('#definitionsContracts')).toHaveAttribute('aria-selected','true');
 await page.locator('#definitionsContracts').focus();await page.keyboard.press('Home');await expect(page.locator('#packageYAML')).toHaveValue('# private unsaved draft\n');await expect(page.locator('#packageLibrarySearch')).toHaveValue('retained-record-filter');
 await page.locator('#definitionsMeasurements').click();await expect(page.locator('#measurementDefinitionSearch')).toHaveValue('retained-native-filter');
 for(const width of [320,390,700]){await page.setViewportSize({width,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true)}
 await page.locator('#definitionsRecords').click();await page.locator('[aria-label="Definition families"]').screenshot({path:'docs/journeys/screenshots/221-definition-families-mobile.png'});
 expect(reads).toEqual([]);expect(errors).toEqual([]);
});
test('221 leaving App contracts cancels its source read within Definitions',async({page})=>{
 let release,started;const gate=new Promise(r=>release=r),ready=new Promise(r=>started=r);let id;const cancels=[];
 page.on('request',r=>{if(new URL(r.url()).pathname.endsWith('/cancel'))cancels.push(new URL(r.url()).pathname)});
 await page.goto(base+'/#definitions');await page.locator('#definitionsContracts').click();
 await page.route('**/api/http-package?*',async route=>{id=route.request().headers()['x-goo-query-id'];const response=await route.fetch();started();await gate;await route.fulfill({response})});
 try{await page.locator('#runHttpPackage').click();await ready;expect(id).toMatch(/^[a-f0-9-]{36}$/);await page.locator('#definitionsRecords').click();await expect.poll(()=>cancels.includes('/api/queries/'+id+'/cancel')).toBe(true);release();await page.unrouteAll({behavior:'wait'});await expect(page.locator('#httpPackageView')).toBeEmpty();await expect(page.locator('#httpPackageStatus')).toContainText('Read stopped.');
 await page.locator('#definitionsContracts').click();const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/http-package');await page.locator('#runHttpPackage').click();expect((await response).status()).toBe(200);await expect(page.locator('#httpPackageView table')).toBeVisible();
 }finally{release();await page.unrouteAll({behavior:'wait'})}
});
test('221 Bombadil explores family switching and definition inspection',async()=>{
 test.skip(process.env.GOO_DEFINITION_FAMILIES_BOMBADIL!=='1','Opt-in bounded browser campaign');
 const {startBombadil}=require('./support/cross-app'),fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');const dir=await fs.mkdtemp(path.join(os.tmpdir(),'goo-definition-families-'));const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#definitions',specification:'verification/bombadil/definition-library.ts',seconds:25});let passed=false;try{await actor.finish();passed=true}finally{await actor.stop();if(passed)await fs.rm(dir,{recursive:true,force:true})}
});
