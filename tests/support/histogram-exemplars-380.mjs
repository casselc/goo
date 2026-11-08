import {createRequire} from 'node:module';
import * as Context from '../../ui/native-exemplar-context.mjs';
const {result}=createRequire(import.meta.url)('../fixtures/native-histogram.cjs');
export function exemplarPage(){
 const r=result(),stream=r.histograms.series[0].stream;stream.resource['service.namespace']='';r.question.fromNS='1000000001';r.question.toNS='5000000000';r.question.definition.source.namespace={mode:'value',value:''};
 const raw={resource_schema_url:'',resource:structuredClone(stream.resource),scope:stream.scope,version:stream.version,schema_url:stream.schemaURL,name:stream.name,unit:stream.unit,attributes:structuredClone(stream.attributes),start_ns:'1000000000',end_ns:'3000000000',temporality:2,flags:0,exemplar_total:'1',exemplar_span_total:'1',exemplar_time_total:'1',exemplar_value_total:'1',exemplar_attribute_total:'1',exemplar_times:['0'],exemplar_values:['NaN'],exemplar_trace_ids:['a'.repeat(32)],exemplar_span_ids:['b'.repeat(16)]};
 r.rows=[raw];r.histogramExemplars={version:'goo/histogram-exemplar-references/1',valueEncoding:'source-float64',originalValueType:'unknown',limited:false,omittedExports:0,exports:[{sourceRow:0,stream:structuredClone(stream),startNS:raw.start_ns,endNS:raw.end_ns,temporality:'cumulative',state:'available',items:[{index:0,timeNS:'0',traceId:'a'.repeat(32),spanId:'b'.repeat(16),storedValue:'NaN'}],total:'1',omitted:'0',diagnostics:[]}]};
 return{receipt:{id:'c'.repeat(64),question:{kind:'measurement',measurement:structuredClone(r.question)}},measurement:r};
}
export function exemplarDetail(page=exemplarPage()){
 const r=page.measurement,stream=r.histograms.series[0].stream;
 return{...Context.plan(r,stream,r.rows[0].end_ns,0),measurement:r.question,stream,basis:r.basis,readAt:r.readAt};
}
