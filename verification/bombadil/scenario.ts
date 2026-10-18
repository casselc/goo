import { eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const scenarioResult = extract(({ document, window }) => {
  const scenario = new URLSearchParams(window.location.hash.slice(1)).get('scenario') || '';
  const text = document.querySelector('#result')?.textContent?.trim() || '';
  let result: { id?: string; scenario?: string } = {};
  try { result = text ? JSON.parse(text) : {}; } catch { /* retain failure text */ }
  return { scenario, text, id: result.id || '', returnedScenario: result.scenario || '',
    canSubmit: !!document.querySelector('#submit') && !text && ['retry', 'reject', 'missing'].includes(scenario) };
});
export const requestIsAccepted = eventually(() => !!scenarioResult.current.id &&
  scenarioResult.current.returnedScenario === scenarioResult.current.scenario).within(5, 'seconds');

const submit = registerCustomAction('sample.submit-scenario', async (document, window, scenario: string) => {
  const select = document.querySelector<HTMLSelectElement>('#scenario');
  const button = document.querySelector<HTMLButtonElement>('#submit');
  const result = document.querySelector('#result');
  if (!select || !button || !result || result.textContent?.trim() || !['retry', 'reject', 'missing'].includes(scenario)) {
    throw new Error('Scenario submission is unavailable');
  }
  select.value = scenario;
  select.dispatchEvent(new Event('change', { bubbles: true }));
  // This sample does not disable its button. Await the real result so another
  // generated action cannot submit again while the first request is in flight.
  await new Promise<void>((resolve, reject) => {
    const observer = new MutationObserver(() => {
      if (result.textContent?.trim()) { observer.disconnect(); window.clearTimeout(timer); resolve(); }
    });
    const timer = window.setTimeout(() => { observer.disconnect(); reject(new Error('Scenario request timed out')); }, 5000);
    observer.observe(result, { childList: true, subtree: true, characterData: true });
    button.click();
  });
});
export const scenarioActions = actions(() => scenarioResult.current.canSubmit
  ? [submit(scenarioResult.current.scenario)] : ['Wait']);
