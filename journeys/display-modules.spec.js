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
