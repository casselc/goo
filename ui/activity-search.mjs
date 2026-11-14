/* Explicit local Find -> fixed source-message question; never a source reader. */
import * as UI from './components.mjs';
import * as Environment from './record-environment.mjs';
import * as Shape from './intent-shape.mjs';
import {identity} from './question-session.mjs';
import {validateRecordQuestion} from './record-intent.mjs';
import {findQuery} from './record-presentation.mjs';
import {nativeControlReady} from './dom-visibility.mjs';
function checked(context){
 if(context?.active!==true||context.clean!==true||context.staged===true||!['unknown','partial','paused'].includes(context.phase))throw Error('Use the applied Activity log scope and a completed source window.');
 Shape.object(context.selector,['signal','environment','environmentSelector','where']);
 if((context.selector.signal??'logs')!=='logs')throw Error('Message search requires logs.');
 const find=context.find;findQuery(find?.text);
 if(!find.text.length||!Number.isSafeInteger(find.revision)||find.revision<0)throw Error('Enter a literal message to search.');
 const read=context.completedRead;
 if(!read||!Number.isSafeInteger(read.fromMs)||!Number.isSafeInteger(read.toMs)||read.toMs-read.fromMs!==60000||!Number.isSafeInteger(read.revision)||read.revision<1||typeof read.limited!=='boolean'||!Number.isSafeInteger(read.rowCount)||read.rowCount<0||typeof read.observedAt!=='string'||!Number.isFinite(Date.parse(read.observedAt))||Date.parse(read.observedAt)<read.toMs)throw Error('Wait for a completed source window.');
 if(context.readWindow?.fromMs!==read.fromMs||context.readWindow?.toMs!==read.toMs||!Number.isSafeInteger(context.epoch)||context.epoch<0||typeof context.connection!=='string'||!/^[a-f0-9]{64}$/.test(context.connection))throw Error('Activity observation ownership is unavailable.');
 const predicate={field:{kind:'column',key:'body'},op:'contains',value:find.text};
 const query={signal:'logs',...Environment.questionFields(context.selector),fromMs:read.fromMs,toMs:read.toMs,limit:100,offset:0,where:context.selector.where?{all:[structuredClone(context.selector.where),predicate]}:predicate};
 validateRecordQuestion(query);return {query,read,find};
}
function plan(context){return checked(context).query}
function lease(context){try{const {query,read,find}=checked(context);return JSON.stringify([context.epoch,context.connection,read.revision,read.observedAt,find.revision,identity(query)]);}catch{return null}}
function mount(host,{current,pause,layout=()=>null,onStage,onError=()=>{}}={}){
 const n=UI.node,section=n('div',undefined,'activity-source-search'),button=n('button','Search log messages'),facts=n('span');button.type='button';section.hidden=true;section.append(button,facts);host.append(section);
 let disposed=false,version=0,bound=null;
 function active(key,revision){return !disposed&&version===revision&&bound===key&&!!key&&lease(current())===key&&nativeControlReady(button)}
 function sync(){
  if(disposed)return;const context=current(),key=lease(context);if(key!==bound){bound=key;version++;}section.hidden=!key;if(!key)return;
  const q=plan(context);section.dataset.fromMs=String(q.fromMs);section.dataset.toMs=String(q.toMs);button.title='Stage an exact, case-sensitive literal search of log messages in this fixed event-time window. Run explicitly to read the source.';
  const text='Message only · case-sensitive · '+new Date(q.fromMs).toISOString()+' → '+new Date(q.toMs).toISOString();if(facts.textContent!==text)facts.textContent=text;
  const revision=version;button.onclick=()=>{if(!active(key,revision))return;try{const question=plan(current()),recordLayout=structuredClone(layout());pause?.();if(!active(key,revision))return;onStage?.(question,recordLayout,button);}catch{sync();onError();}};
 }
 sync();return {sync,dispose(){disposed=true;version++;section.remove()}};
}
export {plan,lease,mount};
