import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import * as Native from '../ui/native-view-declaration.mjs';
const bank=JSON.parse(readFileSync(new URL('../testdata/protocol/native-histogram-view-questions-373.json',import.meta.url)));
test('finite native configuration wire cases agree with browser admission without altering inputs',()=>{
 assert.equal(bank.version,'goo.native-histogram-question-cases/1');assert(bank.cases.length>=50);
 for(const {name,question,runtimeValid}of bank.cases){const before=JSON.stringify(question);let accepted=true;try{Native.question(question)}catch{accepted=false}assert.equal(accepted,runtimeValid,name);assert.equal(JSON.stringify(question),before,name+' mutated')}
});
