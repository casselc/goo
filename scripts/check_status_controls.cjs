// Focused ordinary Distribution adapter checks; SDK/ClickHouse journey is separate.
const {chromium}=require('playwright');
const fs=require('node:fs/promises'),path=require('node:path'),assert=require('node:assert/strict');
(async()=>{let browser;try{
 browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
 const page=await browser.newPage({viewport:{width:390,height:844}}),requests=[],errors=[];
 page.on('request',request=>requests.push(new URL(request.url()).pathname));page.on('pageerror',error=>errors.push(error.message));
 await page.route('http://distribution.invalid/**',async route=>{const name=new URL(route.request().url()).pathname;
  if(name==='/'){await route.fulfill({contentType:'text/html',body:'<!doctype html><html lang="en"><meta name="viewport" content="width=device-width"><title>Distribution controls</title><link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/chart-frame.css"><link rel="stylesheet" href="/scope-view.css"><div id="distribution"></div><div id="drecordsPanel"></div>'});return}
  const local=path.resolve(__dirname,'../ui'+name);assert(local.startsWith(path.resolve(__dirname,'../ui')+path.sep));await route.fulfill({contentType:name.endsWith('.css')?'text/css':'text/javascript',body:await fs.readFile(local)});
 });
 await page.goto('http://distribution.invalid/');
 const state=await page.evaluate(async()=>{
 const D=await import('/distributions.mjs'),{createSourceAdapter}=await import('/snapshots.mjs');
 const q={signal:'spans',environment:'test',fromMs:0,toMs:20000,spanStatus:true,groupBy:[{kind:'column',key:'service'}],limit:100,offset:0};
 const fields=[{key:'category',label:'Service',type:'string'},{key:'count',label:'Captured records',type:'number',unit:'records'},...['error','ok','unset','unknown'].map(k=>({key:'status.'+k,label:k,type:'number',unit:'records',nullable:true,aggregation:{family:'span-status',kind:'count',method:'exact'}})),{key:'status.error_fraction',label:'Recorded error fraction',type:'rational',unit:'1',nullable:true,aggregation:{family:'ratio',kind:'fraction',method:'exact-rational',numerator:'status.error',denominator:'count'}}];
 const row=(key,count,error,ok)=>({key,values:{category:key,count,'status.error':error,'status.ok':ok,'status.unset':'0','status.unknown':'0','status.error_fraction':error===count?'1':'1/100'},cells:Object.fromEntries(fields.filter(f=>f.aggregation).map(f=>[f.key,{state:'derived'}]))});
 const raw={version:'goo.dataset/1',id:'groups',revision:'stable',origin:'source-read',shape:'ranked',scope:{id:'q',generation:'stable',signal:'spans',selector:q},dimensions:['category'],partition:{kind:'disjoint-record-groups',measure:'count',total:'101'},quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'stable',scopeId:'q',aspects:['groups']},fields,rows:[row('frontend','100','1','99'),row('worker','1','1','0')]};
 const source=createSourceAdapter({id:'status-check'}),snapshot=source.adoptSnapshot(raw,{question:q});
 const summaryRaw={...raw,id:'summary',shape:'summary',fields:fields.slice(1),rows:[{...row('population','101','2','99'),values:{...row('population','101','2','99').values,'status.error_fraction':'2/101'}}],completeness:{...raw.completeness,aspects:['summary']}};delete summaryRaw.dimensions;delete summaryRaw.partition;delete summaryRaw.rows[0].values.category;
 const summary=source.adoptSnapshot(summaryRaw,{question:q});
 const view=D.mount(document.querySelector('#distribution'),{prefix:'d'});view.render({question:q,basis:'stable',count:'101',state:'observed',rows:[],fields:[],groups:[],timeSeries:[],spanStatus:{count:'101',error:'2',ok:'99',unset:'0',unknown:'0'},viewDataset:snapshot,summaryDataset:summary});view.select('Distribution');window.adapter=view;
 return {canonical:D.usesCanonicalDistribution({question:q,viewDataset:snapshot}),summary:!!document.querySelector('#dnumericSummary goo-measure-summary'),legacyHidden:document.querySelector('#dspanStatusSummary').hidden};
 });
 for(const [key,value]of Object.entries(state))assert(value,key);
 await page.getByRole('combobox',{name:'Breakdown measure',exact:true}).selectOption('status.error_fraction');
 assert.equal(await page.getByRole('combobox',{name:'Count scale',exact:true}).isVisible(),false);
 let widths=await page.locator('#ddistributionContent .scope-view-fill').evaluateAll(nodes=>nodes.map(n=>parseFloat(n.style.width)).sort((a,b)=>a-b));assert.deepEqual(widths,[1,100]);
 await page.locator('#dfacetSearch').fill('frontend');assert.equal(await page.locator('#ddistributionContent .scope-view-fill').evaluate(n=>parseFloat(n.style.width)),1);
 await page.getByRole('button',{name:/^Select "frontend":/}).click();assert.match(await page.locator('.scoped-breakdown-selection').textContent(),/ok: 99/);assert.match(await page.locator('.scoped-breakdown-selection').textContent(),/unknown: 0/);
 await page.getByRole('combobox',{name:'Breakdown measure',exact:true}).selectOption('status.unknown');assert.equal(await page.locator('#ddistributionContent .scope-view-fill').evaluate(n=>parseFloat(n.style.width)),0);
 assert(!requests.some(name=>name.startsWith('/api/')));assert.deepEqual(errors,[]);console.log('status controls: PASS (fixed fraction axis, status-only summary, class context, local no IO)');
 }finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
