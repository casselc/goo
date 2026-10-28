const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto');
const {execFile}=require('node:child_process');
const {promisify}=require('node:util');
const fs=require('node:fs/promises'),os=require('node:os'),path=require('node:path');
const {startRuntime}=require('./191-generic-runtime');
const {source,verifyWithGo,startBombadil}=require('./support/cross-app');
const {normalize}=require('../ui/filter-editor.mjs');
const shared={all:[
 {field:{kind:'column',key:'service'},op:'eq',value:'scope-api'},
 {any:[{field:{kind:'attribute',key:'bucket'},op:'missing'},{field:{kind:'attribute',key:'bucket'},op:'eq',value:''}]},
 {field:{kind:'resource',key:'region'},op:'eq',value:'west'},
 {field:{kind:'attribute',key:'region'},op:'eq',value:'east'}
]};

test('239 Activity keeps complete predicates and drafts across signals, refusing late old-signal publication',async({page,request})=>{
 test.setTimeout(180000);page.setDefaultTimeout(15000);
 const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-signal-scope-'));
 const dir=path.resolve('artifacts/signal-scope',randomBytes(4).toString('hex'));
 let runtime;
 try{
  const binary=process.env.GOO_GENERIC_BINARY||path.join(owned,'goo');
  if(!process.env.GOO_GENERIC_BINARY)await promisify(execFile)(process.env.GOO_TEST_GO||'go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});
  runtime=await startRuntime(binary,dir);
  const environment='signal-scope-'+randomBytes(4).toString('hex');
  const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
  const records=[{service:'scope-api',label:'missing',attrs:{region:'west'}},{service:'scope-api',label:'empty',attrs:{region:'east',bucket:''}},{service:'other-api',label:'other',attrs:{region:'east',bucket:''}}];
  const now=BigInt(Date.now()-500)*1000000n;
  const resources=record=>({attributes:attrs({'service.name':record.service,'deployment.environment.name':environment,region:'west'})});
  const logs=await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:records.map((record,i)=>({resource:resources(record),scopeLogs:[{logRecords:[{timeUnixNano:String(now+BigInt(i)),body:{stringValue:'scope log '+record.label},severityText:'INFO',attributes:attrs(record.attrs)}]}]}))}});
  expect(logs.ok()).toBe(true);
  const spans=await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:records.map((record,i)=>({resource:resources(record),scopeSpans:[{spans:[{traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'scope span '+record.label,kind:2,startTimeUnixNano:String(now+BigInt(i)),endTimeUnixNano:String(now+1000000n+BigInt(i)),attributes:attrs(record.attrs)}]}]}))}});
  expect(spans.ok()).toBe(true);
  for(const table of ['otel_logs','otel_traces'])await expect.poll(async()=>Number((await source(request,`SELECT count() AS n FROM default.${table} WHERE ResourceAttributes['deployment.environment.name']={env:String} FORMAT JSONEachRow`,{param_env:environment}))[0].n),{timeout:45000}).toBe(3);
  const saved=await request.post(runtime.base+'/api/activity-views',{data:{version:1,title:'Shared signal scope',selector:{signal:'logs',environment,where:shared},columns:['utc','service','summary']}});
  expect(saved.status(),await saved.text()).toBe(201);const view=await saved.json();
  const streams=[],errors=[],unexpected=[],events=[];
  page.on('pageerror',error=>errors.push(error.message));
  page.on('request',req=>{if(req.method()!=='POST')return;const route=new URL(req.url()).pathname;if(route==='/api/tail')streams.push(req.postDataJSON());else if(['/api/explore','/api/measurements','/api/captures'].includes(route))unexpected.push(route)});
  const shell=page.locator('.stream-entry'),tail=shell.locator('.live-tail'),rows=tail.locator('tbody tr'),signal=shell.getByLabel('Live signal',{exact:true});
  const event=async action=>events.push({action,reads:streams.length,retained:Number(await tail.getAttribute('data-retained')),gap:await tail.getAttribute('data-gap')==='true',coherent:!errors.length&&!unexpected.length});
  await page.goto(runtime.base+'/#activity-view='+view.id);
  await expect(shell.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();expect(streams).toHaveLength(0);
  await shell.getByRole('button',{name:'Start stream',exact:true}).click();await expect(rows).toHaveCount(1);await expect(tail).toContainText('scope log empty');await event('connect');
  expect(streams[0]).toEqual({environment,where:normalize(shared)});
  await signal.selectOption('spans');await expect(rows).toHaveCount(1);await expect(tail).toContainText('scope span empty');await event('connect');
  expect(streams[1]).toEqual({signal:'spans',environment,where:normalize(shared)});
  await signal.selectOption('logs');await expect(tail).toContainText('scope log empty');await event('connect');expect(streams).toHaveLength(3);

  // An unapplied environment edit cannot be committed by a signal gesture.
  await shell.getByLabel('Live environment',{exact:true}).fill(environment+'-draft');await event('pause');
  await signal.selectOption('spans');expect(streams).toHaveLength(3);await event('hold');
  await expect(shell.getByLabel('Live environment',{exact:true})).toHaveValue(environment+'-draft');
  await shell.locator('.entry-investigate-options>summary').click();await shell.getByRole('button',{name:'Analyze this scope/window',exact:true}).click();
  await expect(shell).toContainText('Apply scope before analyzing this window.');expect(streams).toHaveLength(3);await event('hold');
  await shell.locator('.entry-investigate-options>summary').click();
  await shell.getByLabel('Live environment',{exact:true}).fill(environment);

  // Preserve nested OR and qualified namespaces when an incompatible leaf is added.
  await signal.selectOption('logs');await expect(tail).toContainText('scope log empty');await event('connect');
  await shell.locator('.entry-filter-options>summary').click();
  const editor=shell.locator('.entry-filter-options .filter-editor');
  await editor.getByRole('button',{name:'Add condition',exact:true}).click();
  await editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last().fill('body');
  await editor.locator('[aria-label^="Condition "][aria-label$=" value"]').last().fill('scope log');
  const before=streams.length;await event('pause');
  await signal.selectOption('spans');await expect(shell).toContainText('Filters need review for spans');await expect(shell).toContainText('column:body');
  expect(streams).toHaveLength(before);await event('hold');
  await expect(editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last()).toHaveValue('body');
  await expect(editor.locator('[aria-label^="Condition "][aria-label$=" value"]').last()).toHaveValue('scope log');
  await expect(shell.getByRole('button',{name:'Apply scope',exact:true})).toBeDisabled();
  await page.setViewportSize({width:390,height:844});
  await page.screenshot({path:'docs/journeys/screenshots/239-signal-filters-review.png',fullPage:true});
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
  await page.setViewportSize({width:1440,height:1100});
  await signal.selectOption('logs');expect(streams).toHaveLength(before);await event('hold');
  await expect(editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last()).toHaveValue('body');
  // Invalid editor input must survive too; it must not fall back to the last valid tree.
  await editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last().fill('');
  await signal.selectOption('spans');expect(streams).toHaveLength(before);await event('hold');
  await expect(editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last()).toHaveValue('');
  await expect(shell.getByRole('button',{name:'Apply scope',exact:true})).toBeDisabled();
  await editor.locator('[aria-label^="Condition "][aria-label$=" field"]').last().fill('service');
  await editor.locator('[aria-label^="Condition "][aria-label$=" value"]').last().fill('scope-api');
  expect(streams).toHaveLength(before);await event('hold');
  await shell.getByRole('button',{name:'Apply scope',exact:true}).click();await expect(tail).toContainText('scope span empty');await event('connect');
  expect(streams.at(-1)).toEqual({signal:'spans',environment,where:normalize({all:[...shared.all,{field:{kind:'column',key:'service'},op:'eq',value:'scope-api'}]})});

  // Deliberately ignore the old request's abort to exercise a late real SSE body.
  await page.goto(runtime.base+'/#activity-view='+view.id);await page.reload();
  await expect(shell.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();
  await page.evaluate(()=>{
   const original=window.fetch.bind(window);let delay=true;
   window.fetch=async(input,init)=>{
    if(delay&&String(input)==='/api/tail'){
     delay=false;window.scopeLateAbort=new AbortController();
     const response=await original(input,{...init,signal:window.scopeLateAbort.signal});
     window.scopeLateHeaders=true;
     return new Promise(resolve=>{window.scopeReleaseLate=()=>resolve(response)});
    }
    return original(input,init);
   };
  });
  await shell.getByRole('button',{name:'Start stream',exact:true}).click();
  await expect.poll(()=>page.evaluate(()=>!!window.scopeLateHeaders)).toBe(true);await event('connect');
  await signal.selectOption('spans');await expect(tail).toContainText('scope span empty');await event('connect');
  const lastReads=streams.length;await page.evaluate(()=>window.scopeReleaseLate());
  await page.waitForTimeout(2200);await expect(tail).toContainText('scope span empty');await expect(tail).not.toContainText('scope log');expect(streams).toHaveLength(lastReads);await event('hold');
  await page.evaluate(()=>window.scopeLateAbort.abort());
  expect(streams.slice(-2)).toEqual([{environment,where:normalize(shared)},{signal:'spans',environment,where:normalize(shared)}]);
  expect(errors).toEqual([]);expect(unexpected).toEqual([]);
  const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({streams,errors,unexpected,events,firstRowY:await rows.first().evaluate(row=>row.getBoundingClientRect().top)}));
  await verifyWithGo(dir,'TestQuintActivityNavigationLedger',{GOO_ACTIVITY_NAVIGATION_LEDGER:ledger});
  if(process.env.GOO_SIGNAL_SCOPE_BOMBADIL==='1'){
   const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:runtime.base+'/#activity-view='+view.id,specification:'verification/bombadil/activity-signal-scope.ts',seconds:20});
   try{await actor.finish()}finally{await actor.stop()}
  }
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
