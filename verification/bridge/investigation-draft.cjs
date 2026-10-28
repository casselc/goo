// Replay actual production Journal; the oracle stores only accepted caller intent.
const assert = require('node:assert/strict');
const {Journal, storageKey, TTL_MS} = require('../../ui/investigation-draft.mjs');
const readline = require('node:readline').createInterface({input:process.stdin});
const clone = x => JSON.parse(JSON.stringify(x));
const id = n => n.toString(16).padStart(64,'0');
let values, journal, expected, sequence, now, network;
const storage = {getItem:k=>values.get(k)??null,setItem:(k,v)=>values.set(k,v),removeItem:k=>values.delete(k)};
global.fetch = () => {network++; throw Error('Journal attempted network request')};
function newJournal(path='/') {return new Journal({storage,path,now:()=>now,id:()=>id(++sequence)})}
function draft(n) {
 return {parent:{id:id(900),recordedAt:'2026-10-24T00:00:00.000Z'},document:{title:'Draft '+n,question:'why?',finding:'',openQuestions:'next',cards:[{kind:'capture',reference:id(901),caption:'exact receipt'},{kind:'note',role:'finding',text:'applied note'}]},composer:{role:'finding',text:'pending '+n,editing:1,captureID:'',caption:'',destination:'',comparisonID:'',comparisonFinding:''},openedQuestion:''};
}
function reset(){values=new Map();sequence=0;now=1792800000000;network=0;expected=null;journal=newJournal();assert.equal(journal.inspect().state,'empty')}
function check() {
 const actual=journal.inspect();
 if(expected===null)assert.equal(actual.state,'empty');
 else {assert.equal(actual.state,'ready');assert.deepEqual(actual.entry,expected);const returned=actual.entry;returned.draft.document.title='caller mutation';assert.deepEqual(journal.inspect().entry,expected)}
 assert.equal(network,0);
 return {count:expected?1:0};
}
function save(unknown=false) {
 const input=draft(sequence+1), previous=expected;
 const entry=unknown?journal.beforePublish(input):journal.checkpoint(input);
 const publication=unknown||previous?.publication==='unknown'?'unknown':'draft';
 expected={version:1,path:'/',id:id(sequence),savedAt:new Date(now).toISOString(),publication,draft:clone(input)};
 assert.deepEqual(entry,expected);
 input.document.cards[0].reference=id(999);entry.draft.composer.text='mutated returned buffer';
}
reset();
readline.on('line',line=>{try{
 const {action}=JSON.parse(line);
 if(action==='reset')reset();
 if(action==='checkpoint')save();
 if(action==='beforePublish')save(true);
 if(action==='reload'){const raw=values.get(storageKey('/'));journal=newJournal();assert.deepEqual(values.get(storageKey('/')),raw)}
 if(action==='oldAck'){save(true);const old=expected.id;save();const raw=values.get(storageKey('/'));assert.equal(journal.acknowledge(old),false);assert.equal(values.get(storageKey('/')),raw)}
 if(action==='ack'){if(expected){assert.equal(journal.acknowledge(expected.id),true);expected=null}else assert.equal(journal.acknowledge(id(9999)),false)}
 if(action==='discard'){const read=journal.inspect();assert.equal(journal.discard(read.token),true);expected=null}
 if(action==='staleDiscard'){save();const stale=journal.inspect().token;save();assert.equal(journal.discard(stale),false)}
 if(action==='foreignWriter'){save();const observed=journal.inspect().token;const second=newJournal();second.inspect();const fresh=second.checkpoint(draft(99));assert.throws(()=>journal.checkpoint(draft(100),{expected:observed}),/changed/);expected=clone(fresh);journal=newJournal()}
 if(action==='malformed'){const input=draft(1);input.document.cards[0].telemetryRows=[{secret:'payload'}];const raw=values.get(storageKey('/'));assert.throws(()=>journal.checkpoint(input));assert.equal(values.get(storageKey('/')),raw)}
 if(action==='wrongParent'){const input=draft(1);input.parent.id='not-an-immutable-id';assert.throws(()=>journal.checkpoint(input))}
 if(action==='expired'){save();const old=values.get(storageKey('/'));now+=TTL_MS;assert.equal(journal.inspect().state,'invalid');assert.throws(()=>journal.checkpoint(draft(1)));assert.equal(values.get(storageKey('/')),old);assert.equal(journal.discard(old),true);expected=null}
 if(action==='wrongPath'){const other=newJournal('/other');assert.equal(other.inspect().state,'empty');assert.equal(other.acknowledge(expected?.id||id(1)),false)}
 if(action==='deniedStorage'){const denied=new Journal({storage:{getItem(){throw Error('denied')},setItem(){throw Error('denied')},removeItem(){throw Error('denied')}},path:'/'});assert.equal(denied.inspect().state,'unavailable');assert.throws(()=>denied.checkpoint(draft(1)))}
 if(action==='quota'){const quota=new Journal({storage:{...storage,setItem(){throw Error('quota')}},path:'/',now:()=>now,id:()=>id(++sequence)});quota.inspect();const raw=values.get(storageKey('/'));assert.throws(()=>quota.checkpoint(draft(1)),/could not be stored/);assert.equal(values.get(storageKey('/')),raw)}
 console.log(JSON.stringify(check()));
 }catch(e){console.log(JSON.stringify({error:e.stack}))}});
