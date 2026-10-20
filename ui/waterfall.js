/* Passive span geometry and native-button display. App adapters own selection intent,
 * detail panels, source reads and navigation. Subtract exact ns before Number. */
(function(global){
  function layout(rows) {
    const spans = rows.map((row,index) => ({ row,index, start: BigInt(row.timestamp_ns), duration: BigInt(row.duration_ns) }));
    if (!spans.length) return { spans: [], start: 0n, end: 0n };
    let start = spans[0].start, end = start;
    for (const s of spans) { if (s.start < start) start = s.start; if (s.start + s.duration > end) end = s.start + s.duration; }
    const range = Number(end - start) || 1;
    const byKey = new Map(spans.map(s => [s.row.TraceId + ':' + s.row.SpanId, s]));
    function depth(s) {
      const seen = new Set([s]); let current = s, n = 0;
      while (current.row.ParentSpanId && !/^0+$/.test(current.row.ParentSpanId)) {
        const parent = byKey.get(current.row.TraceId + ':' + current.row.ParentSpanId);
        if (!parent) { s.missingParent = true; break; }
        if (seen.has(parent)) { s.cycle = true; break; }
        seen.add(parent); n++; current = parent;
      }
      return n;
    }
    for (const s of spans) {
      s.depth = depth(s); s.left = 100 * Number(s.start - start) / range;
      s.width = 100 * Number(s.duration) / range;
    }
    // Parent-first traversal, stable by exact start time; links never add edges.
    const compare = (a,b) => a.start < b.start ? -1 : a.start > b.start ? 1 : a.row.SpanId.localeCompare(b.row.SpanId);
    const ordered = [], visited = new Set();
    const children = new Map();
    for (const s of spans) {
      const key = s.row.TraceId + ':' + s.row.ParentSpanId;
      if (!children.has(key)) children.set(key, []);
      children.get(key).push(s);
    }
    function visit(s) { if (visited.has(s)) return; visited.add(s); ordered.push(s);
      for (const child of (children.get(s.row.TraceId + ':' + s.row.SpanId) || []).sort(compare)) visit(child); }
    const traces = [...new Set(spans.map(s => s.row.TraceId))].sort();
    for (const trace of traces) {
      const group = spans.filter(s => s.row.TraceId === trace).sort(compare);
      group.filter(s => !byKey.has(trace + ':' + s.row.ParentSpanId)).forEach(visit);
      group.forEach(visit); // Cycles remain inspectable and explicitly labelled.
    }
    return { spans: ordered, start, end };
  }
 function validData(data){
  const ns=v=>typeof v==='string'&&/^(0|[1-9][0-9]{0,19})$/.test(v)&&BigInt(v)<=18446744073709551615n;
  return !!data&&Array.isArray(data.rows)&&data.rows.length<=1000&&(!Object.hasOwn(data,'truncated')||typeof data.truncated==='boolean')&&data.rows.every(r=>r&&ns(r.timestamp_ns)&&ns(r.duration_ns)&&['TraceId','SpanId'].every(k=>typeof r[k]==='string')&&['ParentSpanId','ServiceName','SpanName','StatusCode'].every(k=>r[k]===undefined||typeof r[k]==='string'));
 }
 function node(tag,text,cls){const n=document.createElement(tag);if(text!==undefined)n.textContent=text;if(cls)n.className=cls;return n}
 if(global.customElements&&global.HTMLElement){
  class GooTraceWaterfall extends global.HTMLElement{
   static get observedAttributes(){return ['data','selected-index']}
   set data(value){try{this._data=structuredClone(value)}catch{this._data=null}if(this.isConnected)this.render()}
   get data(){if(this._data!==undefined)return structuredClone(this._data);try{const raw=this.getAttribute('data');return raw&&raw.length<=65536?JSON.parse(raw):null}catch{return null}}
   set selectedIndex(value){this.setAttribute('selected-index',value)}get selectedIndex(){const raw=this.getAttribute('selected-index');return raw!==null&&/^(0|[1-9][0-9]*)$/.test(raw)&&Number(raw)<1000?Number(raw):-1}
   connectedCallback(){for(const key of ['data','selectedIndex'])if(Object.hasOwn(this,key)){const value=this[key];delete this[key];this[key]=value}this.render()}
   attributeChangedCallback(name){if(name==='data')this._data=undefined;if(this.isConnected){if(name==='selected-index')this.updateSelection();else this.render()}}
   updateSelection(){for(const row of this.querySelectorAll('.waterfall-row'))row.setAttribute('aria-pressed',String(Number(row.dataset.rowIndex)===this.selectedIndex))}
   render(){const input=this.data;this.replaceChildren();if(!validData(input)){this.dataset.state='unavailable';this.append(node('p','Waterfall unavailable: unsupported span data.','waterfall-status'));return}
    const data=layout(input.rows);this.dataset.state=input.rows.length?'observed':'empty';
    this.append(node('p',`Displayed span extent: ${(Number(data.end-data.start)/1e6).toFixed(2)} ms. Relative to earliest returned span. Evidence completeness unknown${input.truncated?'; span list truncated':''}. Async links are references, not parent edges.`,'waterfall-status'));
    if(!input.rows.length){this.append(node('p','No spans in this selection.'));return}
    this.append(node('p','Parent layout follows reported IDs in supplied rows; duplicate IDs can be ambiguous. These rows are not unique invocation counts.','waterfall-status waterfall-identity-note'));
    let trace;
    for(const span of data.spans){const r=span.row;if(trace!==r.TraceId){trace=r.TraceId;this.append(node('h3','Trace '+(trace||'(unspecified)')))}
     const line=node('button',undefined,'waterfall-row');line.type='button';line.dataset.rowIndex=String(span.index);line.dataset.spanId=r.SpanId||'';line.dataset.traceId=r.TraceId||'';
     const suffix=span.cycle?' · parent cycle':span.missingParent?' · parent outside results':'';
     const title=`${r.ServiceName||'(service unspecified)'} · ${r.SpanName||'(span unnamed)'}${suffix}`;line.setAttribute('aria-label',title);
     const label=node('span',title,'span-label');label.style.paddingLeft=Math.min(span.depth,12)*12+'px';
     const track=node('span',undefined,'span-track'),bar=node('span',undefined,'span-bar');track.setAttribute('aria-hidden','true');bar.style.left=span.left+'%';bar.style.width=span.width+'%';bar.dataset.error=r.StatusCode==='Error'?'true':'false';track.append(bar);
     line.append(label,track,node('span',`${Number(span.duration)/1e6} ms${r.StatusCode==='Error'?' · error':''}`,'span-duration'));this.append(line);
    }this.updateSelection();
   }
  }
  if(!global.customElements.get('goo-trace-waterfall'))global.customElements.define('goo-trace-waterfall',GooTraceWaterfall);
 }
 const api={layout,validData};if(typeof module!=='undefined')module.exports=api;else global.GooWaterfall=api;
})(typeof window==='undefined'?globalThis:window);
