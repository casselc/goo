/* Renderer-neutral span geometry. Adapted from observability/lakeui's relative
 * time waterfall approach; subtract exact ns BEFORE converting to Number. */
(function (global) {
  const edges = [0, 10, 50, 100, 500, 1000, 10000, 18446744073710];
  const labels = ['0–10 ms', '10–50 ms', '50–100 ms', '100–500 ms', '0.5–1 s', '1–10 s', '≥10 s'];
  const waterfallAPI=()=>typeof module!=='undefined'?require('./waterfall'):global.GooWaterfall;
  const layout=rows=>waterfallAPI().layout(rows);
  function node(tag, text, cls) { const n = document.createElement(tag); if (text !== undefined) n.textContent = text; if (cls) n.className = cls; return n; }
  function waterfall(container, rows, truncated, relatedActions, detailsId = 'spanDetails') {
    const element=node('goo-trace-waterfall');element.data={rows,truncated:!!truncated};container.append(element);
    if(!rows.length)return element;
    const details=node('section');details.id=detailsId;container.append(details);const inspector=global.GooUI.inspector(details,{headingLevel:3});
    element.addEventListener('click',event=>{
      const line=event.target.closest('.waterfall-row');if(!line||!element.contains(line))return;
      const index=Number(line.dataset.rowIndex),r=rows[index];if(!r)return;
      element.selectedIndex=index;
      const key=String(index);if(inspector.has(key)){inspector.show({key,title:r.SpanName,opener:line});return}
      const content=node('div'),description=node('div');description.append(node('span',r.ServiceName+' · '+global.GooUI.durationMS(r.duration_ns)+' · '+(r.StatusCode||'unset status')));if(r.timestamp_ns)description.append(' · ',global.GooEvidenceSummary.timestamp(r.timestamp_ns));
      global.GooAttributes.mount(content,r);
      const actions=[];for(const action of relatedActions?.(r)||[]){const button=node('button',action.label);button.onclick=action.select;actions.push(button)}
      for(let i=0;i<(r.linked_traces||[]).length;i++){
        const targetTrace=r.linked_traces[i],targetSpan=(r.linked_spans||[])[i];
        const target=[...element.querySelectorAll('.waterfall-row')].find(n=>n.dataset.spanId===targetSpan&&n.dataset.traceId===targetTrace);
        if(target){const button=node('button','Follow async link → '+targetSpan);button.onclick=()=>{target.focus();target.click()};actions.push(button)}
        else content.append(node('p',`Async link outside returned spans: ${targetTrace} / ${targetSpan||'(span unspecified)'}`));
      }
      inspector.show({key,title:r.SpanName,description,content,actions,opener:line});
    });
    return element;
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
