/* Shared local search over service identity and discovery metadata. */
import {label} from './service-identity.mjs';
function matches(row,query=''){
 const text=[row.service,label(row.namespace),...Object.keys(row.signals||{}),...(row.declarations||[]).map(d=>d.title+' '+d.name)].join(' ');
 return text.toLowerCase().includes(query.toLowerCase());
}
export {matches};
