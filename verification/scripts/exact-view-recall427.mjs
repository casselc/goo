import fs from 'node:fs';
import {Journal,normalizeIntent,identity} from '../../ui/scope-history.mjs';
const operations=JSON.parse(fs.readFileSync(0,'utf8'));
let clock=1000000;const values=new Map(),storage={getItem:key=>values.get(key)??null,setItem:(key,value)=>values.set(key,value),removeItem:key=>values.delete(key)};
const create=()=>new Journal({storage:()=>storage,path:'/state427',now:()=>clock});
let journal=create();journal.persistence(true);
const snapshots=[];
for(const op of operations){
 clock+=op.action==='expire'?31*86400000:1;
 const intent={kind:'saved-view',id:op.id,title:op.title};
 switch(op.action){
 case 'remember':journal.remember(intent);break;
 case 'pin':journal.pin(intent);break;
 case 'unpin':journal.unpin(intent);break;
 case 'forget':journal.forget(intent);break;
 case 'clear':journal.clearRecent();break;
 case 'reload':journal=create();break;
 case 'expire':break;
 default:throw Error('Unknown generated operation');
 }
 snapshots.push({items:journal.list(),stored:[...values.values()].map(value=>JSON.parse(value))});
}
const candidate={kind:'saved-view',id:'a'.repeat(64),title:'A'};
const unchanged=identity(candidate)===identity({...candidate,title:'Different display title'});
const forbidden=['question','queries','rows','source','authoring','receipt','binding'].map(key=>{try{normalizeIntent({...candidate,[key]:{private:'not a reference'}});return false}catch{return true}});
process.stdout.write(JSON.stringify({snapshots,unchanged,forbidden}));
