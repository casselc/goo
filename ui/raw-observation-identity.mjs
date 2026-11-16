/* Local supplied-data identity only; not a cryptographic source receipt. */
export const canonical=value=>JSON.stringify(value,(_,entry)=>entry&&typeof entry==='object'&&!Array.isArray(entry)?Object.fromEntries(Object.entries(entry).sort(([a],[b])=>a.localeCompare(b))):entry);
export function identity(value){let a=2166136261,b=2246822519;for(const byte of new TextEncoder().encode(canonical(value))){a=Math.imul(a^byte,16777619);b=Math.imul(b^byte,3266489917)}return (a>>>0).toString(16).padStart(8,'0')+(b>>>0).toString(16).padStart(8,'0')}
