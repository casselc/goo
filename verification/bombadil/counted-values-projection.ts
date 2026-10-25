import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({
  phase: Number(document.documentElement.dataset.countedProjectionPhase || 0),
  source: Number(document.documentElement.dataset.countedProjectionReads || 0),
  lookups: Number(document.documentElement.dataset.countedProjectionLookups || 0),
  legacy: Number(document.documentElement.dataset.countedProjectionLegacy || 0),
  applied: document.querySelector('#exploration')?.getAttribute('data-question-phase') === 'applied',
  loaded: document.querySelector('#exploreFilterEditor .filter-counted-values')?.textContent?.includes('143 matching records') || false,
  selected: document.querySelector('#exploreCount > span')?.textContent === '1',
  terminal: document.documentElement.dataset.countedProjectionTerminal === 'true',
  overflow: document.documentElement.scrollWidth > innerWidth,
}));
export const bounded = always(() => state.current.source <= 3 && state.current.legacy === 0 && !state.current.overflow);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.countedProjectionStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.countedProjectionPhase || 0);
  if (phase === 0) {
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (url.pathname.startsWith('/api/') && method === 'POST') {
        root.dataset.countedProjectionReads = String(Number(root.dataset.countedProjectionReads || 0) + 1);
        const body = typeof init?.body === 'string' ? JSON.parse(init.body) : null;
        if (url.pathname === '/api/explore/counted-values') root.dataset.countedProjectionLookups = String(Number(root.dataset.countedProjectionLookups || 0) + 1);
        if (url.pathname === '/api/explore' && body?.limit === 1) root.dataset.countedProjectionLegacy = '1';
        if (url.pathname === '/api/explore') root.dataset.countedProjectionLast = JSON.stringify(body);
      }
      return original(input, init);
    };
    document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  } else if (phase === 1) {
    document.querySelector<HTMLDetailsElement>('#filterComposer')!.open = true;
    const panel = document.querySelector<HTMLDetailsElement>('#exploreFilterEditor .filter-observations')!; panel.open = true;
    const search = panel.querySelector<HTMLInputElement>('input')!; search.value = 'attribute:release'; search.dispatchEvent(new Event('input', { bubbles: true }));
    panel.querySelectorAll<HTMLDetailsElement>(':scope>div>details').forEach(element => { element.open = true; });
    document.querySelector<HTMLButtonElement>('#exploreFilterEditor button[aria-label="Load scoped values for attribute:release"]')!.click();
  } else if (phase === 2) {
    if (root.dataset.countedProjectionReads !== '2' || root.dataset.countedProjectionLookups !== '1') throw Error('Counted-values endpoint budget differs');
    const search = document.querySelector<HTMLInputElement>('#exploreFilterEditor input[aria-label="Search counted values"]')!;
    search.value = 'canary'; search.dispatchEvent(new Event('input', { bubbles: true }));
  } else if (phase === 3) document.querySelector<HTMLButtonElement>('#exploreFilterEditor button[aria-label=\'Stage counted attribute:release "canary"\']')!.click();
  else if (phase === 4) {
    if (root.dataset.countedProjectionReads !== '2') throw Error('Staging counted value read telemetry');
    document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  } else if (phase === 5) {
    const last = JSON.parse(root.dataset.countedProjectionLast || 'null');
    if (!JSON.stringify(last?.where).includes('9007199254740993') || !JSON.stringify(last?.where).includes('canary')) throw Error('Staged counted value lost original numeric selector');
    if (root.dataset.countedProjectionReads !== '3') throw Error('Explicit selected Run budget differs');
    root.dataset.countedProjectionTerminal = 'true';
  }
  root.dataset.countedProjectionPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.phase >= 6) return ['Wait'];
  if (value.phase === 1 && !value.applied || value.phase === 2 && !value.loaded || value.phase === 5 && !value.selected) return ['Wait'];
  return [step()];
});
