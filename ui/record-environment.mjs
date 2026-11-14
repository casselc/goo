import {object as closed} from './intent-shape.mjs';
/* Pure wire/selection helpers. No reads, navigation, state, or publication. */
const version='goo/environment-selector/v1',currentKey='deployment.environment.name',legacyKey='deployment.environment';
const own=(object,key)=>Object.prototype.hasOwnProperty.call(object,key);
function object(value,keys){return closed(value,keys,'Invalid environment selector fields.')}
function normalizeScope(value){object(value,['mode','value']);if(!['named','empty','unreported'].includes(value.mode)||own(value,'value')&&typeof value.value!=='string')throw Error('Invalid environment scope.');const name=value.value??'';if(value.mode==='named'){if(!name||name.includes('\0')||!name.isWellFormed()||new TextEncoder().encode(name).length>128)throw Error('Invalid environment name.');return {mode:'named',value:name}}if(name!=='')throw Error('Only named scope accepts a value.');return {mode:value.mode}}
function normalize(value){
 object(value,['version','policy','scope']);if(value.version!==version||!['current-only','effective'].includes(value.policy))throw Error('Unsupported environment selector.');
 let scope;
 if(value.scope?.mode==='all'){object(value.scope,['mode','value']);if(own(value.scope,'value')&&value.scope.value!=='')throw Error('All environments accepts no name.');scope={mode:'all'}}
 else scope=normalizeScope(value.scope);
 if(value.policy==='current-only'&&scope.mode!=='named')throw Error('Current-only selection requires a name.');return {version,policy:value.policy,scope};
}
function effective(scope){return normalize({version,policy:'effective',scope})}
function current(name){return normalize({version,policy:'current-only',scope:{mode:'named',value:name}})}
function equal(a,b){try{return JSON.stringify(normalize(a))===JSON.stringify(normalize(b))}catch{return false}}
function matches(selector,resource){const value=normalize(selector);if(value.scope.mode==='all')return true;if(!resource||typeof resource!=='object'||Array.isArray(resource)||Object.values(resource).some(v=>typeof v!=='string'))return false;const c=own(resource,currentKey),l=own(resource,legacyKey);if(value.policy==='current-only')return c&&resource[currentKey]===value.scope.value;if(!c&&!l)return value.scope.mode==='unreported';const winning=c?resource[currentKey]:resource[legacyKey];return winning===''?value.scope.mode==='empty':value.scope.mode==='named'&&value.scope.value===winning}
function isAll(question){try{return selection(question).scope.mode==='all'}catch{return false}}
function requireNarrow(question){const value=selection(question);if(value.scope.mode==='all')throw Error('Narrow to one environment before using this workflow.');return value}
function observe(resource){
 if(!resource||typeof resource!=='object'||![Object.prototype,null].includes(Object.getPrototypeOf(resource))||Object.getOwnPropertySymbols(resource).length||Object.values(Object.getOwnPropertyDescriptors(resource)).some(d=>!d.enumerable||!own(d,'value')||typeof d.value!=='string'))throw Error('Reported environment metadata unavailable.');
 const c=own(resource,currentKey),l=own(resource,legacyKey),winningKey=c?currentKey:l?legacyKey:null,value=winningKey?resource[winningKey]:null;
 const scope={mode:value===null?'unreported':value===''?'empty':'named',...(value?{value}:{})};
 return {scope,winningKey,current:c?resource[currentKey]:null,legacy:l?resource[legacyKey]:null,conflict:c&&l&&resource[currentKey]!==resource[legacyKey]};
}
function observedLabel(fact){return fact.scope.mode==='named'?'Environment '+JSON.stringify(fact.scope.value):fact.scope.mode==='empty'?'Empty environment label':'Unreported environment'}
function observedQuestion(question,resource){if(!isAll(question)||question.signal!=='logs')throw Error('Choose an All-environments log question.');return {...structuredClone(question),environment:'',environmentSelector:effective(observe(resource).scope)}}
function assertPreview(selector,scope){const expected=effective(normalizeScope(scope));if(!equal(selector,expected))throw Error('Returned environment selector differs from the requested preview.');return normalize(selector)}
function label(selector){const value=normalize(selector),scope=value.scope;return scope.mode==='all'?'All environments':scope.mode==='unreported'?'Unreported environment':scope.mode==='empty'?'Empty environment label':'Environment '+JSON.stringify(scope.value)+(value.policy==='current-only'?' · current key only':' · current/legacy precedence')}
// Preserve representation for immutable question identity; normalize only for checks.
function selection(question){
  if(!question||typeof question!=='object'||Array.isArray(question))throw Error('An explicit record environment is required.');
  if(question.environmentSelector!=null){if(question.environment!==undefined&&question.environment!=='')throw Error('Choose an environment name or selector, not both.');return normalize(question.environmentSelector)}
  return current(question.environment);
}
function questionFields(question){
  selection(question);
  return {...(own(question,'environment')?{environment:question.environment}:{}),...(own(question,'environmentSelector')?{environmentSelector:structuredClone(question.environmentSelector)}:{})};
}
function sameQuestionEnvironment(a,b){try{return equal(selection(a),selection(b))}catch{return false}}
function questionLabel(question){return question.environmentSelector!=null?label(question.environmentSelector):question.environment}
// Scalar-only adapters may project a name only after this exact policy proof.
function currentName(question){try{const value=selection(question);return value.policy==='current-only'&&value.scope.mode==='named'?value.scope.value:null}catch{return null}}
function hasQuestionSelector(question){return !!question&&typeof question==='object'&&(question.environmentSelector!=null||['explore','comparison','explanation','baseline','selection'].some(key=>hasQuestionSelector(question[key])))}
export {version,currentKey,legacyKey,normalizeScope,normalize,effective,current,equal,matches,assertPreview,label,selection,questionFields,sameQuestionEnvironment,questionLabel,currentName,hasQuestionSelector,isAll,requireNarrow,observe,observedLabel,observedQuestion};
