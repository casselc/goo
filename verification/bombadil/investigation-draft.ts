// Same-tab local recovery exploration. No Save/Run/Capture action is offered.
import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document,window})=>{
 const raw=window.sessionStorage.getItem('goo.investigation-draft.v1:'+window.location.pathname);
 return {ready:!!document.querySelector('#documentNew'),stage:Number(window.sessionStorage.getItem('goo.bombadil.draft.stage')||0),publication:raw?JSON.parse(raw).publication:'none',requests:window.performance.getEntriesByType('resource').filter(r=>/\/api\/(documents|captures|explore|measurements)(?:[/?]|$)/.test(r.name)).length,overflow:document.documentElement.scrollWidth>window.innerWidth};
});
export const localOnly=always(()=>state.current.publication!=='unknown'&&state.current.requests===0&&!state.current.overflow);
export const exercised=eventually(()=>state.current.stage===7).within(30,'seconds');
const advance=registerCustomAction('goo.draftRecovery',async(document,window,stage:number)=>{
 const button=(text:string)=>[...document.querySelectorAll<HTMLButtonElement>('#documentPanel button')].find(b=>b.textContent===text&&!b.hidden);
 const input=(id:string,value:string)=>{const field=document.getElementById(id)as HTMLInputElement;field.value=value;field.dispatchEvent(new Event('input',{bubbles:true}))};
 if(stage===0||stage===4){(document.getElementById('documentPanel')as HTMLDetailsElement).open=true;button('New investigation')!.click();input('documentTitle','Local recovery exploration');input('documentQuestion','What authored work survives an explicit reload?');input('documentText','Pending note remains unapplied.');}
 if(stage===1||stage===5){window.sessionStorage.setItem('goo.bombadil.draft.stage',String(stage+1));window.location.reload();return;}
 if(stage===2)button('Restore local draft')!.click();
 if(stage===3)(document.getElementById('documentRevert')as HTMLButtonElement).click();
 if(stage===6)button('Discard local draft')!.click();
 window.sessionStorage.setItem('goo.bombadil.draft.stage',String(stage+1));
});
export const choices=actions(()=>!state.current.ready?['Wait']:state.current.stage<7?[advance(state.current.stage)]:['Wait']);
