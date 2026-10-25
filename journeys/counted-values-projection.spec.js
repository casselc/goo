const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { questionLink } = require('../ui/workspace');
const { fixture, question, logs, field } = require('./193-counted-values-fixture');
const { openFilters } = require('./support/explore');
const base = 'http://127.0.0.1:8088';
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
async function json(response) { expect(response.ok(), await response.text()).toBe(true); return response.json(); }
async function seed(request) {
  const facts = fixture(); const ack = await json(await request.post('http://127.0.0.1:4318/v1/logs', { data: logs(facts) }));
  let ready;
  await expect.poll(async () => { ready = await json(await request.post(base + '/api/explore', { data: question(facts) })); return ready.count; }, { timeout: 60000 }).toBe('143');
  return { ...facts, ack, ready };
}
async function observedField(page) {
  await openFilters(page); const editor = page.locator('#exploreFilterEditor'), panel = editor.locator('.filter-observations');
  if (!await panel.evaluate(element => element.open)) await panel.locator(':scope>summary').click();
  await panel.getByRole('searchbox', { name: 'Search observed fields' }).fill('attribute:release');
  const entry = panel.locator(':scope>div>details'); if (!await entry.evaluate(element => element.open)) await entry.locator(':scope>summary').click();
  return editor;
}
async function load(page, editor) {
  const pending = page.waitForResponse(response => response.url().endsWith('/api/explore/counted-values'));
  await editor.getByRole('button', { name: 'Load scoped values for attribute:release', exact: true }).click();
  const response = await pending, data = await json(response);
  await expect(editor.locator('.filter-counted-values')).toContainText('matching records');
  return { request: response.request().postDataJSON(), data };
}
test('193 counted-values projection preserves exact scope with full groups and stale ownership rejection', async ({ page, request }) => {
  test.setTimeout(120000);
  const facts = await seed(request), requests = [], errors = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', req => { if (new URL(req.url()).pathname.startsWith('/api/')) requests.push({ method: req.method(), endpoint: new URL(req.url()).pathname, payload: req.postDataJSON() }); });
  const posts = () => requests.filter(item => item.method === 'POST');
  await page.goto(base + '/' + questionLink(question(facts)));
  let pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const original = await json(await pending);
  expect(original.count).toBe('143'); expect(original.rows).toHaveLength(100);
  const beforeBrowse = posts().length, editor = await observedField(page);
  expect(posts()).toHaveLength(beforeBrowse);
  const large = await load(page, editor);
  expect(posts()).toHaveLength(beforeBrowse + 1);
  expect(large.request).toEqual({ ...question(facts), limit: 1, groupBy: [field] });
  expect(large.data.version).toBe('goo/counted-values/v1'); expect(large.data.count).toBe('143');
  expect(large.data.groups).toEqual(expect.arrayContaining([{ g0: 'common', count: '140' }, { g0: 'canary', count: '1' }, { g0: '', count: '1' }, { g0: null, count: '1' }]));
  expect(large.data.groups).toHaveLength(4);
  expect(Object.keys(large.data).sort()).toEqual(['version', 'question', 'capturedAt', 'basis', 'count', 'state', 'groups', 'reads'].sort());
  for (const read of Object.values(large.data.reads)) { expect(read.basis).toBeTruthy(); expect(Number.isFinite(Date.parse(read.readAt))).toBe(true); }
  const counted = editor.locator('.filter-counted-values'), beforeLocal = posts().length;
  await counted.getByRole('searchbox', { name: 'Search counted values' }).fill('canary');
  await expect(counted.locator('[data-facet-item]:visible')).toHaveCount(1);
  await counted.getByRole('searchbox', { name: 'Search counted values' }).fill('');
  expect(posts()).toHaveLength(beforeLocal);
  await counted.screenshot({ path: 'docs/journeys/screenshots/193a-counted-values-full-population.png' });
  await page.setViewportSize({ width: 390, height: 1000 });
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await counted.screenshot({ path: 'docs/journeys/screenshots/193b-mobile-counted-values.png' });
  await page.setViewportSize({ width: 1440, height: 1100 });
  await counted.getByRole('button', { name: 'Stage counted attribute:release Missing key', exact: true }).click();
  expect(posts()).toHaveLength(beforeLocal); await expect(counted).toBeHidden();
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); const selected = await json(await pending);
  expect(selected.count).toBe('1'); expect(selected.question.where.all).toContainEqual({ field, op: 'missing' });
  expect(JSON.stringify(selected.question.where)).toContain('9007199254740993');
  const smallRequest = { ...question(facts, 'small'), limit: 1, groupBy: [field] };
  const small = await json(await request.post(base + '/api/explore/counted-values', { data: smallRequest }));
  expect(small.count).toBe('5'); expect(small.groups).toHaveLength(4);
  // Delay a real positive response; changing scope must clear and refuse it.
  await page.goto(base + '/' + questionLink(question(facts)));
  pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
  await page.locator('#exploreRun').click(); await json(await pending); await observedField(page);
  let release, entered, settled, heldData;
  const gate = new Promise(resolve => { release = resolve; }), started = new Promise(resolve => { entered = resolve; }), done = new Promise(resolve => { settled = resolve; });
  await page.route('**/api/explore/counted-values', async route => { const response = await route.fetch(); heldData = await response.json(); entered(); await gate; try { await route.fulfill({ response }); } catch {} finally { settled(); } });
  await editor.getByRole('button', { name: 'Load scoped values for attribute:release', exact: true }).click(); await started;
  await page.locator('#exploreEnvironment').fill(facts.environment + '-changed');
  await expect(counted).toBeHidden(); release(); await done; await page.unroute('**/api/explore/counted-values'); await expect(counted).toBeHidden();
  expect(errors).toEqual([]);
  const dir = path.resolve('artifacts/counted-values-projection', facts.environment); await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, 'ledger.json'), JSON.stringify({ version: 1, facts, original, large, selected, smallRequest, small, heldData, beforeBrowse, beforeLocal, requests, errors, stalePublished: false }, null, 2));
  console.log('193 ledger:', path.join(dir, 'ledger.json'));
});
module.exports = { seed };
