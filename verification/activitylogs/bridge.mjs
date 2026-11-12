import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Logs from '../../ui/activity-log-model.mjs';
import {ReadGate} from '../../ui/aggregate-read-gate.mjs';
const inputs=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Pure log projection acquired a source')};
const copy=x=>structuredClone(x);
const output=inputs.map(c=>{
 const before=copy(c),q=c.request||Logs.query(c.context),snapshot=Logs.admit(q,c.dto);
 const selector={signal:q.signal,...(q.environment!==undefined?{environment:q.environment}:{}),...(q.environmentSelector?{environmentSelector:q.environmentSelector}:{}),...(q.where?{where:q.where}:{})};
 const context={selector,readWindow:{fromMs:q.fromMs,toMs:q.toMs}},later={selector,readWindow:{fromMs:q.fromMs+1000,toMs:q.toMs+1000}};
 assert.equal(Logs.scopeKey(context),Logs.scopeKey(later),'scope identity must not follow rolling time');
 const a=c.a??1,b=c.b??3,selection=Logs.select(snapshot,String(a),String(b));
 const gate=new ReadGate(),states=(c.commands||[]).map(command=>{let result;if(command.name==='bind')result=gate.bind(command.scope);else if(command.name==='start')result=gate.start();else if(command.name==='accept')result=gate.accept(command.token,command.scope);else if(command.name==='pause')gate.pause();else if(command.name==='resume')gate.resume();else throw Error('Unknown ownership command');return{scope:gate.scope,generation:gate.generation,loading:gate.loading,frozen:gate.frozen,result:result??null}});
 const out={query:q,timeline:Logs.timeline(snapshot),severity:Logs.severity(snapshot),rate:Logs.rate(snapshot),selection,gate:states};
 assert.deepEqual(Logs.select(snapshot,String(b),String(a)),selection,'reverse brush changed membership');
 for(const mutate of [d=>d.question.toMs++,d=>d.state='changed_between_reads',d=>d.count=String(BigInt(d.count)+1n),d=>d.timeSeries[0].toMs++,d=>d.timeSeries[0].count='01',...(c.dto.groups.length?[d=>d.groups[0].g0=null,d=>d.groups.push(copy(d.groups[0]))]:[])]){
  const bad=copy(c.dto);mutate(bad);assert.throws(()=>Logs.admit(q,bad));
 }
 assert.throws(()=>Logs.select(snapshot,'-1','5'));assert.throws(()=>Logs.select(snapshot,'0','6'));
 assert.deepEqual(c,before,'pure projection rewrote its evidence');
 return out;
});
process.stdout.write(JSON.stringify(output));
