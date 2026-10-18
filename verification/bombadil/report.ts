// A bounded sample-app adapter. A successful run submits exactly one real UI
// request; the coordinator obtains the returned identity from captured DOM state.
import { eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const reportResult = extract(({ document }) => {
  const text = document.querySelector('#result')?.textContent?.trim() || '';
  let result: { id?: string; outcome?: string } | null = null;
  try { result = text ? JSON.parse(text) : null; } catch { /* failure text stays observable */ }
  const button = document.querySelector<HTMLButtonElement>('#generate');
  return { text, id: result?.id || '', outcome: result?.outcome || '', canSubmit: !!button && !button.disabled && !text };
});
export const reportBecomesReady = eventually(() =>
  !!reportResult.current.id && reportResult.current.outcome === 'ready',
).within(5, 'seconds');

const submit = registerCustomAction('sample.generate-report', async document => {
  const button = document.querySelector<HTMLButtonElement>('#generate');
  if (!button || button.disabled || document.querySelector('#result')?.textContent?.trim()) {
    throw new Error('Report submission is not currently available');
  }
  button.click();
});
export const reportActions = actions(() => reportResult.current.canSubmit ? [submit()] : ['Wait']);
