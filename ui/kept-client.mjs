/* Local keep transport and bounded browser retry/discovery journal. No source reads. */
import {Journal} from './pending-saves.mjs';
import * as Protocol from './kept-protocol.mjs';
const pendingKey='goo.pending-keeps.v1',recentKey='goo.recent-keeps.v1';
const key=()=>Array.from(crypto.getRandomValues(new Uint8Array(32)),v=>v.toString(16).padStart(2,'0')).join('');
function createClient({storage=()=>globalThis.localStorage,fetch=(...a)=>globalThis.fetch(...a),notify=()=>globalThis.dispatchEvent(new CustomEvent('goo:keeps-changed'))}={}){
 const inflight=new Map();
 const journal=()=>new Journal(storage(),{key:pendingKey,validate:Protocol.entry});
 function recent(){const raw=storage().getItem(recentKey);if(!raw)return [];if(raw.length>32768)throw Error('Kept history exceeds bound');const items=JSON.parse(raw);if(!Array.isArray(items)||items.length>16||items.some(v=>!v||!/^[a-f0-9]{64}$/.test(v.id)||!Number.isInteger(v.count)||v.count<1||v.count>20||!Number.isFinite(Date.parse(v.keptAt))))throw Error('Invalid kept history');return items}
 async function publish(body){const item=journal().begin(body.version==='goo/keep-records/v1'?'/records':'/tail',body);notify();const response=await fetch('/api/kept-observations'+item.path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(item.body)}),value=await response.json();if(!response.ok){const error=Error(value.error||'Keep acknowledgment unavailable');error.status=response.status;throw error}Protocol.validate(value,item.body);try{storage().setItem(recentKey,JSON.stringify([{id:value.id,keptAt:value.keptAt,count:value.members.length},...recent().filter(v=>v.id!==value.id)].slice(0,16)));journal().complete(value.id)}catch{}notify();return value}
 function send(body){const normalized=Protocol.intent(body),existing=inflight.get(normalized.requestID),signature=JSON.stringify(normalized);if(existing){if(existing.signature!==signature)return Promise.reject(Error('Request ID already has a different pending keep'));return existing.promise}const promise=Promise.resolve().then(()=>publish(normalized)).finally(()=>{inflight.delete(normalized.requestID);notify()});inflight.set(normalized.requestID,{signature,promise});return promise}
 async function open(id){if(!/^[a-f0-9]{64}$/.test(id))throw Error('Invalid kept observation ID');const response=await fetch('/api/kept-observations/'+id),value=await response.json();if(!response.ok)throw Error(value.error||'Kept observations unavailable');Protocol.validate(value);if(value.id!==id)throw Error('Returned kept observation ID differs');return value}
 return {send,open,recent,busy:id=>inflight.has(id),pending:()=>journal().list(),forget:id=>{journal().complete(id);notify()},key};
}
const client=createClient();
export {createClient,key,client};
