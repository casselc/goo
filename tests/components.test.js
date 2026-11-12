const {test}=require('node:test'),assert=require('node:assert/strict');
const {latencyModel,durationMS}=require('../ui/components.mjs');
test('latency display preserves exact strings, zero and absent quantiles without inventing p90',()=>{
 const summary={state:'observed',count:'130',sumNs:'3250000000',p50Ns:'10000000',p95Ns:'40000000',p99Ns:'40000000',minNs:'10000000',maxNs:'40000000'},before=JSON.stringify(summary),model=latencyModel(summary);
 assert.equal(model.mean,'25.000 ms');assert.deepEqual(model.quantiles.map(q=>[q.label,q.ns,q.position]),[['p50','10000000',25],['p95','40000000',100],['p99','40000000',100]]);assert.equal(JSON.stringify(summary),before);
 assert.deepEqual(latencyModel({...summary,p95Ns:undefined}).quantiles.map(q=>q.label),['p50','p99']);assert.equal(latencyModel({...summary,count:'0'}).state,'unavailable');assert.equal(latencyModel({...summary,count:'invalid'}).state,'unavailable');assert.equal(latencyModel({state:'no_data',count:'0'}).state,'no_data');assert.deepEqual(latencyModel({state:'observed',count:'2',sumNs:'0',p50Ns:'0'}).quantiles[0],{label:'p50',ns:'0',value:'0.000 ms',position:0});assert.equal(durationMS('18446744073709551615'),'18446744073709.552 ms');
});
test('caller-supplied latency domain aligns independent summaries without rounding geometry',()=>{
 const base={state:'observed',count:'2',sumNs:'4',p50Ns:'2',p95Ns:'2',p99Ns:'2',minNs:'2',maxNs:'2'},next={...base,sumNs:'16',p50Ns:'8',p95Ns:'8',p99Ns:'8'};
 assert.equal(latencyModel(base,'8').quantiles[0].position,25);assert.equal(latencyModel(next,'8').quantiles[0].position,100);assert.equal(latencyModel(base,'8').ceilingNs,'8');assert.equal(latencyModel(next,'4').state,'unavailable');assert.equal(latencyModel(base,'invalid').state,'unavailable');assert.equal(latencyModel({...base,sumNs:'0',p50Ns:'0',p95Ns:'0',p99Ns:'0'},'0').quantiles[0].position,0);
});


test('inspector Escape preserves editor cancellation and dismisses only an unhandled detail gesture',t=>{
 const previousDocument=globalThis.document,previousNode=globalThis.Node;
 class Element{
  constructor(tag){this.tagName=tag.toUpperCase();this.children=[];this.dataset={};this.listeners=new Map();this.attributes={};this.isConnected=true;this.classList={add(){}}}
  append(...children){this.children.push(...children)}replaceChildren(...children){this.children=[...children]}get firstChild(){return this.children[0]}
  setAttribute(name,value){this.attributes[name]=String(value)}
  addEventListener(type,listener){this.listeners.set(type,listener)}focus(){this.focused=true}click(){this.onclick?.()}
 }
 globalThis.Node=Element;globalThis.document={createElement:tag=>new Element(tag)};
 t.after(()=>{if(previousDocument===undefined)delete globalThis.document;else globalThis.document=previousDocument;if(previousNode===undefined)delete globalThis.Node;else globalThis.Node=previousNode});
 const {inspector}=require('../ui/components.mjs');
 const editor=(type,value='')=>{const target={type,value};target.closest=selector=>selector==='input'?(type==='search'||type==='text'?target:null):['textarea','select','contenteditable','combobox','listbox'].includes(type)?target:null;return target};
 const plain={closest:()=>null};
 const cases=[
  ['heading',{},true],['button',{},true],['empty search',{target:editor('search')},true],
  ['nonempty search',{target:editor('search','query')},false],['text input',{target:editor('text')},false],
  ...['textarea','select','contenteditable','combobox','listbox'].map(type=>[type,{target:editor(type)},false]),
  ['IME composition',{isComposing:true},false],['legacy IME composition',{keyCode:229},false],
  ['handled Escape',{defaultPrevented:true},false],['other key',{key:'Enter'},false]
 ];
 for(const [label,overrides,dismiss]of cases){
  const host=new Element('section'),opener=new Element('button');let closed=0,prevented=0,stopped=0;const view=inspector(host,{onClose:()=>{closed++}});
  view.show({key:'occurrence',title:'Selected record',content:new Element('div'),opener});
  host.listeners.get('keydown')({key:'Escape',defaultPrevented:false,target:plain,preventDefault(){prevented++},stopPropagation(){stopped++},...overrides});
  assert.equal(host.hidden,dismiss,label);assert.equal(closed,Number(dismiss),label);assert.equal(prevented,Number(dismiss),label);assert.equal(stopped,Number(dismiss),label);assert.equal(!!opener.focused,dismiss,label);
 }
});
