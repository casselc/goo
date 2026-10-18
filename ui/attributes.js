/* Structured record inspection. Telemetry is always rendered as text, never HTML. */
(function(global){
 const maps=new Set(['ResourceAttributes','Attributes','SpanAttributes','LogAttributes']);
 const columns={ServiceName:'service',SpanName:'name',TraceId:'trace_id',SpanId:'span_id',ScopeName:'scope',ScopeVersion:'scope_version',StatusCode:'status',SpanKind:'kind',duration_ns:'duration_ns',duration_bucket:'duration_bucket',Body:'body',SeverityText:'severity'};
 const node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
 function type(value){return value===null?'null':Array.isArray(value)?'array':typeof value}
 function text(value){return typeof value==='string'?value:JSON.stringify(value)}
 function groups(record){
  const entries=(object,kind)=>Object.entries(object||{}).map(([key,value])=>({key,value,...(kind?{field:{kind,key}}:Object.hasOwn(columns,key)?{field:{kind:'column',key:columns[key]}}:{})}));
  const recordFields=Object.fromEntries(Object.entries(record).filter(([key])=>!maps.has(key)));
  return [{name:'Record fields',entries:entries(recordFields)},{name:'Resource attributes',entries:entries(record.ResourceAttributes,'resource')},
   {name:record.Body!==undefined?'Log attributes':'Span attributes',entries:entries(record.Attributes||record.SpanAttributes||record.LogAttributes,'attribute')}];
 }
 function matches(entry,query){const q=query.toLocaleLowerCase();return entry.key.toLocaleLowerCase().includes(q)||(text(entry.value)||'').toLocaleLowerCase().includes(q)}
 function valueNode(value,depth=0){
  const kind=type(value);
  if(kind==='object'||kind==='array'){
   const entries=Object.entries(value),detail=node('details',undefined,'attribute-object');detail.append(node('summary',`${kind==='array'?'Array':'Object'} · ${entries.length} ${kind==='array'?'items':'fields'}`));
   let built=false;detail.addEventListener('toggle',()=>{if(!detail.open||built)return;built=true;
    if(depth>=8){detail.append(node('p','Further nesting is available in Raw record.'));return}
    const values=node('dl');let shown=0;
    const more=node('button','Show more values');more.type='button';
    function append(){for(const [key,item]of entries.slice(shown,shown+50)){values.append(node('dt',key),node('dd'));values.lastChild.append(valueNode(item,depth+1))}shown=Math.min(shown+50,entries.length);more.hidden=shown===entries.length;more.textContent=`Show more values (${shown} of ${entries.length})`}
    more.onclick=append;append();detail.append(values,more);
   });return detail;
  }
  const display=value===''?'Empty string':value===null?'Null':String(value);
  if(display.length>240){const detail=node('details',undefined,'attribute-long');detail.append(node('summary',display.slice(0,160)+'…'),node('span',display,'attribute-value-text'));return detail}
  return node('span',display,'attribute-value-text'+(value===''||value===null?' attribute-empty':''));
 }
 function mount(container,record,{onFilter,onBreakdown,canRefine=()=>true,raw=true}={}){
  const model=groups(record),total=model.reduce((n,g)=>n+g.entries.length,0),root=node('div',undefined,'attribute-viewer');
  const tools=node('div',undefined,'attribute-tools'),label=node('label','Find an attribute'),search=node('input');search.type='search';search.placeholder='Search keys and values';search.setAttribute('aria-label','Search attribute keys and values');label.append(search);
  const scopeLabel=node('label','Scope'),scope=node('select');scope.setAttribute('aria-label','Attribute scope');
  for(const name of ['All scopes',...model.map(g=>g.name)]){const option=node('option',name);option.value=name;scope.append(option)}scopeLabel.append(scope);
  const count=node('span',undefined,'attribute-count');count.setAttribute('aria-live','polite');tools.append(label,scopeLabel,count);
  const content=node('div'),status=node('p',undefined,'attribute-feedback');status.setAttribute('role','status');root.append(tools,content,status);container.append(root);
  const copy=async(value,key)=>{try{await navigator.clipboard.writeText(text(value));status.textContent='Copied value for '+key}catch{status.textContent='Copy unavailable. Select the value to copy it manually.'}};
  function render(){content.replaceChildren();let found=0;
   for(const group of model){if(scope.value!=='All scopes'&&scope.value!==group.name)continue;const entries=group.entries.filter(e=>matches(e,search.value));found+=entries.length;if(!entries.length)continue;
    const section=node('div',undefined,'attribute-group'),heading=node('h4',group.name);heading.append(node('span',String(entries.length),'attribute-group-count'));section.append(heading);
    for(const entry of entries){const row=node('div',undefined,'attribute-row'),key=node('span',entry.key,'attribute-key'),value=node('div',undefined,'attribute-value'),actions=node('div',undefined,'attribute-actions');row.dataset.key=entry.key;row.setAttribute('role','group');row.setAttribute('aria-label',group.name+': '+entry.key);
     value.append(valueNode(entry.value),node('span',type(entry.value),'attribute-type'));
     const copyButton=node('button','Copy');copyButton.type='button';copyButton.setAttribute('aria-label','Copy value for '+entry.key);copyButton.onclick=()=>copy(entry.value,entry.key);actions.append(copyButton);
     if(entry.field&&typeof entry.value==='string'&&onFilter){const filter=node('button','Filter');filter.type='button';filter.disabled=!canRefine();filter.title=filter.disabled?'Choose All filters to narrow this question.':'Add an exact filter';filter.setAttribute('aria-label','Filter by '+entry.field.kind+':'+entry.field.key);filter.onclick=()=>onFilter(entry.field,entry.value);actions.append(filter)}
     if(entry.field&&onBreakdown){const breakdown=node('button','Break down');breakdown.type='button';breakdown.setAttribute('aria-label','Break down by '+entry.field.kind+':'+entry.field.key);breakdown.onclick=()=>onBreakdown(entry.field);actions.append(breakdown)}
     row.append(key,value,actions);section.append(row);
    }content.append(section);
   }count.textContent=`${found} of ${total} fields`;if(!found)content.append(node('p','No attributes match this search.','attribute-no-match'));
  }
  search.oninput=render;scope.onchange=render;render();
  if(raw){const detail=node('details',undefined,'raw-record');detail.append(node('summary','Raw record'),node('pre',JSON.stringify(record,null,2)));root.append(detail)}
  return root;
 }
 function disclosure(record,options={}){
  const detail=node('details',undefined,'record-attributes'),host=node('div');detail.append(node('summary','Captured attributes'),host);
  // Keep raw evidence available without constructing thousands of primary-view
  // nodes for unopened records on a 100-row page.
  const raw=node('details',undefined,'raw-record');raw.append(node('summary','Raw record'),node('pre',JSON.stringify(record,null,2)));detail.append(raw);
  let built=false;detail.addEventListener('toggle',()=>{if(detail.open&&!built){built=true;mount(host,record,{...options,raw:false})}});return detail;
 }
 const api={groups,type,text,matches,mount,disclosure};if(typeof module!=='undefined')module.exports=api;else global.GooAttributes=api;
})(typeof window==='undefined'?globalThis:window);
