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
    key: 'charge.amount'
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
  if (f.kind === 'resource') return f.key === 'declared.amount' ? r.resourceRaw : null;
  if (f.kind === 'column') return f.key === 'service' ? r.service : null;
  return {
    'charge.amount': r.raw,
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
  const environment = 'decimal-numerics-' + randomBytes(5).toString('hex'),
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
              'declared.amount': r.resourceRaw
            })
          })
        },
        attributes = attrs(Object.fromEntries([['charge.amount', r.raw], ['amount.scale1', r.scale1Raw], ['amount.scale9', r.scale9Raw], ['work.items', r.integerRaw]].filter(([, value]) => value !== null))),
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
  return {
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
    if (['/api/explore', '/api/captures', '/api/measurements', '/api/discovery', '/api/inventory'].includes(endpoint)) {
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
    writes: state.writes
  });
  return state;
}
async function execute(page, records) {
  const next = page.waitForResponse(r => isResponse('explore', r));
  await page.locator('#exploreRun').click();
  const result = await json(await next);
  assertResult(result, records);
  return result;
}
async function brush(page) {
  const section = page.locator('#numericSummary .numeric-distribution'),
    bar = section.locator('button[data-key="72"]');
  await bar.focus();
  await page.keyboard.press('Escape');
  await page.keyboard.press('Space');
  await page.keyboard.press('Shift+ArrowRight');
  await page.keyboard.press('Enter');
  return section;
}
function expectedHuman(summary, metric, scale) {
  const numerator = BigInt(summary[metric === 'mean' ? 'sum' : metric]),
    denominator = (metric === 'mean' ? BigInt(summary.valid) : 1n) * 10n ** BigInt(scale);
  return rational(numerator, denominator);
}
for (const signal of ['spans', 'logs']) test('180 ' + signal + ' exact decimal filters, summaries, ranges and comparisons survive capture and reuse', async ({
  page,
  request
}) => {
  test.setTimeout(420000);
  page.setDefaultTimeout(15000);
  page.setDefaultNavigationTimeout(45000);
  const f = await fixture(request, signal),
    state = observe(page),
    initial = {
      ...f.q,
      where: {
        field: binding.field,
        op: 'exists'
      }
    };
  await page.goto(base + '/' + questionLink(initial));
  await expect(page.locator('#exploreStatus')).toContainText('No evidence loaded');
  expect(state.reads).toBe(0);
  state.event('stage');
  await openMeasures(page);
  const numericEditor = page.locator('#measureComposer .numeric-binding-editor');
  if (!(await numericEditor.evaluate(e => e.open))) await numericEditor.locator(':scope>summary').click();
  await page.locator('#numericEnabled').check();
  await page.locator('#numericKey').fill(binding.field.key);
  await page.locator('#numericUnit').fill(binding.unit);
  await page.locator('#numericType').selectOption('decimal64');
  await expect(page.locator('#numericScale')).toHaveValue('');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  await page.locator('#numericScale').fill('2.0');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  await page.locator('#numericScale').fill('2');
  await page.locator('#numericHistogram').check();
  await openFilters(page);
  const editor = page.locator('#exploreFilterEditor');
  await editor.getByLabel('Condition 1 type', {
    exact: true
  }).selectOption('decimal64');
  await editor.getByLabel('Condition 1 unit', {
    exact: true
  }).fill('USD');
  await editor.getByLabel('Condition 1 operator', {
    exact: true
  }).selectOption('gte');
  await editor.getByLabel('Condition 1 value', {
    exact: true
  }).fill('1.20');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  await editor.getByLabel('Condition 1 scale', {
    exact: true
  }).fill('2');
  await expect(page.locator('#exploreRun')).toBeEnabled();
  expect(state.reads).toBe(0);
  state.event('author');
  const filtered = await execute(page, f.records);
  expect(filtered.question.numeric).toEqual(binding);
  expect(filtered.question.where).toEqual({
    numeric: binding,
    op: 'gte',
    value: '1.20'
  });
  state.event('filter-gte');
  await openFilters(page);
  await editor.getByLabel('Condition 1 scale', {
    exact: true
  }).fill('1');
  await expect(editor.getByLabel('Condition 1 value', {
    exact: true
  })).toHaveValue('1.20');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  const beforeRepair = state.reads;
  await editor.getByLabel('Condition 1 scale', {
    exact: true
  }).fill('2');
  await editor.getByLabel('Condition 1 value', {
    exact: true
  }).fill('1.200');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  await editor.getByLabel('Condition 1 value', {
    exact: true
  }).fill('-0.00');
  await expect(page.locator('#exploreRun')).toBeDisabled();
  await editor.getByLabel('Condition 1 value', {
    exact: true
  }).fill('-0.01');
  await editor.getByLabel('Condition 1 operator', {
    exact: true
  }).selectOption('lt');
  expect(state.reads).toBe(beforeRepair);
  state.event('invalid-and-repair');
  const negative = await execute(page, f.records);
  expect(negative.question.where.value).toBe('-0.01');
  state.event('filter-negative');
  await openFilters(page);
  await editor.getByLabel('Condition 1 namespace', {
    exact: true
  }).selectOption('resource');
  await editor.getByLabel('Condition 1 field', {
    exact: true
  }).fill('declared.amount');
  await editor.getByLabel('Condition 1 operator', {
    exact: true
  }).selectOption('gte');
  await editor.getByLabel('Condition 1 value', {
    exact: true
  }).fill('1.20');
  const resourceFilter = await execute(page, f.records);
  expect(resourceFilter.question.where.numeric).toEqual({
    ...binding,
    field: {
      kind: 'resource',
      key: 'declared.amount'
    }
  });
  state.event('filter-resource');
  await openFilters(page);
  await editor.getByLabel('Condition 1 namespace', {
    exact: true
  }).selectOption('attribute');
  await editor.getByLabel('Condition 1 field', {
    exact: true
  }).fill(binding.field.key);
  const coverage = [];
  for (const operator of ['valid', 'missing', 'empty', 'invalid']) {
    await editor.getByLabel('Condition 1 operator', {
      exact: true
    }).selectOption(operator);
    const result = await execute(page, f.records);
    expect(result.question.where).toEqual({
      numeric: binding,
      op: operator
    });
    coverage.push({
      operator,
      result
    });
    state.event('filter-' + operator);
    await openFilters(page);
  }
  expect(coverage.find(x => x.operator === 'valid').result.rows).toHaveLength(100);
  expect(coverage.find(x => x.operator === 'invalid').result.numeric.state).toBe('no_valid_values');
  const parentQuery = {
      ...f.q,
      numeric: binding,
      numericHistogram: {
        binning: 'powers-of-two'
      }
    },
    beforeParent = state.reads;
  await page.goto(base + '/' + questionLink(parentQuery));
  await page.reload();
  await expect(page.locator('#exploreRun')).toBeEnabled();
  expect(state.reads).toBe(beforeParent);
  state.event('parent-stage');
  const result = await execute(page, f.records);
  expect(result.count).toBe('400');
  expect(result.rows).toHaveLength(100);
  expect(result.numeric).toMatchObject({
    scale: 2,
    valid: '160',
    missing: '10',
    empty: '10',
    invalid: '220'
  });
  expect(result.numeric.min).toBe(String(lower));
  expect(result.numeric.max).toBe(String(upper - 1n));
  await expect(page.locator('#numericSummary')).toContainText(human(upper - 1n, 2) + ' USD');
  await expect(page.locator('#numericSummary')).toContainText(human(lower, 2) + ' USD');
  state.event('parent-run');
  const section = await brush(page),
    view = await section.locator('goo-time-bars').evaluate(e => e.data);
  expect(view.points.some(p => p.label.includes('[1.28, 2.56)'))).toBe(true);
  expect(state.reads).toBe(beforeParent + 1);
  state.event('brush');
  await page.setViewportSize({
    width: 1440,
    height: 1500
  });
  await section.evaluate(e => {
    e.scrollIntoView({
      block: 'start'
    });
    window.scrollBy(0, -120);
  });
  await section.screenshot({
    path: 'docs/journeys/screenshots/180a-' + signal + '-decimal-range.png'
  });
  const selectedPlan = page.waitForResponse(r => isResponse('numeric-histograms/selection', r)),
    selectedRead = page.waitForResponse(r => isResponse('explore', r));
  await section.getByRole('button', {
    name: 'Run selected population',
    exact: true
  }).click();
  const selectionResponse = await selectedPlan,
    selection = {
      request: selectionResponse.request().postDataJSON(),
      response: await json(selectionResponse)
    },
    selected = await json(await selectedRead);
  assertResult(selected, f.records);
  expect(selection.response.fromBin).toBe(72);
  expect(selection.response.toBin).toBe(74);
  expect(selection.response.from).toBe('128');
  expect(selection.response.to).toBe('512');
  expect(selected.question).toEqual(selection.response.selected);
  const serialized = JSON.stringify(selected.question.where);
  expect(serialized).toContain('1.28');
  expect(serialized).toContain('5.12');
  expect(serialized).not.toContain('"value":"128"');
  state.event('selected-run');
  const rest = await json(await request.post(base + '/api/explore', {
    data: selection.response.rest
  }));
  assertResult(rest, f.records);
  expect(BigInt(selected.count) + BigInt(rest.count)).toBe(400n);
  expect(rest.numeric.invalid).toBe(result.numeric.invalid);
  expect(rest.numeric.missing).toBe(result.numeric.missing);
  expect(rest.numeric.empty).toBe(result.numeric.empty);
  const beforeBack = {
    reads: state.reads,
    plans: state.plans
  };
  await page.getByRole('button', {
    name: 'Back to question',
    exact: true
  }).click();
  const compareBrush = await brush(page);
  expect({
    reads: state.reads,
    plans: state.plans
  }).toEqual(beforeBack);
  state.event('back-and-brush');
  const partitionResponse = page.waitForResponse(r => isResponse('numeric-histograms/selection', r));
  await compareBrush.getByRole('button', {
    name: 'Stage numeric comparison with rest',
    exact: true
  }).click();
  const plan = await json(await partitionResponse);
  expect(plan.fromBin).toBe(72);
  expect(plan.toBin).toBe(74);
  expect(plan.numericComparison.selection.numeric).toEqual(binding);
  expect(plan.numericComparison.baseline.numeric).toEqual(binding);
  expect(state.reads).toBe(beforeBack.reads);
  state.event('comparison-stage');
  const sourceResults = [];
  const collect = async response => {
    if (isResponse('explore', response) && response.ok()) sourceResults.push(await response.json());
  };
  page.on('response', collect);
  const compared = page.waitForResponse(r => isResponse('compare', r));
  await page.locator('.population-run').click();
  const cr = await compared;
  await expect.poll(() => sourceResults.length).toBe(2);
  page.off('response', collect);
  const partition = {
    plan,
    selected: sourceResults.find(x => JSON.stringify(x.question.where) === JSON.stringify(plan.numericComparison.selection.where)),
    rest: sourceResults.find(x => JSON.stringify(x.question.where) === JSON.stringify(plan.numericComparison.baseline.where)),
    request: cr.request().postDataJSON(),
    response: await json(cr)
  };
  assertResult(partition.selected, f.records);
  assertResult(partition.rest, f.records);
  assertPair(partition, f.records);
  state.event('comparison-run');
  const comparison = page.locator('.population-breakdown .numeric-comparison');
  await expect(comparison).toHaveAttribute('data-numeric-state', 'comparable');
  const signed = await comparison.locator('.signed-comparison tbody>tr').evaluateAll(rows => rows.map(row => [...row.querySelectorAll('.signed-comparison-cell')].map(cell => ({
    value: cell.dataset.value,
    exact: cell.querySelector('.comparison-value')?.getAttribute('title'),
    text: cell.querySelector('.comparison-value')?.textContent
  }))));
  expect(signed).toHaveLength(4);
  for (const [index, metric] of ['mean', 'sum', 'min', 'max'].entries()) for (const [side, summary] of [partition.rest.numeric, partition.selected.numeric].entries()) {
    const expected = expectedHuman(summary, metric, 2);
    const [n, d = '1'] = signed[index][side].value.split('/');
    expect(rational(BigInt(n), BigInt(d))).toBe(expected);
    expect(signed[index][side].exact).toBe(signed[index][side].value + ' USD (exact)');
  }
  await page.setViewportSize({
    width: 390,
    height: 2200
  });
  const metricFigure = comparison.locator('goo-metric-comparison');
  await metricFigure.evaluate(element => {
    element.scrollIntoView({
      block: 'start'
    });
    window.scrollBy(0, -120);
  });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  const mobileRows = await metricFigure.locator('tbody > tr').evaluateAll(rows => rows.map(row => {
    const figure = row.closest('figure').getBoundingClientRect();
    const axis = row.querySelector('.comparison-row-axis');
    return {
      axis: axis.textContent,
      axisFits: axis.scrollWidth <= axis.clientWidth + 1 && axis.scrollHeight <= axis.clientHeight + 1,
      cells: [...row.querySelectorAll('.signed-comparison-cell')].map(cell => {
        const bounds = cell.getBoundingClientRect(),
          text = cell.querySelector('.comparison-value');
        return {
          inside: bounds.left >= figure.left - 1 && bounds.right <= figure.right + 1 && bounds.top >= figure.top - 1 && bounds.bottom <= figure.bottom + 1,
          inViewport: bounds.left >= 0 && bounds.right <= innerWidth + 1 && bounds.top >= 80 && bounds.bottom <= innerHeight,
          noClip: text.scrollWidth <= text.clientWidth + 1 && text.scrollHeight <= text.clientHeight + 1,
          exact: text.title,
          text: text.textContent
        };
      })
    };
  }));
  expect(mobileRows).toHaveLength(4);
  for (const row of mobileRows) {
    expect(row.axisFits).toBe(true);
    expect(row.cells).toHaveLength(2);
    for (const cell of row.cells) {
      expect(cell.inside).toBe(true);
      expect(cell.inViewport).toBe(true);
      expect(cell.noClip).toBe(true);
    }
  }
  expect(mobileRows[1].axis).toContain('10719.3');
  expect(mobileRows[1].axis).not.toContain('1071930/100');
  await metricFigure.screenshot({
    path: 'docs/journeys/screenshots/180b-' + signal + '-decimal-comparison.png'
  });
  await page.setViewportSize({
    width: 1440,
    height: 1100
  });
  // Capture and grammar reuse preserve explicit scale; capture rereads are named.
  await page.goto(base + '/' + questionLink(parentQuery));
  const parentAgain = await execute(page, f.records);
  await page.locator('#exploreSaveCapture').click();
  await page.locator('#captureDraftTitle').fill('Exact ' + signal + ' decimal amount evidence');
  let response = page.waitForResponse(r => isResponse('captures', r));
  await page.locator('#captureDraftConfirm').click();
  const saved = await json(await response);
  expect(saved.explore.numeric).toEqual(result.numeric);
  expect(saved.explore.numericHistogram).toEqual(result.numericHistogram);
  expect(saved.receipt.question.explore.numeric).toEqual(binding);
  state.event('capture');
  const beforeReopen = state.reads;
  await page.reload();
  await expect(page.locator('#savednumericSummary')).toContainText(human(upper - 1n, 2) + ' USD');
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  expect(state.reads).toBe(beforeReopen);
  state.event('reopen');
  await page.locator('#savedexploreCreatePackage').click();
  await page.locator('#packageName').fill(f.environment);
  await page.locator('#packageTitle').fill('Reusable exact decimal amount');
  await page.locator('#packageLookback').fill('5');
  response = page.waitForResponse(r => isResponse('telemetry-packages/draft', r));
  await page.locator('#packageBuild').click();
  const draft = await json(await response);
  expect(draft.question.numeric).toEqual(binding);
  await expect(page.locator('#packageYAML')).toHaveValue(/decimal64/);
  await expect(page.locator('#packageYAML')).toHaveValue(/scale: 2/);
  response = page.waitForResponse(r => isResponse('telemetry-packages', r));
  await page.locator('#packageSave').click();
  const reusable = await json(await response);
  expect(state.reads).toBe(beforeReopen);
  state.event('package-save');
  response = page.waitForResponse(r => isResponse('explore', r), {
    timeout: 45000
  });
  await page.locator('#packageUse').click();
  const reused = await json(await response);
  assertResult(reused, f.records);
  expect(reused.question.numeric).toEqual(binding);
  expect(reused.numeric).toEqual(result.numeric);
  expect(reused.numericHistogram).toEqual(result.numericHistogram);
  state.event('package-use');
  const edgeResults = [];
  for (const [kind, numeric] of [['scale1', {
    ...binding,
    scale: 1,
    field: {
      kind: 'attribute',
      key: 'amount.scale1'
    }
  }], ['scale9', {
    ...binding,
    scale: 9,
    field: {
      kind: 'attribute',
      key: 'amount.scale9'
    }
  }], ['resource', {
    ...binding,
    field: {
      kind: 'resource',
      key: 'declared.amount'
    }
  }], ['integer', {
    field: {
      kind: 'attribute',
      key: 'work.items'
    },
    type: 'int64',
    unit: 'items',
    parse: 'canonical-decimal'
  }]]) {
    const actual = await json(await request.post(base + '/api/explore', {
      data: {
        ...parentQuery,
        numeric
      }
    }));
    assertResult(actual, f.records);
    edgeResults.push({
      kind,
      result: actual
    });
  }
  const lastBinPlan = await json(await request.post(base + '/api/numeric-histograms/selection', {
    data: {
      parent: parentQuery,
      fromBin: 127,
      toBin: 128
    }
  }));
  const lastBinResult = await json(await request.post(base + '/api/explore', {
    data: lastBinPlan.selected
  }));
  assertResult(lastBinResult, f.records);
  expect(lastBinResult.numeric.max).toBe(String(upper - 1n));
  expect(JSON.stringify(lastBinPlan.selected.where)).not.toContain(human(upper, 2));
  const refusals = [];
  for (const [kind, numeric] of [['missing-scale', {
    ...binding,
    scale: undefined
  }], ['scale-zero', {
    ...binding,
    scale: 0
  }], ['scale-ten', {
    ...binding,
    scale: 10
  }], ['integer-scale', {
    ...binding,
    type: 'int64'
  }]]) {
    const actual = await request.post(base + '/api/explore', {
      data: {
        ...parentQuery,
        numeric
      }
    });
    expect(actual.ok()).toBe(false);
    refusals.push({
      kind,
      status: actual.status(),
      response: await actual.json()
    });
  }
  const mismatch = structuredClone(partition.request);
  mismatch.baseline.question.numeric.scale = 3;
  const mismatchResponse = await request.post(base + '/api/compare', {
    data: mismatch
  });
  expect(mismatchResponse.ok()).toBe(false);
  refusals.push({
    kind: 'scale-mismatch',
    status: mismatchResponse.status(),
    response: await mismatchResponse.json()
  });
  expect(state.errors).toEqual([]);
  const dir = path.resolve('artifacts/decimal-numerics', f.environment);
  await fs.mkdir(dir, {
    recursive: true
  });
  await fs.writeFile(path.join(dir, 'ledger.json'), JSON.stringify({
    version: 1,
    ...f,
    binding,
    initial,
    filtered,
    negative,
    resourceFilter,
    coverage,
    parentQuery,
    result,
    view,
    selection,
    selected,
    rest,
    partition,
    signed,
    mobileRows,
    parentAgain,
    saved,
    reopened,
    draft,
    reusable,
    reused,
    edgeResults,
    lastBinPlan,
    lastBinResult,
    refusals,
    events: state.events,
    requests: state.requests
  }, null, 2));
});
test('180 Bombadil explores an exact decimal population and a retained numeric range', async ({
  request
}) => {
  test.skip(process.env.GOO_DECIMAL_BOMBADIL !== '1', 'Optional serialized ordinary OTel campaign.');
  test.setTimeout(180000);
  const fixtureData = await fixture(request, 'spans');
  const query = {
    ...fixtureData.q,
    numeric: binding,
    numericHistogram: {
      binning: 'powers-of-two'
    }
  };
  const directory = path.resolve('artifacts/decimal-numerics', fixtureData.environment, 'bombadil');
  await fs.mkdir(directory, {
    recursive: true
  });
  await fs.writeFile(path.join(directory, 'fixture.json'), JSON.stringify({
    ...fixtureData,
    binding,
    query
  }, null, 2));
  const actor = await startBombadil({
    dir: directory,
    url: base + '/' + questionLink(query),
    specification: 'verification/bombadil/decimal-numerics.ts',
    seconds: 55
  });
  try {
    await actor.finish();
  } finally {
    await actor.stop();
  }
});
