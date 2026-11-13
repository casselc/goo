// Generated native-tree conformance, not a browser/owner substitute.
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {nativeControlReady,completionControlReady,revealFocus} from '../../ui/dom-visibility.mjs';
globalThis.fetch=()=>{throw Error('Passive readiness or return attempted source I/O')};
class Element {
 constructor(tag,parent){this.tag=tag;this.parentElement=parent;this.children=[];parent?.children.push(this);this.open=true;this.hidden=false;this.inert=false;this.disabled=false;this.aria=false;this.connected=true;this.cssHidden=false;this.focused=0;}
 get isConnected(){return this.connected&&(!this.parentElement||this.parentElement.isConnected)}
 contains(value){for(let n=value;n;n=n.parentElement)if(n===this)return true;return false}
 closest(selector){for(let n=this;n;n=n.parentElement){if(selector==='details:not([open])'&&n.tag==='details'&&!n.open)return n;if(selector==='fieldset:disabled'&&n.tag==='fieldset'&&n.disabled)return n;if(selector.includes('[hidden]')&&(n.hidden||n.ariaHidden))return n;if(selector.includes('[inert]')&&n.inert)return n;}return null}
 querySelector(selector){return this.children.find(e=>e.tag===(selector.includes('legend')?'legend':'summary'))??null}
 getAttribute(key){return key==='aria-disabled'&&this.aria?'true':null}
 matches(selector){if(selector==='details:not([open])')return this.tag==='details'&&!this.open;if(selector!==':disabled')return false;if(this.disabled)return true;for(let n=this.parentElement;n;n=n.parentElement)if(n.tag==='fieldset'&&n.disabled&&!n.querySelector(':scope > legend')?.contains(this))return true;return false}
 getClientRects(){for(let n=this;n;n=n.parentElement){if(n.hidden||n.inert||n.cssHidden||n.ariaHidden)return [];if(n.tag==='details'&&!n.open&&!n.querySelector(':scope > summary')?.contains(this))return []}return this.isConnected?[{}]:[]}
 focus(){this.focused++}
 scrollIntoView(){}
}
globalThis.getComputedStyle=element=>({display:element.cssHidden?'none':'block',visibility:'visible'});
const cases=JSON.parse(readFileSync(0,'utf8'));
for(const [index,{Condition:c,Reveal:reveal,Expected:e}]of cases.entries()){
 globalThis.document={hidden:c.DocumentHidden};
 const root=new Element('section'),fieldset=new Element('fieldset',root),legend=new Element('legend',fieldset),outer=new Element('details',c.Legend?legend:fieldset);new Element('summary',outer);const inner=new Element('details',outer);new Element('summary',inner);const origin=new Element('button',inner),fallback=new Element('button',root);
 origin.connected=c.Connected;origin.disabled=c.Disabled;origin.aria=c.ARIA;fieldset.disabled=c.Fieldset;fieldset.inert=c.Inert;fieldset.hidden=c.Hidden;fieldset.cssHidden=c.CSSHidden;inner.open=!c.Closed;outer.open=!c.OuterClosed;
 const before=[inner.open,outer.open];
 assert.equal(c.Current&&nativeControlReady(origin),e.Invoke,'invoke '+index);
 assert.equal(c.Current&&completionControlReady(origin),e.Complete,'completion '+index);
 assert.deepEqual([inner.open,outer.open],before,'admission must not reveal disclosure '+index);
 // Owner currency gates effects, not the passive focus helper. The actual
 // production owners additionally check their lease before calling it.
 const target=revealFocus(origin,fallback,{disclosures:reveal});
 assert.equal(target===origin?'origin':target===fallback?'fallback':'none',e.Focus,'focus '+index);
 assert.equal(Number(!before[0]&&inner.open)+Number(!before[1]&&outer.open),e.Opened,'revealed ancestors '+index);
 assert.equal(target?.focused??0,e.Focus==='none'?0:1,'one explicit focus '+index);
}
process.stdout.write(JSON.stringify(cases.length));
