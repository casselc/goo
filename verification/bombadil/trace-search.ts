import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => {
  const root = document.documentElement;
  const panel = document.querySelector('.trace-search:not([hidden])');
  const proof = root.dataset.traceSearchProof;
  return { phase: Number(root.dataset.traceSearchPhase || 0), source: Number(root.dataset.traceSearchReads || 0),
    calls: Number(root.dataset.traceSearchCalls || 0), ready: !!document.querySelector('#exploreRun'), parentApplied: document.querySelector('#exploreStatus')?.textContent?.includes('135 matching') || false,
    parent: root.dataset.traceSearchParent ? JSON.parse(root.dataset.traceSearchParent) : null,
    proof: proof ? JSON.parse(proof) : null, panel: !!panel,
    operation: [...document.querySelectorAll('button')].some(button => button.textContent === 'Find traces with this operation' && button.getClientRects().length > 0),
    conditions: panel?.querySelectorAll('.trace-search-condition').length || 0,
    result: !!panel?.querySelector('tr[data-trace-id]'),
    searchReady: [...(panel?.querySelectorAll('button') || [])].some(button => button.textContent === 'Run trace search' && !(button as HTMLButtonElement).disabled),
    logs: document.querySelector('#exploreLogStatus')?.textContent?.includes('captured logs') || false,
    contextClosed: (document.querySelector('#exploreContext') as HTMLElement)?.hidden ?? true,
    terminal: root.dataset.traceSearchTerminal === 'true' };
});
export const bounded = always(() => state.current.source <= 4);
export const complete = eventually(() => state.current.phase === 15 && state.current.source === 4 && state.current.terminal &&
  !state.current.panel && state.current.contextClosed && state.current.proof?.matched === '15' &&
  state.current.proof?.crossService && state.current.proof?.sameSpan && state.current.proof?.repeat120).within(50, 'seconds');
const step = registerCustomAction('goo.traceSearchStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.traceSearchPhase || 0);
  const next = (value: number) => { root.dataset.traceSearchPhase = String(value); };
  const panel = () => document.querySelector('.trace-search:not([hidden])') as HTMLElement;
  const button = (label: string, host: Element | Document = document) => [...host.querySelectorAll('button')].find(item => item.textContent === label) as HTMLButtonElement;
  const local = (fn: () => void) => { const before = root.dataset.traceSearchCalls || '0'; fn(); if ((root.dataset.traceSearchCalls || '0') !== before) throw Error('Local trace-search action issued an API request'); };
  const editor = (index: number) => panel().querySelectorAll('details[data-condition-editor]')[index] as HTMLDetailsElement;
  if (phase === 0) {
    const original = window.fetch;
    root.dataset.traceSearchReads = root.dataset.traceSearchCalls = '0';
    window.fetch = async function (...args: any[]) {
      const request = args[0] instanceof Request ? args[0] : null;
      const url = new URL(String(request ? request.url : args[0]), window.location.href);
      const method = (args[1]?.method || request?.method || 'GET').toUpperCase();
      if (url.pathname.startsWith('/api/')) root.dataset.traceSearchCalls = String(Number(root.dataset.traceSearchCalls || 0) + 1);
      if (method === 'POST' && ['/api/explore', '/api/trace-search'].includes(url.pathname)) root.dataset.traceSearchReads = String(Number(root.dataset.traceSearchReads || 0) + 1);
      const response = await original.apply(window, args as any);
      if (method === 'POST' && response.ok && url.pathname === '/api/explore') {
        const result = await response.clone().json();
        if (!root.dataset.traceSearchParent) {
          root.dataset.traceSearchParent = JSON.stringify({ count: result.count, rows: result.rows.length, question: result.question });
        } else if (result.question.signal === 'spans') {
          const query = result.question, parent = JSON.parse(root.dataset.traceSearchParent).question;
          const expected = root.dataset.traceSearchCross;
          if (query.environment !== parent.environment || query.fromMs !== parent.fromMs || query.toMs !== parent.toMs ||
            query.where?.all?.length !== 1 || query.where.all[0].field?.key !== 'trace_id' || query.where.all[0].value !== expected) throw Error('Open trace changed the original fixed scope or selected reported ID');
        }
      }
      if (method === 'POST' && response.ok && url.pathname === '/api/trace-search') {
        const result = await response.clone().json(), q = result.request;
        const parent = JSON.parse(root.dataset.traceSearchParent).question;
        const cross = result.items.find((item: any) => item.witnesses[0].service !== item.witnesses[1].service);
        const sameSpan = result.items.some((item: any) => item.witnesses[0].spanId === item.witnesses[1].spanId);
        const repeat120 = result.items.some((item: any) => item.witnesses.every((w: any) => w.matchingRecords === '120'));
        if (q.environment !== parent.environment || q.fromMs !== parent.fromMs || q.toMs !== parent.toMs || q.conditions.length !== 2 ||
          q.conditions[1].field?.kind !== 'column' || q.conditions[1].field?.key !== 'status_class' || q.conditions[1].value !== 'error' || !cross || !sameSpan || !repeat120) throw Error('Actual trace search lost cross-span conditions or bounded witness evidence');
        root.dataset.traceSearchCross = cross.traceId;
        root.dataset.traceSearchProof = JSON.stringify({ matched: result.matchedTraces, crossService: !!cross, sameSpan, repeat120 });
      }
      return response;
    };
    (document.querySelector('#exploreRun') as HTMLButtonElement).click(); next(1);
  } else if (phase === 1) {
    const parent = JSON.parse(root.dataset.traceSearchParent || '{}');
    if (parent.count !== '135' || parent.rows !== 100) throw Error('Actor did not read the real bounded parent fixture');
    local(() => (document.querySelector('#exploreRows .record-row button[aria-label="Inspect record"]') as HTMLButtonElement).click()); next(2);
  } else if (phase === 2) {
    local(() => button('Find traces with this operation').click()); next(3);
  } else if (phase === 3) {
    if (editor(0).open) throw Error('Seeded editor should start compact and closed');
    local(() => button('Add span condition', panel()).click()); next(4);
  } else if (phase === 4) {
    if (!editor(1).open) throw Error('New span condition editor did not open');
    local(() => { const field = editor(1).querySelector('input[aria-label^="Condition "][aria-label$=" field"]') as HTMLInputElement;
      field.value = 'status_class'; field.dispatchEvent(new window.Event('input', { bubbles: true })); }); next(5);
  } else if (phase === 5) {
    local(() => { const value = editor(1).querySelector('input[aria-label^="Condition "][aria-label$=" value"]') as HTMLInputElement;
      value.value = 'error'; value.dispatchEvent(new window.Event('input', { bubbles: true })); }); next(6);
  } else if (phase === 6) {
    button('Run trace search', panel()).click(); next(7);
  } else if (phase === 7) {
    if (editor(0).open || editor(1).open) throw Error('Successful Run left the full editors ahead of results');
    local(() => { const row = panel().querySelector('tr[data-trace-id="' + root.dataset.traceSearchCross + '"]') as HTMLElement;
      (row.querySelector('details[data-condition-index="0"] > summary') as HTMLElement).click(); }); next(8);
  } else if (phase === 8) {
    const witness = panel().querySelector('tr[data-trace-id="' + root.dataset.traceSearchCross + '"] details[data-condition-index="0"]') as HTMLElement;
    const description = witness.querySelector('p') as HTMLElement;
    if (!description.textContent?.includes(' ms') || !/\d{4}-\d{2}-\d{2}T/.test(description.textContent) || !description.title.includes(' ns (exact)') || !witness.querySelector('goo-field-value') || !witness.textContent?.includes('column:service')) throw Error('Qualified UTC/human-duration witness proof missing');
    local(() => (editor(0).querySelector(':scope > summary') as HTMLElement).click()); next(9);
  } else if (phase === 9) {
    const field = editor(0).querySelector('input[aria-label^="Condition "][aria-label$=" field"]') as HTMLInputElement;
    root.dataset.traceSearchOriginalField = field.value;
    root.dataset.traceSearchOriginalSummary = panel().querySelector('.trace-condition-summary')?.textContent || '';
    local(() => { (editor(0).querySelector(':scope > summary') as HTMLElement).click();
      field.value = 'trace_id'; field.dispatchEvent(new window.Event('input', { bubbles: true })); }); next(10);
  } else if (phase === 10) {
    if (panel().querySelector('.trace-condition-summary')?.textContent !== root.dataset.traceSearchOriginalSummary) throw Error('Closed editor changed accepted predicate summary');
    local(() => (editor(0).querySelector(':scope > summary') as HTMLElement).click()); next(11);
  } else if (phase === 11) {
    const field = editor(0).querySelector('input[aria-label^="Condition "][aria-label$=" field"]') as HTMLInputElement;
    if (field.value !== root.dataset.traceSearchOriginalField) throw Error('Reopened editor failed to restore accepted predicate');
    local(() => (editor(0).querySelector(':scope > summary') as HTMLElement).click()); next(12);
  } else if (phase === 12) {
    button('Open trace (fresh read)', panel().querySelector('tr[data-trace-id="' + root.dataset.traceSearchCross + '"]') as HTMLElement).click(); next(13);
  } else if (phase === 13) {
    local(() => (document.querySelector('#exploreContextClose') as HTMLButtonElement).click()); next(14);
  } else if (phase === 14) {
    const retainedRun = button('Run trace search', panel()), retainedOpen = button('Open trace (fresh read)', panel());
    local(() => { button('Close trace search', panel()).click(); retainedRun.click(); retainedOpen.click(); });
    root.dataset.traceSearchTerminal = 'true'; next(15);
  }
});
export const choices = actions(() => {
  const s = state.current;
  if (!s.ready || s.phase === 15) return ['Wait'];
  if (s.phase === 1 && (!s.parent || !s.parentApplied) || s.phase === 2 && !s.operation || s.phase >= 3 && s.phase <= 6 && !s.panel ||
    s.phase === 5 && s.conditions !== 2 || s.phase === 6 && !s.searchReady || s.phase === 7 && (!s.result || !s.proof) ||
    s.phase === 13 && (s.source !== 4 || !s.logs)) return ['Wait'];
  return [step()];
});
