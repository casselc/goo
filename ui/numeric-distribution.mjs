import './scope-view.mjs';
import {isSnapshot} from './snapshots.mjs';
import {validateSnapshot} from './view-model.mjs';
/* Supplied exact numeric bins; source execution stays in existing population owners. */
import * as Explanation from './explanation.mjs';
import * as NumericBins from './numeric-bins.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
import * as QuestionSession from './question-session.mjs';
import * as UI from './components.mjs';
const global=globalThis;
const protocol=()=>QuestionProtocol;
const {edges,displayBound}=NumericBins;
class Planner{
  constructor({request=(...args)=>global.fetch(...args),canAct=()=>true}={}){this.request=request;this.canAct=canAct;this.generation=0;this.active=true}
  cancel(){this.generation++;this.controller?.abort()}
  invalidate(){this.cancel();this.active=false}
  async plan(body,path='/api/numeric-histograms/selection'){if(!this.active||!this.canAct())return null;this.cancel();const token=this.generation;this.controller=new AbortController();const request=this.request,response=await request(path,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body),signal:this.controller.signal}),value=await response.json();if(!this.active||token!==this.generation||!this.canAct())return null;if(!response.ok)throw Error(value.error||'Numeric selection unavailable');return value}
 }
// Admit planner structure and scope; Go remains the sole range evaluator.
function admitSelectionPlan(plan,parent){
  const same=(a,b)=>QuestionSession.identity(a)===QuestionSession.identity(b);
  if(!same(plan.parent,parent))throw Error('Returned parent question changed.');
  protocol().validateRecordQuestion(parent);
  const configured=edges(parent.numeric?.type),{fromBin,toBin}=plan;
  if(!parent.numericHistogram||plan.binning!==parent.numericHistogram.binning||!Number.isInteger(fromBin)||!Number.isInteger(toBin)||fromBin<0||toBin<=fromBin||toBin>=configured.length||plan.from!==configured[fromBin]||plan.to!==configured[toBin])throw Error('Returned selection differs from the configured bins.');
  const leaf=(op,value)=>({numeric:structuredClone(parent.numeric),op,...(value===undefined?{}:{value:parent.numeric.type==='decimal64'?displayBound(value,parent.numeric):value})});
  const selectedParts=[],restParts=[leaf('missing'),leaf('empty'),leaf('invalid')];
  if(fromBin>0){selectedParts.push(leaf('gte',plan.from));restParts.push(leaf('lt',plan.from))}
  if(toBin<configured.length-1){selectedParts.push(leaf('lt',plan.to));restParts.push(leaf('gte',plan.to))}
  const predicates={selected:selectedParts.length===0?leaf('valid'):selectedParts.length===1?selectedParts[0]:{all:selectedParts},rest:{any:restParts}},children={};
  for(const name of ['selected','rest']){
    const q=plan[name];protocol().validateRecordQuestion(q);
    const generated=predicates[name],expected={...structuredClone(parent),offset:0,where:parent.where?{all:[structuredClone(parent.where),generated]}:generated};
    if(!same(q,expected))throw Error('Returned '+name+' population changed its parent scope or exact selection predicates.');
    children[name]=q;
  }
  for(const [key,keepNumeric,keepHistogram]of [['comparison',false,false],['numericComparison',true,false],['distributionComparison',true,true]]){
    const pair=plan[key];if(!pair){if(key==='comparison')throw Error('Returned count comparison is missing.');continue}
    for(const [side,child]of [['selection','selected'],['baseline','rest']]){const expected=structuredClone(children[child]);if(!keepNumeric)delete expected.numeric;if(!keepHistogram)delete expected.numericHistogram;delete expected.summary;delete expected.spanStatus;expected.groupBy=expected.groupBy?.length?expected.groupBy.slice(0,1):[{kind:'column',key:'service'}];protocol().validateRecordQuestion(pair[side]);if(!same(pair[side],expected))throw Error('Returned '+key+' '+side+' changed its admitted population.')}
  }
  return plan;
}
function mount(host,data,{open,readOnly=false,canAct=()=>true,resolveRange}={}){
  if(!data.question?.numericHistogram)return false;const node=UI.node,snapshot=data.histogramDataset,section=node('section',undefined,'numeric-distribution');section.setAttribute('aria-label','Numeric value distribution');host.append(section);
  if(!isSnapshot(snapshot)||snapshot.shape!=='histogram'||!validateSnapshot(snapshot).ok){section.append(node('p','Numeric distribution unavailable: captured bins or read consistency could not be verified.'));return false}
  const chart=node('goo-chart');chart.snapshot=snapshot;chart.display={kind:'histogram',title:'Numeric value distribution'};section.append(chart);
  if(readOnly||!open||!resolveRange)return true;
  section.append(node('small','Drag to select. Keyboard: Space, Shift + arrows, Enter; Escape cancels.'));
  const tray=node('div',undefined,'numeric-distribution-selection'),status=node('p'),run=node('button','Inspect selected population'),compare=node('button','Stage count comparison with rest'),numeric=node('button','Stage numeric comparison with rest');status.setAttribute('role','status');tray.hidden=true;tray.append(status,run,compare,numeric);section.append(tray);let chosen=null,revision=0,withdrawn=false,explanationOwner=null;const explanationHost=node('div');section.append(explanationHost);const allowed=()=>!withdrawn&&canAct()&&UI.actionVisible(section),planner=new Planner({canAct:allowed});
  function choose(intent){if(!allowed())return;const resolved=resolveRange(intent);if(!resolved?.ok)return;explanationOwner?.dispose?.();explanationHost.replaceChildren();const {fromBin,toBin,from,to}=resolved;chosen={fromBin,toBin,from,to,count:String(snapshot.rows.slice(fromBin,toBin).reduce((n,row)=>n+BigInt(row.values.count),0n)),intent:Object.freeze({...intent}),question:resolved.question};revision++;planner.cancel();tray.hidden=false;run.disabled=compare.disabled=numeric.disabled=false;status.textContent=chosen.count+' of '+data.count+' records selected · ['+displayBound(chosen.from,data.question.numeric)+', '+displayBound(chosen.to,data.question.numeric)+') '+data.question.numeric.unit+' · rest includes missing, empty and invalid values.';const range={...chosen};explanationOwner=Explanation?.mount(explanationHost,{title:'Explain this selection',signal:data.question.signal,available:data.fields||[],canAct:allowed,resolve:async signal=>{const checked=resolveRange(range.intent);if(!checked?.ok)throw Error('Selection paused. Run again before selecting.');const response=await fetch('/api/numeric-histograms/selection',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({parent:checked.question,fromBin:checked.fromBin,toBin:checked.toBin}),signal}),plan=await response.json();if(!response.ok)throw Error(plan.error||'Selection planning unavailable');if(!resolveRange(range.intent)?.ok)throw Error('Selection paused. Run again before selecting.');return {version:'goo/explanation/v1',baseline:plan.rest,selection:plan.selected}}})}
  chart.addEventListener('goo-histogram-range',e=>choose(e.detail));
  async function proceed(comparing){const opener=comparing==='numeric'?numeric:comparing?compare:run;if(opener.disabled||!UI.actionVisible(opener)||opener.closest('[inert]')||!chosen||!allowed())return;const expected=revision,range={...chosen};run.disabled=compare.disabled=numeric.disabled=true;status.textContent='Planning the selected range; no telemetry read yet…';try{const checked=resolveRange(range.intent);if(!checked?.ok)return;const plan=await planner.plan({parent:checked.question,fromBin:checked.fromBin,toBin:checked.toBin});if(!plan||expected!==revision||!allowed()||!UI.actionVisible(opener)||opener.closest('[inert]')||!resolveRange(range.intent)?.ok)return;if(plan.fromBin!==range.fromBin||plan.toBin!==range.toBin||plan.from!==range.from||plan.to!==range.to)throw Error('Returned selection differs from the staged bins.');admitSelectionPlan(plan,range.question);for(const q of [plan.selected,plan.rest,plan.comparison?.selection,plan.comparison?.baseline])protocol().validateLink({version:1,question:q,view:'Records'});if(comparing==='numeric'){if(!plan.numericComparison)throw Error('Numeric comparison planning is unavailable.');for(const q of [plan.numericComparison.selection,plan.numericComparison.baseline])protocol().validateLink({version:1,question:q,view:'Records'});open(plan.numericComparison.selection,null,plan.numericComparison.baseline,numeric)}else if(comparing)open(plan.comparison.selection,null,plan.comparison.baseline,compare);else open(plan.selected,undefined,undefined,run)}catch(e){if(expected===revision&&allowed()&&e.name!=='AbortError')status.textContent=e.message}finally{if(expected===revision&&allowed())run.disabled=compare.disabled=numeric.disabled=false}}
  run.onclick=()=>proceed(false);compare.onclick=()=>proceed(true);numeric.onclick=()=>proceed('numeric');return {suspend(){explanationOwner?.invalidate?.();revision++;planner.cancel();run.disabled=compare.disabled=numeric.disabled=true;if(chosen)status.textContent='Selection paused. Run the fixed question explicitly before selecting again; no telemetry was read.'},invalidate(){explanationOwner?.dispose?.();withdrawn=true;revision++;planner.invalidate();run.disabled=compare.disabled=numeric.disabled=true;for(const b of chart.querySelectorAll('button'))b.disabled=true}};
 }

export {displayBound,edges,mount,Planner,admitSelectionPlan};
