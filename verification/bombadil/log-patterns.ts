import {always, eventually} from '@antithesishq/bombadil';
import {extract, actions, registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions, noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';

const state = extract(({document}) => {
  const root = document.documentElement;
  const parent = root.dataset.patternParent ? JSON.parse(root.dataset.patternParent) : null;
  const child = root.dataset.patternChild ? JSON.parse(root.dataset.patternChild) : null;
  const pair = root.dataset.patternPair ? JSON.parse(root.dataset.patternPair) : null;
  const panel = [...document.querySelectorAll('.log-pattern-results')].find(element => element.getClientRects().length > 0);
  const rows = panel ? [...panel.querySelectorAll('tr[data-pattern]')] : [];
  const counts = rows.map(row => BigInt(row.querySelector('td').textContent.split(' · ')[0]));
  const ascending = counts.every((count, index) => !index || counts[index - 1] <= count);
  const descending = counts.every((count, index) => !index || counts[index - 1] >= count);
  const activeApplied = [...document.querySelectorAll('.population-workspace[data-phase="applied"]')].some(element => element.getClientRects().length > 0);
  const raw = document.querySelector('#exploreRows .record-detail .raw-record>pre');
  const inspected = raw ? JSON.parse(raw.textContent) : null;
  return {
    phase: root.dataset.patternPhase || 'initial', reads: Number(root.dataset.patternReads || 0),
    protocols: Number(root.dataset.patternProtocols || 0), staged: document.getElementById('exploreStatus')?.textContent?.includes('No evidence loaded') || false,
    parent: !!parent, panel: !!panel, rows: rows.length, ascending, descending,
    groupsComplete: !!parent && parent.count === '400' && parent.rows.length === 100 && parent.groups.length > 1 && parent.groups.reduce((total, group) => total + BigInt(group.count), 0n) === 400n && parent.groups.some(group => group.g0 === 'lp1:e') && parent.groups.some(group => group.g0 === 'lp1:o'),
    inspected: !!inspected, rawUnchanged: !inspected || !!parent && inspected.Body === parent.rows[0].Body && inspected.SpanId === parent.rows[0].SpanId,
    child: !!child, childCount: child?.count || '',
    runPair: [...document.querySelectorAll('.population-run')].some(button => button.getClientRects().length > 0 && !button.disabled),
    pair: !!pair, partitioned: !!pair && !!parent && BigInt(pair.request.baseline.count) + BigInt(pair.request.selection.count) === BigInt(parent.count),
    pairApplied: activeApplied && !!pair,
    injected: document.defaultView.__patternInjected === true
  };
});
export const completes = eventually(() => state.current.phase === 'complete').within(45, 'seconds');
export const bounded = always(() => state.current.reads <= 4 && state.current.protocols <= 1);
export const noLabelExecution = always(() => !state.current.injected);
export const unchangedRawBody = always(() => state.current.rawUnchanged);
export const fullPopulation = always(() => !state.current.parent || state.current.groupsComplete);
export const localSort = always(() => !['sorting-least', 'sorting-most', 'inspecting', 'closing', 'trend-stage'].includes(state.current.phase) || state.current.reads === 1);
export const localBackAndComparison = always(() => !['back', 'comparison-stage'].includes(state.current.phase) || state.current.reads === 2);
export const terminal = always(() => state.current.phase !== 'complete' || state.current.reads === 4 && state.current.protocols === 1 && state.current.partitioned && state.current.pairApplied);

const step = registerCustomAction('goo.logPatterns', async (document, window, phase) => {
  const root = document.documentElement;
  const activePanel = () => [...document.querySelectorAll('.log-pattern-results')].find(element => element.getClientRects().length > 0);
  const frequentID = 'lp1:p:726571756573742069643dn20636f6d706c6574656420696e20n6d73';
  if (phase === 'initialized') {
    root.dataset.patternReads = root.dataset.patternProtocols = '0';
    const original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = typeof input === 'string' ? input : input.url;
      if (init?.method === 'POST' && /\/api\/(explore|captures|measurements|inventory|discovery)$/.test(url)) root.dataset.patternReads = String(Number(root.dataset.patternReads) + 1);
      if (init?.method === 'POST' && url.endsWith('/api/compare')) root.dataset.patternProtocols = String(Number(root.dataset.patternProtocols) + 1);
      const response = await original(input, init);
      if (response.ok && url.endsWith('/api/explore')) {
        const result = await response.clone().json();
        if (!root.dataset.patternParent) root.dataset.patternParent = JSON.stringify(result);
        else if (!root.dataset.patternChild) root.dataset.patternChild = JSON.stringify(result);
      }
      if (response.ok && url.endsWith('/api/compare')) root.dataset.patternPair = JSON.stringify({request: JSON.parse(init.body), response: await response.clone().json()});
      return response;
    };
  }
  if (phase === 'parent-run' || phase === 'trend-run') document.getElementById('exploreRun').click();
  if (phase === 'sorting-least' || phase === 'sorting-most') {
    const select = activePanel().querySelector('[aria-label="Sort log patterns"]');
    select.value = phase === 'sorting-least' ? 'least' : 'most';
    select.dispatchEvent(new window.Event('change', {bubbles: true}));
  }
  if (phase === 'inspecting') {
    document.getElementById('resultRecords').click();
    document.querySelector('#exploreRows [aria-label="Inspect record"]').click();
  }
  if (phase === 'closing') {
    document.querySelector('#exploreRows [aria-label="Close selected item details"]').click();
    document.getElementById('resultDistribution').click();
  }
  if (phase === 'trend-stage') activePanel().querySelector('tr[data-pattern="' + frequentID + '"] a').click();
  if (phase === 'back') [...document.querySelectorAll('button')].find(button => button.textContent === 'Back to question' && !button.disabled).click();
  if (phase === 'comparison-stage') activePanel().querySelector('tr[data-pattern="' + frequentID + '"] button').click();
  if (phase === 'comparison-run') [...document.querySelectorAll('.population-run')].find(button => button.getClientRects().length > 0 && !button.disabled).click();
  root.dataset.patternPhase = phase;
});
export const exploration = actions(() => {
  const current = state.current;
  if (current.phase === 'initial') return current.staged ? [step('initialized')] : ['Wait'];
  if (current.phase === 'initialized') return [step('parent-run')];
  if (current.phase === 'parent-run') return current.parent && current.panel ? [step('sorting-least')] : ['Wait'];
  if (current.phase === 'sorting-least') return current.ascending ? [step('sorting-most')] : ['Wait'];
  if (current.phase === 'sorting-most') return current.descending ? [step('inspecting')] : ['Wait'];
  if (current.phase === 'inspecting') return current.inspected ? [step('closing')] : ['Wait'];
  if (current.phase === 'closing') return [step('trend-stage')];
  if (current.phase === 'trend-stage') return current.staged ? [step('trend-run')] : ['Wait'];
  if (current.phase === 'trend-run') return current.child ? [step('back')] : ['Wait'];
  if (current.phase === 'back') return current.panel ? [step('comparison-stage')] : ['Wait'];
  if (current.phase === 'comparison-stage') return current.runPair ? [step('comparison-run')] : ['Wait'];
  if (current.phase === 'comparison-run') return current.pairApplied && current.partitioned ? [step('complete')] : ['Wait'];
  return ['Wait'];
});
