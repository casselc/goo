import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mount,ready,labelled,rows,search,text,key,union} from './support/facet-pagination-369.mjs';
const item=(value,count='1')=>({value,count});
const batch=(prefix,n=1000)=>Array.from({length:n},(_,i)=>item(prefix+String(i).padStart(4,'0'),String(i%19+1)));
const state=element=>({keys:rows(element).map(row=>row.dataset.facetValueKey),matches:labelled(element,'Loaded facet matches').textContent,range:labelled(element,'Shown facet values').textContent});
function expectRows(element,expected){assert.deepEqual(rows(element).map(row=>row.dataset.facetValueKey),expected.map(item=>key(item.value)));assert(rows(element).length<=100)}
test('bounded local pages search the entire loaded partition including missing, empty and full Unicode text',()=>{
 const supplied=[item(null),item(''),...batch('ordinary-',997),item('🙂'.repeat(170)+' terminal-雪-0999')],element=mount({values:ready(supplied)}),original=structuredClone(element.data),events=[];
 element.addEventListener('goo:facet-stage',event=>events.push(event.detail));expectRows(element,supplied.slice(0,100));assert.equal(labelled(element,'Loaded facet matches').textContent,'1000 of 1000 loaded values');
 const input=search(element,'terminal-雪-0999');expectRows(element,supplied.slice(-1));assert.equal(input.value,'terminal-雪-0999');assert.equal(document.activeElement,input);assert.equal(labelled(element,'Loaded facet matches').textContent,'1 of 1000 loaded values');assert.equal(rows(element)[0].querySelector('.categorical-facet-value').title,text(supplied.at(-1).value));
 for(const [find,value]of [['missing key',null],['present empty string','']]){search(element,find);expectRows(element,[item(value)]);const action=labelled(element,'Include attribute:cohort '+text(value));action.onclick();assert.deepEqual(events.at(-1),{value,mode:'include'})}
 search(element,'not supplied');expectRows(element,[]);assert.equal(labelled(element,'Shown facet values').textContent,'Showing 0 of 0 matches');assert.equal(labelled(element,'Next facet page').hidden,true);search(element,'');expectRows(element,supplied.slice(0,100));assert.deepEqual(element.data,original);
});
test('all2000 paired values preserve full-union scale and exact side membership on every page',()=>{
 const selected=batch('selected-'),baseline=batch('baseline-');baseline[999].count='9007199254740993';const expected=union(selected,baseline),element=mount({values:ready(selected),baseline:ready(baseline)}),seen=[];
 for(let page=0;page<20;page++){
  const slice=expected.items.slice(page*100,page*100+100);expectRows(element,slice);assert.equal(labelled(element,'Loaded facet matches').textContent,'2000 of 2000 loaded values');
  for(const [index,row]of rows(element).entries()){seen.push(row.dataset.facetValueKey);const cells=row.querySelectorAll('.comparison-bar-cell');assert.deepEqual(cells.map(cell=>cell.dataset.count),[slice[index].selected,slice[index].baseline]);assert(cells.every(cell=>cell.dataset.scaleMax==='9007199254740993'))}
  const next=labelled(element,'Next facet page');assert.equal(next.disabled,page===19);if(!next.disabled)next.onclick();
 }
 assert.equal(new Set(seen).size,2000);assert.deepEqual(seen,expected.items.map(item=>key(item.value)));assert.equal(document.activeElement,labelled(element,'Previous facet page'));
 search(element,'selected-0999');expectRows(element,expected.items.filter(item=>item.value==='selected-0999'));assert.equal(labelled(element,'Loaded facet matches').textContent,'1 of 2000 loaded values');assert.equal(rows(element)[0].querySelectorAll('.comparison-bar-cell')[1].dataset.presence,'absent');
});
test('page and Find retire row actions even when old nodes are maliciously reattached',()=>{
 const element=mount({values:ready(batch('value-',250))}),events=[];element.addEventListener('goo:facet-stage',event=>events.push(event.detail));const old=rows(element)[0].querySelector('button'),saved=old.onclick;labelled(element,'Next facet page').onclick();assert.equal(old.isConnected,false);element.append(old);assert.equal(old.isConnected,true);saved();assert.equal(events.length,0);
 const second=rows(element)[0].querySelector('button'),secondFn=second.onclick;search(element,'value-0249');element.append(second);secondFn();assert.equal(events.length,0);const current=rows(element)[0].querySelector('button');current.onclick();assert.deepEqual(events,[{value:'value-0249',mode:'include'}]);element.interaction={current:()=>false};current.onclick();assert.equal(events.length,1);
});
test('local search keeps input/caret and same-read republishing restores focus/page without rereading',()=>{
 const element=mount({values:ready(batch('needle-',350))}),input=search(element,'needle-');input.setSelectionRange(2,5,'backward');const sameInput=input;labelled(element,'Next facet page').onclick();assert.equal(document.activeElement,labelled(element,'Next facet page'));assert.equal(labelled(element,'Search loaded facet values'),sameInput);assert.deepEqual([input.selectionStart,input.selectionEnd,input.selectionDirection],[2,5,'backward']);input.focus();element.data=element.data;const restored=labelled(element,'Search loaded facet values');assert.equal(restored.value,'needle-');assert.equal(document.activeElement,restored);assert.deepEqual([restored.selectionStart,restored.selectionEnd,restored.selectionDirection],[2,5,'backward']);assert.equal(labelled(element,'Shown facet values').textContent,'Showing 101–200 of 350 matches');
 const next=element.data;next.values.readAt='2026-11-07T00:00:02Z';element.data=next;assert.equal(labelled(element,'Search loaded facet values').value,'');assert.equal(labelled(element,'Shown facet values').textContent,'Showing 1–100 of 350 matches');
});
test('deterministic mixed page/Find schedules match a fullmodel and produce no source reads',()=>{
 const previous=globalThis.fetch;globalThis.fetch=()=>{throw Error('Passive local paging attempted a source read')};
 try{for(let seed=1;seed<=12;seed++){
  const selected=batch('left-'+seed+'-',333),baseline=batch('right-'+seed+'-',391),oracle=union(selected,baseline),element=mount({values:ready(selected),baseline:ready(baseline)});let query='',page=0,n=seed;
  for(let step=0;step<30;step++){n=(n*1664525+1013904223)>>>0;const action=n%4;
   if(action<2){query=['','left','right','-00','absent'][n%5];search(element,query);page=0}else{const matching=oracle.items.filter(item=>text(item.value).toLocaleLowerCase().includes(query.toLocaleLowerCase())),direction=action===2?1:-1,button=labelled(element,direction>0?'Next facet page':'Previous facet page');if(!button.disabled){button.onclick();page+=direction}assert(page>=0&&page<=Math.max(0,Math.ceil(matching.length/100)-1))}
   const matched=oracle.items.filter(item=>text(item.value).toLocaleLowerCase().includes(query.toLocaleLowerCase()));expectRows(element,matched.slice(page*100,page*100+100));assert.equal(state(element).matches,matched.length+' of '+oracle.items.length+' loaded values');
  }
 }}finally{globalThis.fetch=previous}
});

test('inactive owners, detached and hidden pager callbacks cannot change the held page',()=>{
 const element=mount({values:ready(batch('value-',250))}),button=labelled(element,'Next facet page'),saved=button.onclick,before=state(element);button.isConnected=false;saved();assert.deepEqual(state(element),before);button.isConnected=true;button.hidden=true;saved();assert.deepEqual(state(element),before);button.hidden=false;element.inert=true;saved();assert.deepEqual(state(element),before);element.inert=false;
 let active=false;element.interaction={current:()=>active};const next=labelled(element,'Next facet page'),input=labelled(element,'Search loaded facet values'),held=state(element);next.onclick();input.value='value-0249';input.oninput();assert.deepEqual(state(element),held);active=true;input.oninput();assert.equal(rows(element).length,1);assert.equal(rows(element)[0].dataset.facetValueKey,'1value-0249');
 const oldNext=next.onclick;search(element,'');oldNext();assert.equal(labelled(element,'Shown facet values').textContent,'Showing 1–100 of 250 matches');
});
