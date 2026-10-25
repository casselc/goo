const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { fixture, question, logs } = require('./193-counted-values-fixture');
const { questionLink } = require('../ui/workspace');
const { startBombadil } = require('./support/cross-app');
test('193 bounded counted-values projection campaign', async ({ request }) => {
  test.skip(process.env.GOO_COUNTED_PROJECTION_BOMBADIL !== '1', 'Explicit ordinary OTLP campaign'); test.setTimeout(100000);
  const facts = fixture(), base = 'http://127.0.0.1:8088', dir = path.resolve('artifacts/counted-values-projection', facts.environment);
  expect((await request.post('http://127.0.0.1:4318/v1/logs', { data: logs(facts) })).ok()).toBe(true);
  await expect.poll(async () => { const response = await request.post(base + '/api/explore', { data: question(facts) }); expect(response.ok()).toBe(true); return (await response.json()).count; }, { timeout: 60000 }).toBe('143');
  await fs.mkdir(dir, { recursive: true }); await fs.writeFile(path.join(dir, 'fixture.json'), JSON.stringify(facts, null, 2));
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/' + questionLink(question(facts)), specification: 'verification/bombadil/counted-values-projection.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value)), phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
  expect(phases).toEqual([0, 1, 2, 3, 4, 5, 6]); expect(states.at(-1)).toMatchObject({ phase: 6, source: 3, lookups: 1, legacy: 0, terminal: true });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2)); console.log('193 actor:', dir);
});
