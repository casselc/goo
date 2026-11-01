// Saved fixed-query pair: only explicit Run may acquire either operand.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0);
 const visible=(selector:string)=>[...document.querySelectorAll<HTMLElement>(selector)].find(e=>e.getClientRects().length>0);
 const count=(path:string)=>window.performance.getEntriesByType('resource').filter(e=>new URL(e.name).pathname===path).length;
 return {version:'goo.pairedWindow309',phase:document.documentElement.dataset.pairedWindowPhase||'initial',
  population:panel?.dataset.phase,reads:count('/api/explore'),compares:count('/api/compare'),
  runDisabled:(visible('.population-run') as HTMLButtonElement)?.disabled,
  comparison:!!visible('[data-composition-role="comparison"] goo-measure-comparison,[data-composition-role="comparison"] goo-group-count-comparison'),
  saveDisabled:panel?.querySelector<HTMLButtonElement>('.population-view-publish')?.disabled,
  editor:!!panel?.querySelector('.population-comparison-window'),
  overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const completes=eventually(()=>state.current.phase==='complete').within(30,'seconds');
export const localDraftsDoNotRead=always(()=>!['draft','discarded','staged','fresh'].includes(state.current.phase)||state.current.reads===2&&state.current.compares===1);
export const boundedReads=always(()=>state.current.reads<=4&&state.current.compares<=2);
export const unappliedDraftBlocksRunAndSave=always(()=>state.current.phase!=='draft'||state.current.runDisabled&&state.current.saveDisabled);
export const stagedPairHasNoOldAnswer=always(()=>!['staged','fresh'].includes(state.current.phase)||state.current.population==='staged'&&!state.current.comparison);
export const noViewportOverflow=always(()=>!state.current.overflow);
export const terminalReadPair=always(()=>state.current.phase!=='complete'||state.current.population==='applied'&&state.current.reads===4&&state.current.compares===2&&state.current.comparison);
const step=registerCustomAction('goo.pairedWindow',async(document,window,phase:string)=>{
 const panel=[...document.querySelectorAll<HTMLElement>('.population-workspace')].find(e=>!e.hidden&&e.getClientRects().length>0)!;
 const button=(name:string)=>[...panel.querySelectorAll<HTMLButtonElement>('button')].find(b=>b.textContent===name)!;
 const editor=()=>{panel.querySelectorAll<HTMLDetailsElement>('.population-scope-context,.population-comparison-window').forEach(e=>e.open=true);return panel.querySelector('.population-comparison-window')!};
 const edit=()=>{const form=editor();for(const label of ['Comparison from (UTC)','Comparison to (UTC, exclusive)']){const input=form.querySelector<HTMLInputElement>('[aria-label="'+label+'"]')!;input.value=new Date(Date.parse(input.value+'Z')+1000).toISOString().slice(0,23);input.dispatchEvent(new Event('input',{bubbles:true}))}};
 if(phase==='initial')button('Run population view').click();
 else if(phase==='reading')edit();
 else if(phase==='draft')button('Discard window draft').click();
 else if(phase==='discarded'){edit();button('Stage window').click()}
 else if(phase==='staged')button('Use fresh comparison window').click();
 else if(phase==='fresh')button('Run population view').click();
 document.documentElement.dataset.pairedWindowPhase=({initial:'reading',reading:'draft',draft:'discarded',discarded:'staged',staged:'fresh',fresh:'rereading',rereading:'complete'} as Record<string,string>)[phase];
});
export const explore=actions(()=>{
 const s=state.current;
 if(s.phase==='initial'&&s.population==='staged'&&s.editor)return[step('initial')];
 if(s.phase==='reading'&&s.population==='applied'&&s.comparison)return[step('reading')];
 if(['draft','discarded','staged','fresh'].includes(s.phase))return[step(s.phase)];
 if(s.phase==='rereading'&&s.population==='applied'&&s.comparison)return[step('rereading')];
 return['Wait'];
});
