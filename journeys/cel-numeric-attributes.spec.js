'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path'),{randomBytes}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {source}=require('./support/cross-app');
const {openFilters}=require('./support/explore');
const {recordStep,stepLink}=require('../ui/question-protocol.mjs');
const attributes=values=>Object.entries(values).filter(([,v])=>v!==undefined).map(([key,stringValue])=>({key,value:{stringValue}}));
const quantity=[undefined,'','0','-1','1','9007199254740993','18446744073709551615','18446744073709551616','01','1.20'];
const price=[undefined,'','0.00','-1.20','1.20','9007199254740993.00','92233720368547758.07','92233720368547758.08','01.20','1.2'];
const numeric=(kind,key,type,op,value,unit='1',scale)=>({numeric:{field:{kind,key},type,unit,parse:'canonical-decimal',...(scale?{scale}:{})},op,...(value===undefined?{}:{value})});
const pure=new Set(['/api/cel/context','/api/cel/plan']);
const reads=new Set(['/api/explore','/api/explore/page','/api/tail','/api/discovery','/api/inventory','/api/source-capabilities','/api/record-aggregates','/api/explore/counted-values']);

test('443 exact numeric attribute CEL shares typed conditions and real source execution',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);if(!process.env.GOO_GENERIC_BINARY)throw Error('GOO_GENERIC_BINARY required');
 // The lab TTL is toDate(event time)+1 day. Near midnight place synthetic
 // fixture events just inside the next day so a run cannot lose yesterday's
 // part between its independent oracle read and its last source assertion.
 const now=Date.now(),dayEnd=(Math.floor(now/86400000)+1)*86400000,clock=dayEnd-now<120000?dayEnd+2000:now;
 const token=randomBytes(6).toString('hex'),environment='numeric443-'+token,dir=path.resolve('artifacts/cel-numeric443',token),report={version:'goo.cel-numeric443/1',token,environment,fixture:[],raw:{},contexts:[],calls:[],errors:[],consoleRefusals:[]};await fs.mkdir(dir,{recursive:true});let runtime;
 page.on('pageerror',e=>report.errors.push(e.message));page.on('console',m=>{if(m.type()!=='error')return;const location=m.location();if(m.text()==='Failed to load resource: the server responded with a status of 400 (Bad Request)'&&location.url?.endsWith('/api/cel/plan'))report.consoleRefusals.push({message:m.text(),url:location.url});else report.errors.push(m.text())});page.on('request',r=>{if(r.method()==='POST'){const p=new URL(r.url()).pathname;if(p.startsWith('/api/'))report.calls.push({path:p,body:r.postDataJSON()})}});
 const post=p=>page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===p),editor=()=>page.locator('#exploreFilterEditor'),text=()=>editor().getByRole('textbox',{name:'Additional condition (CEL)',exact:true});
 async function local(context,name,action){const before=report.calls.length,result=await action();await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));const calls=report.calls.slice(before);expect(calls.filter(c=>!pure.has(c.path)&&!/^\/api\/queries\/[^/]+\/cancel$/.test(c.path)),name+' stays source-free').toEqual([]);context.steps.push({name,sourceReads:0});return result}
 async function stage(base){await page.goto('about:blank');await page.goto(runtime.base+stepLink(recordStep(base)));await expect(page.locator('#exploreRun')).toBeVisible();await openFilters(page);await editor().getByRole('combobox',{name:'Filter mode',exact:true}).selectOption('expression')}
 async function check(expression,status=200){await text().fill(expression);const pending=post('/api/cel/plan');await editor().getByRole('button',{name:'Check condition',exact:true}).click();const response=await pending;expect(response.status(),await response.text()).toBe(status);return response.json()}
 try{
  report.fixture=quantity.map((value,i)=>({i,traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),timestamp:String(BigInt(clock-1000+i)*1000000n),resource:{'service.name':'ordinary-worker','deployment.environment.name':environment,'numeric443.bank':token,'example.quantity':quantity[quantity.length-1-i]},record:{'example.quantity':value,'example.price':price[i]}}));
  const outside=[{...report.fixture[5],spanId:randomBytes(8).toString('hex'),resource:{...report.fixture[5].resource,'deployment.environment.name':environment+'-outside'}},{...report.fixture[6],spanId:randomBytes(8).toString('hex'),resource:{...report.fixture[6].resource,'numeric443.bank':token+'-outside'}}],all=[...report.fixture,...outside];
  const logs={resourceLogs:all.map(f=>({resource:{attributes:attributes(f.resource)},scopeLogs:[{scope:{name:'ordinary.numeric'},logRecords:[{timeUnixNano:f.timestamp,severityNumber:9,body:{stringValue:'numeric case '+f.i},traceId:f.traceId,spanId:f.spanId,attributes:attributes(f.record)}]}]}))};
  const spans={resourceSpans:all.map(f=>({resource:{attributes:attributes(f.resource)},scopeSpans:[{scope:{name:'ordinary.numeric'},spans:[{name:'numeric case '+f.i,traceId:f.traceId,spanId:f.spanId,kind:2,startTimeUnixNano:f.timestamp,endTimeUnixNano:String(BigInt(f.timestamp)+1000000n),attributes:attributes(f.record),status:{code:1}}]}]}))};
  for(const [endpoint,data]of [['logs',logs],['traces',spans]]){const r=await request.post('http://127.0.0.1:4318/v1/'+endpoint,{data});expect(r.ok(),await r.text()).toBe(true)}
  for(const signal of ['logs','spans']){const table=signal==='logs'?'otel_logs':'otel_traces',column=signal==='logs'?'LogAttributes':'SpanAttributes',sql='SELECT SpanId,ResourceAttributes,'+column+" AS Attributes FROM "+table+" WHERE ResourceAttributes['numeric443.bank']={bank:String} AND ResourceAttributes['deployment.environment.name']={environment:String} FORMAT JSONEachRow",params={param_bank:token,param_environment:environment};await expect.poll(async()=>(await source(request,sql,params)).length,{timeout:45000}).toBe(quantity.length);report.raw[signal]=await source(request,sql,params);for(const f of report.fixture){const raw=report.raw[signal].find(r=>r.SpanId===f.spanId);for(const [key,value]of Object.entries(f.record)){expect(Object.hasOwn(raw.Attributes,key)).toBe(value!==undefined);if(value!==undefined)expect(raw.Attributes[key]).toBe(value)}}}
  runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,dir);
  const q='example.quantity',p='example.price',cases=[
   ['exact-uint',`record.attributes.uint64("${q}", "1").eq("9007199254740993")`,numeric('attribute',q,'uint64','eq','9007199254740993'),[5]],
   ['uint-range',`record.attributes.uint64("${q}", "1").gte("1")`,numeric('attribute',q,'uint64','gte','1'),[4,5,6]],
   ['signed-negative',`record.attributes.int64("${q}", "1").lt("0")`,numeric('attribute',q,'int64','lt','0'),[3]],
   ['valid-only-ne',`record.attributes.uint64("${q}", "1").ne("0")`,numeric('attribute',q,'uint64','ne','0'),[4,5,6]],
   ['invalid',`record.attributes.uint64("${q}", "1").invalid()`,numeric('attribute',q,'uint64','invalid'),[3,7,8,9]],
   ['missing',`record.attributes.uint64("${q}", "1").missing()`,numeric('attribute',q,'uint64','missing'),[0]],
   ['empty',`record.attributes.uint64("${q}", "1").empty()`,numeric('attribute',q,'uint64','empty'),[1]],
   ['decimal-exact',`record.attributes.decimal64("${p}", "USD", 2).eq("1.20")`,numeric('attribute',p,'decimal64','eq','1.20','USD',2),[4]],
   ['decimal-boundary',`record.attributes.decimal64("${p}", "USD", 2).gte("92233720368547758.07")`,numeric('attribute',p,'decimal64','gte','92233720368547758.07','USD',2),[6]],
   ['decimal-valid',`record.attributes.decimal64("${p}", "USD", 2).valid()`,numeric('attribute',p,'decimal64','valid',undefined,'USD',2),[2,3,4,5,6]],
   ['decimal-invalid',`record.attributes.decimal64("${p}", "USD", 2).invalid()`,numeric('attribute',p,'decimal64','invalid',undefined,'USD',2),[7,8,9]],
   ['resource-namespace',`resource.attributes.uint64("${q}", "1").eq("9007199254740993")`,numeric('resource',q,'uint64','eq','9007199254740993'),[4]],
  ];
  for(const signal of ['logs','spans']){
   const parent={field:{kind:'resource',key:'numeric443.bank'},op:'eq',value:token},base={signal,environment,fromMs:clock-60000,toMs:clock+1000,limit:100,offset:0,where:parent},context={signal,steps:[],runs:[],checks:[]};report.contexts.push(context);
   for(const [name,expression,predicate,indexes]of cases){const plan=await local(context,'check-'+name,async()=>{await stage(base);return check(expression)});expect(plan.predicate).toEqual(predicate);expect(plan.question.where).toEqual({all:[parent,predicate]});context.checks.push({name,expression,plan});const before=report.calls.filter(c=>reads.has(c.path)).length,pending=post('/api/explore');await page.locator('#exploreRun').click();const response=await pending;expect(response.ok(),await response.text()).toBe(true);const result=await response.json(),expected=indexes.map(i=>report.fixture[i].spanId).sort();expect(result.rows.map(r=>r.SpanId).sort()).toEqual(expected);expect(result.count).toBe(String(expected.length));expect(response.request().postDataJSON()).toEqual({...plan.question,groupBy:plan.question.groupBy||[]});expect(report.calls.filter(c=>reads.has(c.path)).length).toBe(before+1);context.runs.push({name,request:plan.question,expected,result});}
   await local(context,'refusals-and-recovery',async()=>{await stage(base);for(const expression of [`record.attributes.uint64("${q}", "1").gte(9007199254740993)`,`record.attributes.uint64("${q}", "1").eq("-1")`,`record.attributes.decimal64("${p}", "USD", 2).eq("1.2")`,`record.attributes.decimal64("${p}", "USD", 1+1).valid()`,`record.attributes.int64(record.service, "1").valid()`,`record.attributes.uint64("${q}", "bad unit").valid()`]){const refused=await check(expression,400);expect(refused.diagnostic.span).toBeTruthy();expect(refused.question).toBeUndefined();await expect(text()).toHaveAttribute('aria-invalid','true');context.checks.push({expression,refused})}const plan=await check(cases[0][1]);expect(plan.predicate).toEqual(cases[0][2]);await expect(page.locator('#exploreRun')).toBeEnabled()});
   // The checked ordinary predicate remains editable using the existing numeric
   // controls; this is a representation change, not another source read.
   await local(context,'numeric-visual-edit',async()=>{await editor().getByRole('combobox',{name:'Filter mode',exact:true}).selectOption('conditions');const discard=editor().getByRole('button',{name:'Discard expression',exact:true});if(await discard.isVisible())await discard.click();const key=editor().locator('input[aria-label$=" field"]').filter({visible:true});let id;for(let i=0;i<await key.count();i++)if(await key.nth(i).inputValue()===q)id=(await key.nth(i).getAttribute('aria-label')).match(/Condition (\d+)/)[1];expect(id).toBeTruthy();await expect(editor().getByLabel('Condition '+id+' value',{exact:true})).toHaveValue('9007199254740993');await expect(editor().getByLabel('Condition '+id+' type',{exact:true})).toHaveValue('uint64');});
   if(signal==='logs')for(const width of [1440,390,320]){await page.setViewportSize({width,height:1000});await editor().scrollIntoViewIfNeeded();expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/443-numeric-condition-'+width+'.png'})}await page.setViewportSize({width:1440,height:1100});
  }
  expect(report.consoleRefusals).toHaveLength(12);expect(report.errors).toEqual([]);report.status='passed';
 }catch(error){report.status='failed';report.failure=error.message;await page.screenshot({path:path.join(dir,'failure.png')}).catch(()=>{});throw error}
 finally{await page.goto('about:blank').catch(()=>{});await runtime?.stop();report.cleanup=runtime?.receipt.cleanup||false;await fs.writeFile(path.join(dir,'report.json'),JSON.stringify(report,null,2))}
});
