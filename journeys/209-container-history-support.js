const shared=require('./208-container-resources-support'),fixture=require('./209-container-history-fixture'),{expect}=require('@playwright/test'),fs=require('node:fs/promises'),path=require('node:path');
async function seed(request){
 const facts=fixture.fixture(),rawOTLP={metrics:fixture.metrics(facts),traces:fixture.traces(facts),logs:fixture.logs(facts)},acknowledgements=[];
 for(const [signal,data]of Object.entries(rawOTLP))acknowledgements.push({signal,response:await shared.json(await request.post('http://127.0.0.1:4318/v1/'+signal,{data}))});
 let ready;await expect.poll(async()=>{ready=await shared.json(await request.post(shared.base+'/api/pod-discovery',{data:{environment:facts.environment,fromMs:facts.fromMs,toMs:facts.toMs}}));return ready.observations?.reduce((sum,r)=>sum+r.families.reduce((n,f)=>n+Number(f.count),0),0);},{timeout:60000}).toBe(facts.points.length+facts.records.length);
 shared.assertInventory(facts,ready);return {...facts,rawOTLP,acknowledgements,ready};
}
async function write(facts,name,data){const dir=path.resolve('artifacts/container-history',facts.environment);await fs.mkdir(dir,{recursive:true});await fs.writeFile(path.join(dir,name),JSON.stringify(data,null,2));console.log('209 artifact:',path.join(dir,name));return dir;}
const history=(panel,family)=>panel.locator('.container-resource-history[data-history-family="'+family+'"]');
async function historySnapshot(target){return target.evaluate(e=>({text:e.textContent,focus:e.querySelector('select').value,plot:e.querySelector('goo-value-timeline').data,selectedKey:e.querySelector('goo-value-timeline').selectedKey,marks:[...e.querySelectorAll('.value-timeline-mark')].map(n=>({...n.dataset,label:n.getAttribute('aria-label'),pressed:n.getAttribute('aria-pressed'),left:n.style.left,top:n.style.top,width:n.style.width})),statLists:[...e.querySelectorAll('goo-stat-list')].map(n=>n.items),context:[...e.querySelectorAll('a')].map(n=>({text:n.textContent,href:n.href})),range:e.querySelector('goo-time-range').data}));}
function linkedQuestion(href){return JSON.parse(decodeURIComponent(new URL(href,shared.base).hash.slice(6))).record;}
function assertContext(facts,resource,signal,question,range){
 expect(question.environment).toBe(facts.environment);expect(question.signal).toBe(signal);
 expect(question.fromMs).toBe(Number(BigInt(range.fromNS)/1000000n));expect(question.toMs).toBe(Number((BigInt(range.toNS)+999999n)/1000000n));
 const keys=['container.id','k8s.pod.uid','k8s.container.name','k8s.cluster.uid','k8s.cluster.name','k8s.namespace.name'];
 expect(question.where.all).toHaveLength(9);
 for(const k of [...keys,'deployment.environment.name'])expect(question.where.all).toContainEqual({field:{kind:'resource',key:k},op:'eq',value:resource[k]});
 expect(question.where.all).toContainEqual({field:{kind:'column',key:'timestamp_ns'},op:'gte',value:range.fromNS});expect(question.where.all).toContainEqual({field:{kind:'column',key:'timestamp_ns'},op:'lte',value:String(BigInt(range.toNS)-1n)});
 return facts.records.filter(r=>r.family===signal&&BigInt(r.timeNS)>=BigInt(range.fromNS)&&BigInt(r.timeNS)<BigInt(range.toNS)&&keys.every(k=>r.resource[k]===resource[k])).map(r=>r.id).sort();
}
async function captureBytes(facts,id,name){const source=await shared.captureBytes(facts,id,name),old=path.resolve('artifacts/container-resources',facts.environment,name),dir=path.resolve('artifacts/container-history',facts.environment);await fs.mkdir(dir,{recursive:true});await fs.rename(old,path.join(dir,name));return source;}
module.exports={...shared,seed,write,history,historySnapshot,linkedQuestion,assertContext,captureBytes};
