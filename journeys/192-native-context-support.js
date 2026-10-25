const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { fixture, traces, logs, metrics } = require('./192-native-context-fixture');
const base = 'http://127.0.0.1:8088';
async function json(response) { expect(response.ok(), await response.text()).toBe(true); return response.json(); }
function question(facts, kind = 'counter', unavailable = false) {
  return { definition: { apiVersion: 'goo/measurement-package/v1alpha1', name: unavailable ? 'unavailable-work' : kind === 'counter' ? 'work-count' : 'work-latency',
    title: unavailable ? 'Unknown export interval' : kind === 'counter' ? 'Native work intervals' : 'Native latency intervals', kind,
    environment: facts.environment, lookbackMinutes: 1, maxGapMs: 10000, grain: 'per-stream',
    source: { services: ['native-worker'], scope: 'ordinary.native', version: '9', schemaURL: 'https://opentelemetry.io/schemas/1.30.0',
      metric: unavailable ? 'work.unavailable' : kind === 'counter' ? 'work.completed' : 'work.latency', unit: kind === 'counter' ? '{work}' : 's' } },
    fromMs: Number(BigInt(facts.fromNS) / 1000000n), toMs: Number((BigInt(facts.emptyToNS) + 999999n) / 1000000n), fromNS: facts.fromNS, toNS: facts.emptyToNS };
}
async function seed(request) {
  const facts = fixture();
  const traceAck = await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: traces(facts) }));
  const logAck = await json(await request.post('http://127.0.0.1:4318/v1/logs', { data: logs(facts) }));
  const metricAck = await json(await request.post('http://127.0.0.1:4318/v1/metrics', { data: metrics(facts) }));
  let ready;
  await expect.poll(async () => {
    ready = await json(await request.post(base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: Number(BigInt(facts.fromNS) / 1000000n) - 1,
      toMs: Number((BigInt(facts.emptyToNS) + 999999n) / 1000000n), limit: 100 } }));
    return ready.count;
  }, { timeout: 60000 }).toBe(String(facts.records.length));
  let native;
  await expect.poll(async () => {
    native = await json(await request.post(base + '/api/measurements', { data: question(facts) }));
    return native.counters?.series.length || 0;
  }, { timeout: 60000 }).toBe(3);
  return { ...facts, traceAck, logAck, metricAck, ready, native };
}
function observe(page) {
  const requests = [], responses = [], errors = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    const url = new URL(request.url());
    if (url.pathname.startsWith('/api/')) requests.push({ endpoint: url.pathname, method: request.method(), payload: request.postDataJSON(),
      effect: request.method() !== 'POST' ? 'metadata' : ['/api/measurement-packages/plan', '/api/dashboards/plan', '/api/compare'].includes(url.pathname) ? 'pure' : 'source' });
  });
  page.on('response', async response => {
    if (response.request().method() === 'POST' && new URL(response.url()).pathname.startsWith('/api/')) {
      try { const raw = await response.text(); responses.push({ endpoint: new URL(response.url()).pathname, status: response.status(), raw, body: JSON.parse(raw) }); } catch {}
    }
  });
  return { requests, responses, errors, source: () => requests.filter(item => item.effect === 'source').length };
}
async function write(facts, name, data) {
  const dir = path.resolve('artifacts/native-record-context', facts.environment);
  await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('192 artifact:', path.join(dir, name)); return dir;
}
module.exports = { base, json, question, seed, observe, write };
