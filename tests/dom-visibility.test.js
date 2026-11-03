const {test}=require('node:test'),assert=require('node:assert/strict'),{actionVisible,revealFocus}=require('../ui/dom-visibility.mjs');
class E{
 constructor(tag,parent){this.tag=tag;this.parentElement=parent;this.children=[];parent?.children.push(this);this.open=false;this.hidden=false;this.ariaHidden=false;this.isConnected=true;this.focused=0;this.scrolled=0;this.rect=true;}
 closest(selector){for(let n=this;n;n=n.parentElement){if(selector==='details:not([open])'?n.tag==='details'&&!n.open:n.hidden||n.ariaHidden)return n}return null}
 querySelector(){return this.children.find(e=>e.tag==='summary')||null}
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
