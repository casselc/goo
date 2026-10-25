import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.stockPhase || 0), reads: Number(document.documentElement.dataset.stockReads || 0),
  ready: !!document.querySelector('#measureResult goo-value-timeline'), terminal: document.documentElement.dataset.stockTerminal === 'true' }));
export const bounded = always(() => state.current.reads <= 1);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.stockStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.stockPhase || 0);
  if (phase === 0) {
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.stockReads = String(Number(root.dataset.stockReads || 0) + 1);
      return original(input, init);
    };
    document.querySelector<HTMLButtonElement>('#measureRun')!.click();
  } else if (phase === 1) {
    const input = document.querySelector<HTMLInputElement>('#measureResult input[aria-label="Find a stream"]')!;
    input.value = 'generic'; input.dispatchEvent(new Event('input', { bubbles: true }));
  } else if (phase === 2) {
    const stream = document.querySelector<HTMLDetailsElement>('#measureResult .native-stream-list > details')!;
    if (!stream.open) stream.querySelector<HTMLElement>(':scope > summary')!.click();
  } else if (phase === 3) {
    document.querySelector<HTMLButtonElement>('#measureResult goo-value-timeline button[data-key="1"]')!.click();
  } else if (phase === 4) {
    const text = document.querySelector('#measureResult .native-stream-list')!.textContent || '';
    if (!text.includes('Accumulation epoch:') || !text.includes('-1.25')) throw Error('Selected stock value or epoch missing');
    if (document.querySelector('#measureResult [data-native-record-context]')) throw Error('Stock observation invented an interval');
    if (root.dataset.stockReads !== '1') throw Error('Local stock controls read telemetry');
    root.dataset.stockTerminal = 'true';
  }
  root.dataset.stockPhase = String(phase + 1);
});
export const workflow = actions(() => state.current.phase >= 5 ? ['Wait'] : state.current.phase === 1 && !state.current.ready ? ['Wait'] : [step()]);
