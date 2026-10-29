import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({stage:document.documentElement.dataset.adoptionProbe||'initial',mounted:!!document.querySelector('.environment-adoption'),status:document.querySelector('.environment-adoption [role=status]')?.textContent||'',rows:document.querySelectorAll('.environment-adoption [aria-label="Environment record preview"] tbody tr').length,preview:document.querySelector('.environment-adoption [aria-label="Environment record preview"]')?.textContent||'',detail:!!document.querySelector('.environment-adoption .ui-inspector:not([hidden])'),returned:document.activeElement?.textContent==='Inspect record',unexpected:document.documentElement.dataset.adoptionUnexpected==='true'}));
export const onlyAdoptionReads=always(()=>!state.current.unexpected);
export const explicitScope=always(()=>state.current.rows===0||state.current.preview.includes('Unreported environment'));
export const completes=eventually(()=>state.current.stage==='complete').within(40,'seconds');
const act=registerCustomAction('goo.environmentAdoption',async(document,window,phase:string)=>{
 const box=document.querySelector<HTMLDetailsElement>('.environment-adoption')!;
 if(phase==='open'){
  const native=window.fetch.bind(window);window.fetch=(input,init)=>{const path=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href).pathname;if(init?.method==='POST'&&path!=='/api/environment-scopes'&&path!=='/api/environment-scopes/records'&&!path.startsWith('/api/queries/'))document.documentElement.dataset.adoptionUnexpected='true';return native(input,init)};
  box.open=true;
 }else if(phase==='read')box.querySelector<HTMLButtonElement>('button')!.click();
 else if(phase==='preview'){const b=Array.from(box.querySelectorAll<HTMLButtonElement>('button')).find(b=>b.textContent==='Preview unreported logs');if(!b)throw Error('Owned unreported fixture missing');b.click()}
 else if(phase==='inspect')box.querySelector<HTMLButtonElement>('[aria-label="Environment record preview"] tbody button')!.click();
 else if(phase==='close')box.querySelector<HTMLButtonElement>('[aria-label="Close selected item details"]')!.click();
 document.documentElement.dataset.adoptionProbe=phase==='complete'?'complete':phase;
});
export const explore=actions(()=>{const s=state.current;if(!s.mounted)return['Wait'];if(s.stage==='initial')return[act('open')];if(s.stage==='open')return[act('read')];if(s.stage==='read'&&s.status.includes('observed label groups'))return[act('preview')];if(s.stage==='preview'&&s.rows>0)return[act('inspect')];if(s.stage==='inspect'&&s.detail)return[act('close')];if(s.stage==='close'&&!s.detail&&s.returned)return[act('complete')];return['Wait']});
