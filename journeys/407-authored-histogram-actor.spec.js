'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path'),{randomBytes}=require('node:crypto');
const F=require('./support/histogram407');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {startBombadil}=require('./support/cross-app');
test('407 Bombadil reuses a numeric histogram table without replacing destination scope',async({request})=>{
 test.setTimeout(120000);const binary=process.env.GOO_GENERIC_BINARY;if(!binary)throw Error('Reviewed binary required');
 const root=path.resolve('artifacts/histogram407'),main=JSON.parse(await fs.readFile(path.join(root,'summary.json'),'utf8')),c=main.cases[0],dir=path.join(root,'actor-'+randomBytes(5).toString('hex'));
 await fs.mkdir(dir,{recursive:true});const report={version:'goo.histogram407ActorReport/1',sourceBank:main.fixture.token,question:c.destinationQuestion,setup:[]};let runtime,proxy;
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(binary,dir,{clickHouseURL:proxy.url});let destination;
  for(const [kind,body]of[['source-template',c.saved.request],['destination-configuration',c.destinationSetup.request]]){const response=await request.post(runtime.base+'/api/exploration-views',{data:body});expect(response.status()).toBe(201);const result=await response.json();report.setup.push({kind,request:body,response:result,status:201});if(kind==='destination-configuration')destination=result;}
  const expected=F.expected(main.raw.spans,c.destinationQuestion),url=new URL(runtime.base+'/#exploration-view='+destination.id);url.searchParams.set('expected',JSON.stringify({question:c.destinationQuestion,title:c.title,bins:expected.bins,count:expected.count}));
  report.sqlBefore=proxy.calls.length;const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:String(url),specification:'verification/bombadil/histogram407.ts',seconds:35,width:390,height:844});
  try{
   const campaign=await actor.finish({snapshot:{name:'state',version:'goo.histogram407Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete'),observed=campaign.states.find(s=>s.phase==='read'&&s.population==='applied'&&s.histogram?.snapshot);
   expect(complete).toBeTruthy();expect(observed).toBeTruthy();expect(complete.calls.filter(c=>c.path==='/api/explore')).toHaveLength(1);expect(complete.calls.filter(c=>c.path==='/api/view-templates/expand')).toHaveLength(1);expect(complete.calls.some(c=>c.method==='POST'&&['/api/exploration-views','/api/captures','/api/documents'].includes(c.path))).toBe(false);
   report.actor={snapshots:campaign.states.length,phases:[...new Set(campaign.states.map(s=>s.phase))],complete,observed,run:campaign.run};
  }finally{await actor.stop()}
  report.sqlAfter=proxy.calls.length;report.proxyCalls=proxy.calls;report.directory=dir;
 }finally{await fs.writeFile(path.join(root,'actor.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
