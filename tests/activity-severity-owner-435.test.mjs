import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import vm from 'node:vm';

test('fixed editor retains a requested numeric severity projection when editing log filters',()=>{
 const source=readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
 const body=source.match(/function withFilters\(q\)\{([^\n]+)\}/)?.[0];assert.ok(body);
 const state={question:{logSeverity:true},filterError:null,numericControls:{error:()=>null,get:()=>undefined,getHistogram:()=>undefined},draftWhere:()=>undefined,FilterEditor:{validate(){},normalize:()=>undefined},edgeScope:null,splitInput:{checked:false},bucketInput:{value:'10000'},statusInput:{checked:false},summaryInput:{checked:false},el:()=>({value:'logs'}),filterEditor:{grouping:()=>[]}};
 vm.runInNewContext(body+';result=withFilters({signal:"logs",fromMs:1,toMs:60001})',state);
 assert.equal(state.result.logSeverity,true);assert.equal(state.result.fromMs,1);assert.equal(state.result.toMs,60001);
 vm.runInNewContext('result=withFilters({signal:"spans"})',state);assert.equal(state.result.logSeverity,undefined);
});

test('Activity numeric band handoff only stages complete predicates and focuses explicit Apply',()=>{
 const source=readFileSync(new URL('../ui/stream-entry.mjs',import.meta.url),'utf8'),begin=source.indexOf('logContext=ActivityLogContext.mount('),statement=source.slice(begin,source.indexOf('\n',begin));assert.ok(begin>=0);
 const parent={any:[{field:{kind:'column',key:'service'},op:'eq',value:'worker'},{field:{kind:'column',key:'service'},op:'eq',value:'api'}]},where={all:[parent,{all:[{field:{kind:'column',key:'severity_number'},op:'gte',value:'17'},{field:{kind:'column',key:'severity_number'},op:'lte',value:'20'}]}]};
 for(const [clean,native] of [[true,true],[false,true],[true,false]]){
  const calls=[],apply={},opener={},state={volumeHost:{},toolbar:{},ActivityLogContext:{mount(_host,options){state.options=options;return {}}},groupable:()=>clean,nativeControlReady:()=>native,filters:{replace:value=>calls.push(['stage',value])},draftScope:(...args)=>calls.push(['draft',...args]),revealFocus:element=>calls.push(['focus',element]),apply};
  vm.runInNewContext(statement,state);state.options.onStage(where,opener);
  assert.deepEqual(calls,clean&&native?[['stage',where],['draft',true,true],['focus',apply]]:[]);
 }
});
