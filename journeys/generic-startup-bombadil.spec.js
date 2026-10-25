const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { startRuntime } = require('./191-generic-runtime');
const { fixture, traces } = require('./191-generic-fixture');
const { startBombadil } = require('./support/cross-app');
test('191 bounded generic startup campaign', async ({ request }) => {
  test.skip(process.env.GOO_GENERIC_BOMBADIL !== '1', 'Explicit generic startup actor');
  test.setTimeout(100000);
  const facts = fixture(), dir = path.resolve('artifacts/generic-startup', facts.environment);
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, dir);
  try {
    const ack = await request.post('http://127.0.0.1:4318/v1/traces', { data: traces(facts) }); expect(ack.ok()).toBe(true);
    await expect.poll(async () => {
      const response = await request.post(runtime.base + '/api/explore', { data: { signal: 'spans', environment: facts.environment, fromMs: facts.fromMs, toMs: facts.toMs, limit: 100 } });
      expect(response.ok()).toBe(true); return (await response.json()).count;
    }, { timeout: 60000 }).toBe('150');
    const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: runtime.base + '/#' + new URLSearchParams({ genericEnvironment: facts.environment }), specification: 'verification/bombadil/generic-startup.ts', seconds: 30 });
    try { await actor.finish(); } finally { await actor.stop(); }
    const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(line => JSON.parse(line));
    const states = entries.flatMap(entry => entry.snapshots.filter(value => value.name === 'state').map(value => value.value));
    const phases = [...new Set(states.map(value => value.phase))].sort((a, b) => a - b);
    expect(phases).toEqual([0, 1, 2, 3, 4, 5]);
    expect(states.at(-1)).toMatchObject({ phase: 5, terminal: true, profile: 'generic', proof: 'service:140|fixed:140' });
    await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, final: states.at(-1) }, null, 2));
    console.log('191 actor:', dir);
  } finally { await runtime.stop(); }
});
