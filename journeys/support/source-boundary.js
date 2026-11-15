const http=require('node:http');
const {once}=require('node:events');
const {performance}=require('node:perf_hooks');

// Forward real source bytes. The only invented response is a deliberate refusal.
// Held successes are still read from ClickHouse after the browser cancels its lane.
async function sourceBoundary(upstream){
 const calls=[],sockets=new Set();let armed=null;
 const server=http.createServer(async(req,res)=>{
  const began=performance.now();const chunks=[];for await(const chunk of req)chunks.push(chunk);const body=Buffer.concat(chunks),sql=body.toString(),call={sql,index:calls.length,status:null,startedMs:began,family:/\botel_logs\b/.test(sql)?'logs':/\botel_traces\b/.test(sql)?'spans':'other'};calls.push(call);
  res.on('close',()=>{if(!res.writableEnded)call.downstreamClosed=true});
  const fault=armed&&armed.matches(sql)?armed:null;if(fault){armed=null;call.fault=fault.kind;fault.startedResolve(call);if(fault.kind==='hold')await fault.releasePromise;if(fault.kind==='refuse'){call.status=403;call.elapsedMs=performance.now()-began;res.writeHead(403,{'content-type':'text/plain'});res.end('Planted fixture access refusal');return}}
  try{const response=await fetch(new URL(req.url,upstream),{method:req.method,headers:{Authorization:req.headers.authorization||'','Content-Type':'text/plain'},body}),bytes=Buffer.from(await response.arrayBuffer());call.status=response.status;call.bytes=bytes.length;call.elapsedMs=performance.now()-began;if(!res.destroyed){res.writeHead(response.status,{'Content-Type':response.headers.get('content-type')||'text/plain'});res.end(bytes)}}catch(error){call.error=error.message;call.elapsedMs=performance.now()-began;if(!res.destroyed){res.writeHead(502);res.end('Fixture forwarding unavailable')}}
 });server.on('connection',socket=>{sockets.add(socket);socket.on('close',()=>sockets.delete(socket))});server.listen(0,'127.0.0.1');await once(server,'listening');
 return {url:'http://127.0.0.1:'+server.address().port,calls,arm(kind,family){if(armed)throw Error('Unconsumed source fault');let startedResolve,release;const started=new Promise(resolve=>startedResolve=resolve),releasePromise=new Promise(resolve=>release=resolve);armed={kind,matches:sql=>new RegExp('\\botel_'+(family==='logs'?'logs':'traces')+'\\b').test(sql),startedResolve,releasePromise};return {started,release}},async stop(){for(const socket of sockets)socket.destroy();await new Promise(resolve=>server.close(resolve))}};
}

module.exports={sourceBoundary};
