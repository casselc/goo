const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { base, seed, write } = require('./194-service-projection-support');
const { startBombadil } = require('./support/cross-app');
test('194 bounded rich service dashboard campaign', async ({ request }) => {
  test.skip(process.env.GOO_SERVICE_PROJECTION_BOMBADIL !== '1', 'Explicit ordinary OTLP campaign'); test.setTimeout(100000);
  const facts = await seed(request), dir = await write(facts, 'actor-fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ serviceProjection: facts.environment }), specification: 'verification/bombadil/service-dashboard-projection.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
  const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value)), phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
  expect(phases).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8]); expect(states.at(-1)).toMatchObject({ phase: 8, source: 11, aggregates: 3, terminal: true });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
