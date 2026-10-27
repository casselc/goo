/* Passive supplied scalar observations/intervals. No source or measurement semantics. */
(function(global){
 function decimal(value){if(typeof value!=='string'||value.length>512)return null;const m=value.match(/^(-?)(\d+)(?:\.(\d+))?(?:e([+-]?\d+))?$/i);if(!m)return null;const e=Number(m[4]||0)-(m[3]||'').length;if(Math.abs(e)>600)return null;return {n:BigInt(m[2]+(m[3]||''))*(m[1]?-1n:1n)*(e>0?10n**BigInt(e):1n),d:e<0?10n**BigInt(-e):1n}}
 const compare=(a,b)=>a.n*b.d-b.n*a.d;
 const ns=v=>typeof v==='string'&&/^(0|[1-9]\d{0,18})$/.test(v)&&BigInt(v)<=9223372036854775807n;
 function model(data){
  const bad={state:'unavailable',points:[]};if(!data||!ns(data.fromNS)||!ns(data.toNS)||BigInt(data.toNS)<=BigInt(data.fromNS)||!Array.isArray(data.points)||data.points.length>256)return bad;
  const from=BigInt(data.fromNS),to=BigInt(data.toNS),keys=new Set(),points=[];let min=decimal('0'),max=decimal('0'),minLabel='0',maxLabel='0';const supplied=data.domain!==undefined;if(supplied){min=decimal(data.domain?.min);max=decimal(data.domain?.max);if(!min||!max||compare(min,decimal('0'))>0n||compare(max,decimal('0'))<0n||compare(min,max)>0n)return bad;minLabel=data.domain.min;maxLabel=data.domain.max}if(data.scaleLabel!==undefined&&(typeof data.scaleLabel!=='string'||data.scaleLabel.length>256))return bad;
  for(const p of data.points){
   if(!p||p.series!==undefined&&(typeof p.series!=='string'||!/^[a-z][a-z0-9-]{0,63}$/.test(p.series))||typeof p.key!=='string'||!p.key||p.key.length>512||keys.has(p.key)||typeof p.label!=='string'||p.label.length>8192||!ns(p.fromNS)||!ns(p.toNS))return bad;keys.add(p.key);
   const a=BigInt(p.fromNS),b=BigInt(p.toNS),state=p.state||'observed',known=['observed','stale'].includes(state),v=known?decimal(p.value):null;
   if(!['observed','stale','unknown','no_data'].includes(state)||a<from||b>to||b<a||known&&!v||!known&&p.value!==undefined&&p.value!==null)return bad;
   if(v&&supplied&&(compare(v,min)<0n||compare(v,max)>0n))return bad;
   if(!supplied&&v&&compare(v,min)<0n){min=v;minLabel=p.value}if(!supplied&&v&&compare(v,max)>0n){max=v;maxLabel=p.value}
   points.push({key:p.key,...(p.series!==undefined?{series:p.series}:{}),label:p.label,fromNS:p.fromNS,toNS:p.toNS,state,value:known?p.value:null,left:Number((a-from)*100000n/(to-from))/1000,width:Number((b-a)*100000n/(to-from))/1000,v});
  }
  const range=max.n*min.d-min.n*max.d;
  for(const p of points){p.y=p.v?(range===0n?0:Number((p.v.n*min.d-min.n*p.v.d)*max.d*100000n/(p.v.d*range))/1000):null;delete p.v}
  return {state:points.length?'observed':'no_data',numeric:points.some(p=>p.y!==null),min:minLabel,max:maxLabel,zeroY:range===0n?0:Number(-min.n*max.d*100000n/range)/1000,points};
 }
 // Pixel fractions choose supplied time bounds; all timestamp arithmetic remains exact.
 function brushRange(data,start,end){if(!data||!ns(data.fromNS)||!ns(data.toNS)||BigInt(data.toNS)<=BigInt(data.fromNS)||![start,end].every(x=>Number.isFinite(x)&&x>=0&&x<=1))return null;const a=BigInt(Math.round(Math.min(start,end)*1000000)),b=BigInt(Math.round(Math.max(start,end)*1000000)),from=BigInt(data.fromNS),width=BigInt(data.toNS)-from;const lo=from+width*a/1000000n,hi=from+width*b/1000000n;return hi>lo?{fromNS:String(lo),toNS:String(hi)}:null}
 function node(tag,text,cls){const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e}
 if(global.customElements&&global.HTMLElement){
  class GooValueTimeline extends global.HTMLElement {
   static get observedAttributes(){return ['data','selected-key','brush','touch-brush']}
   set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}
   get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=2097152?JSON.parse(raw):null}catch{return null}}
   set selectedKey(value){if(value===null||value===undefined)this.removeAttribute('selected-key');else this.setAttribute('selected-key',value)}get selectedKey(){return this.getAttribute('selected-key')}
   connectedCallback(){if(!this.hasAttribute('role'))this.setAttribute('role','group');if(!this.hasAttribute('aria-label'))this.setAttribute('aria-label','Supplied value timeline');for(const key of ['data','selectedKey'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
   attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected){if(name==='selected-key')this.syncSelection();else this.render()}}
   disconnectedCallback(){this._cancelBrush?.()}
   syncSelection(){for(const mark of this.querySelectorAll(':scope > .value-timeline-plot > .value-timeline-mark'))mark.setAttribute('aria-pressed',String(mark.dataset.key===this.selectedKey))}
   render(){
    this._cancelBrush?.();const m=model(this.data);this.replaceChildren();this.dataset.state=m.state;
    if(m.state!=='observed'){this.append(node('p',m.state==='no_data'?'No retained observations. Absence is not zero.':'Timeline unavailable.'));return}
    const scale=node('p',m.numeric?(this.data.scaleLabel||m.min+' to '+m.max+' · dashed baseline is zero'):'No numeric observations to scale','value-timeline-scale'),plot=node('div',undefined,'value-timeline-plot'),zero=node('div',undefined,'value-timeline-zero');zero.style.top=(100-m.zeroY)+'%';zero.hidden=!m.numeric;zero.setAttribute('aria-hidden','true');plot.append(zero);scale.title=m.min+' to '+m.max+' · exact displayed domain';
    for(const p of m.points){const mark=node('button',p.y===null?'?':'','value-timeline-mark');mark.type='button';mark.dataset.key=p.key;if(p.series!==undefined)mark.dataset.series=p.series;mark.dataset.state=p.state;mark.dataset.fromNs=p.fromNS;mark.dataset.toNs=p.toNS;if(p.value!==null)mark.dataset.value=p.value;mark.dataset.interval=String(p.fromNS!==p.toNS);mark.style.left=p.left+'%';mark.style.width=p.width+'%';mark.style.top=p.y===null?'0':(100-p.y)+'%';mark.setAttribute('aria-label',p.label);mark.title=p.label;mark.setAttribute('aria-pressed',String(this.selectedKey===p.key));mark.onclick=()=>this.dispatchEvent(new CustomEvent('goo-timeline-select',{bubbles:true,detail:{key:p.key}}));plot.append(mark)}
    this.append(scale,plot);
    if(this.hasAttribute('brush'))this.bindBrush(plot);

   }
   bindBrush(plot){
    plot.tabIndex=0;plot.setAttribute('aria-label','Drag to preview a time range; Escape cancels');let gesture=null,suppress=null;const overlay=node('div',undefined,'value-timeline-brush');overlay.hidden=true;overlay.setAttribute('aria-hidden','true');plot.append(overlay);
    const emit=(phase,bounds)=>this.dispatchEvent(new CustomEvent('goo-timeline-range',{bubbles:true,detail:{phase,...bounds}}));
    const fraction=(x,rect)=>Math.max(0,Math.min(1,(x-rect.left)/rect.width));
    const finish=(phase)=>{const g=gesture;if(!g)return;gesture=null;overlay.hidden=true;suppress=performance.now()+500;if(plot.hasPointerCapture(g.id))plot.releasePointerCapture(g.id);emit(phase,phase==='commit'?g.bounds:undefined);};
    this._cancelBrush=()=>finish('cancel');
    plot.addEventListener('click',e=>{if(suppress&&performance.now()<suppress&&e.isTrusted&&e.detail>0){e.preventDefault();e.stopImmediatePropagation()}},true);
    plot.addEventListener('pointerdown',e=>{if(e.button!==0||!e.isPrimary||e.pointerType==='touch'&&!this.hasAttribute('touch-brush'))return;const rect=plot.getBoundingClientRect();if(!rect.width)return;suppress=null;plot.focus({preventScroll:true});gesture={id:e.pointerId,start:e.clientX,rect,mark:e.target.closest('.value-timeline-mark'),bounds:null,moved:false};plot.setPointerCapture(e.pointerId);emit('start');});
    plot.addEventListener('pointermove',e=>{const g=gesture;if(!g||g.id!==e.pointerId)return;if(Math.abs(e.clientX-g.start)<4&&!g.moved)return;g.moved=true;g.bounds=brushRange(this.data,fraction(g.start,g.rect),fraction(e.clientX,g.rect));if(!g.bounds)return;const a=fraction(g.start,g.rect),b=fraction(e.clientX,g.rect);overlay.hidden=false;overlay.style.left=Math.min(a,b)*100+'%';overlay.style.width=Math.abs(a-b)*100+'%';emit('preview',g.bounds)});
    plot.addEventListener('pointerup',e=>{const g=gesture;if(!g||g.id!==e.pointerId)return;finish(g.moved&&g.bounds?'commit':'cancel');if(!g.moved&&g.mark){g.mark.focus({preventScroll:true});g.mark.click()}});
    plot.addEventListener('pointercancel',()=>finish('cancel'));plot.addEventListener('lostpointercapture',()=>finish('cancel'));plot.addEventListener('keydown',e=>{if(e.key==='Escape'&&gesture){e.preventDefault();finish('cancel')}});
   }
  }
  if(!global.customElements.get('goo-value-timeline'))global.customElements.define('goo-value-timeline',GooValueTimeline);
 }
 const api={decimal,model,brushRange};if(typeof module!=='undefined')module.exports=api;else global.GooValueTimeline=api;
})(typeof window==='undefined'?globalThis:window);
