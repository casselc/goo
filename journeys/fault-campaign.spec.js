const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const { randomUUID, createHash } = require('node:crypto');
const { startBombadil, verifyWithGo } = require('./support/cross-app');
const { capturedScenario, inspectScenario } = require('./support/fault-evidence');
const goo = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const app = process.env.GOO_SAMPLE_URL || 'http://127.0.0.1:8081';
const scenarios = ['retry', 'reject', 'missing'];

test('28 correlate generated retry, rejection and missing-terminal workloads', async ({ page, request }, testInfo) => {
  test.setTimeout(240000);
  test.skip(process.env.GOO_FAULT_CAMPAIGN !== '1', 'opt-in: npm run test:fault-campaign pauses and restores owned demo traffic');
  const run = randomUUID(), dir = path.resolve('artifacts/bombadil/fault-campaigns', run);
  await fs.mkdir(dir, { recursive: true });
  await fs.mkdir('docs/journeys/screenshots', { recursive: true });
  const planFile = path.join(dir, 'plan.json');
  await verifyWithGo(dir, 'TestQuintGenerateBrowserCampaign', { GOO_CAMPAIGN_PLAN_OUTPUT: planFile });
  const plan = JSON.parse(await fs.readFile(planFile));
  expect(plan.version).toBe(1); expect(plan.checkpoints).toHaveLength(9);
  const observed = { version: 1, run, seed: plan.seed, scenarios, checkpoints: [], requests: [], inspections: [] };
  const persist = () => fs.writeFile(path.join(dir, 'observations.json'), JSON.stringify(observed, null, 2));
  const explorer = await startBombadil({ dir: path.join(dir, 'goo'), url: goo,
    specification: 'verification/bombadil/workspace.ts', seconds: 120 });
  const checked = new Set(); let revisits = 0, done = false;
  try {
    await explorer.ready('goo.navigate');
    for (const checkpoint of plan.checkpoints) {
      const target = checkpoint.target;
      expect(Number.isInteger(target) && target >= 0 && target < 3).toBe(true);
      if (checkpoint.action === 'issue') {
        expect(target).toBe(observed.requests.length);
        expect(explorer.active()).toBe(true);
        const startedAt = Date.now(), appDir = path.join(dir, 'app-' + target);
        const url = new URL(app); url.hash = new URLSearchParams({ scenario: scenarios[target] }).toString();
        const actor = await startBombadil({ dir: appDir, url: url.href,
          specification: 'verification/bombadil/scenario.ts', seconds: 12 });
        try {
          const outcome = await actor.finish(), result = await capturedScenario(appDir, scenarios[target]);
          expect(observed.requests.some(previous => previous.result.id === result.id)).toBe(false);
          observed.requests.push({ result, startedAt, actor: outcome });
        } finally { await actor.stop(); }
      } else if (checkpoint.action === 'inspect' || checkpoint.action === 'revisit') {
        expect(!!observed.requests[target]).toBe(true);
        expect(checked.has(target)).toBe(checkpoint.action === 'revisit');
        const issued = observed.requests[target];
        const inspection = await inspectScenario({ page, request, ...issued,
          frameName: '28-' + issued.result.scenario + '-captured-evidence.png' });
        observed.inspections.push(inspection);
        checked.add(target); if (checkpoint.action === 'revisit') revisits++;
      } else if (checkpoint.action === 'finish') {
        expect(observed.requests).toHaveLength(3); expect(checked.size).toBe(3); expect(revisits).toBe(2);
        done = true;
      } else if (checkpoint.action !== 'idle') throw new Error('Unknown model action: ' + checkpoint.action);
      const actual = { action: checkpoint.action, target, issued: observed.requests.length, checked: checked.size, revisits, done };
      expect(actual).toEqual(checkpoint); observed.checkpoints.push(actual); await persist();
    }
    observed.exploration = await explorer.finish();
    const records = (await fs.readFile(path.join(dir, 'goo/trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
    const steps = records.filter(record => record.action?.Custom?.name === 'goo.navigate').length;
    expect(steps).toBeGreaterThan(0); observed.exploration.steps = steps;
    await persist();
    await verifyWithGo(dir, 'TestQuintBrowserCampaignLiveObservations', { GOO_CAMPAIGN_OBSERVATIONS: path.join(dir, 'observations.json') });
    for (let i = 0; i < observed.inspections.length; i++) {
      const evidence = observed.inspections[i], reads = evidence.reads;
      // Hashes compare the two recorded projections; they are not authenticated
      // provenance or a replacement for the archived raw questions and rows.
      const { summarizeCaptured } = require('./support/fault-evidence');
      const sourceFacts = summarizeCaptured(reads.spans.oracle.rows, reads.logs.oracle.rows);
      const signature = projection => createHash('sha256').update(JSON.stringify(projection)).digest('hex');
      const sourceProjection = ['spans', 'logs'].map(signal => ({ question: reads[signal].question,
        rows: reads[signal].oracle.rows.map(row => [row.TraceId, row.SpanId, row.timestamp_ns]).sort() }));
      const displayProjection = ['spans', 'logs'].map(signal => ({ question: reads[signal].displayedQuestion,
        rows: reads[signal].rendered.map(row => [row.TraceId, row.SpanId, row.timestamp_ns]).sort() }));
      const ledger = { version: 1, subject: evidence.result.id, scenario: evidence.result.scenario,
        source: { ...sourceFacts, signature: signature(sourceProjection) },
        display: { ...evidence.observed, signature: signature(displayProjection) }, steps, passed: true, complete: done };
      const inspectionDir = path.join(dir, 'inspection-' + i); await fs.mkdir(inspectionDir);
      const ledgerFile = path.join(inspectionDir, 'ledger.json'); await fs.writeFile(ledgerFile, JSON.stringify(ledger, null, 2));
      await verifyWithGo(inspectionDir, 'TestQuintFaultCampaignLiveLedger', { GOO_FAULT_LEDGER: ledgerFile });
    }
  } finally {
    await explorer.stop(); await persist();
    await testInfo.attach('fault-campaign', { path: path.join(dir, 'observations.json'), contentType: 'application/json' });
  }
});
