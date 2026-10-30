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
 const initial=await page.evaluate(async()=>{
  const Distributions=await import('/distributions.mjs'),{createSourceAdapter}=await import('/snapshots.mjs'),{createActionContext}=await import('/scope-actions.mjs');
  const q={signal:'spans',environment:'test',fromMs:0,toMs:20000,summary:'span-duration',numeric:{field:{kind:'attribute',key:'price'},type:'int64',unit:'USD',parse:'canonical-decimal'},groupBy:[{kind:'column',key:'service'}],limit:100,offset:0};
  const support=n=>({total:n,valid:n,missing:'0',empty:'0',invalid:'0'}),source=createSourceAdapter({id:'distribution-conformance'});
  const raw={version:'goo.dataset/1',id:'groups',revision:'stable',origin:'source-read',shape:'ranked',scope:{id:'q',generation:'stable',signal:'spans',selector:q},dimensions:['category'],partition:{kind:'disjoint-record-groups',measure:'count',total:'4'},quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'stable',scopeId:'q',aspects:['groups']},fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Captured records',type:'number',unit:'records'},{key:'numeric.mean',label:'Numeric mean',type:'rational',nullable:true,unit:'USD',aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}},{key:'duration.p95',label:'p95 duration',type:'rational',nullable:true,unit:'ns',aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}}],rows:[{key:'front',values:{category:'frontend',count:'3','numeric.mean':'-2','duration.p95':'2000000'},cells:{'numeric.mean':{state:'derived',support:support('3')},'duration.p95':{state:'derived',support:support('3')}}},{key:'worker',values:{category:'worker',count:'1','numeric.mean':'4','duration.p95':'1000000'},cells:{'numeric.mean':{state:'derived',support:support('1')},'duration.p95':{state:'derived',support:support('1')}}}]};
  const snapshot=source.adoptSnapshot(raw,{question:q});let mounts=0,stages=0,phase='applied';const context=createActionContext({epoch:()=>1,active:()=>true,canStage:()=>phase==='applied'});
  const view=Distributions.mount(document.querySelector('#distribution'),{prefix:'d',bindDistribution:data=>{mounts++;const binding=context.bind({role:'current',slot:'groups',epoch:1,snapshot:data.viewDataset,question:data.question,coherent:true});return {onScope:intent=>{const result=context.resolve(binding,intent);if(result.ok){stages++;phase='draft';context.revoke()}}}}});
  const data={question:q,basis:'stable',count:'4',state:'observed',rows:[],fields:[],groups:[{g0:'frontend',count:'3'},{g0:'worker',count:'1'}],timeSeries:[],viewDataset:snapshot};view.render(data);view.select('Distribution');
  window.adapter=view;window.original=snapshot;window.counters=()=>({mounts,stages,phase});
  return {ordinary:Distributions.usesCanonicalDistribution(data),same:document.querySelector('#ddistributionContent goo-chart').snapshot===snapshot,oneFind:document.querySelectorAll('#ddistributionPanel input[type=search]').length===1,noFacetPills:!document.querySelector('#ddistributionContent goo-scope-facet'),oldMeasureHidden:document.querySelector('#ddistributionMeasureLabel').hidden};
 });
 for(const [key,value]of Object.entries(initial))assert(value,key);
 await page.locator('#dfacetSearch').fill('worker');
 const population=await page.locator('#ddistributionContent goo-chart').evaluate(chart=>({width:parseFloat(chart.querySelector('.scope-view-fill').style.width),same:chart.snapshot===window.original,scale:chart.display.scale}));assert.equal(population.width,25);assert(population.same);assert.equal(population.scale,'population');
 await page.getByRole('combobox',{name:'Count scale',exact:true}).selectOption('max');
 assert(Math.abs(await page.locator('#ddistributionContent .scope-view-fill').evaluate(fill=>parseFloat(fill.style.width))-100/3)<.01);
 await page.getByRole('combobox',{name:'Breakdown measure',exact:true}).selectOption('numeric.mean');
 assert(await page.locator('#ddistributionContent table').isVisible());assert((await page.locator('#ddistributionContent').textContent()).includes('requires nonnegative'));
 const presentation=await page.evaluate(()=>window.adapter.presentation());assert.equal(presentation.value,'numeric.mean');assert(!Object.hasOwn(presentation,'metric'));
 await page.getByRole('combobox',{name:'Breakdown measure',exact:true}).selectOption('count');
 await page.getByRole('combobox',{name:'Count scale',exact:true}).selectOption('population');
 await page.getByText('More options',{exact:true}).click();await page.getByRole('combobox',{name:'Group order',exact:true}).selectOption('value-asc');
 assert.equal((await page.evaluate(()=>window.counters())).mounts,1);
 await page.locator('#ddistributionContent .scope-view-bar').click();await page.getByRole('button',{name:'Stage selected group',exact:true}).click();
 assert.deepEqual(await page.evaluate(()=>window.counters()),{mounts:1,stages:1,phase:'draft'});
 assert.equal(errors.length,0,errors.join('; '));assert(!requests.some(name=>name.startsWith('/api/')));
 console.log('Distribution controls: one canonical surface/Find, full-population25% vsfull-max33%, immutablehandle, localmeasure/sort, exactsignedfallback, field-key presentation and one capturedstagebinding/no source IO pass');
}finally{await browser?.close()}})().catch(error=>{console.error(error);process.exitCode=1});
