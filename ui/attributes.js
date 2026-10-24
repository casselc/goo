/* Structured record inspection. Telemetry is always rendered as text, never HTML. */
(function(global){
 const presentation=()=>typeof module!=='undefined'?require('./field-presentation'):global.GooFieldPresentation;
 const maps=new Set(['ResourceAttributes','Attributes','SpanAttributes','LogAttributes']);
 const columns={ServiceName:'service',SpanName:'name',TraceId:'trace_id',SpanId:'span_id',ScopeName:'scope',ScopeVersion:'scope_version',StatusCode:'status',SpanKind:'kind',duration_ns:'duration_ns',duration_bucket:'duration_bucket',Body:'body',SeverityText:'severity',severity_number:'severity_number',exception_event:'exception_event',exception_type:'exception_type',exception_type_state:'exception_type_state'};
 const node=(tag,text,cls)=>{const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n};
 function type(value){return value===null?'null':Array.isArray(value)?'array':typeof value}
 function text(value){return typeof value==='string'?value:JSON.stringify(value)}
 function groups(record){
  const entries=(object,kind)=>Object.entries(object||{}).map(([key,value])=>({key,value,...(kind?{field:{kind,key}}:Object.hasOwn(columns,key)?{field:{kind:'column',key:columns[key]}}:{})}));
  const recordFields=Object.fromEntries(Object.entries(record).filter(([key])=>!maps.has(key)));
  return [{name:'Record fields',entries:entries(recordFields)},{name:'Resource attributes',entries:entries(record.ResourceAttributes,'resource')},
   {name:record.Body!==undefined?'Log attributes':'Span attributes',entries:entries(record.Attributes||record.SpanAttributes||record.LogAttributes,'attribute')}];
 }
 // Shared recorded string values only: absence, null and coercion never imply a predicate.
 function selectable(entry){const value=entry.value,key=entry.field?.key;return ['column','resource','attribute'].includes(entry.field?.kind)&&typeof value==='string'&&typeof key==='string'&&key.length>0&&new TextEncoder().encode(value).length<=1024&&new TextEncoder().encode(key).length<=128&&!/[\u0000\uD800-\uDFFF]/u.test(value)&&!/[\u0000\uD800-\uDFFF]/u.test(key)}
 // Bounded supplied observations, never population coverage or inferred map availability.
 function suggestionData(records=[],fields=[]){
  const rows=records.slice(0,100),catalog=new Map();let limited=records.length>100;
  function admit(field){try{if(!selectable({field,value:''}))return;const id=JSON.stringify([field.kind,field.key]);if(!catalog.has(id)){if(catalog.size>=64){limited=true;return}catalog.set(id,{field:structuredClone(field),values:[]})}return catalog.get(id)}catch{return}}
  fields.forEach(admit);const models=rows.map(record=>groups(record));
  for(const model of models)for(const group of model)for(const entry of group.entries)if(selectable(entry))admit(entry.field);
  function value(item,v){if(item.values.includes(v))return;if(item.values.length>=12){limited=true;if(v===null||v===''){const index=item.values.findLastIndex(value=>value!==null&&value!=='');if(index>=0)item.values[index]=v}return}item.values.push(v)}
  rows.forEach((record,index)=>{const entries=new Map(models[index].flatMap(g=>g.entries).filter(e=>e.field).map(e=>[JSON.stringify([e.field.kind,e.field.key]),e]));
   for(const [id,item]of catalog){const entry=entries.get(id);if(entry){if(selectable(entry))value(item,entry.value);continue}
    const map=item.field.kind==='resource'?record.ResourceAttributes:item.field.kind==='attribute'?(record.Attributes||record.SpanAttributes||record.LogAttributes):undefined;
    if(map&&typeof map==='object'&&!Array.isArray(map)&&!Object.hasOwn(map,item.field.key))value(item,null);
   }
  });return {fields:[...catalog.values()],rows:rows.length,limited};
 }
 function commonGroups(records){
  if(!Array.isArray(records)||!records.length||records.length>20)throw Error('Select 1–20 captured page observations.');
  const models=records.map(groups),identity=e=>JSON.stringify([e.field?.kind,e.field?.key,e.value]);
  return models[0].map((group,i)=>{const sets=models.slice(1).map(m=>new Set(m[i].entries.filter(e=>selectable(e)).map(identity)));return {name:group.name,entries:group.entries.filter(e=>selectable(e)&&sets.every(s=>s.has(identity(e)))).map(e=>structuredClone(e))}});
 }
 function matches(entry,query){const q=query.toLocaleLowerCase();return (entry.key+' '+(entry.field?presentation()?.searchText(entry.field)||'':'')).toLocaleLowerCase().includes(q)||(text(entry.value)||'').toLocaleLowerCase().includes(q)}
 function renderValue(value,depth=0){
  const kind=type(value);
  if(kind==='object'||kind==='array'){
   const entries=Object.entries(value),detail=node('details',undefined,'attribute-object');detail.append(node('summary',`${kind==='array'?'Array':'Object'} · ${entries.length} ${kind==='array'?(entries.length===1?'item':'items'):(entries.length===1?'field':'fields')}`));
   let built=false;detail.addEventListener('toggle',()=>{if(!detail.open||built)return;built=true;
    if(depth>=8){detail.append(node('p','Further nesting is available in Raw value.'));return}
    const values=node('dl');let shown=0;
    const more=node('button','Show more values');more.type='button';
    function append(){for(const [key,item]of entries.slice(shown,shown+50)){values.append(node('dt',key),node('dd'));values.lastChild.append(renderValue(item,depth+1))}shown=Math.min(shown+50,entries.length);more.hidden=shown===entries.length;more.textContent=`Show more values (${shown} of ${entries.length})`}
    more.onclick=append;append();detail.append(values,more);
   });return detail;
  }
  const display=value===''?'Empty string':value===null?'Null':String(value);
  if(display.length>240){const detail=node('details',undefined,'attribute-long');detail.append(node('summary',display.slice(0,160)+'…'),node('span',display,'attribute-value-text'));return detail}
  return node('span',display,'attribute-value-text'+(value===''||value===null?' attribute-empty':''));
 }
 // Validate JSON values without invoking getters, toJSON hooks or lossy coercion.
 function isJSONValue(value){
  const ancestors=new Set(),pending=[{value}];
  try{while(pending.length){const item=pending.pop(),v=item.value;if(item.exit){ancestors.delete(v);continue}if(v===null||typeof v==='string'||typeof v==='boolean')continue;if(typeof v==='number'){if(!Number.isFinite(v))return false;continue}if(typeof v!=='object'||ancestors.has(v))return false;const array=Array.isArray(v),proto=Object.getPrototypeOf(v);if(array?proto!==Array.prototype:proto!==Object.prototype&&proto!==null)return false;ancestors.add(v);pending.push({value:v,exit:true});const keys=Reflect.ownKeys(v);if(array&&keys.length!==v.length+1)return false;for(const key of keys){if(array&&key==='length')continue;if(typeof key!=='string')return false;const descriptor=Object.getOwnPropertyDescriptor(v,key);if(!descriptor.enumerable||!Object.hasOwn(descriptor,'value'))return false;if(array&&(!/^(0|[1-9][0-9]*)$/.test(key)||Number(key)>=v.length))return false;pending.push({value:descriptor.value})}}
   // Reject structures deeper than the platform can safely serialize as raw JSON.
   JSON.stringify(value);return true;
  }catch{return false}
 }
 function fieldModel(value,state='present'){
  if(state!=='present')return {state:['missing','unknown','unavailable'].includes(state)?state:'unknown',kind:null,label:state==='missing'?'Missing value':state==='unavailable'?'Value unavailable':'Value unknown'};
  if(value===undefined)return {state:'missing',kind:null,label:'Missing value'};
  if(!isJSONValue(value))return {state:'unavailable',kind:null,label:'Value unavailable'};
  return {state:'present',kind:type(value),label:value===''?'Empty string':value===null?'Null':typeof value==='object'?null:String(value)};
 }
 function valueNode(value){const element=node('goo-field-value');element.value=value;return element}
 if(global.customElements&&global.HTMLElement){
  class GooFieldValue extends global.HTMLElement{
   static get observedAttributes(){return ['value','state']}
   set value(value){this._invalid=value!==undefined&&!isJSONValue(value);try{this._value=this._invalid?undefined:structuredClone(value)}catch{this._invalid=true;this._value=undefined}this._hasValue=true;if(this.isConnected)this.render()}
   get value(){return structuredClone(this._hasValue?this._value:this.attributeValue().value)}
   get state(){const declared=this.getAttribute('state');if(declared&&declared!=='present')return ['missing','unknown','unavailable'].includes(declared)?declared:'unknown';const input=this._hasValue?{value:this._value,invalid:this._invalid}:this.attributeValue();return input.invalid?'unavailable':input.value===undefined?'missing':'present'}set state(value){this.setAttribute('state',value)}
   attributeValue(){const json=this.getAttribute('value');if(json===null)return {value:undefined};if(json.length>65536)return {invalid:true};try{const value=JSON.parse(json);return isJSONValue(value)?{value}:{invalid:true}}catch{return {invalid:true}}}
   connectedCallback(){for(const key of ['value','state'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
   attributeChangedCallback(name){if(name==='value'){this._hasValue=false;this._value=undefined;this._invalid=false}if(this.isConnected)this.render()}
   render(){const input=this._hasValue?{value:this._value,invalid:this._invalid}:this.attributeValue(),model=fieldModel(input.value,input.invalid?'unavailable':this.state);this.replaceChildren();this.dataset.valueState=model.state;if(model.state!=='present'){this.append(node('span',model.label,'attribute-value-text attribute-empty'));return}this.append(renderValue(input.value),node('span',model.kind,'attribute-type'));
    if(model.kind==='object'||model.kind==='array'){const raw=node('details',undefined,'raw-value');raw.append(node('summary','Raw value'));let built=false;raw.addEventListener('toggle',()=>{if(raw.open&&!built){built=true;raw.append(node('pre',JSON.stringify(input.value,null,2)))}});this.append(raw)}
   }
  }
  // Caller owns fields, search, scopes and actions; this shell never maps records or replaces children.
  class GooAttributeList extends global.HTMLElement{}
  if(!global.customElements.get('goo-field-value'))global.customElements.define('goo-field-value',GooFieldValue);
  if(!global.customElements.get('goo-attribute-list'))global.customElements.define('goo-attribute-list',GooAttributeList);
 }
 function mountModel(container,record,model,{onFilter,onBreakdown,onFacet,canRefine=()=>true,raw=true,facetHint,renderField}={}){
  const total=model.reduce((n,g)=>n+g.entries.length,0),root=node('goo-attribute-list',undefined,'attribute-viewer');
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
     const h=entry.field&&presentation()?.hint(entry.field);if(h){key.prepend(node('span',h.label+(h.legacy?' (legacy)':''),'field-hint-label'));key.title=h.description;row.setAttribute('aria-description',h.description);row.setAttribute('aria-label',group.name+': '+entry.key+' · '+h.label+(h.legacy?' (legacy)':''))}
     value.append(renderField?.(entry)||valueNode(entry.value));
     const copyButton=node('button','Copy');copyButton.type='button';copyButton.setAttribute('aria-label','Copy value for '+entry.key);copyButton.onclick=()=>copy(entry.value,entry.key);actions.append(copyButton);
     if(selectable(entry)&&onFacet)for(const mode of ['include','exclude']){const action=node('button',mode==='include'?'Include':'Exclude');action.type='button';action.disabled=!canRefine();action.setAttribute('aria-label',(mode==='include'?'Include ':'Exclude ')+entry.field.kind+':'+entry.field.key+' from record');action.onclick=()=>{if(!canRefine())return;onFacet(structuredClone(entry.field),entry.value,mode);status.textContent=facetHint||'Staged '+mode+' for '+entry.field.kind+':'+entry.field.key+'. Apply or Open in the selection tray; displayed measurements are unchanged.'};actions.append(action)}
     if(entry.field&&typeof entry.value==='string'&&onFilter){const filter=node('button','Filter and run');filter.type='button';filter.disabled=!canRefine();filter.title=filter.disabled?'Choose All filters to narrow this question.':'Read a fresh result with this exact value in the fixed window';filter.setAttribute('aria-label','Filter and run by '+entry.field.kind+':'+entry.field.key);filter.onclick=()=>onFilter(entry.field,entry.value);actions.append(filter)}
     if(entry.field&&onBreakdown){const breakdown=node('button','Break down and run');breakdown.type='button';if(h)breakdown.title='Group by '+h.label+' · '+entry.field.kind+':'+entry.field.key;breakdown.setAttribute('aria-label','Break down and run by '+entry.field.kind+':'+entry.field.key);breakdown.onclick=()=>onBreakdown(entry.field);actions.append(breakdown)}
     row.append(key,value,actions);section.append(row);
    }content.append(section);
   }count.textContent=`${found} of ${total} fields`;if(!found)content.append(node('p','No attributes match this search.','attribute-no-match'));
  }
  search.oninput=render;scope.onchange=render;render();
  if(raw){const detail=node('details',undefined,'raw-record');detail.append(node('summary','Raw record'),node('pre',JSON.stringify(record,null,2)));root.append(detail)}
  return root;
 }
 function mount(container,record,options={}){return mountModel(container,record,groups(record),options)}
 function mountShared(container,records,options={}){return mountModel(container,null,commonGroups(records),{...options,raw:false})}
 // Native identity stays separate from record time and measurement values.
 function streamRecord(stream){const s=structuredClone(stream);return {name:s.name,unit:s.unit,ResourceAttributes:s.resource,Attributes:s.attributes,scope:s.scope,version:s.version,schema:s.schemaURL,resourceSchema:s.resourceSchemaURL}}
 function mountStream(container,stream){return mount(container,streamRecord(stream))}
 // Application adapter for an already supplied native stream; no source client.
 function streamInspection(stream){
  const snapshot=structuredClone(stream),root=node('div',undefined,'stream-inspection'),button=node('button','Inspect stream'),panel=node('section'),inspector=global.GooUI.inspector(panel,{onClose:()=>button.setAttribute('aria-expanded','false')});
  button.type='button';button.setAttribute('aria-expanded','false');
  // The shared shell owns close-to-opener focus and local body/scroll retention.
  button.onclick=()=>{if(!inspector.has('stream')){const content=node('div');content.append(node('p','Supplied stream identity only; inspection does not reread telemetry.','evidence-note'));mountStream(content,snapshot);inspector.show({key:'stream',title:'Native stream identity',description:snapshot.name+' ['+snapshot.unit+']',content,opener:button})}else inspector.show({key:'stream',title:'Native stream identity',opener:button});button.setAttribute('aria-expanded','true')};root.append(button,panel);return root;
 }
 function disclosure(record,options={}){
  const detail=node('details',undefined,'record-attributes'),host=node('div');detail.append(node('summary','Captured attributes'),host);
  // Keep raw evidence available without constructing thousands of primary-view
  // nodes for unopened records on a 100-row page.
  const raw=node('details',undefined,'raw-record');raw.append(node('summary','Raw record'),node('pre',JSON.stringify(record,null,2)));detail.append(raw);
  let built=false;detail.addEventListener('toggle',()=>{if(detail.open&&!built){built=true;mount(host,record,{...options,raw:false})}});return detail;
 }
 const api={suggestionData,groups,commonGroups,selectable,type,text,matches,fieldModel,isJSONValue,mount,mountShared,streamRecord,mountStream,streamInspection,disclosure};if(typeof module!=='undefined')module.exports=api;else global.GooAttributes=api;
})(typeof window==='undefined'?globalThis:window);
