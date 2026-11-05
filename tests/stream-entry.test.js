const{test}=require('node:test'),assert=require('node:assert/strict');const{fresh,EntryGate}=require('../ui/stream-entry.mjs');
test('only a fresh visible record entry can claim its one automatic monitor',()=>{for(const hash of ['','#explore']){assert.equal(fresh(hash),true);const g=new EntryGate(hash);for(const patch of [{active:false},{visible:false},{records:false}])assert.equal(g.claim({active:true,visible:true,records:true,...patch}),false);assert.equal(g.claim({active:true,visible:true,records:true}),true);assert.equal(g.claim({active:true,visible:true,records:true}),false);g.pause();assert.equal(g.paused,true);assert.equal(g.claim({active:true,visible:true,records:true}),false);g.explicit();assert.equal(g.paused,false)}});
test('saved and fixed intent never auto-starts even after visibility changes',()=>{for(const hash of ['#questions','#step=x','#question=x','#exploration-view=x','#capture=x','#document=x','#services','#investigations']){const g=new EntryGate(hash);assert.equal(fresh(hash),false);for(let i=0;i<3;i++){g.pause();assert.equal(g.claim({active:true,visible:true,records:true}),false)}g.explicit();assert.equal(g.started,true)}});

const{ScopeFilters}=require('../ui/stream-entry.mjs');
test('live predicates preserve scope namespaces, empty strings and immutable applied state',()=>{const s=new ScopeFilters(),f={kind:'attribute',key:'region'};s.stage(f,'','include');s.stage({kind:'resource',key:'region'},'west','exclude');s.apply();const old=s.where(true);s.stage(f,'','exclude');assert.equal(s.dirty,true);assert.equal(old.all[0].op,'eq');assert.equal(s.where(true).all[0].op,'eq');assert.equal(s.where().all[0].op,'ne');assert.equal(s.where().all[0].value,'');old.all[0].value='mutated';assert.equal(s.where(true).all[0].value,'');s.remove(0);assert.equal(s.where().all[0].field.kind,'resource');s.apply();assert.equal(s.dirty,false);s.reset();assert.equal(s.where(),undefined);assert.equal(s.where(true),undefined)});
test('live filter drafts reject missing/coerced values and enforce a bounded selector',()=>{const s=new ScopeFilters();for(const value of [undefined,null,1,{},[]])assert.throws(()=>s.stage({kind:'attribute',key:'a'},value,'include'));for(let i=0;i<16;i++)s.stage({kind:'attribute',key:'k'+i},'v','include');assert.throws(()=>s.stage({kind:'attribute',key:'overflow'},'v','include'));s.stage({kind:'attribute',key:'k0'},'v','exclude');assert.equal(s.draft.length,16);assert.equal(s.draft[0].op,'ne');assert.equal(s.applied.length,0)});

// Generic startup retains Activity entry while requiring an explicit first selector.
test('explicit startup gate never claims on route or visibility changes and can start deliberately',()=>{for(const hash of ['','#explore']){const gate=new EntryGate(hash,{automatic:false});assert.equal(gate.eligible,true);for(let i=0;i<3;i++)assert.equal(gate.claim({active:true,visible:true,records:true}),false);assert.equal(gate.started,false);gate.explicit();assert.equal(gate.started,true);assert.equal(gate.paused,false);assert.equal(gate.claim({active:true,visible:true,records:true}),false)}});

test('private service return lease survives hidden source but rejects retirement and ABA',()=>{
 const {serviceReturnCurrent}=require('../ui/stream-entry.mjs'),anchor={},expected={revision:2,anchor,selectorKey:'scope',scopeEpoch:4},actual={...expected,retained:true,connected:true,enabled:false};
 assert.equal(serviceReturnCurrent(expected,actual),true);
 for(const patch of [{revision:3},{anchor:{}},{selectorKey:'other'},{scopeEpoch:5},{retained:false},{connected:false},{enabled:true}])assert.equal(serviceReturnCurrent(expected,{...actual,...patch}),false);
 // Restoring identical selector text cannot revive an older scope epoch or anchor.
 assert.equal(serviceReturnCurrent(expected,{...actual,selectorKey:'scope',scopeEpoch:6}),false);
});

const {previousWindowView}=require('../ui/stream-entry.mjs');
function comparisonContext(){return {active:true,clean:true,phase:'paused',selector:{signal:'logs',environmentSelector:{version:'goo/environment-selector/v1',policy:'effective',scope:{mode:'unreported'}},where:{all:[{field:{kind:'resource',key:'tenant'},op:'eq',value:''},{any:[{field:{kind:'attribute',key:'region'},op:'eq',value:'west'},{field:{kind:'attribute',key:'region'},op:'missing'}]}]}},readWindow:{fromMs:120000,toMs:180000},rows:Array.from({length:100},()=>({record:{Body:'sample'}})),evicted:7}}
test('Activity previous-window planner freezes full checked scopes and preserves typed policies and predicates',()=>{
 const input=comparisonContext(),before=structuredClone(input),layout={columns:['utc','summary'],fieldColumns:[{kind:'resource',key:'tenant'}]},view=previousWindowView(input,layout),q=view.queries[0],pair=view.composition.children.find(c=>c.comparison).comparison;
 assert.deepEqual(q,{...input.selector,...input.readWindow,groupBy:[{kind:'column',key:'service'}],limit:100,offset:0});
 assert.deepEqual(pair.baselineQuery,{...q,fromMs:60000,toMs:120000});
 assert.deepEqual(view.composition.children.find(c=>c.records).records.display,layout);
 assert.equal(pair.kind.name,'query-count-comparison');assert.equal(JSON.stringify(view).includes('sample'),false);assert.equal(Object.hasOwn(view,'count'),false);
 require('../ui/view-template-handoff.mjs').stagedView(view,q);
 q.where.all[0].value='mutated';pair.baselineQuery.environmentSelector.scope.mode='empty';layout.columns.push('service');assert.deepEqual(input,before);assert.equal(view.composition.children.find(c=>c.records).records.display.columns.length,2);
});
test('Activity previous-window planner refuses dirty, expired and incomplete owners before staging',()=>{
 const c=comparisonContext();for(const patch of [{active:false},{active:undefined},{clean:false},{clean:undefined},{selector:null},{readWindow:null},...['loading','reconnecting','stale','stopped','unavailable','no_data',undefined].map(phase=>({phase}))])assert.throws(()=>previousWindowView({...c,...patch}));
 for(const readWindow of [{fromMs:0,toMs:60000},{fromMs:-1,toMs:59999},{fromMs:120000,toMs:120000},{fromMs:120000,toMs:180001},{fromMs:Number.MAX_SAFE_INTEGER,toMs:Number.MAX_SAFE_INTEGER+1}])assert.throws(()=>previousWindowView({...c,readWindow}));
 for(const phase of ['unknown','partial','paused'])assert.equal(previousWindowView({...c,phase}).queries[0].toMs,180000);
});
test('Activity previous-window planner keeps named, empty and missing typed environment selection exact',()=>{
 for(const policy of ['current-only','effective'])for(const scope of [{mode:'named',value:'prod'},{mode:'empty'},{mode:'unreported'}]){const c=comparisonContext();c.selector.environmentSelector={version:'goo/environment-selector/v1',policy,scope};if(policy==='current-only'&&scope.mode!=='named'){assert.throws(()=>previousWindowView(c));continue}const v=previousWindowView(c),q=v.queries[0],b=v.composition.children.find(c=>c.comparison).comparison.baselineQuery;assert.deepEqual(q.environmentSelector,c.selector.environmentSelector);assert.deepEqual(b.environmentSelector,c.selector.environmentSelector)}
});
