// Scoped fault boundary for native journeys. Successful SQL is forwarded unchanged
// to the real OTLP-backed source; no successful telemetry or UI response is faked.
const http=require('node:http');
const {once}=require('node:events');
async function startSourceFaultProxy(upstream='http://127.0.0.1:8123'){
 const calls=[],sockets=new Set();let armed=null;
 const server=http.createServer(async(req,res)=>{
  const chunks=[];for await(const chunk of req)chunks.push(chunk);const body=Buffer.concat(chunks),url=new URL(req.url,'http://fixture.invalid');
  const call={index:calls.length,sql:body.toString(),parameters:Object.fromEntries(url.searchParams),fault:null,status:null,closed:false};calls.push(call);
  res.on('close',()=>{call.closed=true});
  if(armed&&armed.matches(call)){
   const fault=armed;armed=null;call.fault=fault.kind;fault.startedResolve(call);
   if(fault.kind==='hold')await fault.releasePromise;
   if(!res.destroyed){const status=fault.status||503;call.status=status;res.writeHead(status,{'content-type':'text/plain'});res.end(fault.body||'Planted upstream unavailable');}
   return;
  }
  const target=new URL(req.url,upstream),forward=http.request(target,{method:req.method,headers:{...req.headers,host:target.host}},response=>{
   call.status=response.statusCode;res.writeHead(response.statusCode,response.headers);response.pipe(res);
  });
  forward.on('error',()=>{if(!res.destroyed){call.status=502;res.writeHead(502);res.end('Scoped upstream forwarding failed')}});
  res.on('close',()=>{if(!res.writableEnded)forward.destroy()});forward.end(body);
 });
 server.on('connection',socket=>{sockets.add(socket);socket.on('close',()=>sockets.delete(socket))});server.listen(0,'127.0.0.1');await once(server,'listening');
 return{url:'http://127.0.0.1:'+server.address().port,calls,
  arm({kind='unavailable',matches,status,body}){if(armed)throw Error('Unconsumed source fault');let startedResolve,release;const started=new Promise(resolve=>startedResolve=resolve),releasePromise=new Promise(resolve=>release=resolve);armed={kind,matches,status,body,startedResolve,releasePromise};return{started,release}},
  async stop(){for(const socket of sockets)socket.destroy();await new Promise(resolve=>server.close(resolve))}
 };
}
module.exports={startSourceFaultProxy};
