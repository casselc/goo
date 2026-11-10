import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFile} from 'node:fs/promises';

test('display entry imports a closed passive graph without publishing namespaces', async () => {
  const before = new Set(Object.getOwnPropertyNames(globalThis));
  const display = await import('../ui/display.mjs');
  assert.equal(display.UI, await import('../ui/components.mjs'));
  assert.equal(display.ComparisonVisual, await import('../ui/comparison-visual.mjs'));
  assert.equal(display, await import('../ui/display.mjs'));
  assert.deepEqual(Object.keys(display).sort(), ['AlignedHistogram', 'Attributes', 'BucketHeatmap', 'CategoricalFacet', 'ChartFrame', 'ComparisonSeries', 'ComparisonVisual', 'EvidenceSummary', 'ExceptionDetail', 'FacetDeclaration', 'FieldPresentation', 'GroupingControl', 'RecordColumns', 'RecordList', 'RecordPresentation', 'RecordSpanWaterfall', 'ServiceIdentity', 'StateObservations', 'TagList', 'TimeBars', 'TimeRange', 'TraceComparisonView', 'TracePageBreakdown', 'TracePageBreakdownView', 'TracePageComparison', 'UI', 'ValueTimeline', 'Waterfall']);
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
  // Native aligned plots add only the shared pure histogram-value formatter.
  // Count bars and canonical temporal charts share the passive key-range gesture helper.
  assert.ok(visited.has(new URL('../ui/interval-range.mjs', import.meta.url).href));
  // Record lists add bounded page and pure declaration validation, never an app owner.
  for (const name of ['categorical-facet', 'categorical-pair', 'facet-declaration', 'categorical-selection', 'record-list', 'record-span-waterfall', 'record-page', 'view-composition', 'view-declaration', 'timeline-declaration', 'captured-count-comparison', 'count-cohort-context'])
    assert.ok(visited.has(new URL('../ui/' + name + '.mjs', import.meta.url).href));
  // Declared facets add only their passive element and pure declaration validator;
  // counted-values acquisition stays outside this display graph.
  // Common paired narrowing shares the existing pure categorical AST builder, never an app owner.
  // Operand role names add one shared pure bounded-label validator, with no owner or IO.
  assert.ok(visited.has(new URL('../ui/comparison-role-labels.mjs', import.meta.url).href));
  // Source refusal presentation adds only its closed diagnostic contract and
  // passive status/details renderer. Acquisition and retry owners stay excluded.
  for(const name of ['source-refusal','source-refusal-view'])assert.ok(visited.has(new URL('../ui/'+name+'.mjs',import.meta.url).href));
  assert.deepEqual(Object.keys(await import('../ui/source-refusal.mjs')).sort(),['SourceReadError','failure','reasons','refusal','responseError']);
  assert.deepEqual(Object.keys(await import('../ui/source-refusal-view.mjs')),['renderRefusal']);
  for(const name of ['queries','filter-values','facet-value-reads','population-workspace','exploration','distributions','numeric-measures','grouped-time-series'])assert.equal(visited.has(new URL('../ui/'+name+'.mjs',import.meta.url).href),false);
  // Exact numeric admission is now a shared pure module, without pulling app owners into display.
  assert.ok(visited.has(new URL('../ui/numeric-predicate.mjs',import.meta.url).href));
  // Grouped timelines add only their pure shape validator to this graph.
  assert.ok(visited.has(new URL('../ui/grouped-timeline-schema.mjs', import.meta.url).href));
  // The shared duration formatter now lives in a pure leaf used by components;
  // it adds one module, never the distribution owner or numeric-control adapter.
  const valuesURL = new URL('../ui/distribution-values.mjs', import.meta.url);
  assert.ok(visited.has(valuesURL.href));
  const values = await import(valuesURL.href);
  assert.equal(display.UI.durationMS, values.durationMS);
  assert.deepEqual(Object.keys(values).sort(), ['compareMetric','durationMS','groupMetric','metricWidth','percent','valueLabel']);
  const summaryURL = new URL('../ui/summary-declaration.mjs', import.meta.url);
  assert.ok(visited.has(summaryURL.href));
  const summary = await import(summaryURL.href);
  assert.deepEqual(summary.availableFields({signal:'logs'}), ['count']);
  assert.equal(visited.size, 60);
});

test('feature imports expose the existing exact-data helpers through ESM', async () => {
  const {UI, EvidenceSummary, ValueTimeline, ComparisonVisual} = await import('../ui/display.mjs');
  assert.equal(UI.durationMS('9007199254740993'), '9007199254.741 ms');
  assert.match(EvidenceSummary.unixNs('1700000000000000001').exact, /000000001Z$/);
  assert.equal(ValueTimeline.model({fromNS:'0', toNS:'1', points:[]}).state, 'no_data');
  assert.equal(ComparisonVisual.scale(['0', null]).widths[1], null);
});
