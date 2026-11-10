import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {resolve,dirname,basename} from 'node:path';
import {fileURLToPath,pathToFileURL} from 'node:url';
import {spawnSync} from 'node:child_process';
import * as before from './fixtures/distribution-values399-baseline.mjs';
import * as values from '../ui/distribution-values.mjs';
import {mount as groupedMount} from '../ui/grouped-time-series.mjs';

const root=fileURLToPath(new URL('../ui/',import.meta.url));
function dependencies(path){
 const source=readFileSync(path,'utf8'),out=[];
 for(const match of source.matchAll(/\b(?:import|export)\s+(?:[^'";]*?\s+from\s*)?(['"])([^'"]+)\1/g)){
  assert.ok(match[2].startsWith('.'),'browser module dependency must be local: '+match[2]);
  out.push(resolve(dirname(path),match[2]));
 }
 return out;
}
function reachable(start){const seen=new Set();function visit(path){if(seen.has(path))return;seen.add(path);for(const next of dependencies(path))visit(next)}visit(start);return seen;}
const helper=resolve(root,'distribution-values.mjs'),owner=resolve(root,'distributions.mjs'),grouped=resolve(root,'grouped-time-series.mjs');

test('399 pure display values have a bounded effect-free dependency graph and grouped renderer cannot import its owner',()=>{
 assert.deepEqual([...reachable(helper)].map(path=>basename(path)).sort(),['distribution-values.mjs','field-presentation.mjs','log-pattern-identity.mjs','numeric-values.mjs']);
 for(const path of reachable(helper)){
  const source=readFileSync(path,'utf8');
  assert.doesNotMatch(source,/\b(?:fetch|XMLHttpRequest|WebSocket|EventSource)\s*\(|\bimport\s*\(/,basename(path));
  assert.doesNotMatch(source,/\b(?:document|window|localStorage|sessionStorage|customElements)\s*[.\[]|\bdispatchEvent\s*\(/,basename(path));
 }
 assert.equal(reachable(grouped).has(owner),false,'direct and indirect owner edges must be removed');
 assert.equal(reachable(helper).has(resolve(root,'components.mjs')),false,'formatter must not import component registration');
 assert.equal(dependencies(owner).includes(grouped),true,'the application still composes the grouped renderer');
 const exports=readFileSync(owner,'utf8').match(/export\s*\{([^}]+)\}/g)?.join('\n')||'';
 assert.doesNotMatch(exports,/\b(?:percent|durationMS|groupMetric|compareMetric|metricWidth|valueLabel)\b/,'no helper compatibility facade in the application owner');
});

test('399 helper import and formatting need no browser, transport, store or registration globals',()=>{
 const script=`for(const key of ['document','window','fetch','XMLHttpRequest','WebSocket','EventSource','localStorage','sessionStorage','customElements','location'])Object.defineProperty(globalThis,key,{configurable:true,get(){throw Error('Unexpected effect '+key)}});const h=await import(${JSON.stringify(pathToFileURL(helper).href)});if(h.percent('1','4')!=='25.00'||h.durationMS('1000000')!=='1.000 ms'||h.valueLabel('雪')!=='"雪"')throw Error('Pure import output changed');`;
 const result=spawnSync(process.execPath,['--input-type=module','-e',script],{encoding:'utf8',timeout:15000});
 assert.equal(result.status,0,result.stderr||String(result.error));
});

const outcome=fn=>{try{return {value:fn()}}catch(error){return {error:{name:error.name,message:error.message}}}};
const same=(name,...args)=>assert.deepEqual(outcome(()=>values[name](...structuredClone(args))),outcome(()=>before[name](...structuredClone(args))),name+' '+JSON.stringify(args));
test('399 exact percentages and millisecond formatting preserve frozen production outputs and refusal boundaries',()=>{
 for(const args of [['0','0'],['0','1'],['1','3'],['2','3'],['1','20000'],['18446744073709551615','18446744073709551616'],['9007199254740993','9007199254740994'],['-1','3'],['1','-3'],['invalid','1']])same('percent',...args);
 for(const args of [[],[undefined],[null],['0'],['499'],['500'],['999'],['1000'],['1000000'],['21115000000','205'],['18446744073709551615'],['18446744073709551615','3'],['0','0'],['1','0'],['-1000000'],['1000000',null],['1000000','invalid']])same('durationMS',...args);
 for(let n=0;n<64;n++){const numerator=(2n**63n+BigInt(n)*7919n).toString(),denominator=String(n+1);same('percent',numerator,denominator);same('durationMS',numerator,denominator)}
});

test('399 metric values, exact rational ranking and width keep null, zero, signs, support and full-domain precision',()=>{
 const duration=(count,sum,p95='18446744073709551615')=>({state:'observed',count,sumNs:sum,p95Ns:p95,p99Ns:p95,maxNs:p95});
 const numeric={count:'5',valid:'3',sum:'-18446744073709551615',min:'-9223372036854775808',max:'1',scale:3,unit:'{row}',missing:'1',empty:'0',invalid:'1'};
 const groups=[{g0:null,count:'0'},{g0:'',count:'3',duration:duration('3','0')},{g0:'雪\\"',count:'2',duration:duration('2','18446744073709551616')},{g0:'api',count:'3',duration:duration('3','27670116110564327425'),spanStatus:{error:'1'}},{g0:'negative',count:'5',numeric},{g0:'invalid',count:'5',numeric:{...numeric,valid:'0'}},{g0:'unknown',count:'2',duration:{state:'unknown'}}];
 for(const metric of ['count','mean','p95','p99','max','error-rate','numeric-mean','numeric-sum','numeric-min','numeric-max']){
  for(const group of groups)same('groupMetric',group,metric);
  for(const a of groups)for(const b of groups)same('compareMetric',a,b,metric);
 }
 const rationals=[null,{n:0n,d:1n},{n:1n,d:3n},{n:2n,d:3n},{n:18446744073709551616n,d:2n},{n:27670116110564327425n,d:3n}];
 for(const value of rationals)for(const max of rationals){
  assert.deepEqual(outcome(()=>values.metricWidth(value,max)),outcome(()=>before.metricWidth(value,max)));
 }
 const a=values.groupMetric(groups[2],'mean'),b=values.groupMetric(groups[3],'mean');
 assert.equal(values.metricWidth(a,b),'99.99','rounding labels must not reorder exact rational geometry');
 assert.equal(values.metricWidth(null,b),'0.00');
});

test('399 literal category and pattern labels preserve qualified missing, empty, Unicode and escaping',()=>{
 for(const field of [undefined,{kind:'column',key:'service'},{kind:'attribute',key:'work.class'},{kind:'resource',key:'service.namespace'}])for(const value of [null,'','雪','"quoted"','<script>','a\\b','line\nbreak'])same('valueLabel',value,field);
 for(const id of ['lp1:e','lp1:o','lp1:u','lp1:p:6f7264657220n','lp1:p:','untrusted'])same('valueLabel',id,{kind:'column',key:'body_pattern_v1'});
});

class Element{
 constructor(tag){this.tagName=tag;this.children=[];this.dataset={};this.attributes={};this.hidden=false;this.style={setProperty:(key,value)=>{this.style[key]=value}};this.classList={add:()=>{}};this.listeners={};}
 append(...children){this.children.push(...children)}
 replaceChildren(...children){this.children=children}
 setAttribute(key,value){this.attributes[key]=value}
 addEventListener(name,callback){this.listeners[name]=callback}
 querySelector(tag){return walk(this).find(node=>node!==this&&node.tagName===tag)||null}
 get value(){return this._value??(this.tagName==='select'?this.children[0]?.value:'')??''}
 set value(value){this._value=value}
 set textContent(value){this._text=value;this.children=[]}
 get textContent(){return this._text??this.children.map(child=>typeof child==='string'?child:child.textContent).join('')}
}
function walk(node){return [node,...node.children.filter(child=>child&&typeof child==='object').flatMap(walk)]}
function render(data,options={}){const old=globalThis.document;globalThis.document={createElement:tag=>new Element(tag)};try{const host=new Element('section');groupedMount(host,data,options);return host}finally{globalThis.document=old}}
function supplied(){return {state:'observed',count:'6',question:{signal:'spans',fromMs:1000,toMs:3000,groupBy:[{kind:'attribute',key:'work.class'}]},duration:{state:'observed'},groupedTimeSeries:[{values:['雪'],points:[{fromMs:1000,toMs:2000,count:'2',state:'observed',duration:{state:'observed',count:'2',sumNs:'2000000',p95Ns:'1000000'}},{fromMs:2000,toMs:3000,count:'0',state:'no_data',duration:{state:'no_data'}}]},{values:[null],points:[{fromMs:1000,toMs:2000,count:'4',state:'observed',duration:{state:'observed',count:'4',sumNs:'16000000',p95Ns:'4000000'}},{fromMs:2000,toMs:3000,count:'0',state:'no_data',duration:{state:'no_data'}}]}]}}

test('399 fallback grouped rendering keeps a full-population scale when local Find hides its maximum',()=>{
 const old=globalThis.document;globalThis.document={createElement:tag=>new Element(tag)};
 try{const data=supplied(),host=new Element('section'),beforeData=structuredClone(data);groupedMount(host,data);const all=walk(host),picker=all.find(node=>node.tagName==='select'),search=all.find(node=>node.attributes['aria-label']==='Find a captured series');picker.value='mean';picker.onchange();const cards=walk(host).filter(node=>node.className==='series-card'),small=cards.find(card=>card.dataset.series==='["雪"]'),large=cards.find(card=>card.dataset.series==='[null]');const bar=walk(small).find(node=>node.tagName==='button');assert.equal(bar.style['--bar-height'],'25.00%');assert.match(bar.title,/1\.000 ms/);search.value='雪';search.oninput();assert.equal(large.hidden,true);assert.equal(small.hidden,false);assert.equal(bar.style['--bar-height'],'25.00%');assert.deepEqual(data,beforeData)}finally{globalThis.document=old}
});

test('399 fallback heatmap returns supplied rectangle coordinates to its application owner without transforming scope',()=>{
 const data=supplied();data.question.groupBy=[{kind:'column',key:'duration_bucket'}];data.timeSeries=[{fromMs:1000,toMs:2000,count:'6'}];data.groupedTimeSeries=[{values:['<1ms'],points:[{fromMs:1000,toMs:2000,count:'6'}]}];
 let received;const host=render(data,{onRectangle:range=>received=range}),grid=walk(host).find(node=>node.tagName==='goo-bucket-heatmap'),range={fromColumn:'0',toColumn:'0',fromRow:'<1ms',toRow:'<1ms'};
 assert.ok(grid,'actual fallback heatmap mounted');grid.listeners['goo-heatmap-range']({detail:range});assert.equal(received,range,'renderer must not import or invoke the application selector transform');
 const ownerSource=readFileSync(owner,'utf8');assert.match(ownerSource,/onRectangle:[^\n]*PopulationSelection\.rectangle\(data,range\)/,'owner applies the unchanged typed rectangle transform');
});
