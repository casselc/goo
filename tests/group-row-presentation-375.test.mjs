import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Protocol from '../ui/question-protocol.mjs';
import * as Triage from '../ui/error-triage.mjs';
import {mounted,event} from './support/typed-error-triage-370.mjs';
const clone=structuredClone;
function fixture(typed=false,measures=false){
 const parent={signal:'spans',...(typed?{environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'empty'}}}:{environment:'prod'}),fromMs:1000,toMs:61000,limit:100,offset:0,where:{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''}},question=Triage.triageQuery(parent);
 const labels=[null,'','No unique reported exception type','Present empty string','Missing key','org.example.'+'LongException雪🧪/'.repeat(12)];
 const groups=labels.map((value,index)=>({g0:index===0?null:index===1?'':'service-雪/'+String(index).repeat(30),g1:value,count:index===5?'9007199254740993':String(index+1)}));
 if(measures){for(const group of groups)group.numeric={valid:'0',sum:'0',unit:'items',missing:group.count,empty:'0',invalid:'0'};groups[0].numeric={valid:'1',sum:'0',unit:'items',missing:'0',empty:'0',invalid:'0'};groups[0].duration={count:'1',sumNs:'1000000',p95Ns:'1000000',maxNs:'1000000'}}
 return{question,groups,state:'observed'};
}
function actions(data){return[{label:'Inspect group',href:group=>Protocol.questionLink(Triage.inspectGroup(data.question,group)),current:group=>data.groups.includes(group)},{label:'Compare within service',href:group=>Protocol.questionLink(Triage.compareInService(data.question,group.g0)),current:group=>data.groups.includes(group)}]}
test('shared grouped rows associate every responsive cell with the exact native header for ordinary and typed scopes',()=>{
 for(const typed of [false,true])for(const measures of [false,true]){
  const data=fixture(typed,measures),before=clone(data),view=mounted(data,actions(data)),table=view.host.querySelector('table'),headers=table.querySelector('thead').querySelectorAll('th'),rows=table.querySelector('tbody').querySelectorAll('tr');
  assert.equal(view.host.children[0].getAttribute('mobile-layout'),'stacked');assert.match(table.querySelector('caption').textContent,/separate source read/);assert.equal(rows.length,data.groups.length);
  const expected=['column:service','column:exception_type','Captured record count',...(measures?['Numeric mean','Numeric coverage','Mean span duration','p95 span duration','Max span duration']:[]),'Actions'];assert.deepEqual(headers.map(header=>header.textContent),expected);
  for(const row of rows){const cells=row.querySelectorAll('td');assert.equal(cells.length,headers.length);for(const [i,cell]of cells.entries()){assert.equal(cell.dataset.label,expected[i]);assert.equal(headers[i].parentElement.parentElement.parentElement,table);assert.equal(headers[i].scope,'col')}}
  if(measures){assert.equal(rows[1].children[3].textContent,'Unavailable');assert.match(rows[1].children[4].textContent,/0 valid.*2 missing/);assert.deepEqual(rows[1].children.slice(5,8).map(cell=>cell.textContent),['Not recorded','Not recorded','Not recorded'])}assert.equal(rows[5].children[2].textContent,'9007199254740993');assert(rows[5].children[1].textContent.includes(data.groups[5].g1));assert.deepEqual(data,before);
 }
});
test('missing, empty and lookalike literals remain distinguishable without changing continuation membership',()=>{
 const data=fixture(true),view=mounted(data,actions(data)),rows=view.host.querySelector('tbody').querySelectorAll('tr');
 assert.equal(rows[0].children[1].textContent,'No unique reported exception type');assert.equal(rows[1].children[1].textContent,'Present empty string');assert.notEqual(rows[2].children[1].textContent,rows[0].children[1].textContent);assert.notEqual(rows[3].children[1].textContent,rows[1].children[1].textContent);
 for(const [i,row]of rows.entries()){const links=row.querySelectorAll('a');assert.deepEqual(links.map(link=>link.textContent),['Inspect group','Compare within service']);const expected=Triage.inspectGroup(data.question,data.groups[i]);assert.deepEqual(Protocol.readQuestionLink(links[0].href).question,expected);assert.equal(links[0].onclick(event()),true)}
});
test('separate native grouped tables retain their own headings and presentation changes do not renew retired actions',()=>{
 const data=fixture(true,true),first=mounted(data,actions(data)),second=mounted(data,actions(data));assert.notEqual(first.host.querySelector('table'),second.host.querySelector('table'));for(const view of [first,second])assert.equal(view.host.querySelector('thead').querySelectorAll('th').length,view.host.querySelector('tbody').querySelector('tr').children.length);
 const anchor=first.anchors()[0],original=anchor.onclick,href=anchor.href,row=anchor.parentElement.parentElement;anchor.focus();first.host.children[0].setAttribute('mobile-layout','stacked');assert.equal(first.anchors()[0],anchor);assert.equal(anchor.parentElement.parentElement,row);assert.equal(anchor.href,href);assert.equal(document.activeElement,anchor);assert.equal(original(event({ctrlKey:true})),true);
 first.owner.suspendActions();first.host.children[0].setAttribute('density','comfortable');assert.equal(original(event()),false);first.owner.resumeActions();assert.equal(original(event()),false);assert.equal(anchor.onclick(event()),true);assert.equal(anchor.href,href);assert.equal(document.activeElement,anchor);
});
