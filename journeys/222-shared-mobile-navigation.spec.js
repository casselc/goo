const {test,expect}=require('@playwright/test');const base='http://127.0.0.1:8088';
test.use({actionTimeout:15000,navigationTimeout:45000});
for(const width of [320,390,700])test('222 shared mobile navigation at '+width,async({page})=>{
 await page.setViewportSize({width,height:844});const errors=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(base+'/#definitions');
 const menu=page.getByRole('button',{name:'Menu',exact:true}),nav=page.getByRole('navigation',{name:'Main navigation'});
 await expect(menu).toHaveCount(1);await expect(menu).toBeVisible();await expect(nav).toBeHidden();await expect(menu).toHaveAttribute('aria-expanded','false');
 expect(await page.locator('.app-sidebar').evaluate(el=>el.getBoundingClientRect().height)).toBeLessThanOrEqual(65);
 await page.locator('#packageLibrarySearch').fill('keep this filter');
 let calls=0;page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))calls++});
 await menu.focus();await page.keyboard.press('Enter');await expect(nav).toBeVisible();await page.keyboard.press('Tab');await expect(page.locator('[data-workspace=explore]')).toBeFocused();await expect(menu).toHaveAttribute('aria-expanded','true');
 await page.locator('[data-workspace=services]').focus();await page.keyboard.press('Escape');await expect(menu).toBeFocused();await expect(nav).toBeHidden();expect(calls).toBe(0);
 for(const route of ['services','investigations','definitions']){
  await menu.click();await page.locator('[data-workspace='+route+']').click();await expect(nav).toBeHidden();await expect(page.locator('#workspaceTitle')).toBeFocused();await expect(menu).toHaveAttribute('aria-expanded','false');await expect(page.locator('[data-workspace-panel='+route+']')).toBeVisible();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
 }
 await expect(page.locator('#packageLibrarySearch')).toHaveValue('keep this filter');await expect(menu).toHaveCount(1);
 if(width===390){await page.screenshot({path:'docs/journeys/screenshots/222a-shared-mobile-definitions.png'});await menu.click();await page.screenshot({path:'docs/journeys/screenshots/222b-shared-mobile-menu.png'});await page.keyboard.press('Escape')}
 await page.setViewportSize({width:1024,height:844});await expect(menu).toBeHidden();await expect(nav).toBeVisible();expect(errors).toEqual([]);
});
test('222 mobile workspace Bombadil uses the shared menu',async()=>{
 test.skip(process.env.GOO_MOBILE_NAVIGATION_BOMBADIL!=='1','Opt-in bounded mobile campaign');const {startBombadil}=require('./support/cross-app'),fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');const dir=await fs.mkdtemp(path.join(os.tmpdir(),'goo-mobile-navigation-'));const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#definitions',specification:'verification/bombadil/workspace.ts',seconds:25,width:390,height:844});let passed=false;try{await actor.finish();passed=true}finally{await actor.stop();if(passed)await fs.rm(dir,{recursive:true,force:true})}
});
