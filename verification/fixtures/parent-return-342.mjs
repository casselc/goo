import {retainedDefinitionControls,definitionControlState,restoreDefinitionControlState} from '../../ui/retained-definition-controls.mjs';
export function campaign({rounds=12,duplicates=3,seed=1}={}){
 if(!Number.isInteger(rounds)||rounds<1||rounds>1000||!Number.isInteger(duplicates)||duplicates<1||duplicates>20||!Number.isSafeInteger(seed))throw Error('Bound the return campaign.');
 const question=Object.freeze({policy:'effective',fromMs:1000,toMs:2000}),receipt=Object.freeze({sourceBasis:'held-source',capturedAt:'2026-11-06T00:00:00.123456789Z',count:'9007199254740993'}),original=JSON.stringify({question,receipt});
 let generation=0,currentQuestion=question,connected=true,active=true,child=false,phase='legitimate',draft='unsubmitted  µs',reads=0,legitimateEffects=0,retiredEffects=0,failedRenewals=0,random=seed>>>0;
 const lease=retainedDefinitionControls(()=>({generation,eligible:currentQuestion===question&&connected&&active&&!child})),old=[];let draftPreserved=true;const input=(value,checked,start,end)=>({tagName:'INPUT',type:'text',value,checked,selectionStart:start,selectionEnd:end,selectionDirection:'backward',getAttribute:()=>null,setSelectionRange(a,b,direction){this.selectionStart=a;this.selectionEnd=b;this.selectionDirection=direction;}}),disclosure=open=>({tagName:'DETAILS',open,getAttribute:()=>null});
 const action=()=>{if(phase==='legitimate')legitimateEffects++;else retiredEffects++;};let handler=lease.bind(action);handler();
 const invokeOld=()=>{for(const callback of old)for(let i=0;i<duplicates;i++)callback();};
 for(let round=0;round<rounds;round++){
  old.push(handler);phase='retired';lease.retire();child=true;generation++;invokeOld();if(!lease.renew())failedRenewals++;
  child=false;connected=false;if(!lease.renew())failedRenewals++;invokeOld();connected=true;active=false;if(!lease.renew())failedRenewals++;invokeOld();active=true;
  // Equal text is a different question owner. It cannot resume this retained definition.
  currentQuestion={...question};if(!lease.renew())failedRenewals++;invokeOld();currentQuestion=question;
  if(!lease.renew())throw Error('Current retained definition did not resume.');handler=lease.bind(action);invokeOld();
  random=(Math.imul(random,1664525)+1013904223)>>>0;draft='raw '+random+' · '+round+' λ  ';
  const before=[input(draft,Boolean(random&1),round%draft.length,draft.length),disclosure(Boolean(random&2))],snapshot=definitionControlState({querySelectorAll:()=>before}),after=[input('default',false,0,0),disclosure(!before[1].open)];
  restoreDefinitionControlState({querySelectorAll:()=>after},snapshot);
  draftPreserved&&=['value','checked','selectionStart','selectionEnd','selectionDirection'].every(key=>before[0][key]===after[0][key])&&before[1].open===after[1].open;
  phase='legitimate';for(let duplicate=0;duplicate<duplicates;duplicate++)handler();
 }
 phase='retired';lease.retire();old.push(handler);invokeOld();
 return {rounds,duplicates,seed,legitimateEffects,retiredEffects,failedRenewals,reads,draftPreserved,receiptPreserved:JSON.stringify({question,receipt})===original,oldCallbacks:old.length};
}
if(process.argv[1]===new URL(import.meta.url).pathname){let source='';for await(const chunk of process.stdin)source+=chunk;process.stdout.write(JSON.stringify(campaign(source.trim()?JSON.parse(source):{}))+'\n');}
