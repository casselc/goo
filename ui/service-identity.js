/* Missing, present-empty and wildcard namespace are different selectors. */
(function(global){
 function namespace(value){return value===undefined||value===null?{mode:'missing'}:{mode:'value',...(value!==''?{value}: {})}}
 function key(service,ns){return JSON.stringify([service,ns?.mode||'any',ns?.value||''])}
 function same(a,b){return key('',a)===key('',b)}
 function label(ns,{lowercase=false}={}){const prefix=lowercase?'namespace':'Namespace';return !ns?(lowercase?'all namespaces':'All namespaces'):ns.mode==='missing'?prefix+' missing':ns.value?prefix+' '+JSON.stringify(ns.value):prefix+' present empty'}
 function name(service,ns){return !ns||ns.mode==='missing'?service:(ns.value||'(empty namespace)')+' / '+service}
 function predicate(ns){if(!ns)return null;return {field:{kind:'resource',key:'service.namespace'},op:ns.mode==='missing'?'missing':'eq',...(ns.mode==='missing'?{}:{value:ns.value||''})}}
 function serviceWhere(service,ns){const p={field:{kind:'column',key:'service'},op:'eq',value:service},n=predicate(ns);return n?{all:[p,n]}:p}
 function and(where,...leaves){return {all:[...(where?.all||[where]).filter(Boolean),...leaves]}}
 const api={namespace,key,same,label,name,predicate,serviceWhere,and};if(typeof module!=='undefined')module.exports=api;else global.GooServiceIdentity=api;
})(typeof window==='undefined'?globalThis:window);
