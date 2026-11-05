// Process success and observations admit a campaign for its owning journey's
// completion assertions. They do not establish eventual completion themselves.
function admitBombadil(result, text, {snapshot} = {}) {
  if (snapshot !== undefined && (!snapshot || typeof snapshot !== 'object' || Array.isArray(snapshot) ||
    typeof snapshot.name !== 'string' || !snapshot.name.trim() ||
    (snapshot.version !== undefined && (typeof snapshot.version !== 'string' || !snapshot.version.trim()))))
    throw new Error('Expected a named Bombadil snapshot and optional nonempty model version');
  if (result.code !== 0 || result.signal || result.error) throw new Error('Bombadil process did not exit successfully');
  if (typeof text !== 'string' || !text.trim()) throw new Error('Bombadil emitted no observed trace');
  const lines = text.split('\n');
  if (lines.at(-1) === '') lines.pop();
  const records = lines.map((line, index) => {
    let record;
    try { record = JSON.parse(line); } catch { throw new Error(`Malformed Bombadil trace row ${index + 1}`); }
    if (!record || typeof record !== 'object' || Array.isArray(record) || !Array.isArray(record.snapshots) || !Array.isArray(record.violations))
      throw new Error(`Invalid Bombadil trace row ${index + 1}`);
    if (record.violations.length) throw new Error(`Bombadil recorded violations at row ${index + 1}`);
    for (const value of record.snapshots) {
      if (!value || typeof value !== 'object' || typeof value.name !== 'string' || !Object.hasOwn(value, 'value'))
        throw new Error(`Invalid Bombadil snapshot at row ${index + 1}`);
    }
    return record;
  });
  const observations = records.flatMap(record => record.snapshots);
  const states = observations.filter(value => !snapshot || (value.name === snapshot.name &&
    (snapshot.version === undefined || value.value?.version === snapshot.version))).map(value => value.value);
  if (!states.length) throw new Error('Bombadil emitted no required observed snapshots');
  return {records, states};
}

module.exports = {admitBombadil};
