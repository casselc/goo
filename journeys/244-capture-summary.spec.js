const {test,expect}=require('@playwright/test'),{randomBytes,createHash}=require('node:crypto');
const {execFile}=require('node:child_process'),{promisify}=require('node:util'),fs=require('node:fs/promises'),path=require('node:path'),os=require('node:os');
const {startRuntime}=require('./191-generic-runtime');
test('244 large accepted capture survives summary search and explicit index migration',async({page,request})=>{
 test.setTimeout(120000);const owned=await fs.mkdtemp(path.join(os.tmpdir(),'goo-summary-244-')),dir=path.resolve('artifacts/capture-summary',randomBytes(4).toString('hex'));let runtime;
 try{
  const binary=process.env.GOO_GENERIC_BINARY||path.join(owned,'goo');if(!process.env.GOO_GENERIC_BINARY)await promisify(execFile)('go',['build','-mod=readonly','-o',binary,'./cmd/goo'],{timeout:120000});runtime=await startRuntime(binary,dir);
  const environment='summary-'+randomBytes(4).toString('hex'),query={signal:'logs',environment,fromMs:Date.now()-60000,toMs:Date.now(),limit:100};
  const save=async(title,q)=>{const r=await request.post(runtime.base+'/api/captures',{data:{title,question:{kind:'explore',explore:q}}});expect(r.status()).toBe(201);return r.json()};
  const first=await save('Earlier evidence',query),large=await save('Large <雪> capture',{...query,where:{all:Array.from({length:31},()=>({field:{kind:'column',key:'service'},op:'eq',value:'<'.repeat(340)}))}});
  const file=path.join(runtime.root,'captures','exploration',large.receipt.id+'.json'),before=await fs.readFile(file),sha=b=>createHash('sha256').update(b).digest('hex');
  expect(Buffer.byteLength(JSON.stringify(large.receipt).replaceAll('<','\\u003c'))).toBeGreaterThan(65536);
  let sourceRequests=0;const errors=[];page.on('pageerror',e=>errors.push(e.message));page.on('request',r=>{if(r.method()==='POST')sourceRequests++});
  await page.goto(runtime.base+'/#investigations');await page.getByRole('button',{name:'Browse evidence',exact:true}).click();const rows=page.locator('#captureCatalog tbody tr');await expect(rows).toHaveCount(2);await expect(rows.filter({hasText:'Large <雪> capture'})).toContainText('logs');
  const search=await(await request.get(runtime.base+'/api/captures/search')).json();expect(search.version).toBe('goo/capture-search/v2');expect(search.items.map(s=>s.id).sort()).toEqual([first.receipt.id,large.receipt.id].sort());expect(search.items.every(s=>!('question'in s))).toBe(true);
  await rows.getByRole('button',{name:'Large <雪> capture',exact:true}).click();await expect(page.locator('#captureSavedTitle')).toHaveText('Large <雪> capture');expect((await(await request.get(runtime.base+'/api/captures/'+large.receipt.id)).json()).receipt).toEqual(large.receipt);
  const root=path.dirname(file);await fs.rm(path.join(root,'.catalog-v2'),{recursive:true});await fs.mkdir(path.join(root,'.catalog-v1'));await fs.writeFile(path.join(root,'.catalog-v1','legacy'),'unchanged derived v1');
  expect((await request.get(runtime.base+'/api/captures/search')).status()).toBe(503);expect((await request.get(runtime.base+'/api/captures/'+large.receipt.id)).status()).toBe(200);
  const rebuilt=await promisify(execFile)(binary,['rebuild-capture-catalog','-dir',root]);expect(JSON.parse(rebuilt.stdout).indexed).toBe(2);expect(sha(await fs.readFile(file))).toBe(sha(before));expect(await fs.readFile(path.join(root,'.catalog-v1','legacy'),'utf8')).toBe('unchanged derived v1');
  await page.getByRole('button',{name:'Browse evidence',exact:true}).click();await expect(rows).toHaveCount(2);await page.locator('.capture-search').getByLabel('Capture title',{exact:true}).fill('Large <雪>');await page.getByRole('button',{name:'Find captures',exact:true}).click();await expect(rows).toHaveCount(1);
  await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.screenshot({path:'docs/journeys/screenshots/244-capture-summary-mobile.png'});expect(sourceRequests).toBe(0);expect(errors).toEqual([]);
 }finally{await runtime?.stop();await fs.rm(owned,{recursive:true,force:true})}
});
