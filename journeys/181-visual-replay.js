#!/usr/bin/env node
'use strict';

// Replays retained real evidence. Modes are explicit: saved-desktop performs no
// telemetry reads; resource-right-pan performs exactly one named Explore read.
const {chromium, expect} = require('@playwright/test');
const {questionLink} = require('../ui/workspace');
const fs = require('node:fs/promises');
const path = require('node:path');
const {createHash} = require('node:crypto');
const {parseArgs} = require('node:util');
const repository = path.resolve(__dirname, '..');
const base = (process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088').replace(/\/$/, '');
function human(coefficient, scale) {
  const n = BigInt(coefficient), sign = n < 0n ? '-' : '';
  const digits = (n < 0n ? -n : n).toString().padStart(scale + 1, '0');
  return sign + digits.slice(0, -scale) + '.' + digits.slice(-scale);
}

async function savedDesktop(browser, ledger, directory) {
  const environment = ledger.environment;
  const context = await browser.newContext({viewport: {width: 1920, height: 1600}});
  const page = await context.newPage();
  page.setDefaultTimeout(15000); page.setDefaultNavigationTimeout(45000);
  const requests = [], errors = [];
  page.on('pageerror', e => errors.push(e.message));
  page.on('request', r => {if (r.method() === 'POST' && new URL(r.url()).pathname.startsWith('/api/')) requests.push({url: r.url(), body: r.postDataJSON()})});
  const response = page.waitForResponse(r => r.request().method() === 'GET' && new URL(r.url()).pathname === '/api/captures/' + ledger.saved.receipt.id);
  await page.goto(base + '/#capture=' + ledger.saved.receipt.id);
  const reopened = await (await response).json();
  expect(reopened).toEqual(ledger.saved);
  const plot = page.locator('#savednumericSummary .numeric-distribution');
  await expect(plot).toBeVisible();
  await plot.evaluate(el => {el.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  const geometry = await plot.locator('.numeric-distribution-scroll').evaluate(el => ({width: el.clientWidth, scrollWidth: el.scrollWidth, left: el.getBoundingClientRect().left, right: el.getBoundingClientRect().right, viewportWidth: innerWidth}));
  expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.width);
  expect(geometry.left).toBeGreaterThanOrEqual(0);
  expect(geometry.right).toBeLessThanOrEqual(geometry.viewportWidth);
  const file = 'docs/journeys/screenshots/181c-' + ledger.signal + '-saved-value-distribution.png';
  await plot.screenshot({path: path.join(repository, file)});
  expect(requests).toEqual([]); expect(errors).toEqual([]);
  const metadata = {version: 1, kind: 'source-free saved evidence screenshot', environment, signal: ledger.signal, captureID: ledger.saved.receipt.id, question: reopened.receipt.question.explore, count: reopened.explore.count, file, geometry, sourceRequests: requests, capturedAt: new Date().toISOString()};
  await fs.writeFile(path.join(directory, 'desktop-screenshot.json'), JSON.stringify(metadata, null, 2) + '\n');
  console.log(JSON.stringify(metadata));
  await context.close();
}

async function resourceRightPan(browser, ledger, directory) {
  const environment = ledger.environment;
  const archived = ledger.resourceResult, question = archived.question;
  const context = await browser.newContext({viewport: {width: 390, height: 1800}});
  const page = await context.newPage();
  page.setDefaultTimeout(15000); page.setDefaultNavigationTimeout(45000);
  const sourceRequests = [], errors = [];
  page.on('pageerror', e => errors.push(e.message));
  page.on('request', request => {
    const endpoint = new URL(request.url()).pathname;
    if (request.method() === 'POST' && endpoint.startsWith('/api/')) sourceRequests.push({endpoint, question: request.postDataJSON()});
  });
  await page.goto(base + '/' + questionLink(question, 'Distribution'));
  await expect(page.locator('#exploreStatus')).toContainText('No evidence loaded');
  expect(sourceRequests).toEqual([]);
  const pending = page.waitForResponse(response => response.request().method() === 'POST' && new URL(response.url()).pathname === '/api/explore');
  await page.locator('#exploreRun').click();
  const response = await pending;
  expect(response.ok(), await response.text()).toBe(true);
  const reread = await response.json();
  expect(reread.question).toEqual(question);
  expect(reread.count).toBe(archived.count);
  expect(reread.numeric).toEqual(archived.numeric);
  expect(reread.numericHistogram).toEqual(archived.numericHistogram);
  expect(reread.numericHistogram.bins).toHaveLength(128);
  const plot = page.locator('#numericSummary .numeric-distribution');
  await expect(plot).toBeVisible();
  const exact = plot.locator('details').filter({has: page.getByText('All exact numeric bins', {exact: true})});
  await exact.locator(':scope>summary').click();
  await expect(exact).toHaveAttribute('open', '');
  const tableRows = await exact.locator('tbody>tr').evaluateAll(rows => rows.map(row => ({bin: Number(row.dataset.bin), cells: [...row.querySelectorAll('td')].map(cell => cell.textContent)})));
  expect(tableRows).toHaveLength(128);
  for (const [index, bin] of reread.numericHistogram.bins.entries()) expect(tableRows[index]).toEqual({bin: index, cells: ['[' + human(bin.from, question.numeric.scale) + ', ' + human(bin.to, question.numeric.scale) + ')', bin.count]});
  await exact.locator(':scope>summary').click();
  await plot.evaluate(element => {element.scrollIntoView({block: 'start'}); window.scrollBy(0, -120)});
  const scroll = plot.locator('.numeric-distribution-scroll');
  const before = await scroll.evaluate(element => ({clientWidth: element.clientWidth, scrollWidth: element.scrollWidth, scrollLeft: element.scrollLeft}));
  expect(before.scrollWidth).toBeGreaterThan(before.clientWidth);
  await scroll.evaluate(element => {element.scrollLeft = element.scrollWidth});
  const after = await scroll.evaluate(element => ({clientWidth: element.clientWidth, scrollWidth: element.scrollWidth, scrollLeft: element.scrollLeft}));
  expect(after.scrollLeft).toBe(after.scrollWidth - after.clientWidth);
  const lastBar = plot.locator('goo-time-bars .time-bar').last();
  await expect(lastBar).toBeInViewport();
  const lastBin = reread.numericHistogram.bins.at(-1);
  expect(lastBin.count).not.toBe('0');
  const rightBound = plot.locator('.time-axis>span').last();
  await expect(rightBound).toContainText(human(lastBin.to, question.numeric.scale) + ' USD (exclusive)');
  const rightGeometry = await rightBound.evaluate(element => {
    const box = element.getBoundingClientRect(), viewport = element.closest('.numeric-distribution-scroll').getBoundingClientRect();
    return {text: element.textContent, left: box.left, right: box.right, viewportLeft: viewport.left, viewportRight: viewport.right};
  });
  expect(rightGeometry.left).toBeGreaterThanOrEqual(rightGeometry.viewportLeft);
  expect(rightGeometry.right).toBeLessThanOrEqual(rightGeometry.viewportRight + 1);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  const file = 'docs/journeys/screenshots/181d-' + ledger.signal + '-resource-distribution-right-pan.png';
  await plot.screenshot({path: path.join(repository, file)});
  expect(sourceRequests).toHaveLength(1);
  expect(sourceRequests[0].endpoint).toBe('/api/explore');
  const requested = structuredClone(sourceRequests[0].question);
  // The editor sends an empty grouping array; the Go wire omits it.
  if (Array.isArray(requested.groupBy) && requested.groupBy.length === 0) delete requested.groupBy;
  expect(requested).toEqual(question);
  expect(errors).toEqual([]);
  // Read actual stored bytes again; the existing capture is custody evidence,
  // not a substitute for the freshly run resource distribution above.
  const {execFileSync} = require('node:child_process');
  const env = {...process.env};
  for (const key of ['DOCKER_HOST', 'DOCKER_CONTEXT', 'DOCKER_TLS', 'DOCKER_TLS_VERIFY', 'DOCKER_CERT_PATH']) delete env[key];
  const captureID = ledger.saved.receipt.id;
  expect(captureID).toMatch(/^[a-f0-9]{64}$/);
  const stored = execFileSync('docker', ['--host=unix:///var/run/docker.sock', 'exec', process.env.GOO_QUERY_CONTAINER || 'grammar-observability-packages-1', 'cat', '/app/artifacts/investigations/exploration/' + captureID + '.json'], {env, maxBuffer: 4 * 1024 * 1024});
  const custody = JSON.parse(await fs.readFile(path.join(directory, 'capture-custody.json'), 'utf8'));
  const captureHash = createHash('sha256').update(stored).digest('hex');
  expect(captureHash).toBe(custody.sha256);
  const metadata = {version: 1, kind: 'Fresh explicit same-scope resource histogram reread; mobile right pan', environment, signal: ledger.signal, exactScope: question, archivedCount: archived.count, reread, sourceRequests, requestNormalization: 'Only an empty groupBy array is omitted for comparison with the Go wire; raw request retained.', exactDisclosure: {count: tableRows.length, rows: tableRows}, before, after, rightGeometry, lastPopulatedBin: lastBin, originalCapture: {id: captureID, bytes: stored.length, sha256: captureHash, unchanged: true}, file, capturedAt: new Date().toISOString()};
  await fs.writeFile(path.join(directory, 'mobile-pan-replay.json'), JSON.stringify(metadata, null, 2) + '\n');
  console.log(JSON.stringify({environment, sourceReadCount: sourceRequests.length, count: reread.count, exactRows: tableRows.length, after, rightGeometry, originalCaptureUnchanged: true, file}));
  await context.close();
}

async function main() {
  const {values} = parseArgs({options: {ledger: {type: 'string'}, mode: {type: 'string'}, help: {type: 'boolean'}}});
  const usage = 'node journeys/181-visual-replay.js --ledger <retained ledger.json> --mode saved-desktop|resource-right-pan|both';
  if (values.help) {console.log(usage); return}
  if (!values.ledger || !['saved-desktop', 'resource-right-pan', 'both'].includes(values.mode)) throw Error(usage);
  const ledgerPath = path.resolve(values.ledger), directory = path.dirname(ledgerPath);
  const ledger = JSON.parse(await fs.readFile(ledgerPath, 'utf8'));
  if (!['logs', 'spans'].includes(ledger.signal) || !/^[a-f0-9]{64}$/.test(ledger.saved?.receipt?.id || '') || !ledger.resourceResult?.question) throw Error('Supply a completed181 main ledger with its actual capture and resource result.');
  await fs.mkdir(path.join(repository, 'docs/journeys/screenshots'), {recursive: true});
  const browser = await chromium.launch({executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox']});
  try {
    if (values.mode === 'saved-desktop' || values.mode === 'both') await savedDesktop(browser, ledger, directory);
    if (values.mode === 'resource-right-pan' || values.mode === 'both') await resourceRightPan(browser, ledger, directory);
  } finally {await browser.close()}
}
main().catch(error => {console.error(error); process.exitCode = 1});
