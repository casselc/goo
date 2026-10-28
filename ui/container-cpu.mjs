/* CPU uses the same bounded source/inspection owner as memory, with separate contracts. */
import * as ContainerCPUView from './container-cpu-view.mjs';
import * as ContainerMemory from './container-memory.mjs';
const global=globalThis;
const owner=ContainerMemory,view=ContainerCPUView;
const api=owner.create({family:'cpu',label:'CPU',primary:'usage',primaryKey:'usage',primaryLabel:'Usage',unit:'{cpu}',metric:'container.cpu.usage',panelData:'containerCpu',loadData:'containerCpuLoad',view});
const eligible=api.eligible;
const variants=api.variants;
const request=api.request;
const candidates=api.candidates;
const measurement=api.measurement;
const options=api.options;
const mount=api.mount;
export {eligible,variants,request,candidates,measurement,options,mount};
