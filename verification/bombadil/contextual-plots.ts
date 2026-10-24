import {always, eventually} from '@antithesishq/bombadil';
import {extract, actions, registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions, noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state = extract(({document}) => {
  const root = document.documentElement;
  const d = document.querySelector('dialog[aria-label="Plot attribute values"][open]');
  const parent = root.dataset.plotParent ? JSON.parse(root.dataset.plotParent) : null;
  const result = root.dataset.plotResult ? JSON.parse(root.dataset.plotResult) : null;
  const type = document.getElementById('plotValuesnumericType');
  const unit = document.getElementById('plotValuesnumericUnit');
  const scale = document.getElementById('plotValuesnumericScale');
  const staged = document.getElementById('exploreStatus')?.textContent?.includes('No evidence loaded') || false;
  return {
    phase: root.dataset.plotPhase || 'initial', reads: Number(root.dataset.plotReads || 0),
    dialog: !!d?.open, type: type?.value || '', unit: unit?.value || '', scale: scale?.value || '',
    staged, parent: !!parent, result: !!result,
    plotAction: !!document.querySelector('#exploreRows [aria-label="Plot values of attribute:work.amount"]'),
    inspector: !!document.querySelector('#exploreRows .record-detail'),
    complete: !!document.querySelector('#numericSummary .numeric-distribution') && !!result && result.count === parent.count && result.rows.length === 100 &&
      result.question.environment === parent.question.environment &&
      result.question.fromMs === parent.question.fromMs && result.question.toMs === parent.question.toMs &&
      JSON.stringify(result.question.where) === JSON.stringify(parent.question.where) &&
      !result.question.groupBy?.length && result.question.numeric.type === 'decimal64' &&
      result.question.numeric.scale === 2 && result.question.numeric.field.kind === 'attribute' &&
      result.question.numeric.field.key === 'work.amount' && result.numeric.invalid !== '0' &&
      result.numeric.empty !== '0' && result.numericHistogram.bins.length === 128 &&
      result.numericHistogram.bins.reduce((n, b) => n + BigInt(b.count), 0n).toString() === result.numeric.valid
  };
});
export const completes = eventually(() => state.current.phase === 'complete').within(40, 'seconds');
export const explicitReads = always(() => state.current.reads <= 2);
export const declarationsAreLocal = always(() => !['inspecting', 'opening', 'declaring', 'canceling', 'reopening', 'restaging', 'staged'].includes(state.current.phase) || state.current.reads === 1);
export const blankWithoutInference = always(() => !['opening', 'reopening'].includes(state.current.phase) || state.current.type === '' && state.current.unit === '' && state.current.scale === '');
export const terminalHasFullSourceFacts = always(() => state.current.phase !== 'complete' || state.current.complete && state.current.reads === 2);

const step = registerCustomAction('goo.contextualPlot', async (document, window, phase) => {
  const root = document.documentElement;
  const d = () => document.querySelector('dialog[aria-label="Plot attribute values"][open]');
  const button = (host, text) => [...host.querySelectorAll('button')].find(b => b.textContent === text);
  const set = (id, value, event) => {
    const control = document.getElementById(id);
    control.value = value;
    control.dispatchEvent(new window.Event(event, {bubbles: true}));
  };
  if (phase === 'initialized') {
    root.dataset.plotReads = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = typeof input === 'string' ? input : input.url;
      if (init?.method === 'POST' && /\/api\/(explore|captures|measurements|discovery|inventory)$/.test(url)) root.dataset.plotReads = String(Number(root.dataset.plotReads) + 1);
      const response = await original(input, init);
      if (response.ok && url.endsWith('/api/explore')) {
        const result = await response.clone().json();
        root.dataset[root.dataset.plotParent ? 'plotResult' : 'plotParent'] = JSON.stringify(result);
      }
      return response;
    };
  }
  if (phase === 'parent-reading' || phase === 'plot-reading') document.getElementById('exploreRun').click();
  if (phase === 'inspecting') document.querySelector('#exploreRows [aria-label="Inspect record"]').click();
  if (phase === 'opening' || phase === 'reopening') document.querySelector('#exploreRows [aria-label="Plot values of attribute:work.amount"]').click();
  if (phase === 'declaring' || phase === 'restaging') {
    set('plotValuesnumericType', 'decimal64', 'change');
    set('plotValuesnumericUnit', 'USD', 'input');
    set('plotValuesnumericScale', '2', 'input');
  }
  if (phase === 'canceling') button(d(), 'Cancel').click();
  if (phase === 'staged') button(d(), 'Open value distribution').click();
  root.dataset.plotPhase = phase;
});
export const exploration = actions(() => {
  const s = state.current;
  if (s.phase === 'initial') return s.staged ? [step('initialized')] : ['Wait'];
  if (s.phase === 'initialized') return [step('parent-reading')];
  if (s.phase === 'parent-reading') return s.parent ? [step('inspecting')] : ['Wait'];
  if (s.phase === 'inspecting') return s.plotAction ? [step('opening')] : ['Wait'];
  if (s.phase === 'opening') return s.dialog ? [step('declaring')] : ['Wait'];
  if (s.phase === 'declaring') return s.scale === '2' ? [step('canceling')] : ['Wait'];
  if (s.phase === 'canceling') return !s.dialog ? [step('reopening')] : ['Wait'];
  if (s.phase === 'reopening') return s.dialog ? [step('restaging')] : ['Wait'];
  if (s.phase === 'restaging') return s.scale === '2' ? [step('staged')] : ['Wait'];
  if (s.phase === 'staged') return s.staged && !s.dialog ? [step('plot-reading')] : ['Wait'];
  if (s.phase === 'plot-reading') return s.complete ? [step('complete')] : ['Wait'];
  return ['Wait'];
});
