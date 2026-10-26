const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { base, seed, write } = require('./200-runtime-support');
const { startBombadil } = require('./support/cross-app');
test('200 bounded exact runtime native return campaign', async ({ request }) => {
  test.skip(process.env.GOO_RUNTIME_BOMBADIL !== '1', 'Explicit ordinary OTLP campaign'); test.setTimeout(100000);
  const facts = await seed(request), dir = await write(facts, 'actor-fixture.json', facts);
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ runtimeEnvironment: facts.environment }), specification: 'verification/bombadil/runtime-observations.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse);
  const states = entries.flatMap(e => e.snapshots.filter(s => s.name === 'state').map(s => s.value));
  const phases = [...new Set(states.map(s => s.phase))].sort((a, b) => a - b);
  expect(phases).toEqual([0, 1, 2, 3, 4, 5, 6, 7]);
  expect(states.at(-1).terminal).toBe(true); expect(states.at(-1).source).toBe(states.at(-1).baseline + 1);
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
});
