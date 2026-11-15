/* Local buffer facts only. Presentation never infers population totals or health. */
export function summary({phase,message,gap,retained,evicted,observed,compact}){
 const detail=message+(gap?' · Continuity gap: recent observations restarted; no replay guarantee.':'');
 const facts=retained+' loaded · '+evicted+' evicted · '+observed+' observations received for this buffer. Not a population count.';
 let text=detail;
 if(compact&&['paused','stopped'].includes(phase))text=(phase==='paused'?'Paused':'Stopped')+(gap?' · continuity gap':'');
 return {text,detail,counts:compact?retained+' loaded'+(evicted?' · '+evicted+' evicted':''):retained+' retained observations · '+evicted+' evicted from this view · '+observed+' observations received for this buffer. Not a population count.',facts};
}
