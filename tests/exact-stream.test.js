const {test}=require('node:test'),assert=require('node:assert/strict');
const {streamQuestion}=require('../ui/measurements'),w=require('../ui/workspace');
test('stream focus preserves exact bounds and intent through sharing and capture',()=>{
 const stream={resource:{'service.name':'api','service.instance.id':'one','deployment.environment.name':'prod','service.namespace':''},resourceSchemaURL:'resource-v1',scope:'otel.http',version:'1',schemaURL:'otel-v1',name:'requests',unit:'1',attributes:{route:'',method:'GET'}};
 const q={definition:{kind:'counter',environment:'prod',source:{scope:'otel.http',metric:'requests',unit:'1'}},fromMs:1000,toMs:3001,fromNS:'1000000123',toNS:'3000000123'},r={question:q};
 const focused=streamQuestion(r,stream);assert.equal(q.definition.source.exactStream,undefined);
 const step=w.readStepLink(w.stepLink(w.measurementStep(focused)));assert.deepEqual(w.captureQuestion(step).measurement,focused);
 assert.ok(w.evidenceDifference(w.measurementStep(q),{kind:'measurement',measurement:focused}).includes('Question definition differs'));
 stream.attributes.method='POST';assert.equal(focused.definition.source.exactStream.attributes.method,'GET');
 assert.deepEqual(streamQuestion({question:focused},null),q);assert.equal(focused.definition.source.exactStream.resource['service.namespace'],'');
 assert.throws(()=>streamQuestion({question:{definition:{kind:'lifecycle'}}},stream));
});
