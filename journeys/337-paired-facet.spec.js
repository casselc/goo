const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const {randomBytes}=require('node:crypto');
const {startBombadil,verifyWithGo}=require('./support/cross-app');
const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
const methodField={kind:'attribute',key:'http.request.method'};
const sourcePaths=new Set(['/api/tail','/api/explore','/api/explore/page','/api/record-aggregates','/api/explore/counted-values','/api/measurements','/api/inventory','/api/discovery','/api/environment-scopes','/api/environment-scopes/records','/api/population/plan']);

test('337 narrow both fresh comparison windows with an ordinary OTel categorical facet',async({page,request})=>{
 test.setTimeout(300000);page.setDefaultTimeout(20000);
 const base=process.env.GOO_BASE_URL||'http://127.0.0.1:8088',dir=path.resolve('artifacts/paired-facet337'),token=randomBytes(4).toString('hex');
 await fs.mkdir(dir,{recursive:true});await page.setViewportSize({width:1280,height:1000});
 const calls=[],errors=[],ledger=[],runs=[];
 page.on('pageerror',e=>errors.push(e.message));
 page.on('request',r=>{const pathname=new URL(r.url()).pathname;if(pathname.startsWith('/api/')){let body;try{body=r.postDataJSON()}catch{}calls.push({path:pathname,method:r.method(),body})}});
 const n=p=>calls.filter(c=>c.path===p&&c.method==='POST').length;
 const source=()=>calls.filter(c=>sourcePaths.has(c.path)).length;
 const counters=()=>({reads:n('/api/explore')+n('/api/record-aggregates')+n('/api/explore/page'),lookups:n('/api/explore/counted-values'),compares:n('/api/compare'),saves:n('/api/exploration-views'),encodes:n('/api/view-artifacts/encode')});
 const pop=()=>page.getByRole('region',{name:'Population workspace',exact:true}).filter({visible:true});
 const facet=()=>pop().locator('[data-composition-role=facet] goo-categorical-facet');
 const group=()=>pop().locator('goo-group-count-comparison');
 const tray=()=>pop().getByRole('region',{name:'Facet selection',exact:true});
 const post=p=>page.waitForResponse(r=>r.request().method()==='POST'&&new URL(r.url()).pathname===p);
 const open=async box=>{if(!await box.evaluate(e=>e.open))await box.locator(':scope>summary').click()};
 const close=async box=>{if(await box.evaluate(e=>e.open))await box.locator(':scope>summary').click()};
 const settle=()=>page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
 let budget,recording=false;
 const mark=action=>{if(!recording)return;const now=counters(),out={action,preserved:true};for(const key of Object.keys(now))out[key]=now[key]-budget[key];out.source=out.reads+out.lookups;ledger.push(out)};

 async function emit(signal,environment,clock){
  const resources=[];
  for(const [service,current,previous]of [
   ['api337',[['GET',46],['POST',20],['',1],[null,3]],[['GET',5],['POST',3],[null,1]]],
   ['worker337',[['GET',29],['POST',7],['',1]],[['GET',2],['POST',1],['',1]]],
  ]){
   let serial=0;
   const events=[...current.flatMap(([method,count])=>Array.from({length:count},()=>({method,at:clock-5000+serial++}))),...previous.flatMap(([method,count])=>Array.from({length:count},()=>({method,at:clock-90000+serial++})))];
   const resource={attributes:attrs({'service.name':service,'service.namespace':'http337',...(service==='worker337'?{'deployment.environment':environment}:{'deployment.environment.name':environment})})};
   const attributes=method=>attrs({'http.route':method==='POST'?'/write':'/read',...(method===null?{}:{'http.request.method':method})});
   if(signal==='logs')resources.push({resource,scopeLogs:[{scope:{name:'paired.facet337'},logRecords:events.map(({method,at},i)=>({timeUnixNano:String(BigInt(at)*1000000n),severityText:'INFO',body:{stringValue:'337 '+service+' '+i},attributes:attributes(method)}))}]});
   else resources.push({resource,scopeSpans:[{scope:{name:'paired.facet337'},spans:events.map(({method,at},i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'337 '+service+' '+i,kind:2,startTimeUnixNano:String(BigInt(at)*1000000n),endTimeUnixNano:String(BigInt(at+1)*1000000n),attributes:attributes(method)}))}]});
  }
  const response=await request.post('http://127.0.0.1:4318/v1/'+(signal==='logs'?'logs':'traces'),{data:signal==='logs'?{resourceLogs:resources}:{resourceSpans:resources}});expect(response.ok()).toBe(true);
  await expect.poll(async()=>{const response=await request.post(base+'/api/explore',{data:{signal,environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:environment}},fromMs:clock-60000,toMs:clock+1000,limit:1,offset:0}});return response.ok()&&(await response.json()).count},{timeout:45000}).toBe('107');
 }
 async function enter(signal,environment){
  await page.goto(base+'/');const activity=page.locator('.stream-entry');await activity.getByLabel('Live signal',{exact:true}).selectOption(signal);const adoption=activity.locator('.environment-adoption');await open(adoption);
  let pending=post('/api/environment-scopes');await adoption.getByRole('button',{name:'Read environment scopes',exact:true}).click();await pending;
  pending=post('/api/environment-scopes/records');await adoption.locator('tbody tr').filter({hasText:'Named environment: "'+environment+'"'}).getByRole('button',{name:new RegExp('^Preview named '+signal+' · ')}).first().click();await pending;
  let before=source();await adoption.getByRole('button',{name:'Stage in Activity',exact:true}).click();expect(source()).toBe(before);
  await activity.getByRole('button',{name:'Start stream',exact:true}).click();await expect(activity.locator('.live-tail tbody>tr')).toHaveCount(100);await activity.getByRole('button',{name:'Pause stream',exact:true}).click();await open(activity.locator('.entry-investigate-options'));
  before=source();await activity.getByRole('button',{name:'Compare previous window by service',exact:true}).click();await expect(pop()).toHaveAttribute('data-phase','staged');expect(source()).toBe(before);
  await open(pop().locator('.population-view-settings'));const editor=pop().locator('.population-facet-settings');await open(editor);await editor.getByLabel('Facet field',{exact:true}).fill('attribute:http.request.method');await editor.getByLabel('Facet title',{exact:true}).fill('HTTP methods');
  before=source();await editor.getByRole('button',{name:'Add facet',exact:true}).click();expect(source()).toBe(before);await close(pop().locator('.population-view-settings'));
 }
 async function save(title){
  const box=pop().locator('.population-save');await open(box);await box.getByLabel('Exploration view title',{exact:true}).fill(title);const pending=post('/api/exploration-views'),before=source();await box.getByRole('button',{name:'Save exploration view',exact:true}).click();const saved=await(await pending).json();expect(source()).toBe(before);await close(box);return saved;
 }
 async function runPair(selectedCount,previousCount,expected){
  const before=source(),a=post('/api/explore'),b=post('/api/record-aggregates'),c=post('/api/compare');await pop().locator('.population-run').click();
  const selected=await(await a).json(),previous=await(await b).json(),comparison=await(await c).json();await expect(pop()).toHaveAttribute('data-phase','applied');await expect(pop().locator('.population-comparison-pair')).toBeVisible();expect(source()).toBe(before+2);
  expect(selected.count).toBe(String(selectedCount));expect(previous.count).toBe(String(previousCount));if(expected){expect(selected.question).toEqual(expected.selected);expect(previous.question).toEqual(expected.previous)}
  const display=pop().getByLabel('Comparison view',{exact:true});if(await display.inputValue()!=='groups')await display.selectOption('groups');await expect(group()).toBeVisible();expect(comparison.viewComparison.changes.count.value).toBe(String(selectedCount-previousCount));
  return{selected,previous,comparison};
 }
 async function values(total,groups,baselineTotal,baselineGroups){
  const before=source(),responses=[];const collect=r=>{if(r.request().method()==='POST'&&new URL(r.url()).pathname==='/api/explore/counted-values')responses.push(r)};page.on('response',collect);
  try{await facet().getByRole('button',{name:'Read values',exact:true}).click();await expect.poll(()=>responses.length).toBe(2);const data=await Promise.all(responses.map(async response=>{expect(response.ok()).toBe(true);return response.json()}));await expect(facet()).toHaveAttribute('data-state','ready');expect(source()).toBe(before+2);
   const pair=await group().evaluate(e=>e.comparison),selected=data.find(value=>value.question.fromMs===pair.selection.scope.selector.fromMs),baseline=data.find(value=>value.question.fromMs===pair.baseline.scope.selector.fromMs);expect(selected).toBeTruthy();expect(baseline).toBeTruthy();
   for(const [result,scope,count,expected]of [[selected,pair.selection.scope.selector,total,groups],[baseline,pair.baseline.scope.selector,baselineTotal,baselineGroups]]){expect(result.count).toBe(String(count));expect(result.question.groupBy).toEqual([methodField]);for(const key of ['signal','environmentSelector','fromMs','toMs','where'])expect(result.question[key]).toEqual(scope[key]);expect(result.groups).toEqual(expect.arrayContaining(expected));expect(result.groups).toHaveLength(expected.length);for(const kind of ['count','groups']){expect(result.reads[kind].basis).toMatch(/^[a-f0-9]{64}$/);expect(Number.isFinite(Date.parse(result.reads[kind].readAt))).toBe(true)}}
   expect(await facet().evaluate(e=>e.data.values.total)).toBe(String(total));expect(await facet().evaluate(e=>e.data.baseline.total)).toBe(String(baselineTotal));return{selected,baseline};
  }finally{page.off('response',collect)}
 }
 async function stage(mode,value,action){
  const old=await group().evaluate(e=>e.comparison),before=source(),label=value===null?'Missing key':value===''?'Present empty string':JSON.stringify(value);await pop().locator('.population-run').evaluate(e=>window.pendingRun337={button:e,handler:e.onclick});
  await facet().getByRole('button',{name:(mode==='include'?'Include ':'Exclude ')+'attribute:http.request.method '+label,exact:true}).click();await expect(tray()).toContainText('Both windows');await expect(pop().locator('.population-run')).toBeDisabled();const pendingText=await tray().textContent();await page.evaluate(async()=>{const old=window.pendingRun337;if(old.handler)await old.handler.call(old.button)});await settle();expect(await tray().textContent()).toBe(pendingText);expect(source()).toBe(before);expect(await group().evaluate(e=>e.comparison)).toEqual(old);await expect(tray().getByRole('button',{name:'Apply facet selection',exact:true})).toHaveCount(0);await expect(tray().getByRole('button',{name:'Compare with rest',exact:true})).toHaveCount(0);mark(action);
  return old;
 }
 async function filterBoth(action){const before=source();await tray().getByRole('button',{name:'Filter both windows',exact:true}).click();await expect(pop()).toHaveAttribute('data-phase','staged');await expect(pop().locator('.section-heading h3')).toContainText('Filter both windows');expect(source()).toBe(before);mark(action)}
 const exclude=original=>({...structuredClone(original),offset:0,where:{all:[...(original.where?[structuredClone(original.where)]:[]),{any:[{field:methodField,op:'missing'},{field:methodField,op:'ne',value:'GET'}]}]}});
 const include=original=>({...structuredClone(original),offset:0,where:{all:[structuredClone(original.where),{field:methodField,op:'eq',value:'POST'}]}});

 let actorView;
 for(const signal of ['logs','spans']){
  recording=false;const environment='paired337-'+token+'-'+signal;await emit(signal,environment,Date.now());await enter(signal,environment);
  const originalSaved=await save('Original paired facet '+signal+' '+token);expect(originalSaved.composition.children.find(c=>c.facet).facet.field).toEqual(methodField);if(signal==='logs'){actorView=originalSaved;budget=counters();recording=true;mark('owner')}
  const original=await runPair(107,13);expect(original.selected.rows).toHaveLength(100);expect(original.selected.question.environmentSelector).toEqual({version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'named',value:environment}});mark('runParent');
  const records=pop().locator('[data-composition-role=records]'),beforeRecordGesture=source();await pop().getByLabel('Records view',{exact:true}).selectOption('record-list');await records.getByRole('button',{name:'Inspect record',exact:true}).first().click();const foreignField=records.getByRole('button',{name:'Include attribute:http.route from record',exact:true});await expect(foreignField).toBeEnabled();await foreignField.click();await expect(tray()).toBeHidden();await expect(records.locator('.ui-inspector-body .attribute-feedback')).toHaveText('Not staged.');await expect(pop().locator('.population-status')).toContainText('Use attribute:http.request.method to filter both windows.');await records.getByRole('button',{name:'Close selected item details',exact:true}).click();expect(source()).toBe(beforeRecordGesture);

  const initialValues=await values(107,[{g0:'GET',count:'75'},{g0:'POST',count:'27'},{g0:'',count:'2'},{g0:null,count:'3'}],13,[{g0:'GET',count:'7'},{g0:'POST',count:'4'},{g0:'',count:'1'},{g0:null,count:'1'}]);expect(initialValues.selected.question.fromMs).toBe(original.selected.question.fromMs);expect(initialValues.selected.question.toMs).toBe(original.selected.question.toMs);mark('valuesParent');
  if(signal==='logs')await page.screenshot({path:'docs/journeys/screenshots/337-pair-and-selected-values.png'});
  await stage('exclude','GET','stageExclude');if(signal==='logs')await page.screenshot({path:'docs/journeys/screenshots/337-filter-both-windows.png'});await filterBoth('applyExclude');
  const excludedExpected={selected:exclude(original.selected.question),previous:exclude(original.previous.question)},excluded=await runPair(32,6,excludedExpected);mark('runExcluded');
  const excludedValues=await values(32,[{g0:'POST',count:'27'},{g0:'',count:'2'},{g0:null,count:'3'}],6,[{g0:'POST',count:'4'},{g0:'',count:'1'},{g0:null,count:'1'}]);expect(excludedValues.selected.question.where).toEqual(excludedExpected.selected.where);mark('valuesExcluded');
  await stage('include','POST','stageInclude');await filterBoth('applyInclude');const includedExpected={selected:include(excludedExpected.selected),previous:include(excludedExpected.previous)},included=await runPair(27,4,includedExpected);mark('runIncluded');
  const find=()=>group().getByLabel('Find compared groups',{exact:true}),side=()=>group().locator('tr[data-group-key="1api337"] td[data-side=baseline] button');await find().fill('api337');const held=await group().evaluate(e=>e.comparison),presentation=await group().evaluate(e=>e.presentation),beforeChild=source();await side().click();await expect(pop()).toHaveAttribute('data-phase','staged');await expect(pop().locator('.section-heading h3')).toContainText('Previous window group "api337"');expect(source()).toBe(beforeChild);mark('openGroup');
  const childRead=post('/api/explore/page');await pop().locator('.population-run').click();const child=await(await childRead).json();await expect(pop()).toHaveAttribute('data-phase','applied');expect(source()).toBe(beforeChild+1);expect(child.count).toBe('3');expect(child.rows).toHaveLength(3);await expect(pop().locator('[data-composition-role=comparison], [data-composition-role=aggregate], [data-composition-role=timeline]')).toHaveCount(0);await expect(facet()).toContainText('Applied fixed scope');expect(child.question.fromMs).toBe(includedExpected.previous.fromMs);expect(child.question.toMs).toBe(includedExpected.previous.toMs);expect(child.question.environmentSelector).toEqual(includedExpected.previous.environmentSelector);expect(child.question.where).toEqual({all:[includedExpected.previous.where,{field:{kind:'column',key:'service'},op:'eq',value:'api337'}]});
  for(const row of child.rows){expect(row.ServiceName).toBe('api337');expect(row.Attributes['http.request.method']).toBe('POST');const at=BigInt(row.timestamp_ns);expect(at>=BigInt(child.question.fromMs)*1000000n&&at<BigInt(child.question.toMs)*1000000n).toBe(true)}mark('readGroup');
  const beforeBack=source();await page.getByRole('button',{name:'Back to parent population',exact:true}).click();await expect(pop()).toHaveAttribute('data-phase','applied');expect(source()).toBe(beforeBack);expect(await group().evaluate(e=>e.comparison)).toEqual(held);expect(await group().evaluate(e=>e.presentation)).toEqual(presentation);await expect(find()).toHaveValue('api337');await expect(side()).toBeFocused();mark('backGroup');
  const saved=await save('Narrowed paired facet '+signal+' '+token);expect(saved.queries).toEqual([includedExpected.selected]);expect(saved.composition.children.find(c=>c.comparison).comparison.baselineQuery).toEqual(includedExpected.previous);expect(saved.composition.children.find(c=>c.facet).facet.field).toEqual(methodField);expect(JSON.stringify(saved)).not.toContain('sourceBasis');mark('save');
  const exporting=pop().locator('.population-artifact-export');await open(pop().locator('.population-save'));await open(exporting);await exporting.getByLabel('Artifact namespace',{exact:true}).fill('team337');await exporting.getByLabel('Artifact name',{exact:true}).fill('paired-methods-'+signal);await exporting.getByLabel('Artifact revision',{exact:true}).fill('v1');const beforeExport=source(),encoding=post('/api/view-artifacts/encode');await exporting.getByRole('button',{name:'Export YAML',exact:true}).click();const encoded=await(await encoding).json();expect(encoded.artifact.view.queries).toEqual(saved.queries);expect(encoded.artifact.view.composition).toEqual(saved.composition);expect(await exporting.getByLabel('Exported view YAML',{exact:true}).inputValue()).toBe(encoded.yaml);expect(source()).toBe(beforeExport);mark('encodeArtifact');
  const beforeOpen=source();await page.goto(base+'/#exploration-view='+saved.id);await expect(pop()).toHaveAttribute('data-phase','staged');expect(source()).toBe(beforeOpen);mark('reopen');const reopened=await runPair(27,4,includedExpected);mark('runReopened');await expect(facet()).toHaveAttribute('data-state','unloaded');
  let supplemental=null;
  if(signal==='logs'){
   await page.screenshot({path:'docs/journeys/screenshots/337-reopened-common-scope.png'});
   // Supplemental owner-renewal check after the observed model's terminal step.
   // It performs two explicit operand value lookups; all following gestures are local.
   recording=false;const supplementBefore=source();await values(27,[{g0:'POST',count:'27'}],4,[{g0:'POST',count:'4'}]);await facet().getByLabel('Search loaded facet values',{exact:true}).fill('POST');const facetBefore=await facet().evaluate(e=>e.data),pairBefore=await group().evaluate(e=>e.comparison);await facet().getByRole('button',{name:'Exclude attribute:http.request.method "POST"',exact:true}).evaluate(e=>window.retiredFacet337=e);
   await stage('include','POST','supplementStage');const localBefore=source();await filterBoth('supplementOpen');await page.getByRole('button',{name:'Back to parent population',exact:true}).click();expect(source()).toBe(localBefore);await expect(facet().getByLabel('Search loaded facet values',{exact:true})).toHaveValue('POST');expect(await facet().evaluate(e=>e.data.values)).toEqual(facetBefore.values);expect(await group().evaluate(e=>e.comparison)).toEqual(pairBefore);const pendingText=await tray().textContent();await facet().evaluate(e=>{e.append(window.retiredFacet337);window.retiredFacet337.click();window.retiredFacet337.remove()});await settle();expect(await tray().textContent()).toBe(pendingText);expect(source()).toBe(localBefore);await facet().getByRole('button',{name:'Clear staged facet',exact:true}).click();await expect(pop().locator('.population-run')).toBeEnabled();expect(source()).toBe(localBefore);expect(source()-supplementBefore).toBe(2);supplemental={explicitOperandValueReads:2,localGestureReads:0,findAndValuesRetained:true,oldReattachedControlRefused:true};
  }
  if(signal==='spans'){await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/337-paired-facet-narrow.png'});await page.setViewportSize({width:1280,height:1000})}
  runs.push({signal,environment,originalSaved,original,initialValues,excluded,excludedValues,included,child,saved,encoded,reopened,supplemental});
 }
 await fs.writeFile(path.join(dir,'ledger.json'),JSON.stringify(ledger,null,2));await verifyWithGo(dir,'TestQuintPairedFacetObservedBrowserLedger337',{GOO_PAIRED_FACET_LEDGER:path.join(dir,'ledger.json')});
 if(process.env.GOO_PAIRED_FACET_BOMBADIL==='1'){
  const actorDir=path.join(dir,'bombadil-'+token),actor=await startBombadil({dir:actorDir,url:base+'/#exploration-view='+actorView.id,specification:'verification/bombadil/paired-facet.ts',seconds:35,width:1280,height:1000});
  try{const {run,states}=await actor.finish({snapshot:{name:'state',version:'goo.pairedFacet337'}}),final=states.at(-1);expect(states.some(s=>s.phase===7&&s.verified===4)).toBe(true);expect(final).toMatchObject({phase:7,verified:4,source:6,lookups:2,compares:2,writes:0,immutable:true,overflow:false});await fs.writeFile(path.join(dir,'campaign-summary.json'),JSON.stringify({seconds:35,snapshots:states.length,phases:[...new Set(states.map(s=>s.phase))],run,final},null,2))}finally{await actor.stop()}
 }
 expect(errors).toEqual([]);expect(calls.filter(c=>c.method==='POST'&&['/api/captures','/api/kept-observations/records','/api/documents'].includes(c.path))).toEqual([]);await fs.writeFile(path.join(dir,'summary.json'),JSON.stringify({version:'goo.paired-facet337/1',realOTLP:true,genuineUISaves:true,noSeededDefinitions:true,runs,ledger,calls,errors},null,2));
});
