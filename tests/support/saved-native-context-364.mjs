import {mock} from 'node:test';
import {readFileSync} from 'node:fs';
export class Element{
 constructor(){this.isConnected=true;this.hidden=false;this.children=[];this.onclick=null;this.dataset={};this.selectedKey='held-selection'}
 contains(node){return this.children.includes(node)}
 closest(selector){return selector==='[inert]'&&this.inert?this:null}
 replaceChildren(){for(const c of this.children)c.isConnected=false;this.children=[]}
}
export const renders=[];
mock.module('../../ui/visualizations.mjs',{exports:{actionVisible:node=>node?.isConnected&&!node.hidden}});
mock.module('../../ui/measurements.mjs',{exports:{render(host,result,options){const mount={host,result,options,anchor(){const a=new Element();host.children.push(a);const original=()=>options.onRecordContext(mount.detail,a);a.onclick=options.bindRecordContextAction(original,a);return a}};renders.push(mount)}}});
export const Saved=await import('../../ui/saved-native-record-context.mjs');
export const Context=await import('../../ui/native-record-context.mjs');
export function page(){const f=JSON.parse(readFileSync(new URL('../fixtures/native-record-context.json',import.meta.url))).find(x=>!x.refused&&x.parent.definition.kind==='counter'),q=structuredClone(f.parent),measurement={question:q,rows:[],basis:{source:'retained'},readAt:'2026-11-07T10:00:00Z',counters:{series:[{stream:structuredClone(f.stream),intervals:[{fromNS:f.fromNS,toNS:f.toNS,state:'ok'}]}]}};return{receipt:{version:1,id:'a'.repeat(64),title:'Retained counter',capturedAt:'2026-11-10T07:00:00Z',question:{kind:'measurement',measurement:structuredClone(q)}},measurement,retainedRecords:measurement.rows.length,recordsLimited:false}}
export function detail(p=page()){const r=p.measurement,s=r.counters.series[0],i=s.intervals[0];return{...Context.plan(r.question,s.stream,'logs',i.fromNS,i.toNS),measurement:r.question,stream:s.stream,basis:r.basis,readAt:r.readAt}}
export function fixture(){let active=true;const host=new Element(),opened=[],owner=Saved.mount(host,{active:()=>active,onOpenContext:(...args)=>{opened.push(args);return true}});return{host,owner,opened,active:value=>active=value,render:()=>renders.at(-1)}}
