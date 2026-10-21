const {test}=require('node:test'),assert=require('node:assert/strict');
const {model}=require('../ui/value-timeline'),{timelineSnapshot}=require('../ui/measurements');
const point=(key,value,fromNS='10',toNS=fromNS)=>({key,value,fromNS,toNS,label:key});
test('timeline geometry preserves exact nanoseconds, signed/scientific values, zero and unknown',()=>{
 const data={fromNS:'900719925474099300',toNS:'900719925474099400',points:[point('negative','-1e-3','900719925474099310'),point('zero','0','900719925474099320'),point('positive','0.001','900719925474099340','900719925474099360'),{key:'gap',fromNS:'900719925474099360',toNS:'900719925474099380',label:'Unknown',state:'unknown'}]},before=structuredClone(data),m=model(data);
 assert.equal(m.state,'observed');assert.equal(m.zeroY,50);assert.deepEqual(m.points.map(p=>p.left),[10,20,40,60]);assert.deepEqual(m.points.map(p=>p.width),[0,0,20,20]);assert.deepEqual(m.points.map(p=>p.y),[0,50,100,null]);assert.equal(m.points[3].value,null);assert.deepEqual(data,before);
 const huge=model({fromNS:'0',toNS:'100',points:[point('a','18446744073709551615'),point('b','18446744073709551614','20')]});assert.equal(huge.max,'18446744073709551615');assert.equal(huge.points[0].y,100);assert.equal(huge.points[1].value,'18446744073709551614');assert.ok(huge.points[1].y<=100);
 assert.equal(model({fromNS:'0',toNS:'100',points:[point('tiny','1e-308')]}).points[0].y,100);
});
test('ambiguous supplied points fail closed rather than clipping timestamps or inventing missing values',()=>{
 const valid={fromNS:'0',toNS:'100',points:[point('a','1')]};
 for(const data of [{...valid,toNS:'0'},{...valid,points:[point('a','1','101')]},{...valid,points:[point('a','1','20','10')]},{...valid,points:[point('a','NaN')]},{...valid,points:[point('a','1e9999')]},{...valid,points:[{...point('a','0'),state:'unknown'}]},{...valid,points:[point('a','1'),point('a','2')]},{...valid,points:Array.from({length:257},(_,i)=>point(String(i),'1'))}])assert.equal(model(data).state,'unavailable');
 assert.equal(model({...valid,points:[]}).state,'no_data');const missing=model({...valid,points:[{key:'gap',fromNS:'10',toNS:'20',label:'Unknown interval',state:'unknown'}]});assert.equal(missing.numeric,false);assert.equal(missing.points[0].value,null);const zero=model({...valid,points:[point('zero','0')]});assert.equal(zero.numeric,true);assert.equal(zero.points[0].y,zero.zeroY);
});
test('native adapter retains typed interval rates, stale samples and histogram ranges without computing rates or scalar percentiles',()=>{
 const stream={unit:'{request}'},r={question:{definition:{kind:'counter'}},counters:{fromNS:'0',toNS:'100'}};
 const s={stream,intervals:[{fromNS:'0',toNS:'20',state:'ok',perSecond:5},{fromNS:'20',toNS:'40',state:'unknown',reason:'counter-decrease',perSecond:999},{fromNS:'60',toNS:'80',state:'ok',perSecond:0}]},before=structuredClone(s),m=timelineSnapshot(r,s);assert.equal(m.points[0].value,'5');assert.equal(m.points[1].value,undefined);assert.equal(m.points[2].value,'0');assert.equal(m.points[2].fromNS,'60');assert.deepEqual(s,before);
 const hist={question:{definition:{kind:'histogram'}},histograms:{fromNS:'0',toNS:'100'}},h=timelineSnapshot(hist,{stream:{unit:'s'},intervals:[{fromNS:'0',toNS:'20',state:'ok',distribution:{count:'0',unit:'s',p95:{state:'no_data'}}},{fromNS:'20',toNS:'40',state:'ok',distribution:{count:'4',unit:'s',p95:{state:'unbounded-upper',lower:.5}}}]});assert.equal(h.points[0].value,'0');assert.match(h.points[0].label,/P95: no observations/);assert.match(h.points[1].label,/\(0.5, \+∞\)/);assert.equal(h.unit,'observations per interval');
 const g=timelineSnapshot({question:{definition:{kind:'gauge'}},gauges:{fromNS:'0',toNS:'100'}},{stream:{unit:'1'},samples:[{atNS:'20',state:'stale',value:-2},{atNS:'30',state:'unknown',value:99},{atNS:'40',state:'observed',value:0}]});assert.equal(g.points[0].state,'stale');assert.equal(g.points[0].value,'-2');assert.equal(g.points[1].value,undefined);assert.equal(g.points[2].value,'0');
 const bounded=timelineSnapshot(r,{stream,intervals:Array.from({length:205},(_,i)=>({fromNS:String(i),toNS:String(i+1),state:'unknown'}))});assert.equal(bounded.points.length,200);assert.equal(bounded.omitted,5);assert.equal(bounded.total,205);assert.equal(bounded.points[0].key,'5');
});
test('generated visual positions stay bounded and monotonic without mutating exact values',()=>{
 for(let n=1;n<=100;n++){const points=Array.from({length:20},(_,i)=>point(String(i),String((i-10)*n),String(i*5))),m=model({fromNS:'0',toNS:'100',points});assert.equal(m.state,'observed');for(let i=0;i<m.points.length;i++){assert.ok(m.points[i].y>=0&&m.points[i].y<=100);if(i)assert.ok(m.points[i].y>=m.points[i-1].y);assert.equal(m.points[i].value,points[i].value)}}
});

test('brush geometry uses exact supplied timestamps, clamps no invalid inputs and treats reverse drags identically',()=>{
 const {brushRange}=require('../ui/value-timeline'),data={fromNS:'900719925474099300',toNS:'900719925474099400'},before=structuredClone(data);
 assert.deepEqual(brushRange(data,.2,.8),{fromNS:'900719925474099320',toNS:'900719925474099380'});assert.deepEqual(brushRange(data,.8,.2),brushRange(data,.2,.8));assert.equal(brushRange(data,.5,.5),null);
 for(const pair of [[-1,1],[0,1.1],[NaN,1],[0,Infinity]])assert.equal(brushRange(data,...pair),null);
 for(let i=1;i<100;i++){const r=brushRange(data,0,i/100);assert.equal(BigInt(r.fromNS),BigInt(data.fromNS));assert.ok(BigInt(r.toNS)>BigInt(r.fromNS)&&BigInt(r.toNS)<=BigInt(data.toNS))}assert.deepEqual(data,before);
});
