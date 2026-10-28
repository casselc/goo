/* Pure record-predicate validation shared by owners and resource contexts. */
import * as QuestionProtocol from './question-protocol.mjs';
import * as LogPatternIdentity from './log-pattern-identity.mjs';
const columns=['service','trace_id','span_id','scope','scope_version','name','kind','status','duration_ns','timestamp_ns','duration_bucket','status_class','body','severity','severity_number','exception_event','exception_type','exception_type_state','body_pattern_v1'];
const bytes=s=>new TextEncoder().encode(s).length;
const uintColumn=key=>['duration_ns','timestamp_ns','severity_number'].includes(key);
const protocol=()=>QuestionProtocol;
function validate(where,signal='logs'){
  if(!['logs','spans'].includes(signal))throw Error('Choose logs or spans.');
  let count=0;
  function visit(p,depth){
   if(++count>32||depth>4)throw Error('Use at most 32 conditions/groups and 4 nested levels.');
   if(!p||typeof p!=='object')throw Error('Choose a condition.');
   if(p.all||p.any){const keys=Object.keys(p),list=p.all||p.any;if(keys.length!==1||!Array.isArray(list)||!list.length)throw Error('Each group needs a condition.');list.forEach(x=>visit(x,depth+1));return}
   if(p.numeric!=null){protocol().validateNumericPredicate(p);return}
   const f=p.field;if(!f||!['column','attribute','resource'].includes(f.kind)||typeof f.key!=='string'||!f.key||bytes(f.key)>128||f.key.includes('\0'))throw Error('Choose a qualified field (up to 128 bytes).');
   if(f.kind==='column'&&(!columns.includes(f.key)||(signal==='logs'&&['name','kind','status','duration_ns','duration_bucket','status_class','exception_event','exception_type','exception_type_state'].includes(f.key))||(signal==='spans'&&['body','severity','severity_number','body_pattern_v1'].includes(f.key))))throw Error('Field '+f.kind+':'+f.key+' is unavailable for '+signal+'.');
   const numeric=f.kind==='column'&&uintColumn(f.key),v=p.value??'';
   if(typeof v!=='string'||bytes(v)>1024||v.includes('\0'))throw Error('Use a value up to 1024 bytes.');
   if(!['eq','ne','contains','gte','lte','exists','missing'].includes(p.op)||numeric&&p.op==='contains'||!numeric&&['gte','lte'].includes(p.op))throw Error('Operator does not match the field type.');
   if(f.kind==='column'&&f.key==='body_pattern_v1'){if(!['eq','ne','exists','missing'].includes(p.op))throw Error('Pattern identities support exact equals/not-equals or presence only.');if(['eq','ne'].includes(p.op))(LogPatternIdentity||(null)).decode(v)}
   if(['exists','missing'].includes(p.op)){if(v!=='')throw Error('Presence checks have no value.');return}
   if(numeric&&(!/^[0-9]+$/.test(v)||BigInt(v)>18446744073709551615n))throw Error('Use an unsigned 64-bit decimal integer.');
  }
  if(where)visit(where,0);return structuredClone(where);
 }
export {validate};
