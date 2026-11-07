/* Reported export references only. No bucket attribution, evaluation or reads. */
import * as Environment from './record-environment.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
import {validateHistogramMeasurementQuestion,validateMeasurementWindowSelector,sameNativeStream} from './measurement-scope.mjs';
const object=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const own=(value,key)=>Object.hasOwn(value,key);
const fail=message=>{throw Error(message)};
function closed(value,keys){if(!object(value)||Object.keys(value).some(key=>!keys.includes(key)))fail('Invalid reported exemplar metadata.');}
function ns(value){if(typeof value!=='string'||!/^(0|[1-9][0-9]*)$/.test(value)||value.length>19||BigInt(value)>9223372036854775807n)fail('Invalid reported exemplar timestamp.');return BigInt(value)}
const trace=value=>typeof value==='string'&&/^[a-f0-9]{32}$/.test(value)&&!/^0+$/.test(value);
const span=value=>typeof value==='string'&&/^[a-f0-9]{16}$/.test(value)&&!/^0+$/.test(value);
const text=value=>typeof value==='string'&&value.isWellFormed()&&!value.includes('\0')&&new TextEncoder().encode(value).length<=128;
function storedFloat(value){if(!text(value)||!value||value.trim()!==value||/[xXpP]/.test(value))return false;if(/^(?:nan|[+-]?(?:inf|infinity))$/i.test(value))return true;const digit='[0-9](?:_?[0-9])*',decimal=new RegExp('^[+-]?(?:'+digit+'(?:\\.(?:'+digit+')?)?|\\.'+digit+')(?:[eE][+-]?'+digit+')?$');return decimal.test(value)&&Number.isFinite(Number(value.replaceAll('_','')))}
function checkedParent(parent,stream){const bounds=validateHistogramMeasurementQuestion(parent);if(parent.definition.grain!=='per-stream'||parent.definition.histogramBucketMs!==undefined||parent.definition.environmentSelector===null)fail('Reported exemplars require a per-stream service histogram question.');validateMeasurementWindowSelector({kind:'measurement-window',question:parent,stream,...bounds});return bounds}
function rowStream(row){if(!object(row))fail('Invalid retained source row.');return{resourceSchemaURL:row.resource_schema_url??'',resource:row.resource??{},scope:row.scope,version:row.version,schemaURL:row.schema_url,name:row.name,unit:row.unit,attributes:row.attributes??{}}}
// Projection admission is intentionally independent from distribution validity.
export function select(result,stream,endNS){
 const bounds=checkedParent(result.question,stream);ns(endNS);if(BigInt(endNS)<BigInt(bounds.fromNS)||BigInt(endNS)>BigInt(bounds.toNS))fail('Selected export is outside the retained question.');
 const report=result.histogramExemplars;if(report===undefined)return null;
 closed(report,['version','valueEncoding','originalValueType','exports','limited','omittedExports']);
 if(report.version!=='goo/histogram-exemplar-references/1'||report.valueEncoding!=='source-float64'||report.originalValueType!=='unknown'||!Array.isArray(report.exports)||typeof report.limited!=='boolean'||!Number.isInteger(report.omittedExports)||report.omittedExports<0||report.omittedExports>10000||report.exports.length>10000)fail('Unsupported reported exemplar metadata.');
 if(!Array.isArray(result.rows)||result.rows.length>10000)fail('Retained source rows are unavailable.');
 const matches=[];for(let index=0;index<result.rows.length;index++){const row=result.rows[index];if(sameNativeStream(rowStream(row),stream)&&row.end_ns===endNS)matches.push(index)}
 if(matches.length!==1)fail('Source export association is ambiguous or unavailable.');
 const candidates=report.exports.filter(item=>sameNativeStream(item.stream,stream)&&item.endNS===endNS);if(!candidates.length)return null;if(candidates.length!==1)fail('Reported source export association is ambiguous.');
 const selected=candidates[0],raw=result.rows[matches[0]];
 closed(selected,['sourceRow','stream','startNS','endNS','temporality','state','items','total','omitted','diagnostics']);
 if(!Number.isInteger(selected.sourceRow)||selected.sourceRow!==matches[0]||selected.startNS!==raw.start_ns||selected.endNS!==raw.end_ns||selected.temporality!==(raw.temporality===1?'delta':raw.temporality===2?'cumulative':'')||!['available','unavailable'].includes(selected.state)||!Array.isArray(selected.items)||!Array.isArray(selected.diagnostics))fail('Reported export differs from its retained source row.');
 ns(selected.startNS);ns(selected.endNS);
 const count=value=>{if(typeof value!=='string'||!/^(0|[1-9][0-9]*)$/.test(value)||value.length>20||BigInt(value)>18446744073709551615n)fail('Invalid reported exemplar count.');return BigInt(value)};
 if(selected.state==='available'||selected.total!==''&&selected.omitted!==''){count(selected.total);count(selected.omitted)}else if(selected.total!==''||selected.omitted!=='')fail('Invalid unavailable exemplar counts.');if(selected.diagnostics.some(value=>!['metadata-not-retained','metadata-invalid','no-recorded-value','sample-limit'].includes(value)))fail('Invalid exemplar diagnostic.');
 if(selected.state==='available'){
  const total=count(raw.exemplar_total);for(const key of ['exemplar_span_total','exemplar_time_total','exemplar_value_total','exemplar_attribute_total'])if(count(raw[key])!==total)fail('Retained exemplar arrays are unaligned.');
  const length=Number(total>32n?32n:total);for(const key of ['exemplar_times','exemplar_values','exemplar_trace_ids','exemplar_span_ids'])if(!Array.isArray(raw[key])||raw[key].length!==length)fail('Retained exemplar arrays are unavailable.');
  if(raw.flags&1||selected.diagnostics.join(',')!==(total>BigInt(length)?'sample-limit':'')||selected.total!==raw.exemplar_total||selected.items.length!==length||count(selected.omitted)!==total-BigInt(length))fail('Reported exemplar projection changed.');
  selected.items.forEach((item,index)=>{if(item.index!==index||item.timeNS!==raw.exemplar_times[index]||item.storedValue!==raw.exemplar_values[index]||item.traceId!==raw.exemplar_trace_ids[index]||item.spanId!==raw.exemplar_span_ids[index])fail('Reported exemplar reference differs from its retained ordinal.')});
 }else if(selected.items.length)fail('Unavailable exemplar metadata cannot authorize references.');
 for(const item of selected.items){closed(item,['index','timeNS','traceId','spanId','storedValue']);if(!Number.isInteger(item.index)||item.index<0||!text(item.traceId)||!text(item.spanId)||!storedFloat(item.storedValue))fail('Invalid reported exemplar reference.');ns(item.timeNS)}
 return structuredClone(selected);
}
function planFromSelected(result,stream,selected,ordinal){
 if(!selected||selected.state!=='available'||!Number.isInteger(ordinal)||ordinal<0||ordinal>=selected.items.length)fail('Selected reported exemplar is unavailable.');
 const item=selected.items[ordinal];if(!trace(item.traceId))fail('Reported trace ID is unavailable.');const bounds=checkedParent(result.question,stream),parent=result.question,source=parent.definition.source,resource=stream.resource||{};
 const namespace=source.namespace??(own(resource,'service.namespace')?{mode:'value',value:resource['service.namespace']}:{mode:'missing'}),leaf=(kind,key,op,value)=>({field:{kind,key},op,...(value!==undefined&&value!==''?{value}:{})});
 const question={signal:'spans',environment:parent.definition.environment??'',...Environment.questionFields(parent.definition),fromMs:parent.fromMs,toMs:parent.toMs,where:{all:[leaf('column','trace_id','eq',item.traceId),namespace.mode==='missing'?leaf('resource','service.namespace','missing'):leaf('resource','service.namespace','eq',namespace.value??''),leaf('column','timestamp_ns','gte',bounds.fromNS),leaf('column','timestamp_ns','lte',String(BigInt(bounds.toNS)-1n))]},limit:100,offset:0};
 QuestionProtocol.validateRecordQuestion(question);
 return{relation:'reported-exemplar',question,...bounds,sourceRow:selected.sourceRow,exportStartNS:selected.startNS,exportEndNS:selected.endNS,exemplarIndex:item.index,timeNS:item.timeNS,traceId:item.traceId,...(span(item.spanId)?{spanId:item.spanId}:{})};
}

// One display projection admits the selected source export once; it grants no
// click authority. Navigation always calls public plan again against current rows.
export function inspect(result,stream,endNS){
 const selected=select(result,stream,endNS);if(!selected)return null;
 return{export:selected,references:selected.items.map((_,ordinal)=>{try{return{context:planFromSelected(result,stream,selected,ordinal)}}catch(error){return{unavailable:error.message}}})};
}
export function plan(result,stream,endNS,ordinal){return planFromSelected(result,stream,select(result,stream,endNS),ordinal)}
