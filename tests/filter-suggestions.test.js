const {test}=require('node:test'),assert=require('node:assert/strict'),{suggestionData}=require('../ui/attributes');
const f=(kind,key)=>({kind,key}),field=f('attribute','release');
test('supplied observations preserve qualified empty and missing values without coercing null or absent maps',()=>{
 const records=[{Body:'a',Attributes:{release:''},ResourceAttributes:{release:'resource'}},{Body:'b',Attributes:{},ResourceAttributes:{}},{Body:'c',Attributes:{release:null}},{Body:'d'},{Body:'e',Attributes:{release:'Present empty string'}}];
 const data=suggestionData(records,[field]);assert.deepEqual(data.fields.find(x=>x.field.kind==='attribute'&&x.field.key==='release').values,['',null,'Present empty string']);assert.deepEqual(data.fields.find(x=>x.field.kind==='resource'&&x.field.key==='release').values,['resource',null]);assert.equal(data.rows,5);assert.equal(data.limited,false);assert.equal(records[0].Attributes.release,'');
});
test('unknown map coverage never invents missing observations and bounded suggestions do not define totals',()=>{
 const noMap=suggestionData([{Body:'a'}],[field]);assert.deepEqual(noMap.fields.find(x=>x.field.key==='release').values,[]);
 const data=suggestionData(Array.from({length:101},(_,i)=>({Body:'row '+i,Attributes:{release:String(i)}})));assert.equal(data.rows,100);assert.equal(data.limited,true);assert.equal(data.fields.find(x=>x.field.key==='release').values.length,12);
 const fields=suggestionData([{Attributes:Object.fromEntries(Array.from({length:70},(_,i)=>['key'+i,'v']))}]);assert.equal(fields.fields.length,64);assert.equal(fields.limited,true);
 const detached=suggestionData([{Attributes:{release:'canary'}}]);detached.fields[0].field.key='mutated';assert.equal(suggestionData([{Attributes:{release:'canary'}}]).fields[0].field.key,'release');
});

test('empty and missing choices stay explicit when ordinary literal suggestions reach their budget',()=>{const rows=Array.from({length:14},(_,i)=>({Attributes:{release:String(i)}}));rows.push({Attributes:{release:''}},{Attributes:{}});const data=suggestionData(rows),values=data.fields.find(x=>x.field.key==='release').values;assert.equal(values.length,12);assert.equal(data.limited,true);assert.ok(values.includes(''));assert.ok(values.includes(null))});
