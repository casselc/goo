import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.poolSignalsPhase || 0), source: Number(document.documentElement.dataset.poolSignalsSource || 0), plans: Number(document.documentElement.dataset.poolSignalsPlans || 0),
  catalog: [...document.querySelectorAll('button')].some(b => b.textContent === 'Open service inventory / pool-api'), ready: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded' && !!document.querySelector('.pool-connections'),
  options: !!document.querySelector('.pool-connection-row[data-pool=orders][data-instance=worker-a] [data-pool-signal=capacity]'), run: !!document.querySelector<HTMLElement>('#measureRun')?.getClientRects().length,
  complete: document.documentElement.dataset.poolSignalsCapacity === 'true', terminal: document.documentElement.dataset.poolSignalsTerminal === 'true' }));
export const bounded = always(() => state.current.source <= 10);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.poolSignalsStep', async (document, window) => {
 const root = document.documentElement, phase = Number(root.dataset.poolSignalsPhase || 0);
 const row = () => document.querySelector<HTMLElement>('.pool-connection-row[data-pool=orders][data-instance=worker-a]')!;
 if (phase === 0) {
  const env = new URLSearchParams(window.location.hash.slice(1)).get('poolSignalsEnvironment')!, original = window.fetch.bind(window);
  window.fetch = async (input, init) => { const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href), method = init?.method || (input instanceof Request ? input.method : 'GET');
   if (method === 'POST' && url.pathname === '/api/pool-signals/plan') root.dataset.poolSignalsPlans = String(Number(root.dataset.poolSignalsPlans || 0) + 1);
   else if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.poolSignalsSource = String(Number(root.dataset.poolSignalsSource || 0) + 1);
   const response = await original(input, init);
   if (response.ok && url.pathname === '/api/measurements') { const result = await response.clone().json(); if (result.question.definition.source.metric === 'db.client.connection.max') {
    if (result.question.definition.stockDomain !== 'nonnegative-integer' || result.stocks.series.length !== 1 || result.stocks.series[0].samples.at(-1)?.value !== 0) throw Error('Capacity exact stock or domain lost'); root.dataset.poolSignalsCapacity = 'true';
   }} return response;
  };
  document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click(); const input = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!; input.value = env; input.dispatchEvent(new Event('input', { bubbles: true })); document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
 } else if (phase === 1) [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Open service inventory / pool-api')!.click();
 else if (phase === 2) { root.dataset.poolSignalsHeld = root.dataset.poolSignalsSource; row().querySelector<HTMLElement>('.pool-signals > summary')!.click(); }
 else if (phase === 3) { if (root.dataset.poolSignalsSource !== root.dataset.poolSignalsHeld) throw Error('Planning read telemetry'); row().querySelector<HTMLButtonElement>('[data-pool-signal=capacity]')!.click(); }
 else if (phase === 4) { if (root.dataset.poolSignalsSource !== root.dataset.poolSignalsHeld) throw Error('Preview read telemetry'); document.querySelector<HTMLButtonElement>('#measureRun')!.click(); }
 else if (phase === 5) { if (root.dataset.poolSignalsSource !== '10') throw Error('Unexpected capacity source budget'); }
 else if (phase === 6) [...document.querySelectorAll<HTMLButtonElement>('#serviceReturn button')].find(b => b.textContent === 'Back to service')!.click();
 else if (phase === 7) { if (!row().getClientRects().length || root.dataset.poolSignalsSource !== '10' || root.dataset.poolSignalsPlans !== '1') throw Error('Back failed to retain pool without reread'); root.dataset.poolSignalsTerminal = 'true'; }
 root.dataset.poolSignalsPhase = String(phase + 1);
});
export const workflow = actions(() => { const s = state.current; if (s.phase >= 8) return ['Wait']; if (s.phase === 1 && !s.catalog || s.phase === 2 && !s.ready || s.phase === 3 && !s.options || s.phase === 4 && !s.run || s.phase === 5 && !s.complete) return ['Wait']; return [step()]; });
