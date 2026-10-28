const { test, expect } = require('@playwright/test');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

function supplied() {
  const row = (span, operation, duration, extra = {}) => ({
    TraceId: 'a'.repeat(32), SpanId: span.toString(16).padStart(16, '0'), ParentSpanId: '',
    ServiceName: 'Synthetic API', SpanName: operation, SpanKind: 'Server',
    timestamp_ns: '1800000000000000000', duration_ns: duration,
    Attributes: {}, ResourceAttributes: {}, ...extra,
  });
  const objectA = { a: 1, b: { value: [true, 'x'] } }, objectB = { b: { value: [true, 'x'] }, a: 1 };
  return { baseline: { rows: [
    row(1, 'Root', '100000000', { Attributes: { sameObject: objectA, empty: '', changed: 'old' } }),
    row(2, 'Zero', '0'), row(3, 'Invalid interval', '9', { timestamp_ns: 'invalid' }),
  ], truncated: false }, selection: { rows: [
    row(11, 'Root', '130000000', { TraceId: 'b'.repeat(32), Attributes: { changed: 'new', sameObject: objectB } }),
    row(12, 'Zero', '0', { TraceId: 'b'.repeat(32) }),
    row(13, 'Invalid interval', '11', { TraceId: 'b'.repeat(32), timestamp_ns: 'invalid' }),
  ], truncated: true } };
}

test('185 synthetic comparison supports SSR and detached SPA without fetching or false attribute changes', async ({ page }) => {
  const input = supplied(), errors = [];
  page.on('pageerror', error => errors.push(error.message));
  const html = '<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1">' +
    '<link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/product.css">' +
    '<link rel="stylesheet" href="/trace-comparison-view.css">' +
    '<style>body{margin:16px;background:#10141c;color:#eee}</style>' +
    '<h1>Explicitly synthetic supplied trace pages</h1>' +
    '<goo-trace-comparison id="ssr" data=\'' + JSON.stringify(input) + '\'></goo-trace-comparison>' +
    '<goo-trace-comparison id="upgrade"></goo-trace-comparison>';
  await page.route(base + '/trace-comparison-synthetic', route => route.fulfill({ contentType: 'text/html', body: html }));
  await page.goto(base + '/trace-comparison-synthetic');
  await page.evaluate(data => {
    document.querySelector('#upgrade').data = data;
    window.fixtureFetches = 0;
    window.fetch = () => { fixtureFetches++; return Promise.reject(Error('Synthetic passive fixture must not fetch')); };
    const open = XMLHttpRequest.prototype.open;
    XMLHttpRequest.prototype.open = function (...args) { fixtureFetches++; return open.apply(this, args); };
  }, input);
  await page.addScriptTag({type:'module',url:base+'/trace-comparison-view.mjs'});
  expect(await page.evaluate(()=>Object.keys(globalThis).filter(k=>/^Goo/.test(k)))).toEqual([]);
  for (const id of ['ssr', 'upgrade']) await expect(page.locator('#' + id)).toHaveAttribute('data-state', 'available');
  await page.evaluate(data => {
    const element = document.createElement('goo-trace-comparison'); element.id = 'spa'; element.data = data;
    document.body.append(element); window.pairIntents = [];
    element.addEventListener('goo-trace-comparison-pair', event => pairIntents.push(event.detail));
  }, input);
  const spa = page.locator('#spa');
  const root = spa.locator('.trace-pair').filter({ has: page.getByRole('heading', { name: '"Synthetic API" / "Root"', exact: true }) });
  await root.getByRole('button', { name: 'Inspect pair attributes', exact: true }).focus();
  await root.getByRole('button', { name: 'Inspect pair attributes', exact: true }).press('Enter');
  await expect(spa.locator('.trace-pair-attributes')).toContainText('attribute:empty');
  await expect(spa.locator('.trace-pair-attributes')).not.toContainText('sameObject');
  const empty = spa.locator('.trace-pair-attributes tr').filter({ has: page.getByRole('rowheader', { name: 'attribute:empty', exact: true }) });
  expect(await empty.locator('goo-field-value').evaluateAll(values => values.map(value => ({ state: value.state, value: value.value })))).toEqual([{ state: 'present', value: '' }, { state: 'missing', value: undefined }]);
  const zero = spa.locator('.trace-pair').filter({ has: page.getByRole('heading', { name: '"Synthetic API" / "Zero"', exact: true }) });
  expect(await zero.locator('.trace-pair-bar').evaluateAll(bars => bars.map(bar => ({ width: bar.style.width, pixels: bar.getBoundingClientRect().width })))).toEqual([{ width: '0%', pixels: 0 }, { width: '0%', pixels: 0 }]);
  const invalid = spa.locator('.trace-pair').filter({ has: page.getByRole('heading', { name: '"Synthetic API" / "Invalid interval"', exact: true }) });
  await expect(invalid).toContainText('Placement unavailable');
  await expect(invalid).toContainText('outside observed children: Unavailable');
  const pair = await root.getByRole('button', { name: 'Inspect pair attributes', exact: true }).elementHandle();
  await spa.locator('details > summary').click();
  const count = await page.evaluate(() => pairIntents.length);
  await pair.evaluate(button => button.click());
  expect(await page.evaluate(() => pairIntents.length)).toBe(count);
  await spa.evaluate(element => { window.detachedComparison185 = element; element.remove(); });
  await pair.evaluate(button => button.click());
  expect(await page.evaluate(() => pairIntents.length)).toBe(count);
  await page.evaluate(() => document.body.append(detachedComparison185));
  await spa.locator('details > summary').click();
  const currentPair = await root.getByRole('button', { name: 'Inspect pair attributes', exact: true }).elementHandle();
  await spa.evaluate(element => element.setAttribute('aria-hidden', 'true'));
  await currentPair.evaluate(button => button.click());
  expect(await page.evaluate(() => pairIntents.length)).toBe(count);
  await spa.evaluate(element => element.removeAttribute('aria-hidden'));
  await page.setViewportSize({ width: 390, height: 1800 });
  for (const widths of await spa.locator('.trace-pair').evaluateAll(cards => cards.map(card => [...card.querySelectorAll('.trace-pair-axis')].map(axis => axis.getBoundingClientRect().width)))) {
    expect(Math.abs(widths[0] - widths[1])).toBeLessThanOrEqual(1);
  }
  const unknown = structuredClone(input); delete unknown.selection.rows[0].Attributes;
  await spa.evaluate((element, value) => { element.data = value; }, unknown);
  await root.getByRole('button', { name: 'Inspect pair attributes', exact: true }).click();
  const unknownEmpty = spa.locator('.trace-pair-attributes tr').filter({ has: page.getByRole('rowheader', { name: 'attribute:empty', exact: true }) });
  expect(await unknownEmpty.locator('goo-field-value').last().evaluate(field => field.state)).toBe('unknown');
  await spa.evaluate(element => {
    window.envelopeGetterCalls = 0;
    const value = { selection: { rows: [] } };
    Object.defineProperty(value, 'baseline', { enumerable: true, get() { envelopeGetterCalls++; return { rows: [] }; } });
    element.data = value;
  });
  await expect(spa).toHaveAttribute('data-state', 'unavailable');
  expect(await page.evaluate(() => envelopeGetterCalls)).toBe(0);
  await spa.evaluate(element => { element.remove(); element.data = { baseline: { rows: [] }, selection: { rows: [] } }; document.body.append(element); });
  await expect(spa).toHaveAttribute('data-state', 'available');
  await expect(spa).toContainText('0 label groups');
  expect(await page.evaluate(() => fixtureFetches)).toBe(0);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  expect(errors).toEqual([]);
});
