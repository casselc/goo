import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';
const state = extract(({ document }) => ({ phase: Number(document.documentElement.dataset.podPhase || 0), source: Number(document.documentElement.dataset.podSource || 0),
  terminal: document.documentElement.dataset.podTerminal === 'true', target: document.documentElement.dataset.podTargetKey || '',
  ready: !!document.querySelector('tr[data-pod-variant]'), observed: document.documentElement.dataset.podMembersObserved === 'true',
  runReady: !!document.querySelector('#exploreRun') && !document.querySelector<HTMLButtonElement>('#exploreRun')!.disabled }));
export const bounded = always(() => state.current.source <= 2);
export const complete = eventually(() => state.current.terminal).within(25, 'seconds');
const step = registerCustomAction('goo.podStep', async (document, window) => {
  const root = document.documentElement, phase = Number(root.dataset.podPhase || 0), clusterKey = JSON.stringify([['missing'], ['present', 'shared-cluster']]), namespaceKey = JSON.stringify(['missing']);
  if (phase === 0) {
    const parameters = new URLSearchParams(window.location.hash.slice(1)), environment = parameters.get('podEnvironment')!, expectedIDs = parameters.get('podExpectedLogIDs')!.split(',').sort(), original = window.fetch.bind(window);
    window.fetch = async (input, init) => {
      const url = new URL(typeof input === 'string' ? input : input instanceof URL ? input.href : input.url, window.location.href), method = init?.method || (input instanceof Request ? input.method : 'GET');
      if (method === 'POST' && url.pathname.startsWith('/api/') && !url.pathname.endsWith('/plan') && !url.pathname.endsWith('/cancel')) root.dataset.podSource = String(Number(root.dataset.podSource || 0) + 1);
      const response = await original(input, init);
      if (response.ok && method === 'POST' && url.pathname === '/api/pod-discovery') {
        const data = await response.clone().json(), row = data.observations.find((r: any) => r.resource['k8s.pod.uid'] === 'uid-a' && !Object.hasOwn(r.resource, 'service.name') && !Object.hasOwn(r.resource, 'k8s.cluster.uid') && !Object.hasOwn(r.resource, 'k8s.namespace.name') && r.resourceSchemaURL === 'https://opentelemetry.io/schemas/1.37.0' && r.families.some((f: any) => f.family === 'gauge'));
        if (!row) throw Error('Exact pod resource variant absent'); root.dataset.podTargetKey = row.key;
      }
      if (response.ok && method === 'POST' && url.pathname === '/api/explore') {
        const data = await response.clone().json(), members = data.rows.map((r: any) => r.Attributes?.['work.id'] || r.LogAttributes?.['work.id'] || r.SpanAttributes?.['work.id']).sort();
        if (data.count !== String(expectedIDs.length) || JSON.stringify(members) !== JSON.stringify(expectedIDs)) throw Error('Actual UID-qualified logs differ from emitted records');
        root.dataset.podMembersObserved = 'true';
      }
      return response;
    };
    document.querySelector<HTMLButtonElement>('[data-workspace=services]')!.click(); document.querySelector<HTMLButtonElement>('#entitiesKubernetes')!.click();
    const env = document.querySelector<HTMLInputElement>('[aria-label="Kubernetes environment"]')!; env.value = environment; env.dispatchEvent(new Event('input', { bubbles: true }));
    const minutes = document.querySelector<HTMLInputElement>('#minutes')!; minutes.value = '2'; minutes.dispatchEvent(new Event('input', { bubbles: true }));
    [...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Find pods')!.click();
  } else if (phase === 1 || phase === 2) {
    if (root.dataset.podSource !== '1') throw Error('Unexpected source operation before local facets');
    const select = document.querySelector<HTMLSelectElement>('[aria-label="Reported Kubernetes ' + (phase === 1 ? 'cluster' : 'namespace') + '"]')!;
    select.value = phase === 1 ? clusterKey : namespaceKey; select.dispatchEvent(new Event('change', { bubbles: true }));
  } else if (phase === 3) {
    const search = document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!; search.value = 'uid-a'; search.dispatchEvent(new Event('input', { bubbles: true }));
  } else if (phase === 4) {
    if (root.dataset.podSource !== '1') throw Error('Local filtering read source');
    const link = document.querySelector<HTMLAnchorElement>('tr[data-pod-variant="' + root.dataset.podTargetKey + '"] [data-pod-signal=logs]')!;
    link.scrollIntoView({ block: 'center' }); link.focus(); link.click();
  } else if (phase === 5) {
    if (root.dataset.podSource !== '1') throw Error('Context staging read source'); document.querySelector<HTMLButtonElement>('#exploreRun')!.click();
  } else if (phase === 6) {const minutes=document.querySelector<HTMLInputElement>('#minutes')!;minutes.value='5';minutes.dispatchEvent(new Event('input',{bubbles:true}));[...document.querySelectorAll<HTMLButtonElement>('button')].find(b => b.textContent === 'Back to Kubernetes')!.click();}
  else if (phase === 7) {
    if (!document.querySelector('tr[data-pod-variant]') || root.dataset.podSource !== '2' || document.querySelector<HTMLSelectElement>('[aria-label="Reported Kubernetes cluster"]')!.value !== clusterKey || document.querySelector<HTMLSelectElement>('[aria-label="Reported Kubernetes namespace"]')!.value !== namespaceKey || document.querySelector<HTMLInputElement>('[aria-label="Find reported pod resource"]')!.value !== 'uid-a') throw Error('Retained pod filters or read budget differs');
    if ((document.activeElement as HTMLElement)?.dataset.podSignal !== 'logs') throw Error('Context opener focus not restored');
    root.dataset.podTerminal = 'true';
  }
  root.dataset.podPhase = String(phase + 1);
});
export const workflow = actions(() => { const s = state.current; if (s.phase >= 8 || s.phase === 1 && (!s.ready || !s.target) || s.phase === 5 && !s.runReady || s.phase === 6 && !s.observed) return ['Wait']; return [step()]; });
