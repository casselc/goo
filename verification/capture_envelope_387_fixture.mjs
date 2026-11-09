import fs from 'node:fs';
import assert from 'node:assert/strict';
import {validate,validateReceiptReference} from '../ui/saved-capture-protocol.mjs';
const cases=JSON.parse(fs.readFileSync(0,'utf8'));
globalThis.fetch=()=>{throw Error('Capture admission/reference must not read telemetry')};
for(const [i,c]of cases.entries()){const before=JSON.stringify(c.page);let full=false,reference=false;try{validate(c.page,c.page.receipt?.id);full=true}catch{}try{validateReceiptReference(c.page,c.page.receipt?.id);reference=true}catch{}assert.equal(full,c.Full,'full envelope '+i+' '+c.Label);assert.equal(reference,c.Reference,'reference capability '+i+' '+c.Label);assert.equal(JSON.stringify(c.page),before,'admission mutated immutable transport')}
process.stdout.write(JSON.stringify({cases:cases.length}));
