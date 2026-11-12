import {test} from 'node:test';
import assert from 'node:assert/strict';
import * as Dashboard from '../ui/service-dashboard.mjs';

test('incoming HTTP response context uses literal status groups and independent recorded span Error',()=>{
 const widget={key:'http-server-responses',family:'http-server',query:{signal:'spans',groupBy:[{kind:'attribute',key:'http.response.status_code'}]}},profile=Dashboard.recordFamily(widget),context=Dashboard.dimensionProfile(profile,'http.response.status_code');
 assert.equal(Dashboard.widgetTitle(widget,{profile:'http-server'}),'Reported HTTP response status · HTTP server spans');
 assert.equal(context.diagnosis,'HTTP response context');assert.equal(context.columnLabel,'Reported HTTP response status');assert.equal(context.openLabel,'Open response-status population');
 assert.match(context.dimensionMeaning,/literal reported HTTP response status across routes\/methods/);assert.match(context.dimensionMeaning,/missing and empty remain distinct/);assert.match(context.dimensionMeaning,/separate from recorded span Error/);
 assert.equal(Dashboard.dimensionProfile(profile,'http.route').dimensionTitle,'Routes · HTTP server spans');
 for(const value of [null,'','200','503','0503','Missing key','unexpected']){
  const q={signal:'spans',environment:'prod',fromMs:1000,toMs:2000,limit:3,offset:0,summary:'span-duration',spanStatus:true,where:{all:[{field:{kind:'column',key:'service'},op:'eq',value:'api'},{field:{kind:'resource',key:'service.namespace'},op:'eq',value:''},{field:{kind:'column',key:'kind'},op:'eq',value:'Server'},{field:{kind:'attribute',key:'http.request.method'},op:'ne',value:''}]},groupBy:[{kind:'attribute',key:'http.response.status_code'}]},before=structuredClone(q),child=Dashboard.diagnosisQuestion({question:q},value);
  assert.deepEqual(child.where.all[0],q.where);assert.deepEqual(child.where.all[1],{field:q.groupBy[0],op:value===null?'missing':'eq',...(value===null?{}:{value})});assert.equal(child.limit,3);assert.equal(child.spanStatus,true);assert.equal(child.summary,'span-duration');assert.deepEqual(q,before);
 }
});
