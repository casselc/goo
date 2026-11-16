import {scalarValue} from '../../ui/raw-observation-values.mjs';
export function bank(entries=[{}],{fromMs=1700000000000,toMs=1700000060000,unit='1'}={}){
 const question={environment:'prod',service:'worker',fromMs,toMs},instrument={kind:'gauge',scope:'otel.test',version:'1',schemaURL:'',metric:'temperature',unit};
 const rows=entries.map((entry,index)=>{
  const rawValue=Object.hasOwn(entry,'rawValue')?entry.rawValue:'3',flags=Object.hasOwn(entry,'flags')?entry.flags:0;
  const raw={resource:{'service.name':'worker','deployment.environment.name':'prod'},attributes:entry.attributes??{},resource_schema_url:'',scope:instrument.scope,version:instrument.version,schema_url:'',name:instrument.metric,unit,end_ns:entry.endNS??String(BigInt(fromMs)*1000000n+BigInt(index+1)),start_ns:'0',value:rawValue,...(flags===undefined?{}:{flags})};
  const scalar=scalarValue(rawValue),diagnostics=[];if(scalar.state!=='finite')diagnostics.push({field:'value',code:scalar.state});
  let flagsState='absent',reported=null;if(flags!==undefined&&flags!==null){if(!Number.isInteger(flags)||flags<0||flags>4294967295)flagsState='invalid';else{flagsState='valid';reported=flags;if((BigInt(flags)&~1n)!==0n)diagnostics.push({field:'flags',code:'unknown-bits'});if(flags&1){scalar.value=null;scalar.state='absent';diagnostics.push({field:'value',code:'no-recorded-value'})}}}
  if(flagsState!=='valid')diagnostics.push({field:'flags',code:flagsState});
  return {occurrenceKey:String(index+1),stream:{resourceSchemaURL:'',resource:raw.resource,scope:instrument.scope,version:instrument.version,schemaURL:'',name:instrument.metric,unit,attributes:raw.attributes},startNS:'0',startState:'valid',endNS:raw.end_ns,rawValue,value:scalar.value,valueState:scalar.state,flags:reported,flagsState,diagnostics,rawRow:raw};
 });
 return {version:'goo/raw-gauge-observations/v1',question,instrument,readAt:'2026-11-15T22:00:00.123456789Z',limit:10000,rows};
}
