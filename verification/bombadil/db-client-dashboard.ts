import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const state = extract(({ document }) => {
  const root = document.documentElement;
  const dashboard = document.querySelector<HTMLElement>('.service-dashboard');
  const evidence = [...document.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => {
    try { return JSON.parse(pre.textContent || 'null'); } catch { return null; }
  }).filter(Boolean);
  const spans = evidence.find(item => item.widget.key === 'spans');
  return {
    phase: Number(root.dataset.dbDashboardPhase || 0),
    proof: root.dataset.dbDashboardProof,
    source: Number(root.dataset.dbDashboardSource || 0),
    plans: Number(root.dataset.dbDashboardPlans || 0),
    loaded: dashboard?.dataset.state === 'loaded',
    profile: dashboard?.dataset.profile,
    count: spans?.count,
    service: document.querySelector('#serviceDetail')?.textContent || '',
    catalog: [...document.querySelectorAll<HTMLButtonElement>('#serviceCatalog button')].some(button => button.textContent === 'Open service commerce / db-api'),
    terminal: root.dataset.dbDashboardTerminal === 'true',
    coherent: evidence.every(item => (item.question.environment || item.question.definition?.environment) === root.dataset.dbDashboardEnvironment && item.question.fromMs < item.question.toMs),
    overflow: document.documentElement.scrollWidth > innerWidth,
  };
});
export const scoped = always(() => state.current.coherent && !state.current.overflow);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');

const step = registerCustomAction('goo.dbDashboardStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.dbDashboardPhase || 0);
  const button = (label: string) => [...document.querySelectorAll<HTMLButtonElement>('button')].find(value => value.textContent === label && value.getClientRects().length && !value.disabled);
  const evidence = () => [...document.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => JSON.parse(pre.textContent || 'null')).find(item => item.widget.key === 'spans');
  const check = (expected: string) => { if (evidence()?.count !== expected) throw Error('DB client dashboard population differs: expected ' + expected); };
  if (phase === 0) {
    root.dataset.dbDashboardEnvironment = new URLSearchParams(window.location.hash.slice(1)).get('dbDashboard')!;
    root.dataset.dbDashboardSource = '0'; root.dataset.dbDashboardPlans = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (url.pathname.startsWith('/api/') && method === 'POST') {
        const key = url.pathname === '/api/dashboards/plan' ? 'dbDashboardPlans' : 'dbDashboardSource';
        root.dataset[key] = String(Number(root.dataset[key] || 0) + 1);
      }
      return original(input, init);
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const environment = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!;
    environment.value = root.dataset.dbDashboardEnvironment;
    environment.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) button('Open service commerce / db-api')!.click();
  else if (phase === 2) {
    check('137'); if (Number(root.dataset.dbDashboardSource) !== 9) throw Error('Automatic overview source budget changed'); root.dataset.dbDashboardProof = 'db:137';
    root.dataset.dbDashboardBeforeLocal = root.dataset.dbDashboardSource;
    const details = document.querySelector<HTMLDetailsElement>('[data-widget=spans] > .dashboard-widget-evidence')!;
    details.open = true;
  } else if (phase === 3) {
    if (root.dataset.dbDashboardSource !== root.dataset.dbDashboardBeforeLocal) throw Error('Evidence disclosure read telemetry');
    document.querySelector<HTMLDetailsElement>('[data-widget=spans] > .dashboard-widget-evidence')!.open = false;
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=all-spans]')!.click();
  } else if (phase === 4) {
    check('189'); if (Number(root.dataset.dbDashboardSource) !== 13) throw Error('All-spans switch performed extra source reads'); root.dataset.dbDashboardProof += '|all:189';
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=db-client]')!.click();
  } else if (phase === 5) {
    check('137'); if (Number(root.dataset.dbDashboardSource) !== 17) throw Error('DB client switch performed extra source reads'); root.dataset.dbDashboardProof += '|db:137';
    root.dataset.dbDashboardBeforeLocal = root.dataset.dbDashboardSource;
    const picker = document.querySelector<HTMLDetailsElement>('.dashboard-native-picker')!;
    picker.open = true;
  } else if (phase === 6) {
    const selected = [...document.querySelectorAll<HTMLInputElement>('.dashboard-native-picker input:checked')].map(input => input.getAttribute('aria-label'));
    if (!selected.some(value => value?.startsWith('db.client.operation.duration · histogram [s]'))) throw Error('DB client native duration not selected by default');
    if (root.dataset.dbDashboardSource !== root.dataset.dbDashboardBeforeLocal) throw Error('Native picker disclosure read telemetry');
    document.querySelector<HTMLDetailsElement>('.dashboard-native-picker')!.open = false;
    button('Back to service list')!.click();
  } else if (phase === 7) {
    if (root.dataset.dbDashboardSource !== root.dataset.dbDashboardBeforeLocal) throw Error('Return to retained catalog read telemetry');
    if (document.querySelector<HTMLElement>('#serviceDetail') && !document.querySelector<HTMLElement>('#serviceDetail')!.hidden) throw Error('Service inspector still open after return');
    if (Number(root.dataset.dbDashboardPlans) !== 3) throw Error('Expected exactly three explicit profile plans');
    root.dataset.dbDashboardTerminal = 'true';
  }
  root.dataset.dbDashboardPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.phase >= 8) return ['Wait'];
  if (value.phase === 1 && !value.catalog) return ['Wait'];
  if ([2, 4, 5].includes(value.phase) && !value.loaded) return ['Wait'];
  if (value.phase === 4 && value.profile !== 'all-spans' || value.phase === 5 && value.profile !== 'db-client') return ['Wait'];
  return [step()];
});
