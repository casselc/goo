/* Application-owned reading context. Help never controls telemetry or navigation. */
import {initial,transition,withContext} from './guidance-content.mjs';
import './guidance-view.mjs';
import {nativeControlReady,revealFocus} from './dom-visibility.mjs';
export function contextFor(route,measurements=false){return route==='explore'?(measurements?'measurements':'activity'):({__proto__:null,services:'services',operations:'operations',investigations:'saved',examples:'examples'})[route]||'activity'}
export function mount({route=()=> 'explore'}={}){
 const trigger=document.getElementById('workspaceHelpToggle'),panel=document.getElementById('workspaceHelpPanel'),view=panel?.querySelector('goo-guidance'),work=document.getElementById('workspaceContent'),title=document.getElementById('workspaceTitle');
 if(!trigger||!panel||!view)return {dispose(){}};
 let disposed=false,opened=false,measurements=false,state=initial(contextFor(route()));
 const available=()=>!disposed&&!document.hidden&&panel.isConnected&&trigger.isConnected;
 function update(){const context=contextFor(route(),measurements);if(context===state.context)return;state=withContext(state,context);if(opened)view.data=state}
 function close(){if(!available())return;opened=false;panel.hidden=true;trigger.setAttribute('aria-expanded','false');revealFocus(trigger,title)}
 trigger.hidden=false;trigger.onclick=()=>{
  if(!available()||!nativeControlReady(trigger))return;
  if(opened){close();return}
  update();opened=true;panel.hidden=false;trigger.setAttribute('aria-expanded','true');view.data=state;revealFocus(view.querySelector('h2'),trigger);
 };
 function action(event){
  if(!available()||!opened||!nativeControlReady(view))return;
  const origin=event.composedPath?.().find(element=>element?.localName==='button'&&view.contains(element));if(!origin||!nativeControlReady(origin))return;
  const value=event.detail?.action;
  if(value==='close'){close();return}
  if(value==='work'){opened=false;panel.hidden=true;trigger.setAttribute('aria-expanded','false');revealFocus(work,title);return}
  try{state=transition(state,value);view.data=state}catch{return}
 }
 function workspace(){if(!disposed)update()}
 function mode(event){measurements=event.detail?.measurements===true;if(!disposed)update()}
 function keyboard(event){if(event.key==='Escape'&&!event.defaultPrevented&&!event.isComposing&&event.keyCode!==229&&opened&&available()){event.preventDefault();event.stopPropagation();close()}}
 view.addEventListener('goo-guidance-action',action);panel.addEventListener('keydown',keyboard);globalThis.addEventListener('goo:workspace',workspace);globalThis.addEventListener('goo:explore-mode',mode);
 return {dispose(){if(disposed)return;disposed=true;opened=false;panel.hidden=true;trigger.hidden=true;trigger.setAttribute('aria-expanded','false');trigger.onclick=null;view.removeEventListener('goo-guidance-action',action);panel.removeEventListener('keydown',keyboard);globalThis.removeEventListener('goo:workspace',workspace);globalThis.removeEventListener('goo:explore-mode',mode)}};
}
