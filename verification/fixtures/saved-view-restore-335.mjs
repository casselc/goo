import assert from 'node:assert/strict';
import fs from 'node:fs';
import {fixture,tick,question} from '../../tests/support/saved-view-owner.mjs';
const input=JSON.parse(fs.readFileSync(0,'utf8'));
let reads=0,adoptions=0,refused=0,telemetry=0,writes=0;
for(let round=0;round<input.rounds;round++){
 const f=fixture();await tick();
 const seed=input.seed+round,a=(seed%15).toString(16).repeat(64),b=((seed+1)%15).toString(16).repeat(64),order=seed%2?['hashchange','popstate']:['popstate','hashchange'];
 const duplicates=()=>{for(let n=0;n<input.duplicates;n++)f.emit(order[(seed+n)%2])};
 const expectReads=n=>{assert.equal(f.reads.length,n,'one definition GET per intentional visit');for(const r of f.reads){assert.match(r.url,/^\/api\/exploration-views\/[a-f0-9]{64}$/,'restore attempted non-definition IO');assert.equal(r.method,'GET','restore attempted a write')}};
 const settleRefused=async(request,n)=>{const count=f.adoptions.length,status=f.panel().textContent;if((seed+n)%2)request.resolve();else request.reject();await tick();assert.equal(f.adoptions.length,count,'departed response adopted');assert.equal(f.panel().textContent,status,'departed success/error published status');refused++};
 const resolveCurrent=async(request,n)=>{request.resolve();await tick();assert.equal(f.adoptions.length,n,'current definition did not reach caller admission');assert.deepEqual(f.adoptions.at(-1),question,'saved query changed at admission')};
 try{
  f.route(a,order);duplicates();expectReads(1);const first=f.reads[0];
  f.route(b,order);expectReads(2);assert(first.signal.aborted,'superseded definition was not aborted');await settleRefused(first,0);
  await resolveCurrent(f.reads[1],1);const adoptedStatus=f.panel().textContent;duplicates();expectReads(2);assert.equal(f.adoptions.length,1);assert.equal(f.panel().textContent,adoptedStatus,'passive event replaced post-admission panel');

  // A non-saved hash can remain in the Explore workspace. Cancellation must
  // retire ownership before even an abort-insensitive response is settled.
  f.route(a,order);expectReads(3);const departed=f.reads[2];location.hash=seed%2?'#question=other':'#explore';f.emit(order[0]);assert(departed.signal.aborted,'same-workspace departure did not abort');await settleRefused(departed,1);

  // The same bytes after departure are a new visit. Alternate ordinary and
  // queued ABA departures; both must prevent the original A response adopting.
  f.route(a,order);expectReads(4);const aba=f.reads[3];
  if(seed%2){f.emit('hashchange','#question=other');f.emit(order[0])}
  else {location.hash='#question=other';f.emit(order[0]);f.route(a,order)}
  expectReads(5);assert(aba.signal.aborted,'ABA did not retire original request');await settleRefused(aba,2);await resolveCurrent(f.reads[4],2);duplicates();expectReads(5);

  f.route(b,order);expectReads(6);f.reads[5].reject();await tick();const failedStatus=f.panel().textContent;assert.match(failedStatus,/View unavailable/);duplicates();expectReads(6);assert.equal(f.panel().textContent,failedStatus,'failed passive visit automatically retried');
  location.hash='#explore';f.emit(order[0]);f.route(a,order);expectReads(7);const suspended=f.reads[6];f.owner.suspend();assert(suspended.signal.aborted,'suspension did not retire restore');await settleRefused(suspended,3);
  f.route(b,order);expectReads(8);await resolveCurrent(f.reads[7],3);duplicates();expectReads(8);
  assert.equal(f.adoptions.length,3,'unexpected admission count');
  reads+=f.reads.length;adoptions+=f.adoptions.length;
  telemetry+=f.reads.filter(r=>!r.url.startsWith('/api/exploration-views/')).length;writes+=f.reads.filter(r=>r.method!=='GET').length;
 }finally{f.owner.dispose();await tick()}
}
process.stdout.write(JSON.stringify({rounds:input.rounds,reads,adoptions,refused,telemetry,writes}));
