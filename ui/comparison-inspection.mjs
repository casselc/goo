/* Application owner for explicit stored-side reads; display components stay passive. */
import * as Measurements from './measurements.mjs';
import * as RecordPresentation from './record-presentation.mjs';
import * as UI from './components.mjs';
const global=globalThis;
function mount(host,{sides,current=()=>host.isConnected,actionsHost}={}){
  const n=UI.node,section=n('section',undefined,'comparison-inspection'),panel=n('section'),back=n('button','Back to comparison'),heading=n('h5'),status=n('p'),content=n('div'),selected=n('section');
  section.hidden=!!actionsHost;panel.hidden=true;panel.setAttribute('aria-label','Comparison evidence inspection');status.setAttribute('role','status');panel.append(back,heading,status,content,selected);const valid=()=>section.isConnected&&current();let generation=0,controller=null,opener=null;const pages=new Map(),inspector=UI.inspector(selected,{onClose:(_key,origin)=>origin?.scrollIntoView({block:'nearest'})});
  const buttons=sides.map(({label,side})=>{const button=n('button','Inspect '+label.toLowerCase()+' evidence');button.onclick=async()=>{
   if(!valid())return;const g=++generation;controller?.abort();controller=new AbortController();opener=button;section.hidden=false;panel.hidden=false;heading.textContent=label+' evidence';status.textContent='Reading stored results; no telemetry reread.';content.replaceChildren();inspector.clear();back.focus();
   try{const id=side.receipt.id;let page=pages.get(id);if(!page){const response=await fetch('/api/captures/'+id,{signal:controller.signal});page=await response.json();if(!response.ok)throw Error(page.error||'Stored evidence unavailable');if(page.receipt?.id!==id)throw Error('Stored receipt differs from requested evidence');if(g!==generation||!valid())return;pages.set(id,page)}if(g!==generation||!valid())return;
    if(page.explore){const data=page.explore,table=RecordPresentation.table({signal:data.question.signal,labels:['UTC','Service',data.question.signal==='spans'?'Span / duration':'Severity / body','Inspect'],caption:label+' stored inspection page'});status.textContent=(data.state&&data.state!=='observed'?'Result '+data.state+' · ':'')+data.count+' full matching records · '+data.rows.length+' on this stored page. '+(page.hasMore||data.hasMore||page.nextCursor?'More stored pages are available in the full capture.':'No source reread.');content.append(table.viewport);data.rows.forEach((record,index)=>{const {row,cells}=table.row(record),inspect=n('button','Inspect record');inspect.onclick=()=>{if(g!==generation||!valid())return;inspector.show({key:id+':'+index,title:label+' record',description:RecordPresentation.description(record),content:RecordPresentation.detail(record,data.question.signal,{},data.rows),opener:inspect});selected.querySelector('button[aria-label="Close selected item details"]').focus()};cells[3].append(inspect);table.body.append(row)})}
    else if(page.measurement){status.textContent='Stored native result · no telemetry reread.';Measurements.render(content,page.measurement,{saved:true})}
    else status.textContent='This evidence type has no inline result presenter. Open its retained receipt explicitly.';
   }catch(e){if(g===generation&&valid()&&e.name!=='AbortError')status.textContent='Stored inspection unavailable · '+e.message}
  };return button});
  const actions=UI.actions(buttons,[]);actions.setAttribute('role','group');actions.setAttribute('aria-label','Comparison inspection actions');if(actionsHost){actionsHost.append(...buttons);section.append(panel)}else section.append(actions,panel);host.append(section);
  back.onclick=()=>{generation++;controller?.abort();inspector.clear();panel.hidden=true;section.hidden=!!actionsHost;if(opener?.isConnected)opener.focus()};
  return {dispose(){generation++;controller?.abort();inspector.clear();if(actionsHost)for(const button of buttons)button.remove();section.remove()}};
 }
export {mount};
