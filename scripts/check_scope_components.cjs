// Isolated public-component conformance: real assets, no application or source IO.
const {chromium}=require('playwright');
const fs=require('node:fs/promises');
const path=require('node:path');
const assert=require('node:assert/strict');

const snapshot={
  version:'goo.dataset/1',id:'d',revision:'r',capturedAt:'2026-10-30T02:32:14.794538004Z',
  scope:{id:'s',generation:0,signal:'spans',selector:{}},shape:'ranked',
  fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Count',type:'number'}],
  rows:[{key:'a',values:{category:'api',count:'9007199254740993'}},{key:'b',values:{category:'',count:null}}],
  quality:{extent:'partial',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}
};

(async()=>{
  let browser;
  try{
    browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
    const page=await browser.newPage({viewport:{width:390,height:844}}),requests=[];
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
      chart.snapshot=snapshot;chart.display={kind:'bar',category:'category',value:'count'};
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
    const line=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied);
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
      const snapshot=structuredClone(window.supplied);snapshot.shape='time-series';
      snapshot.interval={fromMs:'0',toMs:'45000',bucketMs:'10000'};
      snapshot.fields=[{key:'fromMs',label:'From UTC',type:'timestamp',unit:'ms'},{key:'toMs',label:'To UTC',type:'timestamp',unit:'ms'},{key:'count',label:'Captured records',type:'number',unit:'records'},{key:'extra',label:'Additional field',type:'string'}];
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
      const table=chart.textContent.includes('Additional field')&&chart.textContent.includes('Kept')&&chart.textContent.includes('Unknown coverage')&&chart.querySelector('td').textContent==='1970-01-01T00:00:00.000Z'&&chart.querySelector('td').title==='0 ms since Unix epoch';
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
    assert(!requests.some(name=>name.startsWith('/api/')||name==='/app.mjs'));
    console.log('Scoped components: exact values, missing cells, quality, complete-only refusal, table/bar/line/time-bar, interval gaps/no-data/unknown/UTC/clipped widths, typed intents, SSR/pre-upgrade,390px layout and no application IO pass');
  }finally{await browser?.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
