// Independent reading driver. Help actions never call application source controls.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const o=(window as any).__guidance460||{},panel=document.querySelector<HTMLElement>('#workspaceHelpPanel'),visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert],[aria-hidden="true"]'),guide=panel?.querySelector<any>('goo-guidance');
 return{version:'goo.guidance460Actor/1',phase:Number(document.documentElement.dataset.guidance460Phase||0),ready:!!document.querySelector('[aria-controls="workspaceHelpPanel"]'),open:visible(panel),page:guide?.data?.guidePage??null,context:guide?.data?.context||null,calls:o.calls||[],hash:window.location.hash,initialHash:o.hash||window.location.hash,overflow:document.documentElement.scrollWidth>window.innerWidth,focusHelp:document.activeElement===document.querySelector('[aria-controls="workspaceHelpPanel"]'),fakeCompletion:/task complete|investigation complete|successfully completed/i.test(guide?.textContent||'')};
});
export const noAcquisition=always(()=>state.current.calls.length===0);
export const noNavigation=always(()=>state.current.hash===state.current.initialHash);
export const readingOnly=always(()=>!state.current.fakeCompletion);
export const fits=always(()=>!state.current.overflow);
export const complete=eventually(()=>state.current.phase===8&&!state.current.open&&state.current.focusHelp).within(16,'seconds');
const step=registerCustomAction('goo.guidance460',async(document,window,next:number)=>{
 const target=window as any,visible=(e:any)=>!!e?.getClientRects().length&&!e.closest('[hidden],[inert],[aria-hidden="true"]'),click=(e:any)=>{if(!e||e.disabled||!visible(e))throw Error('Guidance action unavailable');e.scrollIntoView({block:'nearest'});e.click()},action=(name:string)=>document.querySelector('#workspaceHelpPanel [data-guidance-action="'+name+'"]');
 if(next===1){const o=target.__guidance460={calls:[],hash:window.location.hash},original=window.fetch;window.fetch=function(input,options){const url=new URL(typeof input==='string'?input:input instanceof URL?input.href:input.url,window.location.href),method=String(options?.method||(input instanceof Request?input.method:'GET')).toUpperCase();if(method==='POST'&&url.pathname.startsWith('/api/'))o.calls.push(url.pathname);return Reflect.apply(original,this,[input,options])};click(document.querySelector('[aria-controls="workspaceHelpPanel"]'));}
 else if(next===2)click(action('start-guide'));
 else if(next===3)click(action('next'));
 else if(next===4)click(action('previous'));
 else if(next===5)click(action('replay'));
 else if(next===6)click(action('about'));
 else if(next===7){const summary=document.querySelector<HTMLDetailsElement>('#workspaceHelpPanel details')!;summary.open=!summary.open;summary.open=!summary.open;click(action('start-guide'));}
 else if(next===8)click(action('close'));
 document.documentElement.dataset.guidance460Phase=String(next);
});
export const explore=actions(()=>{const s=state.current;if(s.phase===0)return s.ready?[step(1)]:['Wait'];if(s.phase===1)return s.open?[step(2)]:['Wait'];if(s.phase===2)return s.page===0?[step(3)]:['Wait'];if(s.phase===3)return s.page===1?[step(4)]:['Wait'];if(s.phase===4)return s.page===0?[step(5)]:['Wait'];if(s.phase===5)return[step(6)];if(s.phase===6)return s.page===null?[step(7)]:['Wait'];if(s.phase===7)return[step(8)];return['Wait']});
