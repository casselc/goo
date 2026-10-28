const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { measurementStep, stepLink } = require('../ui/workspace.mjs');
const { startBombadil } = require('./support/cross-app');
test('195 bounded stock control campaign on retained real OTLP facts', async () => {
  test.skip(!process.env.GOO_STOCK_LEDGER, 'Explicit retained stock ledger required'); test.setTimeout(80000);
  const ledgerPath = path.resolve(process.env.GOO_STOCK_LEDGER), ledger = JSON.parse(await fs.readFile(ledgerPath, 'utf8')), dir = path.join(path.dirname(ledgerPath), 'bombadil');
  const actor = await startBombadil({ dir, url: 'http://127.0.0.1:8088/' + stepLink(measurementStep(ledger.question)), specification: 'verification/bombadil/cumulative-stocks.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
  const states = entries.flatMap(e => e.snapshots.filter(s => s.name === 'state').map(s => s.value));
  const phases = [...new Set(states.map(s => s.phase))].sort((a, b) => a - b);
  expect(phases).toEqual([0, 1, 2, 3, 4, 5]); expect(states.at(-1)).toMatchObject({ phase: 5, reads: 1, terminal: true });
  await fs.writeFile(path.join(dir, 'acceptance-summary.json'), JSON.stringify({ sourceLedger: ledgerPath, newIngestion: false, phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
