// Exercise the registered production element's filtering, not a second filter.
// Native journeys separately check real custom-element reactions and painting.
class Element {
 constructor(){this.attributes=new Map();this.rows=[];this.isConnected=false;}
 getAttribute(key){return this.attributes.get(key)??null}
 hasAttribute(key){return this.attributes.has(key)}
 setAttribute(key,value){const old=this.getAttribute(key);this.attributes.set(key,String(value));if(this.isConnected&&this.constructor.observedAttributes?.includes(key))this.attributeChangedCallback(key,old,String(value));}
 querySelectorAll(){return this.rows}
}
const registry=new Map();
globalThis.HTMLElement=Element;
globalThis.customElements={get:key=>registry.get(key),define:(key,value)=>registry.set(key,value)};
await import('../../ui/components.mjs');
const FacetList=registry.get('goo-facet-list');
let input='';for await(const chunk of process.stdin)input+=chunk;
const {size,rounds,seed}=JSON.parse(input);
if(!Number.isInteger(size)||size<1||size>250||!Number.isInteger(rounds)||rounds<1||rounds>12)throw Error('Bound the generated list schedule.');
const list=new FacetList(),history=[];
let serial=0;
const row=()=>({text:'Category Ω🙂 '+seed+' terminal-'+String(serial++).padStart(4,'0'),hidden:false,getAttribute(key){return key==='data-facet-text'?this.text:null}});
list.rows=Array.from({length:size},row);
let scans=0;const refresh=list.refresh;list.refresh=function(){scans++;return refresh.call(this)};
for(let round=0;round<rounds;round++){
 list.isConnected=round%2===0;
 const query=['terminal-00','CATEGORY','absent-'+seed,''][round%4];
 for(const mutation of ['unchanged','insert','text','remove']){
  if(mutation==='insert')list.rows.push(row());
  if(mutation==='text'&&list.rows.length)list.rows[0].text='changed Ω🙂 '+seed;
  if(mutation==='remove'&&list.rows.length>1)list.rows.splice(1,1);
  const identities=[...list.rows],before=scans,result=list.filter(query);
  history.push({query,mutation,connected:list.isConnected,rows:list.rows.map(item=>item.text),hidden:list.rows.map(item=>item.hidden),result,scans:scans-before,identity:list.rows.length===identities.length&&list.rows.every((item,index)=>item===identities[index])});
 }
}
process.stdout.write(JSON.stringify({history})+'\n');
