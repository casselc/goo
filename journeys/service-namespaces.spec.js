const {test,expect}=require('@playwright/test');
const {randomBytes}=require('crypto');
const {readStepLink}=require('../ui/workspace');
const fs=require('node:fs/promises'),path=require('node:path');
const {verifyWithGo,startBombadil}=require('./support/cross-app');
const base='http://127.0.0.1:8088';
test('60 distinguish namespace identities and explore declared operation bindings',async({page,request})=>{
 test.setTimeout(180000);
 const now=Date.now(),environment='namespaces-'+randomBytes(4).toString('hex'),attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}})),trace=randomBytes(16).toString('hex'),at=BigInt(now-10000)*1000000n;
 const resource=(namespace,extra={})=>({attributes:attrs({'service.name':'api','service.instance.id':'one','deployment.environment.name':environment,...(namespace===null?{}:{'service.namespace':namespace}),...extra})});
 const span=(i,parent='')=>({traceId:trace,spanId:i.toString(16).padStart(16,'0'),parentSpanId:parent,name:'GET /items',kind:2,status:{code:1},startTimeUnixNano:String(at),endTimeUnixNano:String(at+10000000n)});
 const namespaces=[null,'','payments','orders'];
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:namespaces.map((n,i)=>({resource:resource(n),scopeSpans:[{scope:{name:'ordinary.sdk',version:'1'},spans:[span(i+1,i===3?'0000000000000003':'')]}]}))}})).ok()).toBe(true);
 expect((await request.post('http://127.0.0.1:4318/v1/metrics',{data:{resourceMetrics:namespaces.map((n,i)=>({resource:resource(n),scopeMetrics:[{scope:{name:'ordinary.runtime',version:'1'},metrics:[{name:'process.memory.utilization',unit:'1',gauge:{dataPoints:[{timeUnixNano:String(at),asDouble:(i+1)/10}]}}]}]}))}})).ok()).toBe(true);
 const post=async(url,data)=>{const r=await request.post(base+url,{data});expect(r.ok(),await r.text()).toBe(true);return r.json()};
 const scope={environment,fromMs:now-60000,toMs:now};
 await expect.poll(async()=>(await post('/api/discovery',scope)).signals.length,{timeout:45000}).toBe(8);
 const discovery=await post('/api/discovery',scope);expect(discovery.dependencies).toContainEqual(expect.objectContaining({parent:'api',parentNamespace:'payments',child:'api',childNamespace:'orders',kind:'edge',count:'1'}));
 for(const namespace of [{mode:'missing'},{mode:'value'},{mode:'value',value:'payments'},{mode:'value',value:'orders'}]){
  const inventory=await post('/api/inventory',{...scope,namespace});expect(inventory.count).toBe('1');
  const native=await post('/api/metric-discovery',{...scope,namespace,service:'api'});expect(native.instruments).toHaveLength(1);expect(native.instruments[0].definition.source.namespace).toEqual(namespace);
 }
 const errors=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(base+'/#services');await page.locator('#inventoryEnvironment').fill(environment);await page.locator('#inventoryRun').click();await expect(page.locator('#serviceCatalog tbody tr')).toHaveCount(4);
 await page.locator('#serviceCatalog').screenshot({path:'docs/journeys/screenshots/60a-namespace-service-identities.png'});
 await page.getByRole('button',{name:'Open service payments / api',exact:true}).click();const dashboard=page.locator('.service-dashboard');await expect(dashboard).toHaveAttribute('data-state','loaded');await expect(dashboard.locator('[data-widget=native-1]')).toContainText('0.3 1');
 const link=readStepLink(await dashboard.locator('[data-widget=spans]').getByRole('link',{name:'Open widget question',exact:true}).getAttribute('href'));expect(link.record.where.all).toContainEqual({field:{kind:'resource',key:'service.namespace'},op:'eq',value:'payments'});
 await dashboard.screenshot({path:'docs/journeys/screenshots/60b-namespace-service-dashboard.png'});
 for(const [label,value] of [['(empty namespace) / api','0.2 1'],['api','0.1 1']]){await page.locator('#serviceDetail').getByRole('button',{name:'Back to service list',exact:true}).click();await page.getByRole('button',{name:'Open service '+label,exact:true}).click();await expect(dashboard).toHaveAttribute('data-state','loaded');await expect(dashboard.locator('[data-widget=native-1]')).toContainText(value);await expect(dashboard.locator('[data-widget=spans]')).toHaveAttribute('data-state','observed');}
 // Registered bindings generate ordinary questions over captured operation span records.
 const packages=await(await request.get(base+'/api/operation-packages')).json(),bound=packages[0],d=bound.definition,n='contract-'+randomBytes(4).toString('hex'),op=d.operations[0];
 const bindingResource={attributes:attrs({'service.name':d.resource.service,'service.instance.id':d.resource.instance,'deployment.environment.name':d.resource.environment,'service.namespace':n})};
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource:bindingResource,scopeSpans:[{scope:{name:d.instrumentation.name,version:d.instrumentation.version},spans:[{...span(9),attributes:attrs({'goo.operation.name':op.name,'goo.operation.role':op.role})},{...span(10),attributes:attrs({'goo.operation.name':op.name,'goo.operation.role':'wrong'})}]}]}]}})).ok()).toBe(true);
 const contractScope={...scope,environment:d.resource.environment,namespace:{mode:'value',value:n}};
 await expect.poll(async()=>(await post('/api/inventory',contractScope)).count,{timeout:45000}).toBe('2');
 const plan=await post('/api/dashboards/plan',{scope:contractScope,service:d.resource.service,signals:['spans'],fields:[],instruments:[]}),widget=plan.widgets.find(w=>w.key==='contract-1');expect(widget.state).toBe('supported');expect(widget.reason).toContain('Schema URL is declared metadata');const result=await post('/api/explore',widget.query);expect(result.count).toBe('1');await page.goto(base+'/#services');await page.locator('#inventoryEnvironment').fill(d.resource.environment);await page.locator('#inventoryRun').click();await page.getByRole('button',{name:'Open service '+n+' / '+d.resource.service,exact:true}).click();await expect(page.locator('[data-widget=contract-1]')).toHaveAttribute('data-state','observed');await page.locator('.service-dashboard').screenshot({path:'docs/journeys/screenshots/60c-declared-operation-question.png'});await page.setViewportSize({width:390,height:844});await expect.poll(()=>page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.locator('.service-dashboard').screenshot({path:'docs/journeys/screenshots/60d-mobile-declared-operation-question.png'});expect(errors).toEqual([]);const dir=path.resolve('artifacts/service-namespaces',String(Date.now()));await fs.mkdir(dir,{recursive:true});const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({discovery,link,plan,result},null,2));await verifyWithGo(dir,'TestQuintNamespaceLive',{GOO_NAMESPACE_LEDGER:ledger});if(process.env.GOO_NAMESPACE_BOMBADIL==='1'){const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:base+'/#'+new URLSearchParams({dashboard:environment}),specification:'verification/bombadil/service-namespace.ts',seconds:30});try{await actor.finish()}finally{await actor.stop()}}
});
