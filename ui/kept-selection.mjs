/* App-owned selection action; reusable record display elements do no IO. */
import * as UI from './components.mjs';
import * as Protocol from './kept-protocol.mjs';
import {client} from './kept-client.mjs';
const n=UI.node;
function mount(host,selection,{active=()=>true,onState=()=>{},expected}={}){
 const frozen=structuredClone(selection),status=n('p','Keep these exact server-observed records without a reread or required label.'),keep=n('button','Keep selected');status.setAttribute('role','status');host.append(keep,status);let intent=null,busy=false,done=false;
 keep.onclick=async()=>{if(busy||done||!host.isConnected||!active())return;const ownFocus=document.activeElement===keep;try{intent ||= Protocol.intent({version:'goo/keep-records/v1',requestID:client.key(),...frozen,...(expected?{expected:structuredClone(expected)}:{})});busy=true;keep.disabled=true;keep.textContent='Keeping…';onState({submitted:true,busy:true});const value=await client.send(intent);if(!host.isConnected||!active())return;done=true;keep.textContent='Kept';status.replaceChildren(n('span','Kept '+value.members.length+' observations. '));const open=n('a','Open saved');open.href='#kept='+value.id;status.append(open);if(ownFocus&&document.activeElement===keep)open.focus()}catch(e){if(host.isConnected&&active()){keep.textContent=intent?'Retry unchanged keep':'Keep selected';status.textContent=e.message+' · Original observations were not reread. Retry reuses this exact request; Clear selection leaves browser retry recovery available.'}}finally{busy=false;if(host.isConnected&&active()){keep.disabled=done;onState({submitted:!!intent,busy:false})}}};
 return {get submitted(){return !!intent}};
}
export {mount};
