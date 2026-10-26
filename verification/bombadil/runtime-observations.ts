import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.runtimePhase || 0),
  source: Number(document.documentElement.dataset.runtimeSource || 0), baseline: Number(document.documentElement.dataset.runtimeBaseline || 0),
  terminal: document.documentElement.dataset.runtimeTerminal === 'true',
  catalog: [...document.querySelectorAll('button')].some(b => b.textContent === 'Open service jobs / runtime-worker'),
  ready: document.querySelector('.service-dashboard')?.getAttribute('data-state') === 'loaded' && document.querySelectorAll('.runtime-observations').length === 2,
  nativeReady: !!document.querySelector('#measureRun') && !document.querySelector<HTMLButtonElement>('#measureRun')!.disabled,
  nativeComplete: !!document.querySelector('#measureResult')?.children.length && !document.querySelector<HTMLButtonElement>('#measureCapture')?.disabled }));
export const bounded = always(() => state.current.source <= 10);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.runtimeStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.runtimePhase || 0);
  if (phase === 0) {
    const environment = new URLSearchParams(window.location.hash.slice(1)).get('runtimeEnvironment')!, original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.runtimeSource = String(Number(root.dataset.runtimeSource || 0) + 1);
      const response = await original(input, init);
      if (response.ok && method === 'POST' && url.pathname === '/api/measurements') {
        const data = await response.clone().json();
        if (data.question.definition.stockDomain !== 'nonnegative-integer' || data.stocks.series.length !== 1 || !data.question.definition.source.exactStream) throw Error('Runtime native identity/domain differs');
      }
      return response;
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click();
    const input = document.querySelector<HTMLInputElement>('#inventoryEnvironment')!; input.value = environment; input.dispatchEvent(new Event('input', { bubbles: true }));
    document.querySelector<HTMLButtonElement>('#inventoryRun')!.click();
  } else if (phase === 1) {
    const button = [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Open service jobs / runtime-worker')!;
    button.scrollIntoView({ block: 'center' }); button.click();
  } else if (phase === 2) {
    root.dataset.runtimeBaseline = root.dataset.runtimeSource;
    const summary = document.querySelector<HTMLElement>('.runtime-observations[data-runtime-metric="go.memory.used"] details > summary')!;
    summary.scrollIntoView({ block: 'center' }); summary.click();
  } else if (phase === 3) {
    if (root.dataset.runtimeSource !== root.dataset.runtimeBaseline) throw Error('Disclosure read telemetry');
    const link = document.querySelector<HTMLAnchorElement>('.runtime-observations[data-runtime-metric="go.memory.used"] a')!;
    link.scrollIntoView({ block: 'center' }); link.click();
  } else if (phase === 4) {
    if (root.dataset.runtimeSource !== root.dataset.runtimeBaseline) throw Error('Native preview read telemetry');
    document.querySelector<HTMLButtonElement>('#measureRun')!.click();
  } else if (phase === 5) {
    if (Number(root.dataset.runtimeSource) !== Number(root.dataset.runtimeBaseline) + 1) throw Error('Explicit native Run budget differs');
    document.querySelector<HTMLButtonElement>('#serviceReturn button')!.click();
  } else if (phase === 6) {
    if (Number(root.dataset.runtimeSource) !== Number(root.dataset.runtimeBaseline) + 1 || document.querySelectorAll('.runtime-observations').length !== 2) throw Error('Retained runtime return reread or lost evidence');
    root.dataset.runtimeTerminal = 'true';
  }
  root.dataset.runtimePhase = String(phase + 1);
});
export const workflow = actions(() => {
  const s = state.current;
  if (s.phase >= 7 || s.phase === 1 && !s.catalog || s.phase === 2 && !s.ready || s.phase === 4 && !s.nativeReady || s.phase === 5 && !s.nativeComplete) return ['Wait'];
  return [step()];
});
