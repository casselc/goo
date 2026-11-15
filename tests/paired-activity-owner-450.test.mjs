import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
const tails=[],filters=[];
let starting;
const canonical=value=>JSON.stringify(value);
mock.module('../ui/components.mjs',{exports:{node:(tag,text,className)=>{const node=new Element(tag);if(text!==undefined)node.textContent=text;if(className)node.className=className;return node},inspector:host=>{
 const close=new Element('button'),heading=new Element('h4');close.setAttribute('aria-label','Close selected item details');close.onclick=()=>{};host.append(close,heading);const query=host.querySelector.bind(host);host.querySelector=selector=>selector==='.ui-inspector-heading>h4'?heading:query(selector);return {clear(){},show(){}};
}}});
mock.module('../ui/tail.mjs',{exports:{canonical,transportIdentity:canonical,mount:(host,options)=>{
 const state={selector:null,rows:['held'],find:{text:'needle',revision:4},connection:'connected',generation:3,starts:[],stages:[]};
 const tail={options,state,ready:()=>true,scope:()=>({selector:state.selector}),snapshot:()=>structuredClone(state),layout:()=>({columns:['utc','service','summary'],fieldColumns:[]}),stage(selector){state.selector=structuredClone(selector);state.stages.push(structuredClone(selector));state.rows=[]},start(selector){state.starts.push(structuredClone(selector));state.selector=structuredClone(selector);state.generation++;starting?.(options);return true},pause(){state.connection='paused'},reset(){state.selector=null;state.rows=[]},setReadBlocked(){},focusStart(){},dispose(){}};tails.push(tail);return tail;
}}});
mock.module('../ui/paired-lane-filter.mjs',{exports:{mount:(host,options)=>{
 let where,dirty=false;const filter={options,resolved:0,retired:0,where:()=>structuredClone(where),dirty:()=>dirty,condition:()=>({where,dirty}),setDirty(value){dirty=value},resolve(){filter.resolved++;return true},reset(value){where=structuredClone(value);dirty=false},retire(){filter.retired++},sync(){},dispose(){},apply(value,opener,capability={current:()=>true}){const accepted=options.onApply(value,opener,capability);if(accepted)where=structuredClone(value);return accepted}};filters.push(filter);return filter;
}}});
const {mount}=await import('../ui/paired-activity.mjs');
const leaf=(key,value='17')=>({field:{kind:'column',key},op:'gte',value});
function fixture(){
 const body=new Element('body');body.root=true;globalThis.Node=Element;globalThis.document={body,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};
 const host=new Element('main'),opener=new Element('button');body.append(host,opener);let active=true;starting=undefined;
 const owner=mount(host,{current:()=>active}),laneTails=tails.slice(-2),laneFilters=filters.slice(-2);
 owner.stage({environment:'prod'});return {owner,opener,logs:laneTails[0],spans:laneTails[1],logFilter:laneFilters[0],spanFilter:laneFilters[1],deactivate:()=>active=false};
}
test('450 actual paired owner Apply starts only its lane and leaves sibling bank/find/connection/generation intact',()=>{
 const f=fixture(),before=structuredClone(f.spans.state);assert.equal(f.logFilter.apply(leaf('severity_number'),f.opener),true);
 assert.equal(f.logs.state.starts.length,1);assert.deepEqual(f.spans.state,before);assert.deepEqual(f.owner.layout().logs.where,leaf('severity_number'));
});
test('450 actual paired owner malformed or composed overbudget conditions refuse before source starts',()=>{
 const f=fixture();for(const condition of [leaf('duration_ns'),{all:Array.from({length:33},()=>leaf('severity_number'))}])assert.throws(()=>f.logFilter.apply(condition,f.opener));
 f.owner.stage({environment:'prod',where:{all:Array.from({length:30},()=>({field:{kind:'resource',key:'region'},op:'eq',value:'west'}))}});
 assert.throws(()=>f.logFilter.apply(leaf('severity_number'),f.opener),/32 conditions/);
 assert.equal(f.logs.state.starts.length,0);assert.equal(f.spans.state.starts.length,0);assert.equal(f.owner.layout().logs.where,undefined);
});
test('450 shared scope replacement rebuilds both selectors with their applied lane conditions',()=>{
 const f=fixture();f.logFilter.apply(leaf('severity_number'),f.opener);f.spanFilter.apply(leaf('duration_ns','100'),f.opener);
 assert.equal(f.owner.start({environment:'next'}),true);
 for(const [tail,key,value]of [[f.logs,'severity_number','17'],[f.spans,'duration_ns','100']]){assert.equal(tail.state.selector.environment,'next');assert.deepEqual(tail.state.selector.where,leaf(key,value));}
 assert.ok(f.logFilter.retired>0);assert.ok(f.spanFilter.retired>0);
});
test('450 synchronous reset or capability withdrawal during own start cannot commit old applied condition',()=>{
 for(const reset of [false,true]){const f=fixture();let alive=true;starting=()=>{if(reset)f.owner.reset();else alive=false};
 assert.equal(f.logFilter.apply(leaf('severity_number'),f.opener,{current:()=>alive}),false);assert.equal(f.owner.layout().logs.where,undefined);
 if(reset)assert.equal(f.owner.snapshot().selector,null);else assert.deepEqual(f.owner.snapshot().selector,{environment:'prod'});
 }
});
test('450 pending lane draft Resume resolves local controls and starts neither lane',()=>{
 const f=fixture();f.logFilter.setDirty(true);assert.equal(f.logs.options.onResume(),false);assert.equal(f.logFilter.resolved,1);assert.equal(f.logs.state.starts.length,0);assert.equal(f.spans.state.starts.length,0);
});
