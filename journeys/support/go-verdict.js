// Required test verdict boundary. Shared protocol cases live in
// tests/fixtures/go-test-verdicts.json and also exercise the Python runner.
function validateGoEvents(output, expected) {
  if (!output.endsWith('\n')) throw new Error('Incomplete Go test event stream');
  const ran = new Set(), passed = new Map(), packages = new Set();
  const actions = new Set(['start', 'run', 'pause', 'cont', 'pass', 'bench', 'fail', 'output', 'skip', 'build-output', 'build-fail']);
  for (const line of output.slice(0, -1).split('\n')) {
    let event;
    try { event = JSON.parse(line); } catch { throw new Error('Non-JSON output in Go test event stream: ' + line.slice(0, 200)); }
    if (!event || typeof event !== 'object' || Array.isArray(event) || !actions.has(event.Action)) throw new Error('Malformed Go test event');
    const {Action: action, Test: name, Package: pkg} = event;
    if (action === 'skip' || action === 'fail' || action === 'build-fail') throw new Error('Required verdict refused: ' + action + ' ' + (name || pkg || 'package'));
    if (['run', 'pass'].includes(action) && (typeof pkg !== 'string' || !pkg || (name !== undefined && (typeof name !== 'string' || !name)))) throw new Error('Malformed Go test verdict');
    if (action === 'run' && name) ran.add(pkg + ':' + name);
    if (action === 'pass') {
      if (name) passed.set(name, pkg);
      else packages.add(pkg);
    }
  }
  const missing = expected.filter(name => !passed.has(name) || !ran.has(passed.get(name) + ':' + name) || !packages.has(passed.get(name)));
  if (missing.length) throw new Error('Required verdict missing: ' + missing.join(', '));
}
module.exports = {validateGoEvents};
