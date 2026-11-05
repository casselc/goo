/* Application-owned independent counted-value reads for one facet scope/pair. */
import * as FilterValues from './filter-values.mjs';
import {identity as questionIdentity} from './question-session.mjs';

export function identity(scope,field){
 if(!scope?.selected)throw Error('Apply the selected scope before reading facet values.');
 FilterValues.query(scope.selected,field);
 if(scope.baseline)FilterValues.query(scope.baseline,field);
 return JSON.stringify([questionIdentity(scope.selected),scope.baseline?questionIdentity(scope.baseline):null,field.kind,field.key]);
}

export function mount(editor,{current,pause=()=>{},initial=null}){
 let bound=initial?{scope:structuredClone(initial.scope),field:structuredClone(initial.field),key:identity(initial.scope,initial.field)}:null,disposed=false,resetting=false,states=initial?{values:structuredClone(initial.values),...(initial.baseline?{baseline:structuredClone(initial.baseline)}:{})}:{values:{state:'unloaded'}},loaders={};
 const matches=()=>{try{return !!bound&&identity(current(),bound.field)===bound.key}catch{return false}};
 const publish=()=>editor.setFacetValues(structuredClone(states));
 const source=side=>matches()?bound.scope[side]:null;
 const sides={selected:'values',baseline:'baseline'};
 const owners=Object.fromEntries(Object.entries(sides).map(([side,name])=>[side,FilterValues.mount({
  setValueLoader(load){loaders[side]=load},
  setCountedValues(value){
   if(disposed||resetting)return;
   if(!matches()){if(bound)invalidate();return;}
   if(value)states[name]=structuredClone(value);
   else if(states[name]?.state==='loading')states[name]={state:'unloaded'};
   publish();
  }
 },{current:()=>source(side)})]));
 function invalidate(){
  bound=null;resetting=true;try{for(const owner of Object.values(owners))owner.invalidate()}finally{resetting=false}
  if(!disposed)editor.setFacetValues();
 }
 async function load(field){
  if(disposed)return;
  pause();if(disposed)return;invalidate();
  let scope,key;
  try{scope=structuredClone(current());key=identity(scope,field)}catch(error){states={values:{state:'error',field:structuredClone(field),error:error.message}};publish();return}
  bound={scope,key,field:structuredClone(field)};
  states={values:{state:'loading',field:structuredClone(field)},...(scope.baseline?{baseline:{state:'loading',field:structuredClone(field)}}:{})};publish();
  await Promise.all([loaders.selected(field),...(scope.baseline?[loaders.baseline(field)]:[])]);
 }
 async function retry(side){
  if(disposed||!Object.hasOwn(sides,side)||!matches()||states[sides[side]]?.state!=='error')return;
  const expected=bound,field=structuredClone(bound.field);pause();
  if(disposed||bound!==expected||!matches()||states[sides[side]]?.state!=='error')return;
  await loaders[side](field);
 }
 editor.setValueLoader(load);
 return {load,retry,invalidate,sync(){if(bound&&!matches())invalidate()},dispose(){if(disposed)return;disposed=true;bound=null;for(const owner of Object.values(owners))owner.dispose();editor.setValueLoader(null)}};
}
