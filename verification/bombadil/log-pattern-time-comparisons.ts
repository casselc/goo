import {always, eventually} from '@antithesishq/bombadil';
import {extract, actions, registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions, noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state = extract(({document}) => {
  const root = document.documentElement;
  const panel = [...document.querySelectorAll('.pattern-time-comparison')].find(element => element.getClientRects().length > 0);
  const choice = [...document.querySelectorAll('.pattern-time-choice')].find(element => element.getClientRects().length > 0);
  const pair = root.dataset.patternTimePair ? JSON.parse(root.dataset.patternTimePair) : null;
  const child = root.dataset.patternTimeChild ? JSON.parse(root.dataset.patternTimeChild) : null;
  const rows = panel ? [...panel.querySelectorAll('tr[data-pattern]')] : [];
  return {
    phase: root.dataset.patternTimePhase || 'initial', sourceOperations: Number(root.dataset.patternTimeReads || 0), protocols: Number(root.dataset.patternTimeProtocols || 0),
    staged: document.getElementById('exploreStatus')?.textContent?.includes('No evidence loaded') || false,
    choice: !!choice, choiceOpen: !!choice?.open,
    parent: !!root.dataset.patternTimeCurrent,
    runPair: [...document.querySelectorAll('.population-run')].some(button => button.getClientRects().length > 0 && !button.disabled),
    known: panel?.dataset.state === 'known', metric: panel?.querySelector('[aria-label="Pattern comparison measure"]')?.value || '',
    rowCount: rows.length,
    completePair: !!pair && pair.request.baseline.count === '400' && pair.request.selection.count === '400' &&
      pair.request.baseline.question.toMs === pair.request.selection.question.fromMs &&
      pair.request.baseline.question.toMs - pair.request.baseline.question.fromMs === 60000 &&
      pair.request.selection.question.toMs - pair.request.selection.question.fromMs === 60000,
    absenceLabels: rows.some(row => row.dataset.presence === 'not-observed-in-current') && rows.some(row => row.dataset.presence === 'not-observed-in-baseline'),
    commonRate: rows.some(row => row.dataset.baselineRate === '8/3' && row.dataset.currentRate === '8/3'),
    child: !!child, childCount: child?.count || '',
    pairedGeometry: rows.every(row => {
      const cells = [...row.querySelectorAll('.comparison-bar-cell')];
      return cells.length === 2 && cells[0].dataset.scaleMax === cells[1].dataset.scaleMax;
    })
  };
});
export const completes = eventually(() => state.current.phase === 'complete').within(45, 'seconds');
export const bounded = always(() => state.current.sourceOperations <= 4 && state.current.protocols <= 1);
export const localStaging = always(() => !['choosing', 'comparison-stage'].includes(state.current.phase) || state.current.sourceOperations === 1);
export const localMeasures = always(() => !['count-view', 'share-view', 'rate-view', 'side-stage'].includes(state.current.phase) || state.current.sourceOperations === 3);
export const sharedScales = always(() => state.current.pairedGeometry);
export const actualTerminal = always(() => state.current.phase !== 'complete' || state.current.known && state.current.completePair && state.current.absenceLabels && state.current.commonRate && state.current.childCount === '160' && state.current.sourceOperations === 4 && state.current.protocols === 1);
const step = registerCustomAction('goo.patternsAcrossTime', async (document, window, phase) => {
  const root = document.documentElement;
  const visible = selector => [...document.querySelectorAll(selector)].find(element => element.getClientRects().length > 0);
  const panel = () => visible('.pattern-time-comparison');
  if (phase === 'initialized') {
    root.dataset.patternTimeReads = root.dataset.patternTimeProtocols = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = typeof input === 'string' ? input : input.url;
      if (init?.method === 'POST' && /\/api\/(explore|captures)$/.test(url)) root.dataset.patternTimeReads = String(Number(root.dataset.patternTimeReads) + 1);
      if (init?.method === 'POST' && url.endsWith('/api/compare')) root.dataset.patternTimeProtocols = String(Number(root.dataset.patternTimeProtocols) + 1);
      const response = await original(input, init);
      if (response.ok && url.endsWith('/api/explore')) {
        const result = await response.clone().json();
        if (!root.dataset.patternTimeCurrent) root.dataset.patternTimeCurrent = JSON.stringify(result);
        if (!result.question.groupBy?.length) root.dataset.patternTimeChild = JSON.stringify(result);
      }
      if (response.ok && url.endsWith('/api/compare')) root.dataset.patternTimePair = JSON.stringify({request: JSON.parse(init.body), response: await response.clone().json()});
      return response;
    };
  }
  if (phase === 'current-run' || phase === 'side-run') document.getElementById('exploreRun').click();
  if (phase === 'choosing') visible('.pattern-time-choice').open = true;
  if (phase === 'comparison-stage') [...visible('.pattern-time-choice').querySelectorAll('button')].find(button => button.textContent === 'Stage pattern comparison').click();
  if (phase === 'comparison-run') visible('.population-run').click();
  if (['count-view', 'share-view', 'rate-view'].includes(phase)) {
    const input = panel().querySelector('[aria-label="Pattern comparison measure"]');
    input.value = phase.split('-')[0]; input.dispatchEvent(new window.Event('change', {bubbles: true}));
  }
  if (phase === 'side-stage') {
    const row = [...panel().querySelectorAll('tr[data-pattern]')].find(element => element.dataset.presence === 'not-observed-in-current');
    row.querySelector('details').open = true;
    [...row.querySelectorAll('a')].find(link => link.textContent === 'Open baseline pattern logs').click();
  }
  if (phase === 'back') [...document.querySelectorAll('button')].find(button => button.textContent === 'Back to question' && !button.disabled).click();
  root.dataset.patternTimePhase = phase;
});
export const exploration = actions(() => {
  const value = state.current;
  if (value.phase === 'initial') return value.staged ? [step('initialized')] : ['Wait'];
  if (value.phase === 'initialized') return [step('current-run')];
  if (value.phase === 'current-run') return value.parent && value.choice ? [step('choosing')] : ['Wait'];
  if (value.phase === 'choosing') return value.choiceOpen ? [step('comparison-stage')] : ['Wait'];
  if (value.phase === 'comparison-stage') return value.runPair ? [step('comparison-run')] : ['Wait'];
  if (value.phase === 'comparison-run') return value.known && value.completePair ? [step('count-view')] : ['Wait'];
  if (value.phase === 'count-view') return value.metric === 'count' ? [step('share-view')] : ['Wait'];
  if (value.phase === 'share-view') return value.metric === 'share' ? [step('rate-view')] : ['Wait'];
  if (value.phase === 'rate-view') return value.metric === 'rate' && value.absenceLabels ? [step('side-stage')] : ['Wait'];
  if (value.phase === 'side-stage') return value.staged ? [step('side-run')] : ['Wait'];
  if (value.phase === 'side-run') return value.child ? [step('back')] : ['Wait'];
  if (value.phase === 'back') return value.known && value.completePair ? [step('complete')] : ['Wait'];
  return ['Wait'];
});
