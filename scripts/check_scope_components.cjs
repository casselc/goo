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
    assert(!requests.some(name=>name.startsWith('/api/')||name==='/app.mjs'));
    console.log('Scoped components: exact values, missing cells, quality, complete-only refusal, table/bar/line, signed/unknown fallback, typed intents, SSR/pre-upgrade,390px layout and no application IO pass');
  }finally{await browser?.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
