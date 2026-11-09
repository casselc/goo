/* Supplied fixed record facts only; execution, navigation and custody stay in owners. */
import * as UI from './components.mjs';
import {formatFact} from './evidence-summary.mjs';
import * as Environment from './record-environment.mjs';
import {selectorText} from './question-session.mjs';
function render(host,result,{receipt,retainedRecords,recordsLimited}={}){
 const q=result.question,node=UI.node,window=formatFact({value:'['+BigInt(q.fromMs)*1000000n+', '+BigInt(q.toMs)*1000000n+')',format:'unix-ns-range'});
 host.replaceChildren();host.classList.add('fixed-record-summary');host.dataset.resultState=result.state;
 const heading=node('strong',(receipt?'Stored evidence':'Applied question')+' · '+result.count+' matching captured '+q.signal+' · '+(result.rows?.length??0)+' displayed');heading.className='fixed-record-count';
 const scope=node('p',Environment.questionLabel(q)+' · '+window.text);scope.title=window.title;scope.className='fixed-record-window';
 const details=node('details'),summary=node('summary','Scope and evidence'),facts=node('dl');details.className='fixed-record-facts';
 const values=[['Signal',q.signal],['Environment',Environment.questionLabel(q)],['Fixed window · end excluded',window.text],['Selector',q.where?selectorText(q.where):'All matching records'],['Page offset',String(q.offset||0)],['Source read',result.capturedAt||'Unavailable in supplied evidence'],['Evidence basis',result.basis],['Coverage','Unverified; counts and rows are separate source reads.']];
 if(q.reportedEdge)values.push(['Reported relationship',JSON.stringify(q.reportedEdge)]);
 if(receipt)values.push(['Capture ID',receipt.id],['Stored custody time',receipt.capturedAt],['Retained inspection records',String(retainedRecords)],['Retention',recordsLimited?'More matching source records were not retained':'All records returned by the capture read were retained; source coverage remains unknown.']);
 for(const [label,value]of values){const group=node('div');group.append(node('dt',label),node('dd',value??'Unavailable'));facts.append(group)}
 details.append(summary,facts);if(result.warnings?.length)details.append(node('p',result.warnings.join(' · ')));
 host.append(heading,scope);
 if(result.state==='changed_between_reads'){const warning=node('p','Source reads disagree. Rerun before treating counts and rows as one population.');warning.className='fixed-record-warning';host.append(warning)}
 else if(result.state==='no_data')host.append(node('p','No matching captured records. Source coverage and application health remain unknown.'));
 const limits=[];if(result.hasMore)limits.push(q.offset>=1000?'Paging budget reached; narrow the question':'More matching records are available');if(result.fieldsLimited)limits.push('Field discovery is limited; unlisted does not mean absent');if(recordsLimited)limits.push('Stored inspection records are limited');if(limits.length){const notice=node('p',limits.join(' · '));notice.className='fixed-record-limits';host.append(notice)}
 host.append(details);return host;
}
export {render};
