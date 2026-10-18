/* Render only Go-checked scope metadata; this module never queries a source. */
(function(global){
 const node=(tag,text)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=String(text);return n};
 const grains={'span-record':'Span records','log-record':'Log records','per-stream':'Per metric stream','scoped-subject':'Scoped subjects','observed-instrument-metadata':'Observed instrument metadata'};
 function summary(t){return `${grains[t.target.grain]||t.target.grain} · ${t.target.environment} · ${new Date(t.target.fromMs).toISOString()} → ${new Date(t.target.toMs).toISOString()}`}
 function render(host,t){
  host.dataset.scopeRelation=t.relation;host.append(node('p',summary(t)));
  for(const [title,values]of [['Keeps',t.preserves],['Changes',t.replaces]]){host.append(node('strong',title));const list=node('ul');for(const v of values)list.append(node('li',v));if(!values.length)list.append(node('li','No change'));host.append(list)}
  const more=node('details');more.className='question-scope-detail';more.append(node('summary','Source binding and OTel semantics'),node('p','Source: '+t.origin.sourceBinding+' → '+t.target.sourceBinding),node('p',t.target.convention),node('p',t.target.schemaMetadata));
  if(t.target.binding){const b=t.target.binding;more.append(node('p','Declared services: '+b.services.join(', ')),node('p',`Instrumentation scope: ${b.scope??'not selected'} · library version: ${b.version??'not selected'} · scope schema URL: ${b.schemaURL??'not selected'}`));}
  if(t.measurePlan){const measures=node('ul');for(const m of t.measurePlan.measures)measures.append(node('li',m.id+' · '+m.operator+' · '+m.outputType+' · '+m.unit+' · merge: '+m.merge));more.append(node('h4','Typed record measures'),measures);if(t.measurePlan.percentilePolicy)more.append(node('p',t.measurePlan.percentilePolicy));}
  host.append(more);
 }
 const api={summary,render};if(typeof module!=='undefined')module.exports=api;else global.GooQuestionScope=api;
})(typeof window==='undefined'?globalThis:window);
