// Isolated public-component conformance: real assets, no application or source IO.
const {chromium}=require('playwright');
const fs=require('node:fs/promises');
const path=require('node:path');
const assert=require('node:assert/strict');

const snapshot={
  version:'goo.dataset/1',id:'d',revision:'r',capturedAt:'2026-10-30T02:32:14.794538004Z',
  origin:'source-read',scope:{id:'s',generation:0,signal:'spans',selector:{signal:'spans',environment:'test',fromMs:0,toMs:45000,groupBy:[{kind:'column',key:'service'}]}},shape:'ranked',dimensions:['category'],
  fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Count',type:'number'},{key:'value',label:'Value',type:'number',nullable:true}],
  rows:[{key:'a',values:{category:'api',count:'9007199254740993',value:'9007199254740993'}},{key:'b',values:{category:'',count:'0',value:null}}],
  quality:{extent:'partial',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}
};

(async()=>{
  let browser;
  try{
    browser=await chromium.launch({executablePath:process.env.CHROMIUM_PATH||'/usr/bin/chromium',args:['--no-sandbox']});
    const page=await browser.newPage({viewport:{width:390,height:844},hasTouch:true}),requests=[],errors=[];
    page.on('request',request=>requests.push(new URL(request.url()).pathname));
    page.on('pageerror',error=>errors.push(error.message));
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
    await page.evaluate(async()=>{const {createSourceAdapter,asSnapshot,hasGuarantee}=await import('/snapshots.mjs');const source=createSourceAdapter({id:'component-conformance'});window.adopt=value=>source.adoptSnapshot(value,{question:value.scope.selector});window.asSnapshot=asSnapshot;window.hasGuarantee=hasGuarantee});
    const result=await page.evaluate(()=>{
      const snapshot=structuredClone(window.supplied),chart=document.createElement('goo-chart');
      const source=window.adopt(snapshot);chart.snapshot=source;chart.display={kind:'bar',category:'category',value:'value'};
      document.querySelector('#root').append(chart);
      let select;chart.addEventListener('goo-select',event=>select=event.detail);chart.querySelector('button').click();
      const bars=chart.querySelectorAll('.scope-view-fill').length,exact=chart.textContent.includes('9007199254740993'),quality=chart.textContent.includes('partial groups');
      chart.display={kind:'table'};const table=!!chart.querySelector('table');
      chart.display={kind:'bar',category:'category',value:'count',requiresComplete:true};
      const refusal=chart.textContent.includes('Complete-only display');
      const facet=document.createElement('goo-scope-facet');facet.snapshot=source;facet.field='category';document.body.append(facet);
      let scope;facet.addEventListener('goo-scope',event=>scope=event.detail);facet.querySelectorAll('button')[1].click();
      const ssr=document.createElement('goo-chart');ssr.setAttribute('snapshot',JSON.stringify(snapshot));ssr.setAttribute('display',JSON.stringify({kind:'table'}));document.body.append(ssr);
      const getters=ssr.snapshot.id.startsWith('goo.local/')&&ssr.display.kind==='table'&&ssr.snapshot===ssr.snapshot&&ssr.snapshot.origin==='browser-held'&&!window.hasGuarantee(ssr.snapshot,'groups');
      const cached=ssr.snapshot;ssr.setAttribute('display',JSON.stringify({kind:'table',title:'Imported data'}));const cachedAttribute=ssr.snapshot===cached;
      const shared=chart.snapshot===source&&facet.snapshot===source&&Object.isFrozen(source.rows[0].values);
      const clock=ssr.querySelector('.chart-frame-window'),utc=clock.textContent==='Captured 2026-10-30T02:32:14.794Z'&&clock.title===snapshot.capturedAt;
      const override=structuredClone(snapshot);override.id='property';ssr.snapshot=override;ssr.setAttribute('snapshot',JSON.stringify(snapshot));
      const precedence=ssr.snapshot.origin==='browser-held'&&ssr.snapshot.id.startsWith('goo.local/');
      return {bars,exact,quality,table,refusal,select,scope,ssr:!!ssr.querySelector('table'),getters,shared,cachedAttribute,precedence,utc,pre:!!document.querySelector('#pre table'),overflow:document.documentElement.scrollWidth>innerWidth};
    });
    assert.equal(result.bars,1);
    for(const key of ['exact','quality','table','refusal','ssr','getters','shared','cachedAttribute','precedence','utc','pre'])assert(result[key],key);
    assert.equal(result.select.key,'a');assert.equal(result.scope.value,'');assert.equal(result.overflow,false);
    const ownership=await page.evaluate(()=>{
      const original=structuredClone(window.supplied),chart=document.createElement('goo-chart');chart.snapshot=original;chart.display={kind:'table'};document.body.append(chart);
      original.rows[0].values.count='123';const assigned=chart.snapshot.rows[0].values.count==='9007199254740993';
      const detached=chart.snapshot;try{detached.rows[0].values.count='456'}catch{}const getter=chart.snapshot===detached&&chart.snapshot.rows[0].values.count==='9007199254740993'&&Object.isFrozen(detached.rows[0].values);
      const empty=structuredClone(window.supplied);empty.rows=[];empty.quality.extent='unknown';chart.snapshot=empty;const unknown=chart.querySelector('.chart-frame-state').dataset.state==='unknown';
      empty.quality.extent='partial';chart.snapshot=window.adopt(empty);const partial=chart.querySelector('.chart-frame-state').dataset.state==='partial';
      empty.quality.extent='complete';empty.completeness={basis:'captured-query',revision:empty.revision,scopeId:empty.scope.id,aspects:['groups']};chart.snapshot=window.adopt(empty);const knownEmpty=chart.querySelector('.chart-frame-state').dataset.state==='no_data';
      const imported=window.asSnapshot(JSON.parse(JSON.stringify(chart.snapshot)));chart.snapshot=imported;chart.display={kind:'table',requiresComplete:true};const noRestoredProof=!chart.querySelector('table')&&!window.hasGuarantee(imported,'groups');chart.display={kind:'table'};const importedValues=!!chart.querySelector('goo-chart-frame');
      return {assigned,getter,unknown,partial,knownEmpty,noRestoredProof,importedValues};
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
      const snapshot=structuredClone(window.supplied);snapshot.shape='time-series';delete snapshot.dimensions;snapshot.scope.selector.bucketMs=10000;
      snapshot.interval={fromMs:'0',toMs:'45000',bucketMs:'10000'};
      snapshot.fields=[{key:'fromMs',label:'From UTC',type:'timestamp',unit:'ms'},{key:'toMs',label:'To UTC',type:'timestamp',unit:'ms'},{key:'count',label:'Captured records',type:'number',nullable:true,unit:'records'},{key:'extra',label:'Additional field',type:'string'}];
      snapshot.rows=[
        {key:'first',state:'observed',values:{fromMs:'0',toMs:'10000',count:'5',extra:'Kept'}},
        {key:'empty',state:'no_data',values:{fromMs:'10000',toMs:'20000',count:'0'}},
        {key:'unknown',state:'unknown',values:{fromMs:'20000',toMs:'30000',count:null}},
        {key:'last',state:'observed',values:{fromMs:'40000',toMs:'45000',count:'18446744073709551615'}}
      ];
      const chart=document.createElement('goo-chart');chart.snapshot=window.adopt(snapshot);chart.display={kind:'time-bar'};document.body.append(chart);
      const buttons=[...chart.querySelectorAll('.scope-time-interval')],gaps=chart.querySelectorAll('.scope-time-gap').length;
      let selection;chart.addEventListener('goo-select',event=>selection=event.detail);buttons[0].click();
      const last=buttons.at(-1),first=buttons[0],positions=Number.parseFloat(last.style.left)>88&&Number.parseFloat(last.style.width)<12&&Number.parseFloat(first.style.width)>22;
      const states=chart.querySelectorAll('[data-state=no_data]').length===1&&chart.querySelectorAll('.scope-time-interval[data-state=unknown]').length===1;
      const title=last.title.includes('1970-01-01T00:00:40.000Z')&&last.title.includes('18446744073709551615')&&last.title.includes('[40000, 45000) ms');
      const noLine=!chart.querySelector('polyline'),quality=chart.textContent.includes('partial intervals')&&chart.textContent.includes('unknown upstream sampling');
      chart.display={kind:'table'};
      const table=chart.textContent.includes('Additional field')&&chart.textContent.includes('Kept')&&chart.textContent.includes('Unknown coverage')&&chart.querySelector('td>span').textContent==='1970-01-01T00:00:00.000Z'&&chart.querySelector('td').title==='0 ms since Unix epoch';
      chart.display={kind:'time-bar',requiresComplete:true};const refuses=chart.textContent.includes('Complete-only display');
      snapshot.rows[2].state='no_data';snapshot.rows[2].values.count='0';snapshot.rows.splice(3,0,{key:'gap-filled',state:'no_data',values:{fromMs:'30000',toMs:'40000',count:'0'}});
      snapshot.quality.extent='complete';snapshot.completeness={basis:'captured-query',revision:snapshot.revision,scopeId:snapshot.scope.id,aspects:['intervals']};chart.snapshot=window.adopt(snapshot);
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
    const shares=await page.evaluate(()=>{
      const raw=structuredClone(window.supplied);raw.fields=raw.fields.filter(field=>['category','count'].includes(field.key));raw.fields.find(field=>field.key==='count').unit='records';raw.fields[0].nullable=true;
      raw.rows=[{key:'large',values:{category:'api',count:'80'}},{key:'small',values:{category:'worker',count:'20'}}];raw.quality.extent='complete';raw.completeness={basis:'captured-query',revision:raw.revision,scopeId:raw.scope.id,aspects:['groups']};raw.partition={kind:'disjoint-record-groups',measure:'count',total:'100'};
      const source=window.adopt(raw),chart=document.createElement('goo-chart');chart.snapshot=source;chart.display={kind:'bar',category:'category',value:'count',scale:'population',query:'WORKER',order:'value-desc'};document.body.append(chart);
      let selected;chart.addEventListener('goo-select',event=>selected=event.detail);chart.querySelector('.scope-view-bar').click();
      const populationWidth=chart.querySelector('.scope-view-fill').style.width,shareTitle=chart.querySelector('.scope-view-bar').title;
      chart.display={...chart.display,scale:'max'};const maxWidth=chart.querySelector('.scope-view-fill').style.width;
      chart.display={...chart.display,kind:'table',scale:'population'};const shareCell=chart.querySelector('td[data-share]'),tableShare=shareCell?.textContent==='20%'&&shareCell.dataset.share==='20/100'&&shareCell.title.includes('20/100');
      chart.display={...chart.display,query:'absent'};const emptyFind=chart.textContent.includes('No matching groups')&&chart.querySelector('.chart-frame-state').dataset.state==='observed';
      chart.display={...chart.display,query:'worker'};const unchanged=chart.snapshot===source&&source.rows.length===2&&source.partition.total==='100';
      chart.snapshot=JSON.parse(JSON.stringify(source));const heldFallback=chart.display.scale==='population'&&!!chart.querySelector('table')&&!chart.querySelector('[data-share]')&&!chart.querySelector('.scope-view-fill')&&chart.textContent.includes('Population shares require verified coverage');
      chart.display={...chart.display,requiresComplete:true};const compactRefusal=chart.textContent.includes('Complete-only display requires verified coverage')&&!chart.textContent.includes('Complete captured-query')&&!chart.textContent.includes('Population shares require')&&!chart.querySelector('table')&&!chart.querySelector('.scope-view-fill');
      const empty=structuredClone(raw);empty.rows=[];empty.partition.total='0';chart.snapshot=window.adopt(empty);chart.display={kind:'bar',category:'category',value:'count',scale:'population'};const zeroRefusal=chart.textContent.includes('No captured population')&&!chart.textContent.includes('0%')&&!!chart.querySelector('table');
      const labels=structuredClone(raw);labels.rows=[null,'','Missing','Present empty string','"Missing"'].map((category,i)=>({key:'label-'+i,values:{category,count:'20'}}));chart.snapshot=window.adopt(labels);chart.display={kind:'table'};const texts=[...chart.querySelectorAll('tbody>tr>td:first-child>span')].map(cell=>cell.textContent),aria=[...chart.querySelectorAll('tbody button')].map(button=>button.getAttribute('aria-label'));
      const facet=document.createElement('goo-scope-facet');facet.snapshot=chart.snapshot;facet.field='category';document.body.append(facet);let rawValue;facet.addEventListener('goo-scope',event=>rawValue=event.detail.value);[...facet.querySelectorAll('button')].find(button=>button.textContent==='"Missing"').click();
      return {populationWidth,maxWidth,shareTitle,tableShare,emptyFind,unchanged,heldFallback,compactRefusal,zeroRefusal,distinctLabels:new Set(texts).size===5&&new Set(aria).size===5,typedValue:rawValue==='Missing',originalIntent:selected.datasetId===source.id&&selected.revision===source.revision&&selected.key==='small'};
    });
    assert.equal(shares.populationWidth,'20%');assert.equal(shares.maxWidth,'25%');assert(shares.shareTitle.includes('20/100'));
    for(const key of ['tableShare','emptyFind','unchanged','heldFallback','compactRefusal','zeroRefusal','distinctLabels','typedValue','originalIntent'])assert(shares[key],key);
    const fractions=await page.evaluate(()=>{
      const raw=structuredClone(window.supplied);raw.scope.selector.spanStatus=true;raw.fields=raw.fields.filter(field=>['category','count'].includes(field.key));raw.fields.find(field=>field.key==='count').unit='records';raw.fields.find(field=>field.key==='count').label='Captured records';
      const classes=['error','ok','unset','unknown'];for(const [key,label]of [['error','Recorded errors'],['ok','Explicit OK'],['unset','Unset status'],['unknown','Unknown status']])raw.fields.push({key:'status.'+key,label,type:'number',unit:'records',nullable:true,aggregation:{family:'span-status',kind:'count',method:'exact'}});
      raw.fields.push({key:'status.error_fraction',label:'Recorded error share',type:'rational',unit:'1',nullable:true,aggregation:{family:'ratio',kind:'fraction',method:'exact-rational',numerator:'status.error',denominator:'count'}});
      const make=(key,count,values,fraction)=>{const row={key,values:{category:key,count,'status.error_fraction':fraction},cells:{'status.error_fraction':{state:count==='0'?'no_data':'derived'}}};classes.forEach((name,i)=>{row.values['status.'+name]=values[i];row.cells['status.'+name]={state:'derived'}});return row};
      raw.rows=[make('api','100',['1','90','8','1'],'1/100'),make('unset','10',['0','0','10','0'],'0')];raw.quality.extent='complete';raw.completeness={basis:'captured-query',revision:raw.revision,scopeId:raw.scope.id,aspects:['groups']};
      const chart=document.createElement('goo-chart');chart.snapshot=raw;chart.display={kind:'bar',category:'category',value:'status.error_fraction',query:'api'};const detached=chart.snapshot.origin==='browser-held'&&chart.snapshot.quality.extent==='unknown';document.body.append(chart);
      const width=chart.querySelector('.scope-view-fill').style.width,percent=chart.textContent.includes('1%')&&chart.textContent.includes('0–100%'),context=chart.querySelector('.scope-view-bar').title;
      const ssr=document.createElement('goo-chart');ssr.setAttribute('snapshot',JSON.stringify(raw));ssr.setAttribute('display',JSON.stringify(chart.display));document.body.append(ssr);const heldSSR=ssr.snapshot.origin==='browser-held'&&ssr.querySelector('.scope-view-fill').style.width==='1%';
      const line=structuredClone(raw);line.dimensions=['x'];line.fields=line.fields.filter(field=>field.key!=='category');line.fields.unshift({key:'x',label:'Index',type:'number',unit:'1'});line.rows=line.rows.slice(0,1);line.rows[0].values.x='1';delete line.rows[0].values.category;line.completeness=undefined;delete line.completeness;line.quality.extent='unknown';chart.snapshot=line;chart.display={kind:'line',category:'x',value:'status.error_fraction'};const circle=chart.querySelector('circle'),lineFixed=Number(circle?.getAttribute('cy'))>150&&chart.querySelector('figcaption')?.textContent.includes('0–100%');
      const time=structuredClone(raw);time.shape='time-series';delete time.dimensions;time.scope.selector.bucketMs=1000;time.scope.selector.fromMs=0;time.scope.selector.toMs=3000;time.fields=time.fields.filter(field=>field.key!=='category');time.fields.find(field=>field.key==='count').nullable=true;time.fields.unshift({key:'fromMs',label:'From',type:'timestamp',unit:'ms'},{key:'toMs',label:'To',type:'timestamp',unit:'ms'});time.rows=[make('known','100',['1','90','8','1'],'1/100'),make('empty','0',['0','0','0','0'],null),make('unknown','0',['0','0','0','0'],null)];time.rows.forEach((row,i)=>{delete row.values.category;row.values.fromMs=String(i*1000);row.values.toMs=String(i*1000+1000);row.state=i===0?'observed':i===1?'no_data':'unknown';if(i===2){row.values.count=null;for(const key of [...classes.map(name=>'status.'+name),'status.error_fraction']){row.values[key]=null;row.cells[key]={state:'unknown'}}}});time.interval={fromMs:'0',toMs:'3000',bucketMs:'1000'};time.quality.extent='unknown';delete time.completeness;
      chart.snapshot=time;chart.display={kind:'time-bar',value:'status.error_fraction'};const labelNode=chart.querySelector('[data-key="known"] .scope-time-count'),labelStyle=getComputedStyle(labelNode),lum=color=>{const values=color.match(/[0-9.]+/g).slice(0,3).map(Number).map(value=>value/255).map(value=>value<=.04045?value/12.92:((value+.055)/1.055)**2.4);return values.reduce((total,value,i)=>total+value*[.2126,.7152,.0722][i],0)},ink=lum(labelStyle.color),backing=lum(labelStyle.backgroundColor),labelReadable=(Math.max(ink,backing)+.05)/(Math.min(ink,backing)+.05)>=4.5&&labelStyle.backgroundColor!=='rgba(0, 0, 0, 0)'&&labelNode.getBoundingClientRect().height>chart.querySelector('[data-key="known"] .scope-time-fill').getBoundingClientRect().height;const timeFixed=chart.querySelector('[data-key="known"] .scope-time-fill').style.height==='1%',noData=chart.querySelector('[data-key="empty"]').dataset.cellState==='no_data'&&chart.querySelector('[data-key="empty"]').title.includes('No data'),unknown=chart.querySelector('[data-key="unknown"]').dataset.cellState==='unknown'&&chart.querySelector('[data-key="unknown"]').title.includes('Unknown');
      const ordinary=structuredClone(raw);ordinary.fields.push({key:'ordinary',label:'Ordinary',type:'number',unit:'1'});ordinary.rows.forEach(row=>row.values.ordinary='.01');ordinary.rows.forEach(row=>row.values.ordinary='0.01');chart.snapshot=ordinary;chart.display={kind:'bar',category:'category',value:'ordinary',query:'api'};const unitOne=chart.textContent.includes('0.01 1')&&!chart.textContent.includes('0–100%')&&!chart.querySelector('.scope-view-value').textContent.includes('%');
      return {detached,width,percent,context,heldSSR,lineFixed,timeFixed,labelReadable,noData,unknown,unitOne};
    });
    assert.equal(fractions.width,'1%');for(const key of ['detached','percent','heldSSR','lineFixed','timeFixed','labelReadable','noData','unknown','unitOne'])assert(fractions[key],key);for(const text of ['Recorded errors: 1','Captured records: 100','Unset status: 8','Unknown status: 1'])assert(fractions.context.includes(text),text);
    const clearing=await page.evaluate(()=>{const chart=document.createElement('goo-chart');chart.snapshot=window.supplied;chart.display={kind:'table'};document.body.append(chart);const before=!!chart.querySelector('table');chart.snapshot={};return before&&!chart.querySelector('table')&&chart.snapshot===null});assert(clearing);
    assert.equal(errors.length,0,errors.join('; '));
    assert(!requests.some(name=>name.startsWith('/api/')||name==='/app.mjs'));
    console.log('Scoped components: fixed-axis status fractions/operand facts/unknowns, full-population count shares, unchanged Find/sort geometry/identity, typed state labels/refusals, exact values/rationals, human units/support/exact quantiles, per-cell gaps/estimates/signed fallback, complete-only refusal, table/bar/line/time-bar, interval gaps/UTC, typed intents, SSR/pre-upgrade,390px layout and no application IO pass');
  }finally{await browser?.close()}
})().catch(error=>{console.error(error);process.exitCode=1});
