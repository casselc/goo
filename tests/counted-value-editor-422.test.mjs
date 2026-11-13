import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Editor from '../ui/filter-editor.mjs';
import {reason as comparisonReason} from '../ui/counted-selection-comparison.mjs';

class Element{
 constructor(tag){this.localName=tag;this.children=[];this.attrs={};this.dataset={};this.style={};this.value='';this.textContent='';this.open=false;this.hidden=false;this.disabled=false;this.inert=false;this.root=false;}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 append(...nodes){for(const node of nodes){if(typeof node==='object'){node.remove();node.parentElement=this;}this.children.push(node)}if(this.localName==='select'&&!this.value)this.value=this.children.find(c=>c.value)?.value||''}
 prepend(...nodes){for(const node of nodes){node.remove();node.parentElement=this;}this.children.unshift(...nodes)}
 before(node){const p=this.parentElement;node.remove();node.parentElement=p;p.children.splice(p.children.indexOf(this),0,node)}
 remove(){if(this.parentElement){const p=this.parentElement;this.parentElement=null;p.children=p.children.filter(c=>c!==this)}}
 replaceChildren(...nodes){for(const child of [...this.children])if(typeof child==='object')child.remove();this.children=[];this.append(...nodes)}
 setAttribute(key,value){this.attrs[key]=String(value)}getAttribute(key){return this.attrs[key]??null}removeAttribute(key){delete this.attrs[key]}
 addEventListener(){}
 contains(node){return node===this||this.children.some(c=>typeof c==='object'&&c.contains(node))}
 matches(selector){if(selector==='details:not([open])')return this.localName==='details'&&!this.open;if(selector===':disabled'){if(this.disabled)return true;for(let p=this.parentElement;p;p=p.parentElement)if(p.localName==='fieldset'&&p.disabled&&!p.querySelector(':scope > legend')?.contains(this))return true;return false}if(selector==='details[open]')return this.localName==='details'&&this.open;return this.localName===selector}
 closest(selector){for(let node=this;node;node=node.parentElement){if(selector==='details:not([open])'&&node.matches(selector)||selector==='fieldset:disabled'&&node.localName==='fieldset'&&node.disabled||selector==='.ui-action-menu'&&node.className?.split(' ').includes('ui-action-menu')||selector.includes('[hidden]')&&node.hidden||selector.includes('[aria-hidden')&&node.getAttribute('aria-hidden')==='true'||selector.includes('[inert]')&&node.inert||selector==='goo-grouping-control'&&node.localName===selector)return node}return null}
 querySelectorAll(selector){if(selector===':scope > summary'||selector===':scope > legend')return this.children.filter(c=>c.localName===selector.slice(9));if(selector==='details[open]>summary')return walk(this).filter(c=>c.localName==='summary'&&c.parentElement?.localName==='details'&&c.parentElement.open);return walk(this).slice(1).filter(c=>selector.split(',').some(s=>c.matches(s.trim())))}
 querySelector(selector){return this.querySelectorAll(selector)[0]||null}
 getClientRects(){return this.isConnected&&!this.closest('[hidden], [aria-hidden="true"]')?[{}]:[]}
 focus(){document.activeElement=this}scrollIntoView(){}
}
const walk=node=>[node,...node.children.filter(c=>typeof c==='object').flatMap(walk)];
const label=(host,text)=>walk(host).find(node=>node.getAttribute('aria-label')===text);
function dom(fn){const before=globalThis.document;globalThis.document={hidden:false,createElement:tag=>new Element(tag)};try{return fn()}finally{globalThis.document=before}}
const field={kind:'attribute',key:'region'},base={field:{kind:'column',key:'service'},op:'eq',value:'api'};
const supplied=()=>({state:'ready',field,question:{signal:'logs',environment:'prod',fromMs:0,toMs:60000,limit:1,offset:0,groupBy:[field],where:base},label:'Original applied UTC window',total:'4',readAt:'2026-11-13T00:01:00Z',basis:'a'.repeat(64),items:[{value:'US',count:'1'},{value:'DE',count:'1'},{value:null,count:'1'},{value:'',count:'1'}]});
function setup({data=supplied(),current=()=>true,where=data.question.where,onCompareCounted,compareCountedReason,canMutate}={}){const host=new Element('section');host.root=true;const changes=[],editor=Editor.mount(host,{signal:'logs',where,onCompareCounted,compareCountedReason,canMutate,onChange:(where,error)=>changes.push({where,error})});editor.setCountedValues(data,{current});const byText=text=>walk(host).find(e=>e.textContent===text),select=value=>label(host,'Select counted attribute:region '+(value===null?'Missing key':value===''?'Present empty string':JSON.stringify(value))),toggle=(value,checked=true)=>{const input=select(value);input.checked=checked;input.onchange();return input};return {host,editor,changes,byText,select,toggle,stage:byText('Stage selected values'),clear:label(host,'Clear selected values'),search:label(host,'Search counted values'),compare:byText('Compare with rest')}};
test('422 selection/search remain local and Stage atomically ORs values within retained scope',()=>dom(()=>{
 const original=supplied(),f=setup({data:original});assert.equal(f.stage.disabled,true);const description=walk(f.host).find(e=>e.id===f.select('US').getAttribute('aria-describedby'));assert(description);assert(description.children.some(e=>e.textContent===' records in original scope'));f.toggle('US');f.search.value='DE';f.search.oninput();assert(walk(f.host).some(e=>e.textContent==='1 selected · 1 hidden by search'));f.toggle('DE');assert.equal(f.changes.length,0);assert.deepEqual(f.editor.snapshot(),base);f.stage.onclick();assert.equal(f.changes.length,1);assert.equal(f.changes[0].error,null);assert.deepEqual(f.editor.snapshot(),{all:[base,{any:[{field,op:'eq',value:'US'},{field,op:'eq',value:'DE'}]}]});f.stage.onclick();assert.equal(f.changes.length,1);assert.equal(f.select('US').disabled,true);assert.deepEqual(original,supplied());
}));
test('422 missing, present-empty and literal values stay distinct; clear is local',()=>dom(()=>{
 const f=setup();f.toggle('US');f.clear.onclick();assert.equal(f.select('US').checked,false);assert.equal(f.changes.length,0);f.toggle('');f.toggle(null);assert.equal(f.changes.length,0);f.stage.onclick();assert.deepEqual(f.editor.snapshot(),{all:[base,{any:[{field,op:'missing'},{field,op:'eq',value:''}]}]});assert.equal(f.changes.length,1);
}));
test('422 stale/native-withdrawn controls refuse and restore authoritative checkbox state',()=>dom(()=>{
 let current=true;const f=setup({current:()=>current}),us=f.select('US');for(const mode of ['hidden','inert','disabled']){us[mode]=true;us.checked=true;us.onchange();assert.equal(us.checked,false);us[mode]=false;}f.toggle('US');current=false;us.checked=false;us.onchange();assert.equal(us.checked,true);f.stage.onclick();assert.equal(f.changes.length,0);current=true;f.editor.set(base);f.stage.onclick();assert.equal(f.changes.length,0);assert.equal(us.checked,false);assert.equal(us.disabled,true);
 const fresh=setup(),old=fresh.select('US');fresh.editor.setCountedValues(supplied());old.checked=true;old.onchange();assert.equal(old.checked,false);assert.equal(fresh.changes.length,0);
}));
test('422 invalid selections preserve tree and unsupported values remain visible-disabled',()=>dom(()=>{
 const data=supplied();data.items.push({value:'x'.repeat(1025),count:'1'});const f=setup({data});assert.equal(f.select('x'.repeat(1025)).disabled,true);
 const crowded={all:Array.from({length:31},(_,i)=>({field:{kind:'attribute',key:'k'+i},op:'exists'}))};f.editor.set(crowded);data.question.where=crowded;f.editor.setCountedValues(data);const input=label(f.host,'Select counted attribute:region "US"');input.checked=true;input.onchange();assert.equal(input.checked,false);assert.deepEqual(f.editor.snapshot(),crowded);assert.equal(f.changes.length,0);
 const wrong=supplied();wrong.question.groupBy=[{kind:'resource',key:'region'}];const bad=setup({data:wrong});assert.equal(bad.select('US'),undefined);assert(walk(bad.host).some(e=>e.textContent==='Reload counted values for the current scope.'));
}));

test('422 an unfiltered applied scope can select and stage an OR predicate',()=>dom(()=>{
 const data=supplied();delete data.question.where;const f=setup({data});assert.equal(f.select('US').disabled,false);f.toggle('US');f.toggle('DE');assert.equal(f.changes.length,0);assert.equal(f.stage.disabled,false);f.stage.onclick();assert.deepEqual(f.editor.snapshot(),{any:[{field,op:'eq',value:'US'},{field,op:'eq',value:'DE'}]});assert.equal(f.changes.length,1);
}));

test('425 comparison stages one detached exact intent including hidden selections without mutating the draft or local selection',()=>dom(()=>{
 const data=supplied(),before=structuredClone(data),calls=[],f=setup({data,onCompareCounted:(...args)=>calls.push(args),compareCountedReason:comparisonReason});
 assert(f.compare);assert.equal(f.compare.disabled,true);f.compare.onclick();assert.equal(calls.length,0);
 f.toggle(null);f.toggle('');f.search.value='US';f.search.oninput();assert.equal(f.compare.disabled,false);assert(walk(f.host).some(e=>e.textContent==='2 selected · 2 hidden by search'));
 f.compare.onclick();assert.equal(calls.length,1);const [question,qualified,values,opener]=calls[0];assert.deepEqual(question,data.question);assert.deepEqual(qualified,field);assert.deepEqual(values,[null,'']);assert.equal(opener,f.compare);assert.equal(f.changes.length,0);assert.deepEqual(f.editor.snapshot(),base);assert.equal(f.select(null).checked,true);assert.equal(f.select('').checked,true);assert.equal(f.stage.disabled,false);
 question.where.value='foreign';qualified.key='foreign';values.push('US');assert.deepEqual(data,before);f.compare.onclick();assert.equal(calls.length,2);assert.deepEqual(calls[1].slice(0,3),[before.question,field,[null,'']]);assert.deepEqual(f.editor.snapshot(),base);assert.equal(f.changes.length,0);
}));

test('425 contextual comparison refuses stale, suspended, withdrawn and retired callbacks',()=>dom(()=>{
 let current=true,mutable=true;const calls=[],f=setup({current:()=>current,canMutate:()=>mutable,onCompareCounted:(...args)=>calls.push(args),compareCountedReason:comparisonReason});f.toggle('US');assert.equal(f.compare.disabled,false);
 const compare=f.compare,handler=compare.onclick,focus={};document.activeElement=focus;
 const denied=()=>{handler.call(compare);assert.equal(calls.length,0);assert.equal(f.changes.length,0);assert.deepEqual(f.editor.snapshot(),base);assert.equal(document.activeElement,focus);assert.equal(f.select('US').checked,true)};
 for(const mode of ['hidden','inert','disabled']){compare[mode]=true;denied();compare[mode]=false;}
 compare.setAttribute('aria-disabled','true');denied();compare.removeAttribute('aria-disabled');
 for(const mode of ['hidden','inert']){f.host[mode]=true;denied();f.host[mode]=false;}
 const disclosure=new Element('details');disclosure.root=true;f.host.root=false;disclosure.append(f.host);disclosure.open=false;denied();disclosure.open=true;
 current=false;denied();current=true;mutable=false;denied();mutable=true;document.hidden=true;denied();document.hidden=false;
 f.editor.setCountedValues(supplied(),{current:()=>current});handler.call(compare);assert.equal(calls.length,0);assert.equal(compare.disabled,true);assert.equal(f.changes.length,0);
 const newer=walk(f.host).find(e=>e.textContent==='Compare with rest');label(f.host,'Select counted attribute:region "US"').checked=true;label(f.host,'Select counted attribute:region "US"').onchange();assert.equal(newer.disabled,false);f.editor.set(base);newer.onclick();assert.equal(calls.length,0);assert.equal(f.changes.length,0);assert.equal(newer.disabled,true);
}));

test('425 comparison-only refusal preserves valid Stage and current reason is checked again at invocation',()=>dom(()=>{
 const data=supplied();data.question.where={all:Array.from({length:28},(_,i)=>({field:{kind:'attribute',key:'retained'+i},op:'exists'}))};const calls=[],f=setup({data,onCompareCounted:(...args)=>calls.push(args),compareCountedReason:comparisonReason});f.toggle('US');f.toggle('DE');assert.equal(f.stage.disabled,false);assert.equal(f.compare.disabled,true);assert.match(f.compare.title,/32|nested/);f.compare.onclick();assert.equal(calls.length,0);assert.equal(f.changes.length,0);assert.deepEqual(f.editor.snapshot(),data.question.where);
 let reason='';const fresh=setup({onCompareCounted:(...args)=>calls.push(args),compareCountedReason:()=>reason});fresh.toggle('US');assert.equal(fresh.compare.disabled,false);reason='Comparison owner changed.';fresh.compare.onclick();assert.equal(calls.length,0);assert.equal(fresh.compare.disabled,true);assert.equal(fresh.stage.disabled,false);assert(walk(fresh.host).some(e=>e.textContent.includes('Comparison owner changed.')));assert.equal(fresh.changes.length,0);
}));
