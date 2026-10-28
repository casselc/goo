const { expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { execFileSync } = require('node:child_process');
const { createHash } = require('node:crypto');
const { questionLink } = require('../ui/workspace.mjs');
const { clearFilters, addFilter } = require('./support/explore');
const { fixture, question, payload } = require('./185-trace-comparison-fixture');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
async function json(response) {
  expect(response.ok(), 'HTTP ' + response.status() + ': ' + (await response.text()).slice(0, 1000)).toBe(true);
  return response.json();
}
async function seed(request) {
  const facts = fixture();
  await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: payload(facts) }));
  for (const trace of facts.traces) await expect.poll(async () => {
    return (await json(await request.post(base + '/api/explore', { data: question(facts, trace) }))).count;
  }, { timeout: 60000 }).toBe(String(trace.records.length));
  return facts;
}
function observe(page) {
  const requests = [], responses = [], events = [], errors = [], network = [];
  let source = 0;
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (endpoint.startsWith('/api/')) network.push({ endpoint, method: request.method(), payload: request.method() === 'POST' ? request.postDataJSON() : null });
    if (request.method() === 'POST' && endpoint.startsWith('/api/')) {
      requests.push({ endpoint, payload: request.postDataJSON(), effect: 'source', ordinal: ++source });
    }
  });
  page.on('response', async response => {
    if (response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore' && response.ok()) {
      responses.push({ question: response.request().postDataJSON(), result: await response.json() });
    }
  });
  return { requests, responses, events, errors, network, get source() { return source; },
    event(action, extra = {}) { events.push({ action, source, ...extra }); } };
}
async function execute(page) {
  const pending = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore');
  await page.locator('#exploreRun').click();
  const result = await json(await pending);
  await expect(page.locator('#exploreStatus')).toContainText(result.count + ' matching');
  return result;
}
async function run(page, q) {
  await page.goto(base + '/' + questionLink(q));
  return execute(page);
}
async function switchTrace(page, trace) {
  if (await page.locator('#exploreContextClose').isVisible()) await page.locator('#exploreContextClose').click();
  await clearFilters(page);
  await addFilter(page, { kind: 'column', key: 'trace_id' }, 'eq', trace.traceId);
  await page.locator('#minutes').fill('1');
  await page.locator('#exploreFreshWindow').click();
  return execute(page);
}
async function expand(page, traceId) {
  const pending = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore' && response.request().postDataJSON().signal === 'spans');
  await page.getByRole('button', { name: 'Expand trace ' + traceId, exact: true }).first().click();
  const result = await json(await pending);
  await expect(page.locator('#exploreLogStatus')).toContainText('captured logs');
  await expect(page.locator('#exploreContextSummary')).toContainText(result.count + ' captured spans');
  return result;
}
async function snapshot(page) {
  return page.locator('goo-trace-comparison:visible').evaluate(async (element)=>{const GooTracePageComparison=await import('/trace-page-comparison.mjs');
    const input = element.data;
    return { input, model: GooTracePageComparison.compare(input.baseline.rows, input.selection.rows,
      !!input.baseline.truncated, !!input.selection.truncated),
      cards: [...element.querySelectorAll('.trace-pair')].map(card => ({
        index: Number(card.dataset.groupIndex), matchState: card.dataset.matchState,
        label: card.querySelector('h4').textContent, text: card.textContent,
        sides: [...card.querySelectorAll('.trace-pair-side')].map(side => {
          const axis = side.querySelector('.trace-pair-axis'), bar = side.querySelector('.trace-pair-bar');
          return { side: side.dataset.side, rowIndex: axis.dataset.rowIndex ?? null,
            trackWidth: axis.getBoundingClientRect().width,
            left: bar?.style.left ?? null, width: bar?.style.width ?? null,
            title: bar?.title ?? null, text: side.textContent };
        }),
      })),
      differences: [...element.querySelectorAll('.trace-pair-attributes tbody tr')].map(row => ({
        field: row.querySelector('th').textContent,
        values: [...row.querySelectorAll('td')].map(cell => cell.textContent),
        fields: [...row.querySelectorAll('goo-field-value')].map(field => ({ state: field.state, value: field.value })),
      })),
    };
  });
}
function stored(id) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const env = { ...process.env };
  for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  return execFileSync('docker', ['--host=unix:///var/run/docker.sock', 'exec',
    process.env.GOO_QUERY_CONTAINER || 'grammar-observability-packages-1', 'cat',
    '/app/artifacts/investigations/exploration/' + id + '.json'], { env, maxBuffer: 8 * 1024 * 1024 });
}
async function output(facts, name, value) {
  const dir = path.resolve('artifacts/trace-comparison', facts.environment);
  await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(value, null, 2));
  console.log('185 artifact:', path.join(dir, name));
  return dir;
}
function custody(bytes, id) { return { id, bytes: bytes.length, sha256: createHash('sha256').update(bytes).digest('hex') }; }
module.exports = { base, json, seed, observe, execute, run, switchTrace, expand, snapshot, stored, output, custody, question };
