const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { randomUUID } = require('node:crypto');
const { inspectReport, startBombadil, verifyWithGo, capturedReport } = require('./support/cross-app');
const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const app = process.env.GOO_REPORT_URL || 'http://127.0.0.1:8091';

test('27 replay a Quint-generated workload with Bombadil over both apps', async ({ page, request }, testInfo) => {
  test.skip(process.env.GOO_GENERATED_CAMPAIGN !== '1', 'opt-in: npm run test:generated-campaign requires Go and quint-go');
  const run = randomUUID();
  // Keep the generated schedule and its actor evidence together even when a
  // later Playwright invocation cleans its ordinary test output directory.
  const dir = path.resolve('artifacts/bombadil/campaigns', run);
  await fs.mkdir(dir, { recursive: true });
  await fs.mkdir('docs/journeys/screenshots', { recursive: true });
  const planFile = path.join(dir, 'plan.json');
  await verifyWithGo(dir, 'TestQuintGenerateBrowserCampaign', { GOO_CAMPAIGN_PLAN_OUTPUT: planFile });
  const plan = JSON.parse(await fs.readFile(planFile));
  expect(plan.version).toBe(1);
  expect(plan.checkpoints).toHaveLength(9);
  const observations = { version: 1, run, seed: plan.seed, checkpoints: [], reports: [], inspections: [] };
  const persist = () => fs.writeFile(path.join(dir, 'observations.json'), JSON.stringify(observations, null, 2));
  const explorer = await startBombadil({ dir: path.join(dir, 'goo'), url: goo,
    specification: 'verification/bombadil/workspace.ts', seconds: 60 });
  let revisits = 0, done = false;
  const checked = new Set();
  try {
    await explorer.ready('goo.navigate');
    for (const checkpoint of plan.checkpoints) {
      const target = checkpoint.target;
      expect(Number.isInteger(target) && target >= 0 && target < 3).toBe(true);
      if (checkpoint.action === 'issue') {
        expect(target).toBe(observations.reports.length);
        expect(explorer.active(), 'goo exploration must overlap each workload request').toBe(true);
        const workloadStartedAt = Date.now(), appDir = path.join(dir, 'app-' + target);
        const actor = await startBombadil({ dir: appDir, url: app,
          specification: 'verification/bombadil/report.ts', seconds: 6 });
        try {
          const outcome = await actor.finish();
          const report = await capturedReport(appDir);
          expect(observations.reports.some(previous => previous.report.id === report.id)).toBe(false);
          observations.reports.push({ report, workloadStartedAt, observedAt: new Date().toISOString(), actor: outcome });
          if (target === 0) {
            // Documentary frame shows the actual returned request, retrieved from
            // Bombadil's screenshot; no second request is made for a screenshot.
            const trace = (await fs.readFile(path.join(appDir, 'trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
            const last = trace.filter(record => record.state?.screenshot).at(-1);
            await fs.copyFile(last.state.screenshot, path.join(dir, 'application-frame.jpeg'));
            await fs.copyFile(last.state.screenshot, 'docs/journeys/screenshots/27c-bombadil-application.jpeg');
            await testInfo.attach('bombadil-application-frame', { path: path.join(dir, 'application-frame.jpeg'), contentType: 'image/jpeg' });
          }
        } finally { await actor.stop(); }
      } else if (checkpoint.action === 'inspect' || checkpoint.action === 'revisit') {
        const issued = observations.reports[target];
        expect(issued, 'model target must have a real captured request result').toBeTruthy();
        if (checkpoint.action === 'inspect') expect(checked.has(target)).toBe(false);
        else expect(checked.has(target)).toBe(true);
        const ledger = { version: 1, run: run + '-observation-' + observations.inspections.length, events: [],
          workloadStartedAt: issued.workloadStartedAt };
        const record = (action, data) => ledger.events.push({ action, run: ledger.run, observedAt: new Date().toISOString(), ...data });
        record('accepted', { actor: 'application', subject: issued.report.id, observedAt: issued.observedAt });
        observations.inspections.push(ledger);
        await inspectReport({ page, request, report: issued.report, ledger, record,
          frameName: checkpoint.action === 'revisit' ? '27b-generated-revisit.png' : '27a-generated-inspection.png' });
        checked.add(target);
        if (checkpoint.action === 'revisit') revisits++;
      } else if (checkpoint.action === 'finish') {
        expect(observations.reports).toHaveLength(3);
        expect(checked.size).toBe(3);
        expect(revisits).toBe(2);
        done = true;
      } else if (checkpoint.action !== 'idle') {
        throw new Error('Unknown model action: ' + checkpoint.action);
      }
      // These are effects established by real handlers, not copied expected counts.
      const actual = { action: checkpoint.action, target, issued: observations.reports.length,
        checked: checked.size, revisits, done };
      expect(actual).toEqual(checkpoint);
      observations.checkpoints.push(actual);
      await persist();
    }
    observations.exploration = await explorer.finish();
    await persist();
    await verifyWithGo(dir, 'TestQuintBrowserCampaignLiveObservations', {
      GOO_CAMPAIGN_OBSERVATIONS: path.join(dir, 'observations.json'),
    });
    const records = (await fs.readFile(path.join(dir, 'goo/trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
    const steps = records.filter(record => record.action?.Custom?.name === 'goo.navigate').length;
    expect(steps).toBeGreaterThan(0);
    observations.exploration.steps = steps;
    if (process.env.GOO_CROSS_APP_QUINT === '1') {
      for (let i = 0; i < observations.inspections.length; i++) {
        const ledger = observations.inspections[i];
        ledger.events.push({ action: 'explored', run: ledger.run, actor: 'bombadil-summary', steps, passed: true });
        const receiptDir = path.join(dir, 'inspection-' + i);
        await fs.mkdir(receiptDir);
        const ledgerFile = path.join(receiptDir, 'ledger.json');
        await fs.writeFile(ledgerFile, JSON.stringify(ledger, null, 2));
        await verifyWithGo(receiptDir, 'TestQuintCrossAppLiveLedger', { GOO_CROSS_APP_LEDGER: ledgerFile });
      }
    }
  } finally {
    await explorer.stop();
    await persist();
    await testInfo.attach('generated-campaign-observations', { path: path.join(dir, 'observations.json'), contentType: 'application/json' });
  }
});
