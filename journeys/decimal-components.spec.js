"use strict";

const {
  test,
  expect
} = require('@playwright/test');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
function ratio(n, d) {
  let a = BigInt(n) * 100n,
    b = BigInt(d),
    x = a,
    y = b;
  while (y) [x, y] = [y, x % y];
  return b / x === 1n ? String(a / x) : a / x + '/' + b / x;
}
function aligned(mode = 'count') {
  const edges = [];
  for (let i = 63; i >= 0; i--) edges.push(-(1n << BigInt(i)));
  edges.push(0n);
  for (let i = 0; i <= 63; i++) edges.push(1n << BigInt(i));
  return {
    type: 'decimal64',
    scale: 2,
    unit: 'USD',
    valid: ['4', '4'],
    mode,
    interactive: true,
    distribution: {
      state: 'comparable',
      binning: 'powers-of-two',
      denominator: 'population-count',
      baselineCount: '8',
      selectionCount: '10',
      bins: edges.slice(0, -1).map((from, i) => {
        const a = {
            64: 1,
            65: 1,
            72: 2
          }[i] || 0,
          b = {
            64: 1,
            65: 2,
            73: 1
          }[i] || 0;
        return {
          from: String(from),
          to: String(edges[i + 1]),
          baselineCount: String(a),
          selectionCount: String(b),
          baselinePercent: ratio(a, 8),
          selectionPercent: ratio(b, 10)
        };
      })
    }
  };
}
function tiny() {
  const summary = n => ({
    state: 'observed',
    scale: 9,
    unit: 'USD',
    count: '1',
    valid: '1',
    missing: '0',
    empty: '0',
    invalid: '0',
    sum: String(n),
    min: String(n),
    max: String(n)
  });
  return {
    binding: {
      field: {
        kind: 'attribute',
        key: 'charge.nano'
      },
      type: 'decimal64',
      scale: 9,
      unit: 'USD',
      parse: 'canonical-decimal'
    },
    state: 'comparable',
    baseline: summary(1),
    selection: summary(2),
    sumDifference: '1',
    minDifference: '1',
    maxDifference: '1',
    meanDifference: '1',
    coverage: []
  };
}
const escape = s => s.replace(/&/g, '&amp;').replace(/"/g, '&quot;').replace(/</g, '&lt;');
test('180 passive decimal components preserve scaled SSR and detached SPA labels, exact means and unchanged shares', async ({
  page
}) => {
  const requests = [],
    errors = [];
  page.on('request', r => {
    if (new URL(r.url()).pathname.startsWith('/api/')) requests.push(r.url());
  });
  page.on('pageerror', e => errors.push(e.message));
  const metric = require('../ui/numeric-comparison').model(tiny());
  expect(metric).not.toBeNull();
  const html = '<!doctype html><meta name="viewport" content="width=device-width,initial-scale=1"><link rel="stylesheet" href="/components.css"><link rel="stylesheet" href="/product.css"><link rel="stylesheet" href="/time-bars.css"><link rel="stylesheet" href="/numeric-distribution.css"><style>body{margin:16px;background:#10141c;color:#dce7f5;font:14px/1.5 system-ui;max-width:900px}goo-aligned-histogram,goo-metric-comparison{display:block}</style><h1>Supplied decimal displays</h1><p>Synthetic SSR/SPA fixture. No application runtime or telemetry source.</p><goo-aligned-histogram id="ssr" data="' + escape(JSON.stringify(aligned())) + '"></goo-aligned-histogram><goo-aligned-histogram id="spa"></goo-aligned-histogram><goo-metric-comparison id="metric" data="' + escape(JSON.stringify(metric)) + '"></goo-metric-comparison><div id="detached"></div>';
  await page.route(base + '/decimal-components-standalone', r => r.fulfill({
    contentType: 'text/html; charset=utf-8',
    body: html
  }));
  await page.goto(base + '/decimal-components-standalone');
  await page.locator('#spa').evaluate((e, data) => e.data = data, aligned('share'));
  for (const script of ['app-display.mjs', 'numeric-measures.js', 'time-bars.mjs', 'numeric-distribution.js', 'aligned-histogram.mjs', 'comparison-visual.mjs', 'numeric-comparison.js']) await page.addScriptTag({type:'module',url: base + '/' + script
  });
  for (const id of ['ssr', 'spa']) {
    await expect(page.locator('#' + id + ' goo-time-bars')).toHaveCount(2);
    const labels = await page.locator('#' + id + ' goo-time-bars').evaluateAll(plots => plots.map(e => e.data));
    expect(labels[0].points[0].key).toBe('64');
    expect(labels[0].points[0].label).toContain('[0.00, 0.01) USD');
    expect(labels[1].points[0].label).toContain('[0.00, 0.01) USD');
    expect(labels[0].points.map(p => p.key)).toEqual(labels[1].points.map(p => p.key));
    if (id === 'spa') {
      expect(labels[0].ceiling).toBe('100');
      expect(labels[0].points[0].value).toBe('25/2');
      expect(labels[1].points[0].value).toBe('10');
    } else expect(labels[0].ceiling).toBe('2');
  }
  const mean = page.locator('#metric tbody>tr').first().locator('.comparison-value');
  await expect(mean).toHaveText(['0.000000001 USD', '0.000000002 USD']);
  expect(await mean.first().getAttribute('title')).toBe('1/1000000000 USD (exact)');
  const axis = page.locator('#metric .comparison-row-axis').first();
  await expect(axis).toHaveText('Shared range: 0 … 0.000000002 USD');
  await expect(axis).not.toContainText('1/1000000000');
  await page.evaluate(data => {
    const element = document.createElement('goo-metric-comparison');
    element.id = 'detachedMetric';
    element.data = GooNumericComparison.model(data);
    data.baseline.sum = '999';
    element.data.rows[0].display[0] = 'mutated getter';
    document.getElementById('detached').append(element);
  }, tiny());
  await expect(page.locator('#detachedMetric tbody>tr').first().locator('.comparison-value')).toHaveText(['0.000000001 USD', '0.000000002 USD']);
  await page.locator('#spa').evaluate(e => {
    window.decimalRanges = [];
    e.addEventListener('goo-aligned-range', event => window.decimalRanges.push(event.detail));
    e.remove();
    document.body.append(e);
  });
  const bar = page.locator('#spa [data-side=selection] button[data-key="64"]');
  await bar.focus();
  await page.keyboard.press('Space');
  await page.keyboard.press('Shift+ArrowRight');
  await page.keyboard.press('Enter');
  expect(await page.evaluate(() => window.decimalRanges)).toEqual([{
    side: 'selection',
    fromBin: 64,
    toBin: 66
  }]);
  await page.setViewportSize({
    width: 390,
    height: 1800
  });
  await page.locator('#metric').evaluate(e => e.scrollIntoView({
    block: 'start'
  }));
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  for (const row of await page.locator('#metric tbody>tr').evaluateAll(rows => rows.map(row => ({
    axis: row.querySelector('.comparison-row-axis').textContent,
    cells: [...row.querySelectorAll('.comparison-value')].map(e => ({
      text: e.textContent,
      clip: e.scrollWidth > e.clientWidth + 1
    }))
  })))) {
    expect(row.axis).toBe('Shared range: 0 … 0.000000002 USD');
    for (const cell of row.cells) expect(cell.clip).toBe(false);
  }
  await page.locator('#metric').screenshot({
    path: 'docs/journeys/screenshots/180c-passive-scaled-decimal-components.png'
  });
  expect(requests).toEqual([]);
  expect(errors).toEqual([]);
  expect(await page.evaluate(() => typeof GooQueries)).toBe('undefined');
});
