const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { spawn } = require('node:child_process');
const { randomUUID } = require('node:crypto');
const { stripVTControlCharacters } = require('node:util');

const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const app = process.env.GOO_REPORT_URL || 'http://127.0.0.1:8091';
const frames = path.resolve('docs/journeys/screenshots');

const { inspectReport } = require('./support/cross-app');

test('26 correlate a real report with goo while Bombadil explores, optionally checking Quint', async ({ page, request }, testInfo) => {
  const run = randomUUID(), dir = testInfo.outputPath('cross-app');
  await fs.mkdir(dir, { recursive: true });
  await fs.mkdir(frames, { recursive: true });
  const ledger = { version: 1, run, events: [], experiment: 'real-report-with-concurrent-goo-navigation' };
  const record = (action, data) => ledger.events.push({ action, run, observedAt: new Date().toISOString(), ...data });
  const persist = () => fs.writeFile(path.join(dir, 'ledger.json'), JSON.stringify(ledger, null, 2));
  const logs = [];
  // Launch the pinned native binary so cancellation reaches Bombadil's SIGINT
  // handler, rather than terminating npm's synchronous wrapper and orphaning it.
  const nativeName = { 'linux-x64': 'bombadil-linux-x64', 'linux-arm64': 'bombadil-linux-arm64',
    'darwin-arm64': 'bombadil-darwin-arm64' }[process.platform + '-' + process.arch];
  expect(nativeName, 'platform supported by the pinned Bombadil package').toBeTruthy();
  const explorer = spawn(path.resolve('node_modules/@antithesishq/bombadil/binaries', nativeName), [
    'browser', 'test', '--headless', '--no-sandbox', '--time-limit', '30s',
    '--exit-on-violation', '--output-path', path.join(dir, 'bombadil'), goo,
    'verification/bombadil/workspace.ts',
  ], { cwd: path.resolve('.'), stdio: ['ignore', 'pipe', 'pipe'], timeout: 65000, killSignal: 'SIGINT' });
  const explorationStartedAt = Date.now();
  explorer.stdout.on('data', data => logs.push(data.toString()));
  explorer.stderr.on('data', data => logs.push(data.toString()));
  let exited = false;
  const finished = new Promise(resolve => {
    explorer.on('error', error => { exited = true; resolve({ error: error.message }); });
    explorer.on('close', (code, signal) => { exited = true; resolve({ code, signal, endedAt: Date.now() }); });
  });
  try {
    await expect.poll(() => stripVTControlCharacters(logs.join('')).includes('goo.navigate('), { timeout: 20000 }).toBe(true);
    ledger.explorationStartedAt = explorationStartedAt;
    ledger.workloadStartedAt = Date.now();
    expect(exited, 'explorer must still be active when the workload starts').toBe(false);
    await page.goto(app);
    await page.getByRole('button', { name: 'Generate report', exact: true }).click();
    await expect(page.locator('#result')).toContainText('ready');
    const report = JSON.parse(await page.locator('#result').innerText());
    expect(report.outcome).toBe('ready');
    expect(report.id).toMatch(/^[A-Za-z0-9]+$/);
    record('accepted', { actor: 'application', subject: report.id });
    await page.screenshot({ path: path.join(frames, '26a-correlated-real-report.png'), fullPage: true });

    await inspectReport({ page, request, report, ledger, record, frameName: '26b-correlated-goo-trace.png' });

    const outcome = await finished;
    const navigationActions = stripVTControlCharacters(logs.join('')).match(/goo\.navigate\(/g)?.length || 0;
    ledger.exploration = { ...outcome, navigationActions };
    expect(outcome.code, logs.join('').slice(-3000)).toBe(0);
    expect(navigationActions).toBeGreaterThan(0);
    expect(outcome.endedAt).toBeGreaterThan(ledger.workloadStartedAt);
    record('explored', { actor: 'bombadil-summary', steps: navigationActions, passed: outcome.code === 0 });
    await persist();

    if (process.env.GOO_CROSS_APP_QUINT === '1') {
      // Optional expensive model validation over this actual run, not a fixture.
      const verifierEnv = { ...process.env, GOO_CROSS_APP_LEDGER: path.join(dir, 'ledger.json'), NO_COLOR: '1' };
      // quint-go parses the CLI's plain verdict lines. Playwright forces color
      // in its worker environment, which must not leak to this subprocess.
      delete verifierEnv.FORCE_COLOR;
      const verifier = spawn('go', ['test', '-count=1', '-v', '-run', '^TestQuintCrossAppLiveLedger$', '.'], {
        cwd: path.resolve('verification'), env: verifierEnv, timeout: 60000,
        stdio: ['ignore', 'pipe', 'pipe'],
      });
      const output = [];
      verifier.stdout.on('data', data => output.push(data.toString()));
      verifier.stderr.on('data', data => output.push(data.toString()));
      const code = await new Promise((resolve, reject) => { verifier.on('error', reject); verifier.on('exit', resolve); });
      await fs.writeFile(path.join(dir, 'quint-validation.log'), output.join(''));
      expect(code, output.join('')).toBe(0);
    }
  } finally {
    if (!exited) { explorer.kill('SIGINT'); await finished; }
    await persist();
    await fs.writeFile(path.join(dir, 'bombadil.log'), logs.join(''));
    await testInfo.attach('cross-app-ledger', { path: path.join(dir, 'ledger.json'), contentType: 'application/json' });
  }
});
