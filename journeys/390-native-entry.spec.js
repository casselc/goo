const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const Fixture=require('./support/native-entry390');
const {geometry390}=require('./support/native-entry390-geometry');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {source,startBombadil}=require('./support/cross-app');
const {stepLink,measurementStep}=require('../ui/workspace.mjs');

test('390 native entry exposes supplied evidence naturally without changing source ownership',async({page,request})=>{
 test.setTimeout(300000);page.setDefaultTimeout(20000);
 const mode=process.env.GOO_NATIVE_ENTRY_390_MODE||'final',baseline=mode==='baseline',binary=process.env.GOO_GENERIC_BINARY;
 if(!binary)throw Error('Reviewed binary required');
 const dir=path.resolve('artifacts/native-entry390-'+mode);await fs.mkdir(dir,{recursive:true});
 const f=Fixture.fixture(),report={version:'goo.nativeEntry390/1',mode,fixture:f,raw:{},reads:[],captures:[],entries:[],facts:[],calls:[],errors:[]};
 let runtime,proxy;const responseTasks=[];
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(binary,dir,{clickHouseURL:proxy.url});
  report.ingestion=[];
  for(const[kind,data]of[['traces',Fixture.traces(f)],['logs',Fixture.logs(f)],['metrics',Fixture.metrics(f)]]){
   const r=await request.post('http://127.0.0.1:4318/v1/'+kind,{data});expect(r.ok()).toBe(true);report.ingestion.push({kind,status:r.status(),body:data});
  }
  const rawQuery=(table,extra)=>source(request,'SELECT ServiceName,ResourceAttributes,ResourceSchemaUrl,ScopeName,ScopeVersion,ScopeSchemaUrl,MetricName,MetricUnit,Attributes,toString(toUnixTimestamp64Nano(TimeUnix)) AS AtNS'+extra+' FROM '+table+" WHERE ResourceAttributes['deployment.environment.name']={environment:String} FORMAT JSONEachRow",{param_environment:f.environment});
  await expect.poll(async()=>{
   report.raw.sums=await rawQuery('otel_metrics_sum',",toString(toUnixTimestamp64Nano(StartTimeUnix)) AS StartNS,toString(Value) AS Value,toString(Flags) AS Flags,toString(AggregationTemporality) AS Temporality,IsMonotonic");
   report.raw.gauges=await rawQuery('otel_metrics_gauge',',toString(Value) AS Value,toString(Flags) AS Flags');
   report.raw.histograms=await rawQuery('otel_metrics_histogram',',toString(toUnixTimestamp64Nano(StartTimeUnix)) AS StartNS,toString(Count) AS Count,BucketCounts,ExplicitBounds,toString(Flags) AS Flags,toString(AggregationTemporality) AS Temporality');
   return[report.raw.sums.length,report.raw.gauges.length,report.raw.histograms.length];
  },{timeout:45000}).toEqual([3,4,1]);
  for(const[signal,table,attributes]of[['spans','otel_traces','SpanAttributes'],['logs','otel_logs','LogAttributes']]){
   await expect.poll(async()=>{report.raw[signal]=await source(request,'SELECT ServiceName,ResourceAttributes,'+attributes+' AS Attributes,TraceId,SpanId,toString(toUnixTimestamp64Nano(Timestamp)) AS AtNS FROM '+table+" WHERE ResourceAttributes['deployment.environment.name']={environment:String} FORMAT JSONEachRow",{param_environment:f.environment});return report.raw[signal].length},{timeout:45000}).toBe(f.records.length);
  }
  const capture=async(name,question)=>{
   const input={title:'390 '+name+' '+f.environment,question},r=await request.post(runtime.base+'/api/captures',{data:input});expect(r.ok(),await r.text()).toBe(true);const body=await r.json(),id=body.receipt.id,file=path.join(runtime.root,'captures','exploration',id+'.json');
   const c={name,id,request:input,status:r.status(),page:body,file,storedBytesBefore:await fs.readFile(file,'utf8')};report.captures.push(c);return c;
  };
  const saved=await capture('counter',{kind:'measurement',measurement:f.counter}),inventory=await capture('inventory',{kind:'inventory',inventory:f.inventory});
  page.on('pageerror',e=>report.errors.push(e.message));
  const callByRequest=new Map();page.on('response',r=>{const c=callByRequest.get(r.request());if(!c)return;const task=(async()=>{try{c.status=r.status();c.response=await r.json()}catch(error){c.responseUnavailable=String(error)}})();responseTasks.push(task)});
  const measurementResponses=[];page.on('response',async r=>{if(r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/measurements'){try{const response=await r.json();if(response.question?.definition?.kind==='histogram'){const row={name:'service-histogram-'+report.reads.filter(x=>x.name.startsWith('service-histogram-')).length,request:r.request().postDataJSON(),response,status:r.status()};report.reads.push(row);measurementResponses.push(row)}}catch{}}});
  page.on('request',r=>{const u=new URL(r.url());if(!u.pathname.startsWith('/api/'))return;let body;try{body=r.postDataJSON()}catch{}const c={method:r.method(),path:u.pathname,url:u.pathname+u.search,body};report.calls.push(c);callByRequest.set(r,c)});
  const sourcePaths=new Set(['/api/measurements','/api/explore','/api/explore/page','/api/discovery','/api/inventory','/api/metric-discovery','/api/record-aggregates']);
  const counters=()=>({calls:report.calls.length,source:report.calls.filter(c=>c.method==='POST'&&sourcePaths.has(c.path)).length,store:report.calls.filter(c=>c.method==='GET'&&/^\/api\/captures\/[a-f0-9]{64}$/.test(c.path)).length,sql:proxy.calls.length,writes:report.calls.filter(c=>c.method==='POST'&&['/api/captures','/api/native-views','/api/documents','/api/measurement-packages'].includes(c.path)).length});
  const settle=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
  const mark=async(name,before,sourceDelta=0,storeDelta=0)=>{await settle();const after=counters();expect(after.source-before.source,name+' source').toBe(sourceDelta);expect(after.store-before.store,name+' store').toBe(storeDelta);if(!sourceDelta)expect(after.sql,name+' SQL').toBe(before.sql);report.facts.push({name,before,after});return after};
  const fresh=async(url,width)=>{await page.goto('about:blank');await page.setViewportSize({width,height:width===390?844:1000});await page.goto(runtime.base+url);await settle()};
  const post=pathname=>page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===pathname);
  const read=async(name)=>{const waiting=post('/api/measurements');await page.locator('#measureRun').click();const r=await waiting,body=await r.json();const item={name,request:r.request().postDataJSON(),response:body,status:r.status()};report.reads.push(item);return item};
  const open=async(q,width)=>{await fresh(stepLink(measurementStep(q)),width);await expect(page.locator('#measureRun')).toBeEnabled()};
  const entry=async(surface,width,rootSelector,state,before,refs={})=>{
   await settle();const after=counters(),geometry=await page.evaluate(geometry390,{rootSelector,state});
   const item={surface,width,state,before,after,geometry,...refs};report.entries.push(item);
   if(!baseline){expect(geometry.root.visible,surface+' owner').toBe(true);expect(geometry.primary.visible,surface+' primary').toBe(true);expect(geometry.primary.intersection.height,surface+' primary height').toBeGreaterThanOrEqual(Math.min(24,geometry.primary.rect.height));expect(geometry.firstItem.intersection.centerInside,surface+' first item viewport').toBe(true);for(const item of[geometry.firstItem,...geometry.marks,...(state==='refused'?[]:[geometry.summary])]){expect(item.visible,surface+' item visible').toBe(true);expect(item.rect.top,surface+' item top').toBeGreaterThanOrEqual(item.clip.top-1);expect(item.rect.bottom,surface+' item bottom').toBeLessThanOrEqual(item.clip.bottom+1);expect(item.rect.left,surface+' item left').toBeGreaterThanOrEqual(item.clip.left-1);expect(item.rect.right,surface+' item right').toBeLessThanOrEqual(item.clip.right+1)}expect(geometry.document.scrollWidth,surface+' no horizontal page overflow').toBeLessThanOrEqual(width);if(geometry.numericPlot&&geometry.unavailablePlot){expect(Math.abs(geometry.numericPlot.left-geometry.unavailablePlot.left)).toBeLessThanOrEqual(1);expect(Math.abs(geometry.numericPlot.width-geometry.unavailablePlot.width)).toBeLessThanOrEqual(1)}}
   const frames=path.join(dir,'screenshots');await fs.mkdir(frames,{recursive:true});const frame=path.join(frames,surface+'-'+width+'.png');await page.screenshot({path:frame});const selected=(surface==='ordinary-counter'||surface==='service-histogram')&&width===1280||(surface==='ordinary-counter'||surface==='saved-counter')&&width===390||!baseline&&width===390&&['ordinary-unknown-only','ordinary-refused'].includes(surface);if(selected)await fs.copyFile(frame,'docs/journeys/screenshots/390-'+mode+'-'+surface+'-'+width+'.png');item.screenshot=frame;return item;
  };
  for(const width of[1280,390]){
   for(const[kind,q]of[['counter',f.counter],['gauge',f.gauge],['unknown-only',f.unknown],['empty',f.empty]]){
    await open(q,width);const before=counters(),result=await read('ordinary-'+kind+'-'+width);expect(result.status).toBe(200);if(kind==='empty')expect(result.response.count).toBe('0');else expect(result.response.count).toBe(kind==='unknown-only'?'1':'3');await expect(page.locator('#measureResult')).not.toBeEmpty();await mark('ordinary-'+kind+'-run-'+width,before,1);
    await entry('ordinary-'+kind,width,'#measureResult',kind==='empty'?'empty':'data',before,{readName:result.name});
    const beforeLocal=counters(),root=page.locator('#measureResult'),evidence=()=>root.evaluate(e=>({timeline:e.querySelector('goo-value-timeline')?.data||null,selectedKey:e.querySelector('goo-value-timeline')?.selectedKey||null,provenance:[...e.querySelectorAll('pre')].map(p=>p.textContent)})),evidenceBefore=await evidence(),details=root.locator('details.native-stream-context');if(await details.count()){await details.first().locator(':scope>summary').click();await details.first().locator(':scope>summary').click()}
    if(!baseline&&['counter','gauge'].includes(kind)){const save=root.locator('.native-display-save').first(),picker=root.getByLabel(kind==='counter'?'Counter display':'Gauge display',{exact:true}).first();await save.locator(':scope>summary').click();const title=save.getByLabel('Native view title',{exact:true}),initial=await title.inputValue(),original=await picker.inputValue();await picker.selectOption('table');expect(await title.inputValue()).toBe(initial);await title.fill('Authored context view');await picker.selectOption(original);expect(await title.inputValue()).toBe('Authored context view');await save.locator(':scope>summary').click();report.localTitles??=[];report.localTitles.push({surface:'ordinary-'+kind,width,initial,authored:await title.inputValue()})}
    const evidenceAfter=await evidence();expect(evidenceAfter).toEqual(evidenceBefore);report.localEvidence??=[];report.localEvidence.push({surface:'ordinary-'+kind,width,before:evidenceBefore,after:evidenceAfter});await mark('ordinary-'+kind+'-local-'+width,beforeLocal);
   }
   await open(f.counter,width);proxy.refuseAll(true);let before=counters(),failure=await read('ordinary-refused-'+width);expect(failure.status).toBe(502);await expect(page.locator('#measureStatus')).toContainText(/unavailable|refused|could not/i);await mark('ordinary-refused-run-'+width,before,1);await entry('ordinary-refused',width,'#measurementWorkbench','refused',before,{readName:failure.name});proxy.refuseAll(false);
   const retry=page.getByRole('button',{name:/Retry measurement/i});if(await retry.count()){const waiting=post('/api/measurements');before=counters();await retry.click();const r=await waiting;report.reads.push({name:'ordinary-retry-'+width,request:r.request().postDataJSON(),response:await r.json(),status:r.status()});await expect(page.locator('#measureResult')).not.toBeEmpty();await mark('ordinary-retry-'+width,before,1)}
   // Capture loading is custody-only, including source unavailable.
   proxy.refuseAll(true);before=counters();await fresh('/#capture='+saved.id,width);await expect(page.locator('#captureMeasurementEvidence')).toBeVisible();await expect(page.locator('#captureMeasurementEvidence goo-value-timeline')).toBeVisible();await entry('saved-counter',width,'#captureMeasurementEvidence','data',before,{captureName:saved.name});expect(counters().source).toBe(before.source);expect(counters().sql).toBe(before.sql);await mark('saved-counter-open-'+width,before,0,1);proxy.refuseAll(false);
   const serviceLoadBefore=counters();await fresh('/#capture='+inventory.id,width);await expect(page.locator('#captureGeneralLive')).toBeEnabled();let waiting=post('/api/discovery');await page.locator('#captureGeneralLive').click();await waiting;waiting=post('/api/dashboards/plan');await page.getByRole('button',{name:'Open service inventory / native-worker',exact:true}).click();report.servicePlan=await(await waiting).json();await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state',/loaded|partial/);await expect(page.locator('.service-dashboard [aria-busy=true]')).toHaveCount(0);
   await expect.poll(()=>measurementResponses.length).toBe(width===1280?1:2);const expectedPaths=['/api/discovery','/api/discovery','/api/explore','/api/metric-discovery','/api/inventory',...report.servicePlan.widgets.filter(w=>w.query||w.measurement).map(w=>w.query?'/api/record-aggregates':'/api/measurements')],sourceCalls=report.calls.slice(serviceLoadBefore.calls).filter(c=>c.method==='POST'&&sourcePaths.has(c.path));expect(sourceCalls.map(c=>c.path)).toEqual(expectedPaths);report.serviceLoads??=[];report.serviceLoads.push({width,fromCall:serviceLoadBefore.calls,toCall:report.calls.length,expectedPaths,plan:report.servicePlan});await mark('service-load-'+width,serviceLoadBefore,expectedPaths.length,1);
   const host=page.locator('.service-dashboard [data-widget]').filter({has:page.locator('.dashboard-native-retained')}).first(),retained=host.locator('.dashboard-native-retained');await expect(retained).toBeAttached();
   before=counters();report.serviceBeforeExpand??=[];report.serviceBeforeExpand.push(await retained.evaluate(e=>({rect:e.getBoundingClientRect().toJSON(),scrollY,open:e.open})));
   if(!await retained.evaluate(e=>e.open))await retained.locator(':scope>summary').click();await expect(retained.locator('goo-value-timeline')).toBeVisible();
   const nativeResponse=await host.locator('pre').evaluateAll(es=>es.map(e=>e.textContent));report.serviceEvidence??=[];report.serviceEvidence.push({width,texts:nativeResponse});
   await entry('service-histogram',width,'.dashboard-native-retained','data',before,{readName:measurementResponses.at(-1).name});await mark('service-expand-local-'+width,before);
  }
  for(const c of report.captures){c.storedBytesAfter=await fs.readFile(c.file,'utf8');expect(c.storedBytesAfter).toBe(c.storedBytesBefore)}
  await mark('complete',counters());expect(report.entries).toHaveLength(14);expect(report.errors).toEqual([]);report.counters=counters();
  if(!baseline&&process.env.GOO_NATIVE_ENTRY_390_BOMBADIL==='1'){
   const url=new URL(runtime.base+stepLink(measurementStep(f.unknown)));url.searchParams.set('expected',JSON.stringify({question:f.unknown,atNS:String(BigInt(f.toNS)+1500000000n)}));
   const sqlBefore=proxy.calls.length,actor=await startBombadil({dir:path.join(dir,'bombadil'),url:String(url),specification:'verification/bombadil/native-entry390.ts',seconds:40,width:1280,height:1000});
   try{const campaign=await actor.finish({snapshot:{name:'entry',version:'goo.nativeEntry390Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete');expect(complete,'Positive native entry selection required').toBeTruthy();for(const key of['observed','current','paint','stateVisible','selected','actionsPresent','budgetSafe'])expect(complete[key],key).toBe(true);expect(complete.source).toBe(1);expect(complete.store).toBe(0);expect(complete.writes).toBe(0);expect(campaign.states.every(s=>s.budgetSafe)).toBe(true);report.actor={snapshots:campaign.states.length,phases:[...new Set(campaign.states.map(s=>s.phase))],complete,run:campaign.run,sqlBefore,sqlAfter:proxy.calls.length};expect(report.actor.sqlAfter-report.actor.sqlBefore).toBe(1)}finally{await actor.stop()}
  }
  for(const c of report.captures){c.storedBytesAfter=await fs.readFile(c.file,'utf8');expect(c.storedBytesAfter).toBe(c.storedBytesBefore)}
 }finally{if(typeof responseTasks!=='undefined')await Promise.allSettled(responseTasks);report.proxyCalls=proxy?.calls;await fs.writeFile(path.join(dir,'summary.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
