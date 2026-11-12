import {test} from 'node:test';
import assert from 'node:assert/strict';
import {nativeControlReady,completionControlReady} from '../ui/dom-visibility.mjs';

class Element {
 constructor(tag,parent){this.tag=tag;this.parentElement=parent;this.children=[];parent?.children.push(this);this.isConnected=true;this.disabled=false;this.attrs={};}
 closest(selector){for(let n=this;n;n=n.parentElement){if(selector==='[hidden], [aria-hidden="true"]'&&(n.hidden||n.attrs['aria-hidden']==='true')||selector==='[inert]'&&n.inert||selector==='fieldset:disabled'&&n.tag==='fieldset'&&n.disabled||selector==='details:not([open])'&&n.tag==='details'&&!n.open)return n}return null}
 querySelector(selector){return this.children.find(n=>n.tag===(selector===':scope > legend'?'legend':'summary'))??null}
 contains(element){for(let n=element;n;n=n.parentElement)if(n===this)return true;return false}
 getClientRects(){return this.rect===false?[]:[{}]}
 getAttribute(name){return this.attrs[name]??null}
 matches(selector){return selector===':disabled'&&(this.disabled||!!this.parentElement?.closest('fieldset:disabled'))}
}
function setup(t){const original=globalThis.document;globalThis.document={hidden:false};t.after(()=>{if(original===undefined)delete globalThis.document;else globalThis.document=original});return new Element('button',new Element('section'))}

test('completion may ignore only the opener’s own pending disabled flag',t=>{
 const opener=setup(t);assert.equal(nativeControlReady(opener),true);opener.disabled=true;
 assert.equal(nativeControlReady(opener),false);assert.equal(completionControlReady(opener),true);
 for(const withdraw of [()=>{opener.attrs['aria-disabled']='true'},()=>{opener.hidden=true},()=>{opener.inert=true},()=>{opener.isConnected=false},()=>{opener.rect=false},()=>{document.hidden=true}]){
  opener.attrs={};opener.hidden=false;opener.inert=false;opener.isConnected=true;opener.rect=true;document.hidden=false;withdraw();assert.equal(completionControlReady(opener),false);
 }
});

test('disabled fieldsets retain their native first-legend exemption at completion',t=>{
 setup(t);const outer=new Element('fieldset');outer.disabled=true;
 const first=new Element('legend',outer),allowed=new Element('button',first),second=new Element('legend',outer),blocked=new Element('button',second);
 assert.equal(completionControlReady(allowed),true);assert.equal(completionControlReady(blocked),false);
 const nested=new Element('fieldset',outer);nested.disabled=true;const innerLegend=new Element('legend',nested),inner=new Element('button',innerLegend);
 assert.equal(completionControlReady(inner),false,'inner legend does not exempt the outer disabled fieldset');
 first.children.push(nested);outer.children=outer.children.filter(n=>n!==nested);nested.parentElement=first;
 assert.equal(completionControlReady(inner),true,'each disabled ancestor independently grants its first-legend exemption');
});
