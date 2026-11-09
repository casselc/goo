import fs from 'node:fs';
import assert from 'node:assert/strict';
import * as Native from '../ui/native-view-declaration.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(input);
globalThis.fetch=()=>{throw Error('Metadata admission must not read telemetry')};
let after='';for(const page of input.pages){Native.page(page,{after});after=page.nextCursor??''}
const reused=Native.reuse(input.source,input.destination);
assert.deepEqual(reused,input.reused);assert.deepEqual(reused.question,input.destination);assert.equal(Object.hasOwn(reused,'id'),false);assert.equal(JSON.stringify(input),before);
process.stdout.write(JSON.stringify({reused,pages:input.pages.length}));
