/* Renderer-neutral span geometry. Adapted from observability/lakeui's relative
 * time waterfall approach; subtract exact ns BEFORE converting to Number. */
(function (global) {
  const edges = [0, 10, 50, 100, 500, 1000, 10000, 18446744073710];
  const labels = ['0–10 ms', '10–50 ms', '50–100 ms', '100–500 ms', '0.5–1 s', '1–10 s', '≥10 s'];
  function layout(rows) {
    const spans = rows.map(row => ({ row, start: BigInt(row.timestamp_ns), duration: BigInt(row.duration_ns) }));
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
  function node(tag, text, cls) { const n = document.createElement(tag); if (text !== undefined) n.textContent = text; if (cls) n.className = cls; return n; }
  function waterfall(container, rows, truncated, relatedActions, detailsId = 'spanDetails') {
    const data = layout(rows);
    container.append(node('p', `Displayed span extent: ${(Number(data.end-data.start)/1e6).toFixed(2)} ms. Relative to earliest returned span. Evidence completeness unknown${truncated ? '; span list truncated' : ''}. Async links are references, not parent edges.`, 'muted'));
    if (!rows.length) { container.append(node('p', 'No spans in this selection.')); return; }
    const details = node('section'); details.id = detailsId; details.append(node('p', 'Select a span to inspect attributes and exact IDs.'));
    let trace;
    for (const s of data.spans) {
      const r = s.row;
      if (trace !== r.TraceId) { trace = r.TraceId; container.append(node('h3', 'Trace ' + trace)); }
      const line = node('button', undefined, 'waterfall-row'); line.type = 'button';
      line.dataset.spanId = r.SpanId; line.dataset.traceId = r.TraceId;
      line.setAttribute('aria-pressed','false');
      const suffix = s.cycle ? ' · parent cycle' : s.missingParent ? ' · parent outside results' : '';
      const title = `${r.ServiceName} · ${r.SpanName}${suffix}`;
      line.setAttribute('aria-label', title);
      const label = node('span', title, 'span-label'); label.style.paddingLeft = Math.min(s.depth, 12)*12 + 'px';
      const track = node('span', undefined, 'span-track'), bar = node('span', undefined, 'span-bar');
      bar.style.left = s.left + '%'; bar.style.width = s.width + '%';
      bar.dataset.error = r.StatusCode === 'Error' ? 'true' : 'false';
      track.append(bar);
      line.append(label, track, node('span', `${Number(s.duration)/1e6} ms${r.StatusCode === 'Error' ? ' · error' : ''}`));
      line.onclick = () => {
        for(const other of container.querySelectorAll('.waterfall-row'))other.setAttribute('aria-pressed',String(other===line));
        details.replaceChildren(node('h3', r.SpanName), node('p', `${r.ServiceName} · ${r.duration_ms} ms · ${r.StatusCode || 'unset status'}`));
        global.GooAttributes.mount(details,r);
        for (const action of relatedActions?.(r) || []) {
          const b = node('button', action.label); b.onclick = action.select; details.append(b);
        }
        for (let i=0; i<(r.linked_traces || []).length; i++) {
          const targetTrace = r.linked_traces[i], targetSpan = (r.linked_spans || [])[i];
          const target = [...container.querySelectorAll('.waterfall-row')].find(n => n.dataset.spanId === targetSpan && n.dataset.traceId === targetTrace);
          if (target) { const b = node('button', 'Follow async link → ' + targetSpan); b.onclick = () => { target.focus(); target.click(); }; details.append(b); }
          else details.append(node('p', `Async link outside returned spans: ${targetTrace} / ${targetSpan || '(span unspecified)'}`));
        }
      };
      container.append(line);
    }
    container.append(details);
  }
  function heatmap(container, rows, data, select) {
    const from = data.evaluatedAt - data.expression.minutes*60000, to = data.evaluatedAt;
    const first = Math.floor(from/60000)*60000, last = Math.floor(to/60000)*60000;
    const cells = new Map(rows.map(r => [`${r.bucket_ms}:${r.band}`, Number(r.calls)]));
    const max = Math.max(1, ...rows.map(r => Number(r.calls)));
    const total = rows.reduce((sum,r) => sum+Number(r.calls),0);
    container.append(node('p', `${total} gRPC server spans · raw durations · UTC time · darker means more calls. Empty cells mean no observed spans; evidence completeness unknown. Select a populated cell to inspect its spans.`));
    const scroll=node('div',undefined,'heatmap-scroll'), table=node('table'); table.className='heatmap';
    const header=node('tr'); header.append(node('th','Duration / UTC'));
    const times=[]; for(let t=first;t<=last;t+=60000) times.push(t);
    for(const t of times) header.append(node('th',new Date(t).toISOString().slice(11,16)));
    table.append(header);
    for(let band=6;band>=0;band--) {
      const tr=node('tr'); tr.append(node('th',labels[band]));
      for(const t of times) {
        const count=cells.get(`${t}:${band}`)||0, td=node('td'), b=node('button',count ? String(count) : '·','heat-cell');
        b.disabled=!count; b.dataset.count=count;
        const begin=Math.max(from,t), end=Math.min(to,t+60000);
        b.setAttribute('aria-label',`${new Date(t).toISOString()} ${labels[band]}: ${count} spans`);
        b.title=b.getAttribute('aria-label');
        if(count) b.style.background=`rgba(57, 201, 173, ${0.18+0.82*Math.log1p(count)/Math.log1p(max)})`;
        b.onclick=()=>select({fromMs:begin,toMs:end,minMs:edges[band],maxMs:edges[band+1]});
        td.append(b); tr.append(td);
      }
      table.append(tr);
    }
    scroll.append(table); container.append(scroll);
  }
  function queue(container, rows, data) {
    container.append(node('p','Redis reservations / inventory · one-minute gauge snapshots per service instance; no averaging across instances. Queue wait p95 uses raw first-attempt observations, not stored percentiles. Evidence completeness unknown.'));
    const groups=[['Waiting and unacknowledged messages','{message}',['lab.queue.depth','lab.queue.pending']],['Active workers and dispatch capacity','{worker}',['lab.worker.active','lab.worker.capacity']],['Worker utilization (active / capacity)','1',['lab.worker.utilization']],['Observed enqueue-to-start wait p95','s',['lab.queue.wait.p95']]];
    const colors=['#67ddbb','#f4b48b','#a6bfff','#dfb2ef'];
    const from=data.evaluatedAt-data.expression.minutes*60000, to=data.evaluatedAt;
    for(const [title,unit,metrics] of groups) {
      const box=node('section');box.append(node('h3',title));
      const relevant=rows.filter(r=>metrics.includes(r.metric)&&r.unit===unit&&Number.isFinite(Number(r.value)));
      if(!relevant.length){box.append(node('p','No observations in this window — unknown, not zero.'));container.append(box);continue;}
      const max=Math.max(unit==='1'?1:0.001,...relevant.map(r=>Number(r.value)));
      const series=new Map();for(const r of relevant){const key=r.metric+' / '+r.instance;if(!series.has(key))series.set(key,[]);series.get(key).push(r);}
      const ns='http://www.w3.org/2000/svg',svg=document.createElementNS(ns,'svg');svg.setAttribute('viewBox','0 0 960 170');svg.setAttribute('role','img');svg.setAttribute('aria-label',title);svg.style.width='100%';
      function mark(tag,attrs,text){const n=document.createElementNS(ns,tag);for(const [k,v]of Object.entries(attrs))n.setAttribute(k,String(v));if(text!==undefined)n.textContent=text;svg.append(n);return n;}
      const x=t=>50+880*(t-from)/(to-from),y=v=>135-110*v/max;
      mark('path',{d:'M50 20 V135 H930',stroke:'#52677e',fill:'none'});
      mark('text',{x:5,y:25,fill:'#9db1c8','font-size':12},max.toFixed(2)+' '+unit);
      mark('text',{x:50,y:160,fill:'#9db1c8','font-size':12},new Date(from).toISOString().slice(11,19)+' UTC');
      mark('text',{x:930,y:160,fill:'#9db1c8','font-size':12,'text-anchor':'end'},new Date(to).toISOString().slice(11,19)+' UTC');
      let i=0;for(const [key,points]of series){points.sort((a,b)=>Number(a.sample_ms)-Number(b.sample_ms));const color=colors[i++%colors.length];let previous;
        for(const p of points){const px=x(Number(p.sample_ms)),py=y(Number(p.value));
          if(previous&&Number(p.minute_ms)-Number(previous.minute_ms)<=60000)mark('line',{x1:x(Number(previous.sample_ms)),y1:y(Number(previous.value)),x2:px,y2:py,stroke:color,'stroke-dasharray':'3 3'});
          const dot=mark('circle',{cx:px,cy:py,r:4,fill:data.evaluatedAt-Number(p.sample_ms)>30000?'none':color,stroke:color});const tip=document.createElementNS(ns,'title');tip.textContent=key+': '+p.value+' '+unit+' / '+p.samples+' samples';dot.append(tip);previous=p;
        }
        const last=points[points.length-1],stale=data.evaluatedAt-Number(last.sample_ms)>30000;
        box.append(node('p',key+' · last '+Number(last.value).toFixed(3)+' '+unit+(stale?' · STALE (>30 seconds); current state unknown':'')));
      }
      box.append(svg);container.append(box);
    }
  }
  const api={layout,waterfall,heatmap,queue};
  if(typeof module!=='undefined') module.exports=api; else global.GooVisuals=api;
})(typeof window==='undefined' ? globalThis : window);
