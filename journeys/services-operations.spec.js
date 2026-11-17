const {test, expect} = require('@playwright/test');
const {spawn} = require('node:child_process');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const {startBombadil} = require('./support/cross-app');

test('Operations is a Services tab with the same read and saved-capture flow', async ({page, request}) => {
  test.setTimeout(120000);
  const binary = process.env.GOO_GENERIC_BINARY;
  if (!binary) throw Error('Build and set GOO_GENERIC_BINARY');

  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'goo-services-ops-'));
  const base = 'http://127.0.0.1:18110';
  const errors = [], reads = [];
  let catalogs = 0, output = '';
  const sourcePaths = new Set(['/api/operations', '/api/counter-rates', '/api/histograms']);
  const proc = spawn(binary, [
    'serve-http', '--profile', 'generic', '--listen', '127.0.0.1:18110',
    '--investigations-dir', dir,
    '--operation-package', 'config/operations/checkout.yaml',
    '--operation-package', 'config/operations/reports.yaml',
  ], {
    cwd: path.resolve('.'),
    env: {...process.env, CLICKHOUSE_URL:'http://127.0.0.1:8123', CLICKHOUSE_USER:'lab', CLICKHOUSE_PASSWORD:'lab-local'},
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  proc.stdout.on('data', value => output += value);
  proc.stderr.on('data', value => output += value);
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    const pathname = new URL(request.url()).pathname;
    if (pathname === '/api/operation-packages') catalogs++;
    if (sourcePaths.has(pathname)) reads.push(pathname);
  });

  try {
    await expect.poll(async () => {
      if (proc.exitCode !== null) throw Error(output);
      try { return (await request.get(base + '/api/ui-config')).status(); }
      catch { return 0; }
    }, {timeout:30000}).toBe(200);

    await page.goto(base + '/#services');
    await expect(page.locator('[data-workspace=operations]')).toHaveCount(0);
    await expect(page.getByRole('tab', {name:'Operations', exact:true})).toBeVisible();
    expect(reads).toEqual([]);
    expect(catalogs).toBe(0);

    await page.getByRole('tab', {name:'Operations', exact:true}).click();
    await expect(page.locator('#operations')).toBeVisible();
    await expect(page.locator('#workspaceTitle')).toHaveText('Services & resources');
    await expect(page.locator('#operationPackage option')).toHaveCount(2);
    expect(catalogs).toBe(1);
    expect(reads).toEqual([]);
    await expect(page.getByLabel('New operation read lookback minutes', {exact:true})).toBeVisible();

    const response = page.waitForResponse(r => new URL(r.url()).pathname === '/api/operations');
    await page.locator('#runOperations').click();
    expect((await response).status()).toBe(200);
    await expect(page.locator('#runOperations')).toBeEnabled();
    await expect(page.locator('#operationMode')).toContainText('Captured operation query');

    for (const width of [1440, 390, 320]) {
      await page.setViewportSize({width, height:1000});
      await page.evaluate(() => scrollTo(0, 0));
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
      await page.screenshot({path:'docs/journeys/screenshots/services-operations-' + width + '.png'});
    }

    await page.setViewportSize({width:1440, height:1000});
    const save = page.waitForResponse(r => new URL(r.url()).pathname === '/api/investigations' && r.request().method() === 'POST');
    await page.locator('#captureSave').click();
    expect((await save).status()).toBe(201);
    await expect(page.locator('#savedInvestigations')).toBeVisible();
    const before = reads.length;
    await page.locator('#savedOperationLive').click();
    await expect(page.getByRole('tab', {name:'Operations', exact:true})).toHaveAttribute('aria-selected', 'true');
    await expect(page.locator('#operations')).toBeVisible();
    expect(reads.length).toBe(before);

    await page.getByRole('tab', {name:'Services', exact:true}).click();
    await expect(page.locator('#operations')).toBeHidden();
    await expect(page.getByLabel('New discovery lookback minutes', {exact:true})).toBeVisible();
    await page.getByRole('tab', {name:'Operations', exact:true}).focus();
    await page.keyboard.press('Enter');
    await expect(page.locator('#operations')).toBeVisible();
    expect(reads.length).toBe(before);
    expect(errors).toEqual([]);

    // Hold a real response so leaving the tab must withdraw its result lease.
    let entered, release;
    const started = new Promise(resolve => entered = resolve);
    const gate = new Promise(resolve => release = resolve);
    await page.route('**/api/operations?*', async route => {
      const response = await route.fetch();
      entered();
      await gate;
      try { await route.fulfill({response}); } catch {}
    });
    await page.locator('#runOperations').click();
    await started;
    // Pending reads must not retain the preceding result's metadata.
    await expect(page.locator('#operationPackageTitle')).toBeEmpty();
    await expect(page.locator('#operationExplain')).toBeEmpty();
    await expect(page.locator('#operationMode')).toHaveText('No operation evidence loaded.');
    await page.getByRole('tab', {name:'Hosts', exact:true}).click();
    await expect(page.locator('#operationStatus')).toContainText('Read stopped');
    release();
    await page.unrouteAll({behavior:'wait'});
    await page.getByRole('tab', {name:'Operations', exact:true}).click();
    await expect(page.locator('#operationView')).toBeEmpty();
    expect(errors).toEqual([]);

    const actor = await startBombadil({
      dir:path.join(dir, 'bombadil'), url:base + '/#services',
      specification:'verification/bombadil/workspace.ts', seconds:30, width:390, height:1000,
    });
    try { await actor.finish(); }
    finally { await actor.stop(); }
  } finally {
    proc.kill('SIGTERM');
    await new Promise(resolve => {
      if (proc.exitCode !== null) return resolve();
      proc.once('exit', resolve);
    });
    await fs.rm(dir, {recursive:true, force:true});
  }
});
