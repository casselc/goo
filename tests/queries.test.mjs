import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createServer} from 'node:http';
import {once} from 'node:events';
import {postJSON,getJSON,jsonResponse} from '../ui/queries.mjs';

const deferred=()=>{let resolve;const promise=new Promise(done=>resolve=done);return {promise,resolve}};

test('abort after headers cancels the matching query and rejects a late JSON body',async t=>{
  const body=deferred(),reading=deferred(),controller=new AbortController(),calls=[];
  t.mock.method(globalThis,'fetch',async(url,options)=>{
    calls.push({url,options});
    return {ok:true,status:200,json:()=>{reading.resolve();return body.promise}};
  });
  const pending=postJSON('/api/explore',{environment:'prod'},controller.signal);
  await reading.promise;
  controller.abort();
  assert.equal(calls.length,2);
  assert.equal(calls[1].url,'/api/queries/'+calls[0].options.headers['X-Goo-Query-ID']+'/cancel');
  assert.equal(calls[1].options.keepalive,true);
  assert.equal(calls[1].options.signal,undefined,'cancel must outlive the aborted source signal');
  const rejected=assert.rejects(pending,{name:'AbortError'});
  body.resolve({rows:['too late']});
  await rejected;
});

test('real streamed JSON is abortable after fetch resolves its headers',{timeout:5000},async t=>{
  const headers=deferred(),cancelled=deferred();let queryID;
  const server=createServer((request,response)=>{
    request.resume();
    if(request.url.startsWith('/api/queries/')){
      assert.equal(request.url,'/api/queries/'+queryID+'/cancel');
      response.writeHead(202,{'Content-Type':'application/json'}).end('{}');
      cancelled.resolve();return;
    }
    queryID=request.headers['x-goo-query-id'];
    response.writeHead(200,{'Content-Type':'application/json'});
    response.write('{"rows":'); // Deliberately leave the JSON body incomplete.
  });
  server.listen(0,'127.0.0.1');await once(server,'listening');
  t.after(async()=>{server.closeAllConnections();await new Promise(done=>server.close(done))});
  const originalFetch=globalThis.fetch,base='http://127.0.0.1:'+server.address().port;
  t.mock.method(globalThis,'fetch',async(url,options)=>{
    const response=await originalFetch(base+url,options);
    if(url==='/api/explore')headers.resolve();
    return response;
  });
  const controller=new AbortController(),pending=postJSON('/api/explore',{},controller.signal);
  const rejected=assert.rejects(pending,{name:'AbortError'});
  await headers.promise;controller.abort();
  await rejected;await cancelled.promise;
});

test('completed, refused and malformed bodies release cancellation ownership',async t=>{
  for(const mode of ['ok','refused','malformed','transport','serialization']){
    const controller=new AbortController(),calls=[];
    const remove=t.mock.method(controller.signal,'removeEventListener');
    t.mock.method(globalThis,'fetch',async(url,options)=>{
      calls.push({url,options});
      if(mode==='transport')throw Error('offline');
      return {ok:mode!=='refused',status:mode==='refused'?429:200,json:async()=>{
        if(mode==='malformed')throw new SyntaxError('invalid JSON');
        return mode==='refused'?{error:'busy'}:{count:'0'};
      }};
    });
    const body={};if(mode==='serialization')body.cycle=body;
    const pending=postJSON('/api/explore',body,controller.signal);
    if(['malformed','transport','serialization'].includes(mode))await assert.rejects(pending);
    else assert.deepEqual(await pending,{ok:mode==='ok',status:mode==='refused'?429:200,data:mode==='ok'?{count:'0'}:{error:'busy'}});
    assert.equal(remove.mock.callCount(),1,mode);
    const before=calls.length;controller.abort();assert.equal(calls.length,before,mode);
    t.mock.restoreAll();
  }
});

test('pre-cancelled reads never start a request',async t=>{
  const fetch=t.mock.method(globalThis,'fetch',()=>assert.fail('unexpected source or cancel request'));
  await assert.rejects(postJSON('/api/explore',{},AbortSignal.abort()),{name:'AbortError'});
  assert.equal(fetch.mock.callCount(),0);
});

test('cancel transport failure cannot replace the original abort',async t=>{
  for(const synchronous of [false,true]){
    const body=deferred(),reading=deferred(),controller=new AbortController();
    t.mock.method(globalThis,'fetch',url=>{
      if(url.includes('/cancel')){
        if(synchronous)throw Error('cancel offline');
        return Promise.reject(Error('cancel offline'));
      }
      return Promise.resolve({ok:true,status:200,json:()=>{reading.resolve();return body.promise}});
    });
    const pending=postJSON('/api/explore',{},controller.signal);
    await reading.promise;controller.abort();
    const rejected=assert.rejects(pending,{name:'AbortError'});body.resolve({});await rejected;
    t.mock.restoreAll();
  }
});

test('pure JSON decoding does not add a query capability or cancellation request',async()=>{
  assert.deepEqual(await jsonResponse(new Response('{"plan":true}',{status:200})),{ok:true,status:200,data:{plan:true}});
});

test('source GET cancellation owns the decoded body without a JSON request body',async t=>{
 const body=deferred(),reading=deferred(),controller=new AbortController(),calls=[];
 t.mock.method(globalThis,'fetch',async(url,options)=>{
  calls.push({url,options});
  return {ok:true,status:200,json:()=>{reading.resolve();return body.promise}};
 });
 const pending=getJSON('/api/operations?minutes=5',controller.signal);
 await reading.promise;
 assert.equal(calls[0].options.method,'GET');
 assert.equal(Object.hasOwn(calls[0].options,'body'),false);
 assert.equal(Object.hasOwn(calls[0].options.headers,'Content-Type'),false);
 assert.match(calls[0].options.headers['X-Goo-Query-ID'],/^[a-f0-9-]{36}$/);
 const rejected=assert.rejects(pending,{name:'AbortError'});controller.abort();body.resolve({late:true});await rejected;
 assert.equal(calls[1].url,'/api/queries/'+calls[0].options.headers['X-Goo-Query-ID']+'/cancel');
 assert.equal(calls[1].options.signal,undefined);
});

test('successful source GET releases cancellation and preserves refused status',async t=>{
 const controller=new AbortController();
 const fetch=t.mock.method(globalThis,'fetch',async()=>({ok:false,status:429,json:async()=>({error:'busy'})}));
 assert.deepEqual(await getJSON('/api/histograms',controller.signal),{ok:false,status:429,data:{error:'busy'}});
 controller.abort();assert.equal(fetch.mock.callCount(),1);
 await assert.rejects(getJSON('/api/operations',AbortSignal.abort()),{name:'AbortError'});
 assert.equal(fetch.mock.callCount(),1);
});
