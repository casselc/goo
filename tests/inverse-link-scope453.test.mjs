import test from 'node:test';
import assert from 'node:assert/strict';
import * as Population from '../ui/population-workspace.mjs';
import * as TraceSearch from '../ui/trace-search.mjs';
import * as Summary from '../ui/summary-declaration.mjs';
import * as Histogram from '../ui/histogram-declaration.mjs';
import * as Facet from '../ui/facet-declaration.mjs';
import * as Template from '../ui/view-template-handoff.mjs';
import * as Declaration from '../ui/view-declaration.mjs';
import * as ExceptionGroups from '../ui/exception-groups.mjs';
import * as Cohort from '../ui/count-cohort-context.mjs';
import * as Package from '../ui/telemetry-package-intent.mjs';
const linksTo={traceId:'a'.repeat(32),spanId:'b'.repeat(16)};
const q={signal:'spans',environment:'prod',fromMs:1000,toMs:61000,limit:100,offset:0,linksTo,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'}};
const group={...q,groupBy:[{kind:'column',key:'service'}]};
test('population response admission includes inverse-link target identity',()=>{
 const plan={version:'goo/population/v1',queries:[q],notices:[]};assert.deepEqual(Population.plannedQueries(q,null,plan),[q]);
 for(const next of [{...q,linksTo:undefined},{...q,linksTo:{...linksTo,spanId:'c'.repeat(16)}}])assert.throws(()=>Population.plannedQueries(q,null,{...plan,queries:[next]}),/changed environment policy or predicates/);
});
test('fixed population transforms retain inverse selector without mutating parent',()=>{
 const before=structuredClone(q);const focused=Population.focusedQuestion([q],'spans');assert.deepEqual(focused.linksTo,linksTo);assert.notEqual(focused.linksTo,linksTo);
 const recent=Population.recentQuestion(q,120000,1);assert.deepEqual(recent.linksTo,linksTo);
 const grouped=Population.composeGrouping(q,{kind:'column',key:'name'},'replace');assert.deepEqual(grouped.linksTo,linksTo);assert.deepEqual(q,before);
});
test('unsupported trace search and specialized projections refuse instead of broadening',()=>{
 const seed={field:{kind:'column',key:'name'},op:'eq',value:'consume'};
 assert.throws(()=>TraceSearch.seedCondition(q,seed),/reported-link/);assert.throws(()=>TraceSearch.buildRequest(q,[seed]),/reported-link/);
 assert.equal(Summary.eligible(q),false);assert.equal(Facet.eligible(q),false);assert.equal(Histogram.eligible(q),false);
 assert.throws(()=>Template.templateQuestion(q));assert.throws(()=>ExceptionGroups.question(q));assert.equal(Cohort.eligible(group),false);
});
test('ordinary aggregate display preserves selector, package definitions project it exactly',()=>{
 const declaration=Declaration.create(group);assert.equal(Declaration.eligible(group),true);assert.equal(Declaration.validate(declaration,group).kind.name,'bar');
 const definition={source:{adapter:'clickhouse.otel/v1',signal:'spans',environment:'prod'},linksTo,where:q.where};const projected=Package.packageQuestion(definition,q);assert.deepEqual(projected.linksTo,linksTo);assert.notEqual(projected.linksTo,linksTo);
});
