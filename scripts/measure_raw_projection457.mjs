/* Synthetic local projection workload; not a browser/source latency benchmark. */
import {readFileSync} from 'node:fs';
import {performance} from 'node:perf_hooks';
import {prepare} from '../ui/raw-observation-dataset.mjs';
import {dataset} from '../ui/raw-histogram-dataset.mjs';
const template=JSON.parse(JSON.parse(readFileSync(new URL('../testdata/http/raw-histogram447.json',import.meta.url))).cases.find(c=>c.name==='observed').response);
const measure=run=>{const start=performance.now();run();return performance.now()-start};
const results=[];
for(const size of [1,100,1000,5000]){
 const data=structuredClone(template);data.rows=Array.from({length:size},(_,i)=>({...structuredClone(template.rows[0]),occurrenceKey:String(i+1)}));
 let bank;const prepareMS=measure(()=>bank=prepare(data)),key=String(size);
 // Warm both paths before taking five individual measurements. No time threshold
 // makes correctness depend on host scheduling or the JavaScript JIT.
 dataset(data,key);bank.histogram(key);
 const convenienceMS=[],preparedMS=[];for(let i=0;i<5;i++){convenienceMS.push(measure(()=>dataset(data,key)));preparedMS.push(measure(()=>bank.histogram(key)))}
 results.push({rows:size,bytes:Buffer.byteLength(JSON.stringify(data)),prepareMS,convenienceMS,preparedMS});
}
console.log(JSON.stringify({version:'goo.rawProjection457Measurement/1',node:process.version,scope:'synthetic Node projection, not browser rendering or source latency',results},null,2));
