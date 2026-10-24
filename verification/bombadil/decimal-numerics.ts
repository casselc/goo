import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({
  document
}) => {
  var _comparison$querySele, _document$getElementB, _document$getElementB2, _document$getElementB3, _document$getElementB4;
  const root = document.documentElement,
    parent = root.dataset.decimalParent ? JSON.parse(root.dataset.decimalParent) : null,
    children = root.dataset.decimalChildren ? JSON.parse(root.dataset.decimalChildren) : [],
    plan = root.dataset.decimalPlan ? JSON.parse(root.dataset.decimalPlan) : null,
    comparison = document.querySelector('.population-breakdown .numeric-comparison'),
    mean = comparison === null || comparison === void 0 || (_comparison$querySele = comparison.querySelectorAll('.signed-comparison tbody>tr')[0]) === null || _comparison$querySele === void 0 || (_comparison$querySele = _comparison$querySele.querySelectorAll('.signed-comparison-cell')[1]) === null || _comparison$querySele === void 0 ? void 0 : _comparison$querySele.getAttribute('data-value'),
    parts = mean === null || mean === void 0 ? void 0 : mean.split('/'),
    exactMean = !!parts && BigInt(parts[0]) * 300n === 983n * BigInt(parts[1] || '1');
  return {
    phase: root.dataset.decimalPhase || 'initial',
    reads: Number(root.dataset.decimalReads || 0),
    plans: Number(root.dataset.decimalPlans || 0),
    protocols: Number(root.dataset.decimalProtocols || 0),
    staged: ((_document$getElementB = document.getElementById('exploreStatus')) === null || _document$getElementB === void 0 || (_document$getElementB = _document$getElementB.textContent) === null || _document$getElementB === void 0 ? void 0 : _document$getElementB.includes('No evidence loaded')) || false,
    scale: ((_document$getElementB2 = document.getElementById('numericScale')) === null || _document$getElementB2 === void 0 ? void 0 : _document$getElementB2.value) || '',
    type: ((_document$getElementB3 = document.getElementById('numericType')) === null || _document$getElementB3 === void 0 ? void 0 : _document$getElementB3.value) || '',
    applied: ((_document$getElementB4 = document.getElementById('exploration')) === null || _document$getElementB4 === void 0 ? void 0 : _document$getElementB4.getAttribute('data-question-phase')) === 'applied',
    parent: !!parent,
    parentCount: (parent === null || parent === void 0 ? void 0 : parent.count) || '',
    valid: (parent === null || parent === void 0 ? void 0 : parent.numeric.valid) || '',
    wireScale: parent === null || parent === void 0 ? void 0 : parent.numeric.scale,
    rows: (parent === null || parent === void 0 ? void 0 : parent.rows.length) || 0,
    bins: (parent === null || parent === void 0 ? void 0 : parent.numericHistogram.bins.length) || 0,
    binSum: (parent === null || parent === void 0 ? void 0 : parent.numericHistogram.bins.reduce((n, b) => n + BigInt(b.count), 0n).toString()) || '',
    maximum: (parent === null || parent === void 0 ? void 0 : parent.numeric.max) || '',
    selected: children.some(x => x.count === '30' && x.numeric.scale === 2),
    range: plan ? plan.fromBin === 72 && plan.toBin === 74 && plan.from === '128' && plan.to === '512' : true,
    parentVisible: !!document.querySelector('#numericSummary .numeric-distribution'),
    selectionVisible: !!document.querySelector('#numericSummary .numeric-distribution-selection'),
    runPair: [...document.querySelectorAll('.population-run')].some(b => b.getClientRects().length > 0 && !b.disabled),
    comparable: (comparison === null || comparison === void 0 ? void 0 : comparison.getAttribute('data-numeric-state')) === 'comparable',
    exactMean
  };
});
export const complete = eventually(() => state.current.phase === 'complete').within(45, 'seconds');
export const bounded = always(() => state.current.reads <= 4 && state.current.plans <= 2 && state.current.protocols <= 1);
export const stageIsLocal = always(() => !['initial', 'initialized'].includes(state.current.phase) || state.current.reads === 0);
export const firstBrushIsLocal = always(() => state.current.phase !== 'brushing' || state.current.reads === 1 && state.current.plans === 0);
export const retainedBrushIsLocal = always(() => !['back', 'brushing-again', 'comparison-staging'].includes(state.current.phase) || state.current.reads === 2);
export const fullSourceCoefficients = always(() => !state.current.parent || state.current.parentCount === '400' && state.current.rows === 100 && state.current.valid === '160' && state.current.wireScale === 2 && state.current.bins === 128 && state.current.binSum === '160' && state.current.maximum === '9223372036854775807');
export const exactRange = always(() => state.current.range);
export const actualTerminal = always(() => state.current.phase !== 'complete' || state.current.reads === 4 && state.current.plans === 2 && state.current.protocols === 1 && state.current.selected && state.current.comparable && state.current.exactMean);
const step = registerCustomAction('goo.decimalNumerics', async (document, window, phase) => {
  const root = document.documentElement,
    click = (host, text) => [...host.querySelectorAll('button')].find(b => b.textContent === text && !b.hidden).click();
  if (phase === 'initialized') {
    root.dataset.decimalReads = root.dataset.decimalPlans = root.dataset.decimalProtocols = '0';
    root.dataset.decimalChildren = '[]';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
      if ((init === null || init === void 0 ? void 0 : init.method) === 'POST') {
        if (/\/api\/(explore|captures|measurements|inventory|discovery)$/.test(url)) root.dataset.decimalReads = String(Number(root.dataset.decimalReads) + 1);
        if (url.endsWith('/api/numeric-histograms/selection')) root.dataset.decimalPlans = String(Number(root.dataset.decimalPlans) + 1);
        if (url.endsWith('/api/compare')) root.dataset.decimalProtocols = String(Number(root.dataset.decimalProtocols) + 1);
      }
      const response = await original(input, init);
      if (response.ok && url.endsWith('/api/explore')) {
        const result = await response.clone().json();
        if (!root.dataset.decimalParent) root.dataset.decimalParent = JSON.stringify(result);else {
          const children = JSON.parse(root.dataset.decimalChildren);
          children.push(result);
          root.dataset.decimalChildren = JSON.stringify(children);
        }
      }
      if (response.ok && url.endsWith('/api/numeric-histograms/selection')) root.dataset.decimalPlan = JSON.stringify(await response.clone().json());
      return response;
    };
  }
  if (phase === 'parent-reading') document.getElementById('exploreRun').click();
  if (phase === 'brushing' || phase === 'brushing-again') {
    const bar = document.querySelector('#numericSummary .numeric-distribution button[data-key="72"]');
    bar.focus();
    for (const [key, shiftKey] of [['Escape', false], [' ', false], ['ArrowRight', true], ['Enter', false]]) (document.activeElement || bar).dispatchEvent(new window.KeyboardEvent('keydown', {
      key,
      shiftKey,
      bubbles: true,
      cancelable: true
    }));
  }
  if (phase === 'selected-reading') click(document.querySelector('#numericSummary .numeric-distribution'), 'Run selected population');
  if (phase === 'back') click(document, 'Back to question');
  if (phase === 'comparison-staging') click(document.querySelector('#numericSummary .numeric-distribution'), 'Stage numeric comparison with rest');
  if (phase === 'comparison-reading') document.querySelector('.population-run').click();
  root.dataset.decimalPhase = phase;
});
export const explore = actions(() => {
  const s = state.current;
  if (s.phase === 'initial') return s.staged && s.type === 'decimal64' && s.scale === '2' ? [step('initialized')] : ['Wait'];
  if (s.phase === 'initialized') return [step('parent-reading')];
  if (s.phase === 'parent-reading') return s.parent && s.applied ? [step('brushing')] : ['Wait'];
  if (s.phase === 'brushing') return s.selectionVisible ? [step('selected-reading')] : ['Wait'];
  if (s.phase === 'selected-reading') return s.selected && s.applied ? [step('back')] : ['Wait'];
  if (s.phase === 'back') return s.parentVisible ? [step('brushing-again')] : ['Wait'];
  if (s.phase === 'brushing-again') return s.selectionVisible ? [step('comparison-staging')] : ['Wait'];
  if (s.phase === 'comparison-staging') return s.runPair ? [step('comparison-reading')] : ['Wait'];
  if (s.phase === 'comparison-reading') return s.comparable && s.exactMean ? [step('complete')] : ['Wait'];
  return ['Wait'];
});
