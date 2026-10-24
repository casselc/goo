const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { startBombadil } = require('./support/cross-app');
const { base, json, seed, observe, run, switchTrace, expand, snapshot, stored, output, custody, question } = require('./185-trace-comparison-support');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const tray = page => page.locator('#exploreContext .trace-reference-tray');
async function compare(page) {
  await tray(page).getByRole('button', { name: 'Compare with reference', exact: true }).click();
  await expect(page.locator('goo-trace-comparison:visible')).toHaveAttribute('data-state', 'available');
  return snapshot(page);
}
async function pin(page) {
  await tray(page).getByRole('button', { name: 'Pin trace page as reference', exact: true }).click();
  await expect(tray(page)).toContainText('Reference: trace');
}
async function frame(page, locator, filename, width = 1440, height = 2200) {
  await page.setViewportSize({ width, height });
  await locator.scrollIntoViewIfNeeded();
  await locator.evaluate(target => {
    const panel = document.querySelector('#exploreContext');
    const top = target.getBoundingClientRect().top, start = panel.getBoundingClientRect().top;
    if (top < start + 15) panel.scrollTop += top - start - 15;
  });
  const bounds = await locator.evaluate(element => {
    const r = element.getBoundingClientRect(), panel = document.querySelector('#exploreContext').getBoundingClientRect();
    return { top: r.top, bottom: r.bottom, left: r.left, right: r.right,
      clipTop: panel.top, clipBottom: panel.bottom, clipLeft: panel.left, clipRight: panel.right };
  });
  expect(bounds.top).toBeGreaterThanOrEqual(bounds.clipTop);
  expect(bounds.bottom).toBeLessThanOrEqual(bounds.clipBottom + 1);
  expect(bounds.left).toBeGreaterThanOrEqual(bounds.clipLeft);
  expect(bounds.right).toBeLessThanOrEqual(bounds.clipRight + 1);
  await locator.screenshot({ path: 'docs/journeys/screenshots/' + filename });
}
async function ownerObservation(page) {
  return page.evaluate(() => {
    const held = window.retained185, host = held?.host;
    if (!host) return null;
    return { connected: host.isConnected, visible: host.getClientRects().length > 0 && !host.closest('[hidden]'),
      comparisonCount: host.querySelectorAll('goo-trace-comparison').length,
      differenceRows: host.querySelectorAll('.trace-pair-attributes tbody tr').length,
      pinDisabled: held.pin.disabled, compareDisabled: held.compare.disabled,
      referenceText: host.querySelector('p')?.textContent ?? '',
      retainedPairConnected: held.pair.isConnected,
      retainedDifferenceRows: held.view.querySelectorAll('.trace-pair-attributes tbody tr').length };
  });
}
function group(view, operation) { return view.model.groups.find(g => g.operation === operation); }
function assertGeometry(view) {
  for (const card of view.cards) {
    expect(Math.abs(card.sides[0].trackWidth - card.sides[1].trackWidth)).toBeLessThanOrEqual(1);
    for (const side of card.sides) if (side.width !== null) {
      expect(parseFloat(side.left)).toBeGreaterThanOrEqual(0);
      expect(parseFloat(side.left) + parseFloat(side.width)).toBeLessThanOrEqual(100.001);
    }
  }
}

test('185 pin two real trace pages and inspect exact timing and qualified attribute differences', async ({ page, request }) => {
  test.setTimeout(180000);
  const f = await seed(request), referenceTrace = f.traces[0], currentTrace = f.traces[1];
  const captureRequest = { title: 'Unchanged original reference trace', question: { kind: 'explore', explore: question(f, referenceTrace) } };
  const saved = await json(await request.post(base + '/api/captures', { data: captureRequest }));
  const before = stored(saved.receipt.id), state = observe(page);
  const referenceSearch = await run(page, question(f, referenceTrace)); state.event('reference-search');
  const reference = await expand(page, referenceTrace.traceId); state.event('reference-expand');
  const beforePinNetwork = state.network.length;
  await pin(page); expect(state.network.length).toBe(beforePinNetwork); expect(state.source).toBe(3); state.event('pin');
  const referenceContext = await tray(page).textContent();
  const currentSearch = await switchTrace(page, currentTrace); state.event('current-search');
  expect(currentSearch.question.fromMs).not.toBe(reference.question.fromMs);
  const current = await expand(page, currentTrace.traceId); state.event('current-expand');
  expect(state.source).toBe(6);
  const beforeCompareNetwork = state.network.length;
  let view = await compare(page); state.event('compare');
  expect(state.network.length).toBe(beforeCompareNetwork);
  const comparisonContext = await tray(page).textContent();
  expect(view.input.baseline.rows).toEqual(reference.rows);
  expect(view.input.selection.rows).toEqual(current.rows);
  expect(view.model.baseline.traceId).toBe(referenceTrace.traceId);
  expect(view.model.selection.traceId).toBe(currentTrace.traceId);
  expect(group(view, 'GET /checkout')).toMatchObject({ matchState: 'unique-label-match', inclusiveDifferenceNs: '30000000', outsideDifferenceNs: '25000000',
    match: { inclusiveDifferenceNs: '30000000', outsideDifferenceNs: '25000000', relativeStartDifferenceNs: '0' } });
  expect(group(view, 'lookup')).toMatchObject({ matchState: 'unique-label-match', inclusiveDifferenceNs: '-10000000' });
  expect(group(view, 'prepare')).toMatchObject({ matchState: 'ambiguous', matchReason: 'repeated-label', inclusiveDifferenceNs: '10000000' });
  expect(group(view, 'prepare').match).toBeUndefined();
  expect(group(view, 'cache lookup').matchState).toBe('baseline-only');
  expect(group(view, 'enqueue').matchState).toBe('selection-only');
  expect(group(view, 'dispatch')).toMatchObject({ matchState: 'unavailable', matchReason: 'span-kind-mismatch-or-unreported' });
  expect(group(view, 'marker')).toMatchObject({ match: { inclusiveDifferenceNs: '0' } });
  const markerIndex = view.model.groups.findIndex(g => g.operation === 'marker');
  expect(view.cards[markerIndex].sides.map(side => side.width)).toEqual(['0%', '0%']);
  assertGeometry(view);
  expect(state.source).toBe(6);
  await expect(tray(page)).toContainText(reference.capturedAt);
  await expect(tray(page)).toContainText(current.capturedAt);
  await expect(tray(page)).toContainText(new Date(reference.question.fromMs).toISOString());
  await expect(tray(page)).toContainText(new Date(current.question.fromMs).toISOString());
  await frame(page, page.locator('goo-trace-comparison:visible'), '185a-two-observed-trace-pages.png');
  const rootIndex = view.model.groups.findIndex(g => g.operation === 'GET /checkout');
  const rootCard = page.locator('goo-trace-comparison:visible .trace-pair').nth(rootIndex);
  await frame(page, rootCard, '185b-mobile-unique-label-pair.png', 390, 1400);
  const mobileView = await snapshot(page); assertGeometry(mobileView);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.setViewportSize({ width: 1440, height: 1800 });
  const beforeAttributesNetwork = state.network.length;
  await rootCard.getByRole('button', { name: 'Inspect pair attributes', exact: true }).click();
  expect(state.network.length).toBe(beforeAttributesNetwork);
  view = await snapshot(page); state.event('pair-attributes');
  const byField = Object.fromEntries(view.differences.map(row => [row.field, row.values]));
  expect(byField['attribute:work.changed'].join(' ')).toContain('old');
  expect(byField['attribute:work.changed'].join(' ')).toContain('new');
  expect(byField['attribute:work.empty'][0]).toContain('Empty string');
  expect(byField['attribute:work.empty'][1]).toContain('Missing value');
  expect(byField['attribute:work.namespace'].join(' ')).toContain('attribute-old');
  expect(byField['resource:work.namespace'].join(' ')).toContain('resource-old');
  expect(byField['attribute:work.same']).toBeUndefined();
  expect(state.source).toBe(6);
  await frame(page, page.locator('goo-trace-comparison:visible .trace-pair-attributes'), '185c-qualified-pair-attribute-differences.png');
  expect(view.input.baseline.rows).toEqual(reference.rows);
  await page.locator('#exploreContextClose').click();
  const beforeReopen = state.source;
  await page.goto(base + '/#capture=' + saved.receipt.id);
  await expect(page.locator('#captureGeneralID')).toHaveValue(saved.receipt.id);
  await expect(page.locator('#savedexploreRows .record-row')).toHaveCount(7);
  expect(state.source).toBe(beforeReopen); state.event('capture-reopen');
  const reopened = await json(await request.get(base + '/api/captures/' + saved.receipt.id));
  expect(reopened).toEqual(saved);
  const after = stored(saved.receipt.id); expect(after.equals(before)).toBe(true);
  expect(state.errors).toEqual([]);
  const dir = await output(f, 'ledger.json', { ...f, captureRequest, saved, reopened,
    custody: custody(before, saved.receipt.id), referenceSearch, reference, referenceContext,
    currentSearch, current, view, mobileView, comparisonContext, events: state.events, requests: state.requests, responses: state.responses, network: state.network });
  await fs.writeFile(path.join(dir, 'scope-capture.json'), before);
  await fs.writeFile(path.join(dir, 'scope-capture.after.json'), after);
});

test('185 partial and outside-parent pages withhold unavailable row timing', async ({ page, request }) => {
  test.setTimeout(150000);
  const f = await seed(request), state = observe(page);
  const referenceSearch = await run(page, question(f, f.traces[2]));
  const reference = await expand(page, f.traces[2].traceId); await pin(page); state.event('problem-pin');
  const currentSearch = await switchTrace(page, f.traces[3]);
  const current = await expand(page, f.traces[3].traceId);
  const problem = await compare(page); state.event('problem-compare');
  expect(group(problem, 'GET /checkout').matchState).toBe('unique-label-match');
  expect(group(problem, 'GET /checkout').inclusiveDifferenceNs).toBe('0');
  expect(group(problem, 'GET /checkout').outsideDifferenceNs).toBeUndefined();
  expect(group(problem, 'GET /checkout').match.outsideDifferenceNs).toBeUndefined();
  expect(group(problem, 'lookup').match.inclusiveDifferenceNs).toBe('10000000');
  expect(problem.model.selection.breakdown.rows.some(row => row.reasons.includes('child-outside-parent'))).toBe(true);
  expect(state.source).toBe(6);
  const partialSearch = await switchTrace(page, f.traces[4]);
  const partial = await expand(page, f.traces[4].traceId);
  expect(partial.count).toBe('126'); expect(partial.rows).toHaveLength(100); expect(partial.hasMore).toBe(true);
  const partialView = await compare(page); state.event('partial-compare');
  expect(partialView.input.selection.truncated).toBe(true);
  await expect(page.locator('goo-trace-comparison:visible')).toContainText('Partial page');
  expect(group(partialView, 'fanout').selection.rowIndexes).toHaveLength(99);
  expect(group(partialView, 'fanout').match).toBeUndefined();
  expect(state.source).toBe(9); expect(state.errors).toEqual([]);
  await output(f, 'partial.json', { ...f, referenceSearch, reference, currentSearch, current, problem,
    partialSearch, partial, partialView, events: state.events, requests: state.requests, responses: state.responses, network: state.network });
});

test('185 pending and hidden owners reject retained pin, compare and pair actions', async ({ page, request }) => {
  test.setTimeout(150000);
  const f = await seed(request), state = observe(page), observations = {};
  await run(page, question(f, f.traces[0])); const reference = await expand(page, f.traces[0].traceId); await pin(page);
  await switchTrace(page, f.traces[1]); const current = await expand(page, f.traces[1].traceId);
  const view = await compare(page); state.event('pair');
  const rootIndex = view.model.groups.findIndex(g => g.operation === 'GET /checkout');
  await page.evaluate(index => {
    const host = document.querySelector('#exploreContext .trace-reference-tray');
    window.retained185 = { host, pin: [...host.querySelectorAll('button')].find(b => b.textContent === 'Pin trace page as reference'),
      compare: [...host.querySelectorAll('button')].find(b => b.textContent === 'Compare with reference'),
      clear: [...host.querySelectorAll('button')].find(b => b.textContent === 'Clear trace reference'),
      pair: host.querySelector('[data-trace-pair="' + index + '"]'), view: host.querySelector('goo-trace-comparison') };
    retained185.view.querySelector('details').open = false;
    retained185.pair.click();
  }, rootIndex);
  expect(await page.locator('.trace-pair-attributes tbody tr').count()).toBe(0);
  expect(state.source).toBe(6); observations.closedPair = await ownerObservation(page); state.event('closed-pair', observations.closedPair);
  await page.goto(base + '/#definitions');
  const beforeURL = page.url();
  await page.evaluate(() => { retained185.pin.click(); retained185.compare.click(); retained185.clear.click(); retained185.pair.click(); });
  const afterURL = page.url(); expect(afterURL).toBe(beforeURL); expect(state.source).toBe(6); observations.hidden = await ownerObservation(page); state.event('hidden-actions', observations.hidden);
  await page.goto(base + '/#explore');
  // Returning never silently re-expands. A fresh explicit read restores accepted authority.
  expect(state.source).toBe(6); state.event('return-stage');
  if (await page.locator('#exploreContextClose').isVisible()) await page.locator('#exploreContextClose').click();
  const returnedPage = await expand(page, f.traces[1].traceId);
  await compare(page); expect(state.source).toBe(8); state.event('returned-compare');
  const returned = await snapshot(page); expect(returned.input.baseline.rows).toEqual(reference.rows);
  await switchTrace(page, f.traces[4]); const partial = await expand(page, f.traces[4].traceId); await compare(page);
  expect(state.source).toBe(11);
  let settle, held;
  const released = new Promise(resolve => { settle = resolve; });
  const ready = new Promise(resolve => { held = resolve; });
  let heldRequest, heldResponse;
  await page.route('**/api/explore', async route => {
    const q = route.request().postDataJSON();
    if (q.signal !== 'spans' || q.offset !== 100) return route.continue();
    heldRequest = q; const response = await route.fetch(); heldResponse = await response.json(); held();
    await released; await route.fulfill({ response });
  });
  await page.locator('#exploreTraceMore').click(); await ready;
  await expect(tray(page).getByRole('button', { name: 'Pin trace page as reference', exact: true })).toBeDisabled();
  await expect(tray(page).getByRole('button', { name: 'Compare with reference', exact: true })).toBeDisabled();
  await page.evaluate(() => {
    const host = document.querySelector('#exploreContext .trace-reference-tray');
    [...host.querySelectorAll('button')].filter(b => ['Pin trace page as reference', 'Compare with reference'].includes(b.textContent)).forEach(b => b.click());
  });
  expect(await page.locator('goo-trace-comparison').count()).toBe(0);
  expect(state.source).toBe(12); observations.pending = await ownerObservation(page); state.event('pending-actions', observations.pending);
  const dirtyEnvironment = f.environment + '-draft';
  await page.locator('#exploreEnvironment').fill(dirtyEnvironment);
  await page.locator('#exploreEnvironment').dispatchEvent('change');
  const stagedURL = page.url();
  settle();
  await expect.poll(() => state.responses.some(response => response.question.offset === 100)).toBe(true);
  await expect(page.locator('#exploreStatus')).toContainText('Draft changes');
  expect(state.source).toBe(12); expect(await page.locator('goo-trace-comparison').count()).toBe(0);
  observations.stale = await ownerObservation(page); state.event('stale-response', observations.stale);
  await page.locator('#exploreEnvironment').fill(f.environment);
  await page.locator('#exploreEnvironment').dispatchEvent('change');
  const { execute } = require('./185-trace-comparison-support');
  const freshSearch = await execute(page);
  const fresh = await expand(page, f.traces[4].traceId); const recovered = await compare(page);
  expect(recovered.input.baseline.rows).toEqual(reference.rows); expect(state.source).toBe(15); state.event('fresh-accepted');
  expect(state.errors).toEqual([]);
  await output(f, 'ownership.json', { ...f, reference, current, view, returnedPage, returned, partial,
    beforeURL, afterURL, stagedURL, dirtyEnvironment, freshSearch, observations, heldRequest, heldResponse, fresh, recovered,
    events: state.events, requests: state.requests, responses: state.responses, network: state.network });
});

test('185 optional real trace comparison Bombadil campaign', async ({ request }) => {
  test.skip(process.env.GOO_TRACE_COMPARISON_BOMBADIL !== '1', 'Explicit ordinary-telemetry campaign');
  test.setTimeout(150000);
  const f = await seed(request), dir = await output(f, 'fixture.json', f);
  const { questionLink } = require('../ui/workspace');
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'),
    url: base + '/' + questionLink(question(f, f.traces[0])),
    specification: 'verification/bombadil/trace-comparison.ts', seconds: 55 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const records = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = records.flatMap(record => record.snapshots.filter(snapshot => snapshot.name === 'state').map(snapshot => snapshot.value));
  const phases = [...new Set(states.map(state => state.phase))];
  expect(phases).toEqual(Array.from({ length: 16 }, (_, index) => index));
  expect(states.at(-1)).toMatchObject({ phase: 15, source: 6, compareDisabled: true,
    proof: { inclusiveDifferenceNs: '30000000', outsideDifferenceNs: '25000000', differentTraceIDs: true, ambiguous: true } });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
