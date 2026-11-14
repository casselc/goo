const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const {randomBytes,createHash}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {startBombadil}=require('./support/cross-app');

// Exporter-shaped telemetry lives only in a uniquely owned database. No lab table
// is replaced, and stream hashes arrange a fixture, never decide eligibility.
test('433 service instrument readiness refuses incomplete instance identities before planning',async({page,request})=>{
 test.setTimeout(240000);
 if(!process.env.GOO_GENERIC_BINARY)throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
 const token=randomBytes(6).toString('hex'),database='goo433_'+token,directory=path.resolve('artifacts/instrument-preflight433',token);
 if(!/^goo433_[a-f0-9]{12}$/.test(database))throw Error('Fixture ownership mismatch');
 await fs.mkdir(directory,{recursive:true});
 const service='ordinary-preflight-'+token,environment='preflight-'+token,namespace='team',scope='ordinary.preflight',report={version:'goo.instrumentPreflight433/1',database,calls:[],errors:[]};
 const base=new URL(process.env.GOO_TEST_CLICKHOUSE_URL||'http://127.0.0.1:8123'),auth='Basic '+Buffer.from((process.env.GOO_TEST_CLICKHOUSE_USER||'lab')+':'+(process.env.GOO_TEST_CLICKHOUSE_PASSWORD||'lab-local')).toString('base64');
 const sql=async text=>{const url=new URL(base);url.searchParams.set('date_time_input_format','best_effort');const r=await request.post(url.href,{data:text,headers:{Authorization:auth}});expect(r.ok(),'Fixture SQL HTTP '+r.status()).toBe(true);return r.text()};
 const resource={'deployment.environment.name':environment,'service.namespace':namespace,'telemetry.sdk.language':'python'},sorted=o=>Object.fromEntries(Object.entries(o).sort(([a],[b])=>a<b?-1:a>b?1:0));
 const hash=(r,a)=>createHash('sha256').update(JSON.stringify({resource:sorted({...r,'service.name':service}),scope,version:'1',schemaURL:'',name:'mixed.gauge',unit:'1',attributes:sorted(a)})).digest('hex');
 const missing=Array.from({length:256},(_,i)=>({resource,attributes:{partition:'missing-'+i}})).sort((a,b)=>hash(a.resource,a.attributes).localeCompare(hash(b.resource,b.attributes))).at(-1);
 const mixed=[];for(let i=0;mixed.length<26&&i<10000;i++){const r={...resource,'service.instance.id':'instance-'+i},a={partition:'valid-'+i};if(hash(r,a)<hash(missing.resource,missing.attributes))mixed.push({resource:r,attributes:a})}expect(mixed).toHaveLength(26);mixed.push(missing);
 const identities=[{metric:'valid.gauge',resource:{...resource,'service.instance.id':'valid-instance'},attributes:{}},{metric:'missing.gauge',resource,attributes:{}},{metric:'empty.gauge',resource:{...resource,'service.instance.id':''},attributes:{}},...mixed.map(x=>({...x,metric:'mixed.gauge'}))];
 let runtime,created=false;const count=p=>report.calls.filter(c=>c.path===p).length,post=p=>page.waitForResponse(r=>new URL(r.url()).pathname===p&&r.request().method()==='POST');
 page.on('pageerror',e=>report.errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.startsWith('/api/'))report.calls.push({path:new URL(r.url()).pathname,body:r.postDataJSON()})});
 try{
  await sql('CREATE DATABASE '+database);created=true;await sql(`CREATE TABLE ${database}.otel_metrics_gauge AS default.otel_metrics_gauge`);
  const now=Date.now(),rows=identities.flatMap(x=>[20000,10000].map((age,i)=>({TimeUnix:new Date(now-age).toISOString(),StartTimeUnix:new Date(now-30000).toISOString(),ServiceName:service,MetricName:x.metric,MetricUnit:'1',Value:7+i,ScopeName:scope,ScopeVersion:'1',ResourceAttributes:x.resource,Attributes:x.attributes})));
  await sql(`INSERT INTO ${database}.otel_metrics_gauge FORMAT JSONEachRow\n`+rows.map(JSON.stringify).join('\n'));
  const endpoint=new URL(base);endpoint.searchParams.set('database',database);runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,directory,{clickHouseURL:endpoint.href});
  const scopeResponse=await request.post(runtime.base+'/api/metric-scopes',{data:{fromMs:now-60000,toMs:now}});expect(scopeResponse.ok(),await scopeResponse.text()).toBe(true);report.scopes=await scopeResponse.json();expect(report.scopes.choices).toContainEqual(expect.objectContaining({service,namespace:{mode:'value',value:namespace},environment:{mode:'named',value:environment},points:'60',selectable:true}));
  await page.goto(runtime.base+'/');await page.locator('#exploreMeasurementsMode').click();await page.locator('#metricDiscovery > summary').click();await page.locator('#metricEnvironment').fill(environment);await page.locator('#metricService').fill(service);
  let pending=post('/api/metric-discovery');await page.locator('#metricDiscover').click();let response=await pending;expect(response.ok(),await response.text()).toBe(true);report.discovery=await response.json();
  expect(report.discovery.instruments).toHaveLength(4);
  for(const i of report.discovery.instruments){expect(i.records).toBe(i.metric==='mixed.gauge'?54:2);expect(i.capabilityPlan.eligible).toBe(i.metric==='valid.gauge');if(i.metric!=='valid.gauge')expect(i.capabilityPlan.reasons.join(' ')).toMatch(/instance/i)}
  const mixedResult=report.discovery.instruments.find(i=>i.metric==='mixed.gauge'),coverage=mixedResult.capabilityPlan.coverage;expect(coverage.totalStreams).toBe(27);expect(coverage.omittedStreams).toBe(2);expect(coverage.streams).toHaveLength(25);expect(coverage.streams.every(s=>s.instance==='present')).toBe(true);
  const row=metric=>page.locator('#metricInstruments tr[data-instrument="'+metric+'"]');
  for(const metric of ['missing.gauge','empty.gauge','mixed.gauge']){const r=row(metric),disabled=r.getByRole('button',{name:'Build measurement package',exact:true});await expect(disabled).toBeDisabled();await expect(r).toContainText(/instance/i);await disabled.evaluate(e=>e.click())}
  expect(count('/api/measurement-packages/plan')).toBe(0);expect(count('/api/measurements')).toBe(0);
  await row('mixed.gauge').getByText('Metadata and support',{exact:true}).click();
  for(const width of [1440,390,320]){await page.setViewportSize({width,height:1000});await row('mixed.gauge').scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:`docs/journeys/screenshots/instrument-preflight-${width}.png`})}
  await page.setViewportSize({width:1440,height:1000});
  // A retained callback from a closed disclosure cannot stage a package.
  await row('valid.gauge').getByRole('button',{name:'Build measurement package',exact:true}).evaluate(e=>window.__preflightRetired=e);
  await page.locator('#metricDiscovery > summary').click();await page.evaluate(()=>window.__preflightRetired.click());await page.waitForTimeout(100);expect(count('/api/measurement-packages/plan')).toBe(0);
  await page.locator('#metricDiscovery > summary').click();pending=post('/api/measurement-packages/plan');await row('valid.gauge').getByRole('button',{name:'Build measurement package',exact:true}).click();response=await pending;expect(response.ok(),await response.text()).toBe(true);report.plan=await response.json();await expect(page.locator('#measureStatus')).toContainText('Question validated');expect(count('/api/measurements')).toBe(0);
  pending=post('/api/measurements');await page.locator('#measureRun').click();response=await pending;report.result=await response.json();expect(response.ok(),JSON.stringify(report.result)).toBe(true);expect(report.result.count).toBe('2');expect(report.result.gauges.series).toHaveLength(1);
  // The same admission policy is visible through discovered service detail.
  await page.goto(runtime.base+'/#services');await page.locator('#inventoryEnvironment').fill(environment);const disclosure=page.locator('details').filter({has:page.locator('#inventoryWindow')});if(!await disclosure.evaluate(e=>e.open))await disclosure.locator(':scope>summary').click();await page.locator('#inventoryWindow').selectOption('fixed');await page.locator('#inventoryFromUTC').fill(new Date(now-60000).toISOString());await page.locator('#inventoryToUTC').fill(new Date(now).toISOString());await page.locator('#inventoryNamespace').selectOption('value');await page.locator('#inventoryNamespaceValue').fill(namespace);
  pending=post('/api/discovery');await page.locator('#inventoryRun').click();response=await pending;expect(response.ok(),await response.text()).toBe(true);report.services=await response.json();
  await page.locator('#serviceCatalog').getByRole('button',{name:'Open service '+namespace+' / '+service,exact:true}).click();const detail=page.locator('#serviceDetail');await expect(detail.locator('.service-loading')).toHaveCount(0);const native=detail.locator('details.service-support').filter({has:page.locator('summary',{hasText:'Native instruments'})});if(!await native.evaluate(e=>e.open))await native.locator(':scope>summary').click();
  await expect(detail.getByRole('button',{name:'Preview measurement',exact:true}).and(page.locator(':disabled'))).toHaveCount(3);await expect(detail.getByRole('button',{name:'Preview measurement',exact:true}).and(page.locator(':enabled'))).toHaveCount(1);
  await expect(detail.locator('.service-dashboard .dashboard-widget[data-widget^="native-"][data-state="observed"]')).toHaveCount(1);report.nativeWidgetReads=report.calls.filter(c=>c.path==='/api/measurements').map(c=>c.body.definition.source.metric);expect(report.nativeWidgetReads).toEqual(['valid.gauge','valid.gauge']);
  const mixedCard=detail.locator('article.definition-card').filter({has:page.getByRole('heading',{name:'mixed.gauge',exact:true})});
  const cardMetadata=mixedCard.locator(':scope > details');
  await expect(cardMetadata).toHaveCount(1);expect(await cardMetadata.evaluate(e=>e.open)).toBe(false);
  for(const width of [1440,390,320]){await page.setViewportSize({width,height:1000});await mixedCard.scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await expect(mixedCard.getByRole('button',{name:'Preview measurement',exact:true})).toBeVisible();await expect(mixedCard).toContainText(/instance/i);await page.screenshot({path:`docs/journeys/screenshots/instrument-preflight-services-${width}.png`})}
  const beforeInspect=report.calls.length;await cardMetadata.locator(':scope>summary').click();await expect(cardMetadata).toHaveJSProperty('open',true);await cardMetadata.locator(':scope>summary').click();expect(report.calls.length).toBe(beforeInspect);await page.setViewportSize({width:1440,height:1000});
  const beforePlan=count('/api/measurement-packages/plan'),beforeRun=count('/api/measurements');await detail.getByRole('button',{name:'Preview measurement',exact:true}).and(page.locator(':disabled')).evaluateAll(elements=>elements.forEach(e=>e.click()));expect(count('/api/measurement-packages/plan')).toBe(beforePlan);expect(count('/api/measurements')).toBe(beforeRun);
  pending=post('/api/measurement-packages/plan');await detail.getByRole('button',{name:'Preview measurement',exact:true}).and(page.locator(':enabled')).click();response=await pending;expect(response.ok(),await response.text()).toBe(true);await expect(page.locator('#measureStatus')).toContainText('Question validated');expect(count('/api/measurements')).toBe(beforeRun);
  if(process.env.GOO_INSTRUMENT_PREFLIGHT433_BOMBADIL==='1'){const url=new URL(runtime.base+'/');url.searchParams.set('preflight433',JSON.stringify({environment,service}));const actor=await startBombadil({dir:path.join(directory,'bombadil'),url:String(url),specification:'verification/bombadil/instrument-preflight433.ts',seconds:30,width:390,height:1000});try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.instrumentPreflight433Actor/1'}});report.bombadil=campaign.states.find(s=>s.phase==='complete');expect(report.bombadil).toBeTruthy();expect(report.bombadil.safe).toBe(true)}finally{await actor.stop()}}
  expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;await page.screenshot({path:path.join(directory,'failure.png')}).catch(()=>{});throw error}finally{try{await page.goto('about:blank',{timeout:5000}).catch(()=>{})}finally{try{await runtime?.stop()}finally{if(created){await sql('DROP DATABASE '+database+' SYNC');report.cleaned=true}await fs.writeFile(path.join(directory,'report.json'),JSON.stringify(report,null,2))}}}
});
