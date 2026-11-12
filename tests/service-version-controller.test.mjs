import {test} from 'node:test';
import assert from 'node:assert/strict';
import {fixture,observed,tick,walk} from './support/dashboard-owner-354.mjs';

const scope={environment:'prod',namespace:{mode:'missing'},fromMs:0,toMs:60000};
function setup(){const stages=[],f=fixture(1,{onCompare:(...args)=>stages.push(args)},scope);f.plan.widgets[0].key='spans';f.plan.widgets[0].query={signal:'spans',environment:'prod',fromMs:0,toMs:60000,limit:100,offset:0,groupBy:[{kind:'column',key:'name'}],where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'missing'}]}};return {f,stages};}
const byText=(host,text)=>walk(host).find(x=>x.textContent===text),byLabel=(host,label)=>walk(host).find(x=>x.getAttribute?.('aria-label')===label);
function discovered(request){const response=observed(request);response.data.groups=[{g0:null,count:'1'},{g0:'',count:'1'},{g0:'a',count:'3'},{g0:'b',count:'4'}];response.data.count='9';return response;}

test('415 real dashboard discovery is explicit, source-free staging preserves choices and retires hidden handlers',async()=>{
 const {f,stages}=setup();await tick();assert.equal(f.requests.length,1);f.requests[0].resolve(observed(f.requests[0]));await f.loaded;
 const root=walk(f.host).find(x=>x.className==='service-version-comparison');assert(root);assert.equal(f.requests.length,1);root.open=true;
 const discover=byText(root,'Discover versions'),stage=byText(root,'Stage version comparison');assert(stage.disabled);const pending=discover.onclick();assert.equal(f.requests.length,2);assert.equal(f.pending.at(-1),true);discover.onclick();assert.equal(f.requests.length,2);f.requests[1].resolve(discovered(f.requests[1]));await pending;
 const baseline=byLabel(root,'Baseline version'),selection=byLabel(root,'Comparison version'),measure=byLabel(root,'Version comparison measure');assert.equal(stage.disabled,true);assert.equal(baseline.children.length,5);assert.equal(selection.children.length,5);const receipt=byText(root,'Read details');assert(receipt);assert.equal(receipt.hidden,false);const receiptText=walk(receipt).map(x=>x.textContent).join(' ');assert(receiptText.includes('Count basis'));assert(receiptText.includes('Groups basis'));assert(receiptText.includes('2026-11-06T00:00:00Z'));assert(receiptText.includes('resource service.version'));assert(walk(root).some(x=>x.textContent.includes('1 unreported · 1 empty')));baseline.value='3';selection.value='4';measure.value='count';selection.onchange();assert.equal(stage.disabled,false);const before=f.requests.length;stage.onclick();assert.equal(stages.length,1);assert.equal(f.requests.length,before);assert.equal(stages[0][0].queries[0].where.all.at(-1).value,'b');
 for(const mode of ['disabled','hidden','inert']){stage[mode]=true;stage.onclick();assert.equal(stages.length,1);stage[mode]=false;}
 root.open=false;stage.onclick();assert.equal(stages.length,1);root.open=true;const old=stage.onclick;root.remove();f.host.append(root);f.retire();stage.disabled=false;old();assert.equal(stages.length,1);
});

test('415 late closed and failed discoveries do not retain options or stage old populations',async()=>{
 const {f,stages}=setup();await tick();f.requests[0].resolve(observed(f.requests[0]));await f.loaded;const root=walk(f.host).find(x=>x.className==='service-version-comparison');root.open=true;const discover=byText(root,'Discover versions'),stage=byText(root,'Stage version comparison');const request=discover.onclick();assert.equal(f.requests.length,2);root.open=false;f.requests[1].resolve(discovered(f.requests[1]));await request;assert.equal(stage.disabled,true);stage.onclick();assert.equal(stages.length,0);
 root.open=true;const next=discover.onclick();assert.equal(f.requests.length,3);f.requests[2].resolve({ok:false,data:{error:'Source unavailable'}});await next;assert.equal(root.dataset.state,'unavailable');assert.equal(byText(root,'Read details').hidden,true);assert.equal(stage.disabled,true);assert.equal(stages.length,0);f.retire();
});
