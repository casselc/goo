const {test} = require('node:test');
const assert = require('node:assert/strict');
const {admitBombadil} = require('../journeys/support/bombadil-verdict');
const run = {code:0, signal:null};
const snapshot = {name:'state',version:'test.model'};
const row = (phase = 'initial') => ({timestamp:1, action:'Wait', state:{}, snapshots:[
  {name:'default',value:false}, {name:'state',value:{version:'test.model',phase}}
], violations:[]});
const encode = rows => rows.map(value => JSON.stringify(value)).join('\n');
const admit = (text, result=run) => admitBombadil(result,text,{snapshot});

test('admission retains observations without claiming eventual completion', () => {
  const result = admit(encode([row()])+'\n');
  assert.equal(result.states[0].phase,'initial');
  assert.equal(result.records.length,1);
  assert.equal(Object.hasOwn(result,'completed'),false);
});
test('unsuccessful process cannot be admitted even with completed observations', () => {
  for (const result of [{code:1},{code:null},{code:0,signal:'SIGINT'},{code:0,error:'spawn failed'}])
    assert.throws(()=>admit(encode([row('complete')]),result),/process/);
});
test('empty, torn, malformed and invalid trace records fail closed', () => {
  for (const text of ['', '  ', '\n', '{', '{}', 'null', '[]', encode([row()])+'\n\n',
    encode([row()])+'\n{', JSON.stringify({...row(),snapshots:{}}), JSON.stringify({...row(),violations:null}),
    JSON.stringify({...row(),snapshots:[null]}), JSON.stringify({...row(),snapshots:[{name:'state'}]})])
    assert.throws(()=>admit(text));
});
test('a zero exit with a recorded violation is rejected anywhere in the trace', () => {
  for(let position=0;position<12;position++) {
    const rows=Array.from({length:12},()=>row('complete'));
    rows[position].violations=[{property:'custodyOnly',violation:{False:{}}}];
    assert.throws(()=>admit(encode(rows)),/violations/);
  }
});
test('required observation cannot be replaced by another extractor or model version', () => {
  for(const snapshots of [[],[{name:'default',value:true}],[{name:'state',value:{version:'other'}}],
    [{name:'other',value:{version:'test.model'}}],[{name:'state',value:null}]])
    assert.throws(()=>admit(JSON.stringify({...row(),snapshots})),/required/);
});
test('invalid required snapshot contracts are rejected', () => {
  for (const snapshot of [null,[],{},'state',{name:''},{name:' '},{name:1},{name:'state',version:2},
    {name:'state',version:''},{name:'state',version:' '}])
    assert.throws(()=>admitBombadil(run,encode([row()]),{snapshot}),/named Bombadil snapshot/);
});
test('every truncated JSON suffix and inserted malformed row rejects an otherwise valid campaign', () => {
  const complete=encode([row('complete')]);
  for(let offset=1;offset<complete.length;offset++) assert.throws(()=>admit(complete.slice(0,offset)));
  for(let position=0;position<=8;position++) {
    const rows=Array.from({length:8},()=>encode([row()]));
    rows.splice(position,0,'{"snapshots":');
    assert.throws(()=>admit(rows.join('\n')));
  }
});
