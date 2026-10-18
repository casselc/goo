const {test,expect}=require('@playwright/test');
const {execFileSync}=require('node:child_process');
const {startBombadil,verifyWithGo}=require('./support/cross-app');
const fs=require('node:fs/promises');const path=require('node:path');
test('39 discover ordinary OTel metrics and build a reusable measurement',async({page,request})=>{
 execFileSync('/workspace/.venvs/goo/bin/python',['scripts/independent_metrics.py']);
 const base='http://127.0.0.1:8088';
 await expect.poll(async()=>{const toMs=Date.now();const r=await request.post(base+'/api/metric-discovery',{data:{environment:'local',service:'independent-metrics',fromMs:toMs-120000,toMs}});if(!r.ok())return 0;return (await r.json()).records},{timeout:45000}).toBeGreaterThan(0);
 await page.goto(base+'/#explore');await page.evaluate(()=>dispatchEvent(new CustomEvent('goo:measurements-tab')));
 await page.locator('#metricDiscovery summary').click();await page.locator('#metricService').fill('independent-metrics');const discovered=page.waitForResponse(r=>r.url().endsWith('/api/metric-discovery'));await page.locator('#metricDiscover').click();const discovery=await (await discovered).json();
 await expect(page.locator('#metricInstruments')).toContainText('jobs.completed');await expect(page.locator('#metricInstruments')).toContainText('version omitted');
 await page.screenshot({path:'docs/journeys/screenshots/39a-metric-discovery.png',fullPage:true});
 await page.getByRole('button',{name:'Build measurement package',exact:true}).click();await expect(page.locator('#measureStatus')).toContainText('Question validated');
 await page.locator('#measureStore').click();await expect(page.locator('#measureStatus')).toContainText('Saved immutable');
 const response=page.waitForResponse(r=>r.url().endsWith('/api/measurements'));await page.locator('#measureRun').click();const result=await (await response).json();expect(result.counters.series).toHaveLength(2);expect(result.counters.series.every(s=>s.stream.version===''&&s.stream.schemaURL==='')).toBe(true);
 await expect(page.locator('#measureResult')).toContainText('operation=import');await expect(page.locator('#measureResult')).toContainText('outcome=error');
 await page.screenshot({path:'docs/journeys/screenshots/39b-discovered-measurement.png',fullPage:true});
 await page.locator('#measureCapture').click();await page.locator('#captureDraftTitle').fill('Independent OTel metric discovery');const saved=page.waitForResponse(r=>r.url().endsWith('/api/captures')&&r.request().method()==='POST');await page.locator('#captureDraftConfirm').click();expect((await saved).ok()).toBe(true);await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();await page.reload();await expect(page.locator('#captureMeasurementEvidence')).toContainText('jobs.completed');
 await page.locator('#captureGeneralNotes').click();await page.locator('#documentTitle').fill('Independent metric investigation');const document=page.waitForResponse(r=>r.url().endsWith('/api/documents')&&r.request().method()==='POST');await page.locator('#documentSave').click();expect((await document).ok()).toBe(true);
 const dir=path.resolve('artifacts/metric-discovery');await fs.mkdir(dir,{recursive:true});const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({discovery,result}));await verifyWithGo(dir,'TestQuintMetricDiscoveryLedger',{GOO_METRIC_DISCOVERY_LEDGER:ledger});
 if(process.env.GOO_METRIC_DISCOVERY_BOMBADIL==='1'){const actor=await startBombadil({dir:path.join(dir,'bombadil-'+require('node:crypto').randomUUID()),url:base+'/#explore',specification:'verification/bombadil/metric-discovery.ts',seconds:30});try{await actor.finish()}finally{await actor.stop()}}
});
