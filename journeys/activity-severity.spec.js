const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path'),{randomBytes}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {source,startBombadil}=require('./support/cross-app');

const bands=['trace','debug','info','warn','error','fatal','unspecified','invalid'];
const bandOf=n=>n===0?'unspecified':n>24?'invalid':bands[Math.floor((n-1)/4)];
const attrs=object=>Object.entries(object).map(([key,stringValue])=>({key,value:{stringValue}}));
const numericRange={all:[{field:{kind:'column',key:'severity_number'},op:'gte',value:'17'},{field:{kind:'column',key:'severity_number'},op:'lte',value:'20'}]};

test('435 Activity shows complete numeric severity independently of text and stages exact bands',async({page,request})=>{
 test.setTimeout(240000);page.setDefaultTimeout(15000);
 if(!process.env.GOO_GENERIC_BINARY)throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
 const token=randomBytes(6).toString('hex'),dir=path.resolve('artifacts/activity-severity435',token);
 await fs.mkdir(dir,{recursive:true});
 const report={version:'goo.activitySeverity435/1',token,calls:[],steps:[],errors:[]},runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir);
 const telemetry=new Set(['/api/tail','/api/record-aggregates','/api/explore','/api/explore/page','/api/population/explain','/api/explore/counted-values','/api/measurements','/api/inventory','/api/discovery','/api/source-capabilities','/api/environment-scopes','/api/environment-scopes/records','/api/metric-scopes','/api/metric-discovery']);
 const reads=()=>report.calls.filter(c=>c.method==='POST'&&telemetry.has(c.path)).length,step=(name,before)=>report.steps.push({name,sourceReads:reads()-before});
 page.on('pageerror',e=>report.errors.push(e.message));
 page.on('request',req=>{const pathname=new URL(req.url()).pathname;if(pathname.startsWith('/api/')){let body;try{body=req.postDataJSON()}catch{}report.calls.push({path:pathname,method:req.method(),body});}});
 async function emit(environment,bank,numbers,{age=1000}={}){
  const now=Date.now();
  const logs=numbers.map((severityNumber,i)=>({timeUnixNano:String(BigInt(now-age-(i%3)*10000)*1000000n),severityNumber,severityText:severityNumber===9?'FATAL':severityNumber===21?'info':'producer-label',body:{stringValue:'ordinary severity '+bank+' '+i},traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex')}));
  const response=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource:{attributes:attrs({'service.name':'ordinary-severity435','deployment.environment.name':environment,'severity435.bank':bank})},scopeLogs:[{scope:{name:'ordinary.severity'},logRecords:logs}]}]}});
  expect(response.ok(),await response.text()).toBe(true);
 }
 async function saved(environment,bank,title){
  const selector={signal:'logs',environment,where:{any:[{field:{kind:'resource',key:'severity435.bank'},op:'eq',value:bank},{field:{kind:'column',key:'service'},op:'eq',value:'never-'+bank}]}};
  const response=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title,selector,columns:['utc','service','summary']}});expect(response.status()).toBe(201);return {view:await response.json(),selector};
 }
 try{
  const environment='severity435-'+token,bank=token+'-main',numbers=[0,1,4,5,8,9,12,13,16,17,20,21,24,25,255,...Array(110).fill(17)];
  await emit(environment,bank,numbers);await emit(environment,'outside-'+bank,[17]);await emit(environment,bank,[17],{age:5*60000});await emit(environment+'-outside',bank,[17]);
  await expect.poll(async()=>Number((await source(request,"SELECT count() AS n FROM default.otel_logs WHERE ResourceAttributes['severity435.bank']={bank:String} FORMAT JSONEachRow",{param_bank:bank}))[0].n),{timeout:45000}).toBe(numbers.length+2);
  const {view,selector}=await saved(environment,bank,'Numeric severity · ordinary OTLP');report.selector=selector;
  await page.goto(runtime.base+'/#activity-view='+view.id);
  const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),volume=shell.getByRole('region',{name:'Activity log volume',exact:true});
  let before=reads();await shell.getByRole('button',{name:'Start stream',exact:true}).click();await expect(tail.locator('[data-tail-key]')).toHaveCount(100);await shell.getByRole('button',{name:'Pause stream',exact:true}).click();step('loaded',before);report.loaded={retained:Number(await tail.getAttribute('data-retained')),limited:await tail.getAttribute('data-gap')==='true'};
  before=reads();const pending=page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/record-aggregates');await shell.getByRole('button',{name:'Show volume',exact:true}).click();const response=await pending;expect(response.ok(),await response.text()).toBe(true);const aggregate=await response.json();await expect(volume).toHaveAttribute('data-phase','ready');step('volume',before);
  const query=report.calls.filter(c=>c.path==='/api/record-aggregates'&&c.method==='POST').at(-1).body;report.aggregate={question:query,result:aggregate};expect(query.logSeverity).toBe(true);expect(query.where).toEqual(selector.where);expect(aggregate.count).toBe(String(numbers.length));expect(report.loaded.retained).toBeLessThan(Number(aggregate.count));
  report.raw=await source(request,"SELECT toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns, SeverityNumber, SeverityText, Body, ServiceName, ResourceAttributes, LogAttributes FROM default.otel_logs WHERE ResourceAttributes['severity435.bank']={bank:String} AND ResourceAttributes['deployment.environment.name']={environment:String} AND Timestamp >= fromUnixTimestamp64Milli({from:Int64}) AND Timestamp < fromUnixTimestamp64Milli({to:Int64}) ORDER BY Timestamp, Body FORMAT JSONEachRow",{param_bank:bank,param_environment:environment,param_from:query.fromMs,param_to:query.toMs,max_result_rows:1000});
  expect(report.raw).toHaveLength(numbers.length);
  report.heldRecords=await tail.locator('goo-record-summary').evaluateAll(nodes=>nodes.map(node=>node.record));expect(report.heldRecords).toHaveLength(100);expect(report.heldRecords.every(row=>row.SeverityNumber!==undefined)).toBe(true);
  const expected=Object.fromEntries(bands.map(band=>[band,'0']));for(const row of report.raw)expected[bandOf(Number(row.SeverityNumber))]=String(BigInt(expected[bandOf(Number(row.SeverityNumber))])+1n);expect(aggregate.logSeverity).toEqual({count:String(numbers.length),...expected});
  const misleading=aggregate.groups.find(g=>g.g0==='FATAL');expect(misleading.logSeverity.info).toBe('1');expect(misleading.logSeverity.fatal).toBe('0');report.expected=expected;
  const chooser=volume.getByRole('combobox',{name:'Numeric severity',exact:true}),chart=volume.locator('goo-time-bars'),all=await chart.evaluate(e=>e.data);report.allChart=all;
  for(const band of bands){
   before=reads();await chooser.selectOption(band);await expect(volume).toHaveAttribute('data-severity-band',band);const shown=await chart.evaluate(e=>e.data);expect(shown.ceiling).toBe(all.ceiling);expect(shown.points.map(p=>p.value)).toEqual(aggregate.timeSeries.map(p=>p.logSeverity[band]));expect(shown.points.reduce((n,p)=>n+BigInt(p.value),0n).toString()).toBe(expected[band]);expect(reads()).toBe(before);report.steps.push({name:'band',sourceReads:reads()-before,band});
  }
  await chooser.selectOption('error');report.selectedBand={band:'error',points:(await chart.evaluate(e=>e.data)).points.map(p=>p.value),ceiling:(await chart.evaluate(e=>e.data)).ceiling};
  const filter=volume.getByRole('button',{name:'Filter Error',exact:true});
  await filter.evaluate(e=>window.__severity435Filter={element:e,handler:e.onclick});before=reads();await chooser.selectOption('info');await page.evaluate(()=>window.__severity435Filter.handler.call(window.__severity435Filter.element));await expect(shell.getByRole('button',{name:'Apply scope',exact:true})).toBeHidden();expect(reads()).toBe(before);step('stale_refused',before);await chooser.selectOption('error');
  for(const width of [1440,390,320]){await page.setViewportSize({width,height:900});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await volume.scrollIntoViewIfNeeded();await page.screenshot({path:`docs/journeys/screenshots/activity-severity-${width}.png`});}
  await page.setViewportSize({width:1440,height:1000});
  // A supplied band and time bucket define the explanation's own exact scope.
  const bucket=aggregate.timeSeries.findIndex(point=>point.logSeverity.error!=='0');expect(bucket).toBeGreaterThanOrEqual(0);
  before=reads();await chart.locator('.time-bar').nth(bucket).focus();await page.keyboard.press(' ');await page.keyboard.press('Enter');await volume.getByRole('button',{name:'Explain interval',exact:true}).click();const explanation=shell.getByRole('region',{name:'Activity explanation',exact:true});await expect(explanation).toBeVisible();await expect(explanation).toHaveAttribute('data-severity-band','error');expect(reads()).toBe(before);report.explanation={fromMs:Number(await explanation.getAttribute('data-selected-from-ms')),toMs:Number(await explanation.getAttribute('data-selected-to-ms'))};expect(report.explanation).toEqual({fromMs:aggregate.timeSeries[bucket].fromMs,toMs:aggregate.timeSeries[bucket].toMs});
  const heldCounts=Array(6).fill(0);for(const row of report.heldRecords){const ms=Number(BigInt(row.timestamp_ns)/1000000n);if(bandOf(Number(row.SeverityNumber))==='error'&&ms>=query.fromMs&&ms<query.toMs)heldCounts[Math.floor((ms-query.fromMs)/10000)]++;}const heldChart=await explanation.locator('goo-time-bars').evaluate(e=>e.data);expect(heldChart.points.map(p=>p.value)).toEqual(heldCounts.map(String));expect(heldCounts.reduce((a,b)=>a+b,0)).toBeGreaterThan(0);report.explanation.heldPoints=heldCounts.map(String);step('explain',before);
  await page.setViewportSize({width:390,height:900});await explanation.scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/activity-severity-explanation-390.png'});await page.setViewportSize({width:1440,height:1000});await explanation.getByRole('button',{name:'Close explanation',exact:true}).click();
  before=reads();await filter.click();await expect(shell.getByRole('button',{name:'Apply scope',exact:true})).toBeVisible();expect(reads()).toBe(before);step('filter',before);report.filterWhere={all:[selector.where,numericRange]};
  before=reads();await shell.getByRole('button',{name:'Apply scope',exact:true}).click();await expect(tail).toHaveAttribute('data-retained','100');await expect.poll(()=>report.calls.filter(c=>c.path==='/api/tail'&&c.method==='POST').length).toBe(2);report.appliedSelector=report.calls.filter(c=>c.path==='/api/tail'&&c.method==='POST').at(-1).body;expect(report.appliedSelector.where).toEqual(report.filterWhere);step('apply',before);await shell.getByRole('button',{name:'Pause stream',exact:true}).click();
  // Retained owners cannot change a replacement/dirty scope.
  await volume.getByRole('button',{name:'Refresh log volume',exact:true}).evaluate(e=>window.__severity435Refresh={element:e,handler:e.onclick});before=reads();await shell.getByLabel('Live environment',{exact:true}).fill(environment+'-dirty');await page.evaluate(()=>{window.__severity435Refresh.handler.call(window.__severity435Refresh.element);window.__severity435Filter.handler.call(window.__severity435Filter.element)});expect(reads()).toBe(before);step('retire',before);
  if(process.env.GOO_SEVERITY435_BOMBADIL==='1'){
   const actorBank=token+'-actor',actorEnvironment='severity435actor-'+token;await emit(actorEnvironment,actorBank,[0,1,5,9,13,17,21,25]);await expect.poll(async()=>Number((await source(request,"SELECT count() AS n FROM default.otel_logs WHERE ResourceAttributes['severity435.bank']={bank:String} FORMAT JSONEachRow",{param_bank:actorBank}))[0].n),{timeout:45000}).toBe(8);
   const made=await saved(actorEnvironment,actorBank,'Actor numeric severity');await page.goto('about:blank');const params=new URLSearchParams({severity435Env:actorEnvironment,severity435Bank:actorBank});const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/?'+params+'#activity-view='+made.view.id,specification:'verification/bombadil/activity-severity435.ts',seconds:30,width:390,height:900});
   try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.activitySeverity435Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete');expect(complete).toBeTruthy();expect(complete.localSafe).toBe(true);expect(complete.scopeSafe).toBe(true);expect(complete.forbidden).toEqual([]);expect([complete.tailReads,complete.aggregateReads]).toEqual([2,1]);report.bombadil={complete,snapshots:campaign.states.length};}finally{await actor.stop();}
  }
  expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;report.ui=await page.locator('.stream-entry').innerText().catch(()=>null);await page.screenshot({path:path.join(dir,'failure.png')}).catch(()=>{});throw error;}
 finally{await page.goto('about:blank').catch(()=>{});await runtime.stop();report.cleanup=runtime.receipt.cleanup;await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2));}
});
