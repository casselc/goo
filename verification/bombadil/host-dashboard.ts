import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({
  phase: Number(document.documentElement.dataset.hostPhase || 0), source: Number(document.documentElement.dataset.hostSource || 0),
  plans: Number(document.documentElement.dataset.hostPlans || 0), terminal: document.documentElement.dataset.hostTerminal === 'true',
  hostKey: document.documentElement.dataset.hostKey || '', inventory: !!document.querySelector('button[data-host-key]'),
  ready: document.querySelectorAll('.host-widget').length === 2, nativeComplete: !!document.querySelector('#measureResult')?.children.length && !document.querySelector<HTMLButtonElement>('#measureCapture')?.disabled, nativeReady: !document.querySelector<HTMLButtonElement>('#measureRun')?.disabled
}));
export const bounded = always(() => state.current.source <= 4 && state.current.plans <= 1);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.hostStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.hostPhase || 0);
  if (phase === 0) {
    const environment = new URLSearchParams(window.location.hash.slice(1)).get('hostEnvironment')!;
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname === '/api/host-dashboard/plan') root.dataset.hostPlans = String(Number(root.dataset.hostPlans || 0) + 1);
      else if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.hostSource = String(Number(root.dataset.hostSource || 0) + 1);
      const response = await original(input, init);
      if (response.ok && url.pathname === '/api/host-discovery') {
        const data = await response.clone().json(), host = data.hosts.find((h: any) => h.resource['host.id'] === 'host-a' && h.resource['cloud.account.id'] === 'account-a' && !Object.hasOwn(h.resource, 'service.name'));
        if (!host) throw Error('Raw missing-service host identity absent'); root.dataset.hostKey = host.key;
      }
      return response;
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    document.querySelector<HTMLButtonElement>('#entitiesHosts')!.click();
    const input = document.querySelector<HTMLInputElement>('[aria-label="Host environment"]')!;
    input.value = environment;
    input.dispatchEvent(new Event('input', { bubbles: true }));
    [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Find hosts')!.click();
  } else if (phase === 1) {
    const button = document.querySelector<HTMLButtonElement>('button[data-host-key="' + root.dataset.hostKey + '"]')!;
    button.scrollIntoView({ block: 'center' }); button.click();
  }
  else if (phase === 2) {
    if (root.dataset.hostSource !== '3' || root.dataset.hostPlans !== '1') throw Error('Host defaults read outside the bounded plan');
    document.querySelector<HTMLElement>('.host-widget details > summary')!.click();
  } else if (phase === 3) {
    if (root.dataset.hostSource !== '3') throw Error('Evidence disclosure read source');
    [...document.querySelectorAll<HTMLButtonElement>('.host-widget button')].find(b => b.textContent === 'Inspect measurement')!.click();
  } else if (phase === 4) {
    if (root.dataset.hostSource !== '3') throw Error('Preview read source');
    document.querySelector<HTMLButtonElement>('#measureRun')!.click();
  } else if (phase === 5) {
    if (root.dataset.hostSource !== '4') throw Error('Explicit native Run read budget differs');
    [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Back to host')!.click();
  } else if (phase === 6) {
    if (root.dataset.hostSource !== '4' || document.querySelectorAll('.host-widget').length !== 2) throw Error('Retained host return reread or lost observations');
    root.dataset.hostTerminal = 'true';
  }
  root.dataset.hostPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const s = state.current;
  if (s.phase >= 7 || s.phase === 1 && (!s.inventory || !s.hostKey) || s.phase === 2 && !s.ready || s.phase === 4 && !s.nativeReady || s.phase === 5 && !s.nativeComplete) return ['Wait'];
  return [step()];
});
