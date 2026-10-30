import {nativeRationalCoordinate} from './measurement-scope.mjs';
/* Family-aware display facts. Numeric parse support is never invented for spans. */
import {validateSnapshot} from './view-schema.mjs';
import {rationalDisplay} from './numeric-values.mjs';
export function histogramFacts(snapshot){const h=snapshot.histogram;return h.kind==='numeric'?{kind:h.kind,unit:h.binding.unit,total:h.support.total,represented:h.support.valid,support:h.support}:{kind:h.kind,unit:h.unit,total:h.total,represented:h.total}}
export function histogramBoundText(histogram,value){
 if(value==='+inf')return '+∞';
 if(value==='-inf')return '−∞';
 if(histogram.kind==='otel-explicit'){let exact=value;if(value.length<=256){try{exact=rationalDisplay(value)}catch{}}if(exact.length<=18&&!exact.includes('/'))return exact+' '+histogram.unit;const coordinate=nativeRationalCoordinate(value);return (coordinate===null?'Unavailable':'≈ '+Number(coordinate.toPrecision(6)).toString())+' '+histogram.unit;}
 if(histogram.kind==='numeric')return rationalDisplay(value)+' '+histogram.binding.unit;
 const [numerator,denominator='1']=value.split('/'),n=BigInt(numerator),d=BigInt(denominator),absolute=n<0n?-n:n;
 const [factor,unit]=absolute>=1000000000n*d?[1000000000n,'s']:absolute>=1000000n*d||n===0n?[1000000n,'ms']:absolute>=1000n*d?[1000n,'µs']:[1n,'ns'];
 return rationalDisplay(n+'/'+(d*factor))+' '+unit;
}
export const histogramRangeText=(histogram,from,to)=>histogram.kind==='otel-explicit'?'('+histogramBoundText(histogram,from)+', '+histogramBoundText(histogram,to)+(to==='+inf'?')':']'):histogram.kind==='numeric'?'['+rationalDisplay(from)+', '+rationalDisplay(to)+') '+histogram.binding.unit:'['+histogramBoundText(histogram,from)+', '+histogramBoundText(histogram,to)+')';
// Supplied duration-range facts only. This grants no permission to stage or read.
export function durationRangeFacts(snapshot,fromKey,toKey=fromKey){
 if(!validateSnapshot(snapshot).ok||snapshot?.shape!=='histogram'||snapshot.histogram?.kind!=='span-duration')throw TypeError('A validated span-duration histogram is required');
 const from=snapshot.rows.findIndex(row=>row.key===fromKey),to=snapshot.rows.findIndex(row=>row.key===toKey);
 if(from<0||to<from)throw TypeError('Choose ordered supplied duration bin keys');
 const fromNS=snapshot.rows[from].values.from,toNS=snapshot.rows[to].values.to,count=snapshot.rows.slice(from,to+1).reduce((n,row)=>n+BigInt(row.values.count),0n).toString();
 return {fromBin:from,toBin:to+1,fromNS,toNS,minMS:String(BigInt(fromNS)/1000000n),maxMS:toNS==='+inf'?'':String(BigInt(toNS)/1000000n),count,total:snapshot.histogram.total};
}
// Compact duration labels share an explicit millisecond axis unit in the view.
export function histogramCompactRangeText(histogram,from,to){
 if(histogram.kind!=='span-duration')return histogramRangeText(histogram,from,to);
 const ms=value=>{const n=BigInt(value);return n>=1000000000n&&n%1000000000n===0n?String(n/1000000000n)+'k':rationalDisplay(n+'/1000000')};
 return ms(from)+(to==='+inf'?'+':'–'+ms(to));
}
// Short native bin labels share the instrument unit and retain right closure.
export function nativeHistogramCompactRangeText(histogram,from,to){
 const compact=value=>histogramBoundText(histogram,value).replace(' '+histogram.unit,'').replace('≈ ','≈');
 if(from==='-inf'&&to==='+inf')return 'All values';
 if(from==='-inf')return '≤'+compact(to);
 if(to==='+inf')return '>'+compact(from);
 return compact(from)+'–'+compact(to);
}
