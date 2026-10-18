import {always,eventually} from '@antithesishq/bombadil';
import {extract,actions,registerCustomAction} from '@antithesishq/bombadil/browser';
export {noUncaughtExceptions,noUnhandledPromiseRejections} from '@antithesishq/bombadil/browser/defaults/properties';
const state=extract(({document})=>({stage:document.documentElement.dataset.gooMetricStage||'initial',ready:!!document.getElementById('metricDiscover'),catalog:document.getElementById('metricInstruments')?.textContent||'',status:document.getElementById('metricDiscoveryStatus')?.textContent||'',evidence:document.getElementById('measureResult')?.textContent||'',captureDisabled:(document.getElementById('measureCapture') as HTMLButtonElement)?.disabled}));
export const discoveryDoesNotEvaluate=always(()=>state.current.evidence==='');
export const discoveryDoesNotCapture=always(()=>state.current.captureDisabled!==false);
export const done=eventually(()=>state.current.stage==='complete').within(25,'seconds');
const act=registerCustomAction('goo.metric-discovery',async(document,window,phase:string)=>{window.dispatchEvent(new CustomEvent('goo:measurements-tab'));if(phase==='abandon'){(document.getElementById('metricService') as HTMLInputElement).value='independent-metrics';document.getElementById('metricDiscover')?.click();document.getElementById('exploreRecordsMode')?.click();document.getElementById('exploreMeasurementsMode')?.click()}else if(phase==='discover'){(document.getElementById('metricService') as HTMLInputElement).value='independent-metrics';document.getElementById('metricDiscover')?.click()}else{const input=document.getElementById('metricService') as HTMLInputElement;input.value='different-service';input.dispatchEvent(new Event('input',{bubbles:true}))}document.documentElement.dataset.gooMetricStage=phase==='abandon'?'abandoned':phase==='discover'?'waiting':'complete'});
export const explore=actions(()=>{const s=state.current;if(!s.ready)return['Wait'];if(s.stage==='initial')return[act('abandon')];if(s.stage==='abandoned')return[act('discover')];if(s.stage==='waiting'&&s.catalog.includes('jobs.completed'))return[act('clear')];return['Wait']});
export const changedScopeClearsCatalog=always(()=>state.current.stage!=='complete'||state.current.catalog==='');

export const abandonedReadCannotPublish=always(()=>state.current.stage!=='abandoned'||state.current.catalog==='');
