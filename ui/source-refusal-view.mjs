/* Display only. The application supplies diagnostics and owns recovery gestures. */
import {node} from './components.mjs';
import {failure,refusal,reasons} from './source-refusal.mjs';
export function renderRefusal(host,value,{prefix=''}={}){
 const supplied=value instanceof Error?failure(value):value;let checked;try{checked=supplied?.refusal?refusal(supplied.refusal):null}catch{checked=null}
 if(!checked){host.textContent=prefix+(supplied?.error||'Read unavailable.');return null;}
 const state=node('goo-result-state');if(host.localName==='goo-result-state'||host.getAttribute?.('role')==='status'||host.closest?.('[role="status"],[aria-live="polite"],[aria-live="assertive"],goo-result-state')){state.setAttribute('role','group');state.setAttribute('announce','off');}state.state='unavailable';state.className='source-refusal';state.dataset.sourceRefusalKind=checked.kind;
 state.append(node('span',prefix+reasons[checked.kind],'source-refusal-reason'));
 const details=node('details',undefined,'source-refusal-details');details.append(node('summary','Technical details'),node('p',checked.detail));state.append(details);host.replaceChildren(state);return state;
}
