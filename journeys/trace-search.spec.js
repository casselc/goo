const { test, expect } = require('@playwright/test');
const { questionLink } = require('../ui/workspace');
const { base, json, query, assertResult, seed, observe, output } = require('./186-trace-search-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const panel = page => page.locator('.trace-search:visible');
async function apply(page, q) {
  await page.goto(base + '/' + questionLink(q));
  const pending = page.waitForResponse(response => new URL(response.url()).pathname === '/api/explore' && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const result = await json(await pending);
  await expect(page.locator('#exploreStatus')).toContainText(result.count + ' matching');
  return result;
}
async function openSearch(page) {
  await page.locator('#exploreRows .record-row').first().getByRole('button', { name: 'Inspect record', exact: true }).click();
  await page.getByRole('button', { name: 'Find traces with this operation', exact: true }).click();
  await expect(panel(page)).toBeVisible();
}
async function addErrorCondition(page) {
  await panel(page).getByRole('button', { name: 'Add span condition', exact: true }).click();
  const condition = panel(page).locator('.trace-search-condition').nth(1);
  await condition.locator('input[aria-label^="Condition "][aria-label$=" field"]').fill('status_class');
  await condition.locator('select[aria-label^="Condition "][aria-label$=" operator"]').selectOption('eq');
  await condition.locator('input[aria-label^="Condition "][aria-label$=" value"]').fill('error');
}
async function search(page) {
  const pending = page.waitForResponse(response => new URL(response.url()).pathname === '/api/trace-search' && response.request().method() === 'POST');
  await panel(page).getByRole('button', { name: 'Run trace search', exact: true }).click();
  const response = await pending, result = await json(response), request = response.request().postDataJSON();
  await expect(panel(page).locator(':scope > [role=status]')).toContainText(result.matchedTraces + ' distinct reported trace IDs');
  return { request, result };
}
async function witnesses(page) {
  return panel(page).evaluate(element => ({
    status: element.querySelector(':scope > [role=status]').textContent,
    scope: [...element.querySelectorAll(':scope > p')].map(value => value.textContent),
    rows: [...element.querySelectorAll('tr[data-trace-id]')].map(row => ({ traceId: row.dataset.traceId,
      witnesses: [...row.querySelectorAll('details[data-condition-index]')].map(details => ({
        conditionIndex: Number(details.dataset.conditionIndex), summary: details.querySelector('summary').textContent,
        text: details.querySelector('p').textContent, title: details.querySelector('p').title,
        fields: [...details.querySelectorAll('tbody tr')].map(field => ({
          key: field.querySelector('th').textContent, state: field.querySelector('goo-field-value').state,
          value: field.querySelector('goo-field-value').value,
        })),
      })),
    })),
  }));
}
async function trace(page, traceId) {
  const pending = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore' && response.request().postDataJSON().signal === 'spans');
  await panel(page).locator('tr[data-trace-id="' + traceId + '"]').getByRole('button', { name: 'Open trace (fresh read)', exact: true }).click();
  const result = await json(await pending);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  await expect(page.locator('#exploreContextSummary')).toContainText(result.count + ' captured spans');
  return result;
}
function assertTraceScope(result, searchRequest, traceId) {
  expect(result.question.environment).toBe(searchRequest.environment);
  expect(result.question.fromMs).toBe(searchRequest.fromMs); expect(result.question.toMs).toBe(searchRequest.toMs);
  expect(result.question.where).toEqual({ all: [{ field: { kind: 'column', key: 'trace_id' }, op: 'eq', value: traceId }] });
}
async function shot(page, target, name, width = 1440, height = 2000) {
  await page.setViewportSize({ width, height }); await target.scrollIntoViewIfNeeded();
  const initial = await target.boundingBox();
  if (initial.y < 150 && initial.height < height - 180) await page.evaluate(top => window.scrollBy(0, top - 150), initial.y);
  const bounds = await target.boundingBox();
  expect(bounds.x).toBeGreaterThanOrEqual(0); expect(bounds.x + bounds.width).toBeLessThanOrEqual(width + 1);
  expect(bounds.y).toBeGreaterThanOrEqual(0); expect(bounds.y + bounds.height).toBeLessThanOrEqual(height + 1);
  await target.screenshot({ path: 'docs/journeys/screenshots/' + name });
}

async function controlsFrame(page) {
  await page.setViewportSize({ width: 1440, height: 2600 });
  await panel(page).evaluate(element => element.scrollIntoView({ block: 'start' }));
  await page.evaluate(() => {
    const top = document.querySelector('.trace-search').getBoundingClientRect().top;
    if (top < 150) window.scrollBy(0, top - 150);
  });
  const geometry = await panel(page).evaluate(element => {
    const bounds = element.getBoundingClientRect(), first = element.querySelector('tr[data-trace-id]').getBoundingClientRect();
    return { left: bounds.left, top: bounds.top, width: bounds.width, panelHeight: bounds.height,
      controlsToFirstResult: first.top - bounds.top, firstBottom: first.bottom };
  });
  expect(geometry.top).toBeGreaterThanOrEqual(0); expect(geometry.firstBottom).toBeLessThanOrEqual(2600);
  await page.screenshot({ path: 'docs/journeys/screenshots/186d-controls-to-first-result.png',
    clip: { x: geometry.left, y: geometry.top, width: geometry.width, height: geometry.firstBottom - geometry.top } });
  return geometry;
}

test('186 find trace-wide span conditions and continue into original-window trace evidence', async ({ page, request }) => {
  test.setTimeout(180000);
  const f = await seed(request), state = observe(page);
  const parent = await apply(page, query(f)); state.event('parent-run');
  expect(parent.count).toBe('135'); expect(parent.rows).toHaveLength(100);
  const beforeStage = state.network.length;
  await openSearch(page); await addErrorCondition(page);
  expect(state.source).toBe(1); expect(state.network.length).toBe(beforeStage); state.event('stage');
  const observed = await search(page); assertResult(f, observed.request, observed.result); state.event('search');
  expect(observed.result.matchedTraces).toBe('15'); expect(state.source).toBe(2);
  const layout = await controlsFrame(page);
  const cross = f.traces.find(trace => trace.role === 'cross-service'), same = f.traces.find(trace => trace.role === 'same-span'), repeated = f.traces.find(trace => trace.role === 'repeated-witness');
  const crossItem = observed.result.items.find(item => item.traceId === cross.traceId);
  expect(crossItem.witnesses.map(value => value.service)).toEqual(['db-service', 'queue-worker']);
  expect(crossItem.witnesses[0].spanId).not.toBe(crossItem.witnesses[1].spanId);
  const sameItem = observed.result.items.find(item => item.traceId === same.traceId);
  expect(sameItem.witnesses[0].spanId).toBe(sameItem.witnesses[1].spanId);
  const repeatItem = observed.result.items.find(item => item.traceId === repeated.traceId);
  expect(repeatItem.witnesses.map(value => value.matchingRecords)).toEqual(['120', '120']);
  for (const control of f.traces.filter(trace => ['condition-a-only', 'condition-b-only', 'fast-database-error', 'wrong-environment', 'outside-window'].includes(trace.role))) {
    expect(observed.result.items.some(item => item.traceId === control.traceId)).toBe(false);
  }
  const crossRow = panel(page).locator('tr[data-trace-id="' + cross.traceId + '"]');
  for (const details of await crossRow.locator('details').all()) await details.locator(':scope > summary').click();
  const dom = await witnesses(page); expect(state.source).toBe(2); state.event('witness-details');
  await shot(page, crossRow, '186a-cross-service-condition-witnesses.png');
  const sameRow = panel(page).locator('tr[data-trace-id="' + same.traceId + '"]');
  for (const details of await sameRow.locator('details').all()) await details.locator(':scope > summary').click();
  await shot(page, sameRow, '186b-mobile-same-span-witnesses.png', 390, 1800);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.setViewportSize({ width: 1440, height: 1800 });
  const boundedRequest = { ...observed.request, limit: 2 };
  const bounded = await json(await request.post(base + '/api/trace-search', { data: boundedRequest }));
  assertResult(f, boundedRequest, bounded); expect(bounded.hasMore).toBe(true);
  const emptyRequest = { ...observed.request, conditions: [observed.request.conditions[0], { field: { kind: 'column', key: 'service' }, op: 'eq', value: f.environment + '-absent-service' }] };
  const empty = await json(await request.post(base + '/api/trace-search', { data: emptyRequest }));
  assertResult(f, emptyRequest, empty); expect(empty.state).toBe('no_data');
  const reference = await trace(page, cross.traceId); assertTraceScope(reference, observed.request, cross.traceId);
  expect(reference.count).toBe('3'); state.event('open-cross'); expect(state.source).toBe(4);
  const tray = page.locator('#exploreContext .trace-reference-tray');
  await tray.getByRole('button', { name: 'Pin trace page as reference', exact: true }).click();
  expect(state.source).toBe(4); state.event('pin-cross');
  await page.locator('#exploreContextClose').click();
  const current = await trace(page, same.traceId); assertTraceScope(current, observed.request, same.traceId);
  expect(current.count).toBe('2'); state.event('open-same'); expect(state.source).toBe(6);
  const beforeCompare = state.network.length;
  await tray.getByRole('button', { name: 'Compare with reference', exact: true }).click();
  const comparison = await page.locator('goo-trace-comparison:visible').evaluate(element => ({ input: element.data,
    model: GooTracePageComparison.compare(element.data.baseline.rows, element.data.selection.rows, false, false) }));
  expect(comparison.input.baseline.rows).toEqual(reference.rows); expect(comparison.input.selection.rows).toEqual(current.rows);
  expect(comparison.model.groups.find(group => group.operation === 'GET /checkout')).toMatchObject({ inclusiveDifferenceNs: '100000000', outsideDifferenceNs: '80000000' });
  expect(state.network.length).toBe(beforeCompare); state.event('compare');
  await shot(page, page.locator('goo-trace-comparison:visible'), '186c-matching-trace-page-comparison.png', 1440, 2200);
  await page.locator('#exploreContextClose').click();
  const repeatedPage = await trace(page, repeated.traceId); assertTraceScope(repeatedPage, observed.request, repeated.traceId);
  expect(repeatedPage.count).toBe('241'); expect(repeatedPage.rows).toHaveLength(100); expect(repeatedPage.hasMore).toBe(true);
  expect(repeatedPage.rows.some(row => row.SpanId === repeatItem.witnesses[0].spanId)).toBe(true);
  expect(repeatedPage.rows.some(row => row.SpanId === repeatItem.witnesses[1].spanId)).toBe(false);
  state.event('open-repeated'); expect(state.source).toBe(8);
  const nextResponse = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore' && response.request().postDataJSON().signal === 'spans');
  await page.locator('#exploreTraceMore').click(); const repeatedNext = await json(await nextResponse);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  expect(repeatedNext.question.offset).toBe(100); expect(repeatedNext.rows).toHaveLength(100);
  expect(repeatedNext.rows.some(row => row.SpanId === repeatItem.witnesses[1].spanId)).toBe(true);
  expect(new Set([...repeatedPage.rows, ...repeatedNext.rows].map(row => row.SpanId)).size).toBe(200);
  expect(state.source).toBe(10); state.event('next-trace-page'); expect(state.errors).toEqual([]);
  await output(f, 'ledger.json', { ...f, parent, observed, dom, layout, bounded: { request: boundedRequest, result: bounded },
    empty: { request: emptyRequest, result: empty }, directSourceReads: 2,
    reference, current, comparison, repeatedPage, repeatedNext,
    events: state.events, requests: state.requests, responses: state.responses, network: state.network });
});

async function ownerDOM(page) {
  return page.evaluate(() => {
    const element = document.querySelector('.trace-search');
    return { hidden: element.hidden, connected: element.isConnected,
      rows: element.querySelectorAll('tr[data-trace-id]').length,
      conditions: element.querySelectorAll('.trace-search-condition').length,
      status: element.querySelector(':scope > [role=status]').textContent,
      runDisabled: [...element.querySelectorAll('button')].find(button => button.textContent === 'Run trace search').disabled,
      cancelVisible: ![...element.querySelectorAll('button')].find(button => button.textContent === 'Cancel trace search').hidden };
  });
}

test('186 closed, hidden and superseded trace searches cannot publish stale evidence', async ({ page, request }) => {
  test.setTimeout(180000);
  const f = await seed(request), state = observe(page);
  const parent = await apply(page, query(f)); await openSearch(page); await addErrorCondition(page);
  const initial = await search(page); assertResult(f, initial.request, initial.result); state.event('initial', { observed: await ownerDOM(page) });
  const closedCondition = panel(page).locator('.trace-search-condition').nth(1);
  const closedInput = closedCondition.locator('input[aria-label^="Condition "][aria-label$=" value"]');
  await closedInput.evaluate(input => { input.value = 'ok'; input.dispatchEvent(new Event('input', { bubbles: true })); });
  await expect(panel(page).locator('tr[data-trace-id]')).toHaveCount(15);
  expect(state.source).toBe(2); state.event('closed-editor', { observed: await ownerDOM(page) });
  const closedEditorRerun = await search(page); assertResult(f, closedEditorRerun.request, closedEditorRerun.result);
  expect(closedEditorRerun.request).toEqual(initial.request);
  expect(state.source).toBe(3); state.event('closed-editor-rerun', { observed: await ownerDOM(page) });
  await closedCondition.getByText('Edit span condition 2', { exact: true }).click();
  await expect(closedCondition.locator('input[aria-label^="Condition "][aria-label$=" value"]')).toHaveValue('error');

  await page.evaluate(() => {
    const panel = document.querySelector('.trace-search');
    window.retainedSearch186 = { panel, run: [...panel.querySelectorAll('button')].find(button => button.textContent === 'Run trace search'),
      open: panel.querySelector('tr[data-trace-id] button'), add: [...panel.querySelectorAll('button')].find(button => button.textContent === 'Add span condition') };
  });
  await panel(page).getByRole('button', { name: 'Close trace search', exact: true }).click();
  await page.evaluate(() => { retainedSearch186.run.click(); retainedSearch186.open.click(); retainedSearch186.add.click(); });
  expect(state.source).toBe(3); expect(await page.locator('.trace-search:visible').count()).toBe(0); state.event('closed', { observed: await ownerDOM(page) });
  await openSearch(page); await addErrorCondition(page);
  await page.goto(base + '/#definitions'); const hiddenURL = page.url();
  await page.evaluate(() => { retainedSearch186.run.click(); retainedSearch186.open.click(); retainedSearch186.add.click(); });
  expect(page.url()).toBe(hiddenURL); expect(state.source).toBe(3); state.event('hidden', { observed: await ownerDOM(page) });
  await page.goto(base + '/#explore'); await openSearch(page); await addErrorCondition(page);
  let release, ready, finished;
  const held = new Promise(resolve => { ready = resolve; }), delivery = new Promise(resolve => { release = resolve; }), settled = new Promise(resolve => { finished = resolve; });
  let heldRequest, heldResult, deliveryError = null;
  await page.route('**/api/trace-search', async route => {
    heldRequest = route.request().postDataJSON();
    const response = await route.fetch(); heldResult = await response.json(); ready(); await delivery;
    try { await route.fulfill({ response }); } catch (error) { deliveryError = error.message; }
    finished();
  });
  await panel(page).getByRole('button', { name: 'Run trace search', exact: true }).click(); await held;
  assertResult(f, heldRequest, heldResult); state.event('held', { observed: await ownerDOM(page) }); expect(state.source).toBe(4);
  const condition = panel(page).locator('.trace-search-condition').nth(1);
  const editorDisclosure = condition.locator('details[data-condition-editor]');
  if (!await editorDisclosure.evaluate(element => element.open)) await editorDisclosure.locator(':scope > summary').click();
  await condition.locator('input[aria-label^="Condition "][aria-label$=" value"]').fill('ok');
  release(); await settled; await page.unroute('**/api/trace-search');
  await expect(panel(page).locator('tr[data-trace-id]')).toHaveCount(0);
  expect(state.source).toBe(4); state.event('superseded', { observed: await ownerDOM(page) });
  await condition.locator('input[aria-label^="Condition "][aria-label$=" value"]').fill('error');
  const fresh = await search(page); assertResult(f, fresh.request, fresh.result);
  expect(state.source).toBe(5); state.event('fresh', { observed: await ownerDOM(page) }); expect(state.errors).toEqual([]);
  await output(f, 'ownership.json', { ...f, parent, initial, closedEditorRerun, hiddenURL, heldRequest, heldResult, deliveryError, fresh,
    events: state.events, requests: state.requests, responses: state.responses, network: state.network });
});

test('186 optional real trace-search Bombadil campaign', async ({ request }) => {
  test.skip(process.env.GOO_TRACE_SEARCH_BOMBADIL !== '1', 'Explicit ordinary-telemetry campaign');
  test.setTimeout(150000);
  const fs = require('node:fs/promises'), path = require('node:path');
  const { startBombadil } = require('./support/cross-app');
  const f = await seed(request), dir = await output(f, 'fixture.json', f);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'),
    url: base + '/' + questionLink(query(f)), specification: 'verification/bombadil/trace-search.ts', seconds: 55 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(snapshot => snapshot.name === 'state').map(snapshot => snapshot.value));
  const phases = [...new Set(states.map(state => state.phase))];
  expect(phases).toEqual(Array.from({ length: 16 }, (_, index) => index));
  expect(states.at(-1)).toMatchObject({ phase: 15, source: 4, terminal: true, panel: false, contextClosed: true,
    parent: { count: '135', rows: 100 }, proof: { matched: '15', crossService: true, sameSpan: true, repeat120: true } });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
