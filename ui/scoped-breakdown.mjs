/* Application adapter: supplied data and local display/selection; host owns effects. */
import './scope-view.mjs';
import {eligibleKinds,validateSnapshot,groupIntent,intervalIntent} from './view-model.mjs';
import {node} from './components.mjs';
const preferences=new WeakMap();
const categoryLabel=value=>value===null?'Missing key':value===''?'Present empty string':String(value);
const utc=value=>new Date(Number(value)).toISOString();

export function mount(target,snapshot,{onScope,onWindow,onSelect}={}){
 const check=validateSnapshot(snapshot);
 if(!check.ok){target.replaceChildren(node('p','Display unavailable: '+check.reasons.join('; ')));return}
 const temporal=snapshot.shape==='time-series',dimensions=snapshot.dimensions||[],multi=dimensions.length>1;
 const section=node('section'),toolbar=node('div'),label=node('label','View as '),pick=node('select'),chart=node('goo-chart'),status=node('p');
 const stage=node('button',temporal?'Stage selected interval':'Stage selected group'),completeLabel=node('label'),complete=node('input');
 section.className='scoped-breakdown';section.setAttribute('aria-label',temporal?'Scoped timeline':'Scoped breakdown');
 toolbar.className='ui-actions';pick.setAttribute('aria-label',temporal?'Timeline view':'Breakdown view');
 const choices=eligibleKinds(snapshot),labels={table:'Table',bar:'Bars',line:'Line','time-bar':'Intervals'};
 for(const kind of choices){const option=node('option',labels[kind]);option.value=kind;pick.append(option)}
 const preference=preferences.get(target),preferred=temporal?'time-bar':multi?'table':'bar';
 pick.value=choices.includes(preference?.kind)?preference.kind:choices.includes(preferred)?preferred:'table';label.append(pick);
 complete.type='checkbox';complete.checked=preference?.requiresComplete??false;completeLabel.className='scoped-breakdown-complete';
 completeLabel.append(complete,document.createTextNode(temporal?'Complete intervals only':'Complete groups only'));
 stage.type='button';stage.disabled=true;stage.hidden=temporal?!onWindow:!onScope;
 status.className='scoped-breakdown-selection';status.setAttribute('aria-live','polite');let selected=null;
 const display=()=>{
  preferences.set(target,{kind:pick.value,requiresComplete:complete.checked});
  chart.display={kind:pick.value,title:temporal?'Captured record intervals':multi?'Captured groups':'Scoped breakdown',requiresComplete:complete.checked,...(!temporal&&!multi?{category:'category',value:'count'}:{})};
 };
 chart.snapshot=snapshot;display();pick.addEventListener('change',display);complete.addEventListener('change',display);
 chart.addEventListener('goo-select',event=>{
  const intent=event.detail;
  if(intent?.type!=='select'||intent.datasetId!==snapshot.id||intent.revision!==snapshot.revision||intent.scopeId!==snapshot.scope.id||intent.generation!==snapshot.scope.generation)return;
  selected=snapshot.rows.find(row=>row.key===intent.key);stage.disabled=!selected||selected.state==='unknown';
  status.textContent=!selected?'Selection unavailable':temporal?`${utc(selected.values.fromMs)} → ${utc(selected.values.toMs)} · ${selected.values.count??'Unknown'} captured records${selected.state==='no_data'?' · no captured data; health unknown':''}`:`${(dimensions.length?dimensions:['category']).map(field=>categoryLabel(selected.values[field])).join(' / ')} · ${selected.values.count} captured records`;
  if(selected&&selected.state!=='unknown')onSelect?.(selected,snapshot);
 });
 toolbar.append(label,completeLabel,stage);section.append(toolbar,chart);
 if(!temporal&&onScope)for(const field of dimensions){const facet=node('goo-scope-facet');facet.snapshot=snapshot;facet.field=field;facet.addEventListener('goo-scope',event=>onScope(event.detail,snapshot));section.append(facet)}
 stage.addEventListener('click',()=>{
  if(!selected||selected.state==='unknown')return;
  if(temporal)onWindow?.(intervalIntent(snapshot,selected.key),snapshot);
  else onScope?.(groupIntent(snapshot,selected.key),snapshot);
 });
 section.append(status);target.replaceChildren(section);return {snapshot,chart};
}
