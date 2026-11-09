import {test} from 'node:test';
import assert from 'node:assert/strict';
import {diagnosis} from '../ui/service-dashboard.mjs';
import {readQuestionLink} from '../ui/question-protocol.mjs';
class Element{
 constructor(localName){this.localName=localName;this.children=[];this.parentElement=null;this.dataset={};this.style={setProperty(){}};this.attributes={};this.hidden=false;this.open=false;this.value='';this.textContent='';}
 get isConnected(){return !!this.root||!!this.parentElement?.isConnected}
 append(...nodes){for(const node of nodes){if(node&&typeof node==='object')node.parentElement=this;this.children.push(node)}}
 replaceChildren(...nodes){for(const node of this.children)if(node&&typeof node==='object')node.parentElement=null;this.children=[];this.append(...nodes)}
 setAttribute(key,value){this.attributes[key]=String(value)}
 getAttribute(key){return this.attributes[key]??null}
 getClientRects(){return this.isConnected?[{}]:[]}
 contains(node){return node===this||this.children.some(child=>child&&typeof child==='object'&&child.contains(node))}
 querySelector(selector){return selector===':scope > summary'?this.children.find(child=>child.localName==='summary')||null:null}
 closest(selector){for(let node=this;node;node=node.parentElement){if(selector==='details:not([open])'&&node.localName==='details'&&!node.open)return node;if(selector.includes('hidden')&&(node.hidden||node.attributes['aria-hidden']==='true'))return node;if(selector==='[inert]'&&node.inert)return node}return null}
}
const walk=host=>[host,...host.children.filter(x=>x&&typeof x==='object').flatMap(walk)];
test('qualified context links refuse hidden and retired callbacks, while an explicit visible drill stages literal missing/empty values',t=>{
 const previous={document:globalThis.document,Option:globalThis.Option};globalThis.document={hidden:false,createElement:tag=>new Element(tag)};globalThis.Option=class extends Element{constructor(text,value){super('option');this.textContent=text;this.value=value}};t.after(()=>Object.assign(globalThis,previous));
 for(const value of [null,'','Missing key','503']){const root=new Element('section');root.root=true;const disclosure=new Element('details');disclosure.append(new Element('summary'));root.append(disclosure);let alive=true;const q={signal:'spans',environment:'prod',fromMs:0,toMs:1000,limit:100,offset:0,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'column',key:'kind'},op:'eq',value:'Client'},{field:{kind:'attribute',key:'http.request.method'},op:'ne',value:''}]},groupBy:[{kind:'attribute',key:'http.response.status_code'}]},data={question:q,state:'observed',duration:{state:'observed',count:'1',sumNs:'2'},groups:[{g0:value,count:'1',duration:{state:'observed',count:'1',sumNs:'2',p95Ns:'2'}}]},before=structuredClone(data);
  diagnosis(disclosure,data,{family:'http-client',current:()=>alive});const link=walk(disclosure).find(e=>e.localName==='a'&&e.textContent==='Open response-status population');assert(link);const denied={prevented:false,stopped:false,preventDefault(){this.prevented=true},stopImmediatePropagation(){this.stopped=true}};assert.equal(link.onclick(denied),false);assert(denied.prevented&&denied.stopped);disclosure.open=true;assert.equal(link.onclick({}),true);const staged=readQuestionLink(link.href).question;assert.deepEqual(staged.where.all[0],q.where);assert.deepEqual(staged.where.all[1],{field:q.groupBy[0],op:value===null?'missing':'eq',...(value===null?{}:{value})});assert.deepEqual(data,before);alive=false;assert.equal(link.onclick({}),false);
 }
});
