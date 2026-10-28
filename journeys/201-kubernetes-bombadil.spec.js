const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path');
const { base, seed, write } = require('./201-kubernetes-support');
const { startBombadil } = require('./support/cross-app');
test('201 bounded Kubernetes local facets and retained UID context campaign', async ({ request }) => {
  test.skip(process.env.GOO_KUBERNETES_BOMBADIL !== '1', 'Explicit ordinary OTLP campaign'); test.setTimeout(100000);
  const facts = await seed(request), dir = await write(facts, 'actor-fixture.json', facts);
  const expectedIDs = facts.records.filter(r => r.family === 'logs' && Object.entries(facts.primary).every(([k, v]) => r.resource[k] === v) && !Object.hasOwn(r.resource, 'k8s.cluster.uid') && !Object.hasOwn(r.resource, 'k8s.namespace.name')).map(r => r.id).sort();
  const actor = await startBombadil({ dir: path.join(dir, 'bombadil'), url: base + '/#' + new URLSearchParams({ podEnvironment: facts.environment, podExpectedResource: JSON.stringify(facts.primary), podExpectedLogIDs: expectedIDs.join(',') }), specification: 'verification/bombadil/kubernetes-resources.ts', seconds: 30 });
  try { await actor.finish(); } finally { await actor.stop(); }
  const entries = (await fs.readFile(path.join(dir, 'bombadil/trace.jsonl'), 'utf8')).trim().split('\n').map(JSON.parse), states = entries.flatMap(e => e.snapshots.filter(s => s.name === 'state').map(s => s.value));
  const phases = [...new Set(states.map(s => s.phase))].sort((a, b) => a - b); expect(phases).toEqual([0,1,2,3,4,5,6,7,8]);
  expect(states.at(-1)).toMatchObject({ phase: 8, source: 2, terminal: true, observed: true });
  await fs.writeFile(path.join(dir, 'bombadil/acceptance-summary.json'), JSON.stringify({ phases, snapshots: states.length, expectedIDs, final: states.at(-1) }, null, 2));
});
