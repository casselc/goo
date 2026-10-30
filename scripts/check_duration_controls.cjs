// Canonical duration owner checks; SDK journeys exercise actual telemetry.
const {chromium}=require('playwright');
const fs=require('node:fs/promises'),path=require('node:path'),assert=require('node:assert/strict');
(async()=>{let browser;try{
 browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
 const page=await browser.newPage({viewport:{width:390,height:844}}),requests=[],errors=[];
 page.on('request',request=>requests.push(new URL(request.url()).pathname));page.on('pageerror',error=>errors.push(error.message));
 await page.route('http://distribution.invalid/**',async route=>{const name=new URL(route.request().url()).pathname;
  if(name==='/'){await route.fulfill({contentType:'text/html',body:'<!doctype html><html lang="en"><meta name="viewport" content="width=device-width"><title>Distribution controls</title><link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/chart-frame.css"><link rel="stylesheet" href="/scope-view.css"><link rel="stylesheet" href="/numeric-distribution.css"><div id="distribution"></div><div id="drecordsPanel"></div>'});return}
  const local=path.resolve(__dirname,'../ui'+name);assert(local.startsWith(path.resolve(__dirname,'../ui')+path.sep));await route.fulfill({contentType:name.endsWith('.css')?'text/css':'text/javascript',body:await fs.readFile(local)});
 });
 await page.goto('http://distribution.invalid/');
 const state=await page.evaluate(async()=>{
 const {mount}=await import('/latency-histogram.mjs'),{createSourceAdapter}=await import('/snapshots.mjs'),{createActionContext}=await import('/scope-actions.mjs');
 const q={signal:'spans',environment:'test',fromMs:1000,toMs:61000,limit:1,offset:0,groupBy:[{kind:'column',key:'duration_bucket'}]},edges=['0','1000000','5000000','10000000','50000000','100000000','500000000','1000000000','+inf'];
 const raw={version:'goo.dataset/1',id:'duration',revision:'stable',origin:'source-read',capturedAt:'2026-10-30T00:00:00Z',shape:'histogram',scope:{id:'duration',generation:'stable',signal:'spans',selector:q},quality:{extent:'complete',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'stable',scopeId:'duration',aspects:['bins']},fields:[{key:'from',label:'From',type:'histogram-bound',unit:'ns'},{key:'to',label:'To',type:'histogram-bound',unit:'ns'},{key:'count',label:'Captured records',type:'number',unit:'records'}],histogram:{kind:'span-duration',binning:'goo-duration-v1',unit:'ns',total:'9007199254740994',closure:'left-closed-right-open'},rows:edges.slice(0,-1).map((from,i)=>({key:'bin:'+i,values:{from,to:edges[i+1],count:i===2?'9007199254740993':i===7?'1':'0'},state:[2,7].includes(i)?'observed':'no_data'}))};
 const snapshot=createSourceAdapter({id:'duration-check'}).adoptSnapshot(raw,{question:q}),context=createActionContext({epoch:()=>1,active:()=>true,canStage:()=>true}),binding=context.bind({role:'current',slot:'histogram',epoch:1,snapshot,question:q,coherent:true});let opened=0;
 const data={question:q,count:raw.histogram.total,state:'observed',histogramDataset:snapshot};window.owner=mount(document.querySelector('#distribution'),data,{open:()=>opened++,canAct:()=>context.canAct(binding),resolveRange:intent=>context.duration(binding,intent)});window.retire=()=>context.revoke();window.opened=()=>opened;window.raw=raw;window.data=data;
 return {same:document.querySelector('goo-chart').snapshot===snapshot,clear:typeof document.querySelector('goo-chart').clearRange==='function'};
 });assert(state.same);assert(state.clear);
 await page.locator('goo-time-bars button[data-key="bin:7"]').click();await page.getByText('Selection bounds and predicates',{exact:true}).click();assert.equal(await page.getByLabel('Minimum span duration (ms)',{exact:true}).inputValue(),'1000');assert.equal(await page.getByLabel('Maximum span duration (ms, exclusive)',{exact:true}).inputValue(),'');
 await page.getByRole('button',{name:'Run selected population',exact:true}).click();assert.equal(await page.evaluate(()=>window.opened()),1);
 await page.getByLabel('Minimum span duration (ms)',{exact:true}).fill('1001');assert.equal(await page.locator('[data-range-selected=true]').count(),0);
 await page.evaluate(()=>window.retire());await page.locator('.population-selection form').dispatchEvent('submit');assert.equal(await page.evaluate(()=>window.opened()),1);await page.evaluate(()=>window.owner.suspend());assert(await page.getByRole('button',{name:'Compare with rest',exact:true}).isDisabled());
 // Activity's onBegin pauses its read gate, but keeps the captured rendered population usable.
 await page.evaluate(async()=>{
  if(!crypto.randomUUID)crypto.randomUUID=()=> '00000000-0000-4000-8000-000000000001';const A=await import('/activity-latency.mjs'),host=document.createElement('div'),controls=document.createElement('div');document.body.append(controls,host);let stages=0,pauses=0;
  const c={active:true,clean:true,phase:'partial',selector:{signal:'spans',environment:'test'},readWindow:{fromMs:1000,toMs:61000}};
  window.fetch=async(path,options)=>{const q=JSON.parse(options.body),raw=structuredClone(window.raw);raw.scope.selector=q;return {ok:true,json:async()=>({question:q,basis:'stable',capturedAt:raw.capturedAt,count:raw.histogram.total,state:'observed',groups:[{g0:'[5,10) ms',count:'9007199254740993'},{g0:'[1000,+inf) ms',count:'1'}],histogramDataset:raw})}};
  window.activity=A.mount(host,{controlsHost:controls,current:()=>c,pauseStream:()=>pauses++,onStage:()=>stages++,onCompare:()=>{},onBaselineChange:()=>{}});window.activityCounts=()=>({stages,pauses});
 });
 await page.getByRole('button',{name:'Show latency',exact:true}).click();await page.locator('.activity-latency goo-time-bars button[data-key="bin:2"]').click();assert.equal((await page.evaluate(()=>window.activityCounts())).stages,1);assert((await page.evaluate(()=>window.activityCounts())).pauses>0);
 assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);assert.deepEqual(errors,[]);console.log('duration controls: PASS (canonical+infinitebound, exactmanual/stale guards, Activityselection survivespause,390px)');
 }finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
