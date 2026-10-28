const {test, expect} = require('@playwright/test');
const {moduleFixture} = require('./module-fixture');
const base = 'http://127.0.0.1:8088';

test('native display imports upgrade SSR and detached SPA data with no application runtime', async ({page}) => {
  await moduleFixture(page);
  const errors = [], reads = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    if (new URL(request.url()).pathname.startsWith('/api/')) reads.push(request.url());
  });
  await page.setContent('<goo-latency id="ssr" heading="Span latency" summary=\'{"state":"observed","count":"1","sumNs":"1000000","p50Ns":"1000000"}\'><p>SSR fallback</p></goo-latency><goo-chart-frame id="frame"><button id="kept">Caller action</button></goo-chart-frame><div id="spa-host"></div>');
  await page.evaluate(() => {
    const span = document.createElement('goo-trace-waterfall');
    span.id = 'spa';
    span.data = {rows:[{SpanId:'a',TraceId:'b',timestamp_ns:'1700000000000000001',duration_ns:'1',SpanName:'Supplied span',ServiceName:'example'}],truncated:false};
    span.selectedIndex = 0;
    window.detachedSpan = span;
    window.keptAction = document.querySelector('#kept');
  });
  await expect(page.locator('#ssr')).toContainText('SSR fallback');
  await page.evaluate(async () => {
    const first = await import('/display.mjs');
    const repeated = await import('/display.mjs');
    const direct = await import('/components.mjs');
    if (first !== repeated || first.UI !== direct) throw Error('Module identity changed');
    document.querySelector('#spa-host').append(window.detachedSpan);
  });
  await expect(page.locator('#ssr')).toContainText('1.000 ms');
  await expect(page.locator('#spa .waterfall-row')).toContainText('Supplied span');
  await expect(page.locator('#spa .waterfall-row')).toHaveAttribute('aria-pressed','true');
  expect(await page.evaluate(() => document.querySelector('#kept') === window.keptAction)).toBe(true);
  await page.locator('#spa').evaluate(element => {
    element.remove();
    element.data = {rows:[],truncated:false};
    document.querySelector('#spa-host').append(element);
  });
  await expect(page.locator('#spa')).toHaveAttribute('data-state','empty');
  expect(await page.evaluate(() => Object.keys(globalThis).filter(k => /^Goo/.test(k)))).toEqual([]);
  expect(reads).toEqual([]);
  expect(errors).toEqual([]);
});

test('record actions and authored tags work without application namespaces', async ({page}) => {
  await moduleFixture(page);
  const reads = [], errors = [];
  page.on('request', request => {
    if (new URL(request.url()).pathname.startsWith('/api/')) reads.push(request.url());
  });
  page.on('pageerror', error => errors.push(error.message));
  await page.setContent('<div id="attributes"></div><div id="tags"></div>');
  await page.evaluate(() => {
    const tags = document.createElement('goo-tag-list');
    tags.id = 'early-tags';
    tags.tags = ['team:payments'];
    window.earlyTags = tags;
    window.traceIntents = [];
  });
  await page.evaluate(async () => {
    const Attributes = await import('/attributes.mjs');
    await import('/tag-list.mjs');
    document.querySelector('#tags').append(window.earlyTags);
    Attributes.mount(document.querySelector('#attributes'), {SpanAttributes:{'http.route':'/items'}}, {
      canTraceSearch: () => true,
      onTraceSearch: intent => window.traceIntents.push(intent),
    });
  });
  await expect(page.locator('#early-tags')).toHaveText('team:payments');
  await page.locator('#early-tags').evaluate(element => {
    element.remove();
    element.tags = ['review:latency'];
    document.querySelector('#tags').append(element);
  });
  await expect(page.locator('#early-tags')).toHaveText('review:latency');
  const action = page.getByRole('button',{name:'Find traces containing attribute:http.route',exact:true});
  await action.click();
  expect(await page.evaluate(() => window.traceIntents)).toEqual([
    {field:{kind:'attribute',key:'http.route'},op:'eq',value:'/items'},
  ]);
  await action.evaluate(button => {
    const host = document.querySelector('#attributes');
    host.hidden = true;
    button.click();
    host.remove();
    button.click();
  });
  expect(await page.evaluate(() => window.traceIntents.length)).toBe(1);
  expect(await page.evaluate(() => Object.keys(globalThis).filter(k => /^Goo/.test(k)))).toEqual([]);
  expect(reads).toEqual([]);
  expect(errors).toEqual([]);
});

test('the whole supplied-data gallery uses the public graph without application owners', async ({page}) => {
  const paths = [], errors = [];
  page.on('request', request => paths.push(new URL(request.url()).pathname));
  page.on('pageerror', error => errors.push(error.message));
  await page.goto(base+'/ui-components');
  await expect(page.locator('#component-tags')).toContainText('team:payments');
  await expect(page.locator('#component-record')).toContainText('GET /items');
  await expect(page.locator('#component-aligned goo-time-bars')).toHaveCount(2);
  expect(paths.filter(path => path.startsWith('/api/'))).toEqual([]);
  expect(paths).not.toContain('/app-display.mjs');
  for (const owner of ['meaning','numeric-measures','numeric-distribution','log-patterns','queries','retention']) {
    expect(paths).not.toContain('/'+owner+'.js');
  }
  expect(await page.evaluate(() => Object.keys(globalThis).filter(k => /^Goo/.test(k)))).toEqual([]);
  expect(errors).toEqual([]);
});

test('application owners register their tag-created widgets without a namespace bridge',async({page})=>{
 const errors=[];page.on('pageerror',error=>errors.push(error.message));
 await page.goto(base+'/#questions');
 await expect(page.locator('#exploreRun')).toHaveCount(1);
 const missing=await page.evaluate(()=>['goo-chart-frame','goo-value-timeline','goo-bucket-heatmap','goo-grouping-control','goo-trace-comparison','goo-record-columns','goo-comparison-series','goo-time-range'].filter(tag=>!customElements.get(tag)));
 expect(missing).toEqual([]);
 expect(await page.evaluate(()=>Object.keys(globalThis).filter(key=>/^Goo/.test(key)))).toEqual([]);
 expect(errors).toEqual([]);
});
