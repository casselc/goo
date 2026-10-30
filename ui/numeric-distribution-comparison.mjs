import {PendingScope} from './request-owner.mjs';
/* App adapter for supplied aligned bins and explicit checked side continuations. */
import './aligned-histogram.mjs';
import * as NumericDistribution from './numeric-distribution.mjs';
import {asHistogramComparison} from './snapshots.mjs';
import {histogramIntent} from './view-model.mjs';
import {createHistogramComparisonActions} from './histogram-comparison-actions.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
import * as QuestionSession from './question-session.mjs';
import * as UI from './components.mjs';
const global=globalThis;

function checkedSelected(parent,selected){
  QuestionProtocol.validateRecordQuestion(parent);QuestionProtocol.validateRecordQuestion(selected);
  const shape=q=>{const out=structuredClone(q);delete out.where;delete out.offset;return QuestionSession.identity(out)};
  if(shape(parent)!==shape(selected))throw Error('Returned selection changed the retained side scope or measures.');
  if(parent.where&&(!Array.isArray(selected.where?.all)||selected.where.all.length!==2||QuestionSession.identity({where:parent.where})!==QuestionSession.identity({where:selected.where.all[0]})))throw Error('Returned selection dropped the parent predicates.');
  return selected;
}
function rangeCapability(questions){if(!Array.isArray(questions)||questions.length!==2||questions.some(q=>!q))return {allowed:false,reason:'Exact side questions are unavailable for range inspection.'};return questions.some(q=>(q.groupBy||[]).length>1)?{allowed:false,reason:'Range inspection supports zero or one breakdown; open a side question to edit grouping.'}:{allowed:true}}
// The deferred receipt load must reproduce the population shown by this mount.
// A matching question alone does not establish matching retained observations.
function matchesOperand(snapshot,cohort){
 try{
  if(!cohort||QuestionSession.identity(cohort.question)!==QuestionSession.identity(snapshot.scope.selector))return false;
  if(snapshot.sourceBasis!==undefined&&cohort.basis!==snapshot.sourceBasis)return false;
  if(snapshot.capturedAt!==undefined&&cohort.capturedAt!==snapshot.capturedAt)return false;
  const support=snapshot.histogram.support,numeric=cohort.numeric,binding=snapshot.histogram.binding;
  if(cohort.state==='changed_between_reads'||cohort.count!==support.total||!numeric||numeric.count!==support.total||numeric.unit!==binding.unit||(numeric.scale??0)!==(binding.scale??0))return false;
  if(['valid','missing','empty','invalid'].some(key=>numeric[key]!==support[key]))return false;
  const state=support.valid!=='0'?'observed':support.total==='0'?'no_data':'no_valid_values';
  if(numeric.state!==state)return false;
  const bins=cohort.numericHistogram?.bins,edges=NumericDistribution.edges(binding.type);
  return cohort.numericHistogram?.binning==='powers-of-two'&&Array.isArray(bins)&&bins.length===snapshot.rows.length&&bins.every((bin,index)=>bin.from===edges[index]&&bin.to===edges[index+1]&&bin.count===snapshot.rows[index].values.count);
 }catch{return false}
}
function mount(host,envelope,{comparison,loadComparison,open,canAct=()=>true,readOnly=false,route='explore'}={}){
  if(!envelope)return null;let captured;try{captured=asHistogramComparison(envelope)}catch{return null}const n=UI.node,section=n('section',undefined,'numeric-distribution-comparison'),label=n('label','Distribution scale'),mode=n('select'),plot=n('goo-aligned-histogram'),status=n('p'),tray=n('div',undefined,'controls'),baseline=n('button','Open baseline range'),selection=n('button','Open comparison range');mode.setAttribute('aria-label','Distribution scale');for(const [value,text]of [['count','Record count'],['population-share','Share of all captured records'],['valid-share','Share of valid numeric records']]){const option=n('option',text);option.value=value;mode.append(option)}label.append(mode);status.setAttribute('role','status');tray.append(baseline,selection);tray.hidden=true;section.append(label,plot,status,tray);host.append(section);const sideQuestions=()=>[captured.baseline.scope.selector,captured.selection.scope.selector],capability=rangeCapability(sideQuestions());if(!capability.allowed)section.append(n('p',capability.reason));let chosen=null,revision=0,withdrawn=false,loadController,pendingScope;const allowed=()=>!readOnly&&!!open&&!withdrawn&&section.isConnected&&section.getClientRects().length>0&&canAct()&&rangeCapability(sideQuestions()).allowed,planner=new NumericDistribution.Planner({canAct:allowed});
  const onWorkspace=event=>{if(event.detail.route!==route){cancel();chosen=null;tray.hidden=true}},onDisplay=()=>{cancel();chosen=null;tray.hidden=true};const onToggle=event=>{if(event.target?.localName==='details'&&!event.target.open&&event.target.contains(section))onDisplay()};
  function requestScope(){return new PendingScope({events:[{target:global,type:'goo:workspace',callback:onWorkspace},...['goo:show-document','goo:show-general-capture','goo:show-retained-evidence','goo:show-operation-capture'].map(type=>({target:global,type,callback:onDisplay})),{target:global.document,type:'toggle',callback:onToggle,capture:true}],observe:onRemove=>{const observer=new MutationObserver(()=>{if(!section.isConnected)onRemove()});observer.observe(global.document.body,{childList:true,subtree:true});return ()=>observer.disconnect()},onRemove:cancel}).start()}

  const actions=createHistogramComparisonActions({comparison:captured,active:allowed});
  function update(){plot.comparison=captured;plot.display={scale:mode.value}}
  function cancel(){revision++;pendingScope?.stop();pendingScope=null;planner.cancel();loadController?.abort();baseline.disabled=selection.disabled=true}
  mode.onchange=()=>{cancel();chosen=null;tray.hidden=true;status.textContent='Display changed locally; range selection cleared.';update()};
  plot.addEventListener('goo-aligned-range',event=>{if(!allowed())return;const resolved=actions.resolve(event.detail);if(!resolved.ok)return;cancel();const {fromBin,toBin,from,to}=resolved,intents=Object.fromEntries(['baseline','selection'].map(side=>[side,Object.freeze(histogramIntent(captured[side],event.detail.intent.fromKey,event.detail.intent.toKey))]));chosen={fromBin,toBin,from,to,intents,counts:['baseline','selection'].map(side=>String(captured[side].rows.slice(fromBin,toBin).reduce((sum,row)=>sum+BigInt(row.values.count),0n)))};const binding=captured.baseline.histogram.binding;status.textContent='['+NumericDistribution.displayBound(from,binding)+', '+NumericDistribution.displayBound(to,binding)+') '+binding.unit+' · baseline '+chosen.counts[0]+' records · comparison '+chosen.counts[1]+' records. Open a side explicitly to read its fixed population.';baseline.disabled=chosen.counts[0]==='0';selection.disabled=chosen.counts[1]==='0';tray.hidden=false});
  async function proceed(side){if(!chosen||!allowed()||chosen.counts[side==='baseline'?0:1]==='0')return;const range={...chosen},effect=Object.freeze({side,intent:chosen.intents[side]});if(!actions.resolve(effect).ok)return;cancel();const expected=revision,scope=requestScope();pendingScope=scope;baseline.disabled=selection.disabled=true;loadController=new AbortController();try{const request=loadComparison?await loadComparison(loadController.signal):typeof comparison==='function'?comparison():comparison;if(expected!==revision||!allowed()||!actions.resolve(effect).ok)return;for(const name of ['baseline','selection'])if(!matchesOperand(captured[name],request?.[name]))throw Error('Loaded comparison differs from the retained side observations.');const plan=await planner.plan({comparison:request,side,fromBin:range.fromBin,toBin:range.toBin},'/api/numeric-histograms/comparison-selection');if(!plan||expected!==revision||!allowed()||!actions.resolve(effect).ok)return;if(plan.fromBin!==range.fromBin||plan.toBin!==range.toBin||plan.from!==range.from||plan.to!==range.to)throw Error('Returned range differs from the staged bins.');const parent=captured[side].scope.selector;if(QuestionSession.identity(plan.parent)!==QuestionSession.identity(parent))throw Error('Returned side question changed.');open(checkedSelected(parent,plan.selected))}catch(e){if(expected===revision&&allowed()&&e.name!=='AbortError')status.textContent=e.message}finally{scope.stop();if(pendingScope===scope)pendingScope=null;if(expected===revision&&allowed()){baseline.disabled=chosen.counts[0]==='0';selection.disabled=chosen.counts[1]==='0'}}}
  baseline.onclick=()=>proceed('baseline');selection.onclick=()=>proceed('selection');update();return {suspend(){cancel();chosen=null;tray.hidden=true},invalidate(){cancel();withdrawn=true;actions.revoke();chosen=null;tray.hidden=true}};
 }

export {checkedSelected,mount,rangeCapability,PendingScope,matchesOperand};
