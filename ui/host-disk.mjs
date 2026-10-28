/* Pinned host byte profile; app-owned loading is shared. */
import * as HostCounterPair from './host-counter-pair.mjs';
const global=globalThis;
const factory=HostCounterPair;
const api=factory.create({"metric": "system.disk.io", "dimension": "system.device", "direction": "disk.io.direction", "roles": ["read", "write"], "title": "Disk", "noun": "Device", "css": "host-disk"});
const candidates=api.candidates;
const choices=api.choices;
const result=api.result;
const mount=api.mount;
export {candidates,choices,result,mount};
