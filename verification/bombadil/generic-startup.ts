import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({
  phase: Number(document.documentElement.dataset.genericPhase || 0),
  profile: document.documentElement.dataset.profile,
  source: Number(document.documentElement.dataset.genericSource || 0),
  plans: Number(document.documentElement.dataset.genericPlans || 0),
  terminal: document.documentElement.dataset.genericTerminal === 'true',
  proof: document.documentElement.dataset.genericProof || '',
  catalog: !!document.querySelector('#serviceCatalog button'),
  loaded: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded',
  fixed: document.querySelector('#exploreCount')?.textContent?.includes('140') || false,
  overflow: document.documentElement.scrollWidth > innerWidth,
}));
export const coherent = always(() => !state.current.overflow && (!state.current.profile || state.current.profile === 'generic'));
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.genericStartupStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.genericPhase || 0);
  const button = (text: string) => [...document.querySelectorAll<HTMLButtonElement>('button')].find(value => value.textContent === text && value.getClientRects().length && !value.disabled);
  if (phase === 0) {
    if (root.dataset.profile !== 'generic') throw Error('Generic profile was not loaded');
    root.dataset.genericSource = '0'; root.dataset.genericPlans = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (url.pathname.startsWith('/api/') && method === 'POST') {
        const key = url.pathname === '/api/dashboards/plan' ? 'genericPlans' : 'genericSource';
        root.dataset[key] = String(Number(root.dataset[key] || 0) + 1);
      }
      return original(input, init);
    };
    const selectedEnvironment = new URLSearchParams(window.location.hash.slice(1)).get('genericEnvironment')!;
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const environment = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!;
    environment.value = selectedEnvironment;
    environment.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) button('Open service ordinary-python / python-orders')!.click();
  else if (phase === 2) {
    const entry = [...document.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => JSON.parse(pre.textContent || 'null')).find(item => item.widget.key === 'spans');
    if (entry.count !== '140') throw Error('Python service population differs');
    root.dataset.genericProof = 'service:140';
    root.dataset.genericBeforeStage = root.dataset.genericSource;
    document.querySelector<HTMLAnchorElement>('[data-widget=spans] a')!.click();
  } else if (phase === 3) {
    if (root.dataset.genericSource !== root.dataset.genericBeforeStage) throw Error('Staging read telemetry');
    document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  }
  else if (phase === 4) {
    if (!document.querySelector('#exploreCount')?.textContent?.includes('140')) throw Error('Fixed population differs');
    if (Number(root.dataset.genericSource) !== Number(root.dataset.genericBeforeStage) + 1) throw Error('Fixed Run source budget differs');
    root.dataset.genericProof += '|fixed:140';
    root.dataset.genericTerminal = 'true';
  }
  root.dataset.genericPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.profile !== 'generic') return ['Wait'];
  if (value.phase >= 5) return ['Wait'];
  if (value.phase === 1 && !value.catalog) return ['Wait'];
  if (value.phase === 2 && !value.loaded) return ['Wait'];
  if (value.phase === 4 && !value.fixed) return ['Wait'];
  return [step()];
});
