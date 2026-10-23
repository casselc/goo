/* Pure presentation of Go capability plans. No browser convention planner. */
(function(global){
 const version='goo/instrument-capability-plan/v1alpha1';
 function describe(i){const p=i?.capabilityPlan;if(p?.version===version&&p.verification==='not-run'&&Array.isArray(p.reasons)&&Array.isArray(p.hints))return structuredClone(p);return {version,profile:'unavailable',title:'Capability plan unavailable',eligible:false,verification:'not-run',reasons:['Discover again with a server that supplies a supported capability plan.'],hints:[],coverage:{state:'unavailable',streams:[],dimensions:[]}}}
 function catalog(items){return items.some(i=>describe(i).profile==='http-server-duration')?'HTTP server duration metadata observed. Convention plans do not verify data coverage.':'HTTP server duration not observed in this bounded window. This is not proof that the application lacks HTTP instrumentation.'}
 const api={describe,catalog,version};if(typeof module!=='undefined')module.exports=api;else global.GooInstrumentCapabilities=api;
})(typeof window==='undefined'?globalThis:window);
