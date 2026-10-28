import * as ActivityViews from './activity-views.mjs';
import * as Captures from './captures.mjs';
import * as Documents from './documents.mjs';
import * as ExampleQuestions from './example-questions.mjs';
import * as Fleet from './fleet.mjs';
import * as GuidedMeasurements from './guided-measurements.mjs';
import * as Histograms from './histograms.mjs';
import * as HostBrowser from './host-browser.mjs';
import * as Inventory from './inventory.mjs';
import * as InvestigationTrail from './investigation-trail.mjs';
import * as Investigations from './investigations.mjs';
import * as KubernetesBrowser from './kubernetes-browser.mjs';
import * as Meaning from './meaning.mjs';
import * as Measurements from './measurements.mjs';
import * as NativeRecordNavigation from './native-record-navigation.mjs';
import * as Operations from './operations.mjs';
import * as Ratios from './ratios.mjs';
import * as Queries from './queries.mjs';
import * as ResourceTabs from './resource-tabs.mjs';
import * as Retention from './retention.mjs';
import * as SavedWorkspace from './saved-workspace.mjs';
import * as StreamEntry from './stream-entry.mjs';
import * as TelemetryPackages from './telemetry-packages.mjs';
import * as Visuals from './visualizations.mjs';
import * as Workflows from './workflows.mjs';
import * as Workspace from './workspace.mjs';
const el=id=>document.getElementById(id);
let selected='';
let spanFilter;
let controlRevision=0,queryGeneration=0;
for(const id of ['view','group','minutes','scenario'])for(const event of ['input','change'])el(id).addEventListener(event,()=>controlRevision++);
const initial={view:'completion',minutes:30,groupBy:'release',reservationId:'',scenario:''};
function controls(){return{view:el('view').value,minutes:Number(el('minutes').value),groupBy:el('group').value,reservationId:selected,scenario:el('scenario').value,...(spanFilter&&el('view').value==='traces'?{spanFilter}: {})}}
function textNode(tag,value){const n=document.createElement(tag);n.textContent=value;return n}
async function configurationJSON(url){const response=await fetch(url),data=await response.json();if(!response.ok)throw Error(data.error||'Configuration unavailable');return data}
const uiConfiguration=configurationJSON('/api/ui-config').then(config=>{
 if(config.version!=='goo/ui-config/v1'||!['lab','generic'].includes(config.profile)||!['configured','unconfigured'].includes(config.examples?.state))throw Error('Invalid UI configuration');
 document.documentElement.dataset.profile=config.profile;document.documentElement.dataset.examples=config.examples.state;
 const examples=config.examples.state==='configured';el('sampleApplicationLink').hidden=!examples;document.querySelector('[data-workspace=examples]').hidden=!examples;el('legacyExamples').hidden=!examples;
 for(const link of document.querySelectorAll('.operation-sample-links'))link.hidden=!examples;
 if(!examples){el('exampleQuestions').append(textNode('h2','Sample applications not configured'),textNode('p','Explore your telemetry or define reusable questions in Definitions.'));el('package').textContent='Example grammar is not configured.';el('prebuilts').textContent='Example catalog is not configured.';}
 return config;
}).catch(error=>{el('package').textContent='Example configuration unavailable · '+error.message;el('prebuilts').textContent='Example configuration unavailable · '+error.message;return null});
el('runOperations').disabled=true;
el('runHttpPackage').disabled=true;
const runtimeCapabilities=configurationJSON('/api/capabilities').then(config=>{
 if(config.version!=='goo/runtime-capabilities/v1'||!['lab','generic'].includes(config.profile)||![config.preparedHTTP?.state,config.declaredOperations?.state].every(state=>['configured','unconfigured'].includes(state)))throw Error('Invalid runtime capabilities');
 const operations=config.declaredOperations.state==='configured',prepared=config.preparedHTTP.state==='configured';
 el('runOperations').disabled=!operations;el('operationPackage').disabled=!operations;el('captureSave').disabled=!operations;el('runHttpPackage').disabled=!prepared;
 if(!operations){el('operationStatus').textContent='Declared operations are not configured. Observed services remain available in Services.';el('definitionCatalog').append(textNode('p','No operation packages configured. General record and measurement definitions remain available.'));}
 if(!prepared)el('httpPackageStatus').textContent='Prepared HTTP package is not configured. Discover HTTP services in Services.';
 return config;
}).catch(error=>{el('operationStatus').textContent='Operation configuration unavailable · '+error.message;el('httpPackageStatus').textContent='Prepared HTTP configuration unavailable · '+error.message;return null});
async function run(spec){
 const generation=++queryGeneration,revision=controlRevision;
 el('expression').value=JSON.stringify(spec,null,2);el('status').textContent='Querying…';el('run').disabled=true;
 try{const r=await fetch('/api/query',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(spec)});const data=await r.json();if(generation!==queryGeneration)return;if(!r.ok)throw Error(data.error);
 const expression=data.expression;if(revision===controlRevision){spanFilter=expression.spanFilter;el('view').value=expression.view;el('group').value=expression.groupBy;el('minutes').value=expression.minutes;el('scenario').value=expression.scenario||'';selected=expression.reservationId||'';workflow.update(expression);el('selection').textContent=(selected?'Reservation: '+selected:'')+(spanFilter?' · Heatmap selection: '+new Date(spanFilter.fromMs).toISOString()+' to '+new Date(spanFilter.toMs).toISOString()+' · '+spanFilter.minMs+'–'+spanFilter.maxMs+' ms (upper bounds exclusive)':'');}
 el('explain').textContent=data.sql+'\n\n'+JSON.stringify(data.parameters,null,2);el('cards').replaceChildren();el('results').replaceChildren();el('chart').replaceChildren();
 if(data.summary){const s=data.summary;for(const [label,value] of [['Good',s.good],['Bad',s.bad],['Pending',s.pending],['Success preview',s.successRatio===null?'Unknown':(s.successRatio*100).toFixed(1)+'%']]){let c=textNode('div',label);c.className='card muted';c.append(textNode('b',value));el('cards').append(c)}
 el('status').textContent=`Burn preview: ${s.burnRate===null?'unknown':s.burnRate.toFixed(2)+'×'} · Alarm: ${s.alarmPreview} · Unknown subjects: ${s.unknown}. ${data.possiblyTruncated?'ROW LIMIT REACHED; values cover only displayed subjects.':''}`;
 const groups={};for(const row of data.rows){const g=groups[row.cohort]??={good:0,bad:0,pending:0,unknown:0};g[row.classification]++}for(const [name,g]of Object.entries(groups)){const line=document.createElement('div');line.append(textNode('span',name+' '));const b=document.createElement('span');b.className='bar';b.style.width=Math.min(70,Math.max(1,g.bad/(g.good+g.bad||1)*70))+'%';line.append(b,textNode('span',`${g.good} good / ${g.bad} bad / ${g.pending} pending / ${g.unknown} unknown`));el('chart').append(line)}el('chartSection').hidden=false;el('chartTitle').textContent='Bad share of mature known subjects by cohort';
 }else{el('chartSection').hidden=true;el('status').textContent=`${data.rows.length} rows${data.possiblyTruncated?' · row limit reached':''}${expression.view==='metrics'?' · Native metrics are exported point values, not derived rates.':''}`}
 if(['traces','heatmap','queue'].includes(expression.view)){
 el('chartSection').hidden=false;el('chartTitle').textContent=expression.view==='traces'?'Trace waterfall':expression.view==='queue'?'Queue and worker evidence':'gRPC duration heatmap';
 if(expression.view==='traces') Visuals.waterfall(el('chart'),data.rows,data.possiblyTruncated,row=>{const id=(row.SpanAttributes||{})['app.reservation.id'];return id?[{label:'Open reservation waterfall',select:()=>run({...expression,reservationId:id,spanFilter:undefined})}]:[]});
 else if(expression.view==='queue') Visuals.queue(el('chart'),data.rows,data);
 else Visuals.heatmap(el('chart'),data.rows,data,selection=>run({...expression,view:'traces',spanFilter:selection}));
 }
 if(!data.rows.length){el('results').textContent='No rows yet. Allow 15–30 seconds for traffic and ingestion.';return}
 const keys=Object.keys(data.rows[0]);const table=document.createElement('table'),head=document.createElement('tr');keys.forEach(k=>head.append(textNode('th',k)));table.append(head);
 for(const row of data.rows){const tr=document.createElement('tr');for(const key of keys){const td=document.createElement('td');const value=row[key];if(key==='reservation_id'){const b=textNode('button',String(value).slice(0,8)+' → spans');b.onclick=()=>{selected=value;el('view').value='traces';el('selection').textContent='Reservation: '+value;run(controls())};td.append(b)}else td.textContent=typeof value==='object'?JSON.stringify(value):value;tr.append(td)}table.append(tr)}el('results').append(table);
 }catch(e){if(generation===queryGeneration)el('status').textContent=e.message}finally{if(generation===queryGeneration)el('run').disabled=false}
}
let operationController=null,httpController=null,httpGeneration=0;
function stopOperationRead(message='Read stopped. Inspect operations to read again.'){if(!operationController)return;operationController.abort();operationController=null;operationGeneration++;el('runOperations').disabled=false;el('operationStatus').textContent=message}
function stopHTTPRead(message='Read stopped. Evaluate the package to read again.'){if(!httpController)return;httpController.abort();httpController=null;httpGeneration++;el('runHttpPackage').disabled=false;el('httpPackageStatus').textContent=message}
let operationGeneration=0,activeOperation=null,activeExplorer=null,activeRatio=null,activeComparison=null;
function renderOperationData(data){
 el('operationMode').textContent='Captured operation query · '+new Date(data.fromMs).toISOString()+' to '+new Date(data.toMs).toISOString();
 activeOperation=data;el('operationPackage').value=data.package.definition.name;
 el('operationPackageTitle').textContent=data.package.definition.title;
 el('operationStatus').textContent=data.conformance.warnings.join(' ');
 activeExplorer=Operations.overview(el,data);return activeExplorer;
}
function clearOperationEvidence(){activeRatio=null;activeComparison=null;for(const id of ['operationView','operationExplorer','fleetCoverageView','errorRatioView','counterRateView','histogramView']){el(id).replaceChildren();if(el(id).tagName==='DETAILS'){el(id).hidden=true;el(id).open=false;}}}
function showSavedCapture(c){
 globalThis.dispatchEvent(new CustomEvent('goo:show-operation-capture'));
 appWorkspace.navigate('investigations',{writeHash:false});
 const saved=id=>el('saved-'+id),data=c.evidence.operations;
 const host=el('savedOperationEvidence');host.hidden=false;
 for(const id of ['operationView','operationExplorer','fleetCoverageView','errorRatioView','counterRateView','histogramView']){saved(id).replaceChildren();if(saved(id).tagName==='DETAILS'){saved(id).hidden=true;saved(id).open=false}}
 saved('operationPackageTitle').textContent=data.package.definition.title;
 const explorer=Operations.overview(saved,data,{prefix:'saved-'}),rates=c.evidence.rates,hist=c.evidence.histograms,views=new Set(data.package.definition.views);let ratio=null,comparison=null;
 if(rates){if(views.has('counter-rates'))Operations.renderRates(saved('counterRateView'),rates,explorer.choose);if(views.has('error-ratios'))ratio=Ratios.render(saved('errorRatioView'),rates,explorer.choose,{prefix:'saved-'});if(rates.fleet&&views.has('fleet-coverage'))comparison=Fleet.render(saved('fleetCoverageView'),rates.fleet,{prefix:'saved-'});}
 if(hist)Histograms.render(saved('histogramView'),hist,explorer.choose);
 for(const [component,error]of Object.entries(c.evidence.errors||{})){const target=saved(component==='rates'?'counterRateView':'histogramView');target.hidden=false;target.append(textNode('summary',component+' unavailable in this capture'),textNode('p',error));}
 saved('operationMode').textContent='Saved capture '+c.id.slice(0,12)+' · '+c.capturedAt+' · fixed evidence';explorer.restore(c.selection);ratio?.restore(c.selection.ratio);comparison?.restore(c.selection.comparison);
 el('savedOperationLive').onclick=()=>appWorkspace.navigate('operations',{focus:true});
}
let operationCatalogStarted=false;
function loadOperationCatalog(){if(operationCatalogStarted)return;operationCatalogStarted=true;
runtimeCapabilities.then(config=>config?.declaredOperations.state==='configured'?configurationJSON('/api/operation-packages'):null).then(packages=>{if(!packages)return;const select=el('operationPackage');select.replaceChildren();for(const p of packages){const option=textNode('option',p.definition.title);option.value=p.definition.name;select.append(option);const card=textNode('div');card.className='definition-card';card.append(textNode('h3',p.definition.title),textNode('p',p.definition.resource.service+' · '+p.definition.resource.environment),textNode('p',p.definition.views.join(' · ')));const details=textNode('details');details.append(textNode('summary','Definition and revision'),textNode('pre',JSON.stringify(p,null,2)));card.append(details);el('definitionCatalog').append(card)}if(activeOperation)select.value=activeOperation.package.definition.name}).catch(error=>{el('operationStatus').textContent=error.message});
}

el('operationPackage').onchange=()=>{stopOperationRead();el('operationMode').textContent='No operation evidence loaded.';activeOperation=null;activeExplorer=null;activeRatio=null;activeComparison=null;operationGeneration++;el('runOperations').disabled=false;el('operationView').replaceChildren();el('operationExplorer').replaceChildren();el('fleetCoverageView').replaceChildren();el('fleetCoverageView').hidden=true;el('fleetCoverageView').open=false;el('errorRatioView').replaceChildren();el('errorRatioView').hidden=true;el('errorRatioView').open=false;el('counterRateView').replaceChildren();el('counterRateView').hidden=true;el('counterRateView').open=false;el('histogramView').replaceChildren();el('histogramView').hidden=true;el('histogramView').open=false;el('operationExplain').textContent='';el('operationPackageTitle').textContent='';el('operationStatus').textContent='Package changed. Inspect operations to capture its evidence.'};
el('runOperations').onclick=async()=>{
 appWorkspace.navigate('operations',{writeHash:false});stopOperationRead();const controller=operationController=new AbortController();activeOperation=null;activeExplorer=null;activeRatio=null;activeComparison=null;const generation=++operationGeneration;const current=()=>generation===operationGeneration&&!controller.signal.aborted&&appWorkspace.current()==='operations';const button=el('runOperations');button.disabled=true;el('operationStatus').textContent='Reading operation evidence…';el('operationView').replaceChildren();el('operationExplorer').replaceChildren();el('fleetCoverageView').replaceChildren();el('fleetCoverageView').hidden=true;el('fleetCoverageView').open=false;el('errorRatioView').replaceChildren();el('errorRatioView').hidden=true;el('errorRatioView').open=false;el('counterRateView').replaceChildren();el('counterRateView').hidden=true;el('counterRateView').open=false;el('histogramView').replaceChildren();el('histogramView').hidden=true;el('histogramView').open=false;
 try{const response=await Queries.getJSON('/api/operations?minutes='+encodeURIComponent(el('minutes').value)+(el('operationPackage').value?'&package='+encodeURIComponent(el('operationPackage').value):''),controller.signal);const data=response.data;if(!current())return;if(!response.ok)throw Error(data.error);
 const explorer=renderOperationData(data);if(!current())return;
 if(data.package.definition.views.includes('counter-rates')||data.package.definition.views.includes('error-ratios')||data.package.definition.views.includes('fleet-coverage')){
  try{const rateResponse=await Queries.getJSON('/api/counter-rates?package='+encodeURIComponent(data.package.definition.name)+'&fromMs='+data.fromMs+'&toMs='+data.toMs,controller.signal);const rates=rateResponse.data;if(!current())return;if(!rateResponse.ok)throw Error(rates.error);if(rates.packageRevision!==data.package.revision)throw Error('Package changed; inspect operations again.');if(data.package.definition.views.includes('counter-rates'))Operations.renderRates(el('counterRateView'),rates,explorer.choose);if(data.package.definition.views.includes('error-ratios'))activeRatio=Ratios.render(el('errorRatioView'),rates,explorer.choose);if(data.package.definition.views.includes('fleet-coverage')){if(!rates.fleet)throw Error('Fleet evidence missing');activeComparison=Fleet.render(el('fleetCoverageView'),rates.fleet)}}catch(error){if(current()){if(data.package.definition.views.includes('fleet-coverage')){el('fleetCoverageView').hidden=false;el('fleetCoverageView').replaceChildren(textNode('summary','Fleet coverage unavailable'),textNode('p',error.message))}el('counterRateView').hidden=false;el('counterRateView').replaceChildren(textNode('summary','Native counter rates unavailable'),textNode('p',error.message));if(data.package.definition.views.includes('error-ratios')){el('errorRatioView').hidden=false;el('errorRatioView').replaceChildren(textNode('summary','Operation error ratios unavailable'),textNode('p',error.message))}}}
 }

 if(!current())return;
 if(data.package.definition.views.includes('histogram-heatmap')){
  try{const response=await Queries.getJSON('/api/histograms?package='+encodeURIComponent(data.package.definition.name)+'&fromMs='+data.fromMs+'&toMs='+data.toMs,controller.signal);const histogram=response.data;if(!current())return;if(!response.ok)throw Error(histogram.error);if(histogram.packageRevision!==data.package.revision)throw Error('Package changed; inspect operations again.');Histograms.render(el('histogramView'),histogram,explorer.choose)}catch(error){if(current()){el('histogramView').hidden=false;el('histogramView').replaceChildren(textNode('summary','Native histogram unavailable'),textNode('p',error.message))}}
 }

 }catch(error){if(current())el('operationStatus').textContent=error.message}finally{if(operationController===controller)operationController=null;if(generation===operationGeneration)button.disabled=false}
};
el('runHttpPackage').onclick=async()=>{
 appWorkspace.navigate('definitions',{writeHash:false});definitionTabs.select('contracts');stopHTTPRead();const controller=httpController=new AbortController(),generation=++httpGeneration;const current=()=>generation===httpGeneration&&!controller.signal.aborted&&appWorkspace.current()==='definitions'&&definitionTabs.current()==='contracts';
 const button=el('runHttpPackage');button.disabled=true;el('httpPackageStatus').textContent='Evaluating Go package…';el('httpPackageView').replaceChildren();
 try{const response=await Queries.getJSON('/api/http-package?minutes='+encodeURIComponent(el('minutes').value),controller.signal);const data=response.data;if(!current())return;if(!response.ok)throw Error(data.error);
 el('httpPackageStatus').textContent=data.warnings.join(' ');
 const descriptor=data.expansion.views[0];if(descriptor.kind!=='aggregate-table')throw Error('Unsupported shared view descriptor: '+descriptor.kind);
 const columns=[...descriptor.groupBy,...descriptor.measures];const table=document.createElement('table'),header=document.createElement('tr');columns.forEach(c=>header.append(textNode('th',c)));table.append(header);
 for(const row of data.results){const tr=document.createElement('tr');for(const column of columns){const value=row.dimensions[column]??row.measures[column];let display=value;if(typeof value==='object')display=value.state==='ok'?(value.count??value.ratio):value.state;tr.append(textNode('td',display))}table.append(tr)}
 el('httpPackageView').append(textNode('h3',descriptor.id),table);if(!data.results.length)el('httpPackageView').append(textNode('p','No observed requests in this scope/window.'));
 el('httpPackageExplain').textContent=JSON.stringify({revision:data.expansion.revision,inventoryBasis:data.expansion.inventoryBasis,source:data.source,fromMs:data.fromMs,toMs:data.toMs,policy:data.expansion.binding.policy,query:data.query,parameters:data.parameters,duplicates:data.duplicateRecords,recipe:data.expansion.recipe,diagnostics:data.expansion.diagnostics},null,2);
 }catch(error){if(current())el('httpPackageStatus').textContent=error.message}finally{if(httpController===controller)httpController=null;if(generation===httpGeneration)button.disabled=false}
};
el('run').onclick=()=>run(controls());
el('runExpression').onclick=()=>{try{run(JSON.parse(el('expression').value))}catch(e){el('status').textContent=e.message}};
el('clear').onclick=()=>{selected='';spanFilter=undefined;el('selection').textContent='';el('scenario').value='';run(controls())};
el('view').onchange=()=>{spanFilter=undefined;if(['metrics','queue'].includes(el('view').value)){selected='';el('scenario').value='';el('selection').textContent=''}};
const workflow=Workflows.mount(el('workflows'),run,controls);

// Declaration catalogs and hidden sample UI belong to their first opened workspace.
let definitionExamplesStarted=false,exampleCatalogsStarted=false;
function loadDefinitionExamples(){if(definitionExamplesStarted)return;definitionExamplesStarted=true;
uiConfiguration.then(config=>config?.examples.state==='configured'?configurationJSON('/api/prebuilts'):null).then(catalog=>{if(!catalog)return;for(const profile of catalog.profiles){const d=document.createElement('details');d.append(textNode('summary',profile.id+' · '+profile.stability));d.append(textNode('p',profile.description));d.append(textNode('pre',JSON.stringify({requirements:profile.requirements,features:profile.features,caveats:profile.caveats},null,2)));el('prebuilts').append(d)}}).catch(error=>{el('prebuilts').textContent='Example catalog unavailable · '+error.message});
uiConfiguration.then(config=>config?.examples.state==='configured'?configurationJSON('/api/package'):null).then(p=>{if(p)el('package').textContent=JSON.stringify(p,null,2)}).catch(error=>{el('package').textContent='Example grammar unavailable · '+error.message});
}
function loadExampleCatalogs(){
 uiConfiguration.then(config=>{
  if(config?.examples.state!=='configured'||exampleCatalogsStarted)return;
  exampleCatalogsStarted=true;
  const catalog=configurationJSON('/api/example-questions');
  ExampleQuestions.mount(el('exampleQuestions'),{catalog:catalog.then(items=>structuredClone(items))});
  GuidedMeasurements.mount(el('guidedMeasurements'),{exampleState:config.examples.state,telemetryCatalog:catalog.then(items=>structuredClone(items))});
 });
}
function openWorkspaceCatalogs(route){
 if(route==='examples')loadExampleCatalogs();
 if(route==='definitions')loadDefinitionExamples();
 if(route==='definitions'||route==='operations')loadOperationCatalog();
}
let examplesStarted=false;
el('legacyExamples').addEventListener('toggle',()=>{if(document.documentElement.dataset.examples==='configured'&&el('legacyExamples').open&&!examplesStarted){examplesStarted=true;run(initial)}});
const appWorkspace=Workspace.mount();
openWorkspaceCatalogs(appWorkspace.current());
// Source owners stop hidden/superseded work; saved publications are independent.
globalThis.addEventListener('goo:workspace',event=>{if(event.detail.route!=='operations')stopOperationRead();if(event.detail.route!=='definitions')stopHTTPRead();openWorkspaceCatalogs(event.detail.route)});
document.addEventListener('visibilitychange',()=>{if(document.visibilityState==='hidden'){stopOperationRead();stopHTTPRead()}});
globalThis.addEventListener('pagehide',()=>{stopOperationRead();stopHTTPRead()});
for(const event of ['input','change'])el('minutes').addEventListener(event,()=>{stopOperationRead('Scope changed. Inspect operations to read again.');stopHTTPRead('Scope changed. Evaluate the package to read again.')});

el('defineMeasurement').onclick=()=>{dispatchEvent(new CustomEvent('goo:navigate-explore'));dispatchEvent(new CustomEvent('goo:measurements-tab'));el('measureEditor').open=true;el('measureLibrary').focus()};
const liveExploration=InvestigationTrail.mount(el('exploration'),{lazy:true});
uiConfiguration.then(config=>{const lab=config?.profile==='lab';if(!lab)el('inventoryEnvironment').value='';StreamEntry.mount(el('exploration'),{environment:lab?'local':'',autoStart:lab,onLeaveFixed:()=>liveExploration.suspend(),onEnterFixed:()=>liveExploration.resume(),onOpenFixed:(q,recordLayout,options)=>liveExploration.openQuestion(q,[],{stage:true,activity:true,recordLayout,...options}),onOpenTrace:(q,row)=>{dispatchEvent(new CustomEvent('goo:navigate-explore'));liveExploration.openTrace(q,row.TraceId,row.SpanId)}})});
const nativeRecordNavigation=NativeRecordNavigation.mount(el('nativeRecordReturn'),{nativeHost:el('measurementWorkbench'),active:()=>appWorkspace.current()==='explore'&&el('exploreMeasurementsMode').getAttribute('aria-selected')==='true',stageQuestion:q=>{appWorkspace.navigate('explore',{writeHash:false});exploreMode(false);liveExploration.openQuestion(q,[],{stage:true})},showNative:()=>{appWorkspace.navigate('explore',{writeHash:false});exploreMode(true)},suspendRecords:()=>liveExploration.suspend()});
Measurements.mount(el('measurementWorkbench'),{onRecordContext:(detail,opener)=>nativeRecordNavigation.stage(detail,opener),onRestoreStep:hash=>nativeRecordNavigation.restore(hash)});
function exploreMode(measures){el('exploration').hidden=measures;el('measurementWorkbench').hidden=!measures;el('exploreRecordsMode').setAttribute('aria-selected',String(!measures));el('exploreMeasurementsMode').setAttribute('aria-selected',String(measures));el('exploreRecordsMode').tabIndex=measures?-1:0;el('exploreMeasurementsMode').tabIndex=measures?0:-1;dispatchEvent(new CustomEvent('goo:explore-mode',{detail:{measurements:measures}}))}
el('exploreRecordsMode').onclick=()=>exploreMode(false);
el('exploreMeasurementsMode').onclick=()=>exploreMode(true);
globalThis.addEventListener('goo:measurements-tab',()=>exploreMode(true));
globalThis.addEventListener('goo:navigate-explore',()=>exploreMode(false));
for(const id of ['exploreRecordsMode','exploreMeasurementsMode'])el(id).onkeydown=e=>{if(['ArrowLeft','ArrowRight','Home','End'].includes(e.key)){e.preventDefault();const measures=e.key==='End'||(e.key!=='Home'&&id==='exploreRecordsMode');exploreMode(measures);el(measures?'exploreMeasurementsMode':'exploreRecordsMode').focus()}};
const liveInventory=Inventory.mount(el('serviceInventory'));
let entityKind='services';
const stageResourceQuestion=q=>{appWorkspace.navigate('explore');exploreMode(false);history.pushState(null,'',Workspace.questionLink(q));liveExploration.openQuestion(q,[],{stage:true})};
const hostBrowser=HostBrowser.mount(el('hostEntityPanel'),{active:()=>entityKind==='hosts'&&appWorkspace.current()==='services',onReturn:()=>{appWorkspace.navigate('services');entityTabs.select('hosts')},onRecordQuestion:stageResourceQuestion});
const kubernetesBrowser=KubernetesBrowser.mount(el('kubernetesEntityPanel'),{active:()=>entityKind==='kubernetes'&&appWorkspace.current()==='services',onReturn:()=>{appWorkspace.navigate('services');entityTabs.select('kubernetes')},onRecordQuestion:stageResourceQuestion});
const entityTabs=ResourceTabs.mount([{key:'services',button:el('entitiesServices'),panel:el('serviceEntityPanel'),leave:()=>liveInventory.suspend()},{key:'hosts',button:el('entitiesHosts'),panel:el('hostEntityPanel'),leave:()=>hostBrowser.cancel()},{key:'kubernetes',button:el('entitiesKubernetes'),panel:el('kubernetesEntityPanel'),leave:()=>kubernetesBrowser.cancel()}],{onSelect:key=>entityKind=key});
const definitionTabs=ResourceTabs.mount([
 {key:'records',button:el('definitionsRecords'),panel:el('definitionRecordsPanel')},
 {key:'measurements',button:el('definitionsMeasurements'),panel:el('definitionMeasurementsPanel')},
 {key:'contracts',button:el('definitionsContracts'),panel:el('definitionContractsPanel'),leave:()=>stopHTTPRead()},
],{initial:'records'});
for(const event of ['goo:draft-telemetry-package','goo:preview-telemetry-package'])globalThis.addEventListener(event,()=>{appWorkspace.navigate('definitions');definitionTabs.select('records')});
TelemetryPackages.mount(el('telemetryPackages'));
globalThis.addEventListener('goo:navigate-explore',()=>appWorkspace.navigate('explore'));
Captures.mount({openLive:page=>{if(page.measurement){globalThis.dispatchEvent(new CustomEvent('goo:open-measurement',{detail:{yaml:JSON.stringify(page.measurement.question.definition)}}))}else if(page.explore){appWorkspace.navigate('explore');liveExploration.openQuestion(page.receipt.question.explore,page.explore.fields,{stage:true})}else{appWorkspace.navigate('services');liveInventory.openQuestion(page.receipt.question.inventory)}}});
globalThis.addEventListener('goo:show-general-capture',()=>{el('savedOperationEvidence').hidden=true});
globalThis.addEventListener('goo:navigate-investigations',()=>appWorkspace.navigate('investigations',{writeHash:false}));
Documents.mount(el('investigationDocuments'));
Meaning.index(el('savedAnnotations'));
Retention.mount(el('savedRetainedWorkspace'),{catalogHost:el('savedRetainedCatalog'),recoveryHost:el('savedRecovery')});
SavedWorkspace.mount();
ActivityViews.library(el('savedActivityViews'));
Investigations.mount({generation:()=>operationGeneration,question:()=>activeOperation?{package:activeOperation.package.definition.name,packageRevision:activeOperation.package.revision,fromMs:activeOperation.fromMs,toMs:activeOperation.toMs}:null,selection:()=>({...activeExplorer?.getState(),...(activeRatio?.getState()?{ratio:activeRatio.getState()}:{}),...(activeComparison?.getState()?{comparison:activeComparison.getState()}: {})}),show:showSavedCapture});

el('expression').value=JSON.stringify(initial,null,2);
