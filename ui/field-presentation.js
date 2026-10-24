/* Record-field display only. Validation, query ownership and native bindings stay separate. */
(function(global){
 const operators={eq:'=',ne:'≠',gte:'≥',lte:'≤',contains:'contains',exists:'is present',missing:'is missing'},columns={service:'Service',duration_ns:'Span duration (ns)',timestamp_ns:'Event time (UTC)',status:'Recorded status',body:'Log message',severity:'Severity'};
 // Curated display hints, checked against OTel semconv v1.43.0 (legacy spellings
 // against earlier SDK registries). Exact namespace/key lookup; no alias rewriting.
 const conventions=new Map([
  ['resource','service.name','Service','Service','Logical service name shared by its instances.'],
  ['resource','service.namespace','Service namespace','Service','Groups services within a shared naming namespace.'],
  ['resource','service.instance.id','Service instance','Service','Identifies one service instance; distinct from the service name.'],
  ['resource','service.version','Service version','Service','Reported version of the service.'],
  ['resource','deployment.environment.name','Environment','Service','Reported deployment environment, such as production or staging.'],
  ['attribute','http.route','HTTP route','HTTP','Route template such as /orders/{id}; group by handler rather than individual URLs.'],
  ['attribute','http.request.method','HTTP request method','HTTP','Reported request method, such as GET or POST.'],
  ['attribute','http.response.status_code','HTTP response status','HTTP','HTTP response code; separate from reported span status and business outcome.'],
  ['attribute','rpc.system.name','RPC system','RPC','Reported RPC framework or system.'],
  ['attribute','rpc.method','RPC method','RPC','Logical RPC method; its naming convention depends on the instrumentation schema.'],
  ['attribute','rpc.response.status_code','RPC response status','RPC','RPC response code; interpretation depends on the RPC system, not span status.'],
  ['attribute','db.system.name','Database system','Database','Database product, such as PostgreSQL or Redis.'],
  ['attribute','db.namespace','Database namespace','Database','Database namespace being accessed; interpretation depends on the database system.'],
  ['attribute','db.response.returned_rows','Database returned rows','Database','Reported returned-row count. Declare the numeric type and unit explicitly; this key alone does not validate its values.'],
  ['attribute','db.operation.name','Database operation','Database','Reported database operation, such as SELECT or GET.'],
  ['attribute','db.collection.name','Database collection','Database','Collection or table being accessed, where reported.'],
  ['attribute','messaging.system','Messaging system','Messaging','Messaging product or system used for this operation.'],
  ['attribute','messaging.destination.name','Message destination','Messaging','Reported queue, topic or destination; not a causal trace relationship.'],
  ['attribute','messaging.operation.type','Messaging operation','Messaging','Operation type, such as send, receive or process.'],
  ['attribute','error.type','Error type','Error','Reported error classification; absence does not establish success.'],
  ['attribute','http.method','HTTP request method','HTTP','Earlier HTTP method spelling; kept as its own exact field.',true],
  ['attribute','http.status_code','HTTP response status','HTTP','Earlier HTTP response-code spelling; separate from span status and business outcome.',true],
  ['attribute','rpc.system','RPC system','RPC','Earlier RPC system spelling; kept as its own exact field.',true],
  ['attribute','rpc.service','RPC service','RPC','Earlier RPC service field; not interchangeable with resource service.name or rpc.method.',true],
  ['attribute','db.system','Database system','Database','Earlier database-system spelling; kept as its own exact field.',true],
  ['attribute','db.name','Database namespace','Database','Earlier database-name spelling; kept as its own exact field.',true],
  ['attribute','db.operation','Database operation','Database','Earlier database-operation spelling; kept as its own exact field.',true],
  ['attribute','messaging.destination','Message destination','Messaging','Earlier destination spelling; kept as its own exact field.',true],
  ['resource','deployment.environment','Environment','Service','Earlier environment spelling; kept as its own exact field.',true],
  ['column','status','Reported span status','Span','Exporter-reported OTel span status; Unset or absent status does not establish business success.']
 ].map(([kind,key,label,group,description,legacy=false])=>[JSON.stringify([kind,key]),Object.freeze({label,group,description,legacy})]));
 function hint(field){const value=conventions.get(JSON.stringify([field?.kind,field?.key]));return value?{...value}:null}
 function searchText(field){const h=hint(field);return [qualified(field),h?.label,h?.group,h?.description,h?.legacy?'legacy':''].filter(Boolean).join(' ')}
 function qualified(field){return field.kind+':'+field.key}
 function label(field){return field.kind==='column'?(columns[field.key]||field.key):field.kind+' '+field.key}
 function operator(op){return operators[op]||op}
 function valueText(field,value){
  if(field?.kind==='column'&&field.key==='timestamp_ns'&&typeof value==='string'&&/^(0|[1-9][0-9]*)$/.test(value)){
   const ns=BigInt(value);if(ns<=18446744073709551615n)return new Date(Number(ns/1000000n)).toISOString().slice(0,19)+'.'+String(ns%1000000000n).padStart(9,'0')+'Z';
  }
  return (value??'')===''?'Empty string':JSON.stringify(value);
 }
 function parts(p){return {field:qualified(p.field),operator:operator(p.op),...(['exists','missing'].includes(p.op)?{}:{value:valueText(p.field,p.value??'')})}}
 function selectorText(p){if(!p)return 'No additional selectors';if(p.all||p.any)return '('+(p.all||p.any).map(selectorText).join(p.all?' AND ':' OR ')+')';return label(p.field)+' '+operator(p.op)+(['exists','missing'].includes(p.op)?'':' '+(p.field?.kind==='column'&&p.field.key==='timestamp_ns'?valueText(p.field,p.value??''):JSON.stringify(p.value??'')))}
 const api={qualified,label,operator,parts,selectorText,hint,searchText};if(typeof module!=='undefined')module.exports=api;else global.GooFieldPresentation=api;
})(typeof window==='undefined'?globalThis:window);
