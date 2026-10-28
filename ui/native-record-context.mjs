/* Pure projection of a validated native stream into coincident record context.
 * Keep admission aligned with exploration.PlanNativeRecordContext; no fetches. */

const global=globalThis;
const has=(o,k)=>Object.prototype.hasOwnProperty.call(o||{},k),plain=x=>x!==null&&typeof x==='object'&&!Array.isArray(x),integer=Number.isSafeInteger;
const fail=()=>{throw new Error('Native record context requires a valid exact stream and bounded native interval')};
const bytes=s=>new TextEncoder().encode(s).length;
function text(s,max){return typeof s==='string'&&bytes(s)<=max&&!s.includes('\0')&&!/[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/u.test(s)}
function str(o,k){const v=o?.[k];if(v===undefined||v===null)return '';if(typeof v!=='string')fail();return v}
function mapping(x,max,keyMax,valueMax){if(x==null)return {};if(!plain(x)||Object.keys(x).length>max)fail();for(const [k,v]of Object.entries(x))if(!k||!text(k,keyMax)||!text(v,valueMax))fail();return x}
function streamValue(s){if(!plain(s))fail();return {resourceSchemaURL:str(s,'resourceSchemaURL'),resource:mapping(s.resource,64,256,2048),scope:str(s,'scope'),version:str(s,'version'),schemaURL:str(s,'schemaURL'),name:str(s,'name'),unit:str(s,'unit'),attributes:mapping(s.attributes,64,256,2048),_attributesNil:s.attributes==null,_resourceNil:s.resource==null}}
function same(a,b){for(const k of ['resourceSchemaURL','scope','version','schemaURL','name','unit'])if(a[k]!==b[k])return false;for(const k of ['resource','attributes']){if(Object.keys(a[k]).length!==Object.keys(b[k]).length)return false;for(const [key,value]of Object.entries(a[k]))if(!has(b[k],key)||b[k][key]!==value)return false}return true}
function ns(v){if(typeof v!=='string'||!/^(0|[1-9][0-9]*)$/.test(v))fail();const n=BigInt(v);if(n>18446744073709551615n)fail();return n}
function namespace(n){if(n==null)return null;if(!plain(n)||!['missing','value'].includes(n.mode)||!text(str(n,'value'),128)||n.mode==='missing'&&str(n,'value')!=='')fail();return n}
function matches(n,r){return !n||n.mode==='missing'?!n||!has(r,'service.namespace'):has(r,'service.namespace')&&r['service.namespace']===str(n,'value')}
function validateStream(s,d,b,n){
  if(!text(s.resourceSchemaURL,1024))fail();
  // Match Go's stream wire encoding size, including HTML/line-separator escaping.
  const wire={...(s.resourceSchemaURL?{resourceSchemaURL:s.resourceSchemaURL}:{}),resource:s._resourceNil?null:s.resource,scope:s.scope,version:s.version,schemaURL:s.schemaURL,name:s.name,unit:s.unit,attributes:s._attributesNil?null:s.attributes};
  if(bytes(JSON.stringify(wire).replace(/[<>&\u2028\u2029]/g,c=>'\\u'+c.charCodeAt(0).toString(16).padStart(4,'0')))>16384)fail();
  if((b.resource?.kind==='host'?s.resource['host.id']!==b.resource.id:!b.services.includes(s.resource['service.name'])||!s.resource['service.instance.id'])||s.resource['deployment.environment.name']!==d.environment||!matches(n,s.resource)||s.scope!==b.scope||s.version!==str(b,'version')||s.schemaURL!==str(b,'schemaURL')||s.name!==b.metric||s.unit!==b.unit)fail();
  for(const [k,v]of Object.entries(b.attributes||{}))if(!has(s.attributes,k)||s.attributes[k]!==v)fail();
 }
function hostCounterDefinition(d){const b=d?.source;return b?.resource?.kind==='host'&&d.kind==='counter'&&d.counterDomain==='nonnegative-decimal'&&b.schemaURL==='https://opentelemetry.io/schemas/1.37.0'&&(b.metric==='system.cpu.time'&&b.unit==='s'||['system.network.io','system.disk.io'].includes(b.metric)&&b.unit==='By')}
function plan(parent,actual,signal,fromNS,toNS){
  if(!['logs','spans'].includes(signal)||!plain(parent)||!plain(parent.definition))fail();
  const d=parent.definition,b=d.source;if(!plain(b))fail();const n=namespace(b.namespace),hostCPU=hostCounterDefinition(d);if(b.resource&&!hostCPU)fail();if(d.counterDomain!==undefined&&(d.counterDomain!=='nonnegative-decimal'||d.kind!=='counter'||d.grain!=='per-stream'))fail();if(hostCPU&&(!b.resource.id||!text(b.resource.id,512)||!b.exactStream||n||d.grain!=='per-stream'||str(d,'stockDomain')!==''||str(d,'stateProfile')!==''))fail();
  if(d.apiVersion!=='goo/measurement-package/v1alpha1'||typeof d.name!=='string'||!/^[A-Za-z][A-Za-z0-9_.-]{0,127}$/.test(d.name)||!d.title||!text(d.title,160)||!d.environment||!text(d.environment,128)||!integer(d.lookbackMinutes)||d.lookbackMinutes<1||d.lookbackMinutes>120||!integer(d.maxGapMs)||d.maxGapMs<1||d.maxGapMs>600000||!['gauge','counter','histogram'].includes(d.kind)||!['per-stream','consumer-group'].includes(d.grain)||d.lifecycle!=null)fail();
  if(!Array.isArray(b.services)||(hostCPU?b.services.length!==0:b.services.length<1||b.services.length>8)||new Set(b.services).size!==b.services.length||b.services.some(s=>!s||!text(s,128))||!b.scope||!text(b.scope,256)||!text(str(b,'version'),128)||!text(str(b,'schemaURL'),1024)||!b.metric||!text(b.metric,256)||!b.unit||!text(b.unit,64))fail();
  mapping(b.attributes,16,128,512);
  const a=parent.fromMs,z=parent.toMs;if(!integer(a)||!integer(z)||a<0||z<=a||z-a>7200000||z>=4102444800000)fail();
  let lo=BigInt(a)*1000000n,hi=BigInt(z)*1000000n;
  if(str(parent,'fromNS')!==''||str(parent,'toNS')!==''){lo=ns(parent.fromNS);hi=ns(parent.toNS);if(hi<=lo||hi-lo>7200000000000n||hi>=4102444800000000000n||BigInt(a)!==lo/1000000n||BigInt(z)!==(hi+999999n)/1000000n)fail()}
  if(d.histogramBucketMs!=null){const w=d.histogramBucketMs;if(d.kind!=='histogram'||!integer(w)||w<1||w>7200000||(hi-lo-1n)/(BigInt(w)*1000000n)+1n>200n)fail()}
  const s=streamValue(actual);if(b.exactStream!=null){const old=streamValue(b.exactStream);validateStream(old,d,b,n);if(!same(old,s))fail()}validateStream(s,d,b,n);
  const from=ns(fromNS),to=ns(toNS);if(to<=from||from<lo||to>hi||d.kind==='gauge'&&to-from!==1n)fail();
  const fromMs=Number(from/1000000n),toMs=Number((to+999999n)/1000000n);if(toMs-fromMs>7200000)fail();
  const r=s.resource,namespaceValue=r['service.namespace'];if(!hostCPU&&(!text(r['service.instance.id'],1024)||has(r,'service.namespace')&&!text(namespaceValue,1024)))fail();
  const leaf=(kind,key,op,value)=>({field:{kind,key},op,...(value!==undefined&&value!==''?{value}: {})});
  return {question:{signal,environment:d.environment,fromMs,toMs,where:{all:[...(hostCPU?[leaf('resource','host.id','eq',r['host.id'])]:[leaf('column','service','eq',r['service.name']),has(r,'service.namespace')?leaf('resource','service.namespace','eq',namespaceValue):leaf('resource','service.namespace','missing'),leaf('resource','service.instance.id','eq',r['service.instance.id'])]),leaf('column','timestamp_ns','gte',fromNS),leaf('column','timestamp_ns','lte',String(to-1n))]},limit:100,offset:0},fromNS,toNS,grain:d.kind==='gauge'?'point-context':'interval-context'};
 }

export {plan,hostCounterDefinition};
