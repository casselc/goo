/* Application-owned interactions shared by record presentations. Display data grants no source authority. */
import * as Attributes from './attributes.mjs';
import * as Discovery from './discovery.mjs';
import * as RecordColumns from './record-columns.mjs';
import * as RecordPresentation from './record-presentation.mjs';
import * as Keeps from './kept-selection.mjs';
import * as Retention from './retention.mjs';
import * as TraceSearch from './trace-search.mjs';
import * as UI from './components.mjs';
import * as Visuals from './visualizations.mjs';
import './record-list.mjs';
const node=UI.node,mounts=new WeakMap();

// Positions identify occurrences in this page, including identical/repeated records.
// This state has no event identity, source token, query or publication authority.
export function createSelection(size){
 if(!Number.isSafeInteger(size)||size<0)throw Error('Use a bounded page size.');
 const selected=new Set();let revision=0,submitted=false,busy=false;
 return Object.freeze({
  indexes:()=>Object.freeze([...selected].sort((a,b)=>a-b)),
  has:index=>selected.has(index),revision:()=>revision,
  state:()=>Object.freeze({selected:selected.size,submitted,busy}),
  set(index,checked){if(!Number.isInteger(index)||index<0||index>=size||typeof checked!=='boolean'||submitted||busy||checked&&!selected.has(index)&&selected.size>=20)return false;if(checked)selected.add(index);else selected.delete(index);revision++;return true},
  clear(){if(busy)return false;selected.clear();submitted=false;revision++;return true},
  submission(value){if(typeof value?.submitted!=='boolean'||typeof value?.busy!=='boolean')throw Error('Use explicit submission state.');submitted=value.submitted;busy=value.busy}
 });
}

function surface(target,rows,signal,options){
 if(options.recordKind==='record-list'){
  const list=node('goo-record-list');list.page=options.recordPage;list.display=options.recordLayout;target.append(list);
  const items=list.recordItems(),page=list.page,display=list.display;
  if(items.length!==rows.length)throw Error('List hosts must match the admitted page.');
  for(const item of items){item.actionsHost.hidden=false;item.details.hidden=true;item.detailHost.classList.add('record-detail-row');item.detailHost.hidden=true}
  return {viewport:list,current:()=>list.page===page&&list.display===display&&list.recordItems()===items,
   rowElements:items.map(item=>item.element),
   row(record,index){const item=items[index];return {row:item.element,time:item.element.querySelector('.record-list-meta')||item.element,actions:item.actionsHost,detailRow:item.detailHost,detailCell:item.detailHost}},
   append(){},finish(){}
  };
 }
 const records=RecordPresentation.table({signal,mobileCompact:true,caption:'Captured '+signal+' · inspection page; full totals shown separately',label:'Captured '+signal+' inspection page'});
 records.element.className='telemetry-table';records.header.className='record-columns';
 const rowElements=[];
 return {...records,rowElements,current:()=>true,
  row(record){const result=records.row(record),detailRow=node('tr'),detailCell=node('td');detailRow.className='record-detail-row';detailRow.hidden=true;detailCell.colSpan=4;detailRow.append(detailCell);rowElements.push(result.row);return {row:result.row,time:result.cells[0],actions:result.cells[3],detailRow,detailCell}},
  append(element){records.body.append(element)},finish(){target.append(records.viewport)}
 };
}

export function mount(target,rows,signal,origin,basis,options={},context={}){
 const {active=()=>true,attributeOptions=()=>({}),readOnly=false,savedPage=()=>null,primary=()=>null,openTrace,openSurrounding,layouts=new Map(),initialLayout,surrounding=()=>({})}=context;
 const surround=surrounding(),owner={};mounts.set(target,owner);target.replaceChildren();
 if(!rows.length){const empty=node('goo-result-state','No captured records in this page; absence is not health evidence.');empty.state='empty';target.append(empty);return}
 const records=surface(target,rows,signal,options),{viewport}=records,isList=options.recordKind==='record-list';
 const canUseRecord=()=>mounts.get(target)===owner&&active()&&target.isConnected&&viewport.isConnected&&target.contains(viewport)&&records.current()&&options.canUseRecord?.()!==false;
 const originalOptions=attributeOptions(target,signal,options,origin,basis);
 const actions={...originalOptions,
  canRefine:()=>canUseRecord()&&originalOptions.canRefine?.()!==false,
  canPlot:()=>canUseRecord()&&(originalOptions.canPlot||originalOptions.canRefine)?.()!==false,
  canTraceSearch:()=>canUseRecord()&&(originalOptions.canTraceSearch||originalOptions.canRefine)?.()!==false,
  onFilter:originalOptions.onFilter?((...args)=>{if(actions.canRefine())originalOptions.onFilter(...args)}):undefined,
  onBreakdown:originalOptions.onBreakdown?((...args)=>{if(actions.canRefine())originalOptions.onBreakdown(...args)}):undefined,
  onTraceSearch:originalOptions.onTraceSearch?((...args)=>{if(actions.canTraceSearch())originalOptions.onTraceSearch(...args)}):undefined,
  onFacet:originalOptions.onFacet?((...args)=>{if(actions.canRefine())originalOptions.onFacet(...args)}):undefined,
  onPlot:originalOptions.onPlot?((field,button)=>{if(actions.canPlot())originalOptions.onPlot(field,button)}):undefined
 };
 const panel=node('section'),inspect=UI.inspector(panel,{onClose:()=>{const row=panel.closest('.record-detail-row');if(row)row.hidden=true;for(const button of target.querySelectorAll('[data-inspect-record]'))button.setAttribute('aria-pressed','false')}});
 const selection=createSelection(rows.length),selectionState=selection.state,inputs=new Map(),recordSelection=node('section',undefined,'record-selection');
 recordSelection.hidden=true;recordSelection.setAttribute('aria-label','Selected '+signal+' records');
 const layoutHost=node('div',undefined,'record-fields-toolbar');
 function selectionChanged(){
  const state=selection.state();for(const [index,input]of inputs){input.checked=selection.has(index);input.disabled=state.submitted||state.busy||!input.checked&&state.selected>=20}
  if(isList){layoutHost.inert=state.selected>0||state.busy;layoutHost.setAttribute('aria-disabled',String(layoutHost.inert))}
  options.onSelectionState?.(state);
 }
 function selectionReference(){const indexes=selection.indexes();return readOnly&&target===primary()?{captureID:savedPage().receipt.id,indexes:indexes.map(index=>Number(origin.offset||0)+index)}:{resultToken:options.resultToken,indexes:[...indexes]}}
 function renderSelection(){
  recordSelection.replaceChildren();recordSelection.hidden=!selection.state().selected;if(recordSelection.hidden)return;
  const revision=selection.revision(),current=()=>canUseRecord()&&revision===selection.revision()&&recordSelection.isConnected;
  recordSelection.append(node('p',selection.state().selected+' selected page observations · maximum 20. Shared recorded fields can define a broader population; these rows are not its total.'));
  const clear=node('button','Clear selected records');clear.onclick=()=>{if(!current()||!recordSelection.contains(clear)||!selection.clear())return;selectionChanged();renderSelection()};recordSelection.append(clear);
  const shared=node('details'),fields=node('div');shared.append(node('summary','Fields shared by selected records'),fields);let built=false;
  shared.addEventListener('toggle',()=>{if(!shared.open||built||!current()||!recordSelection.contains(shared))return;built=true;const allowed=()=>current()&&recordSelection.contains(shared)&&actions.canRefine();Attributes.mountShared(fields,selection.indexes().map(index=>rows[index]),{
   canPlot:()=>current()&&actions.canPlot(),onPlot:actions.onPlot?((field,button)=>{if(current()&&actions.canPlot())actions.onPlot(field,button)}):undefined,
   onFacet:(...args)=>{if(allowed())options.onFacet?.(...args)},canRefine:allowed
  })});
  const provenance=node('details');provenance.append(node('summary','Selected record scope and provenance'),node('p','Page positions identify retained observations in this read only; reported trace/span IDs do not establish unique events, complete traces or causal relationships.'),node('pre',JSON.stringify({question:origin,basis,observations:selection.indexes().map(index=>({pageIndex:index,timestamp_ns:rows[index].timestamp_ns,traceId:rows[index].TraceId,spanId:rows[index].SpanId}))},null,2)));
  recordSelection.append(shared,provenance);
  if(readOnly&&savedPage()&&target===primary()||options.resultToken){
   const host=node('div');recordSelection.append(host);
   Keeps.mount(host,selectionReference(),{expected:{question:readOnly&&target===primary()?savedPage().receipt.question.explore:origin,basis},active:current,onState:state=>{if(!current())return;selection.submission(state);selectionChanged();clear.disabled=state.busy}});
   if(origin.environmentSelector==null){const retention=node('div');recordSelection.append(retention);Retention.selection(retention,selectionReference(),{active:current})}
  }
 }
 function selectRecord(index,input){
  if(!canUseRecord()||!selection.set(index,input.checked)){input.checked=selection.has(index);return}
  selectionChanged();options.onSelectionStart?.();renderSelection();
 }
 for(const [index,row]of rows.entries()){
  const item=records.row(row,index),tr=item.row,cell=item.actions;tr.classList.add('record-row');tr.dataset.recordIndex=String(index);
  if(signal==='logs'){
   if(target===surround.target&&Discovery.sameLog(row,surround.anchor)){tr.classList.add('log-anchor');item.time.append(node('span','Anchor fields match','anchor-badge'))}
   const nearby=node('button','Surrounding logs');nearby.setAttribute('aria-label','Surrounding logs for '+row.ServiceName+' at '+row.timestamp_ns);
   nearby.onclick=()=>{if(canUseRecord())openSurrounding(target===surround.target?surround.origin:origin,row,target===surround.target?surround.basis:basis,nearby,origin,basis)};cell.append(nearby);
  }
  if(typeof row.TraceId==='string'&&/^[a-f0-9]{32}$/.test(row.TraceId)&&!/^0+$/.test(row.TraceId)){
   const expand=node('button','Expand trace');expand.setAttribute('aria-label','Expand trace '+row.TraceId);expand.title='Trace '+row.TraceId;
   expand.onclick=()=>{if(canUseRecord())openTrace(origin,row.TraceId,row.SpanId,basis,expand)};cell.append(expand);
  }else cell.append(node('span','No valid reported trace ID','record-identity-missing'));
  if(options.onFacet||options.resultToken||readOnly&&savedPage()&&target===primary()){
   const input=node('input');input.type='checkbox';input.dataset.recordSelection='true';input.setAttribute('aria-label','Select captured record '+(index+1));input.onchange=()=>selectRecord(index,input);inputs.set(index,input);cell.prepend(input);
  }
  const inspectButton=node('button','Inspect');inspectButton.setAttribute('aria-label','Inspect record');inspectButton.dataset.inspectRecord=String(index);inspectButton.setAttribute('aria-pressed','false');
  inspectButton.onclick=()=>{
   if(!canUseRecord())return;
   const body=panel.querySelector('.ui-inspector-body'),scroll=body.scrollTop;
   for(const detail of target.querySelectorAll('.record-detail-row'))detail.hidden=true;
   item.detailRow.hidden=false;item.detailCell.append(panel);body.scrollTop=scroll;
   for(const button of target.querySelectorAll('[data-inspect-record]'))button.setAttribute('aria-pressed',String(button===inspectButton));
   const key=String(index),description=RecordPresentation.description(row);
   if(inspect.has(key)){inspect.show({key,title:signal==='logs'?'Selected log':'Selected span record',description,opener:inspectButton});return}
   const content=RecordPresentation.detail(row,signal,actions,rows),scope=node('details');
   scope.append(node('summary','Record scope and evidence basis'),node('p','Retained record inspection. Trace expansion and surrounding logs are separate reads.'),node('pre',JSON.stringify({question:origin,basis},null,2)));content.append(scope);
   const follow=[...cell.querySelectorAll('button')].filter(button=>button!==inspectButton).map(button=>{
    const action=node('button',button.textContent);if(button.hasAttribute('aria-label'))action.setAttribute('aria-label',button.getAttribute('aria-label'));if(button.title)action.title=button.title;
    action.onclick=()=>{if(canUseRecord())button.click()};return action;
   });
   if(signal==='spans'&&actions.onTraceSearch){const find=node('button','Find traces with this operation');find.type='button';find.disabled=actions.canTraceSearch?.()===false;find.onclick=()=>{if(!canUseRecord()||actions.canTraceSearch?.()===false||!Visuals.actionVisible(find))return;try{actions.onTraceSearch(TraceSearch.seedFromRow(row),find)}catch(error){find.title=error.message}};follow.push(find)}
   inspect.show({key,title:signal==='logs'?'Selected log':'Selected span record',description,content,actions:follow,opener:inspectButton});
  };
  const input=inputs.get(index);if(input)input.after(inspectButton);else cell.prepend(inspectButton);
  records.append(tr);records.append(item.detailRow);
 }
 const layoutKey=(target.id||'records')+':'+signal,layout=options.recordLayout||layouts.get(layoutKey)||initialLayout||undefined;
 if(isList){
  const picker=node('goo-record-columns');picker.model={layout:layout||{columns:['utc','service','summary','context'],fieldColumns:[]},fields:RecordColumns.fields(rows)};
  picker.addEventListener('goo:record-columns',event=>{if(!canUseRecord()||selection.state().selected||selection.state().busy)return;const next=RecordColumns.validate(event.detail);layouts.set(layoutKey,next);options.onColumns?.(next)});layoutHost.append(picker);
 }else{
  const columns=RecordColumns.mount(layoutHost,records.element,{...(layout?{layout}:{}),onFacet:actions.onFacet,canRefine:actions.canRefine,canChange:()=>canUseRecord()&&!selection.state().selected&&!selection.state().busy,onChange:next=>{if(!canUseRecord())return;layouts.set(layoutKey,next);options.onColumns?.(next)}});
  columns.update(rows.map((record,index)=>({record,element:records.rowElements[index]})));
 }
 target.prepend(layoutHost,recordSelection);records.finish();target.append(panel);
 return {selectionState,dispose(){if(mounts.get(target)===owner)mounts.delete(target)}};
}
