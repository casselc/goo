const {test,expect}=require('@playwright/test');
const {createServer}=require('node:http');
const {once}=require('node:events');
const {readFile}=require('node:fs/promises');

// Transport fixture, not a simulated observability result. Hold real HTTP JSON
// after headers so Chromium exercises the production helper's body lifetime.
test('source cancellation stays active after Chromium receives response headers',async({page})=>{
  const source=await readFile(require.resolve('../ui/queries.mjs'));
  const reads=[],cancellations=[];
  const server=createServer(async(request,response)=>{
    if(request.url==='/queries.mjs'){
      response.writeHead(200,{'Content-Type':'application/javascript'}).end(source);return;
    }
    if(request.url==='/'){
      response.writeHead(200,{'Content-Type':'text/html'}).end('<!doctype html><title>Query transport fixture</title>');return;
    }
    let body='';for await(const chunk of request)body+=chunk;
    if(request.url.startsWith('/api/queries/')){
      cancellations.push({path:request.url,body});
      response.writeHead(202,{'Content-Type':'application/json'}).end('{}');return;
    }
    if(request.url==='/api/explore'){
      reads.push({id:request.headers['x-goo-query-id'],body});
      response.writeHead(200,{'Content-Type':'application/json'});
      response.write('{"rows":');return;
    }
    response.writeHead(404).end();
  });
  server.listen(0,'127.0.0.1');await once(server,'listening');
  try{
    await page.goto('http://127.0.0.1:'+server.address().port);
    await page.evaluate(async()=>{
      const {postJSON}=await import('/queries.mjs'),original=window.fetch;
      window.fetch=async(...args)=>{
        const response=await original(...args);
        if(args[0]==='/api/explore')window.sourceHeadersReceived=true;
        return response;
      };
      window.sourceController=new AbortController();
      window.sourceOutcome='pending';
      window.sourceRead=postJSON('/api/explore',{environment:'transport-fixture'},window.sourceController.signal)
        .then(()=>window.sourceOutcome='published',error=>window.sourceOutcome=error.name);
    });
    await page.waitForFunction(()=>window.sourceHeadersReceived);
    expect(await page.evaluate(()=>window.sourceOutcome)).toBe('pending');
    await page.evaluate(async()=>{sourceController.abort();await sourceRead});
    expect(await page.evaluate(()=>window.sourceOutcome)).toBe('AbortError');
    await expect.poll(()=>cancellations.length).toBe(1);
    expect(reads).toHaveLength(1);
    expect(cancellations).toEqual([{path:'/api/queries/'+reads[0].id+'/cancel',body:'{}'}]);
  }finally{
    await page.close();server.closeAllConnections();await new Promise(done=>server.close(done));
  }
});
