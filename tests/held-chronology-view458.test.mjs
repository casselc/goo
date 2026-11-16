import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Element} from './support/raw-observation-dom.mjs';
import {prepare} from '../ui/held-chronology.mjs';
const events=[],registry=new Map();
class NativeElement extends Element{
 constructor(tag='element'){super(tag)}
 matches(selector){const match=selector.match(/^\[data-([a-z-]+)(?:="([^"]*)")?\]$/);if(match){const value=this.dataset[match[1].replace(/-([a-z])/g,(_,c)=>c.toUpperCase())];return match[2]===undefined?value!==undefined:value===match[2]}return super.matches(selector)}
 setAttribute(key,value){const old=this.getAttribute(key);super.setAttribute(key,value);if(this.constructor.observedAttributes?.includes(key)&&old!==String(value))this.attributeChangedCallback?.(key,old,String(value))}
 removeAttribute(key){const old=this.getAttribute(key);super.removeAttribute(key);if(this.constructor.observedAttributes?.includes(key)&&old!==null)this.attributeChangedCallback?.(key,old,null)}
 dispatchEvent(event){events.push({target:this,event});return true}
}
globalThis.HTMLElement=NativeElement;globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
globalThis.CustomEvent=class{constructor(type,init){this.type=type;Object.assign(this,init)}};
globalThis.document={hidden:false,activeElement:null,createElement:tag=>new NativeElement(tag)};
await import('../ui/held-chronology-view.mjs');
const View=registry.get('goo-held-chronology');
function source(signal,count=2){
 const connection=(signal==='logs'?'a':'b').repeat(64),readWindow={fromMs:1000,toMs:61000},observedAt='1970-01-01T00:01:01.001Z';
 return {phase:'paused',staged:false,selector:{signal,environment:'local'},find:{text:'',revision:0},readWindow,completedRead:{...readWindow,rowCount:count,limited:false,observedAt,revision:1},connection,evicted:0,gap:false,rows:Array.from({length:count},(_,index)=>({key:connection+':'+(index+1),record:{timestamp_ns:String(1001000000+index),ServiceName:'worker',...(signal==='logs'?{Body:'message '+index,SeverityText:'INFO'}:{SpanName:'operation '+index,StatusCode:'Ok',duration_ns:'100'})},readWindow,observedAt,limited:false,late:false}))};
}
function data(logs=2,spans=2){return prepare({logs:source('logs',logs),spans:source('spans',spans)})}
function mount(value=data()){const view=new View();view.localName='goo-held-chronology';view.data=value;view.root=true;view.connectedCallback();return view}
const rows=view=>view.querySelectorAll('[data-chronology-key]');
const controls=view=>view.querySelectorAll('button').filter(button=>!button.dataset.chronologyInspect);
const find=view=>view.querySelector('input');
test('one passive mixed record table preserves exact occurrence references, cloned readback and source metadata',()=>{
 const input=data(),view=mount(input);assert.equal(rows(view).length,4);assert.deepEqual(rows(view).map(row=>row.dataset.signal),['logs','spans','logs','spans']);assert.equal(view.querySelectorAll('table').length,1);assert.equal(view.querySelectorAll('.held-chronology-source').length,2);assert.match(view.textContent,/Logs.*2 held.*Paused/);assert.match(view.textContent,/Latest read UTC/);assert.match(view.textContent,/Original Activity|Latest read window/);assert.equal(view.data.version,'goo.held-activity/1');const read=view.data;read.lanes.logs.rows[0].record.Body='forged';assert.equal(view.data.lanes.logs.rows[0].record.Body,'message 0');
 const inspect=view.querySelectorAll('button').find(button=>button.dataset.chronologyInspect);inspect.onclick();assert.equal(events.at(-1).target,inspect);assert.equal(events.at(-1).event.type,'goo-chronology-inspect');assert.deepEqual(events.at(-1).event.detail,{signal:'logs',key:input.lanes.logs.rows[1].key});assert.equal(events.at(-1).event.bubbles,true);
 view.selectedKey=rows(view)[0].dataset.chronologyKey;assert.equal(rows(view)[0].getAttribute('aria-current'),'true');view.selectedKey=rows(view)[1].dataset.chronologyKey;assert.equal(rows(view)[0].getAttribute('aria-current'),null);assert.equal(rows(view)[1].getAttribute('aria-current'),'true');
});
test('Find is local over held record text, preserves its input focus and accounts for each lane',()=>{
 const input=structuredClone(data());input.lanes.logs.find.text='lane filter';const view=mount(input),search=find(view);search.value='message 0';search.focus();search.oninput();assert.equal(rows(view).length,1);assert.equal(document.activeElement,search);assert.equal(view.dataset.logs,'1');assert.equal(view.dataset.spans,'0');assert.match(view.textContent,/Logs 1\/2 · Spans 0\/2 held/);assert.match(view.textContent,/lane Find not applied/);search.value='not present';search.oninput();assert.equal(rows(view).length,0);assert.match(view.textContent,/No matching held records/);search.value='';search.oninput();assert.equal(rows(view).length,4);
});
test('new supplied data resets local pages and retired, hidden, disconnected controls cannot act',()=>{
 const view=mount(data(100,100)),[previous,next]=controls(view);assert.equal(rows(view).length,100);next.onclick();assert.match(view.textContent,/Rows 101–200/);assert.equal(document.activeElement,previous);const retiredInspect=view.querySelectorAll('button').find(button=>button.dataset.chronologyInspect),before=events.length;view.data=data();assert.equal(rows(view).length,4);next.onclick();retiredInspect.onclick();assert.equal(events.length,before);const inspect=view.querySelectorAll('button').find(button=>button.dataset.chronologyInspect);view.hidden=true;inspect.onclick();assert.equal(events.length,before);view.hidden=false;view.root=false;inspect.onclick();assert.equal(events.length,before);view.root=true;inspect.disabled=true;inspect.onclick();assert.equal(events.length,before);
});
test('unavailable event times remain explicitly separated and inspectable rather than sorting as zero',()=>{
 const input=structuredClone(data());input.lanes.logs.rows[0].record.timestamp_ns='bad';delete input.lanes.spans.rows[0].record.timestamp_ns;const view=mount(input);assert.deepEqual(rows(view).map(row=>row.dataset.placed),['true','true','false','false']);assert.equal(view.dataset.unplaced,'2');assert.match(view.textContent,/Timestamp unavailable/);assert.equal(view.querySelectorAll('[data-chronology-inspect]').length,4);
});
test('SSR, detached properties and malformed data honor the same closed admission without source reads',()=>{
 let reads=0;globalThis.fetch=()=>{reads++;throw Error('Source-free display')};const view=new View();view.setAttribute('data',JSON.stringify(data()));view.root=true;view.connectedCallback();assert.equal(rows(view).length,4);const early=new View();Object.defineProperty(early,'data',{value:data(),configurable:true,writable:true});early.root=true;early.connectedCallback();assert.equal(rows(early).length,4);view.setAttribute('data','{"version":"forged"}');assert.equal(rows(view).length,0);assert.equal(view.dataset.state,'unavailable');view.data=data();assert.equal(rows(view).length,4);assert.equal(reads,0);delete globalThis.fetch;
});
test('one-sided delivery, gaps and evictions remain visible without pretending the missing lane is empty success',()=>{
 const input=structuredClone(data(1,0));input.lanes.logs.gap=true;input.lanes.logs.evicted=3;input.lanes.spans.completedRead=null;input.lanes.spans.readWindow=null;input.lanes.spans.phase='unavailable';const view=mount(input);assert.equal(rows(view).length,1);assert.match(view.textContent,/continuity gap.*3 evicted/);assert.match(view.textContent,/Spans · 0 held · Unavailable · not read/);assert.match(view.textContent,/Latest readUnavailable/);
});

test('retained rows or an earlier window do not become never-read when the latest read is unavailable',()=>{
 const input=structuredClone(data(1,0));input.lanes.logs.completedRead=null;input.lanes.logs.phase='unavailable';input.lanes.logs.gap=true;input.lanes.spans.completedRead=null;input.lanes.spans.phase='paused';const view=mount(input),sources=view.querySelectorAll('.held-chronology-source');assert.match(sources[0].querySelector('summary').textContent,/Logs · 1 held · Unavailable · latest read unavailable/);assert.match(sources[1].querySelector('summary').textContent,/Spans · 0 held · Paused · latest read unavailable/);assert.doesNotMatch(sources[0].querySelector('summary').textContent,/not read/i);assert.doesNotMatch(sources[1].querySelector('summary').textContent,/not read/i);assert.equal(rows(view).length,1);
});

test('an explicitly unconfigured sibling remains distinct from a scoped empty successful read',()=>{
 const input=structuredClone(data(1,0));Object.assign(input.lanes.spans,{selector:null,phase:'paused',staged:false,readWindow:null,completedRead:null,connection:null,evicted:0,gap:false,find:{text:'',revision:0}});const view=mount(input),sources=view.querySelectorAll('.held-chronology-source');assert.equal(rows(view).length,1);assert.match(sources[1].querySelector('summary').textContent,/Spans · 0 held · Not configured/);assert.match(sources[1].textContent,/ScopeNot configured/);assert.doesNotMatch(sources[1].textContent,/Source condition|All records in scope/);assert.equal(view.data.lanes.spans.selector,null);
});

test('one atomic announcement combines current signal matches and page range',()=>{
 const view=mount(data(100,100)),status=view.querySelectorAll('[role="status"]');assert.equal(status.length,1);assert.equal(status[0].getAttribute('aria-atomic'),'true');assert.match(status[0].textContent,/Logs 100\/100 · Spans 100\/100 held · Rows 1–100/);
 const [,next]=controls(view);next.onclick();assert.match(status[0].textContent,/Rows 101–200/);assert.match(status[0].textContent,/Logs 100\/100 · Spans 100\/100 held/);const search=find(view);search.value='message 0';search.oninput();assert.match(status[0].textContent,/Logs 1\/100 · Spans 0\/100 held · Rows 1–1/);assert.equal(view.querySelectorAll('[role="status"]').length,1);
 search.value='x'.repeat(257);search.oninput();assert.equal(search.getAttribute('aria-invalid'),'true');assert.match(status[0].textContent,/at most 256/);
});

test('related held scope ignores Find, retains duplicate observations and restores page, selection, Find and scroll',()=>{
 const input=structuredClone(data(100,100)),trace='a'.repeat(32),span='b'.repeat(16);
 for(const signal of ['logs','spans'])for(const item of input.lanes[signal].rows){item.record.TraceId=trace;item.record.SpanId=span}
 const view=mount(input),search=find(view);search.value='message';search.oninput();search.value='';search.oninput();controls(view)[1].onclick();const chosen=rows(view)[20].dataset.chronologyKey;view.selectedKey=chosen;view.querySelector('goo-table').scrollTop=37;
 const entry=rows(view)[20];assert.equal(view.showRelated(entry.dataset.signal,entry.dataset.sourceKey),true);assert.equal(view.dataset.related,'trace');assert.match(view.textContent,/coverage unknown/);assert.equal(rows(view).length,100);assert.equal(view.querySelectorAll('goo-record-context').length,100);
 const spanControl=view.querySelectorAll('input').find(input=>input.type==='checkbox');spanControl.checked=true;spanControl.onchange();assert.equal(view.dataset.related,'span');assert.equal(rows(view).length,100);
 const back=view.querySelector('[data-related-return]');back.onclick();assert.equal(view.dataset.related,'none');assert.equal(find(view).value,'');assert.match(view.textContent,/Rows 101–200/);assert.equal(view.selectedKey,chosen);assert.equal(view.querySelector('goo-table').scrollTop,37);assert.equal(document.activeElement.dataset.chronologyInspect,chosen);
 view.hidden=true;assert.equal(view.showRelated('logs',input.lanes.logs.rows[0].key),false);view.hidden=false;assert.equal(view.showRelated('logs','forged'),false);back.onclick();assert.equal(view.dataset.related,'none');
});

test('related mode temporarily bypasses chronology Find and resets on a new supplied bank',()=>{
 const input=structuredClone(data());input.lanes.logs.rows[0].record.TraceId='a'.repeat(32);input.lanes.spans.rows[0].record.TraceId='a'.repeat(32);const view=mount(input),search=find(view);search.value='message';search.oninput();const key=input.lanes.logs.rows[0].key;view.selectedKey='logs:'+key;
 assert.equal(view.showRelated('logs',key),true);assert.equal(rows(view).length,2);const span=view.querySelectorAll('input').find(input=>input.type==='checkbox');assert.equal(span.disabled,true);view.querySelector('[data-related-return]').onclick();assert.equal(find(view).value,'message');assert.equal(rows(view).length,2);view.showRelated('logs',key);view.data=data();assert.equal(view.dataset.related,'none');assert.equal(rows(view).length,4);
});
