import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{const p=document.querySelector('[data-kind=comparison] .comparison-breakdown');return {ready:!!p,rows:p?.querySelectorAll('tr').length||0,recapture:!!p?.querySelector('.breakdown-recapture'),links:[...p?.querySelectorAll('a')||[]].every(a=>a.getAttribute('href')?.startsWith('#step='))}});
export const ready=eventually(()=>state.current.ready).within(15,'seconds');
export const frozen=always(()=>!state.current.ready||state.current.rows===5&&!state.current.recapture&&state.current.links);
const provenance=registerCustomAction('goo.toggleComparisonProvenance',async(document)=>{const d=document.querySelector('[data-kind=comparison] .evidence-comparison-result > details')as HTMLDetailsElement;if(d)d.open=!d.open});
const inspect=registerCustomAction('goo.inspectFocusedQuestion',async(document)=>{const a=document.querySelector('[data-kind=comparison] .comparison-breakdown a')as HTMLAnchorElement;if(a){const s=JSON.parse(decodeURIComponent(a.hash.slice(6)));if(s.kind!=='record'||!s.record.where)throw Error('Invalid focused question')}});
export const explore=actions(()=>state.current.ready?[provenance(),inspect()]:['Wait']);
