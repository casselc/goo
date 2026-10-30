/* Stored record-page owner. No query editor, session, tail or source execution on show. */
import * as DOM from './namespace.mjs';
import * as Deployments from './deployments.mjs';
import * as Distributions from './distributions.mjs';
import * as EvidenceScope from './evidence-scope.mjs';
import * as NumericPlot from './numeric-plot.mjs';
import * as Environment from './record-environment.mjs';
import * as SourceSnapshots from './source-snapshots.mjs';
import * as Queries from './queries.mjs';
import * as QuestionProtocol from './question-protocol.mjs';
import * as RecordResults from './record-results.mjs';
import * as TraceJourney from './trace-journey.mjs';
import * as UI from './components.mjs';
const global=globalThis;
const node=(...args)=>UI.node(...args);
function mount(container,{prefix='saved',onPage,summaryProvided=true,active=()=>true,onOpenQuestion}={}){
  let page=null,revision=0;let plot=null;
  container.classList.add('saved-record-evidence');
  container.innerHTML='<goo-result-state id="exploreStatus" state="unknown"></goo-result-state><div class="controls question-result-actions"><button id="explorePrevious" disabled>Previous saved page</button><button id="exploreNext" disabled>Next saved page</button><button id="exploreCreatePackage" disabled>Save reusable question</button></div><div id="savedExamples"></div><div id="exploreResultViews"></div><div id="recordsPanel" role="tabpanel" aria-labelledby="resultRecords"><div id="exploreGroups"></div><div id="exploreRows"></div></div><details><summary>Exploration question, source plan and evidence basis</summary><pre id="exploreExplain"></pre></details>';
  DOM?.namespace(container,prefix);const el=id=>document.getElementById(prefix+id);
  const records=RecordResults.mount(container,{prefix,readOnly:true,active,attributeOptions:(target,signal,options,origin)=>{const observed=revision;return target===el('exploreRows')?{canPlot:()=>active()&&observed===revision&&origin===page?.explore.question,canRefine:()=>active()&&observed===revision&&origin===page?.explore.question,onPlot:(field,button)=>{if(active()&&observed===revision&&origin===page?.explore.question)plot.open(field,button)}}:{}},savedPage:()=>page,returnFocus:()=>el('explorePrevious').focus(),fetchQuery:async(q,signal)=>{const requested=structuredClone(q),r=await Queries.postJSON('/api/explore',requested,signal),data=r.data;if(!r.ok)throw Error(data.error);return SourceSnapshots.adoptExploration(data,requested)}});
  plot=NumericPlot.mount(container,{current:()=>({question:page?.explore.question,active:active(),clean:!!page&&(page.explore.question.environmentSelector==null||!!onOpenQuestion),provenance:'Stages a fresh read of this captured scope; saved evidence is unchanged.'}),onStage:q=>{if(q.environmentSelector!=null){onOpenQuestion?.(structuredClone(q));return}const link=node('a','Open staged value distribution');link.href=QuestionProtocol.questionLink(q);link.title='Run explicitly to read fresh telemetry; saved evidence is unchanged.';container.append(link);link.click();link.remove()}});
  const distributions=Distributions.mount(el('exploreResultViews'),{prefix,readOnly:true,canRefine:()=>false,openTrace:(id,opener)=>{if(page)records.openTrace(page.explore.question,id,'',page.explore.basis,opener)}});
  const deploymentHost=node('div');el('timeChart').after(deploymentHost);const deployments=Deployments.mount(deploymentHost,{prefix});
  const scope=summaryProvided?null:node('div');if(scope)el('exploreStatus').before(scope);
  el('explorePrevious').onclick=()=>{if(active()&&page?.previousCursor){revision++;plot.invalidate();onPage?.(page.previousCursor)}};
  el('exploreNext').onclick=()=>{if(active()&&page?.nextCursor){revision++;plot.invalidate();onPage?.(page.nextCursor)}};
  el('exploreCreatePackage').onclick=()=>{if(active()&&page&&page.explore.question.environmentSelector==null)global.dispatchEvent(new CustomEvent('goo:draft-telemetry-package',{detail:{question:structuredClone(page.explore.question),lookbackMinutes:Number(document.getElementById('minutes').value)}}))};
  function showSaved(value){value=SourceSnapshots.adoptStoredExplore(value);revision++;plot.invalidate();records.hide();page=value;const data=page.explore,q=data.question;container.dataset.questionPhase='frozen';if(scope)scope.replaceChildren(EvidenceScope.element(EvidenceScope.record(page)));deployments.setQuestion(q.environmentSelector!=null?null:q);distributions.invalidate({scopeChanged:true});distributions.render(data);el('savedExamples').hidden=q.signal!=='spans';TraceJourney.examples(el('savedExamples'),data,{openTrace:(id,anchor,opener)=>records.openTrace(q,id,anchor,data.basis,opener)});records.table(el('exploreRows'),data.rows,q.signal,q,data.basis);records.groups(el('exploreGroups'),data);el('exploreCreatePackage').disabled=q.environmentSelector!=null;el('exploreCreatePackage').title=q.environmentSelector!=null?'Reusable selector questions are not supported yet.':'Save this question for fresh reuse.';el('exploreStatus').state=data.state==='changed_between_reads'?'partial':data.state==='no_data'?'empty':'unknown';el('exploreStatus').textContent=`Saved evidence · ${data.count} matching captured ${q.signal} · ${data.rows.length} displayed · offset ${q.offset} · ${page.retainedRecords} stored inspection records${page.recordsLimited?' · more source records were not retained; narrow the question for further inspection':''}. Paging reads this immutable capture. ${data.state==='changed_between_reads'?'Original source reads disagreed; preserved as captured.':'Source completeness remains unknown.'}`;el('explorePrevious').disabled=!page.previousCursor;el('exploreNext').disabled=!page.nextCursor;el('exploreExplain').textContent=JSON.stringify({question:q,basis:data.basis,plan:data.plan,numeric:data.numeric,numericHistogram:data.numericHistogram,duration:data.duration,spanStatus:data.spanStatus,groups:data.groups,timeSeries:data.timeSeries,groupedTimeSeries:data.groupedTimeSeries,warnings:data.warnings,receipt:page.receipt,retainedRecords:page.retainedRecords,recordsLimited:page.recordsLimited},null,2)}
  return {showSaved,suspend(){revision++;plot.invalidate();records.hide();distributions.suspend()},dispose(){plot.dispose();records.dispose();distributions.suspend();deployments.invalidate();container.replaceChildren()}};
 }
export {mount};
