/* One execution lifecycle for record questions. No source or DOM access. */
import * as RecordEnvironment from './record-environment.mjs';
import * as FieldPresentation from './field-presentation.mjs';
const global=globalThis;
const clone=x=>structuredClone(x);
function canonical(value){if(Array.isArray(value))return value.map(canonical);if(value&&typeof value==='object')return Object.fromEntries(Object.keys(value).sort().filter(k=>value[k]!==undefined).map(k=>[k,canonical(value[k])]));return value}
function predicateIdentity(p){if(!p)return p;const out=clone(p);if(out.numeric===null)delete out.numeric;if(out.numeric?.scale===0)delete out.numeric.scale;if(out.value==='')delete out.value;if(out.all)out.all=out.all.map(predicateIdentity);if(out.any)out.any=out.any.map(predicateIdentity);return out}
function identity(q){const v=clone(q);if(v.environmentSelector!=null){v.environmentSelector=RecordEnvironment.normalize(v.environmentSelector);if(v.environment===undefined)v.environment=''}else delete v.environmentSelector;if(v.where)v.where=predicateIdentity(v.where);if(v.numeric?.scale===0)delete v.numeric.scale;for(const key of ['numeric','numericHistogram','where','summary','bucketMs','groupTimeSeries','spanStatus','reportedEdge'])if(!v[key])delete v[key];if(!v.groupBy?.length)delete v.groupBy;return JSON.stringify(canonical(v))}
function combineWhere(base,filters,join,shortcuts){const authored=base&&filters.length?{all:[base,{[join]:clone(filters)}]}:base?clone(base):filters.length?{[join]:clone(filters)}:undefined;if(!shortcuts.length)return authored;return {all:[...(authored?[authored]:[]),...clone(shortcuts)]}}
const selectorText=(...args)=>(FieldPresentation).selectorText(...args);
class Session {
  constructor(observe=()=>{}){this.state={generation:0,phase:'initial',applied:null,requested:null,bounds:null};this.observe=observe}
  snapshot(){return clone(this.state)}
  emit(action,token=null){const value={action,token,...this.snapshot()};this.observe(value);return value}
  edit(resetBounds=false){this.state.generation++;this.state.phase='draft';this.state.requested=null;if(resetBounds)this.state.bounds=null;this.emit('edit');return this.state.generation}
  stage(q){const check=new Session();check.start(q);this.edit(true);this.state.bounds={fromMs:q.fromMs,toMs:q.toMs};this.state.applied=null;this.state.requested=clone(q);this.emit('stage');return this.state.generation}
  start(q){if(!q||!['spans','logs'].includes(q.signal)||!Number.isSafeInteger(q.fromMs)||!Number.isSafeInteger(q.toMs)||q.fromMs<0||q.toMs<=q.fromMs||q.toMs-q.fromMs>7200000)throw Error('Choose a valid signal, environment and fixed window.');RecordEnvironment.selection(q);this.state.generation++;this.state.phase='running';this.state.requested=clone(q);this.state.bounds={fromMs:q.fromMs,toMs:q.toMs};this.emit('start');return this.state.generation}
  publish(token,data){if(token!==this.state.generation||this.state.phase!=='running'){this.emit('reject',token);return false}if(identity(data.question)!==identity(this.state.requested))throw Error('Returned evidence does not match the requested question.');this.state.applied={question:clone(data.question),basis:data.basis,count:data.count,state:data.state};this.state.requested=null;this.state.phase='applied';this.emit('publish',token);return true}
  fail(token){if(token!==this.state.generation||this.state.phase!=='running'){this.emit('reject',token);return false}this.state.phase='failed';this.state.requested=null;this.emit('fail',token);return true}
  cancel(){this.state.generation++;this.state.phase='cancelled';this.state.requested=null;this.emit('cancel');return this.state.generation}
 }

export {Session,identity,combineWhere,selectorText};
