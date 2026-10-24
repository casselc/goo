#!/usr/bin/env node
'use strict';

// Explicit modes: mobile reopens frozen comparison evidence without source reads;
// empty runs the retained current scope once and then its two-scope comparison.
const {chromium, expect} = require('@playwright/test');
const {questionLink} = require('../ui/workspace');
const fs = require('node:fs/promises'), path = require('node:path');
const {parseArgs, promisify} = require('node:util');
const {execFile} = require('node:child_process');
const {createHash} = require('node:crypto');
const repository = path.resolve(__dirname, '..');
const base = (process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088').replace(/\/$/, '');
const isResponse = (endpoint, response) => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/' + endpoint;
async function json(response) {expect(response.ok(), await response.text()).toBe(true); return response.json()}
function hash(bytes) {return createHash('sha256').update(bytes).digest('hex')}
async function stored(id) {
  expect(id).toMatch(/^[a-f0-9]{64}$/);
  const environment = {...process.env}; for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete environment[key];
  const {stdout} = await promisify(execFile)('docker', ['--host=unix:///var/run/docker.sock', 'exec', process.env.GOO_QUERY_CONTAINER || 'grammar-observability-packages-1', 'cat', '/app/artifacts/investigations/exploration/' + id + '.json'], {env: environment, encoding: 'buffer', maxBuffer: 4 * 1024 * 1024});
  return stdout;
}
function watch(page) {
  const requests = [], errors = [];
  page.on('pageerror', error => errors.push(error.message));
  page.on('request', request => {
    if (request.method() !== 'POST') return;
    const endpoint = new URL(request.url()).pathname;
    if (endpoint.startsWith('/api/')) requests.push({endpoint, payload: request.postDataJSON(), effect: endpoint === '/api/compare' ? 'pure-comparison' : ['/api/explore', '/api/captures', '/api/tail', '/api/inventory', '/api/discovery', '/api/measurements'].includes(endpoint) ? 'source' : 'other'});
  });
  return {requests, errors, source: () => requests.filter(request => request.effect === 'source')};
}
async function context(browser, viewport) {
  const context = await browser.newContext({viewport}), page = await context.newPage();
  page.setDefaultTimeout(15000); page.setDefaultNavigationTimeout(45000);
  return {context, page, effects: watch(page)};
}
function exactRows(section) {
  return section.locator('tbody>tr[data-pattern]').evaluateAll(rows => rows.map(row => ({...row.dataset, cells: [...row.querySelectorAll('.comparison-bar-cell')].map(cell => ({value: cell.dataset.value, scaleMax: cell.dataset.scaleMax, exact: cell.querySelector('.comparison-value').title}))})));
}
async function safeFrame(section, file) {
  await section.evaluate(element => {
    element.scrollIntoView({block: 'start'});
    const overlays = [...document.querySelectorAll('.investigation-trail,.workspace-header')].filter(node => {
      const style = getComputedStyle(node); return node.getClientRects().length > 0 && ['fixed', 'sticky'].includes(style.position);
    });
    const margin = Math.max(120, ...overlays.map(node => node.getBoundingClientRect().bottom + 16));
    const top = element.getBoundingClientRect().top;
    // At document bottom the browser may already place the component well below
    // sticky content. Moving it down again would push its last rows out of view.
    if (top < margin) window.scrollBy(0, top - margin);
  });
  const geometry = await section.evaluate(element => {
    const rect = element.getBoundingClientRect(), title = element.querySelector('h4').getBoundingClientRect(), control = element.querySelector('[aria-label="Pattern comparison measure"]').getBoundingClientRect();
    const overlaps = [...document.querySelectorAll('.investigation-trail,.workspace-header')].filter(node => {
      const style = getComputedStyle(node); return node.getClientRects().length > 0 && ['fixed', 'sticky'].includes(style.position);
    }).map(node => ({className: node.className, bottom: node.getBoundingClientRect().bottom}));
    return {top: rect.top, bottom: rect.bottom, titleTop: title.top, titleBottom: title.bottom, controlTop: control.top, controlBottom: control.bottom, viewportHeight: innerHeight, overlays: overlaps, pageOverflow: document.documentElement.scrollWidth > innerWidth};
  });
  const overlayBottom = Math.max(0, ...geometry.overlays.map(overlay => overlay.bottom));
  expect(geometry.titleTop).toBeGreaterThanOrEqual(overlayBottom + 1);
  expect(geometry.controlTop).toBeGreaterThanOrEqual(overlayBottom + 1);
  expect(geometry.bottom).toBeLessThanOrEqual(geometry.viewportHeight);
  expect(geometry.pageOverflow).toBe(false);
  await section.screenshot({path: path.join(repository, file)});
  return geometry;
}
async function mobile(browser, ledger, directory) {
  const originalBefore = await stored(ledger.scopeCapture.receipt.id);
  const custody = JSON.parse(await fs.readFile(path.join(directory, 'capture-custody.json'), 'utf8'));
  expect(hash(originalBefore)).toBe(custody.sha256);
  const {context: owner, page, effects} = await context(browser, {width: 390, height: 3600});
  try {
    const pending = page.waitForResponse(response => response.request().method() === 'GET' && new URL(response.url()).pathname === '/api/captures/' + ledger.saved.receipt.id);
    await page.goto(base + '/#capture=' + ledger.saved.receipt.id);
    const reopened = await json(await pending);
    expect(reopened).toEqual(ledger.saved);
    const section = page.locator('.pattern-time-comparison:visible');
    await expect(section).toHaveAttribute('data-state', 'known');
    await section.getByRole('combobox', {name: 'Pattern comparison measure', exact: true}).selectOption('rate');
    const rows = await exactRows(section);
    expect(rows).toHaveLength(ledger.frozenView.rows.length);
    for (const row of rows) {
      const expected = ledger.frozenView.rows.find(value => value.pattern === row.pattern);
      for (const key of ['baselineCount', 'currentCount', 'baselineRate', 'currentRate', 'baselineShare', 'currentShare', 'countDifference', 'rateDifference', 'shareDifference', 'presence']) expect(row[key]).toBe(expected[key]);
      expect(row.cells[0].scaleMax).toBe(row.cells[1].scaleMax);
    }
    const file = 'docs/journeys/screenshots/183b-mobile-pattern-time-comparison.png';
    const geometry = await safeFrame(section, file);
    expect(effects.source()).toEqual([]); expect(effects.errors).toEqual([]);
    const after = await stored(ledger.scopeCapture.receipt.id);
    expect(after.equals(originalBefore)).toBe(true);
    const result = {version: 1, mode: 'source-free frozen comparison mobile', captureID: ledger.saved.receipt.id, request: reopened.comparison.request, rows, geometry, requests: effects.requests, sourceOperations: 0, originalCaptureHash: hash(after), originalCaptureUnchanged: true, file, capturedAt: new Date().toISOString()};
    await fs.writeFile(path.join(directory, 'mobile-screenshot-replay.json'), JSON.stringify(result, null, 2) + '\n');
    console.log(JSON.stringify({mode: result.mode, sourceOperations: 0, geometry, originalCaptureUnchanged: true, file}));
  } finally {await owner.close()}
}
async function empty(browser, ledger, directory) {
  const originalBefore = await stored(ledger.scopeCapture.receipt.id);
  const {context: owner, page, effects} = await context(browser, {width: 1440, height: 1800});
  try {
    await page.goto(base + '/' + questionLink(ledger.current.question));
    expect(effects.source()).toEqual([]);
    let pending = page.waitForResponse(response => isResponse('explore', response));
    await page.locator('#exploreRun').click();
    const current = await json(await pending);
    expect(current.question).toEqual(ledger.current.question);
    expect(current.count).toBe(ledger.current.count);
    const groups = values => values.map(group => ({g0: group.g0, count: group.count})).sort((a, b) => a.g0.localeCompare(b.g0));
    expect(groups(current.groups)).toEqual(groups(ledger.current.groups));
    const choice = page.locator('.pattern-time-choice:visible');
    // Record actual live entry adjacency without changing the UI or reading again.
    await page.locator('#comparePrevious').evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -160)});
    const entries = await page.locator('#comparePrevious,.pattern-time-choice>summary').evaluateAll(elements => elements.map(element => {
      const box = element.getBoundingClientRect(), style = getComputedStyle(element);
      return {tag: element.localName, id: element.id, text: element.textContent, visible: element.getClientRects().length > 0 && style.visibility !== 'hidden', top: box.top, bottom: box.bottom, left: box.left, right: box.right};
    }));
    if (!await choice.evaluate(element => element.open)) await choice.locator(':scope>summary').click();
    await choice.getByRole('button', {name: 'Stage pattern comparison', exact: true}).click();
    expect(effects.source()).toHaveLength(1);
    pending = page.waitForResponse(response => isResponse('compare', response));
    await page.locator('.population-run:visible').click();
    const response = await pending, pair = {request: response.request().postDataJSON(), response: await json(response)};
    expect(pair.request.baseline.question).toEqual(ledger.pair.request.baseline.question);
    expect(pair.request.selection.question).toEqual(ledger.pair.request.selection.question);
    expect(pair.request.baseline.count).toBe('0'); expect(pair.request.baseline.state).toBe('no_data'); expect(pair.request.selection.count).toBe('400');
    expect(pair.response).toMatchObject({state: 'no_data', baselineCount: '0', selectionCount: '400', rows: []});
    const section = page.locator('.pattern-time-comparison:visible');
    await expect(section).toHaveAttribute('data-state', 'known');
    await section.getByRole('combobox', {name: 'Pattern comparison measure', exact: true}).selectOption('share');
    await expect(section).toContainText('Share differences unavailable; stable pattern order.');
    await expect(section).not.toContainText('largest absolute share change first');
    const rows = await exactRows(section);
    for (const row of rows) {expect(row.baselineRate).toBe('0'); expect(row.baselineShare).toBe('unavailable'); expect(row.shareDifference).toBe('unavailable'); expect(row.presence).toBe('not-observed-in-baseline')}
    const file = 'docs/journeys/screenshots/183c-empty-baseline-share-unavailable.png';
    const geometry = await safeFrame(section, file);
    expect(effects.source()).toHaveLength(3); expect(effects.errors).toEqual([]);
    expect(effects.source().map(request => request.endpoint)).toEqual(['/api/explore', '/api/explore', '/api/explore']);
    const after = await stored(ledger.scopeCapture.receipt.id);
    expect(after.equals(originalBefore)).toBe(true);
    const result = {version: 1, mode: 'explicit successful empty-baseline reread', current, pair, rows, entries, geometry, requests: effects.requests, sourceOperations: 3, originalCaptureHash: hash(after), originalCaptureUnchanged: true, file, capturedAt: new Date().toISOString()};
    await fs.writeFile(path.join(directory, 'empty-screenshot-replay.json'), JSON.stringify(result, null, 2) + '\n');
    console.log(JSON.stringify({mode: result.mode, sourceOperations: 3, entries, geometry, originalCaptureUnchanged: true, file}));
  } finally {await owner.close()}
}
async function main() {
  const {values} = parseArgs({options: {ledger: {type: 'string'}, 'empty-ledger': {type: 'string'}, mode: {type: 'string'}, help: {type: 'boolean'}}});
  const usage = 'node journeys/183-visual-replay.js --ledger <retained main ledger.json> --mode mobile|empty|both [--empty-ledger <retained empty.json>]';
  if (values.help) {console.log(usage); return}
  if (!values.ledger || !['mobile', 'empty', 'both'].includes(values.mode) || values.mode === 'both' && !values['empty-ledger']) throw Error(usage);
  const ledgerPath = path.resolve(values.ledger), ledger = JSON.parse(await fs.readFile(ledgerPath, 'utf8'));
  const emptyPath = values['empty-ledger'] ? path.resolve(values['empty-ledger']) : ledgerPath;
  const browser = await chromium.launch({executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox']});
  try {
    if (values.mode === 'mobile' || values.mode === 'both') await mobile(browser, ledger, path.dirname(ledgerPath));
    if (values.mode === 'empty' || values.mode === 'both') await empty(browser, JSON.parse(await fs.readFile(emptyPath, 'utf8')), path.dirname(emptyPath));
  } finally {await browser.close()}
}
main().catch(error => {console.error(error); process.exitCode = 1});
