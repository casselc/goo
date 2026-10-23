/* Record-field display only. Validation, query ownership and native bindings stay separate. */
(function(global){
 const operators={eq:'=',ne:'≠',gte:'≥',lte:'≤',contains:'contains',exists:'is present',missing:'is missing'},columns={service:'Service',duration_ns:'Span duration (ns)',timestamp_ns:'Event time (UTC)',status:'Recorded status',body:'Log message',severity:'Severity'};
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
 const api={qualified,label,operator,parts,selectorText};if(typeof module!=='undefined')module.exports=api;else global.GooFieldPresentation=api;
})(typeof window==='undefined'?globalThis:window);
