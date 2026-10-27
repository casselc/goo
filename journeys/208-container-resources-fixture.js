// Generated ordinary OTLP; not a running receiver or Kubernetes agent.
const memory = require('./206-container-memory-fixture');
const cpu = require('./207-container-cpu-fixture');
const { randomBytes } = require('node:crypto');
function fixture() {
  const m = memory.fixture(), c = cpu.fixture();
  const environment = 'container-resources-' + randomBytes(5).toString('hex');
  const toMs = Date.now();
  const keep = new Set(['aligned', 'replacement-id', 'wrong-cluster', 'wrong-namespace']);
  const points = [];
  for (const [family, facts] of [['memory', m], ['cpu', c]]) {
    for (const point of facts.points) {
      if (!point.case.startsWith('primary-') && !keep.has(point.case)) continue;
      points.push({ ...point, familyName: family,
        resource: { ...point.resource, 'deployment.environment.name': environment },
        endNS: String(BigInt(toMs - 3000) * 1000000n) });
    }
  }
  const memoryAlternative = points.find(p => p.familyName === 'memory' && p.role === 'workingSet' && p.case.startsWith('primary-'));
  points.push({ ...memoryAlternative, value: '49152', resource: { ...memoryAlternative.resource, 'reported.memory.variant': 'alternative' }, case: 'memory-ambiguity' });
  return { environment, fromMs: toMs - 120000, toMs,
    primary: { ...m.primary, 'deployment.environment.name': environment },
    schemaURL: memory.clusterSchema, points, records: [],
    names: { memory: memory.names, cpu: cpu.names },
    scopes: { memory: memory.scopes, cpu: cpu.scopes } };
}
module.exports = { ...memory, fixture };
