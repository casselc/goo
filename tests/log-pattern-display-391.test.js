const {test}=require('node:test'),assert=require('node:assert/strict');
const D=require('../ui/view-declaration.mjs'),V=require('../ui/view-model.mjs'),Values=require('../ui/view-values.mjs'),S=require('../ui/snapshots.mjs'),T=require('../ui/view-template-handoff.mjs');
const question=()=>({signal:'logs',environment:'prod',fromMs:1000,toMs:2000,limit:100,offset:0,groupBy:[{kind:'column',key:'body_pattern_v1'}]});
const ids=['lp1:p:72657120n','lp1:p:72657120e29faa646967697473e29fab','lp1:e','lp1:o','lp1:u'];
function snapshot(){return {version:'goo.dataset/1',id:'patterns',revision:'r1',origin:'source-read',scope:{id:'q1',generation:'r1',signal:'logs',selector:question()},shape:'ranked',dimensions:['category'],fields:[{key:'category',label:'Log pattern',type:'string',format:'log-pattern-v1'},{key:'count',label:'Captured records',type:'number',unit:'records'}],rows:ids.map((id,i)=>({key:id,values:{category:id,count:String(i+1)}})),quality:{extent:'complete',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'r1',scopeId:'q1',aspects:['groups']},partition:{kind:'disjoint-record-groups',measure:'count',total:'15'}}}
const adopt=raw=>S.createSourceAdapter({id:'patterns-test'}).adoptSnapshot(raw,{question:raw.scope.selector});
test('log-pattern display requires explicit count profile and reuses only presentation over exact destination',()=>{
 const q=question(),d=D.create(q),before=structuredClone(d);assert.equal(d.display.profile,'log-pattern-count-v1');assert.deepEqual(D.availableMeasures(q),['count']);
 for(const change of [d=>delete d.display.profile,d=>d.display.profile='',d=>d.display.profile=null,d=>d.display.profile='log-pattern-v2',d=>d.display.value='numeric.mean',d=>d.display.snapshot={}]){const copy=structuredClone(d);change(copy);assert.throws(()=>D.validate(copy,q));}
 assert.equal(D.eligible(q),true);
 for(const extra of [{numericHistogram:{}},{numeric:{field:{kind:'attribute',key:'size'},type:'int64',unit:'1',parse:'canonical-decimal'}},{summary:'span-duration'},{spanStatus:true},{groupTimeSeries:true},{reportedEdge:{}}]){assert.equal(D.eligible({...q,...extra}),false);assert.throws(()=>D.create({...q,...extra}));assert.throws(()=>D.validate(d,{...q,...extra}));}
 const target={...q,environment:'stage',fromMs:5000,toMs:9000,where:{field:{kind:'column',key:'service'},op:'eq',value:'worker'}};const identity=JSON.stringify(target),next=D.reuse(d,q,target);assert.deepEqual(next,d);assert.equal(JSON.stringify(target),identity);assert.deepEqual(d,before);
 assert.throws(()=>D.reuse(d,q,{...target,groupBy:[{kind:'column',key:'service'}]}));
 const ordinary={...q,groupBy:[{kind:'column',key:'service'}]};for(const profile of [undefined,'',null])assert.doesNotThrow(()=>D.validate({...D.create(ordinary),display:{...D.create(ordinary).display,profile}},ordinary));assert.throws(()=>D.validate(d,ordinary));
});
test('decoded labels feed tables, Find and order while canonical selection and full denominator survive',()=>{
 const raw=snapshot(),before=structuredClone(raw),s=adopt(raw),field=s.fields[0];assert.equal(V.validateSnapshot(s).ok,true);
 assert.equal(Values.formatCell(field,ids[0]),'req ⟪digits⟫');assert.equal(Values.categoryLabel(field,ids[1]),'req \\⟪digits⟫');
 assert.equal(Values.categoryLabel({type:'string'},ids[0]),JSON.stringify(ids[0]));
 const rows=V.displayRows(s,{category:'category',value:'count',query:'req',order:'label'});assert.equal(rows.length,2);
 for(const row of rows){assert.equal(V.populationShare(s,row.key).fraction,row.values.count+'/15');assert.equal(V.groupIntent(s,row.key).values[0].value,row.values.category)}
 for(const id of ids.slice(2)){assert.match(Values.categoryLabel(field,id),/unclassified/);assert.equal(V.groupIntent(s,id).values[0].value,id)}
 assert.deepEqual(raw,before);assert.equal(V.displayRows(s,{category:'category',value:'count',query:'no match'}).length,0);assert.equal(s.partition.total,'15');
 const control='lp1:p:610a09';assert.equal(Values.categoryLabel(field,control),'a\\n\\t');
});
test('inconsistent pattern reads retain counts but cannot obtain shares or complete-only through compatible declarations',()=>{
 const raw=snapshot();raw.quality.extent='unknown';delete raw.completeness;delete raw.partition;const s=adopt(raw),d=D.create(question());assert.equal(V.validateDisplay(s,{...D.presentation(d),kind:'bar',category:'category'}).ok,false);assert.equal(V.populationShare(s,ids[0]).ok,false);assert.equal(V.validateDisplay(s,{kind:'table',category:'category',value:'count',scale:'max'}).ok,true);assert.equal(V.validateDisplay(s,{kind:'table',category:'category',value:'count',requiresComplete:true}).ok,false);assert.equal(s.rows[0].values.count,'1');
});
test('pattern formats reject malformed keys and foreign scope; template expansion stays separate',()=>{
 for(const change of [s=>s.rows[0].values.category='friendly label',s=>delete s.fields[0].format,s=>s.fields[0].nullable=true,s=>s.rows={},s=>s.fields[0].format='log-pattern-v2',s=>s.scope.signal='spans',s=>s.scope.selector.groupBy[0].key='service']){const s=snapshot();change(s);assert.equal(V.validateSnapshot(s).ok,false)}
 assert.throws(()=>T.templateQuestion(question()));
});
test('response admission refuses complete-looking pattern authority from changed or mismatched reads',()=>{
 const Source=require('../ui/source-snapshots.mjs');function response(){const viewDataset=snapshot();return {question:question(),basis:'r1',capturedAt:'2026-11-09T12:00:00Z',state:'observed',count:'15',groups:[],viewDataset:{...viewDataset,capturedAt:'2026-11-09T12:00:00Z'}}}
 for(const mutate of [r=>r.state='changed_between_reads',r=>r.count='16']){const r=response();mutate(r);assert.throws(()=>Source.adoptExploration(r,r.question),/Inconsistent log-pattern/)}
 const r=response();r.state='changed_between_reads';r.viewDataset.quality.extent='unknown';delete r.viewDataset.completeness;delete r.viewDataset.partition;const admitted=Source.adoptExploration(r,r.question);assert.equal(V.populationShare(admitted.viewDataset,ids[0]).ok,false);assert.equal(admitted.viewDataset.rows[0].values.count,'1');
});
test('pattern actions retain canonical unclassified keys and lose authority with coherence or execution ownership',()=>{
 const {createActionContext}=require('../ui/scope-actions.mjs');let epoch=1,active=true;const s=adopt(snapshot()),context=createActionContext({epoch:()=>epoch,active:()=>active,canStage:()=>true});
 const bind=coherent=>context.bind({role:'current',slot:'groups',epoch,snapshot:s,question:s.scope.selector,coherent});let binding=bind(true),intent=V.groupIntent(s,'lp1:e');assert.equal(context.resolve(binding,intent).ok,true);active=false;assert.equal(context.resolve(binding,intent).ok,false);active=true;epoch++;assert.equal(context.resolve(binding,intent).ok,false);binding=bind(false);assert.equal(context.resolve(binding,intent).ok,false);assert.equal(s.rows.find(row=>row.key==='lp1:e').values.category,'lp1:e');
});
test('held pattern staging requires the original coherent population and exact applied basis/question',()=>{
 const {heldPatternMatches}=require('../ui/exploration.mjs');const data={question:question(),basis:'source-a',count:'3',state:'observed',groups:[{g0:'lp1:e',count:'1'},{g0:'lp1:p:72657120n',count:'2'}]},applied={question:structuredClone(data.question),basis:data.basis},before=structuredClone({data,applied});assert.equal(heldPatternMatches(data,applied),true);
 for(const change of [d=>d.state='changed_between_reads',d=>d.count='4',d=>d.groups[1].g0='readable label',d=>d.groups.push({...d.groups[0]}),d=>d.question.groupBy[0].key='service',d=>d.question.environment='stage',d=>d.question.fromMs++,d=>d.basis='another-source']){const copy=structuredClone(data);change(copy);assert.equal(heldPatternMatches(copy,applied),false)}
 const noData={...data,state:'no_data',count:'0',groups:[]};assert.equal(heldPatternMatches(noData,applied),true);assert.equal(heldPatternMatches(data,null),false);assert.deepEqual({data,applied},before);
});
