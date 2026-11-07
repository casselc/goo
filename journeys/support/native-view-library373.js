// Real histogram input comes from the existing native reuse fixture. Additional
// saved declarations exercise metadata pagination; they contain no observed bins.
const Native=require('./native-histogram-reuse371');
function fixture(){const f=Native.fixture();return {...f,sourceTitle:'Reusable native buckets '+f.token,wrongUnitTitle:'Milliseconds display '+f.token}}
function metadataSeeds(f){const unitQuestion=structuredClone(f.source.question);unitQuestion.definition.source.unit='ms';return Array.from({length:26},(_,index)=>({version:1,title:index===0?f.wrongUnitTitle:'Catalog display '+String(index).padStart(2,'0')+' '+f.token,question:index===0?unitQuestion:structuredClone(f.source.question),declaration:Native.declaration(index%2?'histogram':'table')}))}
module.exports={fixture,payload:Native.payload,metadataSeeds};
