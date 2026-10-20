const {test}=require('node:test'),assert=require('node:assert/strict');
const {model,safeHref}=require('../ui/time-bars');
test('count bars preserve exact supplied domain, zero and missing state',()=>{
 const data={ceiling:'900719925474099300',points:[{key:'zero',value:'0',label:'Zero'},{key:'one',value:'900719925474099300',label:'Exact count'},{key:'missing',state:'unknown',label:'Coverage unknown'},{key:'empty',value:'0',state:'no_data',label:'No captured records'}]},before=JSON.stringify(data),result=model(data);
 assert.equal(result.points[0].height,0);assert.equal(result.points[0].positive,false);assert.equal(result.points[1].height,100);assert.equal(result.points[1].value,'900719925474099300');assert.equal(result.points[2].value,null);assert.equal(result.points[2].state,'unknown');assert.equal(result.points[3].state,'no_data');assert.equal(JSON.stringify(data),before);
 const filtered=model({...data,points:[data.points[0]]});assert.equal(filtered.ceiling,result.ceiling);
});
test('invalid domains and ambiguous cells withhold the snapshot rather than clipping or summing',()=>{
 for(const data of [{ceiling:'1',points:[{key:'x',value:'2',label:'overflow'}]},{ceiling:'0',points:[{key:'x',value:'1',label:'positive'}]},{ceiling:'1',points:[{key:'x',value:'1',state:'no_data',label:'contradiction'}]},{ceiling:'1',points:[{key:'x',value:'1',label:'one'},{key:'x',value:'1',label:'duplicate key'}]},{ceiling:1,points:[]},{ceiling:'1',points:Array(257).fill({key:'x',value:'1',label:'over budget'})}])assert.equal(model(data).state,'unavailable');
 assert.equal(model({ceiling:'0',points:[{key:'x',value:'0',label:'zero'}]}).points[0].height,0);
 for(const href of ['javascript:alert(1)','data:text/html,x','  #question',''])assert.equal(safeHref(href),null);for(const href of ['#step=abc','/question','https://example.test/question'])assert.equal(safeHref(href),href);
});
