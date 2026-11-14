import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';
const source=readFileSync(new URL('../ui/documents.mjs',import.meta.url),'utf8');
const between=(a,b)=>source.slice(source.indexOf(a),source.indexOf(b,source.indexOf(a)));
const owned=between('  function fresh(','  el(\'documentNew\')')+between('  function pin(','  el(\'documentAddComparison\')')+between('  controller.linkCapture=','  const contextOpen=');
const id='a'.repeat(64),card={kind:'capture',reference:id,caption:'Pair'};
function fixture({hidden=false}={}){
 const nodes=new Map(),el=key=>{if(!nodes.has(key))nodes.set(key,{hidden:false,open:false,value:'',replaceChildren(){},scrollIntoView(){env.scrolled++}});return nodes.get(key)};
 const env={controller:{},el,contextToken:3,generation:7,current:{id:'old',title:'Original'},cards:[{kind:'note',text:'Keep'}],dirty:false,busy:false,loading:false,awaitingRecovery:false,pending:()=>false,guardRecovery(){},invalidate(){env.generation++;env.loading=false},changed(){env.generation++;env.dirty=true},render(){env.renders++},renders:0,scrolled:0,focused:0,shown:[],events:[],writes:[],message:'unchanged',status:message=>env.message=message,global:{dispatchEvent:event=>env.events.push(event.type)},history:{replaceState(_a,_b,path){env.path=path}},path:'#explore',CustomEvent,structuredClone,idPattern:/^[a-f0-9]{64}$/,newKey:()=> 'question-key',revealFocus(){env.focused++},recoveredParent:null,recovered:false,remoteDetails:true,meaningHost:{replaceChildren(){}},openedQuestion:'',resetComposer(){},checkpoint(){},host:{closest:()=>({hidden:true})}};
 env.show=doc=>{env.shown.push(doc);env.current=doc;env.cards=doc.cards;env.path='#document='+doc.id};
 env.request=(path,body)=>new Promise(resolve=>env.writes.push({path,body,resolve}));
 if(hidden){env.current=null;el('documentEditor').hidden=true;}
 vm.createContext(env);vm.runInContext(owned,env);return env;
}
test('428 current attachment retains authorized card changes without retired navigation, focus or scrolling',async()=>{
 for(const hidden of [false,true]){const env=fixture({hidden}),prior=env.cards.length;await env.controller.attach('current',card,env.contextToken,{navigate:()=>false,focus:true});assert.equal(env.cards.length,(hidden?0:prior)+1);assert.equal(env.cards.at(-1),card);assert.equal(env.path,'#explore');assert.deepEqual(env.events,[]);assert.equal(env.focused,0);assert.equal(env.scrolled,0);}
});
test('428 owned new-draft attachment makes one atomic navigation decision and focuses its visible destination',async()=>{
 const env=fixture({hidden:true});let decisions=0;await env.controller.attach('current',card,env.contextToken,{navigate:()=>{decisions++;return env.path==='#explore'},focus:true});assert.equal(decisions,1);assert.equal(env.path,'#investigations');assert.deepEqual(env.events,['goo:navigate-investigations','goo:show-document']);assert.equal(env.focused,1);
});
test('428 saved-revision attachment writes explicitly but rechecks navigation after both awaited operations',async()=>{
 for(const departure of ['before-base','during-save','none']){const env=fixture(),prior=env.current,priorCards=env.cards;let current=true;const attached=env.controller.attach('saved',card,env.contextToken,{navigate:()=>current,focus:true});assert.equal(env.writes.length,1);if(departure==='before-base')current=false;env.writes[0].resolve({id:'saved',title:'Parent',cards:[]});await new Promise(setImmediate);assert.equal(env.writes.length,2);assert.equal(env.writes[1].body.document.cards[0],card);if(departure==='during-save')current=false;env.writes[1].resolve({id:'child',cards:[card]});const result=await attached;assert.equal(result.id,'child');assert.equal(env.shown.length,departure==='none'?1:0);assert.equal(env.focused,departure==='none'?1:0);if(departure!=='none'){assert.equal(env.current,prior);assert.equal(env.cards,priorCards);assert.equal(env.path,'#explore');assert.equal(env.message,'unchanged');}}
});
test('428 linked-question evidence keeps explicit mutation but a retired completion cannot navigate',()=>{
 const env=fixture();env.cards=[{kind:'question',key:'q',step:{kind:'record'}}];const expected={token:env.contextToken,generation:env.generation};env.controller.linkCapture('q',id,expected,{navigate:()=>false,focus:true});assert.deepEqual([...env.cards[0].evidence],[id]);assert.deepEqual(env.events,[]);assert.equal(env.path,'#explore');assert.equal(env.focused,0);
});
