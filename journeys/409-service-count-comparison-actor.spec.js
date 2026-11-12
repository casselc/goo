'use strict';
const {test,expect}=require('@playwright/test');
const fs=require('node:fs/promises'),path=require('node:path'),{randomBytes}=require('node:crypto');
const {startRuntime}=require('./191-generic-runtime');
const {startSourceFaultProxy}=require('./support/source-fault-proxy');
const {startBombadil}=require('./support/cross-app');
const F=require('./support/service-comparison409');
test('409 Bombadil discovers service comparison and returns without reacquiring its parent',async({request})=>{
 test.setTimeout(120000);const binary=process.env.GOO_GENERIC_BINARY;if(!binary)throw Error('Reviewed binary required');const root=path.resolve('artifacts/service-comparison409'),main=JSON.parse(await fs.readFile(path.join(root,'summary.json'),'utf8')),dir=path.join(root,'actor-'+randomBytes(5).toString('hex'));await fs.mkdir(dir,{recursive:true});const report={version:'goo.serviceComparison409ActorReport/1',sourceBank:main.fixture.token};let runtime,proxy;
 try{
  proxy=await startSourceFaultProxy();runtime=await startRuntime(binary,dir,{captureStore:path.join(root,'capture-store'),clickHouseURL:proxy.url});
  const url=new URL(runtime.base+'/#capture='+main.capture.receipt.id);url.searchParams.set('expected',JSON.stringify({service:main.fixture.service,namespace:main.fixture.namespace,selected:main.expected.selected,baseline:main.expected.baseline}));
  const actor=await startBombadil({dir:path.join(dir,'bombadil'),url:String(url),specification:'verification/bombadil/service-comparison409.ts',seconds:45,width:390,height:844});
  try{const campaign=await actor.finish({snapshot:{name:'state',version:'goo.serviceComparison409Actor/1'}}),complete=campaign.states.find(s=>s.phase==='complete'),observed=campaign.states.find(s=>s.observed&&s.phase==='running');expect(complete).toBeTruthy();expect(observed).toBeTruthy();expect(complete.forbidden).toEqual([]);expect(complete.preserved).toBe(true);expect(complete.compares).toBe(1);F.assertResult(expect,main.raw,main.expected.selected,observed.observed.selected.response,{records:true});F.assertResult(expect,main.raw,main.expected.baseline,observed.observed.baseline.response);report.actor={snapshots:campaign.states.length,phases:[...new Set(campaign.states.map(s=>s.phase))],complete,observed,run:campaign.run};}finally{await actor.stop()}
 }finally{report.proxyCalls=proxy?.calls;await fs.writeFile(path.join(root,'actor.json'),JSON.stringify(report,null,2));await runtime?.stop();await proxy?.stop()}
});
