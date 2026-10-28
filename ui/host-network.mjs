/* Pinned host byte profile; app-owned loading is shared. */
import * as HostCounterPair from './host-counter-pair.mjs';
const global=globalThis;
const factory=HostCounterPair;
const api=factory.create({"metric": "system.network.io", "dimension": "network.interface.name", "direction": "network.io.direction", "roles": ["receive", "transmit"], "title": "Network", "noun": "Interface", "css": "host-network"});
const candidates=api.candidates;
const choices=api.choices;
const result=api.result;
const mount=api.mount;
export {candidates,choices,result,mount};
