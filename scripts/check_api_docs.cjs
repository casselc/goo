// Documentation-only journey against this checkout's embedded Go runtime.
// Uses an isolated generic server unless an explicit running URL is supplied.
const { spawn, execFile } = require('node:child_process');
const { promisify } = require('node:util');
const fs = require('node:fs/promises');
const os = require('node:os');
const net = require('node:net');
const { once } = require('node:events');
const { chromium } = require('playwright');
const assert = require('node:assert/strict');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
const operations = Object.values(require('../api/openapi.json').paths).reduce((n,item)=>n+Object.keys(item).filter(m=>['get','post'].includes(m)).length,0);
(async () => {
  const remote = process.env.GOO_API_DOCS_URL;
  let server, browser, temporary;
  try {
    let base = remote;
    if (!remote) {
      temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'goo-api-docs-'));
      const binary = path.join(temporary, 'goo');
      await promisify(execFile)('go', ['build', '-mod=readonly', '-o', binary, './cmd/goo'], { cwd: root, timeout: 120000 });
      const reservation = net.createServer();
      reservation.listen(0, '127.0.0.1');
      await once(reservation, 'listening');
      const port = reservation.address().port;
      await new Promise(resolve => reservation.close(resolve));
      server = spawn(binary, ['serve-http', '--profile', 'generic', '--listen', `127.0.0.1:${port}`, '--investigations-dir', path.join(temporary, 'evidence')], { cwd: temporary, stdio: ['ignore', 'ignore', 'inherit'] });
      base = `http://127.0.0.1:${port}/api-docs`;
      let ready = false;
      for (let attempt = 0; attempt < 100; attempt++) {
        if (server.exitCode !== null) throw Error(`Go server exited ${server.exitCode}`);
        try { ready = (await fetch(`http://127.0.0.1:${port}/health`)).ok; } catch {}
        if (ready) break;
        await new Promise(resolve => setTimeout(resolve, 50));
      }
      assert.ok(ready, 'Go docs server did not become ready');
    }
    browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox'] });
    const page = await browser.newPage({ viewport: { width: 1280, height: 900 } });
    const errors = []; page.on('pageerror', e => errors.push(e.message));
    await page.goto(base);
    await page.getByRole('status').filter({ hasText: operations+' operations' }).waitFor();
    assert.equal(await page.locator('#operations > details').count(), operations);
    await page.getByRole('searchbox').fill('/tail');
    assert.equal(await page.locator('#operations > details').count(), 2); // stream and exact-retention tail routes
    const tailOp=page.locator('#operations > details').filter({has:page.locator('summary > code').filter({hasText:/^\/tail$/})});
    await tailOp.locator(':scope > summary').click();
    await tailOp.getByText('Operation contract', { exact: true }).click();
    await tailOp.getByText('text/event-stream', { exact: false }).last().waitFor();
    assert.match(await page.locator('#operations').innerText(), /source-read/);
    await page.screenshot({ path: path.join(root, 'docs/journeys/screenshots/api-docs-tail.png'), fullPage: false });
    await page.getByRole('searchbox').fill('/measurements');
    await page.locator('#operations > details > summary').click();
    assert.match(await page.locator('#operations').innerText(), /65,536/);
    await page.getByRole('searchbox').fill('/retained-evidence/{id}/content');
    assert.equal(await page.locator('#operations > details').count(),1);
    await page.locator('#operations > details > summary').click();
    await page.getByText('Operation contract',{exact:true}).click();
    assert.match(await page.locator('#operations').innerText(),/EvidenceContent/);
    await page.getByRole('searchbox').fill('/population/explain');
    assert.equal(await page.locator('#operations > details').count(), 3);
    for (const operation of await page.locator('#operations > details').all()) {
      await operation.locator(':scope > summary').click();
      assert.match(await operation.innerText(), /65,536/);
      await operation.getByText('Operation contract', { exact: true }).click();
      const route = await operation.locator('summary > code').first().innerText();
      assert.match(await operation.innerText(), route === '/population/explain' ? /source-read/ : /pure/);
    }
    await page.setViewportSize({ width: 390, height: 844 });
    await page.getByRole('searchbox').fill('/compare');
    assert.equal(await page.locator('#operations > details').count(), 2); // saved capture compare and supplied-summary compare
    assert.equal(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth), true);
    assert.deepEqual(errors, []);
    console.log(`API docs browser journey passes: ${operations} operations, origin-aware content, filter/disclosure, limits, desktop screenshot, mobile overflow, no page errors`);
  } finally {
    if (browser) await browser.close();
    if (server && server.exitCode === null) {
      const exited = once(server, 'close');
      server.kill('SIGTERM');
      await exited;
    }
    if (temporary) await fs.rm(temporary, {recursive: true, force: true});
  }
})().catch(e => { console.error(e); process.exitCode = 1; });
