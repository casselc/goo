import {mock} from 'node:test';
import {readFileSync} from 'node:fs';
import * as Context from '../../ui/native-record-context.mjs';
class Element {
 constructor(tag='div'){this.localName=tag;this.children=[];this.isConnected=true;this.hidden=false;this.open=false;this.listeners={};this.selectedKey='held-interval';this.dataset={}}
 append(...children){this.children.push(...children);for(const child of children)if(child&&typeof child==='object')child.parentElement=this}
 setAttribute(){}
 addEventListener(type,fn){this.listeners[type]=fn}
 closest(selector){return selector==='[inert]'&&this.inert?this:null}
}
let rendered;
const components=await import('../../ui/components.mjs'),visuals=await import('../../ui/visualizations.mjs'),native=await import('../../ui/native-presentation.mjs'),metrics=await import('../../ui/measurements.mjs');
mock.module('../../ui/components.mjs',{exports:{...components,node:(tag)=>new Element(tag),stats:()=>new Element()}});
mock.module('../../ui/native-presentation.mjs',{exports:{...native,nativeSummary:()=>[]}});
mock.module('../../ui/visualizations.mjs',{exports:{...visuals,actionVisible:node=>!!node?.isConnected&&!node.hidden}});
mock.module('../../ui/measurements.mjs',{exports:{...metrics,streamDiagnostics:()=>({attention:false}),render(host,result,options){rendered={host,result,options,anchor(detail){const a=new Element('a');host.append(a);a.onclick=options.bindRecordContextAction(()=>options.onRecordContext(detail,a),a);return a}}}}});
const {renderNative}=await import('../../ui/service-dashboard.mjs');
function measurement(){
 const f=JSON.parse(readFileSync(new URL('../fixtures/native-record-context.json',import.meta.url))).find(x=>!x.refused&&x.parent.definition.kind==='histogram');
 return{question:f.parent,basis:{source:'clickhouse',coverage:'unknown'},readAt:'2026-11-07T10:00:00Z',histograms:{series:[{stream:f.stream,intervals:[{fromNS:f.fromNS,toNS:f.toNS,state:'ok'}]}]}};
}
function detail(r,signal='spans'){const s=r.histograms.series[0],i=s.intervals[0];return{...Context.plan(r.question,s.stream,signal,i.fromNS,i.toNS),measurement:r.question,stream:s.stream,basis:r.basis,readAt:r.readAt}}
function mount({accept=true,data=measurement()}={}){
 const box=new Element(),opened=[];let current=true,owner;
 renderNative(box,data,{current:()=>current,onContextOwner:value=>owner=value,onRecordContext:(...args)=>{opened.push(args);return accept}});
 const disclosure=box.children.at(-1);disclosure.open=true;disclosure.listeners.toggle();
 return{data,box,disclosure,view:rendered,opened,owner,invalidate:()=>current=false};
}

export {Element,measurement,detail,mount};
