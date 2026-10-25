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
    phase: Number(root.dataset.messagingDashboardPhase || 0),
    proof: root.dataset.messagingDashboardProof,
    source: Number(root.dataset.messagingDashboardSource || 0),
    plans: Number(root.dataset.messagingDashboardPlans || 0),
    loaded: dashboard?.dataset.state === 'loaded',
    profile: dashboard?.dataset.profile,
    count: spans?.count,
    service: document.querySelector('#serviceDetail')?.textContent || '',
    catalog: [...document.querySelectorAll<HTMLButtonElement>('#serviceCatalog button')].some(button => button.textContent === 'Open service commerce / messaging-app'),
    terminal: root.dataset.messagingDashboardTerminal === 'true',
    coherent: evidence.every(item => (item.question.environment || item.question.definition?.environment) === root.dataset.messagingDashboardEnvironment && item.question.fromMs < item.question.toMs),
    overflow: document.documentElement.scrollWidth > innerWidth,
  };
});
export const scoped = always(() => state.current.coherent && !state.current.overflow);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');

const step = registerCustomAction('goo.messagingDashboardStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.messagingDashboardPhase || 0);
  const button = (label: string) => [...document.querySelectorAll<HTMLButtonElement>('button')].find(value => value.textContent === label && value.getClientRects().length && !value.disabled);
  const evidence = () => [...document.querySelectorAll('.dashboard-widget-evidence > details > pre')].map(pre => JSON.parse(pre.textContent || 'null')).find(item => item.widget.key === 'spans');
  const check = (expected: string) => { if (evidence()?.count !== expected) throw Error('Messaging consumer dashboard population differs: expected ' + expected); };
  if (phase === 0) {
    root.dataset.messagingDashboardEnvironment = new URLSearchParams(window.location.hash.slice(1)).get('messagingDashboard')!;
    root.dataset.messagingDashboardSource = '0'; root.dataset.messagingDashboardPlans = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (url.pathname.startsWith('/api/') && method === 'POST') {
        const key = url.pathname === '/api/dashboards/plan' ? 'messagingDashboardPlans' : 'messagingDashboardSource';
        root.dataset[key] = String(Number(root.dataset[key] || 0) + 1);
      }
      return original(input, init);
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const environment = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!;
    environment.value = root.dataset.messagingDashboardEnvironment;
    environment.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) button('Open service commerce / messaging-app')!.click();
  else if (phase === 2) {
    check('137'); if (Number(root.dataset.messagingDashboardSource) !== 9) throw Error('Automatic overview source budget changed'); root.dataset.messagingDashboardProof = 'consumer:137';
    root.dataset.messagingDashboardBeforeLocal = root.dataset.messagingDashboardSource;
    const details = document.querySelector<HTMLDetailsElement>('[data-widget=spans] > .dashboard-widget-evidence')!;
    details.open = true;
  } else if (phase === 3) {
    if (root.dataset.messagingDashboardSource !== root.dataset.messagingDashboardBeforeLocal) throw Error('Evidence disclosure read telemetry');
    document.querySelector<HTMLDetailsElement>('[data-widget=spans] > .dashboard-widget-evidence')!.open = false;
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=all-spans]')!.click();
  } else if (phase === 4) {
    check('309'); if (Number(root.dataset.messagingDashboardSource) !== 13) throw Error('All-spans switch performed extra source reads'); root.dataset.messagingDashboardProof += '|all:309';
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=messaging-consumer]')!.click();
  } else if (phase === 5) {
    check('137'); if (Number(root.dataset.messagingDashboardSource) !== 17) throw Error('Consumer switch performed extra source reads');
    root.dataset.messagingDashboardProof += '|consumer:137';
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=messaging-producer]')!.click();
  } else if (phase === 6) {
    check('120'); if (Number(root.dataset.messagingDashboardSource) !== 21) throw Error('Producer switch performed extra source reads');
    root.dataset.messagingDashboardProof += '|producer:120';
    document.querySelector<HTMLButtonElement>('[data-dashboard-profile=messaging-consumer]')!.click();
  } else if (phase === 7) {
    check('137'); if (Number(root.dataset.messagingDashboardSource) !== 25) throw Error('Consumer return performed extra source reads');
    root.dataset.messagingDashboardProof += '|consumer:137';
    root.dataset.messagingDashboardBeforeLocal = root.dataset.messagingDashboardSource;
    document.querySelector<HTMLDetailsElement>('.dashboard-native-picker')!.open = true;
  } else if (phase === 8) {
    const selected = [...document.querySelectorAll<HTMLInputElement>('.dashboard-native-picker input:checked')].map(input => input.getAttribute('aria-label'));
    if (!selected.some(value => value?.startsWith('messaging.process.duration · histogram [s]'))) throw Error('Processing-duration binding not selected');
    if (root.dataset.messagingDashboardSource !== root.dataset.messagingDashboardBeforeLocal) throw Error('Native picker read telemetry');
    document.querySelector<HTMLDetailsElement>('.dashboard-native-picker')!.open = false;
    button('Back to service list')!.click();
  } else if (phase === 9) {
    if (root.dataset.messagingDashboardSource !== root.dataset.messagingDashboardBeforeLocal) throw Error('Return to catalog read telemetry');
    if (document.querySelector<HTMLElement>('#serviceDetail') && !document.querySelector<HTMLElement>('#serviceDetail')!.hidden) throw Error('Service detail still open');
    if (Number(root.dataset.messagingDashboardPlans) !== 5) throw Error('Expected exactly five explicit plans');
    root.dataset.messagingDashboardTerminal = 'true';
  }
  root.dataset.messagingDashboardPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.phase >= 10) return ['Wait'];
  if (value.phase === 1 && !value.catalog) return ['Wait'];
  if ([2, 4, 5, 6, 7].includes(value.phase) && !value.loaded) return ['Wait'];
  if (value.phase === 4 && value.profile !== 'all-spans' || [5, 7].includes(value.phase) && value.profile !== 'messaging-consumer' || value.phase === 6 && value.profile !== 'messaging-producer') return ['Wait'];
  return [step()];
});
