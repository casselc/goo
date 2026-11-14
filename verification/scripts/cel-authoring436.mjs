import readline from 'node:readline';
import {AuthoringSlot} from '../../ui/cel-authoring.mjs';
const lines=readline.createInterface({input:process.stdin})[Symbol.asyncIterator]();
const receive=async()=>JSON.parse((await lines.next()).value);
const {base,baseIdentity}=await receive();
let sourceReads=0;globalThis.fetch=()=>{sourceReads++;throw Error('Authoring made source IO')};
const options={base,baseIdentity,parentRevision:'parent-1',slot:'condition',epoch:'mount-1'};
const owner=new AuthoringSlot(options),ledger=[];let currentToken=null;
function record(action,accepted=false){const state=owner.snapshot();ledger.push({action,ready:owner.stage(currentToken)!==null,pending:state.phase==='checking',active:state.active,alive:state.phase!=='disposed',lastValid:state.lastValid!==null,accepted,reads:sourceReads})}
async function check(){const attempt=owner.beginCheck();record('check');process.stdout.write(JSON.stringify({request:attempt.request})+'\n');return {attempt,response:await receive()}}
const text='log.body == "a"';owner.edit(text);record('edit');
const old=await check();owner.edit('log.body == "b"');record('edit');owner.edit(text);record('edit');
record('stale',owner.accept(old.attempt.token,old.response.receipt));
const latest=await check();currentToken=latest.attempt.token;record('accept',owner.accept(currentToken,latest.response.receipt));
record('stale',owner.stage(old.attempt.token)!==null);
owner.edit('log.body ==');record('edit');const invalid=await check();record('failed',owner.fail(invalid.attempt.token,invalid.response.diagnostic));
record('stale',owner.accept(latest.attempt.token,latest.response.receipt));
owner.edit(text);record('edit');const hidden=await check();owner.setVisible(false);record('hide');record('stale',owner.accept(hidden.attempt.token,hidden.response.receipt));owner.setVisible(true);record('show');record('stale',owner.fail(hidden.attempt.token,{code:'syntax',phase:'parse',message:'late'}));
currentToken=owner.clear();record('clear');
owner.replaceParent({...options,parentRevision:'parent-2'});record('reparent');record('stale',owner.stage(currentToken)!==null);
owner.edit(text);record('edit');const retired=await check();owner.dispose();record('dispose');record('stale',owner.accept(retired.attempt.token,retired.response.receipt));record('stale',owner.fail(retired.attempt.token,{code:'syntax',phase:'parse',message:'late'}));
process.stdout.write(JSON.stringify({ledger})+'\n');process.exit(0);
