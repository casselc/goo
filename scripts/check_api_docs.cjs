// Documentation-only browser journey against this checkout's actual Python
// handler, on an ephemeral loopback port; does not touch the running lab.
const { spawn } = require('node:child_process');
const { once } = require('node:events');
const { chromium } = require('playwright');
const assert = require('node:assert/strict');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
(async () => {
  const remote = process.env.GOO_API_DOCS_URL;
  const server = remote ? null : spawn(process.env.PYTHON || 'python3', ['-u', '-c',
    'from server import Handler, ThreadingHTTPServer; s=ThreadingHTTPServer(("127.0.0.1",0),Handler); print(s.server_port,flush=True); s.serve_forever()'], { cwd: root, stdio: ['ignore', 'pipe', 'inherit'] });
  let browser;
  try {
    const port = remote ? 0 : await Promise.race([once(server.stdout, 'data').then(([b]) => Number(String(b).trim())), once(server, 'exit').then(([code]) => { throw Error(`Server exited ${code}`); })]);
    assert.ok(remote || port > 0);
    browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox'] });
    const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
    const errors = []; page.on('pageerror', e => errors.push(e.message));
    await page.goto(remote || `http://127.0.0.1:${port}/api-docs`);
    await page.getByRole('status').filter({ hasText: '64 operations' }).waitFor();
    assert.equal(await page.locator('#operations > details').count(), 64);
    await page.getByRole('searchbox').fill('/tail');
    assert.equal(await page.locator('#operations > details').count(), 1);
    await page.locator('#operations > details > summary').click();
    await page.getByText('Operation contract', { exact: true }).click();
    await page.getByText('text/event-stream', { exact: false }).last().waitFor();
    assert.match(await page.locator('#operations').innerText(), /source-read/);
    await page.screenshot({ path: path.join(root, 'docs/journeys/screenshots/api-docs-tail.png'), fullPage: false });
    await page.getByRole('searchbox').fill('/measurements');
    await page.locator('#operations > details > summary').click();
    assert.match(await page.locator('#operations').innerText(), /65,536/);
    await page.setViewportSize({ width: 390, height: 844 });
    await page.getByRole('searchbox').fill('/compare');
    assert.equal(await page.locator('#operations > details').count(), 2); // saved capture compare and supplied-summary compare
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    assert.deepEqual(errors, []);
    console.log('API docs browser journey passes: 64 operations, filter/disclosure, limits, desktop screenshot, mobile overflow, no page errors');
  } finally {
    if (browser) await browser.close();
    if (server) server.kill('SIGTERM');
  }
})().catch(e => { console.error(e); process.exitCode = 1; });
