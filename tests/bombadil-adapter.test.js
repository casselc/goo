const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const os = require('node:os');
const path = require('node:path');
const { capturedReport } = require('../journeys/support/cross-app');

// Protocol fixtures for the artifact reader, not mocked UI/integration evidence.
const submit = { action: { Custom: { name: 'sample.generate-report' } } };
const result = id => ({ snapshots: [{ name: 'reportResult', value: { id, outcome: 'ready' } }] });
async function read(records) {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), 'goo-bombadil-artifact-'));
  try {
    await fs.writeFile(path.join(dir, 'trace.jsonl'), records.map(record => JSON.stringify(record)).join('\n'));
    return await capturedReport(dir);
  } finally { await fs.rm(dir, { recursive: true, force: true }); }
}
test('captured identity is stable across repeated snapshots', async () => {
  assert.deepEqual(await read([submit, result('actual1'), result('actual1')]), { id: 'actual1', outcome: 'ready' });
});
test('a passing run cannot omit submission or the actual returned identity', async () => {
  await assert.rejects(read([result('actual1')]));
  await assert.rejects(read([submit]));
});
test('multiple submissions or identities are rejected even when all outcomes are ready', async () => {
  await assert.rejects(read([submit, submit, result('actual1')]));
  await assert.rejects(read([submit, result('actual1'), result('other2')]));
});
