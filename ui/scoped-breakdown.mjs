/* Application adapter: supplied data and local display/selection; host owns effects. */
import './scope-view.mjs';
import {eligibleKinds,validateSnapshot,groupIntent,intervalIntent} from './view-model.mjs';
import {node} from './components.mjs';
import {formatCell,supportText,exactText} from './view-values.mjs';
const preferences=new WeakMap();
const categoryLabel=value=>value===null?'Missing key':value===''?'Present empty string':String(value);
const utc=value=>new Date(Number(value)).toISOString();

export function mount(target,snapshot,{onScope,onWindow,onSelect}={}){
 const check=validateSnapshot(snapshot);
 if(!check.ok){target.replaceChildren(node('p','Display unavailable: '+check.reasons.join('; ')));return}
 const temporal=snapshot.shape==='time-series',dimensions=snapshot.dimensions||[],multi=dimensions.length>1;
 const section=node('section'),toolbar=node('div'),label=node('label','View as '),pick=node('select'),chart=node('goo-chart'),status=node('p');
 const stage=node('button',temporal?'Stage selected interval':'Stage selected group'),completeLabel=node('label'),complete=node('input'),measureLabel=node('label','Measure '),measure=node('select');
 section.className='scoped-breakdown';section.setAttribute('aria-label',temporal?'Scoped timeline':'Scoped breakdown');
 toolbar.className='ui-actions';pick.setAttribute('aria-label',temporal?'Timeline view':'Breakdown view');
 const choices=eligibleKinds(snapshot),labels={table:'Table',bar:'Bars',line:'Line','time-bar':'Intervals'};
 for(const kind of choices){const option=node('option',labels[kind]);option.value=kind;pick.append(option)}
 const preference=preferences.get(target),preferred=temporal?'time-bar':multi?'table':'bar';
 const measures=snapshot.fields.filter(field=>field.key==='count'||field.aggregation);
 measure.setAttribute('aria-label',temporal?'Timeline measure':'Breakdown measure');
 for(const field of measures){const option=node('option',field.label+(field.unit&&field.aggregation?.family!=='duration'?' · '+field.unit:''));option.value=field.key;measure.append(option)}
 measure.value=measures.some(field=>field.key===preference?.value)?preference.value:'count';measureLabel.append(measure);measureLabel.hidden=measures.length<2||multi;
 pick.value=choices.includes(preference?.kind)?preference.kind:choices.includes(preferred)?preferred:'table';label.append(pick);
 complete.type='checkbox';complete.checked=preference?.requiresComplete??false;completeLabel.className='scoped-breakdown-complete';
 completeLabel.append(complete,document.createTextNode(temporal?'Complete intervals only':'Complete groups only'));
 stage.type='button';stage.disabled=true;stage.hidden=temporal?!onWindow:!onScope;
 status.className='scoped-breakdown-selection';status.setAttribute('aria-live','polite');let selected=null,actionsPaused=false;
 const hasActions=temporal?!!onWindow:!!onScope,facets=[];
 const display=()=>{
  preferences.set(target,{kind:pick.value,value:measure.value,requiresComplete:complete.checked});
  chart.display={kind:pick.value,title:temporal?'Captured record intervals':multi?'Captured groups':'Scoped breakdown',requiresComplete:complete.checked,...(!multi?{...(!temporal?{category:'category'}:{}),value:measure.value}:{})};
 };
 const selectionSummary=()=>{
  if(!selected){status.textContent=actionsPaused&&hasActions?'Run again to use these selections.':'';return}
  const field=snapshot.fields.find(field=>field.key===measure.value),cell=selected.cells?.[measure.value];
  const location=temporal?`${utc(selected.values.fromMs)} → ${utc(selected.values.toMs)}`:(dimensions.length?dimensions:['category']).map(key=>categoryLabel(selected.values[key])).join(' / ');
  status.textContent=`${location} · ${field?.aggregation?formatCell(field,selected.values[field.key],cell):selected.values.count+' captured records'}${cell?' · '+supportText(cell):temporal&&selected.state==='no_data'?' · no captured data; health unknown':''}${actionsPaused&&hasActions?' · Run again to use these selections.':''}`;
  status.title=field?exactText(field,selected.values[field.key],cell):'';
 };
 chart.snapshot=snapshot;display();pick.addEventListener('change',display);complete.addEventListener('change',display);measure.addEventListener('change',()=>{display();selectionSummary()});
 chart.addEventListener('goo-select',event=>{
  const intent=event.detail;
  if(intent?.type!=='select'||intent.datasetId!==snapshot.id||intent.revision!==snapshot.revision||intent.scopeId!==snapshot.scope.id||intent.generation!==snapshot.scope.generation)return;
  selected=snapshot.rows.find(row=>row.key===intent.key);stage.disabled=actionsPaused||!selected||selected.state==='unknown';
  selectionSummary();
  if(selected&&selected.state!=='unknown'&&!actionsPaused)onSelect?.(selected,snapshot);
 });
 toolbar.append(label,measureLabel,completeLabel,stage);section.append(toolbar,chart);
 if(!temporal&&onScope)for(const field of dimensions){const facet=node('goo-scope-facet');facet.snapshot=snapshot;facet.field=field;facet.addEventListener('goo-scope',event=>{if(!actionsPaused)onScope(event.detail,snapshot)});facets.push(facet);section.append(facet)}
 stage.addEventListener('click',()=>{
  if(actionsPaused||!selected||selected.state==='unknown')return;
  if(temporal)onWindow?.(intervalIntent(snapshot,selected.key),snapshot);
  else onScope?.(groupIntent(snapshot,selected.key),snapshot);
 });
 section.append(status);target.replaceChildren(section);return {snapshot,chart,suspendActions(){actionsPaused=true;stage.disabled=true;for(const facet of facets){facet.inert=true;facet.setAttribute('aria-disabled','true');for(const button of facet.querySelectorAll('button'))button.disabled=true}selectionSummary()}};
}
