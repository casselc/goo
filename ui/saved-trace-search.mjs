import {validate as validateSaved} from './saved-capture-protocol.mjs';
/* Held immutable trace-search evidence. This owner never reads telemetry or custody. */
import {validateTraceSearchRequest} from './question-protocol.mjs';
import {validateResult} from './trace-search-result.mjs';
import {render} from './trace-search-view.mjs';
import * as Environment from './record-environment.mjs';
import * as UI from './components.mjs';
import {actionVisible} from './visualizations.mjs';
function checked(page){
 validateSaved(page);
 const {receipt,traceSearch}=page||{};
 if(!/^[a-f0-9]{64}$/.test(receipt?.id)||receipt.question?.kind!=='trace-search'||!traceSearch)throw Error('Stored search differs from its immutable receipt.');
 validateTraceSearchRequest(receipt.question.traceSearch);validateResult(traceSearch,receipt.question.traceSearch,receipt.question.traceSearchPosition);
 return structuredClone({receipt,traceSearch});
}
function traceQuestion(request,traceId){return {signal:'spans',...Environment.questionFields(request),fromMs:request.fromMs,toMs:request.toMs,limit:100,offset:0,where:{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:traceId}]}}}
function mount(host,{active=()=>true,onOpenTrace=()=>false}={}){
 let held=null,revision=0,lease=0,suspended=false,disposed=false;const actions=new Map();
 const current=(page,version)=>!disposed&&host.isConnected&&held===page&&revision===version;
 function bind(action,button){const page=held,version=revision,actionLease=lease;return ()=>{if(actionLease!==lease||!current(page,version)||suspended||!active()||!host.contains(button)||!actionVisible(button)||button.closest?.('[inert]'))return false;return action()}}
 function show(page){const next=checked(page);if(disposed)return false;held=next;revision++;lease++;suspended=false;actions.clear();const version=revision;
  render(host,next.traceSearch,{actionLabel:'Open trace question',onOpenTrace:(item,button)=>{if(!current(next,version)||suspended||!active()||!actionVisible(button))return false;const source={current:()=>current(next,version),receipt:next.receipt,result:next.traceSearch,opener:button};return onOpenTrace({question:traceQuestion(next.traceSearch.request,item.traceId),traceId:item.traceId,witnesses:structuredClone(item.witnesses)},source)===true}});
  // Rebind every action to the current held version, including captured old callbacks.
  for(const button of host.querySelectorAll('button')){const action=button.onclick;actions.set(button,action);button.onclick=bind(action,button)}
  host.append(UI.node('p','Stored search evidence · '+next.traceSearch.state+' · bounded trace list; source completeness unknown.','evidence-note'));
  if(next.traceSearch.warnings.length)host.append(UI.disclosure('Search warnings',next.traceSearch.warnings.map(text=>UI.node('p',text))));
  host.append(UI.disclosure('Captured search plan',[UI.node('pre',JSON.stringify(next.traceSearch.plan,null,2))]));return true;
 }
 function suspend(){suspended=true;lease++;}
 function resume(){if(disposed||!suspended||!host.isConnected||!active())return false;suspended=false;lease++;for(const [button,action]of actions)if(button.isConnected)button.onclick=bind(action,button);return true;}
 // Parent custody remains valid while the child is active; only actions are suspended.
 function hold(){suspend();}
 function dispose(){disposed=true;revision++;held=null;actions.clear();host.replaceChildren()}
 return {show,suspend,resume,hold,dispose};
}
export {checked,traceQuestion,mount};
