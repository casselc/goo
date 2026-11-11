import test from 'node:test';
import assert from 'node:assert/strict';
import {formatCell,exactText,isFraction,rateContextText,defaultSummaryFields} from '../ui/view-values.mjs';
import {timeLegend,timeState} from '../ui/scope-view.mjs';
const field={key:'record.rate',label:'Captured log records per second',type:'rational',unit:'records/s',aggregation:{family:'rate',kind:'records-per-second',method:'exact-rational'}};

test('captured record rates remain unbounded exact rates rather than percentages',()=>{
 assert.equal(isFraction(field),false);
 assert.equal(formatCell(field,'125',{state:'derived'}),'125 records/s');
 assert.equal(formatCell(field,'1/3',{state:'derived'}),'≈ 0.333 records/s');
 assert.equal(formatCell(field,'9007199254740993',{state:'derived'}),'9007199254740993 records/s');
 const text=exactText(field,'2',{state:'derived',rate:{count:'1',elapsedMs:'500'}});
 assert.match(text,/1 captured record over 500 ms/);
 assert.match(text,/count × 1000 \/ elapsed ms/);
 assert.doesNotMatch(text,/bounded fraction|%/);
 assert.equal(rateContextText(undefined),'');
});

test('a known empty interval has a zero rate while an unknown interval stays unknown',()=>{
 assert.equal(formatCell(field,'0',{state:'derived',rate:{count:'0',elapsedMs:'500'}}),'0 records/s');
 assert.equal(formatCell(field,null,{state:'unknown'}),'Unknown');
 assert.equal(timeState({state:'no_data',cells:{'record.rate':{state:'derived'}}},'record.rate',{coordinate:0}),'observed');
 assert.equal(timeState({state:'unknown',cells:{'record.rate':{state:'unknown'}}},'record.rate',{coordinate:null}),'unknown');
 assert.equal(timeLegend({...field,label:'Captured log records/s'},'record.rate')[0][0],'Captured log records/s');
 assert.deepEqual(timeLegend(field,'record.rate').map(([text])=>text),['Captured log records per second · records/s','Known zero','Unknown / unreported']);
});

test('supplied rate fields never add automatic summary measures',()=>{
 const count={key:'count',type:'number',unit:'records'},duration={key:'duration.mean',type:'rational',unit:'ns',aggregation:{family:'duration',kind:'mean'}};
 assert.deepEqual(defaultSummaryFields([count,field]),[count]);
 assert.deepEqual(defaultSummaryFields([count,duration,field]),[count,duration]);
 assert.deepEqual(defaultSummaryFields([field]),[]);
});
