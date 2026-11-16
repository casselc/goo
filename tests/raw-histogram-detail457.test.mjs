import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {Element} from './support/raw-observation-dom.mjs';
import {detail} from '../ui/raw-histogram-observations.mjs';
const fixtures=JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url))).cases;
function mount(name){const bank=JSON.parse(fixtures.find(value=>value.name===name).response),host=new Element('main');host.root=true;globalThis.document={body:host,hidden:false,activeElement:null,createElement:tag=>new Element(tag)};detail(host,bank.rows[0],{bank});return {host,row:bank.rows[0]}}
test('source statistics share one presence qualification and a compact native two-column table',()=>{
 const {host,row}=mount('observed'),stats=host.querySelectorAll('goo-table').find(element=>element.getAttribute('aria-label')==='Source histogram statistics'),table=stats.querySelector('table');assert.equal(stats.getAttribute('mobile-layout'),null);assert.equal(table.querySelector('thead').querySelectorAll('th').length,2);assert.equal(table.querySelector('tbody').querySelectorAll('tr').length,3);assert.equal(stats.textContent.match(/Not preserved by exporter/g).length,1);assert.match(stats.querySelector('caption').textContent,/presence unavailable, including source zero/);assert.equal(table.querySelector('tbody').querySelectorAll('td')[0].textContent,String(row.sum.rawValue));
});
test('source statistic text retains exact raw values and suppressed/nonfinite/absent states',()=>{
 for(const name of ['observed','no-recorded','nonfinite-stat','missing-stat-columns']){
  const {host,row}=mount(name),stats=host.querySelectorAll('goo-table').find(element=>element.getAttribute('aria-label')==='Source histogram statistics'),cells=stats.querySelector('tbody').querySelectorAll('td');
  for(const [index,key]of ['sum','min','max'].entries()){
   const statistic=row[key],raw=statistic.rawValue;assert.equal(cells[index].dataset.valueState,statistic.valueState);if(raw!==null&&raw!==undefined)assert.ok(cells[index].textContent.includes(String(raw)),name+' '+key);else assert.match(cells[index].textContent,/Not reported/);if(statistic.valueState!=='finite')assert.ok(cells[index].textContent.includes(row.recordingState==='no-recorded-value'?'Not recorded':statistic.valueState),name+' '+key);
  }
 }
});
