import './components.mjs';
/* Supplied bucket populations only. Layout compatibility and reads belong to callers. */
const global = globalThis;
 const uint=v=>typeof v==='string'&&/^(0|[1-9]\d*)$/.test(v)&&v.length<=128;
 function model(data){
  const bad={state:'unavailable',rows:[],columns:[]};if(!data||!uint(data.ceiling)||!Array.isArray(data.rows)||!Array.isArray(data.columns)||!data.rows.length||data.rows.length>64||data.columns.length>32)return bad;
  const rowKeys=new Set(),colKeys=new Set(),max=BigInt(data.ceiling),validKey=v=>typeof v==='string'&&!!v&&v.length<=512,validLabel=v=>typeof v==='string'&&v.length<=4096;
  for(const r of data.rows){if(!r||!validKey(r.key)||rowKeys.has(r.key)||!validLabel(r.label))return bad;rowKeys.add(r.key)}
  const columns=[];for(const c of data.columns){if(!c||!validKey(c.key)||colKeys.has(c.key)||!validLabel(c.label)||!Array.isArray(c.cells)||c.cells.length!==data.rows.length)return bad;colKeys.add(c.key);const cells=[];
   for(const p of c.cells){if(!p||!['observed','unknown','outside'].includes(p.state))return bad;const known=p.state==='observed';if(known?(!uint(p.value)||BigInt(p.value)>max):p.value!==undefined&&p.value!==null)return bad;cells.push({state:p.state,value:known?p.value:null,intensity:known&&max>0n?Number(BigInt(p.value)*10000n/max)/10000:0})}columns.push({key:c.key,label:c.label,cells})
  }
  return {state:columns.length?'observed':'no_data',ceiling:data.ceiling,title:typeof data.title==='string'?data.title:'Supplied bucket populations',rows:structuredClone(data.rows),columns};
 }
 const node=(tag,text)=>{const e=document.createElement(tag);if(text!==undefined)e.textContent=text;return e};
 if(global.customElements&&global.HTMLElement){
  class GooBucketHeatmap extends global.HTMLElement{
   static get observedAttributes(){return ['data','selected-column','selected-row','range-select']}
   set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=1048576?JSON.parse(raw):null}catch{return null}}
   set selectedCell(value){this._batch=true;for(const [attr,key]of[['selected-column','columnKey'],['selected-row','rowKey']]){if(value?.[key]!==undefined)this.setAttribute(attr,value[key]);else this.removeAttribute(attr)}this._batch=false;if(this.isConnected)this.syncSelection()}get selectedCell(){return this.hasAttribute('selected-column')&&this.hasAttribute('selected-row')?{columnKey:this.getAttribute('selected-column'),rowKey:this.getAttribute('selected-row')}:null}
   connectedCallback(){for(const key of ['data','selectedCell'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
   attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected&&!this._batch){if(name==='data'||name==='range-select')this.render();else this.syncSelection()}}
   syncSelection(){const selected=this.selectedCell;for(const cell of this.querySelectorAll('.bucket-heatmap-cell'))cell.setAttribute('aria-pressed',String(!cell.disabled&&selected?.columnKey===cell.dataset.column&&selected?.rowKey===cell.dataset.row))}
   render(){this.onpointerdown=this.onpointerover=this.onpointerup=this.onpointercancel=this.onkeydown=null;const m=model(this.data);this.replaceChildren();this.dataset.state=m.state;if(m.state!=='observed'){this.append(node('p',m.state==='no_data'?'No retained intervals.':'Bucket heatmap unavailable; inspect the exact retained values.'));return}
    const legend=node('p','Shared count scale: 0–'+m.ceiling+' · 0 = recorded zero · ? = unknown · — = other layout'),scroll=node('goo-table'),table=node('table'),thead=node('thead'),head=node('tr'),body=node('tbody');legend.className='bucket-heatmap-legend';scroll.setAttribute('density','compact');scroll.setAttribute('role','region');scroll.tabIndex=0;scroll.setAttribute('aria-label',m.title);table.append(node('caption',m.title));const corner=node('th','Bucket');corner.scope='col';head.append(corner);for(const c of m.columns){const th=node('th',c.label);th.scope='col';head.append(th)}thead.append(head);
    for(const [i,row]of m.rows.entries()){const tr=node('tr'),th=node('th',row.label);th.scope='row';tr.append(th);for(const c of m.columns){const p=c.cells[i],td=node('td'),button=node('button',p.value===null?p.state==='outside'?'—':'?':p.value);button.type='button';button.className='bucket-heatmap-cell';button.dataset.column=c.key;button.dataset.row=row.key;button.dataset.state=p.state;button.disabled=p.state==='outside';if(p.value!==null){button.dataset.value=p.value;if(BigInt(p.value)>0n)button.style.background='rgba(57,201,173,'+(.12+.75*p.intensity)+')'}button.setAttribute('aria-label',c.label+' · '+row.label+' · '+(p.value===null?p.state==='outside'?'other bucket layout':'unknown bucket population':p.value+' observations'));button.onclick=()=>this.dispatchEvent(new CustomEvent('goo-heatmap-select',{bubbles:true,detail:{columnKey:c.key,rowKey:row.key}}));td.append(button);tr.append(td)}body.append(tr)}table.append(thead,body);scroll.append(table);this.append(legend,scroll);this.syncSelection();if(this.hasAttribute('range-select'))this.bindRange(m)
   }
   bindRange(m){
    let anchor=null,end=null,pointer=null;
    const cells=()=>[...this.querySelectorAll('.bucket-heatmap-cell')],at=e=>{const b=e.target.closest?.('.bucket-heatmap-cell');return b&&!b.disabled?{column:m.columns.findIndex(c=>c.key===b.dataset.column),row:m.rows.findIndex(r=>r.key===b.dataset.row)}:null};
    const paint=()=>{for(const b of cells()){const c=m.columns.findIndex(c=>c.key===b.dataset.column),r=m.rows.findIndex(r=>r.key===b.dataset.row);b.dataset.rangeSelected=String(!!anchor&&!!end&&c>=Math.min(anchor.column,end.column)&&c<=Math.max(anchor.column,end.column)&&r>=Math.min(anchor.row,end.row)&&r<=Math.max(anchor.row,end.row))}};
    const commit=()=>{if(!anchor||!end)return;this.dispatchEvent(new CustomEvent('goo-heatmap-range',{bubbles:true,detail:{fromColumn:Math.min(anchor.column,end.column),toColumn:Math.max(anchor.column,end.column),fromRow:Math.min(anchor.row,end.row),toRow:Math.max(anchor.row,end.row)}}))};
    this.onpointerdown=e=>{if(e.button!==0||e.pointerType==='touch')return;const p=at(e);if(!p)return;anchor=end=p;pointer=e.pointerId;paint();e.preventDefault()};
    this.onpointerover=e=>{if(pointer===null)return;if(!e.buttons){pointer=null;return;}const p=at(e);if(p){end=p;paint()}};
    this.onpointerup=e=>{if(pointer!==e.pointerId)return;pointer=null;commit();anchor=null};
    this.onpointercancel=()=>{pointer=null;anchor=end=null;paint()};
    this.onkeydown=e=>{const p=at(e);if(!p)return;if(e.key==='Escape'){pointer=null;anchor=end=null;paint();e.preventDefault();return}if(e.key===' '){e.preventDefault();e.stopPropagation();if(!anchor){anchor=end=p;paint()}else{end=p;paint();commit();anchor=null}return}if(e.shiftKey&&['ArrowLeft','ArrowRight','ArrowUp','ArrowDown'].includes(e.key)){e.preventDefault();anchor=anchor||p;end={column:Math.max(0,Math.min(m.columns.length-1,p.column+(e.key==='ArrowRight'?1:e.key==='ArrowLeft'?-1:0))),row:Math.max(0,Math.min(m.rows.length-1,p.row+(e.key==='ArrowDown'?1:e.key==='ArrowUp'?-1:0)))};cells().find(b=>b.dataset.column===m.columns[end.column].key&&b.dataset.row===m.rows[end.row].key)?.focus();paint()}if(e.key==='Enter'&&anchor){e.preventDefault();e.stopPropagation();end=p;paint();commit();anchor=null}};
    {const hint=node('p','Drag across cells to stage a rectangle. Keyboard: Space to anchor, Shift + arrows to extend, Enter to stage; Escape cancels. Exact bounds can be edited below.');hint.className='bucket-heatmap-legend';this.prepend(hint)}
   }

  }
  if(!global.customElements.get('goo-bucket-heatmap'))global.customElements.define('goo-bucket-heatmap',GooBucketHeatmap);
 }
 export {model};
