/* Pure native histogram scope checks. A checked shape is not source authority. */
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const own=(value,key)=>Object.prototype.hasOwnProperty.call(value,key);
const bytes=value=>new TextEncoder().encode(value).length;
const text=(value,max)=>typeof value==='string'&&bytes(value)<=max&&!value.includes('\0')&&!/[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/u.test(value);
const keys=(value,allowed)=>record(value)&&Object.keys(value).every(key=>allowed.includes(key));
const fail=message=>{throw TypeError(message)};
const ns=value=>{if(typeof value!=='string'||!/^(0|[1-9]\d*)$/.test(value)||value.length>19||BigInt(value)>=4102444800000000000n)fail('A canonical bounded native timestamp is required');return BigInt(value)};
function mapping(value,max,keyMax,valueMax){if(value==null)return {};if(!record(value)||Object.keys(value).length>max)fail('Native attribute map exceeds its bounds');for(const [key,item] of Object.entries(value))if(!key||!text(key,keyMax)||!text(item,valueMax))fail('Invalid native attribute');return value}
export function canonicalNativeJSON(value){if(Array.isArray(value))return '['+value.map(canonicalNativeJSON).join(',')+']';if(record(value))return '{'+Object.keys(value).sort().filter(key=>value[key]!==undefined).map(key=>JSON.stringify(key)+':'+canonicalNativeJSON(value[key])).join(',')+'}';return JSON.stringify(value)}
export function sameNativeStream(a,b){return canonicalNativeJSON({...a,resourceSchemaURL:a.resourceSchemaURL||'',resource:a.resource||{},attributes:a.attributes||{}})===canonicalNativeJSON({...b,resourceSchemaURL:b.resourceSchemaURL||'',resource:b.resource||{},attributes:b.attributes||{}})}
function checkedStream(stream,definition){
 const source=definition.source;
 if(!keys(stream,['resourceSchemaURL','resource','scope','version','schemaURL','name','unit','attributes']))fail('A bounded exact native stream is required');
 const resource=mapping(stream.resource,64,256,2048),attributes=mapping(stream.attributes,64,256,2048);
 if(!text(stream.resourceSchemaURL??'',1024)||!text(stream.scope,256)||!text(stream.version,128)||!text(stream.schemaURL,1024)||!text(stream.name,256)||!text(stream.unit,64))fail('Invalid native stream identity');
 const wire={...(stream.resourceSchemaURL?{resourceSchemaURL:stream.resourceSchemaURL}:{}),resource:stream.resource??null,scope:stream.scope,version:stream.version,schemaURL:stream.schemaURL,name:stream.name,unit:stream.unit,attributes:stream.attributes??null};
 if(bytes(JSON.stringify(wire).replace(/[<>&\u2028\u2029]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0')))>16384)fail('Native stream exceeds the wire budget');
 if(!source.services.includes(resource['service.name'])||!resource['service.instance.id']||resource['deployment.environment.name']!==definition.environment||stream.scope!==source.scope||stream.version!==source.version||stream.schemaURL!==source.schemaURL||stream.name!==source.metric||stream.unit!==source.unit)fail('Native stream differs from its measurement binding');
 const namespace=source.namespace;if(namespace&&(namespace.mode==='missing'?own(resource,'service.namespace'):!own(resource,'service.namespace')||resource['service.namespace']!==(namespace.value??'')))fail('Native stream namespace differs from its binding');
 for(const [key,value] of Object.entries(source.attributes||{}))if(!own(attributes,key)||attributes[key]!==value)fail('Native stream attributes differ from its binding');
}
export function validateHistogramMeasurementQuestion(question){
 if(!keys(question,['definition','fromMs','toMs','fromNS','toNS']))fail('A checked histogram measurement question is required');
 const d=question.definition,b=d?.source;
 if(!keys(d,['apiVersion','name','title','kind','environment','lookbackMinutes','maxGapMs','grain','source','histogramBucketMs'])||d.apiVersion!=='goo/measurement-package/v1alpha1'||!text(d.name,128)||!/^[A-Za-z][A-Za-z0-9_.-]{0,127}$/.test(d.name)||!d.title||!text(d.title,160)||!d.environment||!text(d.environment,128)||d.kind!=='histogram'||!['per-stream','consumer-group'].includes(d.grain)||!Number.isInteger(d.lookbackMinutes)||d.lookbackMinutes<1||d.lookbackMinutes>120||!Number.isInteger(d.maxGapMs)||d.maxGapMs<1||d.maxGapMs>600000)fail('Invalid native histogram definition');
 if(!keys(b,['services','scope','version','schemaURL','metric','unit','attributes','namespace','exactStream'])||!Array.isArray(b.services)||b.services.length<1||b.services.length>8||new Set(b.services).size!==b.services.length||b.services.some(value=>!value||!text(value,128))||!b.scope||!text(b.scope,256)||!text(b.version,128)||!text(b.schemaURL,1024)||!b.metric||!text(b.metric,256)||!['s','ms'].includes(b.unit))fail('Invalid native histogram source binding');
 mapping(b.attributes,16,128,512);
 if(b.namespace!==undefined&&b.namespace!==null&&(!keys(b.namespace,['mode','value'])||!['missing','value'].includes(b.namespace.mode)||!text(b.namespace.value??'',128)||b.namespace.mode==='missing'&&(b.namespace.value??'')!==''))fail('Invalid measurement namespace');
 const {fromMs,toMs}=question;if(!Number.isSafeInteger(fromMs)||!Number.isSafeInteger(toMs)||fromMs<0||toMs<=fromMs||toMs-fromMs>7200000||toMs>=4102444800000)fail('Invalid native measurement window');
 let from=BigInt(fromMs)*1000000n,to=BigInt(toMs)*1000000n;
 if(question.fromNS!==undefined||question.toNS!==undefined){from=ns(question.fromNS);to=ns(question.toNS);if(to<=from||to-from>7200000000000n||BigInt(fromMs)!==from/1000000n||BigInt(toMs)!==(to+999999n)/1000000n)fail('Native timestamps differ from their enclosing window')}
 if(d.histogramBucketMs!==undefined){const width=d.histogramBucketMs;if(!Number.isInteger(width)||width<1||width>7200000||(to-from-1n)/(BigInt(width)*1000000n)+1n>200n)fail('Invalid native histogram time bucket width')}
 if(b.exactStream!==undefined&&b.exactStream!==null)checkedStream(b.exactStream,d);
 return {fromNS:String(from),toNS:String(to)};
}
export function validateMeasurementIntervalSelector(selector){
 if(!keys(selector,['kind','question','stream','fromNS','toNS'])||Object.keys(selector).length!==5||selector.kind!=='measurement-interval')fail('An explicit native measurement interval selector is required');
 const bounds=validateHistogramMeasurementQuestion(selector.question);checkedStream(selector.stream,selector.question.definition);
 const exact=selector.question.definition.source.exactStream;if(exact&&!sameNativeStream(exact,selector.stream))fail('Selected stream differs from the exact measurement binding');
 const from=ns(selector.fromNS),to=ns(selector.toNS);if(to<=from||from<BigInt(bounds.fromNS)||to>BigInt(bounds.toNS)||to-from>BigInt(selector.question.definition.maxGapMs)*1000000n)fail('Selected native interval is outside its measurement question');
 return selector;
}
// Exact binary64 boundaries remain rationals in the model, never decimal guesses.
export function binary64Rational(value){
 if(typeof value!=='number'||!Number.isFinite(value))fail('Finite binary64 histogram bounds are required');
 if(value===0)return '0';
 const buffer=new ArrayBuffer(8),view=new DataView(buffer);view.setFloat64(0,value,false);const bits=view.getBigUint64(0,false),negative=(bits>>63n)!==0n,exponent=Number((bits>>52n)&2047n),fraction=bits&((1n<<52n)-1n);
 let numerator=exponent===0?fraction:(1n<<52n)|fraction,shift=(exponent===0?-1022:exponent-1023)-52,denominator=1n;
 if(shift>=0)numerator<<=BigInt(shift);else denominator<<=BigInt(-shift);
 while(denominator>1n&&numerator%2n===0n){numerator/=2n;denominator/=2n}
 if(negative)numerator=-numerator;
 return String(numerator)+(denominator===1n?'':'/'+denominator);
}
// Convert only exactly representable binary64 rationals; avoid Infinity/Infinity
// and decimal mantissa estimates for subnormal or maximum finite boundaries.
export function nativeRationalCoordinate(value){
 if(typeof value!=='string'||value.length>1024||!/^(0|[1-9]\d*)(?:\/[1-9]\d*)?$/.test(value))return null;
 const [raw,denominator='1']=value.split('/');let numerator=BigInt(raw),d=BigInt(denominator);
 if(d&(d-1n)||denominator==='1'&&value.includes('/')||numerator===0n&&d!==1n||d>1n&&numerator%2n===0n)return null;
 let shift=-(d.toString(2).length-1);if(numerator===0n)return 0;
 while(numerator%2n===0n){numerator/=2n;shift++}
 if(numerator.toString(2).length>53||shift< -1074||numerator.toString(2).length-1+shift>1023)return null;
 const coordinate=Number(numerator)*2**shift;
 return Number.isFinite(coordinate)&&coordinate!==0?coordinate:null;
}
