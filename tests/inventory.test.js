const {test}=require('node:test'),assert=require('node:assert/strict');const {question}=require('../ui/inventory');
test('inventory navigation retains fixed scope and exact literal identities without claiming a shared evidence read',()=>{const source={question:{environment:'production',fromMs:1000,toMs:2000},basis:'capture'};const row={service:"api'\\\n",name:'<work>'};const q=question(source,row);assert.deepEqual(q,{signal:'spans',environment:'production',fromMs:1000,toMs:2000,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:row.service},{field:{kind:'column',key:'name'},op:'eq',value:row.name}]},groupBy:[],limit:100,offset:0});assert.equal(q.basis,undefined);assert.equal(source.question.signal,undefined);});

test('metric discovery carries the same fixed inventory window without claiming a span population',()=>{const {metricContext}=require('../ui/inventory');const context=metricContext({question:{environment:'prod',fromMs:1234,toMs:5678}}, {service:'worker',name:'job'});assert.deepEqual(context,{service:'worker',environment:'prod',fromMs:1234,toMs:5678});assert.equal(context.where,undefined)});

const {scopeQuestion,suggestions}=require('../ui/service-browser');
test('service suggestions choose observed exact convention aliases and narrow only matching spans',()=>{
 const scope={environment:'prod',fromMs:1000,toMs:2000}, service="worker'\n";
 const fields=[{kind:'attribute',key:'http.route'},{kind:'attribute',key:'db.system'},{kind:'resource',key:'messaging.system'}];
 const choices=suggestions(scope,service,fields);
 assert.deepEqual(choices.map(s=>s.available),[true,true,false]);
 for(const s of choices.filter(s=>s.available)){
  assert.deepEqual(s.question.where.all[0],scopeQuestion(scope,service).where);
  assert.equal(s.question.where.all[1].op,'exists');
  assert.deepEqual(s.question.groupBy,[s.field]);assert.equal(s.question.fromMs,1000);assert.equal(s.question.toMs,2000);
 }
 assert.equal(choices[1].field.key,'db.system');assert.equal(choices[2].question,null);
 assert.equal(fields.length,3);assert.equal(scope.signal,undefined);
 const newer=suggestions(scope,service,[{kind:'attribute',key:'db.system'},{kind:'attribute',key:'db.system.name'}]);
 assert.equal(newer[1].field.key,'db.system.name');
});

test('catalog retains separate exact signal counts, groups services and leaves source evidence detached',()=>{
 const {catalogRows}=require('../ui/inventory');const source=[{service:'worker',signal:'metrics',count:'18446744073709551615',firstMs:10,lastMs:20},{service:'api',signal:'spans',count:'3',firstMs:11,lastMs:19},{service:'worker',signal:'logs',count:'2',firstMs:5,lastMs:22}];const rows=catalogRows(source);
 assert.equal(rows.length,2);assert.equal(rows[1].signals.metrics.count,'18446744073709551615');assert.equal(rows[1].signals.spans,undefined);assert.equal(rows[1].firstMs,5);assert.equal(rows[1].lastMs,22);rows[1].signals.metrics.count='0';assert.equal(source[0].count,'18446744073709551615');assert.throws(()=>catalogRows([source[0],source[0]]),/Duplicate/);
});
