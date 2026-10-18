// Deliberately bounded smoke exploration: no saves, exports or workload controls.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const shell = extract(({ document }) => ({
  ready: !!document.querySelector('[data-workspace][aria-current="page"]'),
  selected: [...document.querySelectorAll('[data-workspace][aria-current="page"]')].map(e => e.getAttribute('data-workspace')),
  visible: [...document.querySelectorAll<HTMLElement>('[data-workspace-panel]')].filter(e => !e.hidden).map(e => e.dataset.workspacePanel),
}));

export const workspaceBecomesReady = eventually(() => shell.current.ready).within(10, 'seconds');

export const oneWorkspaceMatchesNavigation = always(() => !shell.current.ready || (
  shell.current.selected.length === 1 && shell.current.visible.length === 1 &&
  shell.current.selected[0] === shell.current.visible[0]
));

const navigate = registerCustomAction('goo.navigate', async (document, _window, route: string) => {
  const button = document.querySelector<HTMLButtonElement>(`[data-workspace="${route}"]`);
  if (!button) throw new Error(`Missing workspace: ${route}`);
  button.click();
});

export const workspaceNavigation = actions(() => shell.current.ready
  ? ['explore', 'operations', 'investigations', 'definitions'].map(route => navigate(route))
  : ['Wait']);
