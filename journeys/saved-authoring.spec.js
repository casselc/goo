const {test, expect} = require('@playwright/test');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const {startRuntime} = require('./191-generic-runtime');
const {openDefinitions} = require('./support/definitions');
const {questionLink} = require('../ui/workspace.mjs');
const {startBombadil} = require('./support/cross-app');

test('Saved owns reusable authoring, contextual Build, and immutable reopen', async ({page, request}) => {
  test.setTimeout(150000);
  page.setDefaultTimeout(15000);
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'goo-saved-authoring-'));
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, dir);
  const errors = [], sources = [];
  const sourcePaths = new Set(['/api/explore', '/api/tail', '/api/operations', '/api/http-package', '/api/measurements']);
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', req => { if (sourcePaths.has(new URL(req.url()).pathname)) sources.push(req.url()); });
  try {
    await openDefinitions(page, runtime.base);
    await expect(page.locator('[data-workspace=definitions]')).toHaveCount(0);
    await expect(page.locator('[data-workspace-panel=definitions]')).toHaveCount(0);
    await expect(page.locator('#savedDefinitionsPanel')).toBeVisible();
    await expect(page.locator('#savedEvidencePanel')).toBeHidden();
    await expect(page.locator('.scope-control')).toBeHidden();
    expect(sources).toEqual([]);

    // Import, validate, save, and reopen through the real Go package APIs.
    const now = Date.now();
    const question = {signal:'logs', environment:'local', fromMs:now-60000, toMs:now, limit:100, offset:0};
    const draft = await request.post(runtime.base + '/api/telemetry-packages/draft', {
      data:{question, name:'saved-authoring', title:'Saved authoring', lookbackMinutes:1},
    });
    expect(draft.ok()).toBe(true);
    const definition = await draft.json();
    await page.locator('#packageFile').setInputFiles({name:'authoring.yaml', mimeType:'application/yaml', buffer:Buffer.from(definition.yaml)});
    await expect(page.locator('#packageSave')).toBeEnabled();
    const savedResponse = page.waitForResponse(r => r.request().method()==='POST' && new URL(r.url()).pathname==='/api/telemetry-packages');
    await page.locator('#packageSave').click();
    expect((await savedResponse).status()).toBe(201);
    await expect(page.locator('#packageStatus')).toContainText('Saved package revision');
    expect(sources).toEqual([]);

    await page.locator('#packageYAML').fill('# unsaved draft\n' + definition.yaml);
    await page.locator('#definitionsRecords').focus();
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('#definitionsMeasurements')).toBeFocused();
    await page.locator('#savedEvidenceTab').click();
    await expect(page.locator('#savedDefinitionsPanel')).toBeHidden();
    const before = sources.length;
    // Even a stale direct handler invocation cannot read from a hidden editor.
    await page.locator('#packagePreview').evaluate(button => button.onclick?.());
    expect(sources.length).toBe(before);
    await page.locator('#savedDefinitionsTab').click();
    await page.locator('#definitionsRecords').click();
    await expect(page.locator('#packageYAML')).toHaveValue('# unsaved draft\n' + definition.yaml);
    await page.locator('#packageClose').click();
    await page.locator('#telemetryPackageLibrary .definition-row button').filter({hasText:'Saved authoring'}).first().click();
    await page.getByRole('button', {name:'Edit saved-authoring', exact:true}).click();
    await expect(page.locator('#packageYAML')).toHaveValue(definition.yaml);

    const download = page.waitForEvent('download');
    await page.locator('#packageExport').click();
    expect(await fs.readFile(await (await download).path(), 'utf8')).toBe(definition.yaml);
    const readsBeforeStage = sources.length;
    await page.locator('#packageUse').click();
    await expect(page.locator('[data-workspace-panel=explore]')).toBeVisible();
    expect(sources.length).toBe(readsBeforeStage);

    // A question's existing Build action enters this same Saved editor.
    await page.goto(runtime.base + '/' + questionLink(question));
    await page.locator('#exploreRun').click();
    await expect(page.locator('#exploreRun')).toBeEnabled();
    const editQuestion = page.getByRole('button', {name:'Edit question', exact:true});
    if (await editQuestion.isVisible()) await editQuestion.click();
    const reusable = page.locator('summary').filter({hasText:/^Reusable question package$/});
    await reusable.click();
    const readsBeforeBuild = sources.length;
    await page.locator('#exploreCreatePackage').click();
    await expect(page.locator('#savedDefinitionsPanel')).toBeVisible();
    await expect(page.locator('#packageName')).toBeFocused();
    expect(sources.length).toBe(readsBeforeBuild);
    await page.locator('#packageTitle').fill('From captured question');
    await page.locator('#packageBuild').click();
    await expect(page.locator('#packageSave')).toBeEnabled();
    expect(sources.length).toBe(readsBeforeBuild);

    for (const width of [1440, 390, 320]) {
      await page.setViewportSize({width, height:1000});
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
      await page.screenshot({path:`docs/journeys/screenshots/saved-authoring-${width}.png`});
    }
    // A Saved link reveals evidence even while authoring is selected. An unknown
    // annotation still gets the normal Saved error/recovery surface.
    await page.evaluate(() => location.hash = '#annotation=' + 'f'.repeat(64));
    await expect(page.locator('#savedEvidencePanel')).toBeVisible();
    await expect(page.locator('#savedDefinitionsPanel')).toBeHidden();
    await expect(page.locator('.scope-control')).toBeHidden();
    expect(errors).toEqual([]);
    const actor = await startBombadil({dir:path.join(dir,'bombadil'), url:runtime.base+'/#investigations',
      specification:'verification/bombadil/workspace.ts', seconds:30, width:390, height:1000});
    try { await actor.finish(); } finally { await actor.stop(); }
  } finally {
    await runtime.stop();
    await fs.rm(dir, {recursive:true, force:true});
  }
});
