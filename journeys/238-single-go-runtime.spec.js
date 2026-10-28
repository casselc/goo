const {test, expect} = require('@playwright/test');
const {randomBytes} = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const os = require('node:os');
const {execFile} = require('node:child_process');
const {promisify} = require('node:util');
const {startRuntime} = require('./191-generic-runtime');
const {source} = require('./support/cross-app');

// This acceptance intentionally runs the binary from an empty working directory.
// The only application process is Go; assets and API requests share its listener.
test('238 one Go process serves embedded UI, live OTel, saved evidence and cancellable SSE', async ({page, request}) => {
  test.setTimeout(150000);
  const buildDirectory = process.env.GOO_GENERIC_BINARY ? null : await fs.mkdtemp(path.join(os.tmpdir(), 'goo-runtime-build-238-'));
  const binary = process.env.GOO_GENERIC_BINARY || path.join(buildDirectory, 'goo');
  if (buildDirectory) {
    try {
      await promisify(execFile)(process.env.GOO_TEST_GO || 'go', ['build', '-o', binary, './cmd/goo'], {cwd: path.resolve('.'), timeout: 60000});
    } catch (error) {
      await fs.rm(buildDirectory, {recursive: true, force: true});
      throw error;
    }
  }
  const token = randomBytes(4).toString('hex'), environment = 'single-go-' + token;
  const directory = path.resolve('artifacts/single-go-runtime', environment);
  let runtime;
  try { runtime = await startRuntime(binary, directory); }
  catch (error) {
    if (buildDirectory) await fs.rm(buildDirectory, {recursive: true, force: true});
    throw error;
  }
  const errors = [], failures = [], calls = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('response', response => {
    if (/\.(mjs|css)(?:\?|$)/.test(response.url()) && !response.ok()) failures.push({url: response.url(), status: response.status()});
  });
  page.on('request', req => {
    const url = new URL(req.url());
    if (url.pathname.startsWith('/api/')) calls.push({method: req.method(), path: url.pathname, origin: url.origin});
  });
  const attrs = values => Object.entries(values).map(([key, stringValue]) => ({key, value: {stringValue}}));
  const observedAt = Date.now() - 500;
  const body = 'one Go entry point ' + token;
  try {
    expect(runtime.receipt.commands).toHaveLength(1);
    expect(runtime.receipt.commands[0].label).toBe('go');
    expect(runtime.receipt.files).toEqual([]);
    for (const endpoint of ['/api/capabilities', '/capabilities']) {
      const response = await request.get(runtime.base + endpoint);
      expect(response.ok(), await response.text()).toBe(true);
      expect(await response.json()).toMatchObject({profile: 'generic', preparedHTTP: {state: 'unconfigured'}});
    }
    for (const endpoint of ['/api/package', '/api/query', '/api/explain', '/api/not-a-route', '/server.py', '/grammar.py']) {
      const response = endpoint === '/api/query' || endpoint === '/api/explain'
        ? await request.post(runtime.base + endpoint, {data: {view: 'logs'}})
        : await request.get(runtime.base + endpoint);
      expect(response.status(), endpoint).toBe(404);
    }
    const spec = await request.get(runtime.base + '/openapi.json');
    expect(spec.ok()).toBe(true);
    expect((await spec.json()).paths['/tail']).toBeTruthy();
    const doc = await request.get(runtime.base + '/api-docs');
    expect(await doc.text()).toContain("fetch('/openapi.json')");
    const ack = await request.post('http://127.0.0.1:4318/v1/logs', {data: {resourceLogs: [{resource: {attributes: attrs({'service.name': 'runtime-acceptance', 'deployment.environment.name': environment})}, scopeLogs: [{logRecords: [{timeUnixNano: String(BigInt(observedAt) * 1000000n + 123n), severityText: 'WARN', body: {stringValue: body}}]}]}]}});
    expect(ack.ok(), await ack.text()).toBe(true);
    await expect.poll(async () => Number((await source(request, "SELECT count() AS n FROM default.otel_logs WHERE ResourceAttributes['deployment.environment.name']={environment:String} FORMAT JSONEachRow", {param_environment: environment}))[0].n), {timeout: 45000}).toBe(1);
    await page.goto(runtime.base + '/');
    await expect(page.locator('html')).toHaveAttribute('data-profile', 'generic');
    const entry = page.locator('.stream-entry'), tail = entry.locator('.live-tail');
    await expect(entry.getByRole('button', {name: 'Start stream', exact: true})).toBeVisible();
    expect(calls.filter(call => call.method === 'POST')).toEqual([]);
    await entry.getByLabel('Live environment', {exact: true}).fill(environment);
    await entry.getByRole('button', {name: 'Start stream', exact: true}).click();
    await expect(tail).toContainText(body);
    const readsBeforeInspect = calls.filter(call => call.method === 'POST').length;
    await tail.getByRole('button', {name: 'Inspect', exact: true}).click();
    await expect(tail).toHaveAttribute('data-state', 'paused');
    await expect(entry.locator(':scope>.ui-inspector')).toContainText('runtime-acceptance');
    expect(calls.filter(call => call.method === 'POST')).toHaveLength(readsBeforeInspect);
    await fs.mkdir('docs/journeys/screenshots', {recursive: true});
    await page.screenshot({path: 'docs/journeys/screenshots/238-single-go-activity.png'});
    expect(calls.every(call => call.origin === runtime.base)).toBe(true);
    expect(calls.some(call => ['/api/package', '/api/query', '/api/explain'].includes(call.path))).toBe(false);
    await expect(page.locator('#legacyExamples')).toHaveCount(0);
    expect(failures).toEqual([]);
    expect(errors).toEqual([]);

    // A deliberate capture rereads the fixed question; reopening it must not.
    const question = {signal: 'logs', environment, fromMs: observedAt - 1000, toMs: observedAt + 1000, limit: 100, offset: 0};
    const captureResponse = await request.post(runtime.base + '/api/captures', {data: {title: 'Single Go runtime ' + token, question: {kind: 'explore', explore: question}}});
    expect(captureResponse.ok(), await captureResponse.text()).toBe(true);
    const capture = await captureResponse.json();
    expect(capture.explore.count).toBe('1');
    const beforeReopen = calls.filter(call => call.method === 'POST').length;
    await page.goto(runtime.base + '/#capture=' + capture.receipt.id);
    await expect(page.locator('#savedexploreCount')).toContainText('1');
    await expect(page.locator('#captureExploreEvidence')).toContainText(body);
    expect(calls.filter(call => call.method === 'POST')).toHaveLength(beforeReopen);

    // Prove abort releases the actual Go SSE slot, without a proxy disconnect
    // watcher or a synthetic stream. Four readers fill the dedicated tail budget.
    const controllers = [], readers = [];
    async function stream() {
      const controller = new AbortController(); controllers.push(controller);
      const response = await fetch(runtime.base + '/api/tail', {method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({environment}), signal: controller.signal});
      if (response.status !== 200) { await response.text(); return response.status; }
      expect(response.headers.get('content-type')).toContain('text/event-stream');
      const reader = response.body.getReader(); readers.push(reader);
      const first = await reader.read(); expect(first.done).toBe(false);
      return response.status;
    }
    try {
      for (let index = 0; index < 4; index++) expect(await stream()).toBe(200);
      expect(await stream()).toBe(429);
      controllers[0].abort(); await readers[0].cancel().catch(() => {});
      await expect.poll(stream, {timeout: 5000}).toBe(200);
    } finally {
      controllers.forEach(controller => controller.abort());
      await Promise.all(readers.map(reader => reader.cancel().catch(() => {})));
    }
    expect(errors).toEqual([]);
    await fs.writeFile(path.join(directory, 'acceptance.json'), JSON.stringify({version: 1, environment, processCount: runtime.receipt.commands.length, calls, errors, failures, capture: capture.receipt.id}, null, 2));
  } finally {
    try { await runtime.stop(); }
    finally { if (buildDirectory) await fs.rm(buildDirectory, {recursive: true, force: true}); }
  }
});
