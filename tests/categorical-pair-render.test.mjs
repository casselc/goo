import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag='section'){this.tagName=tag;this.children=[];this.attributes={};this.dataset={};this.isConnected=false;this.style={setProperty(key,value){this[key]=value}};this.listeners=new Map();this.className='';this.classList={add:(name)=>{if(!this.className.split(' ').includes(name))this.className+=' '+name},toggle:(name,on)=>{this.className=this.className.split(' ').filter(value=>value&&value!==name).join(' ');if(on)this.classList.add(name)}};}
 set textContent(value){this._text=value??''}get textContent(){return this._text+this.children.map(child=>child.textContent).join('')}
 append(...children){for(const child of children){child.parentElement=this;child.isConnected=this.isConnected;this.children.push(child)}}
 prepend(...children){for(const child of children.reverse()){child.parentElement=this;child.isConnected=this.isConnected;this.children.unshift(child)}}
 replaceChildren(...children){for(const child of this.children)child.parentElement=null;this.children=[];this._text='';this.append(...children)}
 contains(element){return element===this||this.children.some(child=>child.contains(element))}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}
 matches(selector){if(selector.startsWith('.'))return this.className.split(' ').includes(selector.slice(1));if(selector==='[aria-label]')return this.getAttribute('aria-label')!==null;return this.tagName===selector}
 querySelectorAll(selector){return this.children.flatMap(child=>[...(child.matches(selector)?[child]:[]),...child.querySelectorAll(selector)])}querySelector(selector){return this.querySelectorAll(selector)[0]||null}
 addEventListener(type,handler){this.listeners.set(type,handler)}dispatchEvent(event){this.listeners.get(event.type)?.(event);return true}focus(){globalThis.document.activeElement=this}
}
const node=(tag,text,className)=>{const element=new Element(tag);element.textContent=text;if(className)element.className=className;return element};
globalThis.HTMLElement=Element;const body=node('body');body.isConnected=true;globalThis.document={body,activeElement:body};
mock.module('../ui/components.mjs',{exports:{node}});
const {CategoricalFacet}=await import('../ui/categorical-facet.mjs');
const {create}=await import('../ui/facet-declaration.mjs');
const field={kind:'attribute',key:'http.request.method'},declaration=create({signal:'logs',environment:'prod',fromMs:0,toMs:60000,offset:0,limit:100,groupBy:[]},field),receipt={readAt:'2026-11-05T00:00:01Z',basis:'a'.repeat(64)};
const ready=items=>({state:'ready',field,items,total:String(items.reduce((sum,item)=>sum+BigInt(item.count),0n)),readAt:receipt.readAt,basis:'grouped',reads:{count:receipt,groups:receipt},label:'logs · prod · [2026-11-05T00:00:00Z, 2026-11-05T00:01:00Z)'});
const mount=data=>{const element=new CategoricalFacet();element.data={declaration,...data};element.isConnected=true;element.render();return element};
const button=(element,label)=>element.querySelectorAll('[aria-label]').find(child=>child.getAttribute('aria-label')===label);
test('a successful zero grouping remains a partial witness with its scope and receipts',()=>{
 for(const state of ['error','loading','unloaded']){const element=mount({values:ready([]),baseline:{state,error:'bounded failure'},baselineLabel:'Previous window'});assert.equal(element.dataset.state,'partial');assert.match(element.textContent,/Selected 0 · Previous window unavailable/);assert.match(element.querySelector('.categorical-facet-receipts').textContent,/2026-11-05T00:00:01Z/);assert.match(element.querySelector('.categorical-facet-receipts').textContent,/logs · prod/);assert.doesNotMatch(element.textContent,/No values observed in either window/);}
});
test('partial rows retain exact known counts without absent-zero or comparison bars',()=>{
 const element=mount({values:ready([{value:'POST',count:'27'}]),baseline:{state:'error',error:'Read refused.'},baselineLabel:'Previous window'}),cells=element.querySelectorAll('.comparison-bar-cell');assert.equal(element.dataset.state,'partial');assert.equal(cells[0].dataset.count,'27');assert.equal(cells[1].dataset.count,'unknown');assert.equal(cells[1].dataset.presence,'unknown');assert.equal(cells[1].querySelector('.comparison-value').textContent,'—');assert.equal(element.querySelectorAll('.comparison-bar-fill').length,0);assert(button(element,'Retry baseline values'));
});
test('complete rows share a full-union scale and make baseline-only membership actionable',()=>{
 const element=mount({values:ready([{value:'GET',count:'75'}]),baseline:ready([{value:'DELETE',count:'2'}]),baselineLabel:'Previous window'}),cells=element.querySelectorAll('.comparison-bar-cell');assert.equal(element.dataset.state,'ready');assert(cells.every(cell=>cell.dataset.scaleMax==='75'));assert(cells.every(cell=>cell.getAttribute('role')==='group'));assert(cells.every(cell=>cell.getAttribute('aria-label')?.includes(' count for ')));assert.equal(cells.find(cell=>cell.getAttribute('aria-label')==='Selected count for "DELETE"').dataset.count,'0');assert.equal(cells.find(cell=>cell.getAttribute('aria-label')==='Previous window count for "DELETE"').dataset.count,'2');assert(button(element,'Include attribute:http.request.method "DELETE"'));assert.equal(element.querySelectorAll('.comparison-bar-fill').length,4);
});
test('retry and stage intents obey passive owner leases and retired render epochs',()=>{
 const element=mount({values:ready([{value:'POST',count:'27'}]),baseline:{state:'error'},baselineLabel:'Previous window'}),events=[];for(const type of ['goo:facet-stage','goo:facet-retry'])element.addEventListener(type,event=>events.push({type,detail:event.detail}));const retry=button(element,'Retry baseline values'),stage=button(element,'Include attribute:http.request.method "POST"');retry.onclick();stage.onclick();assert.deepEqual(events,[{type:'goo:facet-retry',detail:{side:'baseline'}},{type:'goo:facet-stage',detail:{value:'POST',mode:'include'}}]);element.data={...element.data,baseline:ready([])};retry.onclick();stage.onclick();assert.equal(events.length,2);element.interaction={current:()=>false};button(element,'Include attribute:http.request.method "POST"').onclick();assert.equal(events.length,2);
});
