import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mount} from '../ui/native-capture-layout.mjs';

class Element{
 constructor(id,document){this.id=id;this.ownerDocument=document;this.children=[];this.attributes=new Map();this.hidden=false;this.moves=0}
 get parentElement(){return this.parent??null}get firstChild(){return this.children[0]??null}get nextSibling(){const list=this.parent?.children;return list?.[list.indexOf(this)+1]??null}
 setAttribute(key,value){this.attributes.set(key,String(value))}getAttribute(key){return this.attributes.get(key)??null}
 insertBefore(node,before){if(node===before)return;node.remove();const at=before===null?this.children.length:this.children.indexOf(before);assert(at>=0);this.children.splice(at,0,node);node.parent=this}
 moveBefore(node,before){this.moves++;this.insertBefore(node,before)}append(...nodes){for(const node of nodes)this.insertBefore(node,null)}prepend(node){this.insertBefore(node,this.firstChild)}before(node){this.parent.insertBefore(node,this)}after(node){this.parent.insertBefore(node,this.nextSibling)}remove(){if(this.parent){this.parent.children=this.parent.children.filter(node=>node!==this);this.parent=null}}
}
function fixture(){const document={createElement(){return new Element('',document)}},node=id=>new Element(id,document),parent=node('saved'),toolbar=node('toolbar'),library=node('catalog'),capture=node('capture'),documents=node('documents'),recovery=node('recovery'),browse1=node('browse-investigations'),browse2=node('browse-evidence'),heading=node('heading'),status=node('status'),scope=node('scope'),measurement=node('measurement'),plot=node('plot');toolbar.append(browse1,browse2,recovery);capture.append(heading,status,scope,measurement);measurement.append(plot);parent.append(toolbar,library,capture,documents);plot.selection='unknown-interval';plot.data={basis:'held-basis',value:null};document.activeElement=plot;return{document,parent,toolbar,library,capture,scope,recovery,plot,layout:mount({capture,library,toolbar,recovery,scope})}}
const visible=node=>node.children.filter(child=>!child.hidden).map(child=>child.id);
const moves=f=>[f.parent,f.toolbar,f.capture].reduce((sum,node)=>sum+node.moves,0);

test('accepted native ordering prioritizes existing evidence and preserves selected custom-element state and focus',()=>{
 const f=fixture(),data=f.plot.data,original=f.plot;f.layout.setNative(true);assert.deepEqual(visible(f.parent),['capture','toolbar','catalog','documents']);assert.deepEqual(visible(f.capture),['recovery','heading','status','measurement','scope']);assert.equal(f.capture.getAttribute('data-native-measurement'),'true');assert.equal(f.document.activeElement,original);assert.equal(f.plot.data,data);assert.equal(f.plot.selection,'unknown-interval');const moved=moves(f);assert(moved>0);f.layout.setNative(true);assert.equal(moves(f),moved,'repeated native presentation must not reconnect existing plots');assert.equal(f.document.activeElement,original);
 f.layout.setNative(false);assert.deepEqual(visible(f.parent),['toolbar','catalog','capture','documents']);assert.deepEqual(visible(f.capture),['heading','status','scope','measurement']);assert.deepEqual(visible(f.toolbar),['browse-investigations','browse-evidence','recovery']);assert.equal(f.capture.getAttribute('data-native-measurement'),'false');assert.equal(f.plot,original);assert.equal(f.plot.data,data);assert.equal(f.plot.selection,'unknown-interval');assert.equal(f.document.activeElement,original);const restored=moves(f);f.layout.setNative(false);assert.equal(moves(f),restored);
});

test('a recovery notice that becomes visible later remains ahead of held evidence and restores without changing visibility',()=>{
 const f=fixture();f.recovery.hidden=true;f.layout.setNative(true);assert.equal(f.recovery.parentElement,f.capture);assert.equal(f.recovery.hidden,true);f.recovery.hidden=false;assert.equal(visible(f.capture)[0],'recovery');f.layout.setNative(false);assert.equal(f.recovery.parentElement,f.toolbar);assert.equal(f.recovery.hidden,false);
});
