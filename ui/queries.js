/* IDs scope local cancellation; they are not authentication or saved questions. */
(function(global){
 async function post(url,body,signal){
  const id=crypto.randomUUID();const cancel=()=>{fetch('/api/queries/'+id+'/cancel',{method:'POST',headers:{'Content-Type':'application/json'},body:'{}',keepalive:true}).catch(()=>{})};
  if(signal?.aborted)throw new DOMException('Query cancelled','AbortError');
  signal?.addEventListener('abort',cancel,{once:true});
  try{return await fetch(url,{method:'POST',headers:{'Content-Type':'application/json','X-Goo-Query-ID':id},body:JSON.stringify(body),signal})}
  finally{signal?.removeEventListener('abort',cancel)}
 }
 global.GooQueries={post};
})(window);
