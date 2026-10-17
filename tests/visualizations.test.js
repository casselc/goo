const {test}=require('node:test');
const assert=require('node:assert/strict');
const {layout}=require('../ui/visualizations.js');
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
