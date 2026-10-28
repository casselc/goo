const { test, expect } = require('@playwright/test');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

test('184 synthetic supplied-page component supports SSR, detached SPA and controlled native selection', async ({ page }) => {
  const rows = [
    { TraceId: 'a'.repeat(32), SpanId: '1'.padStart(16, '0'), ParentSpanId: '',
      ServiceName: 'Synthetic API', SpanName: 'Root', timestamp_ns: '1800000000000000000', duration_ns: '100000000' },
    { TraceId: 'a'.repeat(32), SpanId: '2'.padStart(16, '0'), ParentSpanId: '1'.padStart(16, '0'),
      ServiceName: 'Synthetic API', SpanName: 'Child', timestamp_ns: '1800000000010000000', duration_ns: '20000000' },
  ];
  const supplied = { rows, truncated: true }, errors = [], requests = [];
  page.on('pageerror', e => errors.push(e.message));
  page.on('request', r => { if (r.method() !== 'GET') requests.push(r.url()); });
  const html = ('<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1">' +
    '<link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/product.css">' +
    '<link rel="stylesheet" href="/trace-page-breakdown.css">' +
    '<style>body{margin:16px;background:#10141c;color:#eee}</style>' +
    '<h1>Explicitly synthetic supplied-page component fixture</h1>' +
    '<goo-trace-page-breakdown id="ssr" data=\'' + JSON.stringify(supplied) + '\'></goo-trace-page-breakdown>' +
    '<goo-trace-page-breakdown id="upgrade"></goo-trace-page-breakdown>');
  await page.route(base + '/trace-page-synthetic', route => route.fulfill({ contentType: 'text/html', body: html }));
  await page.goto(base + '/trace-page-synthetic');
  await page.evaluate(data => { document.querySelector('#upgrade').data = data; }, supplied);
  await page.addScriptTag({type:'module',url:base+'/trace-page-breakdown-view.mjs'});
  await expect(page.locator('#ssr')).toHaveAttribute('data-state', 'observed');
  await expect(page.locator('#upgrade')).toHaveAttribute('data-truncated', 'true');
  await page.evaluate(data => {
    const spa = document.createElement('goo-trace-page-breakdown'); spa.id = 'spa'; spa.data = data;
    document.body.append(spa);
    window.intents184 = [];
    spa.addEventListener('goo-trace-page-row', e => intents184.push(e.detail));
  }, supplied);
  const spa = page.locator('#spa');
  await spa.locator('details[data-section=diagnostics] > summary').click();
  const select = spa.getByRole('button', { name: 'Select supplied span row 1', exact: true });
  await select.focus(); await select.press('Enter');
  expect(await page.evaluate(() => intents184)).toEqual([{ rowIndex: 0 }]);
  expect(await spa.evaluate(el => el.selectedIndex)).toBe(-1); // controlled, not internally committed
  await spa.evaluate(el => { el.selectedIndex = 0; });
  await expect(select).toBeFocused(); await expect(select).toHaveAttribute('aria-pressed', 'true');
  expect(await spa.locator('tr[data-row-index="0"] [data-measure="outsideChildrenNs"]').getAttribute('data-value')).toBe('80000000');
  await spa.evaluate((el, data) => { el.remove(); el.data = data; document.body.append(el); }, { rows: [], truncated: false });
  await expect(spa).toHaveAttribute('data-state', 'no_data');
  await spa.evaluate(el => { el.data = { rows: [{ TraceId: 'bad', timestamp_ns: 'x' }], truncated: false }; });
  await expect(spa).toHaveAttribute('data-state', 'partial');
  await spa.locator('details[data-section=diagnostics] > summary').click();
  await expect(spa).toContainText('Invalid or overflowing span interval');
  await page.setViewportSize({ width: 390, height: 1400 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(requests).toEqual([]); expect(errors).toEqual([]);
});
