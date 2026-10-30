const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');
const {randomBytes}=require('node:crypto'),{execFile}=require('node:child_process'),{promisify}=require('node:util');
const {startRuntime}=require('./191-generic-runtime'),{source}=require('./support/cross-app');

test('267 trace inspection uses captured membership independently of the refinement budget',async({page,request})=>{
 const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-trace-inspect-267-')),dir=path.resolve('artifacts/scoped-267-trace');let runtime;
 try{
  const binary=path.join(owned,'goo');await promisify(execFile)('go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});runtime=await startRuntime(binary,dir);
  const token=randomBytes(4).toString('hex'),environment='inspect267-'+token,from=Date.now()-10000,to=from+5000,attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
  const ids=[randomBytes(16).toString('hex'),randomBytes(16).toString('hex')],services=['gateway-'+token,'worker-'+token];
  for(const [i,name]of services.entries()){
   const spans=ids.map((traceId,j)=>{const at=BigInt(from+1000+j*1000+i*100)*1000000n;return {traceId,spanId:randomBytes(8).toString('hex'),name:'trace inspection '+i,kind:2,startTimeUnixNano:String(at),endTimeUnixNano:String(at+1000000n),attributes:attrs({'run.id':token})}});
   expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource:{attributes:attrs({'service.name':name,'deployment.environment.name':environment})},scopeSpans:[{spans}]}]}})).ok()).toBe(true);
  }
  await expect.poll(async()=>Number((await source(request,"SELECT count() AS n FROM default.otel_traces WHERE ResourceAttributes['deployment.environment.name']={env:String} FORMAT JSONEachRow",{param_env:environment}))[0].n),{timeout:45000}).toBe(4);
  const condition={field:{kind:'attribute',key:'run.id'},op:'eq',value:token},q={signal:'spans',environment,fromMs:from,toMs:to,limit:100,offset:0,where:{all:Array.from({length:31},()=>structuredClone(condition))},groupBy:[{kind:'column',key:'trace_id'}]},reads=[],errors=[];
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname==='/api/explore')reads.push(r.postDataJSON())});page.on('pageerror',e=>errors.push(e.message));await page.goto(runtime.base);await page.evaluate(async q=>{location.hash=(await import('/question-protocol.mjs')).questionLink(q)},q);await page.locator('#exploreRun').click();
  const region=page.getByRole('region',{name:'Scoped breakdown',exact:true}),chart=region.locator('goo-chart');await expect(chart.locator('.scope-view-bar')).toHaveCount(2);await chart.locator('.scope-view-bar').first().click();expect(reads).toHaveLength(1);
  await region.getByRole('button',{name:'Stage selected group',exact:true}).click();await expect(page.locator('#exploreStatus')).toContainText('Cannot stage selection:');await expect(page.locator('#exploration')).toHaveAttribute('data-question-phase','applied');expect(reads).toHaveLength(1);
  await region.getByRole('button',{name:'Open selected trace',exact:true}).click();await expect(page.locator('#exploreContext')).toBeVisible();await expect(page.locator('#exploreTrace')).toContainText(services[0]);await expect(page.locator('#exploreTrace')).toContainText(services[1]);await expect.poll(()=>reads.length).toBe(3);
  expect(reads[1].fromMs).toBe(q.fromMs);expect(reads[1].toMs).toBe(q.toMs);expect(reads[1].environment).toBe(environment);expect(reads[1].where.all).toHaveLength(1);expect(reads[1].where.all[0].field).toEqual({kind:'column',key:'trace_id'});expect(ids).toContain(reads[1].where.all[0].value);expect(reads[2].signal).toBe('logs');expect(reads[2].where).toEqual(reads[1].where);expect(reads[2].fromMs).toBe(q.fromMs);expect(reads[2].toMs).toBe(q.toMs);expect(reads[2].environment).toBe(environment);expect(errors).toEqual([]);
  await page.screenshot({path:'docs/journeys/screenshots/267-trace-inspection.png'});await fs.writeFile(path.join(dir,'ledger.json'),JSON.stringify({parentConditions:31,stageBudgetRefused:true,inspectionReads:2,totalExploreRequests:3,sourceServices:2,capturedWindow:true,errors}));
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
