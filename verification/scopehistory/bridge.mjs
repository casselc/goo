import fs from 'node:fs';
import assert from 'node:assert/strict';
import {Journal,normalizeIntent} from '../../ui/scope-history.mjs';
globalThis.fetch=()=>{throw Error('History acquired a source')};
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=structuredClone(input);
let clock=input.now;
const values=new Map(),storage={getItem:key=>values.get(key)||null,setItem:(key,value)=>values.set(key,value),removeItem:key=>values.delete(key)};
let journal=new Journal({storage:()=>storage,path:'/oracle',now:()=>clock});
const state=()=>({durable:journal.durable,items:journal.list().map(({intent,usedAt,pinned})=>({key:intent.kind+':'+(intent.question||intent.selector).environment,usedAt,pinned,offset:intent.kind==='question'?intent.question.offset:null}))});
const states=input.commands.map(command=>{
 let rejected=false;
 try{
  if(command.name==='advance')clock+=command.ms;
  else if(command.name==='reload')journal=new Journal({storage:()=>storage,path:'/oracle',now:()=>clock});
  else if(command.name==='persistence')journal.persistence(command.enabled);
  else if(['clearRecent','reset'].includes(command.name))journal[command.name]();
  else journal[command.name](command.intent);
 }catch{rejected=true;}
 return {...state(),rejected};
});
for(const intent of input.poison||[])assert.throws(()=>normalizeIntent(intent),'unsafe history configuration admitted');
assert.deepEqual(input,before,'history rewrote the caller command');
process.stdout.write(JSON.stringify(states));
