const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const fs=require('node:fs/promises'),path=require('node:path');
const {verifyWithGo,startBombadil}=require('./support/cross-app');
const base=process.env.GOO_QUERY_URL||'http://127.0.0.1:8088';

test('133 native candidates preserve observed unknown and refused evaluation boundaries',async({page,request})=>{
 test.setTimeout(180000);
 const token=randomBytes(5).toString('hex'),environment='acceptance-'+token;
 const end=Date.now()-1000,from=end-4000,mid=from+2000;
 const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 const ns=ms=>String(BigInt(ms)*1000000n);
 const cases=[
  {service:'processing-complete',metric:'messaging.process.duration',instance:'one',state:'observed',points:[[from,mid],[mid,end]]},
  {service:'processing-gap',metric:'messaging.process.duration',instance:'one',state:'unknown',points:[[from,mid]]},
  {service:'processing-unidentified',metric:'messaging.process.duration',state:'refused',points:[[from,mid],[mid,end]],error:'identity/unit/flags mismatch'},
  {service:'custom-fractional',metric:'jobs.completed',instance:'one',state:'refused',error:'counter precision/monotonicity unsupported'}
 ];
 const resourceMetrics=cases.map(c=>({resource:{attributes:attrs({'service.name':c.service,'deployment.environment.name':environment,...(c.instance?{'service.instance.id':c.instance}:{})})},scopeMetrics:[{scope:{name:'ordinary.metrics'},metrics:[c.points?{name:c.metric,unit:'s',histogram:{aggregationTemporality:1,dataPoints:c.points.map(([a,b])=>({startTimeUnixNano:ns(a),timeUnixNano:ns(b),count:'4',bucketCounts:['1','3'],explicitBounds:[1],attributes:attrs({'messaging.system':'kafka','messaging.operation.type':'process'})}))}}:{name:c.metric,unit:'{job}',sum:{aggregationTemporality:1,isMonotonic:true,dataPoints:[{startTimeUnixNano:ns(from),timeUnixNano:ns(mid),asDouble:1.5}]}},...(c.state==='observed'?[{name:'active.messages',unit:'{message}',sum:{aggregationTemporality:1,isMonotonic:false,dataPoints:[{startTimeUnixNano:ns(from),timeUnixNano:ns(mid),asInt:'2'}]}}]:[])]}]}));
 expect((await request.post('http://127.0.0.1:4318/v1/metrics',{data:{resourceMetrics}})).ok()).toBe(true);
 for(const c of cases)await expect.poll(async()=>{const r=await request.post(base+'/api/metric-discovery',{data:{environment,service:c.service,fromMs:from,toMs:end}});return r.ok()&&(await r.json()).instruments.length},{timeout:45000}).toBe(c.state==='observed'?2:1);
 let evaluations=0,writes=0;const errors=[];
 page.on('pageerror',e=>errors.push(e.message));
 page.on('request',r=>{if(r.method()!=='POST')return;const p=new URL(r.url()).pathname;if(p==='/api/measurements')evaluations++;if(['/api/captures','/api/documents','/api/measurement-packages'].includes(p))writes++});
 await page.goto(base+'/#questions');
 const dir=path.resolve('artifacts/instrument-acceptance',token);await fs.mkdir(dir,{recursive:true});
 const outcomes=[];
 for(const c of cases){
  const question={environment,service:c.service,fromMs:from,toMs:end};
  await page.evaluate(q=>dispatchEvent(new CustomEvent('goo:discover-metrics',{detail:q})),question);
  const discovered=page.waitForResponse(r=>r.url().endsWith('/api/metric-discovery'));
  await page.locator('#metricDiscover').click();const discovery=await(await discovered).json();
  const selected=discovery.instruments.find(i=>i.metric===c.metric),candidate=structuredClone(selected.capabilityPlan);
  expect(candidate.eligible).toBe(true);expect(candidate.verification).toBe('not-run');
  const rows=page.locator('#metricInstruments .instrument-catalog-table > table > tbody > tr'),row=rows.filter({hasText:c.metric});
  if(c.state==='observed')await expect(rows.filter({hasText:'active.messages'}).getByRole('button')).toBeDisabled();
  const before=evaluations,previewed=page.waitForResponse(r=>r.url().endsWith('/api/measurement-packages/plan'));
  await row.getByRole('button',{name:'Build measurement package',exact:true}).click();
  const preview=await(await previewed).json();expect(preview.definition).toEqual(selected.definition);
  await expect(page.locator('#measureStatus')).toContainText('Question validated');expect(evaluations).toBe(before);expect(writes).toBe(0);
  const executed=page.waitForResponse(r=>r.url().endsWith('/api/measurements'));
  await page.locator('#measureRun').click();const response=await executed,result=await response.json();
  expect(response.request().postDataJSON()).toEqual({definition:selected.definition,fromMs:from,toMs:end});
  if(c.state==='refused'){
   expect(response.status()).toBe(502);expect(result.error).toContain(c.error);
   await expect(page.locator('#measureStatus')).toContainText('Measurement unavailable');
   await expect(page.locator('#measureResult')).toBeEmpty();await expect(page.locator('#measureCapture')).toBeDisabled();
  }else{
   expect(response.ok()).toBe(true);expect(result.state).toBe(c.state);expect(result.question.definition).toEqual(selected.definition);
   expect(result.histograms.series).toHaveLength(1);expect(result.histograms.series[0].state).toBe(c.state==='observed'?'ok':'unknown');
   const intervals=result.histograms.series[0].intervals;expect(intervals.map(i=>i.distribution.count)).toEqual(c.state==='observed'?['4','4']:['4']);
   await expect(page.locator('#measureStatus')).toContainText(c.state+' ·');
   const ledger=path.join(dir,c.service+'.json');await fs.writeFile(ledger,JSON.stringify({discovery,result},null,2));
   await verifyWithGo(dir,'TestQuintMetricDiscoveryLedger',{GOO_METRIC_DISCOVERY_LEDGER:ledger});
  }
  expect(evaluations).toBe(before+1);expect(writes).toBe(0);
  await expect(row).toHaveAttribute('data-verification','not-run');expect(selected.capabilityPlan).toEqual(candidate);
  await page.locator('#measureStatus').scrollIntoViewIfNeeded();
  await page.screenshot({path:'docs/journeys/screenshots/133-'+c.service+'.png',fullPage:true});
  outcomes.push({service:c.service,expected:c.state,question,discovery,status:response.status(),result});
 }
 expect(errors).toEqual([]);expect(evaluations).toBe(4);expect(writes).toBe(0);
 await fs.writeFile(path.join(dir,'outcomes.json'),JSON.stringify({outcomes,evaluations,writes},null,2));
 const specification=path.join(dir,'campaign.ts'),campaign=await fs.readFile('verification/bombadil/instrument-capabilities.ts','utf8');
 await fs.writeFile(specification,campaign.replace("const fixture={environment:'local',service:'independent-metrics'};",'const fixture='+JSON.stringify({environment,service:cases[2].service})+';'));
 const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#questions',specification,seconds:25});try{await actor.finish()}finally{await actor.stop()}
});
