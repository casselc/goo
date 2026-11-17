import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {Element} from './support/raw-observation-dom.mjs';
import {mount} from '../ui/workspace.mjs';
function fixture(hash){
 const ids=new Map(),element=(tag,id)=>{const value=new Element(tag);if(id){value.setAttribute('id',id);ids.set(id,value)}return value};
 const body=element('body');body.root=true;
 const skip=element('a'),main=element('main','mainContent'),title=element('h1','workspaceTitle'),description=element('p','workspaceDescription'),sidebar=element('aside'),menu=element('button'),topbar=element('div'),scope=element('div'),label=element('label'),minutes=element('input','minutes');
 skip.className='skip-link';skip.setAttribute('href','#mainContent');main.tabIndex=-1;sidebar.className='app-sidebar';menu.className='entry-menu';topbar.className='app-topbar';scope.className='scope-control';label.setAttribute('for','minutes');const scopeQuery=scope.querySelector.bind(scope);scope.querySelector=selector=>selector==='label[for="minutes"]'?label:scopeQuery(selector);scope.append(label,minutes);topbar.append(scope);sidebar.append(menu);main.append(topbar,title,description);body.append(skip,sidebar,main);
 const buttons=['explore','services','investigations'].map(route=>{const button=element('button');button.dataset.workspace=route;sidebar.append(button);return button}),panels=buttons.map(button=>{const panel=element('section');panel.dataset.workspacePanel=button.dataset.workspace;main.append(panel);return panel});
 const events=[],history=[];globalThis.document={activeElement:null,title:'',getElementById:id=>ids.get(id),querySelector:selector=>body.querySelector(selector),querySelectorAll:selector=>selector==='[data-workspace]'?buttons:selector==='[data-workspace-panel]'?panels:body.querySelectorAll(selector)};globalThis.location={hash};globalThis.history={pushState:(...args)=>history.push(args)};globalThis.addEventListener=()=>{};globalThis.dispatchEvent=event=>events.push(event);globalThis.CustomEvent=class{constructor(type,init){this.type=type;Object.assign(this,init)}};
 let scroll;main.scrollIntoView=value=>scroll=value;const owner=mount();return {owner,skip,main,menu,panels,events,history,scroll:()=>scroll};
}
test('skip bypass preserves Services and Saved routes, source state and navigation history',()=>{
 for(const hash of ['#services','#investigation='+'a'.repeat(64)]){
  const f=fixture(hash),before=f.owner.current(),active=f.panels.find(panel=>!panel.hidden),events=f.events.length;let prevented=false;f.skip.focus();f.menu.setAttribute('aria-expanded','true');f.skip.onclick({preventDefault:()=>prevented=true});
  assert.equal(prevented,true);assert.equal(document.activeElement,f.main);assert.deepEqual(f.scroll(),{block:'start',behavior:'instant'});assert.equal(f.menu.getAttribute('aria-expanded'),'false');assert.equal(f.owner.current(),before);assert.equal(f.panels.find(panel=>!panel.hidden),active);assert.equal(location.hash,hash);assert.deepEqual(f.history,[]);assert.equal(f.events.length,events,'bypass does not dispatch a route or data action');
 }
});
test('native bypass has an early anchor and focusable main target without requiring JavaScript',()=>{
 const html=readFileSync(new URL('../ui/index.html',import.meta.url),'utf8');assert.match(html,/<body>\s*<a class="skip-link" href="#mainContent">Skip to main content<\/a>\s*<aside/);assert.match(html,/<main id="mainContent" class="app-main" tabindex="-1">/);
 const f=fixture('#services');const original=document.getElementById;document.getElementById=id=>id==='mainContent'?null:original(id);let prevented=false;f.skip.onclick({preventDefault:()=>prevented=true});assert.equal(prevented,false,'a missing enhancement target retains native anchor behavior');assert.equal(f.owner.current(),'services');
});

test('Services operations lookback is a local tab context without another workspace route',()=>{const f=fixture('#services'),minutes=document.getElementById('minutes'),events=f.events.length;f.owner.setServiceFamily('operations');assert.equal(minutes.getAttribute('aria-label'),'New operation read lookback minutes');assert.equal(f.owner.current(),'services');assert.equal(f.events.length,events);f.owner.setServiceFamily('services');assert.equal(minutes.getAttribute('aria-label'),'New discovery lookback minutes')});
