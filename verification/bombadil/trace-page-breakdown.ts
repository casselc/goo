import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document, window }) => {
  const plot = document.querySelector('goo-trace-page-breakdown') as any;
  return { phase: Number(document.documentElement.dataset.tracePagePhase || 0),
    ready: !!document.querySelector('#exploreRun'),
    applied: document.querySelector('#exploreStatus')?.textContent?.includes('5 matching'),
    trace: !!plot, groups: plot?.querySelectorAll('[data-page-group]').length || 0,
    selected: plot?.selectedIndex ?? -1,
    reads: Number(document.documentElement.dataset.tracePageReads || 0),
    overflow: document.documentElement.scrollWidth > window.innerWidth,
    root: plot?.data?.rows?.findIndex((r: any) => r.SpanName === 'GET /checkout') ?? -1,
    selectedSummary: document.querySelector('.trace-page-span-summary')?.textContent || '',
    rows: plot?.data?.rows?.length || 0,
  };
});
export const bounded = always(() => !state.current.overflow && state.current.reads <= 4 && (!state.current.trace || state.current.rows === 5));
export const completed = eventually(() => state.current.phase === 8 && state.current.selectedSummary.includes('30 ms') && state.current.reads === 4).within(40, 'seconds');
const act = registerCustomAction('goo.tracePageStep', async (document, window) => {
  const phase = Number(document.documentElement.dataset.tracePagePhase || 0);
  const set = (n: number) => { document.documentElement.dataset.tracePagePhase = String(n); };
  if (phase === 0) {
    const original = window.fetch;
    window.fetch = async function (...args: any[]) {
      const url = String(args[0] instanceof Request ? args[0].url : args[0]);
      if (url.includes('/api/explore')) document.documentElement.dataset.tracePageReads = String(Number(document.documentElement.dataset.tracePageReads || 0) + 1);
      return original.apply(window, args as any);
    };
    (document.querySelector('#exploreRun') as HTMLButtonElement).click(); set(1);
  } else if (phase === 1) {
    (document.querySelector('button[aria-label^="Expand trace "]') as HTMLButtonElement).click(); set(2);
  } else if (phase === 2) {
    (document.querySelector('goo-trace-page-breakdown details[data-section="groups"]') as HTMLDetailsElement).open = true;
    (document.querySelector('goo-trace-page-breakdown details[data-section="diagnostics"]') as HTMLDetailsElement).open = true; set(3);
  } else if (phase === 3) {
    (document.querySelector('goo-trace-page-breakdown [data-page-group]') as HTMLButtonElement).click(); set(4);
  } else if (phase === 4) {
    if (document.querySelectorAll('.waterfall-row.span-highlighted').length === 0) throw Error('No accepted group highlight');
    if (Number(document.documentElement.dataset.tracePageReads) !== 3) throw Error('Highlight read source');
    (document.querySelector('goo-trace-page-breakdown [data-page-group]') as HTMLButtonElement).click(); set(5);
  } else if (phase === 5) {
    if (document.querySelectorAll('.span-highlighted').length) throw Error('Group did not clear');
    const p = document.querySelector('goo-trace-page-breakdown') as any;
    const root = p.data.rows.findIndex((r: any) => r.SpanName === 'GET /checkout');
    (p.querySelector('[data-span-row="' + root + '"]') as HTMLButtonElement).click(); set(6);
  } else if (phase === 6) {
    const p = document.querySelector('goo-trace-page-breakdown') as any;
    (p.querySelector('details[data-section="diagnostics"]') as HTMLDetailsElement).open = false;
    (p.querySelector('[data-span-row]') as HTMLButtonElement).click(); set(7);
  } else if (phase === 7) {
    if (Number(document.documentElement.dataset.tracePageReads) !== 4) throw Error('Closed diagnostic read source');
    if (!document.querySelector('.trace-page-span-summary')?.textContent?.includes('30 ms')) throw Error('Wrong selected observed residual');
    set(8);
  }
});
export const choices = actions(() => {
  const s = state.current;
  if (!s.ready || s.phase === 8) return ['Wait'];
  if (s.phase === 1 && !s.applied || s.phase === 2 && (!s.trace || s.reads < 3) || s.phase === 6 && !s.selectedSummary.includes('30 ms')) return ['Wait'];
  return [act()];
});
