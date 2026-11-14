import {node} from './components.mjs';
import {nativeControlReady,completionControlReady} from './dom-visibility.mjs';

/* Contextual authoring adapter. Planning and all effects are injected by the app. */
let serial=0;
const bytes=text=>new TextEncoder().encode(text).length;
const localDiagnostic=message=>({code:'input',phase:'input',message});
const messageOf=error=>error?.diagnostic?.message||error?.message||error?.error||'Condition could not be checked. Try again.';
const conditionHosts=value=>Array.isArray(value)?value:value?[value]:[];
// Summaries use structured compiler facts; full human diagnostics stay available.
function diagnosticPresentation(diagnostic){
 if(!diagnostic)return {summary:'',detail:''};
 const raw=diagnostic.message;
 if(!raw.includes('\n')&&Array.from(raw).length<=180)return {summary:raw,detail:''};
 const messages={syntax:'Check expression syntax',type:'Check field names and value types',unsupported:'This expression is not supported',budget:'Simplify this expression to fit the check limits',binding:'Scope changed. Check the condition again',target:'This condition cannot be planned for this scope'};
 const summary=messages[diagnostic.code]||'Condition could not be checked';
 const span=diagnostic.span,location=Number.isInteger(span?.line)&&span.line>0&&Number.isInteger(span?.column)&&span.column>0?` · line ${span.line}, column ${span.column}`:'';
 return {summary:summary+location+'.',detail:raw};
}


export function mount(host,{owner:prepare,check,onChange=()=>{},current=()=>true,conditionsHost,scopeSummary='',onMode=()=>{},onDiscard=()=>{}}={}){
 if(typeof prepare!=='function'||typeof check!=='function')throw Error('Require supplied condition owner and planning functions.');
 const id='cel-condition-'+(++serial),root=node('div',undefined,'cel-expression-editor'),modeLabel=node('label','Filter mode'),mode=node('select'),pane=node('div',undefined,'cel-expression-pane'),scope=node('p',undefined,'cel-expression-scope'),label=node('label','Additional condition'),text=node('textarea'),hint=node('small','CEL · adds to the current filters'),actions=node('div',undefined,'controls'),verify=node('button','Check condition'),clear=node('button','Clear condition'),discard=node('button','Discard expression'),lockedNote=node('small','Discard expression to edit conditions.'),feedback=node('p',undefined,'cel-expression-diagnostic'),details=node('details',undefined,'cel-expression-diagnostic-details'),detailsTitle=node('summary','Diagnostic details'),detailsBody=node('pre');
 mode.id=id+'-mode';modeLabel.htmlFor=mode.id;mode.setAttribute('aria-label','Filter mode');
 for(const [value,title]of [['conditions','Conditions'],['expression','Expression (CEL)']]){const option=node('option',title);option.value=value;mode.append(option)}
 text.id=id;text.rows=3;text.spellcheck=false;text.autocapitalize='off';text.autocomplete='off';text.setAttribute('aria-label','Additional condition (CEL)');label.htmlFor=text.id;hint.id=id+'-hint';feedback.id=id+'-diagnostic';feedback.setAttribute('role','status');feedback.setAttribute('aria-atomic','true');text.setAttribute('aria-describedby',hint.id+' '+feedback.id);text.placeholder='log.body.contains("timeout")';
 for(const button of [verify,clear,discard])button.type='button';discard.title='Discard expression text; keep the current filters.';
 details.append(detailsTitle,detailsBody);actions.append(verify,clear);pane.append(scope,label,text,hint,actions);root.append(modeLabel,mode,pane,lockedNote,discard,feedback,details);host.append(root);
 let active=true,disposed=false,owned=null,preparing=null,preparation=null,job=null,stageToken=null,verification=null,modeInvocation=null,modeValue='conditions',hasDraft=false,hasCondition=false,needsAdoption=false,diagnostic=null,lastDetail='';
 const available=()=>!disposed&&active&&current();
 const actionReady=control=>available()&&nativeControlReady(control);
 function scopeText(){return typeof scopeSummary==='function'?scopeSummary():scopeSummary}
 function display({announce=false}={}){
  root.dataset.mode=modeValue;root.dataset.phase=diagnostic||needsAdoption?'invalid':job?'checking':hasCondition?'valid':hasDraft?'draft':'clear';
  if(owned)text.placeholder=owned.snapshot().base.signal==='spans'?'span.name.contains("request")':'log.body.contains("timeout")';
  pane.hidden=modeValue!=='expression';pane.inert=pane.hidden||!available();mode.value=modeValue;mode.disabled=!available();
  for(const element of conditionHosts(conditionsHost)){element.hidden=modeValue==='expression';element.inert=modeValue==='expression'||!available()||hasDraft;element.disabled=hasDraft||!available()}
  scope.textContent=scopeText()||'';scope.hidden=!scope.textContent;
  verify.disabled=!available()||needsAdoption||!text.value.trim()||!!job||!!preparing;
  clear.disabled=!available()||needsAdoption||(!hasDraft&&!hasCondition)||!!job||!!preparing;
  lockedNote.hidden=modeValue!=='conditions'||!hasDraft;discard.hidden=!needsAdoption&&(modeValue!=='conditions'||!hasDraft);discard.disabled=!available();
  text.setAttribute('aria-invalid',String(!!diagnostic||needsAdoption));
  // Draft changes clear feedback quietly; only an explicit check announces.
  feedback.setAttribute('aria-live',announce?'polite':'off');
  const presentation=diagnosticPresentation(diagnostic);details.hidden=!presentation.detail;if(lastDetail!==presentation.detail){details.open=false;detailsBody.textContent=presentation.detail;lastDetail=presentation.detail}
  feedback.textContent=needsAdoption?'Scope changed. Discard this draft before adding a new condition.':diagnostic?presentation.summary:job?'Checking condition…':hasCondition?'Condition staged.':'';
 }
 function retire(){verification=null;modeInvocation=null;job?.controller.abort();job=null;stageToken=null;preparation?.controller.abort();preparation=null;preparing=null;owned?.dispose();owned=null}
 async function ensureOwner(){
  if(owned)return owned;
  if(preparing)return preparing;
  if(!available())return null;
  const ticket=preparation={controller:new AbortController()};
  preparing=Promise.resolve().then(()=>prepare(ticket.controller.signal)).then(result=>{
   if(preparation!==ticket||!available()){result?.dispose?.();return null}
   if(!result||typeof result.beginCheck!=='function')throw Error('Condition owner unavailable.');
   owned=result;owned.edit(text.value);owned.setMode(modeValue);owned.setVisible(modeValue==='expression'&&available());return owned;
  }).catch(error=>{if(preparation!==ticket||!available())return null;throw error}).finally(()=>{if(preparation===ticket){preparation=null;preparing=null;display()}});
  display();return preparing;
 }
 function pendingDraft(){verification=null;hasDraft=true;diagnostic=null;hasCondition=false;stageToken=null;job?.controller.abort();job=null;owned?.edit(text.value);onChange(undefined,'Check the expression before applying.',{draft:true});display()}
 mode.onchange=()=>{
  if(!actionReady(mode))return;
  modeValue=mode.value;owned?.setMode(modeValue);owned?.setVisible(modeValue==='expression'&&available());if(modeValue==='conditions'){verification=null;job?.controller.abort();job=null;stageToken=null;hasCondition=false}onMode(modeValue);display();
  const invocation=modeInvocation={};if(modeValue==='expression')ensureOwner().catch(error=>{if(!available()||modeInvocation!==invocation||modeValue!=='expression'||!completionControlReady(mode))return;diagnostic=localDiagnostic(messageOf(error));display({announce:true})});
 };
 text.oninput=()=>{if(actionReady(text))pendingDraft()};
 async function verifyCondition(){
  if(!actionReady(verify))return;
  diagnostic=null;const invocation=verification={};let lease=null;
  try{
   const author=await ensureOwner();if(!author||!available()||verification!==invocation||!completionControlReady(verify))return;
   author.edit(text.value);const attempt=author.beginCheck();if(!attempt)return;
   lease=job={attempt,author,controller:new AbortController()};stageToken=null;hasCondition=false;
   onChange(undefined,'Checking the expression before applying.',{draft:true});display({announce:true});
   const invalid=!text.value.isWellFormed()?'Expression contains malformed Unicode.':!text.value||bytes(text.value)>4096?'Use an expression of 1–4096 UTF-8 bytes.':text.value.includes('\0')?'Expression cannot contain NUL.':null;
   let receipt;
   if(invalid){const failure=localDiagnostic(invalid);if(author.fail(attempt.token,failure)){diagnostic=failure;onChange(undefined,'Check expression before applying.',{diagnostic:failure})}}
   else{
    const response=await check(attempt.request,lease.controller.signal);
    if(job!==lease||!available())return;
    if(!completionControlReady(verify)){job=null;author.setVisible(false);author.setVisible(true);display();return;}
    if(response?.ok===false){const failure=response.data?.diagnostic||response.diagnostic||localDiagnostic(response.error||'Condition could not be checked.');if(author.fail(attempt.token,failure)){diagnostic=failure;onChange(undefined,'Check expression before applying.',{diagnostic:failure})}}
    else{
     receipt=response?.ok===true?response.data:response;
     if(author.accept(attempt.token,receipt)){
      const question=author.stage(attempt.token);if(question){stageToken=attempt.token;hasCondition=true;hasDraft=true;onChange(question,null,{expression:text.value,predicate:receipt.predicate,checked:true})}
     }else{const failure=localDiagnostic('Scope changed. Check the condition again.');diagnostic=failure;onChange(undefined,'Check expression before applying.',{diagnostic:failure})}
    }
   }
   if(job===lease){job=null;display({announce:true})}
  }catch(error){
   if(!available()||verification!==invocation||lease&&job!==lease||error?.name==='AbortError')return;
   if(!completionControlReady(verify)){if(lease&&job===lease)job=null;display();return;}
   const failure=error?.diagnostic||localDiagnostic(messageOf(error));
   if(lease){if(!lease.author.fail(lease.attempt.token,failure))return;job=null}
   diagnostic=failure;onChange(undefined,'Check expression before applying.',{diagnostic:failure});display({announce:true});
  }
 }
 verify.onclick=verifyCondition;
 clear.onclick=async()=>{
  if(!actionReady(clear))return;const invocation=verification={};
  try{const author=await ensureOwner();if(!author||!available()||verification!==invocation||!completionControlReady(clear))return;const token=author.clear(),question=author.stage(token);if(!question)return;text.value='';hasDraft=false;hasCondition=false;diagnostic=null;stageToken=token;onChange(question,null,{clear:true});display({announce:true})}
  catch(error){if(available()&&verification===invocation&&completionControlReady(clear)){diagnostic=localDiagnostic(messageOf(error));display({announce:true})}}
 };
 discard.onclick=()=>{
  if(!actionReady(discard))return;
  retire();text.value='';hasDraft=false;hasCondition=false;needsAdoption=false;diagnostic=null;onDiscard();display();
 };
 display();
 return {
  root,
  canEditConditions:()=>!hasDraft,
  snapshot:()=>({mode:modeValue,text:text.value,hasDraft,hasCondition,needsAdoption,diagnostic:structuredClone(diagnostic),owner:owned?.snapshot()||null}),
  sync(){display()},
  invalidate(){retire();diagnostic=null;hasCondition=false;needsAdoption=hasDraft;if(hasDraft)onChange(undefined,'Scope changed. Discard this draft before adding a new condition.',{draft:true});display()},
  setVisible(value){active=!!value;owned?.setVisible(active);if(!active){verification=null;modeInvocation=null;job?.controller.abort();job=null;stageToken=null;preparation?.controller.abort();preparation=null;preparing=null;hasCondition=false}display()},
  dispose(){if(disposed)return;disposed=true;active=false;retire();display()},
 };
}
