const { expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { fixture, traces, metrics, expected } = require('./187-http-dashboard-fixture');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
async function json(response) {
  expect(response.ok(), 'HTTP ' + response.status() + ': ' + (await response.text()).slice(0, 1000)).toBe(true);
  return response.json();
}
async function seed(request) {
  const facts = fixture();
  const traceAck = await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: traces(facts) }));
  const metricAck = await json(await request.post('http://127.0.0.1:4318/v1/metrics', { data: metrics(facts) }));
  const question = { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100, groupBy: [{ kind: 'column', key: 'service' }] };
  let ready;
  await expect.poll(async () => {
    ready = await json(await request.post(base + '/api/explore', { data: question }));
    return ready.count;
  }, { timeout: 60000 }).toBe(String(facts.records.length));
  let nativeReady;
  await expect.poll(async () => {
    nativeReady = await json(await request.post(base + '/api/metric-discovery', { data: { environment: facts.environment, service: 'edge-api', namespace: { mode: 'value', value: 'commerce' }, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return nativeReady.instruments?.length || 0;
  }, { timeout: 60000 }).toBe(3);
  return { ...facts, traceAck, metricAck, ready, nativeReady };
}
function observe(page) {
  const requests = [], responses = [], errors = [], events = [], pending = new Set();
  const pure = new Set(['/api/dashboards/plan', '/api/measurement-packages/plan', '/api/packages/plan']);
  let source = 0, plans = 0;
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (!endpoint.startsWith('/api/')) return;
    const isPost = request.method() === 'POST';
    const effect = !isPost ? 'metadata' : pure.has(endpoint) ? 'plan' : /^\/api\/queries\/[^/]+\/cancel$/.test(endpoint) ? 'cancel' : 'source';
    if (effect === 'source') source++;
    if (effect === 'plan') plans++;
    requests.push({ endpoint, method: request.method(), effect, source, plans, payload: isPost ? request.postDataJSON() : null });
  });
  page.on('response', response => {
    const endpoint = new URL(response.url()).pathname;
    if (!endpoint.startsWith('/api/') || response.request().method() !== 'POST') return;
    const work = (async () => {
      let body; try { body = await response.json(); } catch { body = await response.text(); }
      responses.push({ endpoint, status: response.status(), request: response.request().postDataJSON(), body });
    })();
    pending.add(work); work.finally(() => pending.delete(work));
  });
  return { requests, responses, errors, events, get source() { return source; }, get plans() { return plans; },
    async settle() { await Promise.all([...pending]); },
    event(action, extra = {}) { events.push({ action, source, plans, ...extra }); } };
}
async function loaded(page, profile) {
  const dash = page.locator('.service-dashboard');
  await expect(page.locator('#serviceDetail .service-loading')).toHaveCount(0);
  await expect(dash).toHaveAttribute('data-state', 'loaded');
  if (profile) await expect(dash).toHaveAttribute('data-profile', profile);
  return dash;
}
async function snapshot(page, observation) {
  await observation.settle();
  return page.locator('.service-dashboard').evaluate(host => ({
    profile: host.dataset.profile, profileState: host.dataset.profileState, state: host.dataset.state,
    title: host.querySelector('h4')?.textContent,
    evidence: [...host.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => JSON.parse(pre.textContent)),
    families: [...host.querySelectorAll('[data-dashboard-profile]')].map(button => ({ key: button.dataset.dashboardProfile, selected: button.getAttribute('aria-pressed'), disabled: button.disabled, text: button.textContent })),
    native: [...host.querySelectorAll('tr[data-native-kind]')].map(row => ({ ...row.dataset, text: row.textContent })),
    rate: host.querySelector('.dashboard-record-rate')?.textContent,
    rateTitle: host.querySelector('.dashboard-record-rate')?.title,
    text: host.textContent,
  }));
}
function assertPopulation(facts, result, profile, service = 'edge-api') {
  const records = expected(facts, profile, service).filter(record => record.timeMs >= result.question.fromMs && record.timeMs < result.question.toMs);
  expect(result.count).toBe(String(records.length));
  expect(result.rows.length).toBe(Math.min(100, records.length));
  const byID = new Map(records.map(record => [record.id, record]));
  const ids = result.rows.map(row => row.SpanId);
  expect(new Set(ids).size).toBe(ids.length);
  for (const row of result.rows) {
    const record = byID.get(row.SpanId);
    expect(record, 'unexpected span ' + row.SpanId).toBeTruthy();
    expect(row.TraceId).toBe(record.traceId);
    expect(row.duration_ns).toBe(record.durationNs);
  }
  if (result.duration) {
    expect(result.duration.count).toBe(String(records.length));
    expect(result.duration.sumNs).toBe(String(records.reduce((sum, record) => sum + BigInt(record.durationNs), 0n)));
  }
  if (result.spanStatus) expect(result.spanStatus.error).toBe(String(records.filter(record => record.status === 2).length));
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/http-server-dashboard', facts.environment);
  await fs.mkdir(dir, { recursive: true });
  await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('187 artifact:', path.join(dir, name));
  return dir;
}
module.exports = { base, json, seed, observe, loaded, snapshot, assertPopulation, write };
