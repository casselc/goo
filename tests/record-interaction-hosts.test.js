const {test}=require('node:test'),assert=require('node:assert/strict');
const {indexRecordItems}=require('../ui/record-interactions.mjs');

test('visual parent-first hosts bind by original delivery position, not DOM order or record identity',()=>{
 const parent={index:2,element:{label:'parent'}},child={index:0,element:{label:'child'}},other={index:1,element:{label:'other'}};
 const visual=[parent,child,other],before=[...visual],indexed=indexRecordItems(visual,3);
 assert.deepEqual(indexed,[child,other,parent]);assert(Object.isFrozen(indexed));assert.deepEqual(visual,before);assert.equal(indexed[2],parent);
 const repeatedRecord={SpanId:'1'.repeat(16)},rows=[repeatedRecord,repeatedRecord];
 const occurrences=[{index:1,record:rows[1]},{index:0,record:rows[0]}];assert.deepEqual(indexRecordItems(occurrences,2).map(item=>item.index),[0,1]);
});

test('missing, duplicate and out-of-range occurrence ports refuse before record controls attach',()=>{
 for(const [items,size]of [[[{index:0}],2],[[{index:0},{index:0}],2],[[{index:0},{index:2}],2],[[{index:-1}],1],[[{index:'0'}],1],[[null],1],[[],NaN],[{},0]])assert.throws(()=>indexRecordItems(items,size));
 assert.deepEqual(indexRecordItems([],0),[]);
});
