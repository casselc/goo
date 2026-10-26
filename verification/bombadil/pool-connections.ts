import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.poolPhase || 0), source: Number(document.documentElement.dataset.poolSource || 0),
  catalog: [...document.querySelectorAll('button')].some(b => b.textContent === 'Open service inventory / pool-api'),
  ready: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded' && !!document.querySelector('.pool-connections'),
  terminal: document.documentElement.dataset.poolTerminal === 'true', overflow: document.documentElement.scrollWidth > innerWidth }));
export const bounded = always(() => state.current.source <= 12 && !state.current.overflow);
export const completed = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.poolConnectionsStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.poolPhase || 0);
  if (phase === 0) {
    const env = new URLSearchParams(window.location.hash.slice(1)).get('poolEnvironment')!, original = window.fetch.bind(window);
    window.fetch = async (input, init) => { const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href), method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.poolSource = String(Number(root.dataset.poolSource || 0) + 1); return original(input, init); };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const input = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!; input.value = env; input.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Open service inventory / pool-api')!.click();
  else if (phase === 2) { root.dataset.poolHeldSource = root.dataset.poolSource; const input = document.querySelector<HTMLInputElement>('.pool-connections [aria-label="Find named pool"]')!; input.value = 'orders'; input.dispatchEvent(new Event('input', { bubbles: true })); }
  else if (phase === 3) { if (document.querySelectorAll('.pool-connection-row').length !== 2) throw Error('Pool identity filter merged instances'); const select = document.querySelector<HTMLSelectElement>('[aria-label="Pool connection state"]')!; select.value = 'state:reserved'; select.dispatchEvent(new Event('change', { bubbles: true })); }
  else if (phase === 4) { const value = document.querySelector<HTMLElement>('.pool-connection-row[data-instance="worker-a"] [data-pool-state="reserved"]')!; if (value.dataset.value !== '2') throw Error('Custom reported state lost'); const select = document.querySelector<HTMLSelectElement>('[aria-label="Pool connection state"]')!; select.value = 'all'; select.dispatchEvent(new Event('change', { bubbles: true })); }
  else if (phase === 5) { document.querySelector<HTMLElement>('.pool-connection-row[data-instance="worker-a"] > details > summary')!.click(); }
  else if (phase === 6) { const row = document.querySelector<HTMLElement>('.pool-connection-row[data-instance="worker-a"]')!; if (!row.textContent?.includes('epoch') || root.dataset.poolSource !== root.dataset.poolHeldSource) throw Error('Local pool disclosure read source or lost epoch'); root.dataset.poolTerminal = 'true'; }
  root.dataset.poolPhase = String(phase + 1);
});
export const workflow = actions(() => state.current.phase >= 7 ? ['Wait'] : state.current.phase === 1 && !state.current.catalog || state.current.phase === 2 && !state.current.ready ? ['Wait'] : [step()]);
