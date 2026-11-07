import fs from 'node:fs';
import * as Triage from '../ui/error-triage.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(input);
globalThis.fetch=()=>{throw Error('A triage continuation must not acquire source data')};
const triage=Triage.triageQuery(input.parent,{rule:'status'}),inspect=Triage.inspectGroup(triage,input.group),compare=Triage.compareInService(triage,input.group.g0);
process.stdout.write(JSON.stringify({triage,inspect,compare,parentUnchanged:JSON.stringify(input)===before}));
