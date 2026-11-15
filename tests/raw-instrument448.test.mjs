import {test} from 'node:test';
import assert from 'node:assert/strict';
import {checkedRequest} from '../ui/raw-observations.mjs';

test('raw instrument keys agree with exact Go admission across all source families',()=>{
 const question={environment:'prod',service:'api',fromMs:1000,toMs:2000};
 for(const family of [{kind:'gauge'},{kind:'counter',temporality:0,monotonic:false},{kind:'histogram',temporality:0}]){
  const instrument={...family,scope:'',version:'',schemaURL:'',metric:'ordinary.metric',unit:''};
  const request={question,instrument},accepted=checkedRequest(request);
  assert.deepEqual(accepted,request);
  accepted.instrument.scope='changed';assert.equal(instrument.scope,'');
  for(const [key,value]of [['Scope','overridden'],['schemaUrl','overridden'],['Unit','overridden'],['cursor','retained']])assert.throws(()=>checkedRequest({question,instrument:{...instrument,[key]:value}}));
  for(const value of [null,false,0])assert.throws(()=>checkedRequest({question,instrument:{...instrument,scope:value}}));
 }
});
