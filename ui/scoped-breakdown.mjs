/* Application adapter: supplied data and local display/selection; host owns effects. */
import './scope-view.mjs';
import {eligibleKinds,validateSnapshot,groupIntent,intervalIntent,displayRows,fractionFacts} from './view-model.mjs';
import {node} from './components.mjs';
import {formatCell,supportText,exactText,categoryText,isFraction,fractionContextText} from './view-values.mjs';
const preferences=new WeakMap();
const utc=value=>new Date(Number(value)).toISOString();

export function mount(target,snapshot,{onScope,onWindow,onSelect,onInspectTrace,distribution=false,facets:showFacets=true,initialPresentation,onPresentation,allowedMeasures}={}){
 const check=validateSnapshot(snapshot);
 if(!check.ok){target.replaceChildren(node('p','Display unavailable: '+check.reasons.join('; ')));return}
 const temporal=snapshot.shape==='time-series',dimensions=snapshot.dimensions||[],multi=dimensions.length>1;
 const section=node('section'),toolbar=node('div'),label=node('label','View as '),pick=node('select'),chart=node('goo-chart'),status=node('p');
 const stage=node('button',temporal?'Stage selected interval':'Stage selected group'),completeLabel=node('label'),complete=node('input'),measureLabel=node('label','Measure '),measure=node('select');
 const scaleLabel=node('label','Scale '),scale=node('select'),orderLabel=node('label','Sort '),order=node('select'),more=node('details'),inspect=node('button','Open selected trace');
 section.className='scoped-breakdown';section.setAttribute('aria-label',temporal?'Scoped timeline':'Scoped breakdown');
 toolbar.className='ui-actions';pick.setAttribute('aria-label',temporal?'Timeline view':'Breakdown view');
 const choices=eligibleKinds(snapshot),labels={table:'Table',bar:'Bars',line:'Line','time-bar':'Intervals'};
 for(const kind of choices){const option=node('option',labels[kind]);option.value=kind;pick.append(option)}
 const preference={...preferences.get(target),...initialPresentation},preferred=temporal?'time-bar':multi?'table':'bar';let query=preference.query||'';
 const measures=snapshot.fields.filter(field=>(field.key==='count'||field.aggregation)&&(!allowedMeasures||allowedMeasures.includes(field.key)));
 measure.setAttribute('aria-label',temporal?'Timeline measure':'Breakdown measure');
 for(const field of measures){const option=node('option',field.label+(isFraction(field)?' · %':field.unit&&field.aggregation?.family!=='duration'?' · '+field.unit:''));option.value=field.key;measure.append(option)}
 const chooseMeasure=value=>{if(!measures.some(field=>field.key===value)&&!Array.from(measure.options).some(option=>option.value===value)){const option=node('option','Unavailable: '+value);option.value=value;option.disabled=true;measure.append(option)}measure.value=value;measureLabel.hidden=measure.options.length<2||multi};
 chooseMeasure(preference.value||'count');measureLabel.append(measure);measureLabel.hidden=measure.options.length<2||multi;
 pick.value=choices.includes(preference?.kind)?preference.kind:choices.includes(preferred)?preferred:'table';label.append(pick);
 complete.type='checkbox';complete.checked=preference?.requiresComplete??false;completeLabel.className='scoped-breakdown-complete';
 completeLabel.append(complete,document.createTextNode(temporal?'Complete intervals only':'Complete groups only'));
 scale.setAttribute('aria-label','Count scale');for(const [value,text] of [['population','Population share'],['max','Captured values']]){const option=node('option',text);option.value=value;scale.append(option)}scale.value=preference.scale||(distribution?'population':'max');scaleLabel.append(scale);
 order.setAttribute('aria-label','Group order');for(const [value,text] of [['source','Source order'],['label','Label'],['value-desc','Value · descending'],['value-asc','Value · ascending']]){const option=node('option',text);option.value=value;order.append(option)}order.value=preference.order||(distribution?'value-desc':'source');orderLabel.append(order);
 more.className='ui-disclosure';more.append(node('summary','More options'),orderLabel,completeLabel);
 inspect.type='button';inspect.disabled=true;inspect.hidden=!onInspectTrace;inspect.title='Read this trace across services in the captured window. Matching records do not establish a complete trace or root latency.';
 stage.type='button';stage.disabled=true;stage.hidden=temporal?!onWindow:!onScope;
 status.className='scoped-breakdown-selection';status.setAttribute('aria-live','polite');let selected=null,actionsPaused=false;
 const hasActions=(temporal?!!onWindow:!!onScope)||!!onInspectTrace,facets=[],pauseNotice=onScope||onWindow?'Run again to use these selections.':'Reopen this view to use its trace link.';
 const presentation=()=>({kind:pick.value,value:measure.value,scale:measure.value==='count'?scale.value:'max',query,order:order.value,requiresComplete:complete.checked});
 const display=()=>{
  preferences.set(target,{...presentation(),scale:scale.value});scaleLabel.hidden=measure.value!=='count';
  chart.display={kind:pick.value,title:preference.title||(temporal?'Captured record intervals':multi?'Captured groups':'Scoped breakdown'),requiresComplete:complete.checked,...(!multi?{...(!temporal?{category:'category'}:{}),value:measure.value}:{}),...(distribution?{scale:measure.value==='count'?scale.value:'max',query,order:order.value}:{})};
  onPresentation?.(presentation());
 };
 const selectionSummary=()=>{
  if(!selected){status.textContent=actionsPaused&&hasActions?pauseNotice:'';return}
  const field=snapshot.fields.find(field=>field.key===measure.value),cell=selected.cells?.[measure.value];
  const location=temporal?`${utc(selected.values.fromMs)} → ${utc(selected.values.toMs)}`:(dimensions.length?dimensions:['category']).map(key=>categoryText(selected.values[key])).join(' / ');
  const context=fractionContextText(fractionFacts(snapshot,selected,measure.value))||supportText(cell);
  status.textContent=`${location} · ${field?.aggregation?formatCell(field,selected.values[field.key],cell):selected.values.count+' captured records'}${context?' · '+context:temporal&&selected.state==='no_data'?' · no captured data; health unknown':''}${actionsPaused&&hasActions?' · '+pauseNotice:''}`;
  status.title=field?exactText(field,selected.values[field.key],cell):'';
 };
 chart.snapshot=snapshot;display();pick.addEventListener('change',display);complete.addEventListener('change',display);scale.addEventListener('change',display);order.addEventListener('change',display);measure.addEventListener('change',()=>{display();selectionSummary()});
 chart.addEventListener('goo-select',event=>{
  const intent=event.detail;
  if(intent?.type!=='select'||intent.datasetId!==snapshot.id||intent.revision!==snapshot.revision||intent.scopeId!==snapshot.scope.id||intent.generation!==snapshot.scope.generation)return;
  selected=snapshot.rows.find(row=>row.key===intent.key);stage.disabled=actionsPaused||!selected||selected.state==='unknown';
  const trace=selected?.values[dimensions[0]];inspect.disabled=actionsPaused||!selected||typeof trace!=='string'||!/^[a-f0-9]{32}$/.test(trace)||/^0+$/.test(trace);
  selectionSummary();
  if(selected&&selected.state!=='unknown'&&!actionsPaused)onSelect?.(selected,snapshot);
 });
 toolbar.append(label,measureLabel);if(distribution)toolbar.append(scaleLabel,stage,inspect,more);else toolbar.append(completeLabel,stage,inspect);section.append(toolbar,chart);
 if(showFacets&&!temporal&&onScope)for(const field of dimensions){const facet=node('goo-scope-facet');facet.snapshot=snapshot;facet.field=field;facet.addEventListener('goo-scope',event=>{if(!actionsPaused)onScope(event.detail,snapshot)});facets.push(facet);section.append(facet)}
 stage.addEventListener('click',()=>{
  if(actionsPaused||!selected||selected.state==='unknown')return;
  if(temporal)onWindow?.(intervalIntent(snapshot,selected.key),snapshot);
  else onScope?.(groupIntent(snapshot,selected.key),snapshot);
 });
 inspect.addEventListener('click',()=>{if(!actionsPaused&&selected&&!inspect.disabled)onInspectTrace?.(selected,inspect)});
 section.append(status);target.replaceChildren(section);return {snapshot,chart,presentation,setQuery(value){query=value;display();return {visible:displayRows(snapshot,chart.display).length,total:snapshot.rows.length}},setPresentation(value){if(value?.title!==undefined)preference.title=value.title;if(value?.value!==undefined)chooseMeasure(value.value);if(value?.kind!==undefined&&choices.includes(value.kind))pick.value=value.kind;if(value?.scale!==undefined)scale.value=value.scale;if(value?.order!==undefined)order.value=value.order;if(value?.query!==undefined)query=value.query;if(value?.requiresComplete!==undefined)complete.checked=value.requiresComplete;display();selectionSummary()},suspendActions(){actionsPaused=true;stage.disabled=true;inspect.disabled=true;for(const facet of facets){facet.inert=true;facet.setAttribute('aria-disabled','true');for(const button of facet.querySelectorAll('button'))button.disabled=true}selectionSummary()}};
}
