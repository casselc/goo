/* Closed raw-read routing metadata. No source access or family admission. */
const common=Object.freeze(['kind','scope','version','schemaURL','metric','unit']);
const sumFields=Object.freeze([...common,'temporality','monotonic']);
const histogramFields=Object.freeze([...common,'temporality']);
const gauge=Object.freeze({version:'goo/raw-gauge-observations/v1',path:'/api/metric-observations',label:'gauge',instrumentFields:common});
const sum=Object.freeze({version:'goo/raw-sum-observations/v1',path:'/api/sum-observations',label:'sum',instrumentFields:sumFields});
const histogram=Object.freeze({version:'goo/raw-histogram-observations/v1',path:'/api/histogram-observations',label:'histogram',instrumentFields:histogramFields});
const profiles=Object.freeze({gauge,counter:sum,stock:sum,histogram});
export function supports(kind){return typeof kind==='string'&&Object.hasOwn(profiles,kind)}
export function profile(kind){if(!supports(kind))throw Error('Raw observations do not support this instrument kind.');return profiles[kind]}
