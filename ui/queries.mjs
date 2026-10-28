/* Source-read cancellation covers headers and JSON consumption. IDs are local
 * execution capabilities, not authentication or saved questions. */
// Decode once; owners retain HTTP error messages and scope validation.
async function jsonResponse(pending){
  const response=await pending;
  return {ok:response.ok,status:response.status,data:await response.json()};
}
async function postJSON(url,body,signal){
  signal?.throwIfAborted();
  const id=crypto.randomUUID();
  const cancel=()=>{
    try{
      fetch('/api/queries/'+id+'/cancel',{method:'POST',headers:{'Content-Type':'application/json'},body:'{}',keepalive:true})
        .then(response=>response.body?.cancel()).catch(()=>{});
    }catch{/* Cancellation is best effort; the original read still aborts. */}
  };
  signal?.addEventListener('abort',cancel,{once:true});
  try{
    const reply=await jsonResponse(fetch(url,{method:'POST',headers:{'Content-Type':'application/json','X-Goo-Query-ID':id},body:JSON.stringify(body),signal}));
    // Also reject transports/fixtures which complete after their signal aborts.
    signal?.throwIfAborted();
    return reply;
  }finally{signal?.removeEventListener('abort',cancel)}
}
export {postJSON,jsonResponse};
