import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';

test('display entry imports a closed passive graph without publishing namespaces', async () => {
  const before = new Set(Object.getOwnPropertyNames(globalThis));
  const display = await import('../ui/display.mjs');
  assert.equal(display.UI, await import('../ui/components.mjs'));
  assert.equal(display.ComparisonVisual, await import('../ui/comparison-visual.mjs'));
  assert.equal(display, await import('../ui/display.mjs'));
  assert.deepEqual(Object.keys(display).sort(), ['AlignedHistogram', 'Attributes', 'BucketHeatmap', 'ChartFrame', 'ComparisonSeries', 'ComparisonVisual', 'EvidenceSummary', 'ExceptionDetail', 'FieldPresentation', 'GroupingControl', 'RecordColumns', 'RecordPresentation', 'ServiceIdentity', 'StateObservations', 'TagList', 'TimeBars', 'TimeRange', 'TraceComparisonView', 'TracePageBreakdown', 'TracePageBreakdownView', 'TracePageComparison', 'UI', 'ValueTimeline', 'Waterfall']);
  assert.deepEqual(Object.getOwnPropertyNames(globalThis).filter(k => !before.has(k)), []);

  // The public graph must remain independent of app owners and their bridge.
  const visited = new Set();
  async function visit(url) {
    if (visited.has(url.href)) return;
    visited.add(url.href);
    const source = await readFile(url, 'utf8');
    assert.doesNotMatch(source, /\brequire\s*\(|module\.exports|global\.Goo|\bfetch\s*\(/);
    for (const [, path] of source.matchAll(/\bfrom\s+['"]([^'"]+)['"]/g)) {
      assert.match(path, /^\.\/[a-z-]+\.mjs$/);
      assert.notEqual(path, './app-display.mjs');
      await visit(new URL(path, url));
    }
  }
  await visit(new URL('../ui/display.mjs', import.meta.url));
  // Canonical comparisons share H0 snapshot/view contracts and pure question
  // validators, including checked measurement-interval scope. The graph still
  // excludes application adapters and all IO.
  assert.equal(visited.size, 39);
});

test('feature imports expose the existing exact-data helpers through ESM', async () => {
  const {UI, EvidenceSummary, ValueTimeline, ComparisonVisual} = await import('../ui/display.mjs');
  assert.equal(UI.durationMS('9007199254740993'), '9007199254.741 ms');
  assert.match(EvidenceSummary.unixNs('1700000000000000001').exact, /000000001Z$/);
  assert.equal(ValueTimeline.model({fromNS:'0', toNS:'1', points:[]}).state, 'no_data');
  assert.equal(ComparisonVisual.scale(['0', null]).widths[1], null);
});
