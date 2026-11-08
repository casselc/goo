import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {exemplarPage} from './support/histogram-exemplars-380.mjs';
const {node,Element}=await import('./support/group-chart-375.mjs');
Element.prototype.closest=function(selector){for(let e=this;e;e=e.parentElement)if(selector==='[inert]'&&e.inert)return e;return null};
const visuals=await import('../ui/visualizations.mjs');
mock.module('../ui/visualizations.mjs',{exports:{...visuals,actionVisible:e=>{for(let v=e;v;v=v.parentElement)if(v.hidden||v.inert)return false;return !!e?.isConnected}}});
const {appendReportedExemplars,renderTimeline}=await import('../ui/measurements.mjs');
globalThis.Node=Element;
const event=()=>({button:0,preventDefault(){},stopPropagation(){}});
function mounted({active=true,held=true}={}){const p=exemplarPage(),r=p.measurement,host=node('div'),calls=[];document.body.append(host);let current=active;appendReportedExemplars(host,r,r.histograms.series[0],{toNS:r.rows[0].end_ns},{heldExemplarContext:held,canAct:()=>current,onRecordContext:(...args)=>{calls.push(args);return true}});host.querySelector('.native-reported-exemplars').open=true;return{host,calls,activate:v=>current=v}}
test('explicit immutable exemplar capability exposes reported trace links without a live source adapter',()=>{
 const none=mounted({held:false});assert.equal(none.host.querySelector('a'),null);assert.match(none.host.textContent,/Trace navigation unavailable/);
 const f=mounted(),a=f.host.querySelector('a');assert(a);assert.equal(a.getAttribute('aria-disabled'),null);a.onclick(event());assert.equal(f.calls.length,1);assert.equal(f.calls[0][0].relation,'reported-exemplar');assert.notEqual(a.getAttribute('aria-disabled'),'true');assert.equal(f.host.querySelector('.native-exemplar-pause').hidden,true);assert.doesNotMatch(f.host.textContent,/Run measurement to reactivate/);a.onclick(event());assert.equal(f.calls.length,2,'the immutable owner, not a live-run pause, controls return leases');
});
test('Saved links rendered before their route activates remain dynamically guarded and become usable without rerender',()=>{
 const f=mounted({active:false}),a=f.host.querySelector('a');let prevented=0;const modified=()=>({button:1,preventDefault(){prevented++},stopPropagation(){},stopImmediatePropagation(){}});a.onauxclick(modified());assert.equal(prevented,1);assert.equal(f.calls.length,0);f.activate(true);a.onclick(event());assert.equal(f.calls.length,1);assert.notEqual(a.getAttribute('aria-disabled'),'true');a.onauxclick(modified());assert.equal(prevented,1,'fresh modified browser gestures keep native behavior');
 f.activate(false);a.onauxclick(modified());assert.equal(prevented,2);assert.equal(f.calls.length,1);
});

test('held Saved timeline selection stays unchanged while inactive, including stale custom selection events',()=>{
 const p=exemplarPage(),r=p.measurement,host=node('div');document.body.append(host);let active=false;renderTimeline(host,r.histograms.series[0],r,{heldExemplarContext:true,canAct:()=>active});
 const plot=host.querySelector('goo-value-timeline'),selection=host.querySelector('.native-timeline-selection');assert.equal(selection.hidden,true);assert.equal(plot.selectedKey,undefined);
 const select=key=>plot.dispatchEvent(new CustomEvent('goo-timeline-select',{detail:{key}}));select('1');assert.equal(selection.hidden,true);active=true;select('1');assert.equal(selection.hidden,false);assert.equal(plot.selectedKey,'1');assert.equal(selection.dataset.key,'1');
 const content=selection.children,selected=selection.dataset.key;active=false;select('0');assert.equal(plot.selectedKey,'1');assert.equal(selection.dataset.key,selected);assert.equal(selection.children,content);active=true;select('0');assert.equal(plot.selectedKey,'0');
});
