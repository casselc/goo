// Deliberately bounded smoke exploration: no saves, exports or workload controls.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const shell = extract(({ document, window }) => ({
  nativeSelected:document.querySelector('#exploreMeasurementsMode')?.getAttribute('aria-selected')==='true',
  savedLookbackHidden: document.querySelector('#minutes')?.closest<HTMLElement>('.scope-control')?.hidden === true,
  liveOwned: document.querySelector('#operations')?.closest('[data-workspace-panel]')?.getAttribute('data-workspace-panel') === 'operations',
  savedHasEditor: !!document.querySelector('[data-workspace-panel=investigations] #operationPackage'),
  examples: document.documentElement.dataset.examples === 'configured',
  visited: (document.documentElement.dataset.catalogVisited || '').split(','),
  catalogs: ['/api/operation-packages','/api/prebuilts','/api/package','/api/example-questions','/api/investigation-workflows'].map(path => window.performance.getEntriesByType('resource').filter(entry => new URL(entry.name).pathname === path).length),
  menuCount: document.querySelectorAll('.entry-menu').length,
  menuExpanded: document.querySelector('.entry-menu')?.getAttribute('aria-expanded') === 'true',
  navExpanded: document.querySelector('.app-sidebar')?.classList.contains('navigation-expanded') === true,
  ready: !!document.querySelector('[data-workspace][aria-current="page"]'),
  selected: [...document.querySelectorAll('[data-workspace][aria-current="page"]')].map(e => e.getAttribute('data-workspace')),
  visible: [...document.querySelectorAll<HTMLElement>('[data-workspace-panel]')].filter(e => !e.hidden).map(e => e.dataset.workspacePanel),
}));

export const sharedMenuState = always(() => !shell.current.ready || (shell.current.menuCount === 1 && shell.current.menuExpanded === shell.current.navExpanded));

export const nativeHasOwnWindow=always(()=>!shell.current.ready||shell.current.selected[0]!=='explore'||!shell.current.nativeSelected||shell.current.savedLookbackHidden);

export const savedHasNoLiveLookback = always(() => !shell.current.ready || shell.current.selected[0] !== 'investigations' || shell.current.savedLookbackHidden);

export const liveAndSavedOwnership = always(() => !shell.current.ready || (shell.current.liveOwned && !shell.current.savedHasEditor));

export const catalogsLoadAtMostOnce = always(() => shell.current.catalogs.every(count => count <= 1));

export const workspaceBecomesReady = eventually(() => shell.current.ready).within(10, 'seconds');

export const oneWorkspaceMatchesNavigation = always(() => !shell.current.ready || (
  shell.current.selected.length === 1 && shell.current.visible.length === 1 &&
  shell.current.selected[0] === shell.current.visible[0]
));

const navigate = registerCustomAction('goo.navigate', async (document, _window, route: string) => {
  const button = document.querySelector<HTMLButtonElement>(`[data-workspace="${route}"]`);
  if (!button) throw new Error(`Missing workspace: ${route}`);
  if (!button.getClientRects().length) document.querySelector<HTMLButtonElement>('.entry-menu')?.click();
  button.click();
  const visited = new Set((document.documentElement.dataset.catalogVisited || '').split(','));
  visited.add(route);document.documentElement.dataset.catalogVisited = [...visited].join(',');
});

export const workspaceNavigation = actions(() => {
  const state = shell.current;
  if (!state.ready) return ['Wait'];
  if (!state.visited.includes('definitions')) return [navigate('definitions')];
  if (state.examples && !state.visited.includes('examples')) return [navigate('examples')];
  return ['explore', 'services', 'operations', 'investigations', 'definitions', ...(state.examples ? ['examples'] : [])].map(route => navigate(route));
});
