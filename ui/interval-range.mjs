/* Passive key-range gestures over a caller-supplied, ordered interval track. */
export function bindIntervalRange(host,{plot,items,keys,onRange}){
 const boundPlot=plot();const live=()=>host.isConnected&&boundPlot?.isConnected&&plot()===boundPlot;
 let anchor=null,end=null,committed=null,pointer=null,ignoreClick=false;
 const at=e=>{const i=items().findIndex(item=>item===e.target||item.contains?.(e.target));return i};
 const paint=()=>{const range=anchor===null?committed:[Math.min(anchor,end),Math.max(anchor,end)];items().forEach((item,i)=>item.dataset.rangeSelected=String(!!range&&i>=range[0]&&i<=range[1]))};
 const release=()=>{const track=plot();if(pointer!==null&&track?.hasPointerCapture(pointer))track.releasePointerCapture(pointer);pointer=null};
 const commit=()=>{if(anchor===null||!live())return;const current=items();if(current.length!==keys.length||!current[anchor]||!current[end]||current.some((item,i)=>item.dataset.key!==keys[i])){anchor=end=null;return}committed=[Math.min(anchor,end),Math.max(anchor,end)];onRange(keys[committed[0]],keys[committed[1]]);anchor=end=null;paint()};
 const cancel=()=>{release();anchor=end=null;paint()};
 host.onpointerdown=e=>{const i=at(e);if(!live()||e.button!==0||e.pointerType==='touch'||i<0)return;ignoreClick=false;anchor=end=i;pointer=e.pointerId;const track=plot();track.onlostpointercapture=()=>{if(pointer!==null)cancel()};track.setPointerCapture(pointer);paint();e.preventDefault()};
 host.onpointermove=e=>{if(pointer!==e.pointerId)return;let nearest=0,distance=Infinity;items().forEach((item,i)=>{const r=item.getBoundingClientRect(),d=Math.abs(e.clientX-(r.left+r.width/2));if(d<distance){distance=d;nearest=i}});end=nearest;paint()};
 host.onpointerup=e=>{if(pointer!==e.pointerId)return;release();ignoreClick=true;commit()};host.onpointercancel=cancel;
 host.onkeydown=e=>{const i=at(e);if(!live()||i<0)return;if(e.key==='Escape'){cancel();e.preventDefault();return}if(e.key===' '){e.preventDefault();e.stopPropagation();if(anchor===null){anchor=end=i;paint()}else{end=i;commit()}return}if(e.shiftKey&&['ArrowLeft','ArrowRight'].includes(e.key)){e.preventDefault();anchor=anchor??i;end=Math.max(0,Math.min(items().length-1,i+(e.key==='ArrowRight'?1:-1)));items()[end].focus();paint()}if(e.key==='Enter'&&anchor!==null){e.preventDefault();e.stopPropagation();end=i;commit()}};
 return {clear(){committed=null;cancel()},ignoreClick(event){if(ignoreClick&&event.detail>0){ignoreClick=false;return true}return false}};
}
