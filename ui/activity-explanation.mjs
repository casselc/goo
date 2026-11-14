import {checkedRecord} from './saved-capture-protocol.mjs';
import * as RecordEnvironment from './record-environment.mjs';
import {identity as questionIdentity} from './question-session.mjs';
import {bandPredicate,classify as severityClass} from './log-severity.mjs';
import {object as closedObject} from './intent-shape.mjs';
import './time-bars.mjs';
/* Activity-owned fixed scopes; held occurrences are a brush guide, never totals. */
import * as Attributes from './attributes.mjs';
import * as Explanation from './explanation.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
import * as UI from './components.mjs';
const global=globalThis;
const copy=x=>structuredClone(x),protocol=()=>QuestionProtocol||(null);
function checked(q){protocol().validateRecordQuestion(q);return q}
function windowQuestion(c){const s=c?.selector,w=c?.readWindow;if(!s||!['logs','spans'].includes(s.signal||'logs')||!w||!Number.isSafeInteger(w.fromMs)||!Number.isSafeInteger(w.toMs)||w.fromMs<0||w.toMs-w.fromMs!==60000)throw Error('Wait for an applied scope and completed 60-second window.');return checked({...copy(s),signal:s.signal||'logs',...copy(w),limit:100,offset:0})}
function previousQuestion(q){const width=q.toMs-q.fromMs;if(!Number.isSafeInteger(width)||width<=0||q.fromMs<width)throw Error('The preceding equal-width window is unavailable.');return checked({...copy(q),fromMs:q.fromMs-width,toMs:q.fromMs})}
function selectWindow(parent,fromMs,toMs){if(!Number.isSafeInteger(fromMs)||!Number.isSafeInteger(toMs)||fromMs<parent.fromMs||toMs>parent.toMs||fromMs>=toMs)throw Error('Choose a nonempty range inside the held window.');return checked({...copy(parent),fromMs,toMs})}
function identity(q){return questionIdentity(q)}
// Only an exact supplied aggregate window and a closed numeric-severity
// refinement may be pinned. Arbitrary replacement selectors remain forbidden.
function severityQuestion(question,band){
 if(band===undefined)return question;
 if(question.signal!=='logs')throw Error('Numeric severity requires a log parent.');
 const predicate=bandPredicate(band);
 return checked({...question,where:question.where?{all:[question.where,predicate]}:predicate});
}
function pinnedWindow(context,value){
 closedObject(value,['parent','selection','logSeverityBand'],'Choose an exact parent and interval.');
 if(!Object.hasOwn(value,'parent')||!Object.hasOwn(value,'selection')||Object.hasOwn(value,'logSeverityBand')&&value.logSeverityBand===undefined)throw Error('Choose an exact parent and interval.');
 const band=Object.hasOwn(value,'logSeverityBand')?value.logSeverityBand:undefined,live=windowQuestion(context),parent=checked(copy(value.parent)),selection=checked(copy(value.selection));
 const expected=severityQuestion(windowQuestion({...context,readWindow:{fromMs:parent.fromMs,toMs:parent.toMs}}),band);
 if(parent.toMs>live.toMs||identity(parent)!==identity(expected))throw Error('The aggregate window or selector differs from this Activity.');
 const narrowed=selectWindow(expected,selection.fromMs,selection.toMs);
 if(identity(selection)!==identity(narrowed))throw Error('The selected interval changes the aggregate selector.');
 return {parent:expected,selection:narrowed,...(band!==undefined?{logSeverityBand:band}:{})};
}
function sameParent(context,parent,pinned,band){try{if(band!==undefined&&!pinned)return false;const expected=windowQuestion(pinned?{...context,readWindow:{fromMs:parent.fromMs,toMs:parent.toMs}}:context);return identity(severityQuestion(expected,band))===identity(parent)}catch{return false}}
function heldBins(context,q,band){
 if(band!==undefined){if(q.signal!=='logs')throw Error('Numeric severity requires a log parent.');bandPredicate(band)} // Validate even for an empty buffer.
 const counts=Array(6).fill(0);
 for(const row of context.rows||[]){
  if(band!==undefined){try{if(severityClass(row.record?.SeverityNumber)!==band)continue}catch{continue}}
  const raw=row.record?.timestamp_ns;if(typeof raw!=='string'||!/^\d+$/.test(raw))continue;const ns=BigInt(raw),start=BigInt(q.fromMs)*1000000n,end=BigInt(q.toMs)*1000000n;if(ns>=start&&ns<end)counts[Number((ns-start)/10000000000n)]++;
 }
 return {title:'Held '+(band===undefined?'observations':band+' observations')+' · sampled occurrences, not full population counts',unit:'held occurrences',interactive:true,ceiling:String(Math.max(1,...counts)),points:counts.map((value,i)=>({key:String(i),label:new Date(q.fromMs+i*10000).toISOString()+' · '+value+' held occurrences',value:String(value)})),fromLabel:new Date(q.fromMs).toISOString(),toLabel:new Date(q.toMs).toISOString()};
}

function savedScope(page,id,selected){const q=page?.receipt?.question;if(!/^[a-f0-9]{64}$/.test(id)||page?.receipt?.id!==id||q?.kind!=='explore'||!q.explore||!Number.isFinite(Date.parse(page.receipt.capturedAt)))throw Error('Load an exact saved explore capture.');const baseline=checked(copy(q.explore));if(!page.explore)throw Error('Stored result differs from its receipt scope.');checkedRecord(baseline,page.explore.question);if(baseline.signal!==selected.signal||!RecordEnvironment.sameQuestionEnvironment(baseline,selected))throw Error('Baseline scope must use the same signal and environment policy.');return baseline}
function mount(host,{current,pauseStream}={}){
  const n=UI.node,panel=n('section',undefined,'activity-explanation'),heading=n('h3','Explain Activity'),close=n('button','Close explanation'),note=n('p'),chart=n('goo-time-bars'),bounds=n('p'),mode=n('select'),captureID=n('input'),load=n('button','Load baseline scope'),provenance=n('p'),status=n('p'),body=n('div');panel.hidden=true;panel.setAttribute('aria-label','Activity explanation');mode.setAttribute('aria-label','Explanation baseline');for(const [value,label]of [['previous','Previous equal-width window'],['saved','Saved capture scope']]){const o=n('option',label);o.value=value;mode.append(o)}captureID.setAttribute('aria-label','Explanation baseline capture ID');captureID.maxLength=64;status.setAttribute('role','status');chart.setAttribute('range-select','');note.textContent='Held observations guide time selection only. Run rereads both fixed populations; this is not an explanation of the held sample.';panel.append(heading,close,note,chart,bounds,mode,captureID,load,provenance,status,body);host.append(panel);
  let parent=null,selected=null,baseline=null,owner=null,opened=false,disposed=false,generation=0,controller=null,suspended=false,pinned=false,logSeverityBand=undefined,entryVersion=0,opener=null;
  function active(){const c=current();const same=!!parent&&sameParent(c,parent,pinned,logSeverityBand);return !disposed&&opened&&!panel.hidden&&c?.active&&c.clean&&!!parent&&same&&!['loading','reconnecting','stopped'].includes(c.phase)}
  function clear(){generation++;controller?.abort();controller=null;load.disabled=false;owner?.dispose();owner=null;body.replaceChildren();status.textContent='';}
  function render(){clear();captureID.hidden=load.hidden=mode.value!=='saved';bounds.textContent='Selection ['+new Date(selected.fromMs).toISOString()+', '+new Date(selected.toMs).toISOString()+')';if(mode.value==='saved'&&!baseline){status.textContent='Load a saved baseline scope explicitly. No saved counts will be reused.';return}try{const b=baseline||previousQuestion(selected),available=Attributes?.suggestionData?.((current().rows||[]).map(x=>x.record));owner=Explanation.mount(body,{title:'Explanation fields and result',signal:selected.signal,request:{version:'goo/explanation/v1',baseline:copy(b),selection:copy(selected)},available:(available?.fields||[]).map(item=>item.field||item),canAct:active,note:mode.value==='saved'?'Saved baseline scope is reread; saved counts are not reused.':'Selection versus the adjacent preceding equal-width window.'});owner.open()}catch(e){status.textContent=e.message}}
  function invalidate(){entryVersion++;opener=null;clear();parent=selected=baseline=null;opened=false;suspended=false;pinned=false;logSeverityBand=undefined;delete panel.dataset.severityBand;panel.hidden=true;provenance.replaceChildren()}
  function sync(){if(!opened||!parent)return;const c=current();const same=sameParent(c,parent,pinned,logSeverityBand);if(!c?.clean||!same||['loading','reconnecting','stopped'].includes(c?.phase)){invalidate();return}if(!c.active){if(!suspended){owner?.invalidate();generation++;controller?.abort();controller=null;load.disabled=false;suspended=true}panel.hidden=true}else{panel.hidden=false;suspended=false}}
  function choose(a,b){if(!active())return;try{selected=selectWindow(parent,parent.fromMs+Number(a)*10000,parent.fromMs+(Number(b)+1)*10000);render()}catch(e){status.textContent=e.message}}
  chart.addEventListener('goo-time-range',e=>choose(e.detail.fromKey,e.detail.toKey));chart.addEventListener('goo-time-select',e=>choose(e.detail.key,e.detail.key));mode.onchange=()=>{baseline=null;provenance.replaceChildren();render()};captureID.oninput=()=>{baseline=null;provenance.replaceChildren();render()};
  load.onclick=async()=>{if(!active())return;clear();const id=captureID.value,g=++generation;controller=new AbortController();load.disabled=true;try{if(!/^[a-f0-9]{64}$/.test(id))throw Error('Enter an exact capture ID.');const response=await fetch('/api/captures/'+id,{signal:controller.signal}),page=await response.json();if(g!==generation||!active())return;if(!response.ok)throw Error(page.error||'Saved scope unavailable.');baseline=savedScope(page,id,selected);const link=n('a',id);link.href='#capture='+id;provenance.replaceChildren(n('span','Baseline scope from '),link,n('span','; explanation rereads both fixed windows. Saved counts are not reused and late arrivals may change them. Custody '+page.receipt.capturedAt+' · original source read '+(page.explore?.capturedAt||'unavailable')));render()}catch(e){if(g===generation&&e.name!=='AbortError')status.textContent=e.message}finally{if(g===generation)load.disabled=false}};
  close.onclick=invalidate;return {open(value,origin){const c=current();if(!c?.active||!c.clean||['loading','reconnecting','stopped'].includes(c.phase))return false;try{const anchored=value===undefined?null:pinnedWindow(c,value),q=anchored?.parent||windowQuestion(c),bins=heldBins(c,q,anchored?.logSeverityBand);invalidate();parent=copy(q);selected=copy(anchored?.selection||q);pinned=!!anchored;logSeverityBand=anchored?.logSeverityBand;panel.dataset.severityBand=logSeverityBand||'all';opened=true;panel.hidden=false;panel.dataset.parentFromMs=String(parent.fromMs);panel.dataset.parentToMs=String(parent.toMs);panel.dataset.selectedFromMs=String(selected.fromMs);panel.dataset.selectedToMs=String(selected.toMs);mode.value='previous';captureID.value='';chart.data=bins;const opening=entryVersion;pauseStream?.();if(opening!==entryVersion||!active()){invalidate();return false}render();opener=anchored&&UI.actionVisible(origin)?origin:null;const entry=entryVersion;close.onclick=()=>{if(entry!==entryVersion||!opened||panel.hidden||!UI.actionVisible(close))return;const origin=opener,restore=active()&&UI.actionVisible(origin);invalidate();if(restore)origin.focus({preventScroll:true})};if(anchored){heading.tabIndex=-1;heading.scrollIntoView({block:'nearest',inline:'nearest'});heading.focus({preventScroll:true})}return true}catch(e){status.textContent=e.message;return false}},sync,invalidate,dispose(){invalidate();disposed=true;panel.remove()}};
 }

export {windowQuestion,previousQuestion,selectWindow,pinnedWindow,sameParent,heldBins,savedScope,mount};
