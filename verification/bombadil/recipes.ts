// Preview-only exploration. No source execution, saved artifacts or workload writes.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>{const plan=document.getElementById('recipePlan') as HTMLElement;return {version:'goo.recipes.v1',stage:document.documentElement.dataset.gooRecipeStage||'initial',selected:(document.getElementById('recipeChoice') as HTMLSelectElement)?.value,ready:!!document.querySelector('#recipeChoice option[value="logs"]'),visible:!!plan&&!plan.hidden,status:document.getElementById('recipeStatus')?.textContent||'',evidence:document.getElementById('exploreExplain')?.textContent||''}});
export const complete=eventually(()=>state.current.stage==='complete').within(25,'seconds');
export const previewCannotInventAppliedQuestion=always(()=>!state.current.visible);
export const noEvidenceFromPreview=always(()=>state.current.evidence==='');
const act=registerCustomAction('goo.recipes',async(document,_window,id:string,next?:string)=>{const select=document.getElementById('recipeChoice') as HTMLSelectElement;document.querySelector<HTMLDetailsElement>('.reusable-questions')!.open=true;select.value=id;select.dispatchEvent(new Event('change',{bubbles:true}));document.getElementById('recipePreview')?.click();if(next)document.documentElement.dataset.gooRecipeStage=next;});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.stage==='initial')return[act('latency','latency')];if(s.stage==='latency')return[act('errors','errors')];if(s.stage==='errors')return[act('logs','complete')];return[act('latency'),act('errors'),act('logs')];});
