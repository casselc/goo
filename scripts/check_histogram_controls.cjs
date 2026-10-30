// Canonical histogram adapter and stale async planning checks; real SDK journey is separate.
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
 const N=await import('/numeric-distribution.mjs'),{histogramEdges,histogramIntent}=await import('/view-model.mjs'),{createSourceAdapter}=await import('/snapshots.mjs'),{createActionContext}=await import('/scope-actions.mjs');
 const q={signal:'logs',environment:'test',fromMs:0,toMs:20000,numeric:{field:{kind:'attribute',key:'items'},type:'uint64',unit:'items',parse:'canonical-decimal'},numericHistogram:{binning:'powers-of-two'},limit:100,offset:0};
 const bounds=histogramEdges(q.numeric),rows=bounds.slice(0,-1).map((from,i)=>({key:'bin:'+i,values:{from,to:bounds[i+1],count:i===3?'2':i===6?'1':'0'},state:[3,6].includes(i)?'observed':'no_data'}));
 const raw={version:'goo.dataset/1',id:'histogram',revision:'stable',origin:'source-read',shape:'histogram',scope:{id:'q',generation:'stable',signal:'logs',selector:q},quality:{extent:'complete',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'stable',scopeId:'q',aspects:['bins']},fields:[{key:'from',label:'From',type:'rational',unit:'items'},{key:'to',label:'To',type:'rational',unit:'items'},{key:'count',label:'Captured records',type:'number',unit:'records'}],histogram:{kind:'numeric',binning:'powers-of-two',binding:q.numeric,support:{total:'6',valid:'3',missing:'1',empty:'1',invalid:'1'},closure:'left-closed-right-open'},rows};
 const source=createSourceAdapter({id:'hist-check'}),snapshot=source.adoptSnapshot(raw,{question:q});let epoch=1,opened=0,requests=0,pending;
 const context=createActionContext({epoch:()=>epoch,active:()=>true,canStage:()=>true}),binding=context.bind({role:'current',slot:'histogram',epoch:1,snapshot,question:q,coherent:true});
 const data={question:q,histogramDataset:snapshot,count:'6',fields:[]};
 const owner=N.mount(document.querySelector('#distribution'),data,{open:()=>opened++,resolveRange:intent=>context.histogram(binding,intent)});
 window.fetch=async(path,options)=>{requests++;return new Promise(resolve=>{pending=()=>resolve({ok:true,json:async()=>({parent:q,fromBin:3,toBin:7,from:'4',to:'64',selected:q,rest:q,comparison:{selection:q,baseline:q},numericComparison:{selection:q,baseline:q}})})})};
 window.state=()=>({opened,requests});window.finish=()=>pending();window.retire=()=>{epoch++;context.revoke()};window.intent=histogramIntent(snapshot,'bin:3','bin:6');window.owner=owner;
 return {same:document.querySelector('goo-chart').snapshot===snapshot,visible:document.querySelectorAll('goo-time-bars button').length};
 });assert(state.same);assert.equal(state.visible,4);
 await page.locator('goo-chart').evaluate(chart=>chart.dispatchEvent(new CustomEvent('goo-histogram-range',{detail:window.intent})));
 assert.match(await page.locator('.numeric-distribution-selection [role=status]').textContent(),/3 of 6 records selected/);
 await page.getByRole('button',{name:'Run selected population',exact:true}).click();assert.deepEqual(await page.evaluate(()=>window.state()),{opened:0,requests:1});
 await page.evaluate(()=>{window.retire();window.finish()});await page.waitForTimeout(50);assert.equal((await page.evaluate(()=>window.state())).opened,0);
 await page.locator('goo-chart').evaluate(chart=>chart.dispatchEvent(new CustomEvent('goo-histogram-range',{detail:window.intent})));await page.getByRole('button',{name:'Run selected population',exact:true}).click();assert.equal((await page.evaluate(()=>window.state())).requests,1);
 assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true,'nested histogram must not overflow the page at390px');assert.deepEqual(errors,[]);console.log('histogram controls: PASS (canonical handle, full indices after outer trimming, same-read coverage, retired async range fenced)');
 }finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
