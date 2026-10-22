/* Record-field display only. Validation, query ownership and native bindings stay separate. */
(function(global){
 const operators={eq:'=',ne:'≠',gte:'≥',lte:'≤',contains:'contains',exists:'is present',missing:'is missing'},columns={service:'Service',duration_ns:'Span duration (ns)',status:'Recorded status',body:'Log message',severity:'Severity'};
 function qualified(field){return field.kind+':'+field.key}
 function label(field){return field.kind==='column'?(columns[field.key]||field.key):field.kind+' '+field.key}
 function operator(op){return operators[op]||op}
 function parts(p){return {field:qualified(p.field),operator:operator(p.op),...(['exists','missing'].includes(p.op)?{}:{value:(p.value??'')===''?'Empty string':JSON.stringify(p.value)})}}
 function selectorText(p){if(!p)return 'No additional selectors';if(p.all||p.any)return '('+(p.all||p.any).map(selectorText).join(p.all?' AND ':' OR ')+')';return label(p.field)+' '+operator(p.op)+(['exists','missing'].includes(p.op)?'':' '+JSON.stringify(p.value??''))}
 const api={qualified,label,operator,parts,selectorText};if(typeof module!=='undefined')module.exports=api;else global.GooFieldPresentation=api;
})(typeof window==='undefined'?globalThis:window);
