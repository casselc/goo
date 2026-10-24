/* Supplied predicate editing only: no source reads, navigation or execution. */
(function(global){
 const columns=['service','trace_id','span_id','scope','scope_version','name','kind','status','duration_ns','timestamp_ns','duration_bucket','status_class','body','severity'];
 let nextID=0;
 const bytes=s=>new TextEncoder().encode(s).length;
 function validate(where,signal='logs'){
  if(!['logs','spans'].includes(signal))throw Error('Choose logs or spans.');
  let count=0;
  function visit(p,depth){
   if(++count>32||depth>4)throw Error('Use at most 32 conditions/groups and 4 nested levels.');
   if(!p||typeof p!=='object')throw Error('Choose a condition.');
   if(p.all||p.any){const keys=Object.keys(p),list=p.all||p.any;if(keys.length!==1||!Array.isArray(list)||!list.length)throw Error('Each group needs a condition.');list.forEach(x=>visit(x,depth+1));return}
   const f=p.field;if(!f||!['column','attribute','resource'].includes(f.kind)||typeof f.key!=='string'||!f.key||bytes(f.key)>128||f.key.includes('\0'))throw Error('Choose a qualified field (up to 128 bytes).');
   if(f.kind==='column'&&(!columns.includes(f.key)||(signal==='logs'&&['name','kind','status','duration_ns','duration_bucket','status_class'].includes(f.key))||(signal==='spans'&&['body','severity'].includes(f.key))))throw Error('Field is unavailable for '+signal+'.');
   const numeric=f.kind==='column'&&['duration_ns','timestamp_ns'].includes(f.key),v=p.value??'';
   if(typeof v!=='string'||bytes(v)>1024||v.includes('\0'))throw Error('Use a value up to 1024 bytes.');
   if(!['eq','ne','contains','gte','lte','exists','missing'].includes(p.op)||numeric&&p.op==='contains'||!numeric&&['gte','lte'].includes(p.op))throw Error('Operator does not match the field type.');
   if(['exists','missing'].includes(p.op)){if(v!=='')throw Error('Presence checks have no value.');return}
   if(numeric&&(!/^[0-9]+$/.test(v)||BigInt(v)>18446744073709551615n))throw Error('Use an unsigned 64-bit decimal integer.');
  }
  if(where)visit(where,0);return structuredClone(where);
 }
 // Match Go omitempty without coercing present-empty to a missing predicate.
 function normalize(p){if(!p)return undefined;if(p.all||p.any){const key=p.all?'all':'any';return {[key]:p[key].map(normalize)}}return {field:structuredClone(p.field),op:p.op,...(p.value?{value:p.value}:{})}}
 function mount(host,{where,signal='logs',onChange=()=>{}}={}){
  const n=global.GooUI.node,namespace='filter-fields-'+(++nextID),root=n('div',undefined,'filter-editor'),body=n('div'),status=n('p'),actions=n('div',undefined,'controls'),add=n('button','Add condition'),group=n('button','Add group'),clear=n('button','Clear filters');host.append(root);status.setAttribute('role','status');root.append(body,actions,status);actions.append(add,group,clear);let tree=structuredClone(where),currentSignal=signal,suppliedFields=[],suggestionsRefresh=[];
  const leaf=()=>({field:{kind:'column',key:'service'},op:'eq',value:''});
  function emit(){try{validate(tree,currentSignal);status.textContent='Changes are staged; apply or run explicitly.';onChange(normalize(tree),null)}catch(e){status.textContent=e.message;onChange(undefined,e.message)}}
  function options(select,items,value,label){select.setAttribute('aria-label',label);for(const [v,t]of items){const o=n('option',t);o.value=v;select.append(o)}select.value=value}
  function append(item){if(!tree)tree=item;else if(tree.all)tree.all.push(item);else tree={all:[tree,item]};render();emit()}
  add.onclick=()=>append(leaf());group.onclick=()=>append({any:[leaf()]});clear.onclick=()=>{tree=undefined;render();emit()};
  function render(){suggestionsRefresh=[];body.replaceChildren();clear.disabled=!tree;let serial=0;
   function row(p,replace,remove,depth){const id=++serial,box=n('fieldset'),legend=n('legend',(p.all||p.any?'Group ':'Condition ')+id);box.append(legend);body.dataset.nodes=String(serial);
    if(p.all||p.any){const key=p.all?'all':'any',pick=n('select');options(pick,[['all','Match all (AND)'],['any','Match any (OR)']],key,'Group '+id+' match');box.append(pick);pick.onchange=()=>{const next={[pick.value]:p[key]};replace(next);render();body.querySelector('[aria-label="Group '+id+' match"]')?.focus();emit()};p[key].forEach((child,i)=>box.append(row(child,next=>p[key][i]=next,()=>{p[key].splice(i,1);if(!p[key].length)remove()},depth+1)));const a=n('button','Add condition to group '+id),g=n('button','Add nested group '+id);a.onclick=()=>{p[key].push(leaf());render();emit()};g.disabled=depth>=3;g.onclick=()=>{p[key].push({any:[leaf()]});render();emit()};box.append(a,g)
    }else{const kind=n('select'),field=n('input'),op=n('select'),value=n('input'),hint=n('small');options(kind,[['column','Column'],['attribute','Attribute'],['resource','Resource']],p.field.kind,'Condition '+id+' namespace');field.value=p.field.key;field.setAttribute('aria-label','Condition '+id+' field');field.placeholder='Field key';field.maxLength=128;const suggestions=n('datalist');suggestions.id=namespace+'-'+id;field.setAttribute('list',suggestions.id);function suggest(){suggestions.replaceChildren();const keys=kind.value==='column'?columns:suppliedFields.filter(f=>f.kind===kind.value).map(f=>f.key);for(const key of [...new Set(keys)]){try{validate({field:{kind:kind.value,key},op:'exists'},currentSignal);const option=n('option');option.value=key;suggestions.append(option)}catch{}}}suggestionsRefresh.push(suggest);suggest();options(op,[['eq','equals'],['ne','does not equal'],['contains','contains'],['gte','≥'],['lte','≤'],['exists','is present'],['missing','is missing']],p.op,'Condition '+id+' operator');value.value=p.value||'';value.setAttribute('aria-label','Condition '+id+' value');value.maxLength=1024;value.placeholder='Empty string is a value';
     function change(){p.field={kind:kind.value,key:field.value};suggest();p.op=op.value;const presence=['exists','missing'].includes(p.op);value.disabled=value.hidden=presence;p.value=presence?'':value.value;hint.textContent=presence?'Checks key presence.':p.field.kind==='column'&&['duration_ns','timestamp_ns'].includes(p.field.key)?'Unsigned integer, nanoseconds.':'Blank value means present empty string.';emit()}
     value.disabled=value.hidden=['exists','missing'].includes(p.op);kind.onchange=op.onchange=change;field.oninput=value.oninput=change;hint.textContent=value.disabled?'Checks key presence.':'Blank value means present empty string.';box.append(kind,field,suggestions,op,value,hint)
    }
    const del=n('button','Remove condition '+id);del.onclick=()=>{remove();render();emit();(body.querySelector('input,select,button')||add).focus()};box.append(del);return box
   }
   if(tree)body.append(row(tree,next=>tree=next,()=>tree=undefined,0));else body.append(n('small','All records in scope.'));body.querySelectorAll('button').forEach(b=>b.type='button');
  }
  root.querySelectorAll('button').forEach(b=>b.type='button');render();return {setFields(fields=[]){suppliedFields=structuredClone(fields);suggestionsRefresh.forEach(refresh=>refresh())},set(where,signal=currentSignal){tree=structuredClone(where);currentSignal=signal;status.textContent='';render()},snapshot:()=>structuredClone(tree)};
 }
 const api={validate,normalize,mount};if(typeof module!=='undefined')module.exports=api;else global.GooFilterEditor=api;
})(typeof globalThis!=='undefined'?globalThis:this);
