// Controlled ordinary OTLP, not application-call or collection-completeness proof.
const Base=require('../192-native-context-fixture');
const {question:baseQuestion}=require('../192-native-context-support');
const attrs=map=>Object.entries(map).map(([key,stringValue])=>({key,value:{stringValue}}));
const canonical=value=>JSON.stringify(value,(_key,item)=>item&&typeof item==='object'&&!Array.isArray(item)?Object.fromEntries(Object.entries(item).sort(([a],[b])=>a.localeCompare(b))):item);
function fixture(){
 const f=Base.fixture();
 f.records=f.records.filter(r=>r.namespace==='inventory'&&r.instance==='worker-a'&&(r.name.startsWith('inside-')&&Number(r.name.slice(7))<4||['at-lower','after-lower','before-upper','at-upper'].includes(r.name)));
 const selector={version:'goo/environment-selector/v1',policy:'current-only',scope:{mode:'named',value:f.environment}};
 const make=kind=>{const q=baseQuestion(f,kind);delete q.definition.environment;q.definition.environmentSelector=selector;q.definition.source.namespace={mode:'value',value:'inventory'};return q};
 f.histogram=make('histogram');f.counter=make('counter');f.counter.definition.source.services=['entry-worker'];f.counter.definition.title='Reported work intervals';
 f.gauge=structuredClone(f.counter);Object.assign(f.gauge.definition,{name:'reported-temperature',title:'Reported chamber observations',kind:'gauge',maxGapMs:500});Object.assign(f.gauge.definition.source,{metric:'work.temperature',unit:'Cel'});
 f.unknown=structuredClone(f.gauge);f.unknown.definition.name='unavailable-temperature';f.unknown.definition.title='Reported unavailable temperature';f.unknown.definition.source.metric='work.unknown-temperature';
 f.empty=structuredClone(f.gauge);f.empty.definition.name='unreported-temperature';f.empty.definition.source.metric='work.absent';
 f.inventory={environmentSelector:selector,namespace:{mode:'value',value:'inventory'},fromMs:Number(BigInt(f.fromNS)/1000000n)-1,toMs:Number((BigInt(f.emptyToNS)+999999n)/1000000n)};
 return f;
}
function metrics(f){
 const p=Base.metrics(f),group=p.resourceMetrics.find(x=>x.resource.attributes.some(a=>a.key==='service.namespace'&&a.value.stringValue==='inventory'));
 const histogram={...group,scopeMetrics:group.scopeMetrics.map(s=>({...s,metrics:s.metrics.filter(m=>m.histogram)}))};
 const ordinary=structuredClone(group);ordinary.resource.attributes=attrs({'service.name':'entry-worker','service.namespace':'inventory','service.instance.id':'worker-a','deployment.environment.name':f.environment});
 const dimensions=attrs({'metric.only':'not-a-record-filter'}),to=BigInt(f.toNS);
 ordinary.scopeMetrics[0].metrics=[
  {name:'work.completed',unit:'{work}',sum:{aggregationTemporality:1,isMonotonic:true,dataPoints:[
   {startTimeUnixNano:f.fromNS,timeUnixNano:f.toNS,asInt:'7',attributes:dimensions},
   {startTimeUnixNano:f.toNS,timeUnixNano:String(to+1000000000n),asInt:'0',attributes:dimensions},
   {startTimeUnixNano:String(to+1000000000n),timeUnixNano:f.emptyToNS,asInt:'7',flags:1,attributes:dimensions},
  ]}},
  {name:'work.temperature',unit:'Cel',gauge:{dataPoints:[
   {timeUnixNano:String(BigInt(f.fromNS)+500000000n),asDouble:-2,attributes:dimensions},
   {timeUnixNano:String(BigInt(f.fromNS)+1000000000n),asDouble:0,attributes:dimensions},
   {timeUnixNano:String(to+1500000000n),asDouble:0,flags:1,attributes:dimensions},
  ]}},
  {name:'work.unknown-temperature',unit:'Cel',gauge:{dataPoints:[{timeUnixNano:String(to+1500000000n),asDouble:0,flags:1,attributes:dimensions}]}},
 ];
 return{resourceMetrics:[histogram,ordinary]};
}
module.exports={fixture,metrics,traces:Base.traces,logs:Base.logs,canonical};
