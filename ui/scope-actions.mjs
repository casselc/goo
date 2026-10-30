/* Pure host-owned staging. A binding is a retained capability, not event
   authentication. Session owns execution epochs; this module performs no IO. */
import {isSnapshot,getSnapshotInfo} from './snapshots.mjs';
import {identity as questionIdentity} from './question-session.mjs';
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import {edges as coefficientEdges} from './numeric-bins.mjs';
import {durationRangeFacts} from './histogram-values.mjs';

const bindings=new WeakMap(),own=(o,k)=>Object.prototype.hasOwnProperty.call(o,k);
const record=value=>value!==null&&typeof value==='object'&&!Array.isArray(value);
const slots=Object.freeze({groups:'ranked',intervals:'time-series',summary:'summary',histogram:'histogram'});
const baseKeys=['type','datasetId','revision','scopeId','generation'];
export class ScopeActionError extends TypeError {constructor(code,message){super(message);this.name='ScopeActionError';this.code=code}}
const fail=(code,message)=>{throw new ScopeActionError(code,message)};
const refusal=(code,reason)=>({ok:false,code,reason});
function closed(value,keys){
 if(!record(value)||![Object.prototype,null].includes(Object.getPrototypeOf(value)))fail('invalid-options','Use a plain options object');
 const out={};for(const key of Reflect.ownKeys(value)){if(typeof key!=='string'||!keys.includes(key))fail('invalid-options','Unknown options field');const d=Object.getOwnPropertyDescriptor(value,key);if(!d||!own(d,'value')||!d.enumerable)fail('invalid-options','Options cannot contain accessors or hidden fields');Object.defineProperty(out,key,{value:d.value,enumerable:true})}return out;
}
function plain(value,{allowUndefined=false,maxBytes=16384,maxNodes=256,maxDepth=8}={}){
 const path=new Set(),encoder=new TextEncoder();let nodes=0,bytes=0;
 const add=value=>{bytes+=value;if(bytes>maxBytes)fail('data-budget','Supplied action exceeds its data budget')};
 function visit(v,depth){
  if(++nodes>maxNodes||depth>maxDepth)fail('data-budget','Supplied action exceeds its data budget');
  if(v===null){add(4);return}
  if(typeof v==='string'){if(v.length>maxBytes)fail('data-budget','Supplied action exceeds its data budget');add(encoder.encode(v).length+2);return}
  if(typeof v==='number'){if(!Number.isFinite(v))fail('invalid-data','Only finite JSON numbers are allowed');add(24);return}
  if(typeof v==='boolean'){add(5);return}
  if(v===undefined&&allowUndefined){add(1);return}
  if(typeof v!=='object')fail('invalid-data','Actions must contain plain data only');
  const array=Array.isArray(v),prototype=Object.getPrototypeOf(v);
  if(array?prototype!==Array.prototype:prototype!==Object.prototype&&prototype!==null)fail('invalid-data','Actions must contain plain JSON objects and arrays');
  if(path.has(v))fail('invalid-data','Actions cannot contain cycles');path.add(v);
  const keys=Reflect.ownKeys(v),length=array?Object.getOwnPropertyDescriptor(v,'length')?.value:0;
  if(keys.length-(array?1:0)>maxNodes-nodes)fail('data-budget','Supplied action exceeds its data budget');
  if(array&&keys.length!==length+1)fail('invalid-data','Action arrays must be dense');
  for(const key of keys){if(array&&key==='length')continue;if(typeof key!=='string')fail('invalid-data','Actions cannot contain symbol fields');const d=Object.getOwnPropertyDescriptor(v,key);if(!d||!own(d,'value')||!d.enumerable)fail('invalid-data','Actions cannot contain accessors or hidden fields');if(array&&(!/^(?:0|[1-9]\d*)$/.test(key)||Number(key)>=length))fail('invalid-data','Action arrays cannot contain named fields');add(key.length+2);visit(d.value,depth+1)}
  path.delete(v);
 }
 visit(value,0);return value;
}
function freeze(value){if(value&&typeof value==='object'&&!Object.isFrozen(value)){for(const child of Object.values(value))freeze(child);Object.freeze(value)}return value}
function checkedQuestion(question){
 plain(question,{allowUndefined:true,maxBytes:65536,maxNodes:2048,maxDepth:24});
 try{validateRecordQuestion(question);validatePredicate(question.where,question.signal)}catch(error){fail('invalid-question',error.message)}
}
const epochValue=value=>Number.isSafeInteger(value)&&value>=0;
const sameIdentity=(intent,snapshot)=>intent.datasetId===snapshot.id&&intent.revision===snapshot.revision&&intent.scopeId===snapshot.scope.id&&intent.generation===snapshot.scope.generation;
function exactKeys(value,keys){return record(value)&&Object.keys(value).length===keys.length&&keys.every(key=>own(value,key))}
function primitive(value){return value===null||typeof value==='string'}
const op=value=>value===null?'missing':'eq';

export function createActionContext(configuration){
 const config=closed(configuration,['epoch','active','canStage']);
 for(const key of ['epoch','active','canStage'])if(typeof config[key]!=='function')fail('invalid-options','Action context requires synchronous epoch, active and canStage callbacks');
 const owner={},roots=new Map();let sourceEpoch=null,currentSource=null;
 const epoch=()=>{const value=config.epoch();if(!epochValue(value))fail('owner-state','Execution epoch must be a nonnegative safe integer');return value};
 const keyFor=(role,slot)=>role+':'+slot;
 function make(entry){const binding=Object.freeze({role:entry.role,slot:entry.slot,epoch:entry.epoch,snapshot:entry.snapshot});entry.binding=binding;entry.owner=owner;entry.root??=binding;bindings.set(binding,entry);return binding}
 function status(binding){
  const entry=bindings.get(binding);
  if(!entry||entry.owner!==owner)return refusal('foreign-binding','This action does not belong to this view');
  if(roots.get(keyFor(entry.role,entry.slot))!==entry.root)return refusal('stale-binding','This view binding has been replaced or withdrawn');
  if(epoch()!==entry.epoch)return refusal('stale-epoch','The applied question has changed');
  return {ok:true,entry};
 }
 function bind(configuration){
  const c=closed(configuration,['role','slot','epoch','snapshot','question','coherent']);
  if(!['current','baseline'].includes(c.role)||typeof c.slot!=='string'||!own(slots,c.slot)||typeof c.coherent!=='boolean')fail('invalid-binding','Bind an explicit role, slot and coherence state');
  if(!isSnapshot(c.snapshot))fail('snapshot-required','Actions require an admitted snapshot handle');
  if(c.snapshot.shape!==slots[c.slot])fail('slot-shape','Snapshot shape does not match its action slot');
  if(c.snapshot.scope.signal==='metrics')fail('scope-family','Instrument observations cannot stage record-population actions');
  if(!epochValue(c.epoch)||c.epoch!==epoch())fail('stale-epoch','Cannot bind an old execution epoch');
  checkedQuestion(c.question);
  if(questionIdentity(c.question)!==questionIdentity(c.snapshot.scope.selector))fail('question-mismatch','Binding question does not match the snapshot question');
  const info=getSnapshotInfo(c.snapshot),question=freeze(structuredClone(c.question));
  if(c.epoch!==sourceEpoch){sourceEpoch=c.epoch;currentSource=null}
  if(c.role==='current'&&info.entry==='source'){if(currentSource!==null&&currentSource!==info.sourceInstance)fail('source-mismatch','Current view slots must belong to one source execution');currentSource=info.sourceInstance}
  const binding=make({...c,question,sourceInstance:info.sourceInstance,entry:info.entry,root:null});
  roots.set(keyFor(c.role,c.slot),binding);return binding;
 }
 function project(parentBinding,subset){
  const current=status(parentBinding);if(!current.ok)fail(current.code,current.reason);
  const parent=current.entry;
  if(parent.root!==parentBinding)fail('projection-parent','Only direct projections of a root binding are supported');
  if(parent.slot==='summary')fail('projection-shape','Population summaries cannot be projected into rows');
  if(parent.slot==='histogram')fail('projection-shape','Histogram geometry requires the full fixed bin schema');
  if(subset===parent.snapshot)return parentBinding;
  if(!isSnapshot(subset)||!['ranked','time-series'].includes(subset.shape))fail('projection-lineage','Use a direct admitted row projection');
  const info=getSnapshotInfo(subset),original=parent.snapshot;
  if(info.entry!=='projection'||info.parent?.id!==original.id||info.parent?.revision!==original.revision||info.sourceInstance!==parent.sourceInstance||subset.scope!==original.scope||subset.fields!==original.fields||subset.dimensions!==original.dimensions||subset.interval!==original.interval||subset.rows.length>=original.rows.length||subset.rows.some(row=>!original.rows.includes(row)))fail('projection-lineage','Projection does not belong to this bound parent');
  return make({...parent,snapshot:subset,entry:'projection',root:parent.root,parentSnapshot:original});
 }
 function actionGuard(binding){
   const current=status(binding);if(!current.ok)return current;
   const entry=current.entry,root=bindings.get(entry.root);
   if(entry.role!=='current')return refusal('read-only-role','Baseline data is read-only');
   if(entry.slot==='summary')return refusal('read-only-slot','Population summaries do not identify groups or intervals');
   const active=config.active(),canStage=config.canStage();
   if(typeof active!=='boolean'||typeof canStage!=='boolean')return refusal('owner-state','Action owner callbacks must return synchronous booleans');
   if(!active)return refusal('inactive','This view is not active');
   if(!canStage)return refusal('not-applied','Run the staged question before refining its results');
   if(!entry.coherent||!root.coherent)return refusal('incoherent','Source reads disagree; run the question again');
   const rootInfo=getSnapshotInfo(root.snapshot);
   if(root.entry!=='source'||rootInfo.entry!=='source'||!entry.sourceInstance||entry.sourceInstance!==rootInfo.sourceInstance)return refusal('source-required','Held data cannot stage a source question');
   return {ok:true,entry,rootInfo};
 }
 function resolve(binding,intent){
  try{
   const guarded=actionGuard(binding);if(!guarded.ok)return guarded;
   const {entry,rootInfo}=guarded;
   if(entry.slot==='histogram')return refusal('intent-type','Histogram ranges require checked source planning');
   plain(intent);
   const snapshot=entry.snapshot;
   if(!sameIdentity(intent,snapshot))return refusal('identity-mismatch','Selection belongs to a different snapshot');
   const question=structuredClone(entry.question);let label;
   if(entry.slot==='groups'){
    const dimensions=snapshot.dimensions??[],groupBy=question.groupBy??[];let values;
    if(!dimensions.length||dimensions.length!==groupBy.length)return refusal('dimension-mismatch','Group dimensions do not match the captured question');
    if(intent.type==='scope-filter'){
     if(!exactKeys(intent,[...baseKeys,'field','value','op'])||!dimensions.includes(intent.field)||!primitive(intent.value)||intent.op!==op(intent.value)||!snapshot.rows.some(row=>row.state!=='unknown'&&own(row.values,intent.field)&&row.values[intent.field]===intent.value))return refusal('invalid-membership','Facet value is not a known supplied group dimension');
     values=[{field:intent.field,value:intent.value}];
    }else if(intent.type==='scope-group'){
     const row=snapshot.rows.find(row=>row.key===intent.key);
     if(!exactKeys(intent,[...baseKeys,'key','values'])||!row||row.state==='unknown'||!Array.isArray(intent.values)||intent.values.length!==dimensions.length||!dimensions.every((field,i)=>exactKeys(intent.values[i],['field','value','op'])&&intent.values[i].field===field&&primitive(intent.values[i].value)&&intent.values[i].value===row.values[field]&&intent.values[i].op===op(row.values[field])))return refusal('invalid-membership','Group tuple does not match a known supplied row');
     values=intent.values;
    }else return refusal('intent-type','Use a group or facet selection for this view');
    const predicates=values.map(({field,value})=>({field:structuredClone(groupBy[dimensions.indexOf(field)]),op:op(value),...(value===null?{}:{value})}));
    question.where={all:[...(question.where?.all??(question.where?[question.where]:[])),...predicates]};label='Group filter';
   }else{
    if(!exactKeys(intent,[...baseKeys,'key','fromMs','toMs'])||intent.type!=='scope-window')return refusal('intent-type','Use a supplied interval selection for this view');
    const row=snapshot.rows.find(row=>row.key===intent.key);
    if(!row||!['observed','no_data'].includes(row.state)||intent.fromMs!==row.values.fromMs||intent.toMs!==row.values.toMs)return refusal('invalid-membership','Interval does not match a known supplied row');
    const from=Number(intent.fromMs),to=Number(intent.toMs);
    if(!Number.isSafeInteger(from)||!Number.isSafeInteger(to)||from<question.fromMs||to>question.toMs||from>=to)return refusal('invalid-window','Selected interval is outside the captured question');
    question.fromMs=from;question.toMs=to;label='Interval';
   }
   question.offset=0;checkedQuestion(question);
   return {ok:true,question:freeze(question),label,role:entry.role,slot:entry.slot,epoch:entry.epoch,sourceId:rootInfo.sourceId,sourceInstance:rootInfo.sourceInstance};
  }catch(error){return refusal(error instanceof ScopeActionError?error.code:'invalid-intent',error.message||'Unable to resolve this selection')}
 }
 function inspect(binding,intent){
  try{
   const guarded=actionGuard(binding);if(!guarded.ok)return guarded;
   if(guarded.entry.slot==='histogram')return refusal('intent-type','Histogram ranges require checked source planning');
   const {entry,rootInfo}=guarded;plain(intent);
   if(!exactKeys(intent,[...baseKeys,'key'])||intent.type!=='select')return refusal('intent-type','Inspect a supplied selected datum');
   if(!sameIdentity(intent,entry.snapshot))return refusal('identity-mismatch','Selection belongs to a different snapshot');
   const row=entry.snapshot.rows.find(row=>row.key===intent.key);
   if(!row||row.state==='unknown')return refusal('invalid-membership','Selection does not match a known supplied datum');
   // This is an aggregate datum, not verified member-record identity. Inspection
   // retains the captured question; a narrower read has its own checked owner.
   return {ok:true,question:entry.question,row,role:entry.role,slot:entry.slot,epoch:entry.epoch,sourceId:rootInfo.sourceId,sourceInstance:rootInfo.sourceInstance};
  }catch(error){return refusal(error instanceof ScopeActionError?error.code:'invalid-intent',error.message||'Unable to inspect this selection')}
 }
 // Resolve geometry for the existing Go planner, never construct a numeric
 // predicate here. Its final max+1 edge is not a representable threshold.
 // Async owners must check this same binding/intent again after awaiting a plan.
 function histogram(binding,intent){
  try{
   const guarded=actionGuard(binding);if(!guarded.ok)return guarded;
   const {entry,rootInfo}=guarded;plain(intent);
   if(entry.slot!=='histogram'||intent.type!=='scope-histogram'||!exactKeys(intent,[...baseKeys,'fromKey','toKey']))return refusal('intent-type','Choose a supplied histogram range');
   const snapshot=entry.snapshot;
   if(snapshot.histogram?.kind!=='numeric')return refusal('histogram-kind','Numeric planning requires a numeric histogram');
   if(!sameIdentity(intent,snapshot))return refusal('identity-mismatch','Selection belongs to a different snapshot');
   const rows=snapshot.rows,a=rows.findIndex(row=>row.key===intent.fromKey),b=rows.findIndex(row=>row.key===intent.toKey);
   if(a<0||b<0||a>b)return refusal('invalid-membership','Range endpoints must identify ordered supplied histogram bins');
   const fromBin=a,toBin=b+1,question=entry.question;
   if(!question.numeric||question.numericHistogram?.binning!==snapshot.histogram?.binning||rows.slice(fromBin,toBin).some(row=>!['observed','no_data'].includes(row.state)))return refusal('invalid-membership','Range does not match known numeric histogram geometry');
   const edges=coefficientEdges(question.numeric.type);
   if(rows.length!==edges.length-1)return refusal('invalid-membership','Histogram requires the complete fixed bin geometry');
   return {ok:true,question,fromBin,toBin,from:edges[fromBin],to:edges[toBin],role:entry.role,slot:entry.slot,epoch:entry.epoch,sourceId:rootInfo.sourceId,sourceInstance:rootInfo.sourceInstance};
  }catch(error){return refusal(error instanceof ScopeActionError?error.code:'invalid-intent',error.message||'Unable to plan this histogram range')}
 }
 function duration(binding,intent){
  try{
   const guarded=actionGuard(binding);if(!guarded.ok)return guarded;
   const {entry}=guarded;plain(intent);
   if(entry.slot!=='histogram'||intent.type!=='scope-histogram'||!exactKeys(intent,[...baseKeys,'fromKey','toKey']))return refusal('intent-type','Choose a supplied duration range');
   const snapshot=entry.snapshot;
   if(snapshot.histogram?.kind!=='span-duration')return refusal('histogram-kind','Duration selection requires captured span-duration bins');
   if(!sameIdentity(intent,snapshot))return refusal('identity-mismatch','Selection belongs to a different snapshot');
   const range=durationRangeFacts(snapshot,intent.fromKey,intent.toKey);
   return Object.freeze({ok:true,question:entry.question,...range});
  }catch(error){return refusal(error instanceof ScopeActionError?error.code:'invalid-intent',error.message||'Unable to resolve this duration range')}
 }
 // Manual exact controls already validate their candidate question. This checks
 // the captured execution before every effect without inventing bin membership
 // for arbitrary user-entered duration bounds.
 function canAct(binding){try{return actionGuard(binding).ok}catch{return false}}
 return Object.freeze({bind,project,resolve,inspect,histogram,duration,canAct,revoke(){roots.clear();sourceEpoch=null;currentSource=null}});
}
