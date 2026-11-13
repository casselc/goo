const {test}=require('node:test'),assert=require('node:assert/strict'),{actionVisible,revealFocus}=require('../ui/dom-visibility.mjs');
class E{
 constructor(tag,parent){this.tag=tag;this.parentElement=parent;this.children=[];parent?.children.push(this);this.open=false;this.hidden=false;this.ariaHidden=false;this.isConnected=true;this.focused=0;this.scrolled=0;this.rect=true;}
 closest(selector){for(let n=this;n;n=n.parentElement){if(selector==='details:not([open])'?n.tag==='details'&&!n.open:selector==='fieldset:disabled'?n.tag==='fieldset'&&n.disabled:(selector.includes('[hidden]')&&n.hidden)||(selector.includes('[aria-hidden')&&n.ariaHidden)||(selector.includes('[inert]')&&n.inert))return n}return null}
 querySelector(selector){return this.children.find(e=>e.tag===(selector?.includes('legend')?'legend':'summary'))||null}
 getAttribute(key){return key==='aria-disabled'?this.ariaDisabled??null:null}
 matches(selector){if(selector==='details:not([open])')return this.tag==='details'&&!this.open;if(selector!==':disabled')return false;if(this.disabled)return true;for(let f=this.parentElement;f;f=f.parentElement)if(f.tag==='fieldset'&&f.disabled&&!f.querySelector(':scope > legend')?.contains(this))return true;return false}
 contains(e){for(let n=e;n;n=n.parentElement)if(n===this)return true;return false}
 getClientRects(){return this.rect?[{}]:[]}
 focus(){this.focused++}
 scrollIntoView(){this.scrolled++}
}
function setup(t){const before=global.document;global.document={hidden:false};t.after(()=>{if(before===undefined)delete global.document;else global.document=before});const root=new E('section'),details=new E('details',root),summary=new E('summary',details),text=new E('span',summary),content=new E('button',details);return {root,details,summary,text,content}}
test('closed disclosure keeps its own first summary visible while its content stays unavailable',t=>{const {details,summary,text,content}=setup(t);assert.equal(actionVisible(summary),true);assert.equal(actionVisible(text),true);assert.equal(actionVisible(content),false);assert.equal(actionVisible(details),false);const second=new E('summary',details);assert.equal(actionVisible(second),false);details.open=true;assert.equal(actionVisible(content),true)});
test('a summary inside closed outer disclosure content is still unavailable',t=>{const {details,content}=setup(t),nested=new E('details',content),summary=new E('summary',nested);assert.equal(actionVisible(summary),false);details.open=true;assert.equal(actionVisible(summary),true);details.hidden=true;assert.equal(actionVisible(summary),false);details.hidden=false;details.ariaHidden=true;assert.equal(actionVisible(summary),false)});
test('disconnected, hidden-document and nonrendered actions remain unavailable',t=>{const {summary}=setup(t);summary.isConnected=false;assert.equal(actionVisible(summary),false);summary.isConnected=true;summary.rect=false;assert.equal(actionVisible(summary),false);summary.rect=true;document.hidden=true;assert.equal(actionVisible(summary),false)});
test('return focus chooses the visible closed-menu summary without opening it',t=>{const {details,summary,content}=setup(t),fallback=new E('button');assert.equal(revealFocus(summary,fallback),summary);assert.equal(summary.focused,1);assert.equal(summary.scrolled,1);assert.equal(details.open,false);assert.equal(fallback.focused,0);assert.equal(revealFocus(content,fallback),fallback);assert.equal(fallback.focused,1)});

test('420 explicit retained return reveals native ancestors outermost first and preserves ordinary fallback',t=>{
 const {details,content}=setup(t),nested=new E('details',details),summary=new E('summary',nested),button=new E('button',nested),fallback=new E('button');
 assert.equal(revealFocus(button,fallback),fallback);assert.equal(details.open,false);assert.equal(nested.open,false);
 assert.equal(revealFocus(button,fallback,{disclosures:true}),button);assert.equal(details.open,true);assert.equal(nested.open,true);assert.equal(button.focused,1);
});
test('420 explicit return cannot revive hidden, inert, disabled, detached or CSS-withdrawn origins',t=>{
 const {root,details,content}=setup(t),fallback=new E('button');
 for(const [target,key,value]of [[root,'hidden',true],[root,'inert',true],[content,'disabled',true],[content,'ariaDisabled','true'],[content,'isConnected',false]]){
  const before=target[key];target[key]=value;assert.equal(revealFocus(content,fallback,{disclosures:true}),fallback);assert.equal(details.open,false);target[key]=before;
 }
 const fieldset=new E('fieldset',root),menu=new E('details',fieldset),button=new E('button',menu);fieldset.disabled=true;
 assert.equal(revealFocus(button,fallback,{disclosures:true}),fallback);assert.equal(menu.open,false);
 const before=global.getComputedStyle;global.getComputedStyle=e=>({display:e===root?'none':'block',visibility:'visible'});t.after(()=>{if(before===undefined)delete global.getComputedStyle;else global.getComputedStyle=before});
 assert.equal(revealFocus(content,fallback,{disclosures:true}),fallback);assert.equal(details.open,false);
 fallback.disabled=true;assert.equal(revealFocus(content,fallback,{disclosures:true}),null);
});

test('420 explicit return to a menu summary keeps that menu closed',t=>{
 const {details,summary}=setup(t);assert.equal(revealFocus(summary,undefined,{disclosures:true}),summary);assert.equal(details.open,false);
});
