const {test, expect} = require('@playwright/test');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const F = require('./support/http-server410');
const {startRuntime} = require('./191-generic-runtime');
const {source, startBombadil} = require('./support/cross-app');

test('HTTP route and method rows preserve joint populations through shared drilldowns', async ({page, request}) => {
  test.setTimeout(150000);
  page.setDefaultTimeout(15000);
  const dir = await fs.mkdtemp(path.join(os.tmpdir(),'goo-http-route-method-'));
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, dir);
  const f = F.fixture(), errors = [], calls = [];
  const payload = F.traces(f);
  for (const [index, resource] of payload.resourceSpans.entries()) {
    const attributes = resource.scopeSpans[0].spans[0].attributes;
    if (index < 4) attributes.find(a => a.key==='http.request.method').value.stringValue = index % 2 ? 'POST' : 'GET';
    if (index === 4) attributes.splice(attributes.findIndex(a => a.key==='http.route'),1);
    if (index === 5) attributes.find(a => a.key==='http.route').value.stringValue = '';
  }
  try {
    expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:payload})).ok()).toBe(true);
    const raw = () => source(request,"SELECT ServiceName,SpanKind,SpanId,ResourceAttributes,SpanAttributes,toString(toUnixTimestamp64Nano(Timestamp)) AS timestamp_ns FROM otel_traces WHERE SpanAttributes['bank410']={bank:String} FORMAT JSONEachRow",{param_bank:f.token});
    await expect.poll(async () => (await raw()).length,{timeout:45000}).toBe(f.rows.length);
    const rows = await raw();
    const made = await request.post(runtime.base+'/api/captures',{data:{title:'Route and method',question:{kind:'inventory',inventory:f.scope}}});
    expect(made.status()).toBe(201);
    const capture = await made.json();
    page.on('pageerror', error => errors.push(error.message));
    page.on('request', req => { if (new URL(req.url()).pathname.startsWith('/api/')) calls.push(req); });
    const reads = () => calls.filter(req => req.method()==='POST' && ['/api/discovery','/api/inventory','/api/metric-discovery','/api/explore','/api/record-aggregates','/api/measurements'].includes(new URL(req.url()).pathname)).length;
    const response = pathname => page.waitForResponse(r => r.request().method()==='POST' && new URL(r.url()).pathname===pathname);
    await page.goto(runtime.base+'/#capture='+capture.receipt.id);
    let pending = response('/api/discovery');
    await page.locator('#captureGeneralLive').click();
    await pending;
    pending = response('/api/dashboards/plan');
    await page.getByRole('button',{name:'Open service '+f.namespace+' / '+f.service,exact:true}).click();
    const plan = await (await pending).json();
    const widget = plan.widgets.find(w => w.key==='dimension');
    expect(widget.query.groupBy.map(f => f.key)).toEqual(['http.route','http.request.method']);
    const dashboard = page.locator('.service-dashboard'), card = dashboard.locator('[data-widget=dimension]');
    await expect(dashboard.locator('[aria-busy=true]')).toHaveCount(0);
    await expect(card.getByRole('columnheader',{name:'Route',exact:true})).toBeVisible();
    await expect(card.getByRole('columnheader',{name:'Method',exact:true})).toBeVisible();
    const groups = new Map();
    for (const row of F.members(rows,widget.query)) {
      const tuple = JSON.stringify([row.SpanAttributes['http.route']??null,row.SpanAttributes['http.request.method']]);
      groups.set(tuple,(groups.get(tuple)||0)+1);
    }
    expect(groups.get(JSON.stringify(['/work','GET']))).toBe(5);
    expect(groups.get(JSON.stringify(['/work','POST']))).toBe(2);
    expect(groups.get(JSON.stringify([null,'GET']))).toBe(1);
    expect(groups.get(JSON.stringify(['','GET']))).toBe(1);
    for (const [tuple,count] of groups) {
      const target = card.locator('tr[data-diagnosis-value='+JSON.stringify(tuple)+']');
      await expect(target).toHaveCount(1);
      expect((await target.locator('td').nth(2).textContent()).trim()).toBe(String(count));
    }
    for (const tuple of [['/work','POST'],[null,'GET'],['','GET']]) {
      const before = reads();
      await card.locator('tr[data-diagnosis-value='+JSON.stringify(JSON.stringify(tuple))+']').getByRole('link',{name:'Open route and method',exact:true}).click();
      await expect(page.locator('#exploreRun')).toBeVisible();
      expect(reads()).toBe(before);
      pending = response('/api/explore');
      await page.locator('#exploreRun').click();
      const reply = await pending, q = reply.request().postDataJSON(), result = await reply.json();
      expect(q.where.all.slice(-2)).toEqual(tuple.map((value,index) => ({field:widget.query.groupBy[index],op:value===null?'missing':'eq',...(value===null?{}:{value})})));
      expect(result.count).toBe(String(groups.get(JSON.stringify(tuple))));
      expect(result.rows.map(r => r.SpanId).sort()).toEqual(F.members(rows,q).map(r=>r.SpanId).sort());
      const after = reads();
      await page.locator('[data-workspace=services]').click();
      await expect(card).toBeVisible();
      expect(reads()).toBe(after);
    }
    for (const width of [1440,390,320]) {
      await page.setViewportSize({width,height:1000});
      await card.scrollIntoViewIfNeeded();
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({path:`docs/journeys/screenshots/http-route-method-${width}.png`});
    }
    const beforeCustomize = reads();
    const expanded = response('/api/view-templates/expand');
    await card.getByRole('button',{name:'Customize route-only view',exact:true}).click();
    const expandedResponse = await expanded;
    expect(expandedResponse.status()).toBe(200);
    expect(expandedResponse.request().postDataJSON().widget.query.groupBy).toEqual([widget.query.groupBy[0]]);
    expect(reads()).toBe(beforeCustomize);
    expect(errors).toEqual([]);
    const question = {...widget.query,offset:0,where:{all:[widget.query.where,
      {field:widget.query.groupBy[0],op:'missing'},
      {field:widget.query.groupBy[1],op:'eq',value:'GET'},
    ]}};
    const actorURL = new URL(runtime.base+'/#capture='+capture.receipt.id);
    actorURL.searchParams.set('expected',JSON.stringify({service:f.service,namespace:f.namespace,question,widget:'dimension',value:[null,'GET'],count:'1'}));
    const actor = await startBombadil({dir:path.join(dir,'bombadil'),url:String(actorURL),specification:'verification/bombadil/http-server410.ts',seconds:45,width:390,height:844});
    try {
      const campaign = await actor.finish({snapshot:{name:'state',version:'goo.httpServer410Actor/1'}});
      expect(campaign.states.some(s=>s.phase==='complete'&&s.preserved)).toBe(true);
      expect(campaign.states.some(s=>s.observed?.response?.count==='1')).toBe(true);
    } finally { await actor.stop(); }
  } finally {
    await runtime.stop();
    await fs.rm(dir,{recursive:true,force:true});
  }
});
