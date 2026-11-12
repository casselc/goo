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
import './record-span-waterfall.mjs';
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

// Display order may differ from delivery order. Only page positions identify the
// controller's occurrences; DOM order and reported IDs never become Keep indexes.
export function indexRecordItems(items,size){
 if(!Array.isArray(items)||!Number.isSafeInteger(size)||size<0||items.length!==size)throw Error('Record hosts must match the admitted page.');
 const indexed=new Array(size);
 for(const item of items){const index=item?.index;if(!Number.isInteger(index)||index<0||index>=size||indexed[index]!==undefined)throw Error('Use each original page position exactly once.');indexed[index]=item}
 return Object.freeze(indexed);
}
function surface(target,rows,signal,options){
 if(['record-list','record-span-waterfall'].includes(options.recordKind)){
  const list=node(options.recordKind==='record-span-waterfall'?'goo-record-span-waterfall':'goo-record-list');if(options.recordKind==='record-span-waterfall')list.summaryProvided=options.summaryProvided===true;list.page=options.recordPage;list.display=options.recordLayout;target.append(list);
  const handles=list.recordItems(),items=indexRecordItems(handles,rows.length),page=list.page,display=list.display;
  for(const item of items){item.actionsHost.hidden=false;item.details.hidden=true;item.detailHost.classList.add('record-detail-row');item.detailHost.hidden=true}
  return {viewport:list,current:()=>list.page===page&&list.display===display&&list.recordItems()===handles,
   rowElements:items.map(item=>item.element),
   row(record,index){const item=items[index];return {row:item.element,time:item.timeHost||item.element.querySelector('.record-list-meta')||item.element,actions:item.actionsHost,detailRow:item.detailHost,detailCell:item.detailHost,timingButton:item.timingButton}},
   append(){},finish(){}
  };
 }
 const records=RecordPresentation.table({signal,summaryCompact:true,mobileCompact:true,labels:['UTC','Service',signal==='spans'?'Span / duration':'Severity / body','Actions'],caption:'Captured '+signal+' · inspection page; full totals shown separately',label:'Captured '+signal+' inspection page'});
 records.element.className='telemetry-table';records.header.className='record-columns';
 const rowElements=[];
 return {...records,rowElements,current:()=>true,
  row(record){const result=records.row(record),detailRow=node('tr'),detailCell=node('td');detailRow.className='record-detail-row';detailRow.hidden=true;detailCell.colSpan=4;detailRow.append(detailCell);rowElements.push(result.row);return {row:result.row,time:result.cells[0],actions:result.cells[3],detailRow,detailCell}},
  append(element){records.body.append(element)},finish(){target.append(records.viewport)}
 };
}

export function mount(target,rows,signal,origin,basis,options={},context={}){
 const {active=()=>true,attributeOptions=()=>({}),readOnly=false,savedPage=()=>null,primary=()=>null,openTrace,openSurrounding,layouts=new Map(),initialLayout,surrounding=()=>({})}=context;
 const inspectorOwners=new Map();
 const surround=surrounding(),owner={};mounts.set(target,owner);target.replaceChildren();
 if(!rows.length){const empty=node('goo-result-state','No captured records in this page; absence is not health evidence.');empty.state='empty';target.append(empty);return}
 const records=surface(target,rows,signal,options),{viewport}=records,isPassive=['record-list','record-span-waterfall'].includes(options.recordKind);
 const canUseRecord=()=>mounts.get(target)===owner&&active()&&target.isConnected&&viewport.isConnected&&target.contains(viewport)&&records.current()&&options.canUseRecord?.()!==false;
 const originalOptions=attributeOptions(target,signal,options,origin,basis);
 const actions={...originalOptions,
  canRefine:()=>canUseRecord()&&originalOptions.canRefine?.()!==false,
  canPlot:()=>canUseRecord()&&(originalOptions.canPlot||originalOptions.canRefine)?.()!==false,
  canTraceSearch:()=>canUseRecord()&&(originalOptions.canTraceSearch||originalOptions.canRefine)?.()!==false,
  onFilter:originalOptions.onFilter?((...args)=>{if(actions.canRefine())originalOptions.onFilter(...args)}):undefined,
  onBreakdown:originalOptions.onBreakdown?((...args)=>{if(actions.canRefine())originalOptions.onBreakdown(...args)}):undefined,
  onTraceSearch:originalOptions.onTraceSearch?((...args)=>{if(actions.canTraceSearch())originalOptions.onTraceSearch(...args)}):undefined,
  onFacet:originalOptions.onFacet?((...args)=>actions.canRefine()?originalOptions.onFacet(...args):false):undefined,
  onPlot:originalOptions.onPlot?((field,button)=>{if(actions.canPlot())originalOptions.onPlot(field,button)}):undefined
 };
 const panel=node('section'),inspect=UI.inspector(panel,{onClose:()=>{for(const extension of inspectorOwners.values())extension.suspend?.();const row=panel.closest('.record-detail-row');if(row)row.hidden=true;for(const button of target.querySelectorAll('[data-inspect-record]'))button.setAttribute('aria-pressed','false')}});
 // Opening a record is an explicit application navigation gesture. Keep the
 // supplied inspector passive; only this current owner moves focus into detail.
 const inspectionHeading=panel.querySelector('.ui-inspector-heading')?.querySelector('h4');
 function focusInspection(opener){if(!inspectionHeading||!canUseRecord()||panel.hidden||!UI.actionVisible(opener))return;inspectionHeading.tabIndex=-1;inspectionHeading.focus({preventScroll:true});inspectionHeading.scrollIntoView({block:'nearest',inline:'nearest'});}
 const selection=createSelection(rows.length),selectionState=selection.state,inputs=new Map(),recordSelection=node('section',undefined,'record-selection');
 recordSelection.hidden=true;recordSelection.setAttribute('aria-label','Selected '+signal+' records');
 const layoutHost=node('div',undefined,'record-fields-toolbar');
 function selectionChanged(){
  const state=selection.state();for(const [index,input]of inputs){input.checked=selection.has(index);input.disabled=state.submitted||state.busy||!input.checked&&state.selected>=20}
  if(isPassive){layoutHost.inert=state.selected>0||state.busy;layoutHost.setAttribute('aria-disabled',String(layoutHost.inert))}
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
   onFacet:actions.onFacet?((...args)=>allowed()?actions.onFacet(...args):false):undefined,canRefine:allowed
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
  const item=records.row(row,index),tr=item.row,cell=node('div',undefined,'record-row-actions'),secondary=[];
  // The shared menu closes after activation; its summary remains a visible
  // return-focus origin. Admission still belongs to the actual clicked button.
  let secondaryOpener=null;
  const secondaryAction=(button,run)=>{secondary.push({button,run});button.onclick=()=>{if(canUseRecord()&&!button.disabled&&UI.actionVisible(button))run(secondaryOpener)};cell.append(button)};cell.setAttribute('role','group');cell.setAttribute('aria-label','Record actions');item.actions.append(cell);tr.classList.add('record-row');tr.dataset.recordIndex=String(index);
  if(signal==='logs'){
   if(target===surround.target&&Discovery.sameLog(row,surround.anchor)){tr.classList.add('log-anchor');item.time.append(node('span','Anchor fields match','anchor-badge'))}
   const nearby=node('button','Surrounding logs');nearby.setAttribute('aria-label','Surrounding logs for '+row.ServiceName+' at '+row.timestamp_ns);
   secondaryAction(nearby,opener=>openSurrounding(target===surround.target?surround.origin:origin,row,target===surround.target?surround.basis:basis,opener,origin,basis));
  }
  if(typeof row.TraceId==='string'&&/^[a-f0-9]{32}$/.test(row.TraceId)&&!/^0+$/.test(row.TraceId)){
   const expand=node('button','Expand trace');expand.dataset.expandTrace=String(index);expand.setAttribute('aria-label','Expand trace '+row.TraceId);expand.title='Trace '+row.TraceId;
   secondaryAction(expand,opener=>openTrace(origin,row.TraceId,row.SpanId,basis,opener,{...(options.onOpenTraceView?{onOpenView:(view,opener)=>canUseRecord()?options.onOpenTraceView(view,opener):false,canOpenView:()=>canUseRecord()&&options.canOpenTraceView?.()!==false}:{})}));
  }else cell.append(node('span','No valid reported trace ID','record-identity-missing'));
  if(options.onFacet||options.resultToken||readOnly&&savedPage()&&target===primary()){
   const input=node('input');input.type='checkbox';input.dataset.recordSelection='true';input.setAttribute('aria-label','Select captured record '+(index+1));input.onchange=()=>selectRecord(index,input);inputs.set(index,input);cell.prepend(input);
  }
  const inspectButton=item.timingButton||node('button','Inspect');if(item.timingButton){const label=inspectButton.getAttribute('aria-label')||RecordPresentation.describe(row,signal).message;if(!label.startsWith('Inspect supplied span'))inspectButton.setAttribute('aria-label','Inspect supplied span · '+label);if(!inspectButton.title.startsWith('Inspect supplied span'))inspectButton.title='Inspect supplied span · '+(inspectButton.title||'')}else inspectButton.setAttribute('aria-label','Inspect record');inspectButton.dataset.inspectRecord=String(index);inspectButton.setAttribute('aria-pressed','false');
  inspectButton.onclick=()=>{
   if(!canUseRecord())return;
   for(const extension of inspectorOwners.values())extension.suspend?.();
   const body=panel.querySelector('.ui-inspector-body'),scroll=body.scrollTop;
   for(const detail of target.querySelectorAll('.record-detail-row'))detail.hidden=true;
   item.detailRow.hidden=false;item.detailCell.append(panel);body.scrollTop=scroll;
   for(const button of target.querySelectorAll('[data-inspect-record]'))button.setAttribute('aria-pressed',String(button===inspectButton));
   const key=String(index),description=RecordPresentation.description(row);
   if(inspect.has(key)){inspect.show({key,title:signal==='logs'?'Selected log':'Selected span record',description,opener:inspectButton});inspectorOwners.get(key)?.resume?.();focusInspection(inspectButton);return}
   const content=RecordPresentation.detail(row,signal,actions,rows),scope=node('details');
   scope.append(node('summary','Record scope and evidence basis'),node('p','Retained record inspection. Trace expansion and surrounding logs are separate reads.'),node('pre',JSON.stringify({question:origin,basis},null,2)));content.append(scope);
   const follow=secondary.map(({button,run})=>{
    const action=node('button',button.textContent);if(button.hasAttribute('aria-label'))action.setAttribute('aria-label',button.getAttribute('aria-label'));if(button.title)action.title=button.title;
    action.onclick=()=>{if(canUseRecord()&&!action.disabled&&UI.actionVisible(action))run(action)};return action;
   });
   if(signal==='spans'&&actions.onTraceSearch){const find=node('button','Find traces with this operation');find.type='button';find.disabled=actions.canTraceSearch?.()===false;find.onclick=()=>{if(!canUseRecord()||actions.canTraceSearch?.()===false||!Visuals.actionVisible(find))return;try{actions.onTraceSearch(TraceSearch.seedFromRow(row),find)}catch(error){find.title=error.message}};follow.push(find)}
   const extension=options.onInspectRecord?.({content,index,opener:inspectButton,current:()=>canUseRecord()&&inspectButton.getAttribute('aria-pressed')==='true'&&!item.detailRow.hidden&&panel.dataset.key===key});if(extension?.dispose)inspectorOwners.set(key,extension);
   inspect.show({key,title:signal==='logs'?'Selected log':'Selected span record',description,content,actions:follow,opener:inspectButton});focusInspection(inspectButton);
  };
  const input=inputs.get(index);if(!item.timingButton){if(input)input.after(inspectButton);else cell.prepend(inspectButton)}
  if(secondary.length){const menu=UI.actions([],secondary.map(({button})=>button),{label:'More'});secondaryOpener=menu.querySelector('summary');secondaryOpener.setAttribute('aria-label','More record actions');cell.append(menu)}
  records.append(tr);records.append(item.detailRow);
 }
 const layoutKey=(target.id||'records')+':'+signal,layout=options.recordLayout||layouts.get(layoutKey)||initialLayout||undefined;
 if(isPassive){
  const picker=node('goo-record-columns');picker.model={layout:layout||{columns:['utc','service','summary','context'],fieldColumns:[]},fields:RecordColumns.fields(rows)};
  picker.addEventListener('goo:record-columns',event=>{if(!canUseRecord()||selection.state().selected||selection.state().busy)return;const next=RecordColumns.validate(event.detail);layouts.set(layoutKey,next);options.onColumns?.(next)});layoutHost.append(picker);
 }else{
  const columns=RecordColumns.mount(layoutHost,records.element,{...(layout?{layout}:{}),onFacet:actions.onFacet,canRefine:actions.canRefine,canChange:()=>canUseRecord()&&!selection.state().selected&&!selection.state().busy,onChange:next=>{if(!canUseRecord())return;layouts.set(layoutKey,next);options.onColumns?.(next)}});
  columns.update(rows.map((record,index)=>({record,element:records.rowElements[index]})));
 }
 target.prepend(layoutHost,recordSelection);records.finish();target.append(panel);
 return {selectionState,dispose(){for(const extension of inspectorOwners.values())extension.dispose();inspectorOwners.clear();if(mounts.get(target)===owner)mounts.delete(target)}};
}
