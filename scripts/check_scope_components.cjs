// Isolated public-component conformance: real assets, no application or source IO.
const {chromium}=require('playwright');
const fs=require('node:fs/promises');
const path=require('node:path');
const assert=require('node:assert/strict');

const snapshot={
  version:'goo.dataset/1',id:'d',revision:'r',capturedAt:'2026-10-30T02:32:14.794538004Z',
  scope:{id:'s',generation:0,signal:'spans',selector:{}},shape:'ranked',dimensions:['category'],
  fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Count',type:'number'},{key:'value',label:'Value',type:'number',nullable:true}],
  rows:[{key:'a',values:{category:'api',count:'9007199254740993',value:'9007199254740993'}},{key:'b',values:{category:'',count:'0',value:null}}],
  quality:{extent:'partial',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}
};

(async()=>{
  let browser;
  try{
    browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
    const page=await browser.newPage({viewport:{width:390,height:844},hasTouch:true}),requests=[];
    page.on('request',request=>requests.push(new URL(request.url()).pathname));
    await page.route('http://scope.invalid/**',async route=>{
      const name=new URL(route.request().url()).pathname;
      if(name==='/'){
        await route.fulfill({contentType:'text/html',body:`<!doctype html><html lang="en"><meta name="viewport" content="width=device-width"><title>Scope components</title><link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/chart-frame.css"><link rel="stylesheet" href="/scope-view.css"><goo-chart id="pre"></goo-chart><script>window.supplied=${JSON.stringify(snapshot)};const pre=document.getElementById('pre');pre.snapshot=window.supplied;pre.display={kind:'table'};</script><script type="module" src="/scope-view.mjs"></script><div id="root"></div>`});
        return;
      }
      const local=path.resolve(__dirname,'../ui'+name);
      assert(local.startsWith(path.resolve(__dirname,'../ui')+path.sep));
      await route.fulfill({contentType:name.endsWith('.css')?'text/css':'text/javascript',body:await fs.readFile(local)});
    });
    await page.goto('http://scope.invalid/');
    await page.waitForFunction(()=>customElements.get('goo-chart'));
    const result=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied),chart=document.createElement('goo-chart');
      chart.snapshot=snapshot;chart.display={kind:'bar',category:'category',value:'value'};
      document.querySelector('#root').append(chart);
      let select;chart.addEventListener('goo-select',event=>select=event.detail);chart.querySelector('button').click();
      const bars=chart.querySelectorAll('.scope-view-fill').length,exact=chart.textContent.includes('9007199254740993'),quality=chart.textContent.includes('partial groups');
      chart.display={kind:'table'};const table=!!chart.querySelector('table');
      chart.display={kind:'bar',category:'category',value:'count',requiresComplete:true};
      const refusal=chart.textContent.includes('Complete captured-query');
      const facet=document.createElement('goo-scope-facet');facet.snapshot=snapshot;facet.field='category';document.body.append(facet);
      let scope;facet.addEventListener('goo-scope',event=>scope=event.detail);facet.querySelectorAll('button')[1].click();
      const ssr=document.createElement('goo-chart');ssr.setAttribute('snapshot',JSON.stringify(snapshot));ssr.setAttribute('display',JSON.stringify({kind:'table'}));document.body.append(ssr);
      const getters=ssr.snapshot.id===snapshot.id&&ssr.display.kind==='table';
      const clock=ssr.querySelector('.chart-frame-window'),utc=clock.textContent==='Captured 2026-10-30T02:32:14.794Z'&&clock.title===snapshot.capturedAt;
      const override=structuredClone(snapshot);override.id='property';ssr.snapshot=override;ssr.setAttribute('snapshot',JSON.stringify(snapshot));
      const precedence=ssr.snapshot.id==='property';
      return {bars,exact,quality,table,refusal,select,scope,ssr:!!ssr.querySelector('table'),getters,precedence,utc,pre:!!document.querySelector('#pre table'),overflow:document.documentElement.scrollWidth>innerWidth};
    });
    assert.equal(result.bars,1);
    for(const key of ['exact','quality','table','refusal','ssr','getters','precedence','utc','pre'])assert(result[key],key);
    assert.equal(result.select.key,'a');assert.equal(result.scope.value,'');assert.equal(result.overflow,false);
    const ownership=await page.evaluate(()=>{
      const original=structuredClone(window.supplied),chart=document.createElement('goo-chart');chart.snapshot=original;chart.display={kind:'table'};document.body.append(chart);
      original.rows[0].values.count='123';const assigned=chart.snapshot.rows[0].values.count==='9007199254740993';
      const detached=chart.snapshot;detached.rows[0].values.count='456';const getter=chart.snapshot.rows[0].values.count==='9007199254740993';
      const empty=structuredClone(window.supplied);empty.rows=[];empty.quality.extent='unknown';chart.snapshot=empty;const unknown=chart.querySelector('.chart-frame-state').dataset.state==='unknown';
      empty.quality.extent='partial';chart.snapshot=empty;const partial=chart.querySelector('.chart-frame-state').dataset.state==='partial';
      empty.quality.extent='complete';empty.completeness={basis:'captured-query',revision:empty.revision,scopeId:empty.scope.id,aspects:['groups']};chart.snapshot=empty;const knownEmpty=chart.querySelector('.chart-frame-state').dataset.state==='no_data';
      return {assigned,getter,unknown,partial,knownEmpty};
    });
    for(const [key,value] of Object.entries(ownership))assert(value,key);
    const line=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied);delete snapshot.dimensions;
      snapshot.fields=[{key:'x',label:'Numeric x',type:'number'},{key:'y',label:'Value',type:'number'}];
      snapshot.rows=[{key:'left',values:{x:'-1000000000000000000000000000000',y:'1'}},{key:'right',values:{x:'1000000000000000000000000000000',y:'2'}}];
      const chart=document.createElement('goo-chart');chart.snapshot=snapshot;chart.display={kind:'line',category:'x',value:'y'};document.body.append(chart);
      const points=chart.querySelector('polyline')?.getAttribute('points'),frame=!!chart.querySelector('goo-chart-frame .chart-frame-header');
      snapshot.quality.value='unknown';chart.snapshot=snapshot;
      const unknown=!chart.querySelector('svg')&&chart.textContent.includes('value knowledge is unknown');
      snapshot.quality.value='derived';snapshot.rows[0].values.y='-1';chart.snapshot=snapshot;chart.display={kind:'bar',category:'x',value:'y'};
      const signed=!chart.querySelector('.scope-view-fill')&&chart.textContent.includes('requires nonnegative');
      return {points,frame,unknown,signed};
    });
    assert(line.points&&!/NaN|Infinity/.test(line.points));assert(line.frame);assert(line.unknown);assert(line.signed);
    const temporal=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied);snapshot.shape='time-series';delete snapshot.dimensions;
      snapshot.interval={fromMs:'0',toMs:'45000',bucketMs:'10000'};
      snapshot.fields=[{key:'fromMs',label:'From UTC',type:'timestamp',unit:'ms'},{key:'toMs',label:'To UTC',type:'timestamp',unit:'ms'},{key:'count',label:'Captured records',type:'number',nullable:true,unit:'records'},{key:'extra',label:'Additional field',type:'string'}];
      snapshot.rows=[
        {key:'first',state:'observed',values:{fromMs:'0',toMs:'10000',count:'5',extra:'Kept'}},
        {key:'empty',state:'no_data',values:{fromMs:'10000',toMs:'20000',count:'0'}},
        {key:'unknown',state:'unknown',values:{fromMs:'20000',toMs:'30000',count:null}},
        {key:'last',state:'observed',values:{fromMs:'40000',toMs:'45000',count:'18446744073709551615'}}
      ];
      const chart=document.createElement('goo-chart');chart.snapshot=snapshot;chart.display={kind:'time-bar'};document.body.append(chart);
      const buttons=[...chart.querySelectorAll('.scope-time-interval')],gaps=chart.querySelectorAll('.scope-time-gap').length;
      let selection;chart.addEventListener('goo-select',event=>selection=event.detail);buttons[0].click();
      const last=buttons.at(-1),first=buttons[0],positions=Number.parseFloat(last.style.left)>88&&Number.parseFloat(last.style.width)<12&&Number.parseFloat(first.style.width)>22;
      const states=chart.querySelectorAll('[data-state=no_data]').length===1&&chart.querySelectorAll('.scope-time-interval[data-state=unknown]').length===1;
      const title=last.title.includes('1970-01-01T00:00:40.000Z')&&last.title.includes('18446744073709551615')&&last.title.includes('[40000, 45000) ms');
      const noLine=!chart.querySelector('polyline'),quality=chart.textContent.includes('partial intervals')&&chart.textContent.includes('unknown upstream sampling');
      chart.display={kind:'table'};
      const table=chart.textContent.includes('Additional field')&&chart.textContent.includes('Kept')&&chart.textContent.includes('Unknown coverage')&&chart.querySelector('td>span').textContent==='1970-01-01T00:00:00.000Z'&&chart.querySelector('td').title==='0 ms since Unix epoch';
      chart.display={kind:'time-bar',requiresComplete:true};const refuses=chart.textContent.includes('Complete captured-query intervals');
      snapshot.rows[2].state='no_data';snapshot.rows[2].values.count='0';snapshot.rows.splice(3,0,{key:'gap-filled',state:'no_data',values:{fromMs:'30000',toMs:'40000',count:'0'}});
      snapshot.quality.extent='complete';snapshot.completeness={basis:'captured-query',revision:snapshot.revision,scopeId:snapshot.scope.id,aspects:['intervals']};chart.snapshot=snapshot;
      const complete=chart.querySelectorAll('.scope-time-interval').length===5&&!chart.querySelector('.scope-time-gap');
      snapshot.revision='new';chart.snapshot=snapshot;const revisionRefuses=!chart.querySelector('.scope-time-track');
      snapshot.quality.extent='unknown';delete snapshot.completeness;snapshot.rows=[];chart.display={kind:'time-bar'};chart.snapshot=snapshot;
      const emptyUnknown=chart.querySelectorAll('.scope-time-gap').length===1&&chart.textContent.includes('Unknown, not zero');
      return {buttons:buttons.length,gaps,positions,states,title,noLine,quality,table,refuses,complete,revisionRefuses,emptyUnknown,selection};
    });
    assert.equal(temporal.buttons,4);assert.equal(temporal.gaps,1);assert.equal(temporal.selection.key,'first');
    for(const key of ['positions','states','title','noLine','quality','table','refuses','complete','revisionRefuses','emptyUnknown'])assert(temporal[key],key);
    const measures=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied);snapshot.rows=[];
      const numeric={key:'numeric.mean',label:'Numeric mean',type:'rational',nullable:true,unit:'USD',aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}};
      const duration={key:'duration.p95',label:'p95 duration',type:'rational',nullable:true,unit:'ns',aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}};
      snapshot.fields.push(numeric,duration);
      const support=(total,valid,missing='0',empty='0',invalid='0')=>({total,valid,missing,empty,invalid});
      const row=(key,count,mean,latency,numericCell,durationCell)=>({key,values:{category:key,count,'numeric.mean':mean,'duration.p95':latency},cells:{'numeric.mean':numericCell,'duration.p95':durationCell}});
      snapshot.rows=[
        row('a','5','7/6','4000000',{state:'estimated',support:support('5','3','1','0','1')},{state:'derived',support:support('5','5')}),
        row('b','2','-2','1500',{state:'derived',support:support('2','2')},{state:'derived',support:support('2','2')}),
        row('c','3',null,null,{state:'no_valid_values',support:support('3','0','1','1','1')},{state:'no_valid_values',support:support('3','0','1','1','1')}),
        row('d','0',null,null,{state:'no_data',support:support('0','0')},{state:'no_data',support:support('0','0')})
      ];
      const chart=document.createElement('goo-chart');chart.snapshot=snapshot;chart.display={kind:'bar',category:'category',value:'numeric.mean'};document.body.append(chart);
      const signed=!chart.querySelector('.scope-view-bars')&&chart.textContent.includes('requires nonnegative');
      const rational=chart.querySelector('td[title*="7/6 USD"]')?.textContent.includes('≈ 1.167 USD');
      const supportVisible=chart.querySelector('td[title*="7/6 USD"]')?.textContent.includes('3/5 valid');
      chart.display={kind:'table'};
      const durationTable=chart.querySelector('td[title*="4000000 ns"]')?.textContent.includes('4 ms');
      const headers=[...chart.querySelectorAll('th')].every(th=>!th.textContent.includes('ns'));
      chart.display={kind:'bar',category:'category',value:'duration.p95'};
      const exactQuantiles=chart.querySelectorAll('.scope-view-fill').length===2&&chart.querySelector('.scope-view-bar').title.includes('quantilesExactHigh')&&!chart.querySelector('.scope-view-bar[data-cell-state=estimated]');
      const unavailable=chart.querySelectorAll('.scope-view-bar[data-cell-state=no_valid_values] .scope-view-fill').length===0&&chart.textContent.includes('No valid values');
      snapshot.rows.splice(1,1);chart.snapshot=snapshot;chart.display={kind:'bar',category:'category',value:'numeric.mean'};
      const estimated=chart.querySelector('.scope-view-bar[data-cell-state=estimated] .scope-view-fill')!==null;
      snapshot.shape='time-series';delete snapshot.dimensions;snapshot.interval={fromMs:'0',toMs:'20000',bucketMs:'10000'};
      snapshot.fields=[{key:'fromMs',label:'From',type:'timestamp',unit:'ms'},{key:'toMs',label:'To',type:'timestamp',unit:'ms'},{key:'count',label:'Count',type:'number',nullable:true,unit:'records'},numeric,duration];
      snapshot.rows=snapshot.rows.slice(0,2).map((item,index)=>({...item,state:'observed',values:{fromMs:String(index*10000),toMs:String((index+1)*10000),count:item.values.count,'numeric.mean':index===0?'-2':null,'duration.p95':item.values['duration.p95']}}));
      chart.snapshot=snapshot;chart.display={kind:'time-bar',value:'numeric.mean'};
      const signedTime=!chart.querySelector('.scope-time-track')&&chart.textContent.includes('requires nonnegative');
      chart.display={kind:'time-bar',value:'duration.p95'};
      const timeMeasure=chart.querySelectorAll('.scope-time-interval[data-state=observed]').length===1&&chart.querySelectorAll('.scope-time-interval[data-cell-state=no_valid_values]').length===1&&chart.querySelector('.scope-time-interval').title.includes('quantilesExactHigh');
      const invalidGap=chart.querySelector('.scope-time-interval[data-cell-state=no_valid_values]').dataset.state==='unknown';
      return {signed,rational,supportVisible,durationTable,headers,exactQuantiles,unavailable,estimated,signedTime,timeMeasure,invalidGap};
    });
    for(const [key,value] of Object.entries(measures))assert(value,key);
    const exactAccess=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied);snapshot.rows=[{key:'precise',values:{category:'Precise',count:'1',mean:'1/3',amount:'9007199254740993.123456789'},cells:{mean:{state:'derived',support:{total:'1',valid:'1',missing:'0',empty:'0',invalid:'0'}}}}];
      snapshot.fields.push({key:'mean',label:'Mean',type:'rational',nullable:true,unit:'USD',aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}},{key:'amount',label:'Amount',type:'number',unit:'USD'});
      const chart=document.createElement('goo-chart');chart.id='exactConformance';chart.snapshot=snapshot;chart.display={kind:'table'};document.body.append(chart);
      const facet=document.createElement('goo-scope-facet');facet.snapshot=snapshot;facet.field='mean';document.body.append(facet);
      let scopeEvents=0;facet.addEventListener('goo-scope',()=>scopeEvents++);
      return {refusal:facet.textContent.includes('Choose a group dimension')&&!facet.querySelector('button'),scopeEvents};
    });
    assert(exactAccess.refusal);assert.equal(exactAccess.scopeEvents,0);
    const rationalDetails=page.locator('#exactConformance td[title^="1/3 USD"] details');
    await rationalDetails.locator('summary').focus();await page.keyboard.press('Enter');
    assert(await rationalDetails.locator('.scope-view-exact').isVisible());
    assert((await rationalDetails.locator('.scope-view-exact').textContent()).includes('1/3 USD'));
    const decimalDetails=page.locator('#exactConformance td[title^="9007199254740993.123456789 USD"] details');
    await decimalDetails.locator('summary').tap();
    assert(await decimalDetails.locator('.scope-view-exact').isVisible());
    assert((await decimalDetails.locator('.scope-view-exact').textContent()).includes('9007199254740993.123456789 USD'));
    assert(!requests.some(name=>name.startsWith('/api/')||name==='/app.mjs'));
    console.log('Scoped components: exact values/rationals, human units/support/exact quantiles, per-cell gaps/estimates/signed fallback, complete-only refusal, table/bar/line/time-bar, interval gaps/UTC, typed intents, SSR/pre-upgrade,390px layout and no application IO pass');
  }finally{await browser?.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
