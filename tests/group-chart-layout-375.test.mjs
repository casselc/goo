import {test} from 'node:test';
import assert from 'node:assert/strict';
import {Chart,node,Breakdown,Results} from './support/group-chart-375.mjs';
import {createSourceAdapter} from '../ui/snapshots.mjs';
function raw(){const question={signal:'logs',environment:'prod',fromMs:1000,toMs:4000,limit:100,offset:0,groupBy:[{kind:'attribute',key:'tenant'},{kind:'attribute',key:'type'}]};return{version:'goo.dataset/1',origin:'source-read',id:'groups375',revision:'read375',capturedAt:'2026-11-07T00:00:00.123456789Z',scope:{id:'q375',generation:'read375',signal:'logs',selector:question},shape:'ranked',dimensions:['category','category1'],fields:[{key:'category',label:'attribute:tenant',type:'string',nullable:true},{key:'category1',label:'attribute:type',type:'string',nullable:true},{key:'count',label:'Captured records',type:'number',unit:'records'}],rows:[{key:'missing',values:{category:null,category1:'Present empty string',count:'1'}},{key:'empty',values:{category:'',category1:'long-雪/'.repeat(12),count:'9007199254740993'}}],quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'read375',scopeId:'q375',aspects:['groups']}}}
const dataset=()=>{const value=raw();return createSourceAdapter({id:'chart375'}).adoptSnapshot(value,{question:value.scope.selector})};
function check(chart,expected){const table=chart.querySelector('table');assert(table,chart.textContent);const frame=chart.querySelector('goo-table');assert.equal(frame.getAttribute('mobile-layout'),expected);const headers=table.querySelector('thead').querySelectorAll('th');assert(headers.every(header=>header.scope==='col'));for(const row of table.querySelector('tbody').querySelectorAll('tr')){assert.equal(row.children.length,headers.length);for(const [i,cell]of row.children.entries())assert.equal(cell.dataset.label,headers[i].textContent)}return table;}
test('initial server attributes and detached SPA inputs opt into the same labeled passive table',()=>{
 for(const mode of ['server','spa']){const chart=node('goo-chart'),snapshot=dataset(),display={kind:'table'};assert(chart instanceof Chart);chart.setAttribute('mobile-layout','stacked');
  if(mode==='server'){chart.setAttribute('snapshot',JSON.stringify(snapshot));chart.setAttribute('display',JSON.stringify(display))}else{chart.snapshot=snapshot;chart.display=display}
  const before=JSON.stringify(snapshot);document.body.append(chart);chart.connectedCallback();const table=check(chart,'stacked');assert(table.textContent.includes('9007199254740993'));assert(table.textContent.includes('long-雪/'.repeat(12)));assert.equal(JSON.stringify(snapshot),before);
 }
});
test('ordinary chart defaults and unrecognized layout attributes do not opt unrelated tables into stacking',()=>{
 for(const layout of [null,'cards','Stacked']){const chart=node('goo-chart');if(layout!==null)chart.setAttribute('mobile-layout',layout);chart.snapshot=dataset();chart.display={kind:'table'};document.body.append(chart);check(chart,null)}
 const chart=node('goo-chart');chart.snapshot=dataset();chart.display={kind:'table'};document.body.append(chart);check(chart,null);chart.setAttribute('mobile-layout','stacked');check(chart,'stacked');chart.removeAttribute('mobile-layout');check(chart,null);
});
test('the canonical grouped caller passes the passive option while direct breakdown callers retain their default',()=>{
 const snapshot=dataset(),defaultHost=node('section');document.body.append(defaultHost);const direct=Breakdown.mount(defaultHost,snapshot,{facets:false});assert(direct);check(direct.chart,null);
 const host=node('section');document.body.append(host);const sourceBefore=JSON.stringify(snapshot),owner=Results.groups(host,{question:snapshot.scope.selector,groups:[],viewDataset:snapshot});assert(owner);const chart=host.querySelector('goo-chart');assert(chart);check(chart,'stacked');assert.equal(JSON.stringify(snapshot),sourceBefore);
});
