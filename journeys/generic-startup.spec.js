const { test, expect } = require('@playwright/test');
const path = require('node:path');
const fs = require('node:fs/promises');
const { spawnSync } = require('node:child_process');
const { startRuntime } = require('./191-generic-runtime');
const { fixture, traces, logs } = require('./191-generic-fixture');
const { readStepLink } = require('../ui/workspace.mjs');
test.use({ actionTimeout: 15000, navigationTimeout: 45000 });
const json = async response => { expect(response.ok(), await response.text()).toBe(true); return response.json(); };

test('191 isolated generic startup observes Python-tagged OTel without lab assets', async ({ page, request }) => {
  test.setTimeout(150000);
  test.skip(!process.env.GOO_GENERIC_BINARY, 'Set GOO_GENERIC_BINARY to an explicit compiled binary after source freeze');
  const facts = fixture(), directory = path.resolve('artifacts/generic-startup', facts.environment);
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, directory);
  const requests = [], responses = [], errors = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', req => { if (new URL(req.url()).pathname.startsWith('/api/')) requests.push({ method: req.method(), endpoint: new URL(req.url()).pathname, payload: req.postDataJSON() }); });
  page.on('response', async response => {
    if (new URL(response.url()).pathname.startsWith('/api/') && response.request().method() === 'POST') {
      try { responses.push({ endpoint: new URL(response.url()).pathname, status: response.status(), body: await response.json() }); } catch {}
    }
  });
  try {
    const config = await json(await request.get(runtime.base + '/api/ui-config'));
    expect(config).toEqual({ version: 'goo/ui-config/v1', profile: 'generic', examples: { state: 'unconfigured' } });
    const capabilities = await json(await request.get(runtime.base + '/api/capabilities'));
    expect(capabilities).toEqual({ version: 'goo/runtime-capabilities/v1', profile: 'generic', preparedHTTP: { state: 'unconfigured' }, declaredOperations: { state: 'unconfigured' } });
    const unavailable = [];
    for (const endpoint of ['/api/http-package', '/api/operations', '/api/operation-packages', '/api/example-questions', '/api/prebuilts']) {
      const response = await request.get(runtime.base + endpoint), body = await response.json();
      expect(response.status()).toBe(503); expect(body.code).toBe('capability_unconfigured');
      unavailable.push({ endpoint, status: response.status(), body });
    }
    await page.goto(runtime.base + '/');
    await expect(page.locator('html')).toHaveAttribute('data-profile', 'generic');
    await expect(page.getByRole('button', { name: 'Start stream', exact: true })).toBeVisible();
    await expect(page.getByLabel('Live environment', { exact: true })).toHaveValue('');
    expect(requests.filter(req => req.method === 'POST' || /\/(tail|trace-tail)(?:\/|$)/.test(req.endpoint))).toEqual([]);
    expect(requests.filter(req => ['/api/package', '/api/example-questions', '/api/prebuilts'].includes(req.endpoint))).toEqual([]);
    const startupRequests = structuredClone(requests);
    const traceAck = await json(await request.post('http://127.0.0.1:4318/v1/traces', { data: traces(facts) }));
    const logAck = await json(await request.post('http://127.0.0.1:4318/v1/logs', { data: logs(facts) }));
    let ready;
    await expect.poll(async () => {
      ready = await json(await request.post(runtime.base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } }));
      return ready.count;
    }, { timeout: 60000 }).toBe('150');
    let logsReady;
    await expect.poll(async () => {
      logsReady = await json(await request.post(runtime.base + '/api/explore', { data: { signal: 'logs', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } }));
      return logsReady.count;
    }, { timeout: 60000 }).toBe('140');
    const live = page.getByRole('region', { name: 'Live activity', exact: true });
    await live.getByLabel('Live environment').fill(facts.environment);
    await live.getByLabel('Live signal').selectOption('spans');
    await page.getByRole('button', { name: 'Start stream', exact: true }).click();
    await expect.poll(() => live.locator('.live-tail tbody tr').count()).toBeGreaterThan(0);
    await page.getByRole('button', { name: 'Pause stream', exact: true }).click();
    const activity = await live.locator('.live-tail').evaluate(host => ({ state: host.dataset.state, retained: host.dataset.retained, text: host.textContent }));
    expect(activity.state).toBe('paused'); expect(activity.text).toContain('python-orders');
    await page.screenshot({ path: 'docs/journeys/screenshots/191a-generic-python-activity.png' });
    await page.locator('[data-workspace=services]').click();
    await page.locator('#inventoryEnvironment').fill(facts.environment);
    await page.locator('#minutes').fill('2');
    let pending = page.waitForResponse(response => response.url().endsWith('/api/discovery'));
    await page.locator('#inventoryRun').click(); const discovery = await json(await pending);
    await expect.poll(async () => JSON.parse(await page.locator('#serviceDeclarationsExplain').textContent()).state).toBe('unconfigured');
    pending = page.waitForResponse(response => response.url().endsWith('/api/dashboards/plan'));
    await page.getByRole('button', { name: 'Open service ordinary-python / python-orders', exact: true }).click();
    const plan = await json(await pending);
    await expect(page.locator('.service-dashboard')).toHaveAttribute('data-state', 'loaded');
    expect(plan.profile).toBe('http-server');
    const overview = [...responses].reverse().find(response => response.endpoint === '/api/explore' && response.body.count === '140')?.body;
    expect(overview).toBeTruthy(); expect(overview.rows).toHaveLength(100);
    for (const row of overview.rows) {
      expect(row.ResourceAttributes['telemetry.sdk.language']).toBe('python');
      expect(row.ResourceAttributes['service.namespace']).toBe('ordinary-python');
    }
    await page.screenshot({ path: 'docs/journeys/screenshots/191b-generic-python-service.png' });
    const link = page.locator('[data-widget=spans]').getByRole('link', { name: 'Open widget question', exact: true });
    const staged = readStepLink(await link.getAttribute('href'));
    const beforeStage = requests.filter(req => req.method === 'POST').length;
    await link.click(); await expect(page.locator('#exploreRun')).toBeVisible();
    expect(requests.filter(req => req.method === 'POST')).toHaveLength(beforeStage);
    pending = page.waitForResponse(response => response.url().endsWith('/api/explore') && response.request().method() === 'POST');
    await page.locator('#exploreRun').click(); const result = await json(await pending);
    expect(result.count).toBe('140'); expect(result.rows).toHaveLength(100);
    expect(result.question.environment).toBe(facts.environment);
    await page.locator('#exploreSaveCapture').click();
    await page.locator('#captureDraftTitle').fill('Ordinary Python HTTP evidence');
    pending = page.waitForResponse(response => response.url().endsWith('/api/captures') && response.request().method() === 'POST');
    await page.locator('#captureDraftConfirm').click(); const saved = await json(await pending);
    await expect(page.locator('#savedexploreCount')).toContainText('140');
    const beforeReopen = requests.filter(req => req.method === 'POST').length;
    await page.goto(runtime.base + '/#capture=' + saved.receipt.id);
    await expect(page.locator('#savedexploreCount')).toContainText('140');
    expect(requests.filter(req => req.method === 'POST')).toHaveLength(beforeReopen);
    await fs.copyFile(path.join(runtime.root, 'captures', 'exploration', saved.receipt.id + '.json'), path.join(directory, 'scope-capture.json'));
    const reopened = await json(await request.get(runtime.base + '/api/captures/' + saved.receipt.id));
    expect(reopened).toEqual(saved);
    await page.setViewportSize({ width: 390, height: 1000 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
    await page.screenshot({ path: 'docs/journeys/screenshots/191c-generic-saved-python-evidence.png' });
    const refusals = [];
    for (const args of [['serve-http', '--profile', 'unknown'], ['serve-http', '--profile', 'generic', '--binding', 'missing.yaml'], ['serve-http', '--profile', 'generic', '--operation-package', 'missing.yaml']]) {
      const run = spawnSync(process.env.GOO_GENERIC_BINARY, args, { cwd: runtime.root, env: runtime.environment, encoding: 'utf8', timeout: 5000 });
      expect(run.status).not.toBe(0); expect(run.error).toBeUndefined(); refusals.push({ args, status: run.status, output: run.stderr + run.stdout });
    }
    const malformedListen = spawnSync(process.env.GOO_GENERIC_BINARY, ['serve-http', '--profile', 'generic', '--listen', '127.0.0.1:not-a-port'], { cwd: runtime.root, env: runtime.environment, encoding: 'utf8', timeout: 5000 });
    expect(malformedListen.status).not.toBe(0); expect(malformedListen.error).toBeUndefined();
    refusals.push({ args: ['serve-http', '--listen', '127.0.0.1:not-a-port'], status: malformedListen.status, output: malformedListen.stderr + malformedListen.stdout });
    expect(runtime.receipt.commands).toHaveLength(1);
    expect(runtime.receipt.commands[0].label).toBe('go');
    expect(runtime.receipt.files.every(file => !/server\.py|\.(?:mjs|css|html)$/.test(file.path))).toBe(true);
    expect(errors).toEqual([]);
    await fs.writeFile(path.join(directory, 'ledger.json'), JSON.stringify({ version: 1, facts, config, capabilities, unavailable, startupRequests, traceAck, logAck, ready, logsReady, activity, discovery, plan, overview, staged, result, saved, reopened, requests, responses, refusals, errors }, null, 2));
    console.log('191 ledger:', path.join(directory, 'ledger.json'));
  } finally { await runtime.stop(); }
});
