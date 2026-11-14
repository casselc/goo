import {test,mock} from 'node:test';
import assert from 'node:assert/strict';
class Element{
 constructor(tag,text=''){this.tag=tag;this.raw=text;this.children=[];this.parentElement=null;this.attributes={};this.dataset={};this.listeners=new Map();this.hidden=false;this.disabled=false;this.value=''}
 get isConnected(){return this.root||!!this.parentElement?.isConnected}
 get textContent(){return this.raw+this.children.map(n=>n.textContent).join('')}
 set textContent(value){this.raw=value;this.replaceChildren()}
 append(...nodes){for(const n of nodes){n.remove();n.parentElement=this;this.children.push(n)}}
 replaceChildren(...nodes){for(const n of this.children)n.parentElement=null;this.children=[];this.append(...nodes)}
 remove(){if(this.parentElement)this.parentElement.children=this.parentElement.children.filter(n=>n!==this);this.parentElement=null}
 setAttribute(key,value){this.attributes[key]=String(value)}getAttribute(key){return this.attributes[key]??null}
 addEventListener(type,callback){this.listeners.set(type,[...(this.listeners.get(type)||[]),callback])}
 dispatch(type,detail={}){for(const callback of this.listeners.get(type)||[])callback({type,detail,...detail})}
 closest(selector){for(let p=this;p;p=p.parentElement)if(selector==='[inert]'&&p.inert)return p;return null}
 focus(){document.activeElement=this}
}
const walk=root=>root.children.flatMap(n=>[n,...walk(n)]),node=(tag,text='',cls='')=>{const e=new Element(tag,text);e.className=cls;return e},visible=e=>{if(!e?.isConnected||document.hidden)return false;for(let p=e;p;p=p.parentElement)if(p.hidden||p.inert)return false;return true};
let reads=[];
mock.module('../ui/components.mjs',{exports:{node}});
mock.module('../ui/time-bars.mjs',{exports:{}});
mock.module('../ui/dom-visibility.mjs',{exports:{actionVisible:visible,nativeControlReady:e=>visible(e)&&!e.disabled}});
mock.module('../ui/queries.mjs',{exports:{postJSON:(url,q,signal)=>new Promise((resolve,reject)=>reads.push({url,q:structuredClone(q),signal,resolve,reject}))}});
const {mount}=await import('../ui/activity-log-context.mjs');
const turn=()=>new Promise(resolve=>setImmediate(resolve));
const summary=(count,band='info')=>Object.fromEntries(['count','trace','debug','info','warn','error','fatal','unspecified','invalid'].map(key=>[key,key==='count'||key===band?String(count):'0']));
function dto(q,{empty=false}={}){const values=empty?[0,0,0,0,0,0]:[2,0,0,0,0,1],count=empty?'0':'3',receipt={basis:'a'.repeat(64),readAt:'2026-11-12T12:00:00Z'};return{version:'goo/record-aggregates/v1',question:structuredClone(q),capturedAt:receipt.readAt,basis:'b'.repeat(64),state:empty?'no_data':'observed',count,logSeverity:empty?summary('0'):{...summary('3'),info:'2',error:'1'},groups:empty?[]:[{g0:'INFO',count:'2',logSeverity:summary('2')},{g0:' ERROR ',count:'1',logSeverity:summary('1','error')}],warnings:[],reads:{count:receipt,groups:receipt,timeSeries:receipt},timeSeries:values.map((count,i)=>({fromMs:q.fromMs+i*10000,toMs:q.fromMs+(i+1)*10000,count:String(count),state:count?'observed':'no_data',logSeverity:summary(count,i===5?'error':'info')}))}}

function setup(t){
 reads=[];globalThis.document={hidden:false,activeElement:null};const root=node('main');root.root=true;const toolbar=node('div'),host=node('div');root.append(toolbar,host);
 const c={selector:{signal:'logs',environment:'prod',where:{any:[{field:{kind:'attribute',key:'tenant'},op:'missing'},{field:{kind:'attribute',key:'tenant'},op:'eq',value:''}]}},readWindow:{fromMs:120000,toMs:180000},active:true,clean:true,phase:'partial'},explanations=[],paused=[],staged=[];let owner;
 owner=mount(host,{controlsHost:toolbar,current:()=>c,pauseStream:()=>{paused.push(true);c.phase='paused';owner.sync()},onExplain:(...args)=>explanations.push(args),onStage:(...args)=>staged.push(args)});t.after(()=>owner.dispose());
 const all=()=>[...walk(toolbar),...walk(host)],button=name=>all().find(e=>e.tag==='button'&&(e.textContent===name||e.attributes['aria-label']===name));
 return{root,toolbar,host,c,owner,paused,explanations,staged,button,band:()=>all().find(e=>e.attributes['aria-label']==='Numeric severity'),chart:()=>all().find(e=>e.tag==='goo-time-bars'),cadence:()=>all().find(e=>e.tag==='select'),section:()=>host.children[0],open:()=>button('Show volume').onclick()};
}
function complete(read,options){read.resolve({ok:true,data:dto(read.q,options)})}

test('volume is opt-in, reads aggregate-only, and keeps exact severity text distinct in a supplied chart',async t=>{
 const f=setup(t);assert.equal(reads.length,0);const work=f.open();assert.equal(reads.length,1);assert.equal(reads[0].url,'/api/record-aggregates');assert.deepEqual(reads[0].q.where,f.c.selector.where);assert.equal(reads[0].q.bucketMs,10000);complete(reads[0]);await work;
 assert.deepEqual(f.chart().data.points.map(p=>p.value),['2','0','0','0','0','1']);assert.match(f.section().textContent,/" ERROR "/);assert.equal(walk(f.host).find(e=>e.dataset.severity===' ERROR ').children[0].textContent,'" ERROR "');assert.equal(f.section().dataset.phase,'ready');assert.equal(reads.length,1);
});
test('interaction during Refresh cancels the pending read and preserves the last admitted snapshot and its old window',async t=>{
 const f=setup(t),first=f.open();complete(reads[0]);await first;const original=f.chart().data;
 f.c.readWindow={fromMs:130000,toMs:190000};f.owner.sync();const refresh=f.button('Refresh log volume').onclick();assert.equal(reads.length,2);assert.equal(reads[1].q.fromMs,130000);assert.deepEqual(f.chart().data,original);
 f.chart().dispatch('pointerdown');assert.equal(reads[1].signal.aborted,true);f.chart().dispatch('goo-time-range',{fromKey:'4',toKey:'5'});f.button('Explain interval').onclick();
 assert.equal(f.explanations.length,1);assert.equal(f.explanations[0][0].parent.fromMs,120000);assert.equal(f.explanations[0][0].parent.toMs,180000);assert.equal(f.explanations[0][0].selection.fromMs,160000);assert.equal(f.explanations[0][0].selection.toMs,180000);assert.deepEqual(f.explanations[0][0].selection.where,f.c.selector.where);
 complete(reads[1]);await refresh;assert.deepEqual(f.chart().data,original);assert.equal(f.section().dataset.fromMs,'120000');assert.equal(reads.length,2);
});
test('cadence schedules without reading, never overlaps, and stops when stream pauses',async t=>{
 const originalSet=globalThis.setTimeout,originalClear=globalThis.clearTimeout,timers=new Map();let next=0;globalThis.setTimeout=(fn,ms)=>{const id=++next;timers.set(id,{fn,ms});return id};globalThis.clearTimeout=id=>timers.delete(id);t.after(()=>{globalThis.setTimeout=originalSet;globalThis.clearTimeout=originalClear});
 const f=setup(t),work=f.open();f.cadence().value='10000';f.cadence().onchange();assert.equal(reads.length,1);assert.equal(timers.size,0);complete(reads[0]);await work;assert.equal(timers.size,1);
 const replaced=[...timers.values()][0];f.cadence().value='30000';f.cadence().onchange();replaced.fn();assert.equal(reads.length,1,'retired cadence callback cannot start a read');const tick=[...timers.values()][0];timers.clear();tick.fn();assert.equal(reads.length,2);tick.fn();assert.equal(reads.length,2);f.c.phase='paused';f.owner.sync();assert.equal(reads[1].signal.aborted,true);assert.equal(timers.size,0);complete(reads[1]);await turn();assert.equal(f.section().dataset.fromMs,'120000');assert.equal(f.section().dataset.auto,'false');f.c.phase='partial';f.owner.sync();tick.fn();assert.equal(reads.length,2,'retired tick cannot resume source reads');
});
test('scope, dirty, hidden, departed and disabled controls retire source work and old callbacks',async t=>{
 const f=setup(t),oldToggle=f.button('Show volume').onclick;f.button('Show volume').disabled=true;oldToggle();assert.equal(reads.length,0);f.owner.sync();const work=f.open(),oldRefresh=f.button('Refresh log volume').onclick;
 f.c.clean=false;f.owner.sync();assert(reads[0].signal.aborted);complete(reads[0]);await work;assert.equal(f.section().dataset.fromMs,undefined);
 f.c.clean=true;f.c.selector.environment='other';f.owner.sync();f.owner.invalidate();oldRefresh();oldToggle();assert.equal(reads.length,1);
 const next=f.button('Refresh log volume').onclick();assert.equal(reads.length,2);document.hidden=true;f.c.active=false;f.owner.sync();assert(reads[1].signal.aborted);complete(reads[1]);await next;document.hidden=false;f.c.active=true;f.owner.sync();assert.equal(f.section().dataset.fromMs,undefined);
 f.root.inert=true;f.button('Refresh log volume').onclick();assert.equal(reads.length,2);f.root.inert=false;
});
test('failure and changed aggregate states never become zero or selectable fallback summaries',async t=>{
 const f=setup(t),work=f.open();const bad=dto(reads[0].q);bad.state='changed_between_reads';reads[0].resolve({ok:true,data:bad});await work;assert.match(f.section().textContent,/Volume unavailable/);assert.equal(f.chart().data,null);assert.equal(f.button('Explain window').disabled,true);
 const next=f.button('Refresh log volume').onclick();reads[1].resolve({ok:false,data:{error:'upstream refused'}});await next;assert.match(f.section().textContent,/upstream refused/);assert.equal(f.chart().data,null);assert.equal(f.section().dataset.auto,'false');
});
test('a completed empty observation retains zero buckets without pretending severity or healthy traffic',async t=>{
 const f=setup(t),work=f.open();complete(reads[0],{empty:true});await work;assert.match(f.section().textContent,/No captured logs/);assert.deepEqual(f.chart().data.points.map(p=>p.value),['0','0','0','0','0','0']);assert.match(f.section().textContent,/Severity text · 0 values/);f.chart().dispatch('goo-time-select',{key:'2'});assert.equal(f.chart().selectedKey,'2');f.button('Explain interval').onclick();assert.equal(f.explanations[0][0].selection.fromMs,140000);
});

test('manual paused Refresh survives unrelated sync and a new stream window keeps the old manual snapshot',async t=>{
 const f=setup(t);f.c.phase='paused';f.owner.sync();const work=f.open();f.owner.sync();assert.equal(reads[0].signal.aborted,false);complete(reads[0]);await work;const original=f.chart().data;
 f.c.phase='loading';f.c.readWindow=null;f.owner.sync();assert.deepEqual(f.chart().data,original);f.c.phase='partial';f.c.readWindow={fromMs:130000,toMs:190000};f.owner.sync();assert.deepEqual(f.chart().data,original);
 f.chart().dispatch('goo-time-select',{key:'1'});f.button('Explain interval').onclick();assert.equal(f.explanations[0][0].parent.fromMs,120000);
});
test('closing held volume remains local and usable with dirty scope',async t=>{
 const f=setup(t),work=f.open();complete(reads[0]);await work;f.c.clean=false;f.owner.sync();assert.equal(f.button('Hide volume').disabled,false);f.button('Hide volume').onclick();assert.equal(f.section().hidden,true);assert.equal(reads.length,1);assert.equal(f.button('Show volume').disabled,true);
});

test('manual controls stay compact and Pause restores Refresh focus when it becomes hidden',async t=>{
 const f=setup(t);assert.equal(f.button('Pause log volume').hidden,true);const work=f.open();assert.equal(f.button('Pause log volume').hidden,false);const pause=f.button('Pause log volume');pause.focus();pause.onclick();assert.equal(pause.hidden,true);assert.equal(document.activeElement,f.button('Refresh log volume'));complete(reads[0]);await work;assert.equal(reads.length,1);
 const next=f.button('Refresh log volume').onclick();complete(reads[1]);await next;const explain=f.button('Explain window');assert.equal(explain.textContent,'Explain');assert.equal(explain.attributes['aria-label'],'Explain window');f.chart().dispatch('goo-time-select',{key:'1'});assert.equal(f.button('Explain interval').textContent,'Explain');assert.equal(f.button('Explain interval').title,'Explain interval');const info=walk(f.host).find(e=>e.title?.startsWith('Window ['));assert.equal(info.textContent.includes('UTC'),false);assert.match(info.title,/source read/);
});


test('single-bar selection clears a previously committed range before setting its selected key',async t=>{
 const f=setup(t),work=f.open();complete(reads[0]);await work;const calls=[];f.chart().clearRange=()=>calls.push({action:'clear',selectedKey:f.chart().selectedKey});
 f.chart().dispatch('goo-time-range',{fromKey:'1',toKey:'3'});assert.equal(f.chart().selectedKey,null);assert.equal(calls.length,0);
 f.chart().dispatch('goo-time-select',{key:'4'});assert.deepEqual(calls,[{action:'clear',selectedKey:null}]);assert.equal(f.chart().selectedKey,'4');f.button('Explain interval').onclick();assert.equal(f.explanations[0][0].selection.fromMs,160000);assert.equal(f.explanations[0][0].selection.toMs,170000);
});


test('numeric band choice replots locally, keeps the total scale, and filters or explains exact numeric scope',async t=>{
 const f=setup(t),work=f.open();assert.equal(reads[0].q.logSeverity,true);complete(reads[0]);await work;const ceiling=f.chart().data.ceiling,band=f.band();assert.equal(band.children.length,9);assert.match(band.textContent,/Unspecified 0 · 0/);assert.match(band.textContent,/Invalid >24 · 0/);
 band.value='error';band.onchange();assert.equal(reads.length,1);assert.equal(f.chart().data.ceiling,ceiling);assert.deepEqual(f.chart().data.points.map(p=>p.value),['0','0','0','0','0','1']);assert.equal(f.c.phase,'paused');assert.equal(f.section().dataset.severityBand,'error');assert.match(f.section().textContent,/1 of 3 logs/);
 f.chart().dispatch('goo-time-select',{key:'5'});f.button('Explain interval').onclick();const exact={all:[f.c.selector.where,{all:[{field:{kind:'column',key:'severity_number'},op:'gte',value:'17'},{field:{kind:'column',key:'severity_number'},op:'lte',value:'20'}]}]};assert.deepEqual(f.explanations[0][0].parent.where,exact);assert.deepEqual(f.explanations[0][0].selection.where,exact);assert.equal(f.explanations[0][0].selection.fromMs,170000);
 f.button('Filter Error').onclick();assert.deepEqual(f.staged[0][0],exact);assert.equal(reads.length,1);
});
test('numeric filter leases retire band ABA, changed scope, disabled/hidden control and pause reentrancy',async t=>{
 const f=setup(t),work=f.open();complete(reads[0]);await work;const band=f.band();band.value='error';band.onchange();const button=f.button('Filter Error'),old=button.onclick;band.value='info';band.onchange();band.value='error';band.onchange();old();assert.equal(f.staged.length,0);
 button.disabled=true;button.onclick();assert.equal(f.staged.length,0);button.disabled=false;button.hidden=true;button.onclick();assert.equal(f.staged.length,0);button.hidden=false;f.c.clean=false;button.onclick();assert.equal(f.staged.length,0);f.c.clean=true;
 const original=f.owner.sync;f.owner.sync=()=>{f.c.clean=false;original()};button.onclick();assert.equal(f.staged.length,0);assert.equal(reads.length,1);
});

test('local severity display survives a parent predicate budget that cannot admit another numeric constraint',async t=>{
 const f=setup(t);f.c.selector.where={all:Array.from({length:29},(_,i)=>({field:{kind:'attribute',key:'tag'+i},op:'eq',value:'yes'}))};f.owner.sync();const work=f.open();complete(reads[0]);await work;const band=f.band();band.value='error';assert.doesNotThrow(()=>band.onchange());assert.equal(f.section().dataset.severityBand,'error');assert.equal(f.chart().data.points[5].value,'1');assert.equal(f.button('Filter Error').disabled,true);assert.equal(f.button('Explain window').disabled,true);assert.doesNotThrow(()=>f.button('Explain window').onclick());assert.equal(f.explanations.length,0);assert.equal(f.staged.length,0);assert.equal(reads.length,1);
});
