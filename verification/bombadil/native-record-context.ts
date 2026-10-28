import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({
  nativeScope:document.querySelector('#exploreMeasurementsMode')?.getAttribute('aria-selected')!=='true'||document.querySelector('#minutes')?.closest<HTMLElement>('.scope-control')?.hidden===true,
  scopeClocks:!document.querySelector<HTMLButtonElement>('#measureRun')||document.querySelector<HTMLButtonElement>('#measureRun')!.disabled||document.querySelectorAll('#measureScope time').length===2,
  phase: Number(document.documentElement.dataset.nativeContextPhase || 0),
  source: Number(document.documentElement.dataset.nativeContextSource || 0),
  ready: !!document.querySelector<HTMLButtonElement>('#measureRun') && !document.querySelector<HTMLButtonElement>('#measureRun')!.disabled,
  streams: document.querySelectorAll('#measureResult .native-stream-list > details').length,
  point: !!document.querySelector('#measureResult .native-stream-list > details goo-value-timeline button[data-key="0"]'),
  context: !!document.querySelector('.native-record-context [data-native-record-context=spans]'),
  count: document.querySelector('#exploreCount')?.textContent || '',
  terminal: document.documentElement.dataset.nativeContextTerminal === 'true',
  proof: document.documentElement.dataset.nativeContextProof || '',
  overflow: document.documentElement.scrollWidth > innerWidth,
}));
export const nativeWindowOwnership=always(()=>state.current.nativeScope&&state.current.scopeClocks);
export const bounded = always(() => !state.current.overflow && state.current.source <= 2);
export const complete = eventually(() => state.current.terminal).within(20, 'seconds');
const step = registerCustomAction('goo.nativeRecordContextStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.nativeContextPhase || 0);
  if (phase === 0) {
    root.dataset.nativeContextSource = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href);
      const method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (url.pathname.startsWith('/api/') && method === 'POST' && !url.pathname.endsWith('/plan')) root.dataset.nativeContextSource = String(Number(root.dataset.nativeContextSource || 0) + 1);
      return original(input, init);
    };
    document.querySelector<HTMLButtonElement>('#measureRun')!.click();
  } else if (phase === 1) {
    const input = document.querySelector<HTMLInputElement>('#measureResult input[aria-label="Find a stream"]')!;
    input.value = 'inventory'; input.dispatchEvent(new Event('input', { bubbles: true }));
    const stream = document.querySelector<HTMLDetailsElement>('#measureResult .native-stream-list > details')!;
    stream.open = true;
  } else if (phase === 2) document.querySelector<HTMLButtonElement>('#measureResult .native-stream-list > details goo-value-timeline button[data-key="0"]')!.click();
  else if (phase === 3) {
    if (root.dataset.nativeContextSource !== '1') throw Error('Native selection read telemetry');
    const link = document.querySelector<HTMLAnchorElement>('.native-record-context [data-native-record-context=spans]')!;
    (window as any).nativeContextActorOpener = link;
    root.dataset.nativeContextProof = document.querySelector('.native-record-context')!.getAttribute('data-from-ns')!;
    link.click();
  } else if (phase === 4) {
    if (root.dataset.nativeContextSource !== '1') throw Error('Context stage read telemetry');
    document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  } else if (phase === 5) {
    if (!document.querySelector('#exploreCount')?.textContent?.includes('115')) throw Error('Exact native interval population differs');
    [...document.querySelectorAll<HTMLButtonElement>('button')].find(button => button.textContent === 'Back to measurement')!.click();
  } else if (phase === 6) {
    if (root.dataset.nativeContextSource !== '2') throw Error('Retained native Back read telemetry');
    if (document.activeElement !== (window as any).nativeContextActorOpener) throw Error('Native interval opener focus was not restored');
    if (!document.querySelector('.native-record-context') || root.dataset.nativeContextProof !== document.querySelector('.native-record-context')!.getAttribute('data-from-ns')) throw Error('Native interval selection changed');
    root.dataset.nativeContextTerminal = 'true';
  }
  root.dataset.nativeContextPhase = String(phase + 1);
});
export const workflow = actions(() => {
  const value = state.current;
  if (value.phase >= 7) return ['Wait'];
  if (value.phase === 0 && !value.ready || value.phase === 1 && !value.streams || value.phase === 2 && !value.point || value.phase === 3 && !value.context || value.phase === 5 && !value.count.includes('115')) return ['Wait'];
  return [step()];
});
