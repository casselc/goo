/* Native retention transport, read lower bounds and exact retry intents. */
(function(global){
 const key=()=>Array.from(crypto.getRandomValues(new Uint8Array(32)),v=>v.toString(16).padStart(2,'0')).join('');
 const bases=new Map();function readToken(id){try{return bases.get(id)||sessionStorage.getItem('goo.evidence.basis.'+id)||''}catch{return bases.get(id)||''}}function remember(id,token){if(typeof token!=='string'||token.length>16384||!token)return;bases.set(id,token);try{sessionStorage.setItem('goo.evidence.basis.'+id,token)}catch{}}
 let journal;function pending(){if(!journal)journal=new global.GooPendingSaves.Journal(global.localStorage);return journal}
 function notify(){global.dispatchEvent(new CustomEvent('goo:pending-native-saves'))}
 async function send(path,body){const item=pending().begin(path,body);notify();const route=item.path==='/'?'':item.path;const r=await fetch('/api/retained-evidence'+route,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(item.body)});const v=await r.json();if(!r.ok)throw Error(v.error||'Native save unavailable');const id=v.manifest?.id||v.target?.id;if(id)remember(id,v.awaitToken);try{pending().complete(item.body.requestID)}catch{}notify();return v}
 async function api(path='',body,basis){if(body!==undefined)return send(path,body);const id=path.match(/^\/([a-f0-9]{64})/)?.[1],token=basis||(id?readToken(id):'');if(token)path+=(path.includes('?')?'&':'?')+'awaitToken='+encodeURIComponent(token);const r=await fetch('/api/retained-evidence'+path),v=await r.json();if(!r.ok)throw Error(v.error||'Evidence retention unavailable');return v}

 global.GooRetentionClient={key,readToken,remember,send,api,pending};
})(window);
