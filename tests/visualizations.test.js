const {test}=require('node:test');
const assert=require('node:assert/strict');
const {layout}=require('../ui/visualizations.mjs');
const span=(id,start,duration,parent='',trace='a')=>({SpanId:id,TraceId:trace,ParentSpanId:parent,timestamp_ns:String(start),duration_ns:String(duration)});
test('exact ns subtraction retains submillisecond offsets at modern epochs',()=>{
 const epoch=1700000000000000000n;
 const result=layout([span('child',epoch+1n,1n,'root'),span('root',epoch,4n)]);
 assert.deepEqual(result.spans.map(s=>s.row.SpanId),['root','child']);
 assert.equal(result.spans[1].left,25);assert.equal(result.spans[1].width,25);
 assert.equal(result.spans[1].depth,1);
});
test('trace identity scopes parents; missing parents and cycles terminate visibly',()=>{
 const result=layout([span('root',0,10),span('other',1,1,'root','b'),span('c',2,1,'d'),span('d',3,1,'c')]);
 assert.equal(result.spans.length,4);
 assert.equal(result.spans.find(s=>s.row.SpanId==='other').missingParent,true);
 assert.ok(result.spans.find(s=>s.row.SpanId==='c').cycle);
});
test('empty and zero-duration traces have finite geometry',()=>{
 assert.equal(layout([]).spans.length,0);
 const s=layout([span('instant',0,0)]).spans[0];
 assert.equal(s.left,0);assert.equal(s.width,0);
});
test('waterfall component withholds malformed or unbounded data without coercing exact times',()=>{
 const {validData}=require('../ui/waterfall.mjs');const row=span('s','18446744073709551615','0');
 assert.equal(validData({rows:[row],truncated:true}),true);
 for(const invalid of [{rows:[{...row,timestamp_ns:'18446744073709551616'}]},{rows:[{...row,duration_ns:'-1'}]},{rows:[{...row,duration_ns:1}]},{rows:[{timestamp_ns:'1',duration_ns:'1'}]},{rows:[row],truncated:'false'},{rows:Array(1001).fill(row)},null])assert.equal(validData(invalid),false);
 assert.equal(row.timestamp_ns,'18446744073709551615');
});
test('duplicate reported IDs retain independently selectable row indices and bounded geometry',()=>{
 const rows=[span('duplicate',10,4),span('duplicate',11,2),span('child',12,0,'duplicate')];const result=layout(rows);
 assert.deepEqual(new Set(result.spans.map(s=>s.index)),new Set([0,1,2]));
 for(const row of result.spans){assert.ok(Number.isFinite(row.left)&&Number.isFinite(row.width));assert.ok(row.left>=0&&row.left+row.width<=100)}
 assert.equal(rows.length,3);
});
