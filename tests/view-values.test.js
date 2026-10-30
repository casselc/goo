import test from 'node:test';
import assert from 'node:assert/strict';
import {formatCell,supportText,exactText} from '../ui/view-values.mjs';

test('rational display preserves exact large integers and marks representational rounding',()=>{
  const field={type:'rational',unit:'USD'};
  assert.equal(formatCell(field,'9007199254740993'),'9007199254740993 USD');
  assert.equal(formatCell(field,'1/4'),'0.25 USD');
  assert.equal(formatCell(field,'1/3'),'≈ 0.333 USD');
  assert.equal(formatCell(field,'-1/3'),'≈ −0.333 USD');
  assert.equal(formatCell(field,'0'),'0 USD');
  assert.equal(formatCell({type:'number'},1e-7),'<0.001');
});

test('duration display converts exact ns rational to compact human units and retains method',()=>{
  const field={type:'rational',unit:'ns',aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}};
  assert.equal(formatCell(field,'4000000'),'4 ms');
  assert.equal(formatCell(field,'1500'),'1.5 µs');
  assert.equal(formatCell(field,'1000000000'),'1 s');
  assert.equal(formatCell(field,'1/10'),'<0.001 µs');
  assert.match(exactText(field,'4000000',{state:'derived'}),/4000000 ns.*quantilesExactHigh.*derived/);
});

test('no-valid, unknown and no-data are distinct from numeric zero with exact support',()=>{
  const field={type:'rational',unit:'USD'},support={total:'9007199254740993',valid:'0',missing:'9007199254740992',empty:'0',invalid:'1'};
  assert.equal(formatCell(field,null,{state:'no_valid_values',support}),'No valid values');
  assert.equal(formatCell(field,null,{state:'unknown'}),'Unknown');
  assert.equal(formatCell(field,null,{state:'no_data'}),'No data');
  assert.equal(formatCell(field,'0',{state:'derived'}),'0 USD');
  assert.equal(supportText({support}),'0/9007199254740993 valid · missing 9007199254740992 · empty 0 · invalid 1');
});
