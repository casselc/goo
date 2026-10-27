const {test,expect}=require('@playwright/test'),fs=require('node:fs/promises'),path=require('node:path'),s=require('./208-container-resources-support'),{startBombadil}=require('./support/cross-app');
test.beforeEach(()=>test.skip(true,'Historical manual Resources entrance; current automatic entry and ownership are exercised in journey 210. Native/API/property/raw contracts remain active.'));
test('208 bounded shared metadata and independent memory CPU workflow',async({request})=>{
 test.skip(process.env.GOO_CONTAINER_RESOURCES_BOMBADIL!=='1','Explicit ordinary OTLP campaign');test.setTimeout(110000);
 const facts=await s.seed(request),dir=await s.write(facts,'actor-fixture.json',facts),primary=facts.points.find(p=>p.case==='primary-usage').resource;
 const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:s.base+'/#'+new URLSearchParams({resourcesEnvironment:facts.environment,resourcesPrimary:JSON.stringify(primary),resourcesCluster:JSON.stringify(facts.primary),resourcesSchema:facts.schemaURL}),specification:'verification/bombadil/container-resources.ts',seconds:30});
 try{await actor.finish();}finally{await actor.stop();}
 const entries=(await fs.readFile(path.join(dir,'bombadil/trace.jsonl'),'utf8')).trim().split('\n').map(JSON.parse),states=entries.flatMap(e=>e.snapshots.filter(s=>s.name==='state').map(s=>s.value)),phases=[...new Set(states.map(s=>s.phase))].sort((a,b)=>a-b);
 expect(phases).toEqual([0,1,2,3,4,5,6,7,8,9,10]);expect(states.at(-1)).toMatchObject({phase:10,source:10,metadata:2,histories:2,compositions:2,terminal:true});
 await fs.writeFile(path.join(dir,'bombadil/acceptance-summary.json'),JSON.stringify({phases,snapshots:states.length,final:states.at(-1)},null,2));
});
