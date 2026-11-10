// Frozen pre-extraction production behavior from 011580e3 (398).
// Test-only oracle: keep independent of the extracted implementation.
import * as NumericMeasures from '../../ui/numeric-values.mjs';
import * as FieldPresentation from '../../ui/field-presentation.mjs';
function percent(count,total){const n=BigInt(count),d=BigInt(total);if(d===0n)return null;const value=(n*10000n+d/2n)/d;return String(value/100n)+'.'+String(value%100n).padStart(2,'0');}
function durationMS(ns,count='1'){if(ns===undefined||ns===null||count===undefined||count===null)return 'Unavailable';const n=BigInt(ns),d=BigInt(count)*1000000n;if(d===0n)return '—';const v=(n*1000n+d/2n)/d;return String(v/1000n)+'.'+String(v%1000n).padStart(3,'0')+' ms'}
function groupMetric(group,metric){if(metric.startsWith('numeric-'))return NumericMeasures.magnitude(NumericMeasures.value(group.numeric,metric.slice(8)));if(metric==='error-rate')return group.spanStatus&&group.count!=='0'?{n:BigInt(group.spanStatus.error)*100n,d:BigInt(group.count)}:null;if(metric==='count')return {n:BigInt(group.count),d:1n};const d=group.duration;if(!d||d.state!=='observed')return null;if(metric==='mean')return {n:BigInt(d.sumNs),d:BigInt(d.count)};return {n:BigInt(d[{p95:'p95Ns',p99:'p99Ns',max:'maxNs'}[metric]]),d:1n};}
function compareMetric(a,b,metric){const x=groupMetric(a,metric),y=groupMetric(b,metric);if(!x||!y)return x?-1:y?1:0;const delta=x.n*y.d-y.n*x.d;if(delta!==0n)return delta>0n?-1:1;const ak=JSON.stringify(a.g0),bk=JSON.stringify(b.g0);return ak<bk?-1:ak>bk?1:0;}
function metricWidth(value,max){if(!value||!max||max.n===0n)return '0.00';const v=value.n*max.d*10000n/(value.d*max.n);return String(v/100n)+'.'+String(v%100n).padStart(2,'0');}
function valueLabel(value,field){if(field?.kind==='column'&&field.key==='body_pattern_v1')return (FieldPresentation).valueText(field,value);return value===null?((FieldPresentation)?.missingLabel?.(field)||'Missing key'):value===''?'Present empty string':JSON.stringify(value);}
export {percent,durationMS,groupMetric,compareMetric,metricWidth,valueLabel};
