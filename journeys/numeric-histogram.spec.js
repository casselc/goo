const {test,expect}=require('@playwright/test');
const {randomBytes}=require('node:crypto'),fs=require('node:fs/promises'),path=require('node:path');
const {questionLink,readQuestionLink}=require('../ui/workspace');
const {openMeasures}=require('./support/explore');
const {verifyWithGo}=require('./support/cross-app');
const base='http://127.0.0.1:8088';
async function json(response){expect(response.ok(),await response.text()).toBe(true);return response.json()}
// Independent integer oracle: never use the production bin builder or JS Number
// for integer parsing, membership, bounds, or count arithmetic.
function oracle(records,type='int64'){
 const bounds=[];if(type==='int64')for(let i=63;i>=0;i--)bounds.push(-(2n**BigInt(i)));
 bounds.push(0n);for(let i=0;i<=(type==='int64'?63:64);i++)bounds.push(2n**BigInt(i));
 const bins=bounds.slice(0,-1).map((n,i)=>({from:String(n),to:String(bounds[i+1]),count:'0'}));
 const coverage={valid:'0',missing:'0',empty:'0',invalid:'0'};
 for(const {raw}of records){let state='valid',value;
  if(raw===null)state='missing';else if(raw==='')state='empty';else if(!/^(0|[1-9]\d*|-[1-9]\d*)$/.test(raw))state='invalid';
  else{value=BigInt(raw);if(value<bounds[0]||value>=bounds.at(-1))state='invalid'}
  coverage[state]=String(BigInt(coverage[state])+1n);if(state!=='valid')continue;
  const index=bounds.findIndex((n,i)=>i<bins.length&&value>=n&&value<bounds[i+1]);expect(index).toBeGreaterThanOrEqual(0);bins[index].count=String(BigInt(bins[index].count)+1n);
 }
 return {bins,coverage};
}
async function suppliedView(section,binding){
 await section.getByText('All exact numeric bins',{exact:true}).click();
 const bins=await section.locator('tbody tr[data-bin]').evaluateAll(rows=>rows.map(row=>{const cells=[...row.cells].map(c=>c.textContent),match=cells[0].match(/^\[(-?\d+), (-?\d+)\)$/);return {from:match?.[1],to:match?.[2],count:cells[1]}}));
 return {binding,bins};
}
test('174 explore exact numeric distributions, brush populations and preserve reusable evidence',async({page,request})=>{
 test.setTimeout(300000);page.setDefaultTimeout(15000);
 const environment='numeric-histogram-'+randomBytes(5).toString('hex'),toMs=Date.now(),fromMs=toMs-180000,key='work.items';
 const binding={field:{kind:'attribute',key},type:'int64',unit:'items',parse:'canonical-decimal'};
 const corpus=['-9223372036854775808','9223372036854775807','9007199254740993','-9007199254740993','-8','-7','-1','0','1','2','3','7','8','16',null,'','01','-0','+1','1e3','1.0','NaN',' 1','9223372036854775808','-9223372036854775809'];
 const records=Array.from({length:200},(_,i)=>({id:(i+1).toString(16).padStart(16,'0'),raw:corpus[i%corpus.length],resourceRaw:'7',timeMs:fromMs+1000+i*500}));
 const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource:{attributes:attrs({'service.name':'ordinary-worker','deployment.environment.name':environment,[key]:'7'})},scopeSpans:[{scope:{name:'ordinary.instrumentation'},spans:records.map(r=>({traceId:randomBytes(16).toString('hex'),spanId:r.id,name:'process work',kind:2,startTimeUnixNano:String(BigInt(r.timeMs)*1000000n),endTimeUnixNano:String(BigInt(r.timeMs+10)*1000000n),attributes:attrs(r.raw===null?{}:{[key]:r.raw}),status:{code:1}}))}]}]}})).ok()).toBe(true);
 const original={signal:'spans',environment,fromMs,toMs,limit:100,offset:0,numeric:binding,numericHistogram:{binning:'powers-of-two'}};
 await expect.poll(async()=>(await json(await request.post(base+'/api/explore',{data:{...original,numericHistogram:undefined}}))).count,{timeout:45000,intervals:[1000]}).toBe(String(records.length));
 let reads=0,plans=0;const errors=[],events=[];
 page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&/\/api\/(explore|captures|measurements|inventory|discovery|metric-discovery)$/.test(r.url()))reads++;if(r.method()==='POST'&&r.url().endsWith('/api/numeric-histograms/selection'))plans++});
 // A fixed link stages a question. Enabling the histogram is also local authoring.
 await page.goto(base+'/'+questionLink({...original,numericHistogram:undefined}));await expect(page.locator('#exploreRun')).toBeVisible();await openMeasures(page);
 const editor=page.locator('#measureComposer > .numeric-binding-editor');if(!await editor.evaluate(n=>n.open))await editor.locator(':scope > summary').click();
 await page.locator('#numericHistogram').check();expect(reads).toBe(0);expect(plans).toBe(0);
 let next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#exploreRun').click();const result=await json(await next),expected=oracle(records);
 expect(result.question).toMatchObject(original);expect(result.count).toBe('200');expect(result.rows).toHaveLength(100);expect(result.numeric).toMatchObject(expected.coverage);expect(result.numericHistogram).toEqual({binning:'powers-of-two',bins:expected.bins});
 const section=page.locator('#numericSummary .numeric-distribution'),view=await suppliedView(section,binding);expect(view.bins).toEqual(expected.bins);
 await expect(section).toContainText('Unequal numeric ranges');await expect(section).toContainText('counts, not density');
 const before=reads;await section.getByText('All exact numeric bins',{exact:true}).click();const afterStage=reads;
 // Actual native keyboard range gesture, across adjacent exact bins above 2^53.
 const fromBin=expected.bins.findIndex(b=>b.from==='9007199254740992'),toBin=fromBin+2;
 const start=section.locator('goo-time-bars button[data-key="'+fromBin+'"]');await start.focus();await start.press('Space');await start.press('Shift+ArrowRight');await page.keyboard.press('Enter');
 const tray=section.locator('.numeric-distribution-selection');await expect(tray).toBeVisible();await expect(tray).toContainText('9007199254740992');await expect(tray).toContainText('rest includes missing, empty and invalid');const afterBrush=reads;expect(afterBrush).toBe(before);expect(plans).toBe(0);
 const selectedCount=expected.bins.slice(fromBin,toBin).reduce((n,b)=>n+BigInt(b.count),0n);await expect(tray).toContainText(String(selectedCount)+' of 200 records selected');
 expect(await section.locator('goo-time-bars button[data-range-selected=true]').count()).toBe(2);
 await section.screenshot({path:'docs/journeys/screenshots/174a-exact-numeric-keyboard-range.png'});events.push({action:'brush',reads,plans,fromBin,toBin});
 // Publication captures the original full population, not the transient brush.
 await page.locator('#exploreSaveCapture').click();await page.locator('#captureDraftTitle').fill('Exact numeric work distribution');next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/captures'));await page.locator('#captureDraftConfirm').click();const saved=await json(await next);expect(saved.explore.numericHistogram).toEqual(result.numericHistogram);
 const savedView=await suppliedView(page.locator('#savednumericSummary .numeric-distribution'),binding);expect(savedView).toEqual(view);expect(await page.locator('#savednumericSummary .numeric-distribution button').count()).toBe(0);
 await page.locator('#savedexploreCreatePackage').click();await page.locator('#packageName').fill(environment);await page.locator('#packageTitle').fill('Reusable exact work distribution');await page.locator('#packageLookback').fill('5');await page.locator('#packageBuild').click();await expect(page.locator('#packageYAML')).toHaveValue(/powers-of-two/);await page.locator('#packageSave').click();await expect(page.locator('#packageStatus')).toContainText('Saved package revision');next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#packageUse').click();const reused=await json(await next);expect(reused.question.numericHistogram).toEqual(original.numericHistogram);expect(reused.numericHistogram).toEqual(result.numericHistogram);
 // Return to the exact fixed link. A reload/link does not silently execute it.
 const beforeLink=reads;expect(readQuestionLink(questionLink(original)).question).toMatchObject(original);await page.goto(base+'/'+questionLink(original));await expect(page.locator('#exploreStatus')).toContainText('No evidence loaded');expect(reads).toBe(beforeLink);
 await page.locator('#exploreShareQuestion').click();const shared=await page.locator('#exploreQuestionLink').inputValue();expect(readQuestionLink(new URL(shared).hash).question).toMatchObject(original);expect(reads).toBe(beforeLink);
 next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#exploreRun').click();await json(await next);
 const current=page.locator('#numericSummary .numeric-distribution');await current.locator('goo-time-bars button[data-key="'+fromBin+'"]').focus();await page.keyboard.press('Space');await page.keyboard.press('Shift+ArrowRight');await page.keyboard.press('Enter');
 const selectionResponse=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/numeric-histograms/selection'));next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await current.getByRole('button',{name:'Run selected population',exact:true}).click();const selection=await json(await selectionResponse),selected=await json(await next);
 expect(selection.fromBin).toBe(fromBin);expect(selection.toBin).toBe(toBin);expect(selected.question).toEqual(selection.selected);expect(selected.count).toBe(String(selectedCount));expect(selected.question.fromMs).toBe(fromMs);expect(selected.question.toMs).toBe(toMs);events.push({action:'run-selected',reads,plans});
 const rest=await json(await request.post(base+'/api/explore',{data:selection.rest}));expect(BigInt(selected.count)+BigInt(rest.count)).toBe(200n);expect(rest.numeric).toMatchObject({missing:expected.coverage.missing,empty:expected.coverage.empty,invalid:expected.coverage.invalid});
 // Compare projects the same partitions to record counts; neither query keeps a numeric measure.
 const beforeBack=reads,plansBeforeBack=plans;await page.getByRole('button',{name:'Back to question',exact:true}).click();const compareBrush=page.locator('#numericSummary .numeric-distribution');await compareBrush.locator('goo-time-bars button[data-key="'+fromBin+'"]').focus();await page.keyboard.press('Escape');await page.keyboard.press('Space');await page.keyboard.press('Shift+ArrowRight');await page.keyboard.press('Enter');await expect(compareBrush.locator('.numeric-distribution-selection')).toContainText('['+selection.from+', '+selection.to+')');const beforeComparisonStage=reads;expect(reads).toBe(beforeBack);expect(plans).toBe(plansBeforeBack);events.push({action:'back-and-brush',reads,plans});
 const countPlanResponse=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/numeric-histograms/selection'));await compareBrush.getByRole('button',{name:'Stage count comparison with rest',exact:true}).click();const countPlan=await json(await countPlanResponse);expect({fromBin:countPlan.fromBin,toBin:countPlan.toBin,from:countPlan.from,to:countPlan.to}).toEqual({fromBin:selection.fromBin,toBin:selection.toBin,from:selection.from,to:selection.to});await expect(page.locator('.population-run')).toBeVisible();expect(reads).toBe(beforeComparisonStage);
 expect(countPlan.comparison.selection.numeric).toBeUndefined();expect(countPlan.comparison.selection.numericHistogram).toBeUndefined();expect(countPlan.comparison.baseline.numeric).toBeUndefined();
 const comparisons=[];const collect=async r=>{if(r.request().method()==='POST'&&r.url().endsWith('/api/explore')&&r.ok())comparisons.push(await r.json())};page.on('response',collect);await page.locator('.population-run').click();await expect.poll(()=>comparisons.length).toBe(2);page.off('response',collect);
 const comparisonSelected=comparisons.find(r=>JSON.stringify(r.question.where)===JSON.stringify(countPlan.comparison.selection.where)),comparisonRest=comparisons.find(r=>JSON.stringify(r.question.where)===JSON.stringify(countPlan.comparison.baseline.where));expect(comparisonSelected.count).toBe(selected.count);expect(comparisonRest.count).toBe(rest.count);await expect(page.locator('.population-breakdown [id$=cohortStatus]')).toContainText(selected.count+' selected records vs '+rest.count+' baseline records');await page.locator('.population-breakdown [id$=comparisonMeasure]').selectOption('count');await expect(page.locator('.population-breakdown .cohort-table')).toContainText('ordinary-worker');events.push({action:'run-comparison',reads,plans});await page.locator('.population-breakdown [id$=comparisonPanel]').screenshot({path:'docs/journeys/screenshots/174b-selected-versus-rest-counts.png'});
 const invalidOnly=await json(await request.post(base+'/api/explore',{data:{...original,where:{numeric:binding,op:'invalid'}}}));expect(invalidOnly.numeric.valid).toBe('0');expect(invalidOnly.numericHistogram.bins.every(b=>b.count==='0')).toBe(true);
 const noData=await json(await request.post(base+'/api/explore',{data:{...original,environment:environment+'-absent'}}));expect(noData.state).toBe('no_data');expect(noData.numericHistogram.bins).toEqual(expected.bins.map(b=>({...b,count:'0'})));
 // Real invalid-only records retain coverage and show no fictitious distribution.
 await page.goto(base+'/'+questionLink(invalidOnly.question));next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#exploreRun').click();await next;await expect(page.locator('#numericSummary .numeric-distribution')).toContainText('No valid numeric values to plot');expect(await page.locator('#numericSummary goo-time-bars').count()).toBe(0);
 await page.goto(base+'/'+questionLink(noData.question));next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#exploreRun').click();await next;await expect(page.locator('#numericSummary')).toContainText('No captured records');await expect(page.locator('#numericSummary .numeric-distribution')).toContainText('No valid numeric values to plot');expect(await page.locator('#numericSummary goo-time-bars').count()).toBe(0);
 const beforeReopen=reads;await page.goto(base+'/#capture='+saved.receipt.id);await expect(page.locator('#savednumericSummary .numeric-distribution goo-time-bars')).toBeVisible();const reopened=await json(await request.get(base+'/api/captures/'+saved.receipt.id));expect(reopened).toEqual(saved);expect(reads).toBe(beforeReopen);const afterReopen=reads;
 await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.locator('#savednumericSummary .numeric-distribution').screenshot({path:'docs/journeys/screenshots/174c-mobile-frozen-numeric-distribution.png'});expect(errors).toEqual([]);
 const plan=await json(await request.post(base+'/api/telemetry-packages/draft',{data:{question:result.question,name:environment+'-oracle',title:'Numeric distribution oracle',lookbackMinutes:5}}));
 const dir=path.resolve('artifacts/numeric-histograms',environment);await fs.mkdir(dir,{recursive:true});const ledger=path.join(dir,'ledger.json');await fs.writeFile(ledger,JSON.stringify({version:1,binding,records,original:result.question,plan,result,selection,selected,rest,countPlan,comparisonSelected,comparisonRest,saved,reopened,reused,view,savedView,invalidOnly,noData,before,afterStage,afterBrush,beforeReopen,afterReopen,events},null,2));
 if(process.env.GOO_NUMERIC_HISTOGRAM_VERIFY==='1')await verifyWithGo(dir,'TestQuintNumericHistogramLiveLedger',{GOO_NUMERIC_HISTOGRAM_LEDGER:ledger});
});

test('174 numeric histogram rejects a late real plan after navigation and restores local brushing',async({page,request})=>{
 test.setTimeout(120000);const environment='numeric-histogram-lifecycle-'+randomBytes(5).toString('hex'),toMs=Date.now(),fromMs=toMs-60000,key='work.items',binding={field:{kind:'attribute',key},type:'int64',unit:'items',parse:'canonical-decimal'};
 const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource:{attributes:attrs({'service.name':'ordinary-worker','deployment.environment.name':environment})},scopeSpans:[{scope:{name:'ordinary.instrumentation'},spans:Array.from({length:120},(_,i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'ordinary lifecycle work',kind:2,startTimeUnixNano:String(BigInt(fromMs+1000+i*100)*1000000n),endTimeUnixNano:String(BigInt(fromMs+1010+i*100)*1000000n),attributes:attrs({[key]:String(2+i%6)})}))}]}]}})).ok()).toBe(true);
 const q={signal:'spans',environment,fromMs,toMs,limit:100,offset:0,numeric:binding,numericHistogram:{binning:'powers-of-two'}};
 await expect.poll(async()=>(await json(await request.post(base+'/api/explore',{data:q}))).count,{timeout:45000}).toBe('120');
 let reads=0;const errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/explore'))reads++});
 await page.goto(base+'/'+questionLink(q));let next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await page.locator('#exploreRun').click();await next;const before=reads,section=page.locator('#numericSummary .numeric-distribution'),bar=section.locator('goo-time-bars button').first();await bar.click();await expect(section.locator('.numeric-distribution-selection')).toBeVisible();
 let arrived,release;const held=new Promise(r=>arrived=r),gate=new Promise(r=>release=r);let actualPlan;
 await page.route('**/api/numeric-histograms/selection',async route=>{const upstream=await route.fetch();expect(upstream.ok()).toBe(true);actualPlan=await upstream.json();arrived();await gate;try{await route.fulfill({response:upstream})}catch{/* navigation may already abort the transport */}});
 await section.getByRole('button',{name:'Run selected population',exact:true}).click();await held;
 // A real source-free planner response is held; it must lose ownership when this visit leaves.
 await page.goto(base+'/#definitions');await page.goto(base+'/#questions');await expect(section).toBeVisible();release();await page.unrouteAll({behavior:'wait'});await expect(section.locator('goo-time-bars button').first()).toBeVisible();expect(reads).toBe(before);await expect(page.locator('#exploreCount')).toContainText('120');
 const restored=section.locator('goo-time-bars button').last();await restored.focus();await restored.press('Space');await restored.press('Enter');await expect(section.locator('.numeric-distribution-selection')).toBeVisible();expect(reads).toBe(before);
 next=page.waitForResponse(r=>r.request().method()==='POST'&&r.url().endsWith('/api/explore'));await section.getByRole('button',{name:'Run selected population',exact:true}).click();const child=await json(await next);expect(reads).toBe(before+1);expect(child.question.numericHistogram).toEqual(q.numericHistogram);expect(child.count).toBe('80');expect(errors).toEqual([]);
 const dir=path.resolve('artifacts/numeric-histograms',environment);await fs.mkdir(dir,{recursive:true});await fs.writeFile(path.join(dir,'lifecycle.json'),JSON.stringify({question:q,actualPlan,child,before,after:reads,latePlanRejected:true,restoredGesture:true},null,2));
});
