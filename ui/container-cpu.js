/* CPU uses the same bounded source/inspection owner as memory, with separate contracts. */
(function(global){
 const owner=typeof module!=='undefined'?require('./container-memory'):global.GooContainerMemory,view=typeof module!=='undefined'?require('./container-cpu-view'):global.GooContainerCPUView;
 const api=owner.create({family:'cpu',label:'CPU',primary:'usage',primaryKey:'usage',primaryLabel:'Usage',unit:'{cpu}',metric:'container.cpu.usage',panelData:'containerCpu',loadData:'containerCpuLoad',view});
 if(typeof module!=='undefined')module.exports=api;else global.GooContainerCPU=api;
})(typeof window==='undefined'?globalThis:window);
