import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({
  phase: Number(document.documentElement.dataset.serviceProjectionPhase || 0),
  source: Number(document.documentElement.dataset.serviceProjectionSource || 0), aggregates: Number(document.documentElement.dataset.serviceProjectionAggregates || 0),
  loaded: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded',
  visible: !!document.querySelector<HTMLElement>('.service-dashboard')?.getClientRects().length,
  catalog: [...document.querySelectorAll('button')].some(button => button.textContent === 'Open service commerce / edge-api'),
  records: document.querySelector('#exploreCount > span')?.textContent === '155',
  terminal: document.documentElement.dataset.serviceProjectionTerminal === 'true', overflow: document.documentElement.scrollWidth > innerWidth,
}));
export const bounded = always(() => state.current.source <= 11 && !state.current.overflow);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.serviceProjectionStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.serviceProjectionPhase || 0);
  const button = (label: string) => [...document.querySelectorAll<HTMLButtonElement>('button')].find(button => button.textContent === label && button.getClientRects().length && !button.disabled);
  const evidence = () => [...document.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => JSON.parse(pre.textContent || 'null'));
  if (phase === 0) {
    const environment = new URLSearchParams(window.location.hash.slice(1)).get('serviceProjection')!;
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href), method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.serviceProjectionSource = String(Number(root.dataset.serviceProjectionSource || 0) + 1);
      if (url.pathname === '/api/record-aggregates') root.dataset.serviceProjectionAggregates = String(Number(root.dataset.serviceProjectionAggregates || 0) + 1);
      const response = await original(input, init);
      if (response.ok && url.pathname === '/api/explore') { const data = await response.clone().json(); root.dataset.serviceProjectionLastFull = JSON.stringify({ count: data.count, rows: data.rows?.length, fields: data.fields?.length }); }
      return response;
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const input = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!; input.value = environment; input.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) button('Open service commerce / edge-api')!.click();
  else if (phase === 2) {
    const entries = evidence();
    for (const [key, count] of [['spans', '137'], ['logs', '155'], ['dimension', '137']]) {
      const entry = entries.find(entry => entry.widget.key === key);
      if (entry?.version !== 'goo/record-aggregates/v1' || entry.count !== count) throw Error('Rich widget contract differs for ' + key);
    }
    if (root.dataset.serviceProjectionAggregates !== '3') throw Error('Three supported record widgets must use the narrow endpoint');
    const supporting = JSON.parse(root.dataset.serviceProjectionLastFull || 'null'); if (supporting.rows !== 100 || !supporting.fields) throw Error('Admission lost its full supporting record page');
  } else if (phase === 3) {
    const sort = document.querySelector<HTMLSelectElement>('[data-widget=dimension] select[aria-label="Rank diagnosis rows"]')!; sort.value = 'p95'; sort.dispatchEvent(new Event('change', { bubbles: true }));
  } else if (phase === 4) { root.dataset.serviceProjectionBeforeStage = root.dataset.serviceProjectionSource; document.querySelector<HTMLAnchorElement>('[data-widget=logs] .ui-actions > a')!.click(); }
  else if (phase === 5) {
    if (root.dataset.serviceProjectionSource !== root.dataset.serviceProjectionBeforeStage) throw Error('Opening the original widget question read telemetry');
    document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  } else if (phase === 6) {
    const full = JSON.parse(root.dataset.serviceProjectionLastFull || 'null'); if (full.count !== '155' || full.rows !== 100 || !full.fields) throw Error('Explicit log drilldown lost its full result');
  } else if (phase === 7) {
    const dashboard = document.querySelector<HTMLElement>('.service-dashboard');
    if (!dashboard?.getClientRects().length) {
      const attempts = Number(root.dataset.serviceProjectionBack || 0); if (attempts >= 7) throw Error('Retained service did not return'); root.dataset.serviceProjectionBack = String(attempts + 1); window.history.back(); return;
    }
    if (root.dataset.serviceProjectionSource !== '11' || root.dataset.serviceProjectionAggregates !== '3') throw Error('Retained Back reread the overview');
    root.dataset.serviceProjectionTerminal = 'true';
  }
  root.dataset.serviceProjectionPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.phase >= 8) return ['Wait'];
  if (value.phase === 1 && !value.catalog || value.phase === 2 && !value.loaded || value.phase === 6 && !value.records) return ['Wait'];
  return [step()];
});
