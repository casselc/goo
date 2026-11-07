const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path');
const {createHash}=require('node:crypto'),{execFileSync}=require('node:child_process');
const {startRuntime}=require('./191-generic-runtime');
const {runLargeFacetBenchmark}=require('./support/large-facet-benchmark');

test('343 supplied large-facet mount and local Find diagnostic',async({page,request})=>{
 test.setTimeout(240000);
 const phase=process.env.GOO_FACET_PHASE||'baseline';if(!/^[a-z0-9-]{1,40}$/.test(phase))throw Error('Use a bounded diagnostic phase.');
 let runtime;const dir=path.resolve('artifacts/large-facets343'),sourceHash={},calls=[],errors=[],cases=[];
 try{
 if(process.env.GOO_GENERIC_BINARY)runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,path.join(dir,'runtime-'+phase));
 const base=runtime?.base||process.env.GOO_BENCHMARK_BASE||'http://127.0.0.1:8088',sourceRef=process.env.GOO_FACET_SOURCE_REF||(phase==='baseline'?'b7adf8d':null);
 const sourceFile=name=>sourceRef?execFileSync('git',['show',sourceRef+':ui/'+name]):fs.readFile(path.resolve('ui',name));
 for(const name of ['components.mjs','categorical-facet.mjs','categorical-pair.mjs','components.css']){const local=createHash('sha256').update(await sourceFile(name)).digest('hex'),response=await request.get(base+'/'+name);expect(response.ok()).toBe(true);const served=createHash('sha256').update(await response.body()).digest('hex');expect(served,name+' served source parity').toBe(local);sourceHash[name]=local;}
 page.on('request',r=>{const url=new URL(r.url());if(url.pathname.startsWith('/api/'))calls.push({method:r.method(),path:url.pathname})});page.on('pageerror',e=>errors.push(e.message));
 for(const viewport of [{width:1280,height:1000},{width:390,height:844}]){
  await page.setViewportSize(viewport);await page.goto('about:blank');const response=await page.goto(base+'/ui-components',{waitUntil:'load'});expect(response.status()).toBe(200);await expect(page).toHaveTitle('goo · UI components');await page.waitForFunction(()=>[...document.styleSheets].some(sheet=>sheet.href?.endsWith('/components.css')&&sheet.cssRules.length>0));
  for(const labelBytes of [384,2048])for(const size of [1000,2000]){
   const before=calls.length,result=await runLargeFacetBenchmark(page,{size,labelBytes,iterations:4,bounded:phase==='final'});
   expect(result.styleFingerprint.display).toBe('grid');expect(result.styleFingerprint.rowDisplay).toBe('grid');expect(result.styleFingerprint.fontSize).toBe('12px');expect(result.rowCount).toBe(phase==='final'?100:size);expect(result.exactSuffixIdentity).toBe(true);expect(result.missingDistinctFromEmpty).toBe(true);expect(result.receiptStable).toBe(true);
   if(size===2000){expect(result.selectedCounts).toEqual(['1','0']);expect(result.previousOnlyCounts).toEqual(['0','2']);expect(result.selectedScales).toEqual(['2','2']);expect(result.selectedWidths).toEqual(['50%','0%']);}
   for(const sample of result.observations){expect(sample.focused).toBe(true);expect(sample.caret).toBe(sample.query.length);const fullMatches=sample.query===''?size:sample.query==='not-an-observed-category'?0:1;expect(sample.visible).toBe(phase==='final'?Math.min(100,fullMatches):fullMatches);expect(sample.matches).toBe(fullMatches+' of '+size+' loaded values');if(phase==='final')expect(sample.refreshPasses).toBe(0);else expect(sample.refreshPasses).toBeGreaterThan(0);}
   expect(calls.length).toBe(before);cases.push({...result,sourceReads:0});
  }
 }
 expect(errors).toEqual([]);expect(calls).toEqual([]);await fs.mkdir(dir,{recursive:true});
 const receipt={phase,sourceRef,styled:true,pageRoute:'/ui-components',stylesheetVerified:true,commit:execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),sourceHash,recordedAt:new Date().toISOString(),kind:'supplied-component diagnostic; synthetic input; live source semantics are checked separately',cases,sourceCalls:calls,errors};
 const target=path.join(dir,'benchmark-'+phase+'.json');try{const previous=JSON.parse(await fs.readFile(target,'utf8'));if(!previous.styled)await fs.writeFile(path.join(dir,'benchmark-'+phase+'-invalid-unstyled.json'),JSON.stringify({...previous,invalid:true,invalidReason:'Historical runner used /components.html (404) without stylesheet verification. Timing and layout results are invalid.'},null,2)+'\n');}catch(error){if(error.code!=='ENOENT')throw error;}
 await fs.writeFile(target,JSON.stringify(receipt,null,2)+'\n');
 console.info('343 supplied facet benchmark',JSON.stringify(cases.map(value=>({viewport:value.viewport,size:value.size,labelBytes:value.labelBytes,mountSyncMs:value.mountSyncMs,mountPaintMs:value.mountPaintMs,domElements:value.domElements,refreshPasses:[...new Set(value.observations.map(sample=>sample.refreshPasses))],maxInputSyncMs:Math.max(...value.observations.map(sample=>sample.syncMs)),maxInputPaintMs:Math.max(...value.observations.map(sample=>sample.paintMs))}))));
 }finally{await runtime?.stop();}
});
