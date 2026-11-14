'use strict';
const fs=require('node:fs/promises'),path=require('node:path');
const {requestLog}=require('./gotogether-driver-upstream.cjs');
const {verifyWithGo}=require('./cross-app');

// Transport completion and DOM adoption remain separate observations.
async function startShadow(page,dir){
 const file=path.join(dir,'gotogether-requests.jsonl');let read=0,epoch=0,closed=false;
 const log=await requestLog(page,{path:file,header:{recorder:'goo439',actor:'acquisition',process:'browser',thread:'page'},
  classify(req){return req.method==='POST'&&new URL(req.url).pathname==='/api/metric-observations'?{action:'read',args:{read:++read,epoch}}:null},
  answer(res,call){return {...call.args}},
 });
 async function show(phase){
  const observed=await page.evaluate(()=>{
   const panel=document.querySelector('[data-raw-observations]');
   const receipt=label=>[...(panel?.querySelectorAll('dt')||[])].find(e=>e.textContent===label)?.nextElementSibling?.textContent||'';
   const status=[...(panel?.querySelectorAll('[role=status]')||[])].at(-1)?.textContent||'';
   return {bank:receipt('Source read UTC'),clock:receipt('Fixed window'),question:JSON.stringify({service:document.querySelector('#metricService')?.value||'',environment:document.querySelector('#metricEnvironment')?.value||'',instrument:document.querySelector('#metricInstruments button[aria-label^="Read observations"]')?.getAttribute('aria-label')||''}),
    page:Number(status.match(/^(\d+)[–-]/)?.[1]||0),find:panel?.querySelector('input[type=search]')?.value||'',
    detail:panel?.querySelector('[aria-label="Selected item details"]')?.getClientRects().length?panel.querySelector('[aria-label="Selected item details"]').textContent:'',
    visible:!!panel?.getClientRects().length};
  });
  await log.step('show',{phase,...observed,epoch},{ref:'actual DOM after '+phase});
 }
 return {depart(){epoch++},show,
  async finish(){
   // The caller first drains its source meter and response-body observers. Two
   // animation frames let the completed fetch handlers publish their actual DOM.
   await page.evaluate(()=>new Promise(r=>requestAnimationFrame(()=>requestAnimationFrame(r))));
   await show('final');
   // close drains the adapter's own answer-body tasks; no EOF implies settle.
   await log.close();closed=true;
   const lines=(await fs.readFile(file,'utf8')).trim().split('\n').map(JSON.parse),pending=new Set();
   for(const line of lines){if(line.event==='request')pending.add(line.req);if(line.event==='response')pending.delete(line.req)}
   if(pending.size)throw Error('Shadow cannot settle with unanswered owned calls: '+[...pending]);
   const last=lines.at(-1).time;
   await fs.appendFile(file,JSON.stringify({event:'step',action:'drained',time:last,args:{pending:0},ref:'source meter, browser response observers and adapter body tasks drained'})+'\n');
  },async close(){if(!closed){await log.close();closed=true}}};
}
async function verifyShadow(dir){await verifyWithGo(dir,'TestRecordedJourney',{GOO_GOTOGETHER_REQUESTS:path.join(dir,'gotogether-requests.jsonl'),GOTOGETHER_STRICT:'1'},{timeoutSeconds:180,moduleDir:'verification/gotogether'});await verifyWithGo(dir,'TestBombadilRecorded',{GOO_GOTOGETHER_BOMBADIL:path.join(dir,'bombadil/trace.jsonl'),GOTOGETHER_STRICT:'1'},{timeoutSeconds:180,moduleDir:'verification/gotogether'});}
module.exports={startShadow,verifyShadow};
