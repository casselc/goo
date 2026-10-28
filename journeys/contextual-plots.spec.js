"use strict";

const {
  test,
  expect
} = require('@playwright/test');
const {
  randomBytes
} = require('node:crypto');
const fs = require('node:fs/promises'),
  path = require('node:path');
const {
  questionLink
} = require('../ui/workspace.mjs');
const {
  openFilters,
  openMeasures
} = require('./support/explore');
const {
  startBombadil
} = require('./support/cross-app');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const binding = {
  field: {
    kind: 'attribute',
    key: 'work.amount'
  },
  type: 'decimal64',
  scale: 2,
  unit: 'USD',
  parse: 'canonical-decimal'
};
const lower = -(1n << 63n),
  upper = 1n << 63n;
const isResponse = (endpoint, r) => r.request().method() === 'POST' && new URL(r.url()).pathname === '/api/' + endpoint;
async function json(response) {
  expect(response.ok(), await response.text()).toBe(true);
  return response.json();
}
function rational(n, d) {
  let a = n < 0n ? -n : n,
    b = d;
  while (b) [a, b] = [b, a % b];
  n /= a;
  d /= a;
  return d === 1n ? String(n) : n + '/' + d;
}
function human(n, scale) {
  const sign = n < 0n ? '-' : '',
    digits = (n < 0n ? -n : n).toString().padStart(scale + 1, '0');
  return sign + digits.slice(0, -scale) + '.' + digits.slice(-scale);
}
// Independent decimal-text oracle. A full-length match is required because JS '$'
// may otherwise accept a trailing newline. No Number/parseFloat conversion occurs.
function scalar(raw, b) {
  if (raw === null) return {
    state: 'missing'
  };
  if (raw === '') return {
    state: 'empty'
  };
  let n;
  if (b.type === 'decimal64') {
    if (!Number.isInteger(b.scale) || b.scale < 1 || b.scale > 9) return {
      state: 'invalid'
    };
    const match = new RegExp('^(-?)(0|[1-9][0-9]*)\\.([0-9]{' + b.scale + '})$').exec(raw);
    if (!match || match[0] !== raw) return {
      state: 'invalid'
    };
    n = BigInt(match[2] + match[3]);
    if (match[1]) {
      if (n === 0n) return {
        state: 'invalid'
      };
      n = -n;
    }
  } else {
    const match = /^(0|[1-9][0-9]*|-[1-9][0-9]*)$/.exec(raw);
    if (!match || match[0] !== raw) return {
      state: 'invalid'
    };
    n = BigInt(raw);
  }
  const min = b.type === 'uint64' ? 0n : lower,
    max = b.type === 'uint64' ? 1n << 64n : upper;
  return n < min || n >= max ? {
    state: 'invalid'
  } : {
    state: 'valid',
    coefficient: n
  };
}
function rawValue(r, f) {
  if (f.kind === 'resource') return f.key === 'work.amount' ? r.resourceRaw : null;
  if (f.kind === 'attribute' && f.key === 'team') return 'core';
  if (f.kind === 'column') return f.key === 'service' ? r.service : null;
  return {
    'work.amount': r.raw,
    'amount.scale1': r.scale1Raw,
    'amount.scale9': r.scale9Raw,
    'work.items': r.integerRaw
  }[f.key] ?? null;
}
function predicate(r, p) {
  if (!p) return true;
  if (p.all) return p.all.every(x => predicate(r, x));
  if (p.any) return p.any.some(x => predicate(r, x));
  if (p.not) return !predicate(r, p.not);
  if (p.numeric) {
    const s = scalar(rawValue(r, p.numeric.field), p.numeric);
    if (['valid', 'missing', 'empty', 'invalid'].includes(p.op)) return s.state === p.op;
    if (s.state !== 'valid') return false;
    const threshold = scalar(p.value, p.numeric);
    expect(threshold.state).toBe('valid');
    const n = s.coefficient,
      z = threshold.coefficient;
    return {
      eq: n === z,
      ne: n !== z,
      lt: n < z,
      lte: n <= z,
      gt: n > z,
      gte: n >= z
    }[p.op];
  }
  const raw = rawValue(r, p.field),
    literal = p.value ?? '';
  if (p.op === 'exists') return raw !== null;
  if (p.op === 'missing') return raw === null;
  if (p.op === 'eq') return raw !== null && raw === literal;
  if (p.op === 'ne') return raw !== null && raw !== literal;
  throw Error('Unknown predicate ' + JSON.stringify(p));
}
const members = (records, q) => records.filter(r => r.environment === q.environment && r.signal === q.signal && r.timeMs >= q.fromMs && r.timeMs < q.toMs && predicate(r, q.where));
function summary(rows, b) {
  const coverage = {
    valid: 0n,
    missing: 0n,
    empty: 0n,
    invalid: 0n
  };
  let sum = 0n,
    min,
    max;
  for (const r of rows) {
    const s = scalar(rawValue(r, b.field), b);
    coverage[s.state]++;
    if (s.state === 'valid') {
      sum += s.coefficient;
      min = min === undefined || s.coefficient < min ? s.coefficient : min;
      max = max === undefined || s.coefficient > max ? s.coefficient : max;
    }
  }
  return {
    state: coverage.valid ? 'observed' : rows.length ? 'no_valid_values' : 'no_data',
    unit: b.unit,
    ...(b.type === 'decimal64' ? {
      scale: b.scale
    } : {}),
    count: String(rows.length),
    ...Object.fromEntries(Object.entries(coverage).map(([k, v]) => [k, String(v)])),
    sum: String(sum),
    ...(min === undefined ? {} : {
      min: String(min),
      max: String(max)
    })
  };
}
function edges(type = 'decimal64') {
  const values = [];
  if (type !== 'uint64') for (let i = 63; i >= 0; i--) values.push(-(1n << BigInt(i)));
  values.push(0n);
  for (let i = 0; i <= (type === 'uint64' ? 64 : 63); i++) values.push(1n << BigInt(i));
  return values;
}
function histogram(rows, b) {
  const bounds = edges(b.type),
    bins = bounds.slice(0, -1).map((v, i) => ({
      from: String(v),
      to: String(bounds[i + 1]),
      count: '0'
    }));
  for (const r of rows) {
    const parsed = scalar(rawValue(r, b.field), b);
    if (parsed.state !== 'valid') continue;
    const i = bounds.findIndex((edge, index) => index < bins.length && parsed.coefficient >= edge && parsed.coefficient < bounds[index + 1]);
    expect(i).toBeGreaterThanOrEqual(0);
    bins[i].count = String(BigInt(bins[i].count) + 1n);
  }
  return {
    binning: 'powers-of-two',
    bins
  };
}
function assertResult(result, records) {
  const q = result.question,
    rows = members(records, q),
    b = q.numeric;
  expect(result.count).toBe(String(rows.length));
  expect(result.rows).toHaveLength(Math.min(q.limit, Math.max(0, rows.length - q.offset)));
  if (b) {
    expect(result.numeric).toEqual(summary(rows, b));
    if (q.numericHistogram) {
      expect(result.numericHistogram).toEqual(histogram(rows, b));
      expect(result.numericHistogram.bins.reduce((n, x) => n + BigInt(x.count), 0n)).toBe(BigInt(result.numeric.valid));
    }
  }
  if (q.groupBy?.length) {
    const keys = new Set(rows.map(r => JSON.stringify(q.groupBy.map(f => rawValue(r, f)))));
    expect(new Set((result.groups || []).map(g => JSON.stringify(q.groupBy.map((_, i) => g['g' + i]))))).toEqual(keys);
    expect((result.groups || []).reduce((n, g) => n + BigInt(g.count), 0n)).toBe(BigInt(rows.length));
  }
  const ids = new Set(rows.map(r => r.id));
  for (const row of result.rows) expect(ids.has(q.signal === 'spans' ? row.SpanId : row.Body.replace('declared amount ', ''))).toBe(true);
  for (const group of result.groups || []) {
    const grouped = rows.filter(r => q.groupBy.every((f, i) => rawValue(r, f) === group['g' + i]));
    expect(group.count).toBe(String(grouped.length));
    if (b) expect(group.numeric).toEqual(summary(grouped, b));
  }
  for (const point of result.timeSeries || []) {
    const grouped = rows.filter(r => r.timeMs >= point.fromMs && r.timeMs < point.toMs);
    expect(point.count).toBe(String(grouped.length));
    if (b) expect(point.numeric).toEqual(summary(grouped, b));
  }
  return rows;
}
function assertNumeric(data, a, z, b) {
  expect(data.binding).toEqual(b);
  expect(data.baseline).toEqual(a);
  expect(data.selection).toEqual(z);
  for (const coverage of data.coverage) {
    const av = BigInt(a?.[coverage.kind] || '0'),
      zv = BigInt(z?.[coverage.kind] || '0');
    expect(coverage.baselineCount).toBe(String(av));
    expect(coverage.selectionCount).toBe(String(zv));
    expect(coverage.difference).toBe(String(zv - av));
    if (a && z && a.count !== '0' && z.count !== '0') expect(coverage.differencePP).toBe(rational(100n * (zv * BigInt(a.count) - av * BigInt(z.count)), BigInt(a.count) * BigInt(z.count)));
  }
  if (!a || !z || a.valid === '0' || z.valid === '0') {
    expect(data.state).toBe('unavailable');
    for (const key of ['sum', 'min', 'max', 'mean']) expect(data[key + 'Difference']).toBeUndefined();
    return;
  }
  expect(data.state).toBe('comparable');
  for (const key of ['sum', 'min', 'max']) expect(data[key + 'Difference']).toBe(String(BigInt(z[key]) - BigInt(a[key])));
  expect(data.meanDifference).toBe(rational(BigInt(z.sum) * BigInt(a.valid) - BigInt(a.sum) * BigInt(z.valid), BigInt(a.valid) * BigInt(z.valid)));
}
function assertPair(pair, records) {
  const a = pair.request.baseline,
    z = pair.request.selection;
  expect(pair.request.numeric).toBe(true);
  expect(a.numeric).toEqual(summary(members(records, a.question), a.question.numeric));
  expect(z.numeric).toEqual(summary(members(records, z.question), z.question.numeric));
  assertNumeric(pair.response.measures.numeric, a.numeric, z.numeric, a.question.numeric);
  for (const row of pair.response.rows) {
    const ag = a.groups.find(g => g.g0 === row.value),
      zg = z.groups.find(g => g.g0 === row.value);
    assertNumeric(row.measures.numeric, ag?.numeric, zg?.numeric, a.question.numeric);
  }
}
const raws = ['-92233720368547758.08', '92233720368547758.07', '90071992547409.93', '-90071992547409.93', '0.00', '0.01', '-0.01', '1.20', '-1.25', '2.50', '3.33', '-8.00', '4.00', '16.00', '64.00', '999.99', null, '', '1', '1.2', '1.200', '01.20', '+1.20', '-0.00', ' 1.20', '1.20 ', '1e2', 'NaN', '1,20', '.20', '-.20', '92233720368547758.08', '-92233720368547758.09', '0.000', '00.00', 'bad', '1.20\n', '1.20\r\n', '1.20\t', '1.20\u2028'];
async function fixture(request, signal) {
  const environment = 'contextual-plot-' + randomBytes(5).toString('hex'),
    toMs = Date.now(),
    fromMs = toMs - 60000,
    coefficients = [null, '', 0n, 1n, -1n, upper - 1n, lower, 9007199254740993n, 120n, 250n, 333n, -125n],
    integerValues = ['-9223372036854775808', '9223372036854775807', '9007199254740993', '0', '1', null, '', '01'];
  const records = Array.from({
    length: 400
  }, (_, i) => {
    const raw = raws[i % raws.length],
      coefficient = coefficients[i % coefficients.length],
      formatted = scale => coefficient === null ? null : coefficient === '' ? '' : human(coefficient, scale);
    return {
      environment,
      signal,
      id: randomBytes(8).toString('hex'),
      traceId: randomBytes(16).toString('hex'),
      timeMs: fromMs + 1000 + i * 100,
      service: raw === '3.33' ? 'decimal-only-worker' : raw === 'bad' ? 'invalid-only-worker' : i % 3 === 0 ? 'checkout-api' : 'queue-worker',
      raw,
      resourceRaw: [null, '', 'bad', '1.20', '-0.01', '92233720368547758.07'][i % 6],
      scale1Raw: i % 13 === 12 ? '1.00' : formatted(1),
      scale9Raw: i % 13 === 12 ? '1.00000000' : formatted(9),
      integerRaw: integerValues[i % integerValues.length]
    };
  });
  const attrs = o => Object.entries(o).map(([key, stringValue]) => ({
      key,
      value: {
        stringValue
      }
    })),
    resources = records.map(r => {
      const resource = {
          attributes: attrs({
            'service.name': r.service,
            'deployment.environment.name': environment,
            ...(r.resourceRaw === null ? {} : {
              'work.amount': r.resourceRaw
            })
          })
        },
        attributes = attrs(Object.fromEntries([['work.amount', r.raw], ['amount.scale1', r.scale1Raw], ['amount.scale9', r.scale9Raw], ['work.items', r.integerRaw], ['team', 'core']].filter(([, value]) => value !== null))),
        at = String(BigInt(r.timeMs) * 1000000n);
      return signal === 'spans' ? {
        resource,
        scopeSpans: [{
          scope: {
            name: 'ordinary.decimal.worker'
          },
          spans: [{
            traceId: r.traceId,
            spanId: r.id,
            name: 'process declared amount',
            kind: 2,
            status: {
              code: 1
            },
            startTimeUnixNano: at,
            endTimeUnixNano: String(BigInt(at) + 1000000n),
            attributes
          }]
        }]
      } : {
        resource,
        scopeLogs: [{
          scope: {
            name: 'ordinary.decimal.worker'
          },
          logRecords: [{
            timeUnixNano: at,
            traceId: r.traceId,
            spanId: r.id,
            severityText: 'INFO',
            body: {
              stringValue: 'declared amount ' + r.id
            },
            attributes
          }]
        }]
      };
    });
  expect((await request.post('http://127.0.0.1:4318/v1/' + (signal === 'spans' ? 'traces' : 'logs'), {
    data: signal === 'spans' ? {
      resourceSpans: resources
    } : {
      resourceLogs: resources
    }
  })).ok()).toBe(true);
  const q = {
    where: {field: {kind: 'attribute', key: 'team'}, op: 'eq', value: 'core'},
    signal,
    environment,
    fromMs,
    toMs,
    limit: 100,
    offset: 0,
    bucketMs: 10000,
    groupBy: [{
      kind: 'column',
      key: 'service'
    }]
  };
  await expect.poll(async () => (await json(await request.post(base + '/api/explore', {
    data: q
  }))).count, {
    timeout: 45000,
    intervals: [500]
  }).toBe('400');
  const selector = {signal, environment, where: q.where};
  const activityView = await json(await request.post(base + '/api/activity-views', {data: {version: 1, title: 'Contextual plot ' + signal, selector, columns: ['utc', 'service', 'summary']}}));
  return {
    selector, activityView,
    environment,
    signal,
    records,
    q
  };
}
function observe(page) {
  const state = {
    reads: 0,
    plans: 0,
    protocols: 0,
    writes: 0,
    streams: 0,
    events: [],
    requests: [],
    errors: []
  };
  page.on('pageerror', e => state.errors.push(e.message));
  page.on('request', r => {
    if (r.method() !== 'POST') return;
    const endpoint = new URL(r.url()).pathname,
      body = r.postDataJSON();
    let effect;
    if (endpoint === '/api/tail') {
      state.streams++; effect = 'stream';
    } else if (['/api/explore', '/api/captures', '/api/measurements', '/api/discovery', '/api/inventory'].includes(endpoint)) {
      state.reads++;
      effect = 'source';
    } else if (endpoint === '/api/numeric-histograms/selection') {
      state.plans++;
      effect = 'plan';
    } else if (endpoint === '/api/compare') {
      state.protocols++;
      effect = 'comparison';
    } else if (['/api/telemetry-packages', '/api/exploration-views'].includes(endpoint)) {
      state.writes++;
      effect = 'publication';
    }
    if (effect) state.requests.push({
      effect,
      endpoint,
      body
    });
  });
  state.event = action => state.events.push({
    action,
    reads: state.reads,
    plans: state.plans,
    protocols: state.protocols,
    writes: state.writes,
    streams: state.streams
  });
  return state;
}

// A context may supply observed strings; only an explicit declaration supplies units.
const { readQuestionLink } = require('../ui/workspace.mjs');
const plotField = kind => ({kind, key: 'work.amount'});
const dialog = page => page.getByRole('dialog', {name: 'Plot attribute values', exact: true});
const plotButton = (host, kind) => host.getByRole('button', {name: 'Plot values of ' + kind + ':work.amount', exact: true});
async function observeTail(page) {
  await page.addInitScript(() => {
    window.__plotTailRows = [];
    window.__plotStages = [];
    window.addEventListener('goo:question-state', event => {
      if (event.detail.action === 'stage') window.__plotStages.push(structuredClone(event.detail.requested));
    });
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const response = await original(input, init);
      const url = typeof input === 'string' ? input : input.url;
      if (url?.endsWith('/api/tail') && response.ok) {
        const reader = response.clone().body.getReader(), decoder = new TextDecoder();
        (async () => {
          let text = '';
          try {
            while (true) {
              const part = await reader.read();
              if (part.done) break;
              text += decoder.decode(part.value, {stream: true});
              let end;
              while ((end = text.indexOf('\n\n')) >= 0) {
                const frame = text.slice(0, end); text = text.slice(end + 2);
                if (frame.includes('event: batch')) {
                  const data = frame.split('\n').filter(x => x.startsWith('data:')).map(x => x.slice(5).trimStart()).join('\n');
                  const batch = JSON.parse(data);
                  window.__plotTailRows.push(...batch.rows.map(row => ({...row, readWindow: {fromMs: batch.fromMs, toMs: batch.toMs}})));
                }
              }
            }
          } catch { /* Pausing aborts the actual stream. */ }
        })();
      }
      return response;
    };
  });
}
async function stagedQuestion(page) {
  await expect(page.locator('#exploreStatus')).toContainText(/No evidence loaded|Analysis staged\. Run analysis reads this fixed window/);
  const staged = await page.evaluate(() => window.__plotStages.at(-1));
  expect(staged).toBeTruthy();
  return staged;
}
function assertTransform(parent, staged, field) {
  const expected = structuredClone(parent);
  for (const key of ['groupBy', 'groupTimeSeries', 'bucketMs', 'summary', 'spanStatus']) delete expected[key];
  expected.numeric = {...binding, field};
  expected.numericHistogram = {binning: 'powers-of-two'};
  expected.limit = 100;
  expected.offset = 0;
  expect(staged).toEqual(expected);
}
async function declare(page, field, {inherit = false, screenshot} = {}) {
  const d = dialog(page);
  await expect(d).toBeVisible();
  await expect(d).toContainText(field.kind + ':' + field.key);
  await expect(d).toContainText('Plots all matching records in this fixed window');
  const type = d.locator('#plotValuesnumericType');
  const unit = d.locator('#plotValuesnumericUnit');
  const open = d.getByRole('button', {name: 'Open value distribution', exact: true});
  if (!inherit) {
    await expect(type).toHaveValue('');
    await expect(unit).toHaveValue('');
    await expect(open).toBeDisabled();
    await type.selectOption('decimal64');
    await expect(d.locator('#plotValuesnumericScale')).toHaveValue('');
    await unit.fill('USD');
    await expect(open).toBeDisabled();
    await d.locator('#plotValuesnumericScale').fill('2.0');
    await expect(open).toBeDisabled();
    await d.locator('#plotValuesnumericScale').fill('2');
  } else {
    await expect(type).toHaveValue('decimal64');
    await expect(unit).toHaveValue('USD');
    await expect(d.locator('#plotValuesnumericScale')).toHaveValue('2');
  }
  await expect(open).toBeEnabled();
  if (screenshot) await d.screenshot({path: screenshot});
}
async function runPlot(page, records) {
  const pending = page.waitForResponse(r => isResponse('explore', r));
  await page.locator('#exploreRun').click();
  const result = await json(await pending);
  assertResult(result, records);
  return result;
}
async function openFixedRecord(page, row = 0) {
  await page.locator('#exploreRows .record-row').nth(row).getByRole('button', {name: 'Inspect record', exact: true}).click();
  return page.locator('#exploreRows .record-detail:visible');
}
async function brush(page) {
  const host = page.locator('#numericSummary .numeric-distribution');
  await host.locator('button[data-key="72"]').focus();
  await page.keyboard.press('Escape');
  await page.keyboard.press('Space');
  await page.keyboard.press('Shift+ArrowRight');
  await page.keyboard.press('Enter');
  return host;
}

for (const signal of ['spans', 'logs']) test('181 ' + signal + ' contextual plots declare and read full fixed populations', async ({page, request}) => {
  test.setTimeout(300000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request, signal), state = observe(page), origins = [];
  await page.setViewportSize({width: 1440, height: 1400});
  await observeTail(page);
  await page.goto(base + '/#activity-view=' + f.activityView.id);
  const activity = page.locator('.stream-entry');
  await activity.getByRole('button', {name: 'Start stream', exact: true}).click();
  const tail = activity.locator('.live-tail');
  await expect(tail).toHaveAttribute('data-last-to', /^\d+$/);
  await expect(tail.locator('tbody>tr')).not.toHaveCount(0);
  state.event('activity-start');
  await tail.locator('tbody>tr').first().getByRole('button', {name: 'Inspect', exact: true}).click();
  const activityDetail = activity.locator('.record-detail:visible');
  const opener = plotButton(activityDetail, 'attribute');
  await opener.click();
  await expect(tail).toHaveAttribute('data-state', 'paused');
  const heldRecord = JSON.parse(await activityDetail.locator('.raw-record > pre').textContent());
  const heldRow = await page.evaluate(record => window.__plotTailRows.find(row =>
    row.record.SpanId === record.SpanId && row.record.timestamp_ns === record.timestamp_ns), heldRecord);
  expect(heldRow).toBeTruthy();
  const parent = {...f.selector, ...heldRow.readWindow, limit: 100, offset: 0};
  const applied = {selector: f.selector, readWindow: heldRow.readWindow, heldRecord, heldRow};
  expect(state.reads).toBe(0);
  await declare(page, plotField('attribute'));
  await dialog(page).getByRole('button', {name: 'Cancel', exact: true}).click();
  await expect(opener).toBeFocused();
  expect(state.reads).toBe(0);
  state.event('activity-cancel');
  await opener.click();
  await declare(page, plotField('attribute'), {screenshot: 'docs/journeys/screenshots/181a-' + signal + '-declare-value-plot.png'});
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  const staged = await stagedQuestion(page);
  assertTransform(parent, staged, plotField('attribute'));
  expect(state.reads).toBe(0);
  state.event('activity-stage');
  const result = await runPlot(page, f.records);
  expect(BigInt(result.count)).toBeGreaterThan(100n);
  expect(result.rows).toHaveLength(100);
  expect(result.numeric.invalid).not.toBe('0');
  expect(result.numeric.empty).not.toBe('0');
  origins.push({kind: 'activity', parent, field: plotField('attribute'), staged, result});
  state.event('activity-run');

  // A declared identical field inherits its declaration; the resource namesake does not.
  const invalidIndex = result.rows.findIndex(row => {
    const attributes = signal === 'spans' ? row.SpanAttributes : row.Attributes;
    return Object.hasOwn(attributes || {}, 'work.amount') && ['empty', 'invalid'].includes(scalar(attributes['work.amount'], binding).state);
  });
  expect(invalidIndex).toBeGreaterThanOrEqual(0);
  let detail = await openFixedRecord(page, invalidIndex);
  await plotButton(detail, 'attribute').click();
  await declare(page, plotField('attribute'), {inherit: true});
  await page.keyboard.press('Escape');
  expect(state.reads).toBe(1);
  state.event('fixed-same-field-cancel');
  const emptyIndex = result.rows.findIndex(row => row.ResourceAttributes?.['work.amount'] === '');
  expect(emptyIndex).toBeGreaterThanOrEqual(0);
  detail = await openFixedRecord(page, emptyIndex);
  await expect(plotButton(detail, 'resource')).toBeEnabled();
  await plotButton(detail, 'resource').click();
  await declare(page, plotField('resource'));
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  const resourceStaged = await stagedQuestion(page);
  assertTransform(result.question, resourceStaged, plotField('resource'));
  expect(state.reads).toBe(1);
  state.event('resource-stage');
  const resourceResult = await runPlot(page, f.records);
  expect(resourceResult.numeric).not.toEqual(result.numeric);
  origins.push({kind: 'fixed', parent: result.question, field: plotField('resource'), staged: resourceStaged, result: resourceResult});
  state.event('resource-run');

  // Observed discovery fields are suggestions from the completed read, never type inference.
  await openFilters(page);
  const observations = page.locator('#filterComposer .filter-observations');
  await observations.locator(':scope>summary').click();
  const fieldObservation = observations.locator('details[data-field="attribute:work.amount"]');
  await fieldObservation.locator(':scope>summary').click();
  await plotButton(fieldObservation, 'attribute').click();
  await declare(page, plotField('attribute'));
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  const discoveryStaged = await stagedQuestion(page);
  assertTransform(resourceResult.question, discoveryStaged, plotField('attribute'));
  expect(state.reads).toBe(2);
  origins.push({kind: 'discovery', parent: resourceResult.question, field: plotField('attribute'), staged: discoveryStaged});
  state.event('discovery-stage');
  const parentAgain = await runPlot(page, f.records);
  state.event('discovery-run');

  const chart = await brush(page);
  expect(state.reads).toBe(3);
  expect(state.plans).toBe(0);
  state.event('brush');
  let planResponse = page.waitForResponse(r => isResponse('numeric-histograms/selection', r));
  let childResponse = page.waitForResponse(r => isResponse('explore', r));
  await chart.getByRole('button', {name: 'Run selected population', exact: true}).click();
  const pr = await planResponse, selection = {request: pr.request().postDataJSON(), response: await json(pr)};
  const selected = await json(await childResponse);
  assertResult(selected, f.records);
  expect(selected.question).toEqual(selection.response.selected);
  state.event('selected-run');
  await page.getByRole('button', {name: 'Back to question', exact: true}).click();
  await brush(page);
  expect(state.reads).toBe(4);
  state.event('back-brush');
  planResponse = page.waitForResponse(r => isResponse('numeric-histograms/selection', r));
  await page.locator('#numericSummary .numeric-distribution').getByRole('button', {name: 'Stage numeric comparison with rest', exact: true}).click();
  const countPlanResponse = await planResponse, countPlan = await json(countPlanResponse);
  expect(state.reads).toBe(4);
  state.event('comparison-stage');
  const pairResponse = page.waitForResponse(r => isResponse('compare', r));
  await page.locator('.population-run:visible').click();
  const pairRaw = await pairResponse, comparison = {request: pairRaw.request().postDataJSON(), response: await json(pairRaw)};
  assertPair(comparison, f.records);
  state.event('comparison-run');

  // Capture the actual plotted parent, then declare from frozen observed attributes.
  await page.goto(base + '/' + questionLink(parentAgain.question, 'Distribution'));
  await runPlot(page, f.records);
  await page.locator('#exploreSaveCapture').click();
  await page.locator('#captureDraftTitle').fill('Contextual ' + signal + ' value distribution');
  const captureResponse = page.waitForResponse(r => isResponse('captures', r));
  await page.locator('#captureDraftConfirm').click();
  const saved = await json(await captureResponse);
  await expect(page.locator('#captureGeneralID')).toHaveValue(saved.receipt.id);
  state.event('capture');
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  const beforeSaved = state.reads;
  const savedResourceIndex = saved.explore.rows.findIndex(row => Object.hasOwn(row.ResourceAttributes || {}, 'work.amount'));
  expect(savedResourceIndex).toBeGreaterThanOrEqual(0);
  await page.locator('#savedexploreRows .record-row').nth(savedResourceIndex).getByRole('button', {name: 'Inspect record', exact: true}).click();
  const savedDetail = page.locator('#savedexploreRows .record-detail:visible');
  await plotButton(savedDetail, 'resource').click();
  await declare(page, plotField('resource'));
  await expect(dialog(page)).toContainText('saved evidence is unchanged');
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  const savedStaged = await stagedQuestion(page);
  assertTransform(saved.receipt.question.explore, savedStaged, plotField('resource'));
  expect(state.reads).toBe(beforeSaved);
  origins.push({kind: 'saved', parent: saved.receipt.question.explore, field: plotField('resource'), staged: savedStaged});
  state.event('saved-stage');
  const savedRun = await runPlot(page, f.records);
  assertResult(savedRun, f.records);
  expect(await json(await request.get(base + '/api/captures/' + saved.receipt.id))).toEqual(saved);
  state.event('saved-run');
  await page.setViewportSize({width: 390, height: 1800});
  const plot = page.locator('#numericSummary .numeric-distribution');
  await plot.evaluate(el => {el.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await plot.screenshot({path: 'docs/journeys/screenshots/181b-' + signal + '-full-value-distribution.png'});
  await page.getByRole('button', {name: 'Return to paused activity', exact: true}).click();
  await expect(tail).toHaveAttribute('data-state', 'paused');
  await expect(opener).toBeFocused();
  state.event('return-paused');
  expect(state.errors).toEqual([]);
  const dir = path.resolve('artifacts/contextual-plots', f.environment);
  await fs.mkdir(dir, {recursive: true});
  await fs.writeFile(path.join(dir, 'ledger.json'), JSON.stringify({version: 1, ...f, applied, origins, result, resourceResult, parentAgain, selection, selected, countPlan, comparison, saved, reopened, savedRun, events: state.events, requests: state.requests}, null, 2));
});

// A retained dialog is a declaration draft, not a source owner or a navigation command.
test('181 contextual plot dialogs lose ownership on dirty scope, navigation and resume', async ({page, request}) => {
  test.setTimeout(180000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request, 'logs'), state = observe(page), observations = [];
  await observeTail(page);
  await page.goto(base + '/#activity-view=' + f.activityView.id);
  const activity = page.locator('.stream-entry'), tail = activity.locator('.live-tail');
  await activity.getByRole('button', {name: 'Start stream', exact: true}).click();
  await expect(tail.locator('tbody>tr')).not.toHaveCount(0);
  async function open() {
    await tail.locator('tbody>tr').first().getByRole('button', {name: 'Inspect', exact: true}).click();
    await plotButton(activity.locator('.record-detail:visible'), 'attribute').click();
    await declare(page, plotField('attribute'));
  }
  async function observeOwnership(reason, action) {
    const beforeURL = page.url();
    await dialog(page).locator('button').filter({hasText: 'Open value distribution'}).evaluate(el => {window.__stalePlotButton = el; window.__stalePlotDialog = el.closest('dialog')});
    await action();
    await expect(dialog(page)).not.toBeVisible();
    const currentURL = page.url();
    await page.evaluate(() => window.__stalePlotButton.click());
    expect(page.url()).toBe(currentURL);
    expect(state.reads).toBe(0);
    observations.push({reason, beforeURL, afterURL: page.url(), dialogOpen: await page.evaluate(() => window.__stalePlotDialog.open), streams: state.streams, reads: state.reads});
    state.event(reason);
  }
  await open();
  state.event('open');
  await observeOwnership('navigation', async () => page.goto(base + '/#definitions'));
  await page.goto(base + '/#explore');
  await expect(tail).toHaveAttribute('data-state', 'paused');
  await open();
  await observeOwnership('dirty-scope', async () => {
    // Programmatic edit models a retained application input event while native modal
    // focus is active. It does not bypass a source guard or alter telemetry.
    await activity.getByRole('textbox', {name: 'Live environment', exact: true}).evaluate(el => {el.value += '-draft'; el.dispatchEvent(new Event('input', {bubbles: true}))});
  });
  await activity.getByRole('textbox', {name: 'Live environment', exact: true}).fill(f.environment);
  // Apply makes a new stream epoch; only a newly inspected row can plot again.
  await activity.getByRole('button', {name: 'Apply scope', exact: true}).click();
  await expect(tail.locator('tbody>tr')).not.toHaveCount(0);
  await open();
  await observeOwnership('resume', async () => {
    await activity.getByRole('button', {name: 'Resume with recent logs', exact: true, includeHidden: true}).evaluate(el => el.click());
  });
  expect(state.errors).toEqual([]);
  const dir = path.resolve('artifacts/contextual-plots', f.environment);
  await fs.mkdir(dir, {recursive: true});
  await fs.writeFile(path.join(dir, 'ownership.json'), JSON.stringify({version: 1, ...f, observations, events: state.events, requests: state.requests}, null, 2));
});

test('181 Bombadil declares contextual numeric plots over ordinary telemetry', async ({request}) => {
  test.skip(process.env.GOO_CONTEXTUAL_PLOT_BOMBADIL !== '1', 'Optional serialized ordinary OTel campaign.');
  test.setTimeout(180000);
  const f = await fixture(request, 'spans');
  const dir = path.resolve('artifacts/contextual-plots', f.environment, 'bombadil');
  await fs.mkdir(dir, {recursive: true});
  await fs.writeFile(path.join(dir, 'fixture.json'), JSON.stringify(f, null, 2));
  const actor = await startBombadil({dir, url: base + '/' + questionLink(f.q), specification: 'verification/bombadil/contextual-plots.ts', seconds: 50});
  try {await actor.finish()} finally {await actor.stop()}
});

test('181 synthetic contextual owner rejects hidden and detached stale confirmations', async ({page}) => {
  page.setDefaultNavigationTimeout(45000);
  const state = observe(page);
  await page.goto(base + '/#definitions');
  await expect(page.locator('#workspaceTitle')).toHaveText('Definitions');
  const evidence = await page.evaluate(async ()=>{const GooNumericPlot=await import('/numeric-plot.mjs');
    const host = document.createElement('section');
    host.setAttribute('aria-label', 'Synthetic contextual declaration lifecycle fixture');
    const opener = document.createElement('button');
    opener.textContent = 'Synthetic Plot values opener';
    host.append(opener); document.body.append(host);
    const question = {signal: 'logs', environment: 'synthetic-owner-only', fromMs: 1000, toMs: 61000, limit: 100, offset: 0};
    const stages = [];
    const owner = GooNumericPlot.mount(host, {current: () => ({active: true, clean: true, question}), onStage: q => stages.push(q)});
    const field = {kind: 'attribute', key: 'work.amount'};
    function declare() {
      if (!owner.open(field, opener)) throw Error('Synthetic owner did not open');
      const d = host.querySelector('dialog');
      for (const [id, value, event] of [['Type', 'decimal64', 'change'], ['Unit', 'USD', 'input'], ['Scale', '2', 'input']]) {
        const input = d.querySelector('#plotValuesnumeric' + id);
        input.value = value; input.dispatchEvent(new Event(event, {bubbles: true}));
      }
      const confirm = [...d.querySelectorAll('button')].find(b => b.textContent === 'Open value distribution');
      if (confirm.disabled) throw Error('Valid synthetic declaration remained disabled');
      return {d, confirm};
    }
    const hidden = declare();
    const descriptor = Object.getOwnPropertyDescriptor(document, 'hidden');
    Object.defineProperty(document, 'hidden', {configurable: true, get: () => true});
    document.dispatchEvent(new Event('visibilitychange'));
    hidden.confirm.click();
    const hiddenClosed = !hidden.d.open;
    if (descriptor) Object.defineProperty(document, 'hidden', descriptor); else delete document.hidden;
    document.dispatchEvent(new Event('visibilitychange'));
    hidden.confirm.click();
    const noResurrection = !hidden.d.open && stages.length === 0;
    const detached = declare();
    host.remove();
    detached.confirm.click();
    const detachedNoStage = stages.length === 0;
    owner.dispose();
    return {fixture: 'Synthetic lifecycle only; no telemetry response is supplied.', hiddenClosed, noResurrection, detachedNoStage, stages};
  });
  expect(evidence).toMatchObject({hiddenClosed: true, noResurrection: true, detachedNoStage: true, stages: []});
  expect(state.reads).toBe(0);
  expect(state.errors).toEqual([]);
});

// The linked population owner must stage a single signal while retaining its companion.
test('181 population record and observed-field plots retain the two-signal parent', async ({page, request}) => {
  test.setTimeout(180000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request, 'spans');
  const logs = f.records.map(record => ({...record, signal: 'logs', id: randomBytes(8).toString('hex')}));
  const attributes = values => Object.entries(values).map(([key, stringValue]) => ({key, value: {stringValue}}));
  const resourceLogs = logs.map(record => ({
    resource: {attributes: attributes({'service.name': record.service, 'deployment.environment.name': f.environment,
      ...(record.resourceRaw === null ? {} : {'work.amount': record.resourceRaw})})},
    scopeLogs: [{scope: {name: 'ordinary.contextual.plot.worker'}, logRecords: [{
      timeUnixNano: String(BigInt(record.timeMs) * 1000000n), traceId: record.traceId, spanId: record.id,
      severityText: 'INFO', body: {stringValue: 'declared amount ' + record.id},
      attributes: attributes({team: 'core', ...(record.raw === null ? {} : {'work.amount': record.raw}),
        ...(record.scale1Raw === null ? {} : {'amount.scale1': record.scale1Raw}),
        ...(record.scale9Raw === null ? {} : {'amount.scale9': record.scale9Raw}),
        ...(record.integerRaw === null ? {} : {'work.items': record.integerRaw})})
    }]}]
  }));
  expect((await request.post('http://127.0.0.1:4318/v1/logs', {data: {resourceLogs}})).ok()).toBe(true);
  const logQuery = {...f.q, signal: 'logs'};
  await expect.poll(async () => (await json(await request.post(base + '/api/explore', {data: logQuery}))).count, {timeout: 45000}).toBe('400');
  const view = await json(await request.post(base + '/api/exploration-views', {data: {
    version: 1, title: 'Contextual plots retain companion signal', queries: [f.q, logQuery], layout: 'charts-records', columns: ['utc', 'service', 'summary', 'context']
  }}));
  const state = observe(page), records = [...f.records, ...logs], parentResults = [], children = [];
  page.on('response', async response => {
    if (isResponse('explore', response) && response.ok()) {
      const result = await response.json();
      if (parentResults.length < 2) parentResults.push(result);
    }
  });
  await page.goto(base + '/#exploration-view=' + view.id);
  const parent = page.locator('.population-workspace').first();
  await expect(parent).toHaveAttribute('data-phase', 'staged');
  expect(state.reads).toBe(0);
  state.event('population-stage');
  await parent.getByRole('button', {name: 'Run population view', exact: true}).click();
  await expect(parent).toHaveAttribute('data-phase', 'applied');
  expect(state.reads).toBe(2);
  await expect.poll(() => parentResults.length).toBe(2);
  for (const result of parentResults) assertResult(result, records);
  state.event('population-run');
  const parentScopes = async () => {
    const scopes = parent.locator('details').filter({has: page.getByText('Signal selectors and independent read scopes', {exact: true})});
    return JSON.parse(await scopes.locator(':scope>pre').textContent()).queries;
  };
  const originalQueries = await parentScopes();
  const companion = parent.locator('goo-table[aria-label="Captured logs inspection page"]');
  await companion.evaluate(element => {window.__plotCompanionTable = element});
  const spanTable = parent.locator('goo-table[aria-label="Captured spans inspection page"]');
  await spanTable.getByRole('button', {name: 'Inspect record', exact: true}).first().click();
  const recordOpener = plotButton(parent.locator('.record-detail:visible'), 'attribute');
  await recordOpener.click();
  await declare(page, plotField('attribute'));
  expect(state.reads).toBe(2);
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  let child = page.locator('.population-child .population-workspace');
  await expect(child).toHaveAttribute('data-phase', 'staged');
  const stagedChild = async () => JSON.parse(await child.locator('details').filter({has: page.getByText('Signal selectors and independent read scopes', {exact: true})}).locator(':scope>pre').textContent()).queries;
  let staged = await stagedChild();
  expect(staged).toHaveLength(1);
  assertTransform(f.q, staged[0], plotField('attribute'));
  expect(state.reads).toBe(2);
  state.event('population-record-stage');
  let pending = page.waitForResponse(response => isResponse('explore', response));
  await child.locator('.population-run').click();
  let result = await json(await pending);
  assertResult(result, records);
  expect(result.question).toEqual(staged[0]);
  children.push({kind: 'population-record', parent: f.q, field: plotField('attribute'), staged: staged[0], result});
  state.event('population-record-run');
  await page.getByRole('button', {name: 'Back to parent population', exact: true}).click();
  await expect(recordOpener).toBeFocused();
  expect(await parentScopes()).toEqual(originalQueries);
  expect(await companion.evaluate(element => element === window.__plotCompanionTable)).toBe(true);
  expect(state.reads).toBe(3);
  state.event('population-record-back');

  await parent.locator('.population-scope-context>summary').click();
  const editor = parent.locator('.population-filter-editor[data-signal="logs"]');
  await editor.locator(':scope>summary').click();
  const observed = editor.locator('.filter-observations');
  await observed.locator(':scope>summary').click();
  const field = observed.locator('details[data-field="resource:work.amount"]');
  await field.locator(':scope>summary').click();
  const observedOpener = plotButton(field, 'resource');
  await observedOpener.click();
  await declare(page, plotField('resource'));
  await dialog(page).getByRole('button', {name: 'Open value distribution', exact: true}).click();
  child = page.locator('.population-child .population-workspace');
  await expect(child).toHaveAttribute('data-phase', 'staged');
  staged = await stagedChild();
  expect(staged).toHaveLength(1);
  assertTransform(logQuery, staged[0], plotField('resource'));
  expect(state.reads).toBe(3);
  state.event('population-field-stage');
  pending = page.waitForResponse(response => isResponse('explore', response));
  await child.locator('.population-run').click();
  result = await json(await pending);
  assertResult(result, records);
  expect(result.question).toEqual(staged[0]);
  children.push({kind: 'population-field', parent: logQuery, field: plotField('resource'), staged: staged[0], result});
  state.event('population-field-run');
  await page.getByRole('button', {name: 'Back to parent population', exact: true}).click();
  await expect(observedOpener).toBeFocused();
  expect(await parentScopes()).toEqual(originalQueries);
  expect(await companion.evaluate(element => element === window.__plotCompanionTable)).toBe(true);
  expect(state.reads).toBe(4);
  state.event('population-field-back');
  expect(state.errors).toEqual([]);
  const dir = path.resolve('artifacts/contextual-plots', f.environment);
  await fs.mkdir(dir, {recursive: true});
  await fs.writeFile(path.join(dir, 'population.json'), JSON.stringify({version: 1, environment: f.environment, records, view, originalQueries, parentResults, children, companionRetained: true, events: state.events, requests: state.requests}, null, 2));
});
