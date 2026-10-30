/* Application adapter: one supplied dataset, local display/selection, explicit scope staging. */
import './scope-view.mjs';
import {eligibleKinds,validateSnapshot,scopeIntent} from './view-model.mjs';
import {node} from './components.mjs';
const preferences=new WeakMap();

export function mount(target,snapshot,{onScope}={}){
 const check=validateSnapshot(snapshot);
 if(!check.ok){target.replaceChildren(node('p','Breakdown unavailable: '+check.reasons.join('; ')));return}
 const section=node('section'),toolbar=node('div'),label=node('label','View as '),pick=node('select'),chart=node('goo-chart'),facet=node('goo-scope-facet'),status=node('p'),stage=node('button','Stage selected group'),completeLabel=node('label'),complete=node('input');
 section.className='scoped-breakdown';section.setAttribute('aria-label','Scoped breakdown');toolbar.className='ui-actions';pick.setAttribute('aria-label','Breakdown view');
 for(const kind of eligibleKinds(snapshot)){const option=node('option',kind==='bar'?'Bars':kind==='line'?'Line':'Table');option.value=kind;pick.append(option)}
 const preference=preferences.get(target);pick.value=eligibleKinds(snapshot).includes(preference?.kind)?preference.kind:eligibleKinds(snapshot).includes('bar')?'bar':'table';label.append(pick);complete.type='checkbox';complete.checked=preference?.requiresComplete??false;completeLabel.className='scoped-breakdown-complete';completeLabel.append(complete,document.createTextNode('Complete groups only'));stage.type='button';stage.disabled=true;stage.hidden=!onScope;status.className='scoped-breakdown-selection';status.setAttribute('aria-live','polite');let selected=null;
 const display=()=>{preferences.set(target,{kind:pick.value,requiresComplete:complete.checked});chart.display={kind:pick.value,category:'category',value:'count',requiresComplete:complete.checked}};
 chart.snapshot=snapshot;display();facet.snapshot=snapshot;facet.field='category';facet.hidden=!onScope;
 pick.addEventListener('change',display);complete.addEventListener('change',display);
 chart.addEventListener('goo-select',e=>{const intent=e.detail;if(intent?.type!=='select'||intent.datasetId!==snapshot.id||intent.revision!==snapshot.revision||intent.scopeId!==snapshot.scope.id||intent.generation!==snapshot.scope.generation)return;selected=snapshot.rows.find(row=>row.key===intent.key);stage.disabled=!selected;status.textContent=selected?`${selected.values.category===null?'Missing key':selected.values.category===''?'Present empty string':selected.values.category} · ${selected.values.count} captured records`:'Selection unavailable';});
 const emit=intent=>{if(onScope)onScope(intent,snapshot)};
 facet.addEventListener('goo-scope',e=>emit(e.detail));stage.addEventListener('click',()=>{if(selected)emit(scopeIntent(snapshot,'category',selected.values.category))});
 toolbar.append(label,completeLabel,stage);section.append(toolbar,chart,facet,status);target.replaceChildren(section);
 return {snapshot,chart};
}
