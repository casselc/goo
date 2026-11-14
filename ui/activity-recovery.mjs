/* Local continuation from a successful empty observation buffer, never a census. */
import * as UI from './components.mjs';
import * as Environment from './record-environment.mjs';
import * as Shape from './intent-shape.mjs';
import {identity} from './question-session.mjs';
import {validateRecordQuestion} from './record-intent.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
const width=15*60000;
function checked(context){
  if(context?.active!==true||context.clean!==true||!['unknown','paused'].includes(context.phase)||context.staged===true||context.retained!==0)throw Error('Use a clean, completed empty Activity observation.');
  const read=context.completedRead;
  if(!read||!Number.isSafeInteger(read.fromMs)||!Number.isSafeInteger(read.toMs)||read.fromMs<0||read.toMs-read.fromMs!==60000||read.toMs>4102444800000||read.rowCount!==0||read.limited!==false||!Number.isSafeInteger(read.revision)||read.revision<1||typeof read.observedAt!=='string'||!/^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(?:\.\d{1,9})?Z$/.test(read.observedAt)||(!Number.isFinite(Date.parse(read.observedAt))||Date.parse(read.observedAt)<read.toMs))throw Error('A successful unlimited empty observation read is required.');
  if(!context.readWindow||context.readWindow.fromMs!==read.fromMs||context.readWindow.toMs!==read.toMs||!Number.isSafeInteger(context.epoch)||context.epoch<0||typeof context.connection!=='string'||!/^[a-f0-9]{64}$/.test(context.connection))throw Error('Activity observation ownership is unavailable.');
  Shape.object(context.selector,['signal','environment','environmentSelector','where']);
  const signal=context.selector.signal??'logs';if(!['logs','spans'].includes(signal))throw Error('Choose logs or spans.');
  const query={signal,...Environment.questionFields(context.selector),fromMs:read.toMs-width,toMs:read.toMs,limit:100,offset:0,...(context.selector.where?{where:structuredClone(context.selector.where)}:{})};
  validateRecordQuestion(query);return {query,read};
}
function plan(context){return checked(context).query}
function lease(context){
  try{const {query,read}=checked(context);return JSON.stringify([context.epoch,context.connection,read.revision,read.fromMs,read.toMs,read.observedAt,identity(query)]);}catch{return null}
}
function mount(host,{current,pause,layout=()=>null,onStage,onError=()=>{}}={}){
  const n=UI.node,section=n('div',undefined,'activity-recovery'),button=n('button','Search wider window'),details=n('details'),summary=n('summary','Read window'),facts=n('p');button.type='button';section.hidden=true;details.append(summary,facts);section.append(button,details);const meta=host.querySelector('.tail-meta');if(meta)meta.after(section);else host.append(section);
  let disposed=false,version=0,bound=null;
  function active(control,key,revision){return !disposed&&version===revision&&bound===key&&!!key&&lease(current())===key&&nativeControlReady(control)}
  function sync(){
    if(disposed)return;const c=current(),key=lease(c);if(key!==bound){bound=key;version++;}section.hidden=!key;if(!key)return;
    const q=plan(c),read=c.completedRead;section.dataset.fromMs=String(q.fromMs);section.dataset.toMs=String(q.toMs);section.dataset.pollFromMs=String(read.fromMs);section.dataset.pollToMs=String(read.toMs);section.dataset.observedAt=read.observedAt;section.dataset.retained=String(c.retained);section.dataset.pollRows=String(read.rowCount);section.dataset.limited=String(read.limited);
    button.title='Stage a 15-minute fixed question ending '+new Date(q.toMs).toISOString()+'. Run explicitly to read telemetry.';
    const text='Event window '+new Date(read.fromMs).toISOString()+' → '+new Date(read.toMs).toISOString()+' · read '+read.observedAt+'. No observations loaded; source coverage is unknown.';if(facts.textContent!==text)facts.textContent=text;
    const revision=version;
    button.onclick=()=>{
      if(!active(button,key,revision))return;try{const question=plan(current()),recordLayout=structuredClone(layout());pause?.();
      // A pause publishes synchronously. Scope edits, replacements and ABA must
      // retire this action even when the selector text later happens to match.
      if(!active(button,key,revision))return;onStage?.(question,recordLayout,button);
      }catch{sync();onError();}
    };
  }
  sync();return {sync,dispose(){disposed=true;version++;section.remove()}};
}
export {plan,lease,mount};
