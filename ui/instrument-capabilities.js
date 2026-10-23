/* Metadata-only capability hints. Preview/Run remain authoritative. */
(function(global){
 const ruleSet='OpenTelemetry semantic conventions v1.30.0',schema='https://opentelemetry.io/schemas/1.30.0';
 function describe(instrument){
  const i=instrument||{},http=i.metric==='http.server.request.duration',eligible=i.supported===true&&!!i.definition,dimensions=new Set(i.dimensions||[]),reasons=[];
  if(!i.scope)reasons.push('Instrumentation scope is omitted.');
  if(!i.unit)reasons.push('Unit is omitted.');
  if(['counter','histogram'].includes(i.kind)&&![1,2].includes(i.temporality))reasons.push('Aggregation temporality is unsupported.');
  if(i.kind==='counter'&&!i.monotonic)reasons.push('Non-monotonic sums are not supported by the counter kernel.');
  if(i.kind==='histogram'&&!['s','ms'].includes(i.unit))reasons.push('The current histogram kernel requires seconds or milliseconds.');
  if(!eligible&&!reasons.length)reasons.push('No valid package binding is available from the current kernels.');
  const hints=[];
  if(http){
   if(i.kind!=='histogram')hints.push('The pinned HTTP duration convention requires a histogram. Use the generic binding.');
   if(i.unit!=='s')hints.push('The pinned HTTP duration convention requires unit s. No automatic conversion is applied.');
   hints.push(i.schemaURL===schema?'Schema URL matches the pinned rule set; this does not verify conformance.':i.schemaURL?'Other schema URL; no schema migration or conformance is inferred.':'Schema URL omitted; convention conformance is unverified.');
   hints.push(dimensions.has('http.request.method')?'HTTP method key observed.':'HTTP method key not observed in this window.');
   hints.push(dimensions.has('http.response.status_code')?'HTTP status key observed; no error policy or denominator has been verified.':'HTTP status key not observed; no HTTP error-rate view is offered.');
   hints.push(i.kind==='histogram'&&eligible&&['s','ms'].includes(i.unit)?'Run can evaluate retained histogram observations and bucket ranges per exact stream. Complete interval coverage is still required; no fleet percentile or request rate is inferred.':'HTTP duration views are not established by this metadata. Retain the generic binding if it is eligible.');
  }else hints.push('Generic exact metric binding; no application convention is inferred.');
  return {profile:http?'http-server-duration':'generic',title:http?'HTTP server duration candidate':'Generic metric',ruleSet:http?ruleSet:null,shapeMatches:http&&i.kind==='histogram'&&i.unit==='s',eligible,verification:'not-run',reasons,hints};
 }
 function catalog(instruments){return instruments.some(i=>i.metric==='http.server.request.duration')?'HTTP server duration metadata observed. Convention hints do not verify data coverage.':'HTTP server duration not observed in this bounded window. This is not proof that the application lacks HTTP instrumentation.'}
 const api={describe,catalog,ruleSet,schema};if(typeof module!=='undefined')module.exports=api;else global.GooInstrumentCapabilities=api;
})(typeof window==='undefined'?globalThis:window);
