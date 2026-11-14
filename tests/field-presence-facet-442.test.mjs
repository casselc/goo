import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mount,ready,labelled} from './support/facet-pagination-369.mjs';

test('facet presence actions are local and independent of sampled or loaded missing values',()=>{
 const element=mount({values:ready([{value:'only-reported',count:'4'}])}),events=[];
 element.addEventListener('goo:facet-stage',event=>events.push(event.detail));
 const reported=labelled(element,'Is reported attribute:cohort'),missing=labelled(element,'Is missing attribute:cohort');
 assert(reported);assert(missing);reported.onclick();missing.onclick();assert.equal(events.length,0,'closed native disclosure cannot stage');reported.closest('details:not([open])').open=true;reported.onclick();missing.onclick();
 assert.deepEqual(events,[{value:null,mode:'exclude'},{value:null,mode:'include'}]);
 const before=structuredClone(element.data);element.data={...before,disabled:true};
 reported.onclick();missing.onclick();assert.equal(events.length,2,'retired callbacks cannot stage');
 labelled(element,'Is missing attribute:cohort').onclick();assert.equal(events.length,2,'disabled owner cannot stage');
 element.data=before;element.hidden=true;labelled(element,'Is missing attribute:cohort').onclick();assert.equal(events.length,2,'hidden owner cannot stage');
 assert.deepEqual(element.data.values.items,[{value:'only-reported',count:'4'}]);
});
