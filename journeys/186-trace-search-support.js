const { expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { fixture, query, payload } = require('./186-trace-search-fixture');
const base = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const kindNames = ['Unspecified', 'Internal', 'Server', 'Client', 'Producer', 'Consumer'];
async function json(response) {
  expect(response.ok(), 'HTTP ' + response.status() + ': ' + (await response.text()).slice(0, 1000)).toBe(true);
  return response.json();
}
function field(record, value) {
  if (value.kind === 'attribute') return Object.hasOwn(record.attributes, value.key) ? record.attributes[value.key] : null;
  if (value.kind === 'resource') {
    const resource = { ...record.resource, 'service.name': record.service, 'deployment.environment.name': record.environment };
    return Object.hasOwn(resource, value.key) ? resource[value.key] : null;
  }
  return ({ service: record.service, name: record.operation, trace_id: record.traceId,
    span_id: record.spanId, duration_ns: record.durationNs, timestamp_ns: record.startNs,
    kind: kindNames[record.kind], status_class: record.status === 2 ? 'error' : record.status === 1 ? 'ok' : 'unset' })[value.key] ?? null;
}
function matches(record, predicate) {
  if (predicate.all) return predicate.all.every(child => matches(record, child));
  if (predicate.any) return predicate.any.some(child => matches(record, child));
  const value = field(record, predicate.field), expected = predicate.value ?? '';
  if (predicate.op === 'exists') return value !== null;
  if (predicate.op === 'missing') return value === null;
  if (predicate.op === 'eq') return value === expected;
  if (predicate.op === 'ne') return value !== null && value !== expected;
  if (predicate.op === 'gte') return value !== null && BigInt(value) >= BigInt(expected);
  if (predicate.op === 'lte') return value !== null && BigInt(value) <= BigInt(expected);
  throw Error('Unimplemented independent fixture predicate: ' + predicate.op);
}
function inScope(record, request) {
  const time = BigInt(record.startNs);
  return record.environment === request.environment && time >= BigInt(request.fromMs) * 1000000n && time < BigInt(request.toMs) * 1000000n;
}
function oracle(facts, request) {
  const groups = new Map();
  for (const record of facts.traces.flatMap(trace => trace.records)) {
    if (!inScope(record, request) || record.traceId.length !== 32 || !/^[a-f0-9]{32}$/.test(record.traceId) || /^0+$/.test(record.traceId)) continue;
    if (!groups.has(record.traceId)) groups.set(record.traceId, []);
    groups.get(record.traceId).push(record);
  }
  return [...groups].filter(([, records]) => request.conditions.every(condition => records.some(record => matches(record, condition)))).sort(([a], [b]) => a < b ? -1 : a > b ? 1 : 0);
}
function assertResult(facts, request, result) {
  const expected = oracle(facts, request);
  expect(result.request).toEqual(request);
  expect(result.matchedTraces).toBe(String(expected.length));
  expect(result.hasMore).toBe(expected.length > request.limit);
  expect(result.state).toBe(expected.length ? 'observed' : 'no_data');
  expect(result.items.map(item => item.traceId)).toEqual(expected.slice(0, request.limit).map(([id]) => id));
  for (const item of result.items) {
    const records = expected.find(([id]) => id === item.traceId)[1];
    expect(item.witnesses).toHaveLength(request.conditions.length);
    for (const witness of item.witnesses) {
      const condition = request.conditions[witness.conditionIndex];
      const matching = records.filter(record => matches(record, condition)).sort((a, b) => BigInt(a.startNs) < BigInt(b.startNs) ? -1 : BigInt(a.startNs) > BigInt(b.startNs) ? 1 : a.spanId < b.spanId ? -1 : a.spanId > b.spanId ? 1 : 0);
      const record = matching[0];
      expect(witness.matchingRecords).toBe(String(matching.length));
      expect(witness.spanId).toBe(record.spanId);
      expect(witness.parentSpanId).toBe(record.parentSpanId);
      expect(witness.service).toBe(record.service);
      expect(witness.operation).toBe(record.operation);
      expect(witness.timestampNs).toBe(record.startNs);
      expect(witness.durationNs).toBe(record.durationNs);
      expect(witness.kind).toBe(kindNames[record.kind]);
      const required = new Set();
      const collect = predicate => {
        if (predicate.all || predicate.any) (predicate.all || predicate.any).forEach(collect);
        else required.add(predicate.field.kind + ':' + predicate.field.key);
      };
      collect(condition);
      expect(witness.fieldValues.map(value => value.field.kind + ':' + value.field.key).sort()).toEqual([...required].sort());
      for (const supplied of witness.fieldValues) expect(supplied.value).toBe(field(record, supplied.field));
    }
  }
}
async function seed(request) {
  const facts = fixture();
  const ingestion = await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: payload(facts) }));
  const parent = query(facts);
  const expected = facts.traces.flatMap(trace => trace.records).filter(record => inScope(record, parent) && matches(record, parent.where)).length;
  let ready;
  await expect.poll(async () => { ready = await json(await request.post(base + '/api/explore', { data: parent })); return ready.count; }, { timeout: 60000 }).toBe(String(expected));
  expect(expected).toBe(135);
  const physicalRequest = { signal: 'spans', environment: facts.environment,
    fromMs: facts.fromMs - 20000, toMs: facts.toMs, limit: 100, offset: 0,
    groupBy: [{ kind: 'column', key: 'trace_id' }] };
  const physical = await json(await request.post(base + '/api/explore', { data: physicalRequest }));
  const otherRequest = { ...physicalRequest, environment: facts.environment + '-other' };
  const other = await json(await request.post(base + '/api/explore', { data: otherRequest }));
  expect(physical.count).toBe('291'); expect(physical.groups).toHaveLength(19);
  expect(other.count).toBe('3'); expect(other.groups).toHaveLength(1);
  for (const result of [physical, other]) {
    const expected = facts.traces.filter(trace => trace.environment === result.question.environment);
    expect(result.groups.map(group => [group.g0, group.count]).sort()).toEqual(expected.map(trace => [trace.traceId, String(trace.records.length)]).sort());
  }
  return { ...facts, ingestion, ready, physicalProbes: { common: physical, other } };
}
function observe(page) {
  const requests = [], responses = [], network = [], events = [], errors = [];
  let source = 0;
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (!endpoint.startsWith('/api/')) return;
    const value = { endpoint, method: request.method(), payload: request.method() === 'POST' ? request.postDataJSON() : null };
    network.push(value);
    if (request.method() === 'POST') {
      const effect = /^\/api\/queries\/[^/]+\/cancel$/.test(endpoint) ? 'cancel' : 'source';
      requests.push({ ...value, effect, source: effect === 'source' ? ++source : source });
    }
  });
  page.on('response', async response => {
    const endpoint = new URL(response.url()).pathname;
    if (response.request().method() === 'POST' && ['/api/explore', '/api/trace-search'].includes(endpoint) && response.ok()) responses.push({ endpoint, request: response.request().postDataJSON(), response: await response.json() });
  });
  return { requests, responses, network, events, errors, get source() { return source; },
    event(action, extra = {}) { events.push({ action, source, ...extra }); } };
}
async function output(facts, name, data) {
  const dir = path.resolve('artifacts/trace-search', facts.environment);
  await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2));
  console.log('186 artifact:', path.join(dir, name)); return dir;
}
module.exports = { base, json, query, matches, inScope, oracle, assertResult, seed, observe, output };
