// Read-only source queries and local inspection. No saves, workload or export actions.
import { always, eventually } from '@antithesishq/bombadil';
import { extract, actions, registerCustomAction } from '@antithesishq/bombadil/browser';
export { noUncaughtExceptions, noUnhandledPromiseRejections } from '@antithesishq/bombadil/browser/defaults/properties';

const inspection = extract(({ document }) => ({
  version:'goo.inspection.v2',
  waterfallStage: document.documentElement.dataset.gooWaterfallStage || 'initial',
  traceAvailable: [...document.querySelectorAll<HTMLButtonElement>('#exploreRows [data-expand-trace]')].some(button=>button.getClientRects().length>0&&!button.disabled),
  fieldMenuStage: document.documentElement.dataset.gooFieldMenuStage || 'initial',
  fieldMenus: [...document.querySelectorAll<HTMLElement>('#exploreRows .ui-inspector:not([hidden]) .attribute-actions .ui-action-menu')].filter(menu=>menu.getClientRects().length>0).length,
  fieldActionsAccessible: [...document.querySelectorAll<HTMLElement>('.attribute-actions>.ui-actions')].every(group=>group.children[0]?.getAttribute('aria-label')?.startsWith('Copy value for ')&&[...group.querySelectorAll('.ui-action-menu>summary')].every(summary=>summary.getAttribute('aria-label')==='Field actions')),
  waterfallRows: document.querySelectorAll('#exploreTrace goo-trace-waterfall .waterfall-row').length,
  waterfalls: [...document.querySelectorAll<any>('goo-trace-waterfall')].every(plot=>{
    const source=plot.data,rows=[...plot.querySelectorAll<HTMLButtonElement>('.waterfall-row')];
    if(plot.dataset.state!=='observed')return rows.length===0;
    return Array.isArray(source?.rows)&&rows.length===source.rows.length&&rows.filter(row=>row.getAttribute('aria-pressed')==='true').length<=1&&rows.every(row=>{
      const index=Number(row.dataset.rowIndex),record=source.rows[index],bar=row.querySelector<HTMLElement>('.span-bar'),track=row.querySelector<HTMLElement>('.span-track');
      const left=parseFloat(bar?.style.left||''),width=parseFloat(bar?.style.width||'');
      return !!record&&row.dataset.traceId===record.TraceId&&row.dataset.spanId===record.SpanId&&row.getAttribute('aria-pressed')===String(index===plot.selectedIndex)&&Number.isFinite(left)&&Number.isFinite(width)&&left>=0&&width>=0&&left+width<=100.000001&&(!row.getClientRects().length||!!track&&track.getBoundingClientRect().width>0);
    });
  }),
  typed: [...document.querySelectorAll('goo-field-value')].every(el=>['present','missing','unknown','unavailable'].includes(el.getAttribute('data-value-state')||'')),
  ready: !!document.querySelector('#exploreRun'),
  busy: document.querySelector<HTMLButtonElement>('#exploreRun')?.disabled,
  records: !!document.querySelector('#exploreRows [data-inspect-record]'),
  viewer: !!document.querySelector('#exploreRows .ui-inspector:not([hidden]) .attribute-viewer'),
  trace: !document.querySelector<HTMLElement>('#exploreContext')?.hidden,
  nearby: !document.querySelector<HTMLElement>('#surroundingContext')?.hidden,
  nearbyAvailable: !!document.querySelector('#exploreRows button[aria-label^="Surrounding logs for"]'),
  selected: document.querySelectorAll('#exploreTrace .waterfall-row[aria-pressed="true"]').length,
  counts: [...document.querySelectorAll('.attribute-count')].map(e => e.textContent || ''),
}));
const recordDetails=extract(({document})=>{const root=document.getElementById('exploreRows');return !root||root.querySelectorAll('.record-attributes').length===0&&root.querySelectorAll('.ui-inspector:not([hidden])').length<=1&&[...root.querySelectorAll<HTMLElement>('.record-detail-row:not([hidden])')].every(row=>!!row.querySelector('.ui-inspector:not([hidden])'));});
export const singleRecordDetail=always(()=>recordDetails.current);
export const fieldActionsKeepAccessibleCopy=always(()=>inspection.current.fieldActionsAccessible);
export const becomesReady = eventually(() => inspection.current.ready).within(10, 'seconds');
export const oneContextInspector = always(() => !(inspection.current.trace && inspection.current.nearby));
export const waterfallRowsMatchSuppliedEvidence = always(() => inspection.current.waterfalls);
export const singleSpanSelection = always(() => inspection.current.selected <= 1);
export const typedPresenceIsExplicit = always(() => inspection.current.typed);
export const attributeCountsAreExplicit = always(() => inspection.current.counts.every(s => /^\d+ of \d+ fields$/.test(s)));

const act = registerCustomAction('goo.inspect', async (document, _window, action: string,stage?: string) => {
  const click = (selector: string) => document.querySelector<HTMLElement>(selector)?.click();
  if (action === 'signal') { const select=document.querySelector<HTMLSelectElement>('#exploreSignal'); if(select){select.value=select.value==='spans'?'logs':'spans';select.dispatchEvent(new Event('change'));} }
  const recordRead=(selector:string)=>{const button=document.querySelector<HTMLButtonElement>(selector),menu=button?.closest<HTMLDetailsElement>('details');if(!button||button.disabled)return;if(menu&&!menu.open)menu.querySelector<HTMLElement>(':scope>summary')?.click();if(!menu||menu.open)button.click()};
  if (action === 'nearby') recordRead('#exploreRows button[aria-label^="Surrounding logs for"]');
  if (action === 'closeNearby') click('#surroundingClose');
  if (action === 'radius') {const select=document.querySelector<HTMLSelectElement>('#surroundingRadius');if(select){select.selectedIndex=(select.selectedIndex+1)%select.options.length;select.dispatchEvent(new Event('change'));}}
  if (action === 'query') click('#exploreRun');
  if (action === 'open') {
    document.querySelector<HTMLButtonElement>('#exploreRows [data-inspect-record]')?.click();
  }
  if (action === 'trace') recordRead('#exploreRows [data-expand-trace]');
  if(action==='openFieldMenu'){const menu=[...document.querySelectorAll<HTMLDetailsElement>('#exploreRows .ui-inspector:not([hidden]) .attribute-actions .ui-action-menu')].find(menu=>menu.getClientRects().length>0);if(menu){menu.querySelector<HTMLElement>('summary')?.click();if(!menu.open)throw Error('Field actions failed to open');document.documentElement.dataset.gooFieldMenuStage='opened'}}
  if(action==='closeFieldMenu'){const menu=document.querySelector<HTMLDetailsElement>('#exploreRows .ui-inspector:not([hidden]) .attribute-actions .ui-action-menu[open]');if(menu){const summary=menu.querySelector<HTMLElement>('summary');summary?.focus();summary?.dispatchEvent(new KeyboardEvent('keydown',{key:'Escape',bubbles:true,cancelable:true}));if(menu.open||document.activeElement!==summary)throw Error('Field actions failed to close and restore focus');document.documentElement.dataset.gooFieldMenuStage='complete'}}
  if (action === 'select' || action === 'nextSpan') {
    const rows=[...document.querySelectorAll<HTMLButtonElement>('#exploreTrace goo-trace-waterfall .waterfall-row')];
    const selected=rows.findIndex(row=>row.getAttribute('aria-pressed')==='true');
    const index=action==='select'?0:(selected+1)%Math.max(1,rows.length);rows[index]?.click();
  }
  if (action === 'verifyWaterfall') {
    const plot=document.querySelector<any>('#exploreTrace goo-trace-waterfall'),selected=plot?.querySelector('.waterfall-row[aria-pressed=true]') as HTMLButtonElement|null;
    if(!plot||!selected)throw Error('Waterfall selection was not controlled by the app adapter');
    const record=plot.data.rows[Number(selected.dataset.rowIndex)];
    if(selected.dataset.traceId!==record.TraceId||selected.dataset.spanId!==record.SpanId||document.querySelector('#exploreSpanDetails .ui-inspector-heading > h3')?.textContent!==record.SpanName)throw Error('Selected waterfall IDs and detail view disagree');
    const close=document.querySelector<HTMLButtonElement>('#exploreSpanDetails button[aria-label="Close selected item details"]');close?.click();if(document.activeElement!==selected||selected.getAttribute('aria-pressed')!=='true')throw Error('Selected inspector lost focus or selection');selected.click();
  }
  if(action==='record')click('#exploreRows [data-inspect-record]');
  if(action==='closeRecord')click('#exploreRows .ui-inspector button[aria-label="Close selected item details"]');
  if(stage)document.documentElement.dataset.gooWaterfallStage=stage;
  if (action === 'close') click('#exploreContextClose');
  if (action === 'scope') {
    const select = document.querySelector<HTMLSelectElement>('#exploreRows .attribute-viewer select');
    if (select) { select.selectedIndex = (select.selectedIndex + 1) % select.options.length; select.dispatchEvent(new Event('change')); }
  }
  if (action.startsWith('search:')) {
    const input = document.querySelector<HTMLInputElement>('#exploreRows .attribute-viewer input[type="search"]');
    if (input) { input.value = action.slice(7); input.dispatchEvent(new Event('input')); }
  }
});
export const inspect = actions(() => {
  const s = inspection.current;
  if (!s.ready || s.busy) return ['Wait'];
  // Once a seeded trace is available, deterministically exercise two row selections
  // before resuming broad read-only exploration. No synthetic trace rows are injected.
  if(s.waterfallStage==='initial'&&s.records&&s.traceAvailable)return[act('trace','opened')];
  if(s.waterfallStage==='opened')return s.trace&&s.waterfallRows?[act('select','selected')]:['Wait'];
  if(s.waterfallStage==='selected')return s.selected?[act('nextSpan','next')]:['Wait'];
  if(s.waterfallStage==='next')return s.selected?[act('verifyWaterfall','complete')]:['Wait'];
  if(s.waterfallStage==='complete'&&s.fieldMenuStage==='initial'&&s.records){if(s.trace)return[act('close')];if(!s.viewer)return[act('record')];if(s.fieldMenus)return[act('openFieldMenu')]}
  if(s.fieldMenuStage==='opened')return[act('closeFieldMenu')];
  const choices = [act('query'),act('signal')];
  if(s.nearbyAvailable)choices.push(act('nearby'));
  if(s.nearby)choices.push(act('radius'),act('closeNearby'));
  if (s.records) choices.push(act('open'), act('trace'),act('record'),act('closeRecord'));
  if (s.viewer) choices.push(act('openFieldMenu'),act('closeFieldMenu'),act('scope'), ...['service', 'missing-key', '<tag>', ''].map(q => act('search:' + q)));
  if(s.trace){choices.push(act('close'));if(s.waterfallRows)choices.push(act('select'),act('nextSpan'));if(s.selected)choices.push(act('verifyWaterfall'))}
  return choices;
});
