const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const { questionLink } = require('../ui/workspace.mjs');
const { fixture, payload } = require('./184-trace-fixture');
const { startBombadil } = require('./support/cross-app');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });

async function json(response) {
  expect(response.ok(), await response.text()).toBe(true);
  return response.json();
}
function question(f, trace) {
  return { signal: 'spans', environment: f.environment, fromMs: f.fromMs,
    toMs: f.toMs, limit: 100, offset: 0,
    where: { all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: trace.traceId }] } };
}
async function seed(request) {
  const f = fixture();
  await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: payload(f) }));
  for (const trace of f.traces) {
    await expect.poll(async () => {
      const result = await json(await request.post(base + '/api/explore', { data: question(f, trace) }));
      return result.count;
    }, { timeout: 60000 }).toBe(String(trace.records.length));
  }
  return f;
}
function observe(page) {
  const requests = [], responses = [], events = [], errors = [];
  let source = 0;
  page.on('pageerror', e => errors.push(e.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (request.method() === 'POST' && ['/api/explore', '/api/captures'].includes(endpoint)) {
      source++;
      requests.push({ endpoint, payload: request.postDataJSON(), effect: 'source', ordinal: source });
    }
  });
  page.on('response', async response => {
    if (response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore' && response.ok()) {
      responses.push({ question: response.request().postDataJSON(), result: await response.json() });
    }
  });
  return { requests, responses, events, errors, get source() { return source; },
    event(action, extra = {}) { events.push({ action, source, ...extra }); } };
}
async function run(page, q) {
  await page.goto(base + '/' + questionLink(q));
  const pending = page.waitForResponse(r => new URL(r.url()).pathname === '/api/explore' && r.request().method() === 'POST');
  await page.locator('#exploreRun').click();
  const result = await json(await pending);
  await expect(page.locator('#exploreStatus')).toContainText(result.count + ' matching');
  return result;
}
async function expand(page, traceId) {
  const pending = page.waitForResponse(r => r.request().method() === 'POST' && new URL(r.url()).pathname === '/api/explore' && r.request().postDataJSON().signal === 'spans');
  await page.getByRole('button', { name: 'Expand trace ' + traceId, exact: true }).first().click();
  const result = await json(await pending);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  await expect(page.locator('#exploreTrace goo-trace-page-breakdown')).toHaveAttribute('data-state', /observed|partial/);
  return result;
}
async function snapshot(page, selector = '#exploreTrace goo-trace-page-breakdown') {
  return page.locator(selector).evaluate(async (element)=>{const GooTracePageBreakdown=await import('/trace-page-breakdown.mjs');return {
    input: element.data,
    model: GooTracePageBreakdown.analyze(element.data.rows, element.data.truncated),
    selectedIndex: element.selectedIndex, selectedGroupIndex: element.selectedGroupIndex,
    groups: [...element.querySelectorAll('tr[data-group-key]')].map(row => ({
      key: JSON.parse(row.dataset.groupKey), indexes: JSON.parse(row.dataset.rowIndexes),
      values: [...row.querySelectorAll('[data-measure]')].map(cell => ({
        measure: cell.dataset.measure, value: cell.dataset.value,
        known: cell.dataset.known, unknown: cell.dataset.unknown,
        bar: cell.querySelector('.comparison-bar-cell') ? {
          ...cell.querySelector('.comparison-bar-cell').dataset,
          width: cell.querySelector('.comparison-bar-fill')?.style.width ?? null,
          title: cell.querySelector('.comparison-value')?.title ?? '',
        } : null,
      })),
    })),
    diagnostics: [...element.querySelectorAll('tr[data-row-index]')].map(row => ({
      rowIndex: Number(row.dataset.rowIndex), state: row.dataset.state,
      reasons: JSON.parse(row.dataset.reasons),
      values: [...row.querySelectorAll('[data-measure]')].map(cell => ({ measure: cell.dataset.measure, value: cell.dataset.value })),
    })),
  }});
}
async function geometry(plot) {
  const value = await plot.evaluate(element => {
    const bounds = element.getBoundingClientRect();
    return { width: bounds.width,
      tracks: [...element.querySelectorAll('tr[data-group-key]')].map(row => [...row.querySelectorAll('.comparison-bar-track')].map(track => track.getBoundingClientRect().width)),
      labels: [...element.querySelectorAll('tr[data-group-key] > th')].map(cell => ({
        text: cell.textContent, left: cell.getBoundingClientRect().left - bounds.left,
        right: cell.getBoundingClientRect().right - bounds.left,
        width: cell.clientWidth, textWidth: cell.scrollWidth,
      })),
      diagnostics: [...element.querySelectorAll('tr[data-row-index] > td:nth-child(2)')].map(cell => ({
        width: cell.clientWidth, textWidth: cell.scrollWidth,
      })),
    };
  });
  for (const tracks of value.tracks) expect(Math.abs(tracks[0] - tracks[1])).toBeLessThanOrEqual(1);
  for (const label of value.labels) {
    expect(label.left).toBeGreaterThanOrEqual(0);
    expect(label.right).toBeLessThanOrEqual(value.width + 1);
    expect(label.textWidth).toBeLessThanOrEqual(label.width + 1);
  }
  for (const cell of value.diagnostics) {
    expect(cell.width).toBeGreaterThanOrEqual(160);
    expect(cell.textWidth).toBeLessThanOrEqual(cell.width + 1);
  }
  return value;
}
function storedBytes(id) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const env = { ...process.env };
  for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  return execFileSync('docker', ['--host=unix:///var/run/docker.sock', 'exec',
    process.env.GOO_QUERY_CONTAINER || 'grammar-observability-packages-1', 'cat',
    '/app/artifacts/investigations/exploration/' + id + '.json'], { env, maxBuffer: 8 * 1024 * 1024 });
}
async function output(f, name, data) {
  const dir = path.resolve('artifacts/trace-page-breakdown', f.environment);
  await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('184 artifact:', path.join(dir, name));
  return dir;
}
async function frame(page, locator, filename, width = 1440, height = 1800) {
  await page.setViewportSize({ width, height });
  await locator.scrollIntoViewIfNeeded();
  await page.evaluate(() => {
    const target = document.querySelector('#exploreTrace goo-trace-page-breakdown');
    if (target && target.getBoundingClientRect().top < 130) window.scrollBy(0, target.getBoundingClientRect().top - 150);
  });
  await locator.screenshot({ path: 'docs/journeys/screenshots/' + filename });
}

test('184 observed trace page separates inclusive intervals and direct-child union', async ({ page, request }) => {
  test.setTimeout(180000);
  const f = await seed(request), trace = f.traces[0], state = observe(page);
  const original = await run(page, question(f, trace)); state.event('run');
  const tracePage = await expand(page, trace.traceId); state.event('expand');
  expect(state.source).toBe(3); // search, explicit trace expansion, existing exact linked-log read
  const plot = page.locator('#exploreTrace goo-trace-page-breakdown');
  await plot.locator('details[data-section=groups] > summary').click();
  await plot.locator('details[data-section=diagnostics] > summary').click();
  const view = await snapshot(page), rootIndex = tracePage.rows.findIndex(r => r.SpanId === trace.records[0].spanId);
  expect(view.input.rows.map(r => r.SpanId)).toEqual(tracePage.rows.map(r => r.SpanId));
  expect(view.model.rows[rootIndex]).toMatchObject({ inclusiveNs: '100000000', childUnionNs: '70000000', outsideChildrenNs: '30000000' });
  expect([...new Set(view.groups.flatMap(group => group.values.map(value => value.bar.scaleMax)))]).toEqual(['100000000']);
  const zeroIndex = tracePage.rows.findIndex(r => r.SpanId === trace.records[4].spanId);
  expect(view.model.rows[zeroIndex]).toMatchObject({ inclusiveNs: '0', childUnionNs: '0', outsideChildrenNs: '0' });
  await plot.locator('[data-page-group]').first().click();
  const highlighted = await snapshot(page);
  const desktopGeometry = await geometry(plot);
  expect(await page.locator('#exploreTrace .waterfall-row.span-highlighted').count()).toBe(highlighted.model.groups[highlighted.selectedGroupIndex].rowIndexes.length);
  expect(state.source).toBe(3); state.event('group-highlight');
  await plot.locator('[data-page-group]').first().press('Enter');
  expect(await page.locator('#exploreTrace .waterfall-row.span-highlighted').count()).toBe(0);
  expect(state.source).toBe(3); state.event('group-clear');
  const logs = page.waitForResponse(r => new URL(r.url()).pathname === '/api/explore' && r.request().postDataJSON()?.signal === 'logs');
  await plot.getByRole('button', { name: 'Select supplied span row ' + (rootIndex + 1), exact: true }).click();
  const selectedLogs = await json(await logs);
  await expect(page.locator('.trace-page-span-summary')).toHaveAttribute('data-row-index', String(rootIndex));
  await expect(page.locator('.trace-page-span-summary')).toContainText('30 ms');
  expect(state.source).toBe(4); state.event('row-select');
  await plot.locator('details[data-section=diagnostics] > summary').click();
  await frame(page, plot, '184a-observed-page-latency.png');
  await frame(page, plot, '184b-mobile-observed-page-latency.png', 390, 2600);
  const mobileGeometry = await geometry(plot);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.locator('#exploreContextClose').click();
  await page.setViewportSize({ width: 1440, height: 1100 });
  const beforeCapture = state.source;
  await page.locator('#exploreSaveCapture').click();
  await page.locator('#captureDraftTitle').fill('Observed trace page timing');
  const pending = page.waitForResponse(r => new URL(r.url()).pathname === '/api/captures' && r.request().method() === 'POST');
  await page.locator('#captureDraftConfirm').click();
  const saved = await json(await pending); state.event('capture');
  expect(state.source).toBe(beforeCapture + 1);
  const before = storedBytes(saved.receipt.id);
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  const afterReopen = state.source;
  await page.reload();
  await expect(page.locator('#captureGeneralID')).toHaveValue(saved.receipt.id);
  await expect(page.locator('#savedexploreRows .record-row')).toHaveCount(5);
  expect(state.source).toBe(afterReopen); state.event('reopen');
  const after = storedBytes(saved.receipt.id);
  expect(after.equals(before)).toBe(true);
  expect(state.errors).toEqual([]);
  const dir = await output(f, 'ledger.json', { ...f, original, tracePage, view, highlighted, desktopGeometry, mobileGeometry,
    selectedLogs, saved, reopened, events: state.events, requests: state.requests, responses: state.responses,
    custody: { id: saved.receipt.id, bytes: before.length, sha256: createHash('sha256').update(before).digest('hex') } });
  await fs.writeFile(path.join(dir, 'scope-capture.json'), before);
  await fs.writeFile(path.join(dir, 'scope-capture.after.json'), after);
});

test('184 graph diagnostics and bounded pages preserve local uncertainty', async ({ page, request }) => {
  test.setTimeout(210000);
  const f = await seed(request), state = observe(page), cases = [];
  const reasons = { 'missing-parent': 'parent-outside-page', 'outside-parent': 'child-outside-parent',
    'grandchild-outside-child': 'child-outside-parent', 'duplicate-identity': 'duplicate-identity', 'parent-cycle': 'parent-cycle' };
  for (const trace of f.traces.slice(1)) {
    const original = await run(page, question(f, trace));
    const tracePage = await expand(page, trace.traceId);
    const plot = page.locator('#exploreTrace goo-trace-page-breakdown');
    await plot.locator('details[data-section=diagnostics] > summary').click();
    const view = await snapshot(page);
    if (trace.case === 'bounded-page') {
      expect(tracePage.count).toBe('126'); expect(tracePage.rows).toHaveLength(100);
      expect(tracePage.hasMore).toBe(true); expect(view.model.truncated).toBe(true);
      await expect(plot).toContainText('partial page');
      const nextResponse = page.waitForResponse(r => new URL(r.url()).pathname === '/api/explore' && r.request().postDataJSON()?.signal === 'spans');
      await page.locator('#exploreTraceMore').click();
      const next = await json(await nextResponse);
      await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
      const nextView = await snapshot(page);
      expect(next.question.offset).toBe(100); expect(next.rows).toHaveLength(26);
      expect(nextView.model.truncated).toBe(true);
      expect(nextView.input.rows.map(row => row.SpanId)).toEqual(next.rows.map(row => row.SpanId));
      expect(new Set([...tracePage.rows, ...next.rows].map(r => r.SpanId)).size).toBe(126);
      cases.push({ case: trace.case, original, tracePage, view, next, nextView });
    } else {
      expect(view.model.rows.some(row => row.reasons.includes(reasons[trace.case]))).toBe(true);
      if (trace.case === 'grandchild-outside-child') {
        const index = tracePage.rows.findIndex(r => r.SpanId === trace.records[0].spanId);
        expect(view.model.rows[index]).toMatchObject({ state: 'available', childUnionNs: '20000000', outsideChildrenNs: '80000000' });
      }
      cases.push({ case: trace.case, original, tracePage, view });
    }
    state.event(trace.case);
  }
  expect(state.errors).toEqual([]);
  await output(f, 'graphs.json', { ...f, cases, events: state.events, requests: state.requests });
});

test('184 stale or closed controls cannot mutate retained trace owners', async ({ page, request }) => {
  test.setTimeout(150000);
  const f = await seed(request), trace = f.traces[0], state = observe(page);
  await run(page, question(f, trace)); const tracePage = await expand(page, trace.traceId);
  const plot = page.locator('#exploreTrace goo-trace-page-breakdown');
  await page.evaluate(() => {
    const p = document.querySelector('#exploreTrace goo-trace-page-breakdown');
    window.old184 = { group: p.querySelector('[data-page-group]'), row: p.querySelector('[data-span-row]'),
      waterfall: document.querySelector('#exploreTrace .waterfall-row'), plot: p };
  });
  const initial = await snapshot(page), sourceBefore = state.source;
  await page.evaluate(() => { old184.group.click(); old184.row.click(); });
  expect(await snapshot(page)).toEqual(initial); expect(state.source).toBe(sourceBefore); state.event('closed-controls');
  await plot.locator('details[data-section=groups] > summary').click();
  await plot.locator('details[data-section=diagnostics] > summary').click();
  await page.goto(base + '/#definitions');
  const beforeURL = page.url();
  await page.evaluate(() => { old184.group.click(); old184.row.click(); old184.waterfall.click(); });
  const afterURL = page.url();
  expect(afterURL).toBe(beforeURL); expect(state.source).toBe(sourceBefore);
  const hidden = await page.evaluate(() => ({ connected: old184.plot.isConnected,
    selectedIndex: old184.plot.selectedIndex, selectedGroupIndex: old184.plot.selectedGroupIndex,
    highlights: [...old184.plot.parentElement.querySelectorAll('.span-highlighted')].length }));
  expect(hidden.selectedIndex).toBe(initial.selectedIndex); expect(hidden.selectedGroupIndex).toBe(initial.selectedGroupIndex);
  expect(hidden.highlights).toBe(0); state.event('hidden-controls', hidden);
  await page.goto(base + '/#explore');
  await expect(plot).toBeVisible();
  await plot.locator('[data-page-group]').first().click();
  expect(await page.locator('#exploreTrace .span-highlighted').count()).toBeGreaterThan(0);
  expect(state.source).toBe(sourceBefore); state.event('return-current');
  expect(state.errors).toEqual([]);
  await output(f, 'ownership.json', { ...f, tracePage, initial, hidden, beforeURL, afterURL,
    events: state.events, requests: state.requests });
});

test('184 optional actor explores supplied-page controls on real telemetry', async ({ request }) => {
  test.skip(process.env.GOO_TRACE_PAGE_BOMBADIL !== '1', 'Explicit ordinary-telemetry actor campaign');
  test.setTimeout(120000);
  const f = await seed(request), dir = await output(f, 'fixture.json', f);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'),
    url: base + '/' + questionLink(question(f, f.traces[0])),
    specification: 'verification/bombadil/trace-page-breakdown.ts', seconds: 55 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const trace = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = trace.flatMap(entry => entry.snapshots.filter(snapshot => snapshot.name === 'state').map(snapshot => snapshot.value));
  const phases = [...new Set(states.map(state => state.phase))];
  expect(phases).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8]);
  expect(states.at(-1)).toMatchObject({ phase: 8, reads: 4, rows: 5 });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
