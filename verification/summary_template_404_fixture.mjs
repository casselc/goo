import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Templates from '../ui/view-template-handoff.mjs';
import * as Composition from '../ui/view-composition.mjs';
const vectors=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Template authoring must not read source data')};
let controls=0;
for(const v of vectors){const before=JSON.stringify(v),source=v.source,q=v.destination;
 for(const mode of ['empty','unreported']){const bad=structuredClone(source);bad.queries[0].environment='';bad.queries[0].environmentSelector={version:'goo/environment-selector/v1',policy:'current-only',scope:{mode}};assert.throws(()=>Templates.templateView(bad));controls++;}
 Templates.templateView(source);const request=Templates.authoredRequest(source,q);assert.deepEqual(request,{source,query:q});const expanded=Templates.authoredResponse(v.expansion,source,q);assert.deepEqual(expanded.queries,[q]);assert.deepEqual(expanded.composition,source.composition);assert.deepEqual(Templates.templateFor('ordinary-summary'),{version:'goo.template/1',group:'none',profile:'ordinary-summary'});
 for(const mutate of [x=>x.queries[0].offset=1,x=>x.queries[0].groupBy=[{kind:'column',key:'service'}],x=>x.queries[0].numeric={field:{kind:'attribute',key:'n'},type:'int64',unit:'{item}',parse:'canonical-decimal'},x=>x.template.group='destination',x=>x.template.baseline='destination',x=>x.template.profile='ordinary',x=>x.composition.children.push({records:Composition.records(x.composition)}),x=>x.composition.children=x.composition.children.filter(c=>!c.summary),x=>x.snapshot={},x=>x.basis='source-basis']){const bad=structuredClone(source);mutate(bad);assert.throws(()=>Templates.templateView(bad));controls++;}
 for(const mutate of [x=>x.offset=1,x=>x.signal=q.signal==='logs'?'spans':'logs',x=>x.groupBy=[{kind:'column',key:'service'}],x=>x.numeric={field:{kind:'attribute',key:'n'},type:'int64',unit:'{item}',parse:'canonical-decimal'}]){const bad=structuredClone(q);mutate(bad);assert.throws(()=>Templates.authoredRequest(source,bad));controls++;}
 const selected=Composition.summary(source.composition).display.fields;if(selected.some(k=>k.startsWith('duration.'))){const bad=structuredClone(q);delete bad.summary;assert.throws(()=>Templates.authoredRequest(source,bad));controls++;}if(selected.some(k=>k.startsWith('status.'))){const bad=structuredClone(q);delete bad.spanStatus;assert.throws(()=>Templates.authoredRequest(source,bad));controls++;}
 for(const mutate of [x=>x.view.queries[0].limit++,x=>x.view.queries[0].where={field:{kind:'column',key:'service'},op:'eq',value:'foreign'},x=>x.view.queries[0].fromMs++,x=>x.view.composition.children.reverse(),x=>x.view.template.profile='ordinary']){const bad=structuredClone(v.expansion);mutate(bad);assert.throws(()=>Templates.authoredResponse(bad,source,q));controls++;}
 assert.equal(JSON.stringify(v),before);
}
console.log(`${vectors.length} Go/JS template carriers; ${controls} destination/lease/profile/role refusal controls`);
