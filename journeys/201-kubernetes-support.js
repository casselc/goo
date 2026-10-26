const { expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const shared = require('./187-http-dashboard-support'), fixture = require('./201-kubernetes-fixture');
function observe(page) { const result = shared.observe(page); return { ...result, source: () => result.source, plans: () => result.plans }; }
function canonical(value) { return JSON.stringify(Array.isArray(value) ? value.map(v => JSON.parse(canonical(v))) : value && typeof value === 'object' ? Object.fromEntries(Object.keys(value).sort().map(k => [k, JSON.parse(canonical(value[k]))])) : value); }
function expected(facts) {
  const groups = new Map();
  const observations = [...facts.points.map(p => ({ family: p.family, resource: p.resource, schema: p.resourceSchemaURL, at: p.endNS })), ...facts.records.map(r => ({ family: r.family, resource: r.resource, schema: r.family === 'spans' ? null : r.resourceSchemaURL, at: r.timeNS }))];
  for (const row of observations) {
    if (row.resource['deployment.environment.name'] !== facts.environment || BigInt(row.at) < BigInt(facts.fromMs) * 1000000n || BigInt(row.at) >= BigInt(facts.toMs) * 1000000n) continue;
    const key = canonical([row.resource, row.schema]);
    if (!groups.has(key)) groups.set(key, { resource: row.resource, resourceSchemaURL: row.schema, families: new Map() });
    const families = groups.get(key).families;
    if (!families.has(row.family)) families.set(row.family, { family: row.family, count: '0', firstNS: row.at, lastNS: row.at });
    const family = families.get(row.family); family.count = String(BigInt(family.count) + 1n);
    if (BigInt(row.at) < BigInt(family.firstNS)) family.firstNS = row.at;
    if (BigInt(row.at) > BigInt(family.lastNS)) family.lastNS = row.at;
  }
  return [...groups.values()].map(row => ({ ...row, families: [...row.families.values()].sort((a, b) => a.family.localeCompare(b.family)) }));
}
function assertInventory(facts, data) {
  const wanted = expected(facts); expect(data.observationsOmitted).toBe(0); expect(data.observations).toHaveLength(wanted.length);
  expect(data.sourceGroups).toBe(wanted.reduce((sum, row) => sum + row.families.length, 0));
  const found = data.observations.map(row => ({ resource: row.resource, resourceSchemaURL: row.resourceSchemaURL, families: [...row.families].sort((a, b) => a.family.localeCompare(b.family)) }));
  expect(found.map(canonical).sort()).toEqual(wanted.map(canonical).sort());
  expect(new Set(data.observations.map(row => row.key)).size).toBe(wanted.length);
}
async function seed(request) {
  const facts = fixture.fixture(), rawOTLP = { metrics: fixture.metrics(facts), traces: fixture.traces(facts), logs: fixture.logs(facts) }, acknowledgements = [];
  for (const [signal, data] of Object.entries(rawOTLP)) acknowledgements.push({ signal, response: await shared.json(await request.post('http://127.0.0.1:4318/v1/' + signal, { data })) });
  let ready; const population = expected(facts).reduce((sum, row) => sum + row.families.reduce((s, family) => s + Number(family.count), 0), 0);
  await expect.poll(async () => {
    ready = await shared.json(await request.post(shared.base + '/api/pod-discovery', { data: { environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs } }));
    return ready.observations?.reduce((sum, row) => sum + row.families.reduce((s, family) => s + Number(family.count), 0), 0);
  }, { timeout: 60000 }).toBe(population);
  assertInventory(facts, ready); return { ...facts, rawOTLP, acknowledgements, ready };
}
async function write(facts, name, data) { const dir = path.resolve('artifacts/kubernetes', facts.environment); await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, name), JSON.stringify(data, null, 2)); console.log('201 artifact:', path.join(dir, name)); return dir; }
async function discover(page, facts) {
  await page.goto(shared.base + '/#services'); await page.locator('#entitiesKubernetes').click();
  await page.getByLabel('Kubernetes environment', { exact: true }).fill(facts.environment); await page.locator('#minutes').fill('2');
  const pending = page.waitForResponse(r => r.url().endsWith('/api/pod-discovery') && r.request().method() === 'POST');
  await page.getByRole('button', { name: 'Find pods', exact: true }).click(); const response = await pending;
  return { request: response.request().postDataJSON(), response: await shared.json(response) };
}
module.exports = { ...shared, observe, seed, write, discover, assertInventory, canonical };
