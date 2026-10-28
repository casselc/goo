const {test,expect}=require('@playwright/test');const base='http://127.0.0.1:8088';test.use({actionTimeout:15000,navigationTimeout:45000});
test('223 only App contracts exposes HTTP lookback and retains the chosen next-read default',async({page})=>{
 const sources=[],errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(['/api/http-package','/api/tail','/api/explore','/api/measurements','/api/operations'].includes(new URL(r.url()).pathname))sources.push(r)});
 await page.goto(base+'/#definitions');const scope=page.locator('.scope-control'),minutes=page.locator('#minutes');await expect(scope).toBeHidden();
 await page.locator('#definitionsMeasurements').click();await expect(scope).toBeHidden();await page.locator('#definitionsContracts').click();await expect(scope).toBeVisible();await expect(minutes).toHaveAccessibleName('HTTP preview lookback minutes');await minutes.fill('7');
 await page.locator('#definitionsRecords').click();await expect(scope).toBeHidden();await page.locator('[data-workspace=services]').click();await expect(scope).toBeVisible();await expect(minutes).toHaveValue('7');await expect(minutes).toHaveAccessibleName('New discovery lookback minutes');
 await page.locator('[data-workspace=definitions]').click();await expect(scope).toBeHidden();await page.locator('#definitionsContracts').click();await expect(minutes).toHaveValue('7');expect(sources).toEqual([]);
 const response=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/http-package');await page.locator('#runHttpPackage').click();expect((await response).status()).toBe(200);expect(new URL(sources.at(-1).url()).searchParams.get('minutes')).toBe('7');await expect(page.locator('#runHttpPackage')).toBeEnabled();
 await page.locator('[data-workspace=investigations]').click();await expect(scope).toBeHidden();await page.locator('[data-workspace=definitions]').click();await expect(scope).toBeVisible();await expect(minutes).toHaveValue('7');
 await page.locator('#definitionsMeasurements').click();await page.setViewportSize({width:390,height:844});await page.screenshot({path:'docs/journeys/screenshots/223-definitions-contextual-scope.png'});expect(errors).toEqual([]);
});
test('223 editing the HTTP default does not alter a workflow fixed question',async({page})=>{
 await page.goto(base+'/#examples');await page.locator('#guidedPackageWorkflow').selectOption('red');const planResponse=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/telemetry-packages/plan');await page.locator('#guidedPackageSteps [data-workflow-package=checkout-spans]').click();const plan=await(await planResponse).json();
 await expect(page.locator('#packageUse')).toBeEnabled();const original=await page.locator('#packagePlan').textContent();await expect(page.locator('.scope-control')).toBeHidden();
 await page.locator('#definitionsContracts').click();await page.locator('#minutes').fill('9');await page.locator('#definitionsRecords').click();expect(await page.locator('#packagePlan').textContent()).toBe(original);
 const preview=page.waitForResponse(r=>new URL(r.url()).pathname==='/api/explore'&&r.request().method()==='POST');await page.locator('#packagePreview').click();const response=await preview;expect(response.status()).toBe(200);expect(response.request().postDataJSON()).toEqual(plan.question);
 await expect(page.locator('#packagePreviewResult')).toContainText('matching captured');
});
