import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.poolViewPhase || 0), source: Number(document.documentElement.dataset.poolViewSource || 0), plans: Number(document.documentElement.dataset.poolViewPlans || 0), views: Number(document.documentElement.dataset.poolViewViews || 0),
 catalog: [...document.querySelectorAll('button')].some(b => b.textContent === 'Open service inventory / pool-api'), ready: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded' && !!document.querySelector('.pool-connections'),
 options: !!document.querySelector('.pool-connection-row[data-pool=orders][data-instance=worker-a] [data-pool-load]'), loaded: !!document.querySelector('.pool-connection-row[data-pool=orders][data-instance=worker-a] .pool-aligned-view'), terminal: document.documentElement.dataset.poolViewTerminal === 'true' }));
export const bounded = always(() => state.current.source <= 11 && state.current.views <= 1);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.poolViewStep', async (document, window) => {
 const root = document.documentElement, phase = Number(root.dataset.poolViewPhase || 0), row = () => document.querySelector<HTMLElement>('.pool-connection-row[data-pool=orders][data-instance=worker-a]')!;
 if (phase === 0) {
  const env = new URLSearchParams(window.location.hash.slice(1)).get('poolViewEnvironment')!, original = window.fetch.bind(window);
  window.fetch = async (input, init) => { const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href), method = init?.method || (input instanceof Request ? input.method : 'GET');
   if (method === 'POST' && url.pathname === '/api/pool-signals/plan') root.dataset.poolViewPlans = String(Number(root.dataset.poolViewPlans || 0) + 1);
   else if (method === 'POST' && url.pathname === '/api/pool-signals/view') root.dataset.poolViewViews = String(Number(root.dataset.poolViewViews || 0) + 1);
   else if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.poolViewSource = String(Number(root.dataset.poolViewSource || 0) + 1);
   const response = await original(input, init); if (response.ok && url.pathname === '/api/pool-signals/view') { const data = await response.clone().json(); if (data.utilization.state !== 'observed' || data.utilization.numerator !== '3' || data.utilization.denominator !== '10') throw Error('Certified ratio differs from actual aligned reports'); } return response;
  };
  document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click(); const input = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!; input.value = env; input.dispatchEvent(new Event('input', { bubbles: true })); document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
 } else if (phase === 1) [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Open service inventory / pool-api')!.click();
 else if (phase === 2) row().querySelector<HTMLElement>('.pool-signals > summary')!.click();
 else if (phase === 3) { if (root.dataset.poolViewSource !== '9') throw Error('Opening planned extra telemetry'); row().querySelector<HTMLInputElement>('[aria-label="Load acquisition wait"]')!.click(); }
 else if (phase === 4) { if (root.dataset.poolViewSource !== '9') throw Error('Selecting read telemetry'); row().querySelector<HTMLInputElement>('[aria-label="Load acquisition wait"]')!.click(); }
 else if (phase === 5) row().querySelector<HTMLButtonElement>('[data-pool-load]')!.click();
 else if (phase === 6) { const ratio = row().querySelector<HTMLElement>('.pool-utilization')!; if (ratio.title !== '3/10 of reported maximum' || root.dataset.poolViewSource !== '11') throw Error('Exact held ratio or read budget differs'); }
 else if (phase === 7) row().querySelector<HTMLElement>('.pool-signals > summary')!.click();
 else if (phase === 8) row().querySelector<HTMLElement>('.pool-signals > summary')!.click();
 else if (phase === 9) { if (!row().querySelector('.pool-aligned-view') || root.dataset.poolViewSource !== '11' || root.dataset.poolViewViews !== '1' || root.dataset.poolViewPlans !== '1') throw Error('Cached reopen reread or lost evidence'); root.dataset.poolViewTerminal = 'true'; }
 root.dataset.poolViewPhase = String(phase + 1);
});
export const workflow = actions(() => { const s=state.current;if(s.phase>=10)return ['Wait'];if(s.phase===1&&!s.catalog||s.phase===2&&!s.ready||s.phase===3&&!s.options||s.phase===6&&!s.loaded)return ['Wait'];return [step()]; });
