'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const F=require('./support/histogram407'),N=require('./support/native-histogram-reuse371');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {source}=require('./support/cross-app');
const {revealControl}=require('./support/explore');
const {questionLink,stepLink,measurementStep}=require('../ui/workspace.mjs');

// A presentation comparison over one persisted, real OTel bank. The existing
// 407/371 journeys remain the source authority and held-control acceptance tests.
test('411 shared histogram chrome over the same supplied observations',async({page,request})=>{
 test.setTimeout(240000);page.setDefaultTimeout(20000);
 const pass=process.env.GOO_411_PASS;if(!['before','after'].includes(pass))throw Error('Set GOO_411_PASS=before or after');
 const binary=process.env.GOO_GENERIC_BINARY;if(!binary)throw Error('Reviewed binary required');
 const dir=path.resolve('artifacts/quantitative411');await fs.mkdir(dir,{recursive:true});
 const bankPath=path.join(dir,'bank.json'),report={pass,cases:[],localEffects:[]};let runtime,proxy;
 let bank;
 if(pass==='before'){
  const native=N.fixture(),c=native.target;
  c.bounds=Array.from({length:20},(_,i)=>(i+1)/100);c.counts=Array.from({length:21},(_,i)=>String(i%4+1));c.count=c.counts.reduce((n,x)=>n+BigInt(x),0n).toString();
  Object.assign(c.points[0],{explicitBounds:c.bounds,bucketCounts:c.counts,count:c.count});Object.assign(c.points[1],{explicitBounds:c.bounds,bucketCounts:c.counts.map(()=> '0')});
  bank={numeric:F.fixture(),native};await fs.writeFile(bankPath,JSON.stringify(bank,null,2));
 }else bank=JSON.parse(await fs.readFile(bankPath,'utf8'));
 const previous=pass==='after'?JSON.parse(await fs.readFile(path.join(dir,'before.json'),'utf8')):null;
 const calls=[];page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))calls.push({path:new URL(r.url()).pathname,method:r.method()})});
 const effects=()=>({sql:proxy.calls.length,posts:calls.filter(c=>c.method==='POST').length});
 const settle=()=>page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
 const post=p=>page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===p);
 const read=async(button,p)=>{const waiting=post(p);await button.click();const response=await waiting;expect(response.ok()).toBe(true);return response.json()};
 async function examine(kind,chart,owner){
  const snapshot=await chart.evaluate(e=>e.snapshot);
  const values={rows:snapshot.rows,histogram:snapshot.histogram,selector:snapshot.scope.selector,fields:snapshot.fields};
  const localBefore=effects();
  for(const width of[1280,390]){
   await page.setViewportSize({width,height:width===390?844:1000});await chart.scrollIntoViewIfNeeded();await settle();
   const geometry=await chart.evaluate((e)=>{
    const frame=e.querySelector('goo-chart-frame'),plot=e.querySelector('goo-time-bars'),scroll=e.querySelector('.scope-histogram-scroll');
    const visible=n=>{const s=getComputedStyle(n),r=n.getBoundingClientRect();return s.display!=='none'&&s.visibility!=='hidden'&&r.width>1&&r.height>1&&s.clipPath==='none'};
    return{headings:[...e.querySelectorAll('h1,h2,h3,h4,.chart-frame-title,figcaption')].filter(visible).map(n=>n.textContent),frames:[frame,plot?.querySelector('figure')].filter(Boolean).filter(n=>parseFloat(getComputedStyle(n).borderTopWidth)>0).length,plotOffset:plot.getBoundingClientRect().top-e.getBoundingClientRect().top,usableWidth:scroll.clientWidth,scrollWidth:scroll.scrollWidth,pageOverflow:document.documentElement.scrollWidth>innerWidth+1,embedded:frame.getAttribute('appearance'),scrollLabel:scroll.getAttribute('aria-label'),tabIndex:scroll.tabIndex};
   });
   const row={kind,width,geometry,values};report.cases.push(row);
   if(pass==='after'){
    const old=previous.cases.find(c=>c.kind===kind&&c.width===width);expect(values).toEqual(old.values);expect(geometry.embedded).toBe('embedded');expect(geometry.pageOverflow).toBe(false);expect(geometry.tabIndex).toBe(0);expect(geometry.scrollLabel).toBeTruthy();expect(geometry.frames).toBeLessThan(old.geometry.frames);expect(geometry.plotOffset).toBeLessThanOrEqual(old.geometry.plotOffset);expect(geometry.usableWidth).toBeGreaterThanOrEqual(old.geometry.usableWidth);
    if(width===390){const scroll=chart.locator('.scope-histogram-scroll');expect(geometry.scrollWidth).toBeGreaterThan(geometry.usableWidth);await scroll.focus();await expect(scroll).toBeFocused();await scroll.press('ArrowRight');await expect.poll(()=>scroll.evaluate(e=>e.scrollLeft)).toBeGreaterThan(0);}
   }
   const exact=chart.locator('details').filter({has:page.getByText('Exact values',{exact:true})}).first();await exact.locator(':scope > summary').focus();await exact.locator(':scope > summary').press('Enter');await expect(exact).toHaveAttribute('open','');await expect(exact.locator('table')).toBeVisible();expect(await exact.locator('tbody tr').count()).toBe(snapshot.rows.length);await exact.locator(':scope > summary').press('Enter');await expect(exact.locator(':scope > summary')).toBeFocused();
   // Cancel any still-running native key scroll before the canonical frame.
   await expect.poll(()=>chart.locator('.scope-histogram-scroll').evaluate(e=>{e.scrollTo({left:0,behavior:'instant'});return e.scrollLeft})).toBe(0);await settle();
   await page.screenshot({path:path.join(dir,`${pass}-${kind}-${width}.png`)});
  }
  await settle();expect(effects()).toEqual(localBefore);report.localEffects.push({kind,before:localBefore,after:effects()});
 }
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(binary,dir,{clickHouseURL:proxy.url});
  if(pass==='before'){expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:F.payload(bank.numeric,'spans')})).ok()).toBe(true);expect((await request.post('http://127.0.0.1:4318/v1/metrics',{data:N.payload(bank.native)})).ok()).toBe(true)}
  const raw=()=>source(request,'SELECT ServiceName,ResourceAttributes,SpanAttributes,SpanId,TraceId,toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns FROM otel_traces WHERE Timestamp>=fromUnixTimestamp64Milli({from:Int64}) AND Timestamp<fromUnixTimestamp64Milli({to:Int64}) AND ServiceName={service:String} FORMAT JSONEachRow',{param_from:String(bank.numeric.fromMs),param_to:String(bank.numeric.toMs),param_service:'histogram-worker'});
  const ids=new Set(bank.numeric.records.filter(r=>r.signal==='spans').map(r=>r.id));await expect.poll(async()=>(await raw()).filter(r=>ids.has(r.SpanId)).length,{timeout:45000}).toBe(ids.size);report.rawNumeric=(await raw()).filter(r=>ids.has(r.SpanId)).sort((a,b)=>a.SpanId.localeCompare(b.SpanId));if(previous)expect(report.rawNumeric).toEqual(previous.rawNumeric);
  const q=F.question(bank.numeric,'spans','source');await page.goto(runtime.base+'/'+questionLink(q));F.assertResult(expect,report.rawNumeric,q,await read(page.locator('#exploreRun'),'/api/explore'));await page.locator('#resultDistribution').click();const customize=page.locator('[data-customize-distribution]');await revealControl(customize);await customize.click();const pop=page.getByRole('region',{name:'Population workspace',exact:true}).filter({visible:true});F.assertResult(expect,report.rawNumeric,q,await read(pop.getByRole('button',{name:'Run population view',exact:true}),'/api/explore'));await examine('numeric',pop.locator('[data-composition-role=histogram] goo-chart'),pop);
  // Saving is an explicit source operation; highlighting its supplied bins is local.
  const captureQuestion={...q,limit:100};const savedResponse=await request.post(runtime.base+'/api/captures',{data:{title:'411 fixed numeric bins',question:{kind:'explore',explore:captureQuestion}}});expect(savedResponse.status()).toBe(201);const saved=await savedResponse.json();F.assertResult(expect,report.rawNumeric,captureQuestion,saved.explore);
  await page.evaluate(id=>location.hash='capture='+id,saved.receipt.id);await page.locator('#savedresultDistribution').click();const savedSection=page.getByRole('region',{name:'Numeric value distribution',exact:true}).filter({visible:true});const savedChart=savedSection.locator('goo-chart');await expect(savedChart).toBeVisible();const savedValues=await savedChart.evaluate(e=>e.snapshot);const savedBefore=effects();const savedBar=savedChart.locator('goo-time-bars button').first();await savedBar.focus();await savedBar.press('Space');await expect(savedBar).toHaveAttribute('data-range-selected','true');await savedBar.press('Enter');await expect(savedBar).toHaveAttribute('data-range-selected','true');await settle();expect(await savedChart.evaluate(e=>e.snapshot)).toEqual(savedValues);await expect(savedSection.getByRole('button',{name:'Inspect selected population',exact:true})).toHaveCount(0);expect(effects()).toEqual(savedBefore);report.localEffects.push({kind:'saved-highlight',before:savedBefore,after:effects()});
  const c=bank.native.target;
  report.rawNative=await source(request,'SELECT toString(Count) AS count,arrayMap(c->toString(c),BucketCounts) AS counts,ExplicitBounds AS bounds FROM otel_metrics_histogram WHERE ServiceName={service:String} AND MetricName={metric:String} ORDER BY TimeUnix FORMAT JSONEachRow',{param_service:c.service,param_metric:c.metric});expect(report.rawNative[0]).toEqual({count:c.count,counts:c.counts,bounds:c.bounds});if(previous)expect(report.rawNative).toEqual(previous.rawNative);
  await page.goto(runtime.base+stepLink(measurementStep(c.question)));await read(page.locator('#measureRun'),'/api/measurements');const stream=page.locator('#measureResult .native-stream-list > details').first();if(!await stream.evaluate(e=>e.open))await stream.locator(':scope > summary').click();await stream.locator('goo-value-timeline button[data-key="0"]').click();const disclosure=stream.locator('details').filter({has:page.locator('summary',{hasText:'Selected native bucket counts'})});await disclosure.locator(':scope > summary').click();const nativeChart=page.locator('#measureResult .native-interval-histogram:visible');await expect(nativeChart).toBeVisible();expect(await nativeChart.evaluate(e=>e.snapshot.rows.map(r=>String(r.values.count)))).toEqual(c.counts);await examine('native',nativeChart,stream);
 }finally{await fs.writeFile(path.join(dir,pass+'.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
