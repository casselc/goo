// Read-only source queries and local inspection. No saves, workload or export actions.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const inspection = extract(({ document }) => ({
  ready: !!document.querySelector('#exploreRun'),
  busy: document.querySelector<HTMLButtonElement>('#exploreRun')?.disabled,
  records: !!document.querySelector('#exploreRows .record-attributes'),
  viewer: !!document.querySelector('#exploreRows .record-attributes[open] .attribute-viewer'),
  trace: !document.querySelector<HTMLElement>('#exploreContext')?.hidden,
  selected: document.querySelectorAll('#exploreTrace .waterfall-row[aria-pressed="true"]').length,
  counts: [...document.querySelectorAll('.attribute-count')].map(e => e.textContent || ''),
}));
export const becomesReady = eventually(() => inspection.current.ready).within(10, 'seconds');
export const singleSpanSelection = always(() => inspection.current.selected <= 1);
export const attributeCountsAreExplicit = always(() => inspection.current.counts.every(s => /^\d+ of \d+ fields$/.test(s)));

const act = registerCustomAction('goo.inspect', async (document, _window, action: string) => {
  const click = (selector: string) => document.querySelector<HTMLElement>(selector)?.click();
  if (action === 'query') click('#exploreRun');
  if (action === 'open') {
    const detail = document.querySelector<HTMLDetailsElement>('#exploreRows .record-attributes');
    if (detail) detail.open = true;
  }
  if (action === 'trace') click('#exploreRows .record-row button');
  if (action === 'select') click('#exploreTrace .waterfall-row');
  if (action === 'close') click('#exploreContextClose');
  if (action === 'scope') {
    const select = document.querySelector<HTMLSelectElement>('#exploreRows .attribute-viewer select');
    if (select) { select.selectedIndex = (select.selectedIndex + 1) % select.options.length; select.dispatchEvent(new Event('change')); }
  }
  if (action.startsWith('search:')) {
    const input = document.querySelector<HTMLInputElement>('#exploreRows .attribute-viewer input[type="search"]');
    if (input) { input.value = action.slice(7); input.dispatchEvent(new Event('input')); }
  }
});
export const inspect = actions(() => {
  const s = inspection.current;
  if (!s.ready || s.busy) return ['Wait'];
  const choices = [act('query')];
  if (s.records) choices.push(act('open'), act('trace'));
  if (s.viewer) choices.push(act('scope'), ...['service', 'missing-key', '<tag>', ''].map(q => act('search:' + q)));
  if (s.trace) choices.push(act('select'), act('close'));
  return choices;
});
