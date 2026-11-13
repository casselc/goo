import {object as closed} from './intent-shape.mjs';
/* Pure exact numeric-predicate admission. No DOM or source reads. */
function validateNumericPredicate(p){
  const object=(v,keys)=>{try{closed(v,keys);return true}catch{return false}},bytes=v=>new TextEncoder().encode(v).length;
  if(!object(p,['numeric','op','value'])||!object(p.numeric,['field','type','unit','parse','scale']))throw Error('Use one numeric binding per condition, without an untyped field or group.');
  const b=p.numeric,f=b.field,wellFormed=s=>![...s].some(c=>{const n=c.codePointAt(0);return n>=0xD800&&n<=0xDFFF});
  if(!object(f,['kind','key'])||!['attribute','resource'].includes(f.kind)||typeof f.key!=='string'||!f.key||!wellFormed(f.key)||bytes(f.key)>128||f.key.includes('\0'))throw Error('Choose an attribute or resource numeric field (up to 128 bytes).');
  if(!['int64','uint64','decimal64'].includes(b.type)||b.parse!=='canonical-decimal')throw Error('Declare int64, uint64 or decimal64 with canonical-decimal parsing.');if(b.type==='decimal64'?!Number.isInteger(b.scale)||b.scale<1||b.scale>9:b.scale!==undefined&&b.scale!==0)throw Error('Declare an explicit decimal scale from 1 to 9; integers have no scale.');
  if(typeof b.unit!=='string'||!b.unit||!wellFormed(b.unit)||bytes(b.unit)>64||/[\p{White_Space}\p{Cc}]/u.test(b.unit))throw Error('Declare a unit without spaces or controls (up to 64 bytes); use 1 for dimensionless.');
  if(['valid','missing','empty','invalid'].includes(p.op)){if(p.value!==undefined&&p.value!=='')throw Error('Numeric state checks have no threshold.');return p}
  if(!['eq','ne','gt','gte','lt','lte'].includes(p.op))throw Error('Choose a numeric comparison or parse-state check.');
  const v=p.value;if(b.type==='decimal64'){if(typeof v!=='string'||v.length>32||/\s/u.test(v)||!new RegExp('^-?(0|[1-9][0-9]*)\\.[0-9]{'+b.scale+'}$').test(v)||/^-[0.]+$/.test(v))throw Error('Use exactly '+b.scale+' fractional digits, without negative zero, exponent or whitespace.');const n=BigInt(v.replace('.',''));if(n< -9223372036854775808n||n>9223372036854775807n)throw Error('Threshold is outside decimal64 coefficient range.');return p}if(typeof v!=='string'||v.length>20||/\s/u.test(v)||!/^(0|[1-9][0-9]*|-[1-9][0-9]*)$/.test(v))throw Error('Use an exact canonical decimal integer threshold.');
  const n=BigInt(v),low=b.type==='int64'?-9223372036854775808n:0n,high=b.type==='int64'?9223372036854775807n:18446744073709551615n;
  if(n<low||n>high)throw Error('Threshold is outside the declared '+b.type+' range.');return p;
 }
export {validateNumericPredicate};
