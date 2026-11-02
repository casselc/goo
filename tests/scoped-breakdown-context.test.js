const {test}=require('node:test'),assert=require('node:assert/strict');
const {boundContext}=require('../ui/scoped-breakdown.mjs');
const snapshot=(groupBy=[{kind:'attribute',key:'http.route'}])=>({
 version:'goo.dataset/1',id:'context',revision:'r1',shape:'ranked',dimensions:['category'],
 scope:{id:'q1',generation:1,signal:'spans',selector:{signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,groupBy,summary:'span-duration'}},
 fields:[{key:'category',label:'attribute:http.route',type:'string',nullable:true},{key:'count',label:'Captured records',type:'number',unit:'records'},
  {key:'duration.p95',label:'p95 duration',type:'rational',unit:'ns',nullable:true,aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}}],
 rows:[],quality:{extent:'partial',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'}
});
test('default breakdown titles and qualified context follow actual route or service binding',()=>{
 const route=snapshot(),service=snapshot([{kind:'column',key:'service'}]),before=structuredClone({route,service});
 assert.deepEqual(boundContext(route),{title:'Captured records by HTTP route',text:'attribute:http.route · Captured records'});
 assert.deepEqual(boundContext(service),{title:'Captured records by Service',text:'column:service · Captured records'});
 assert.deepEqual({route,service},before);
});
test('explicit titles, including an old shipped title, stay literal while effective context changes',()=>{
 const route=snapshot(),service=snapshot([{kind:'resource',key:'service.namespace'}]);
 for(const title of ['HTTP server routes','  Our Ω incident  ']){
  assert.equal(boundContext(route,'count',title).title,title);assert.equal(boundContext(service,'count',title).title,title);
  assert.equal(boundContext(service,'count',title).text,'resource:service.namespace · Captured records');
 }
});
test('measure changes update semantic default and context without persisting or mutating data',()=>{
 const data=snapshot(),before=structuredClone(data);
 assert.deepEqual(boundContext(data,'duration.p95'),{title:'p95 duration by HTTP route',text:'attribute:http.route · p95 duration'});
 assert.deepEqual(boundContext(data,'duration.p95','HTTP server routes'),{title:'HTTP server routes',text:'attribute:http.route · p95 duration'});
 assert.deepEqual(data,before);
});
test('temporal and tuple contexts describe their actual axes instead of a former single category',()=>{
 const temporal={...snapshot(),shape:'time-series',dimensions:[]};
 assert.deepEqual(boundContext(temporal),{title:'Captured records over time',text:'UTC intervals · Captured records'});
 const tuple=snapshot([{kind:'column',key:'service'},{kind:'resource',key:'zone'}]);tuple.dimensions=['category','category1'];
 assert.deepEqual(boundContext(tuple),{title:'Captured records by Service + resource zone',text:'column:service + resource:zone · Captured records'});
 const supplied=snapshot();delete supplied.scope.selector.groupBy;
 assert.deepEqual(boundContext(supplied),{title:'Captured records by attribute:http.route',text:'attribute:http.route · Captured records'});
});
