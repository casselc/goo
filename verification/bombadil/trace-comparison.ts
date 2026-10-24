import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document, window }) => {
  const element = document.querySelector('goo-trace-comparison') as any;
  const host = document.querySelector('#exploreContext .trace-reference-tray');
  const proof = document.documentElement.dataset.traceComparisonProof;
  return { phase: Number(document.documentElement.dataset.traceComparisonPhase || 0),
    source: Number(document.documentElement.dataset.traceComparisonReads || 0),
    ready: !!document.querySelector('#exploreRun'),
    applied: document.querySelector('#exploreStatus')?.textContent?.includes('7 matching') || false,
    context: document.querySelector('#exploreLogStatus')?.textContent?.includes('captured logs') || false,
    pinned: host?.textContent?.includes('Reference: trace') || false,
    groups: element?.querySelectorAll('.trace-pair').length || 0,
    differences: element?.querySelectorAll('.trace-pair-attributes tbody tr').length || 0,
    proof: proof ? JSON.parse(proof) : null,
    compareDisabled: [...(host?.querySelectorAll('button') || [])].find(button => button.textContent === 'Compare with reference')?.disabled ?? true,
    declaration: { kind: (document.querySelector('#exploreFilterEditor [aria-label$=" namespace"]') as HTMLSelectElement)?.value,
      field: (document.querySelector('#exploreFilterEditor [aria-label$=" field"]') as HTMLInputElement)?.value,
      op: (document.querySelector('#exploreFilterEditor [aria-label$=" operator"]') as HTMLSelectElement)?.value,
      value: (document.querySelector('#exploreFilterEditor [aria-label$=" value"]') as HTMLInputElement)?.value },
    currentQuestion: document.documentElement.dataset.traceComparisonQuestion ? JSON.parse(document.documentElement.dataset.traceComparisonQuestion) : null,
    overflow: document.documentElement.scrollWidth > window.innerWidth };
});
export const bounded = always(() => !state.current.overflow && state.current.source <= 6);
export const complete = eventually(() => state.current.phase === 15 && state.current.source === 6 && state.current.compareDisabled &&
  state.current.proof?.inclusiveDifferenceNs === '30000000' && state.current.proof?.outsideDifferenceNs === '25000000' && state.current.proof?.differentTraceIDs).within(50, 'seconds');
const step = registerCustomAction('goo.traceComparisonStep', async (document, window) => {
  const root = document.documentElement;
  const phase = Number(root.dataset.traceComparisonPhase || 0);
  const next = (n: number) => { root.dataset.traceComparisonPhase = String(n); };
  const button = (name: string, parent: Element | Document = document) => [...parent.querySelectorAll('button')].find(value => value.textContent === name) as HTMLButtonElement;
  const local = (fn: () => void) => {
    const before = root.dataset.traceComparisonCalls || '0'; fn();
    if ((root.dataset.traceComparisonCalls || '0') !== before) throw Error('Local comparison action fetched telemetry');
  };
  if (phase === 0) {
    const original = window.fetch;
    window.fetch = async function (...args: any[]) {
      const request = args[0] instanceof Request ? args[0] : null;
      const url = String(request ? request.url : args[0]);
      if (url.includes('/api/')) {
        root.dataset.traceComparisonCalls = String(Number(root.dataset.traceComparisonCalls || 0) + 1);
        if (url.includes('/api/explore') && typeof args[1]?.body === 'string') {
          const q = JSON.parse(args[1].body);
          if (JSON.stringify(q.where).includes('work.role')) root.dataset.traceComparisonQuestion = JSON.stringify(q);
        }
        if ((args[1]?.method || request?.method || 'GET').toUpperCase() === 'POST') root.dataset.traceComparisonReads = String(Number(root.dataset.traceComparisonReads || 0) + 1);
      }
      return original.apply(window, args as any);
    };
    (document.querySelector('#exploreRun') as HTMLButtonElement).click(); next(1);
  } else if (phase === 1 || phase === 10) {
    (document.querySelector('button[aria-label^="Expand trace "]') as HTMLButtonElement).click(); next(phase + 1);
  } else if (phase === 2) {
    local(() => button('Pin trace page as reference').click()); next(3);
  } else if (phase === 3) {
    (document.querySelector('#exploreContextClose') as HTMLButtonElement).click();
    (document.querySelector('#filterComposer') as HTMLDetailsElement).open = true;
    button('Clear filters', document.querySelector('#exploreFilterEditor') as Element).click(); next(4);
  } else if (phase === 4) {
    button('Add condition', document.querySelector('#exploreFilterEditor') as Element).click(); next(5);
  } else if (phase === 5) {
    const kind = document.querySelector('#exploreFilterEditor select[aria-label$=" namespace"]') as HTMLSelectElement;
    kind.value = 'attribute'; kind.dispatchEvent(new Event('change', { bubbles: true })); next(6);
  } else if (phase === 6) {
    const field = document.querySelector('#exploreFilterEditor input[aria-label$=" field"]') as HTMLInputElement;
    field.value = 'work.role'; field.dispatchEvent(new Event('input', { bubbles: true })); next(7);
  } else if (phase === 7) {
    const value = document.querySelector('#exploreFilterEditor input[aria-label$=" value"]') as HTMLInputElement;
    value.value = 'current'; value.dispatchEvent(new Event('input', { bubbles: true })); next(8);
  } else if (phase === 8) {
    (document.querySelector('#minutes') as HTMLInputElement).value = '1';
    (document.querySelector('#exploreFreshWindow') as HTMLButtonElement).click(); next(9);
  } else if (phase === 9) {
    if (Number(root.dataset.traceComparisonReads) !== 3) throw Error('Draft transition read source');
    const editor = document.querySelector('#exploreFilterEditor') as Element;
    const read = (suffix: string) => (editor.querySelector('[aria-label$=" ' + suffix + '"]') as HTMLInputElement)?.value;
    if (read('namespace') !== 'attribute' || read('field') !== 'work.role' || read('operator') !== 'eq' || read('value') !== 'current') throw Error('Current trace declaration not applied');
    (document.querySelector('#exploreRun') as HTMLButtonElement).click(); next(10);
  } else if (phase === 11) {
    const q = JSON.parse(root.dataset.traceComparisonQuestion || 'null'), leaves: any[] = [];
    const walk = (p: any) => { if (p.all || p.any) (p.all || p.any).forEach(walk); else leaves.push(p); };
    if (!q) throw Error('No actual current source request'); walk(q.where);
    if (leaves.length !== 1 || leaves[0].field?.kind !== 'attribute' || leaves[0].field?.key !== 'work.role' || leaves[0].op !== 'eq' || leaves[0].value !== 'current') throw Error('Actual current query differs from authored selector');
    local(() => button('Compare with reference').click());
    const element = document.querySelector('goo-trace-comparison') as any, input = element.data;
    const model = (window as any).GooTracePageComparison.compare(input.baseline.rows, input.selection.rows, input.baseline.truncated, input.selection.truncated);
    const match = model.groups.find((group: any) => group.operation === 'GET /checkout');
    root.dataset.traceComparisonRoot = String(model.groups.indexOf(match));
    root.dataset.traceComparisonProof = JSON.stringify({ inclusiveDifferenceNs: match.inclusiveDifferenceNs,
      outsideDifferenceNs: match.outsideDifferenceNs, differentTraceIDs: model.baseline.traceId !== model.selection.traceId,
      ambiguous: model.groups.some((group: any) => group.operation === 'prepare' && group.matchState === 'ambiguous') });
    next(12);
  } else if (phase === 12) {
    const element = document.querySelector('goo-trace-comparison') as HTMLElement;
    local(() => (element.querySelector('[data-trace-pair="' + root.dataset.traceComparisonRoot + '"]') as HTMLButtonElement).click()); next(13);
  } else if (phase === 13) {
    const element = document.querySelector('goo-trace-comparison') as HTMLElement;
    if (!element.querySelector('.trace-pair-attributes')?.textContent?.includes('attribute:work.empty')) throw Error('Missing qualified attribute difference');
    const details = element.querySelector('details') as HTMLDetailsElement;
    details.open = false;
    const before = element.querySelector('.trace-pair-attributes')?.textContent;
    local(() => (element.querySelector('[data-trace-pair="' + root.dataset.traceComparisonRoot + '"]') as HTMLButtonElement).click());
    if (element.querySelector('.trace-pair-attributes')?.textContent !== before) throw Error('Closed pair action changed evidence');
    details.open = true; next(14);
  } else if (phase === 14) {
    local(() => button('Clear trace reference').click()); next(15);
  }
});
export const choices = actions(() => {
  const s = state.current;
  if (!s.ready || s.phase === 15) return ['Wait'];
  if ((s.phase === 1 || s.phase === 10) && !s.applied || (s.phase === 2 || s.phase === 11) && !s.context || s.phase === 12 && !s.groups || s.phase === 13 && !s.differences) return ['Wait'];
  return [step()];
});
