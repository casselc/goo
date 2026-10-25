const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises'), path = require('node:path'), crypto = require('node:crypto');
const { startRuntime } = require('./191-generic-runtime');
test('191 source failure remains distinct while frozen evidence reopens offline', async ({ request }) => {
  test.skip(!process.env.GOO_GENERIC_LEDGER, 'Explicit retained successful generic ledger');
  const ledgerPath = path.resolve(process.env.GOO_GENERIC_LEDGER), priorDir = path.dirname(ledgerPath);
  const ledger = JSON.parse(await fs.readFile(ledgerPath, 'utf8')), dir = path.join(priorDir, 'offline');
  const stored = await fs.readFile(path.join(priorDir, 'scope-capture.json'));
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, dir, { clickHouseURL: 'http://127.0.0.1:1', captureStore: path.join(priorDir, 'capture-store') });
  try {
    const capabilitiesResponse = await request.get(runtime.base + '/api/capabilities');
    expect(capabilitiesResponse.status()).toBe(200); const capabilities = await capabilitiesResponse.json();
    expect(capabilities.declaredOperations.state).toBe('unconfigured');
    const sourceResponse = await request.post(runtime.base + '/api/discovery', { data: { environment: ledger.facts.environment, fromMs: ledger.facts.fromMs, toMs: ledger.facts.toMs } });
    expect(sourceResponse.status()).toBe(502); const sourceFailure = await sourceResponse.json();
    expect(sourceFailure.code).not.toBe('capability_unconfigured');
    const optionalResponse = await request.get(runtime.base + '/api/operation-packages');
    expect(optionalResponse.status()).toBe(503); const optional = await optionalResponse.json();
    expect(optional.code).toBe('capability_unconfigured');
    const response = await request.get(runtime.base + '/api/captures/' + ledger.saved.receipt.id);
    expect(response.status()).toBe(200); const reopened = await response.json(); expect(reopened).toEqual(ledger.saved);
    const after = await fs.readFile(path.join(runtime.root, 'captures/exploration', ledger.saved.receipt.id + '.json'));
    expect(after.equals(stored)).toBe(true);
    await fs.writeFile(path.join(dir, 'offline.json'), JSON.stringify({ sourceLedger: ledgerPath, clickHouse: 'http://127.0.0.1:1', capabilities, sourceStatus: sourceResponse.status(), sourceFailure, optionalStatus: optionalResponse.status(), optional, reopened, originalCaptureSHA256: crypto.createHash('sha256').update(stored).digest('hex'), captureBytesUnchanged: true }, null, 2));
    console.log('191 offline:', dir);
  } finally { await runtime.stop(); }
});
