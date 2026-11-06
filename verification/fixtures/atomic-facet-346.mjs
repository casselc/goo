// Minimal mounted DOM adapted from categorical-pair-render.test.mjs. Production
// CategoricalFacet and registered GooFacetList own all rendering and filtering.
class Element {
 constructor(tag='section'){this.tagName=tag;this.children=[];this.attributes={};this.dataset={};this.style={setProperty(k,v){this[k]=v}};this.isConnected=false;this.listeners=new Map();this.className='';this._text='';this.classList={add:n=>{if(!this.className.split(' ').includes(n))this.className+=' '+n},toggle:(n,on)=>{this.className=this.className.split(' ').filter(x=>x&&x!==n).join(' ');if(on)this.classList.add(n)}};}
 set textContent(v){this._text=v??''}get textContent(){return this._text+this.children.map(c=>c.textContent).join('')}
 append(...children){for(const c of children){c.parentElement=this;c.isConnected=this.isConnected;this.children.push(c);if(c.isConnected)c.connectedCallback?.()}}
 prepend(...children){for(const c of children.reverse()){c.parentElement=this;c.isConnected=this.isConnected;this.children.unshift(c)}}
 replaceChildren(...children){for(const c of this.children){c.parentElement=null;c.isConnected=false}this.children=[];this._text='';this.append(...children)}
 contains(e){return e===this||this.children.some(c=>c.contains(e))}
 setAttribute(k,v){const old=this.getAttribute(k);this.attributes[k]=String(v);if(this.isConnected&&this.constructor.observedAttributes?.includes(k))this.attributeChangedCallback?.(k,old,String(v))}
 getAttribute(k){if(k.startsWith('data-')){const name=k.slice(5).replace(/-([a-z])/g,(_,c)=>c.toUpperCase());if(Object.hasOwn(this.dataset,name))return String(this.dataset[name])}return this.attributes[k]??null}hasAttribute(k){return this.getAttribute(k)!==null}
 matches(s){if(s.startsWith('.'))return this.className.split(' ').includes(s.slice(1));if(s==='[aria-label]')return this.hasAttribute('aria-label');return this.tagName===s}
 querySelectorAll(s){if(s===':scope > [data-facet-item]')return this.children.filter(c=>Object.hasOwn(c.dataset,'facetItem'));return this.children.flatMap(c=>[...(c.matches(s)?[c]:[]),...c.querySelectorAll(s)])}querySelector(s){return this.querySelectorAll(s)[0]||null}
 addEventListener(t,h){this.listeners.set(t,h)}dispatchEvent(e){this.listeners.get(e.type)?.(e);return true}focus(){globalThis.document.activeElement=this}
}
const registry=new Map();globalThis.HTMLElement=Element;globalThis.customElements={get:k=>registry.get(k),define:(k,v)=>registry.set(k,v)};
const body=new Element('body');body.isConnected=true;globalThis.document={body,activeElement:body,createElement:tag=>{const C=registry.get(tag),e=C?new C():new Element(tag);e.tagName=tag;return e}};
if(!globalThis.CustomEvent)globalThis.CustomEvent=class{constructor(type,{detail}){this.type=type;this.detail=detail}};
await import('../../ui/components.mjs');await import('../../ui/categorical-facet.mjs');
const Facet=registry.get('goo-categorical-facet');
let raw='';for await(const chunk of process.stdin)raw+=chunk;const {rounds,seed}=JSON.parse(raw);
if(!Number.isInteger(rounds)||rounds<2||rounds>6||!Number.isInteger(seed)||seed<0||seed>1000000)throw Error('Bound the publication schedule.');
const field={kind:'attribute',key:'cohort'},declaration={version:'goo.view/1',scope:{kind:'record-query',signal:'logs'},kind:{name:'categorical-facet',version:1},field,display:{basis:'applied'}};
const element=new Facet(),events=[],renders=[],history=[],retired=[];let ownerId=-1;
for(const type of ['goo:facet-stage','goo:facet-load','goo:facet-clear'])element.addEventListener(type,e=>events.push({type,detail:e.detail}));
const originalRender=element.render;element.render=function(){renders.push({owner:this.interaction?.id,adopted:this.data});return originalRender.call(this)};
const button=label=>element.querySelectorAll('[aria-label]').find(e=>e.getAttribute('aria-label')===label),search=()=>button('Search loaded facet values');
const state=()=>({data:element.data,find:search()?.value??null,focused:!!search()&&document.activeElement===search(),details:!!element.querySelector('.categorical-facet-receipts')?.open,rows:element.querySelectorAll('.categorical-facet-row').map(e=>({key:e.dataset.facetValueKey,hidden:!!e.hidden,count:e.querySelector('.categorical-facet-value-count')?.textContent}))});
let basis=0;
for(let turn=0;turn<rounds;turn++){
 const same=turn>0&&(seed+turn)%2===0;if(!same)basis++;
 const value='literal Ω🙂 '+seed+' terminal-'+turn,count=(9007199254740993n+BigInt(turn)).toString(),readAt='2026-11-06T00:00:0'+basis+'Z',receipt={readAt,basis:'a'.repeat(64)},data={declaration,values:{state:'ready',field,items:[{value,count},{value:null,count:'1'},{value:'',count:'2'}],total:(BigInt(count)+3n).toString(),readAt,basis:'source-'+basis,reads:{count:receipt,groups:receipt}},disabled:(seed+turn)%5===0,hasStaged:true};
 const expected=structuredClone(data),connected=(seed+turn)%3!==0,allowed=(seed+turn)%4!==0;ownerId=turn;const owner={id:turn,current:()=>ownerId===turn&&allowed};
 const before=state(),eventStart=events.length,renderStart=renders.length;const old=button('Include attribute:cohort '+JSON.stringify(value));if(old)retired.push(old.onclick);
 // Retain the actual previous render's callback before replacing its data.
 const previous=element.querySelectorAll('button').find(e=>e.getAttribute('aria-label')?.startsWith('Include attribute:cohort '));if(previous)retired.push(previous.onclick);
 element.isConnected=connected;element.publish({data,interaction:owner});
 const adopted=element.data;data.values.items[0].value='caller mutation';data.values.items[0].count='1';const returned=element.data;if(returned)returned.values.items[0].value='getter mutation';
 const detachedCallbacks=events.length;for(const fn of retired)fn();const retiredEvents=events.length-detachedCallbacks;
 const detachedRenders=renders.length-renderStart;
 if(!connected){element.isConnected=true;element.connectedCallback()}
 const after=state();const action=button('Include attribute:cohort '+JSON.stringify(value));action?.onclick();button('Read values')?.onclick();button('Clear staged facet')?.onclick();
 const currentEvents=events.slice(eventStart),rendered=renders.slice(renderStart);
 const input=search();if(input){input.value='terminal';input.focus();input.oninput();element.querySelector('.categorical-facet-receipts').open=true}
 history.push({turn,same,connected,allowed,expected,before,adopted,after,currentEvents,retiredEvents,detachedRenders,rendered});
}
const last=button('Read values');if(last)retired.push(last.onclick);const beforeInvalid=events.length;element.publish({data:{invalid:true},interaction:{id:rounds,current:()=>true}});for(const fn of retired)fn();
process.stdout.write(JSON.stringify({history,invalid:{data:element.data,interaction:element.interaction,state:element.dataset.state,newEvents:events.length-beforeInvalid}})+'\n');
