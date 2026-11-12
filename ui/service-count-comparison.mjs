/* Configuration only: the dashboard retains accepted-result and action ownership. */
import {validateRecordQuestion} from './question-protocol.mjs';
import {validate as validatePredicate} from './predicate-validation.mjs';
import {identity} from './question-session.mjs';
import * as CapturedCount from './captured-count-comparison.mjs';
import * as Declaration from './view-declaration.mjs';
import * as Composition from './view-composition.mjs';
import * as Environment from './record-environment.mjs';
import * as ServiceIdentity from './service-identity.mjs';

export function project(question){
 validateRecordQuestion(question);
 if(question.where)validatePredicate(question.where,question.signal);
 if(!['logs','spans'].includes(question.signal)||(question.offset??0)!==0||(question.limit??100)>100||question.numeric||question.numericHistogram||question.reportedEdge||question.groupTimeSeries||question.groupBy?.length!==1||question.groupBy[0].kind==='column'&&['body_pattern_v1','duration_bucket'].includes(question.groupBy[0].key))throw Error('Choose a page-zero ordinary single-group record widget with a limit up to 100.');
 const query=structuredClone(question);
 // These directives request additional measures; they do not select records.
 delete query.summary;delete query.spanStatus;
 validateRecordQuestion(query);
 CapturedCount.previousQuery(query);
 return query;
}

function assertService(query,{service,scope}){
 const conjuncts=p=>!p||p.any?[]:p.all?p.all.flatMap(conjuncts):[p],leaves=conjuncts(query.where);
 const required=[{field:{kind:'column',key:'service'},op:'eq',value:service},ServiceIdentity.predicate(scope.namespace)].filter(Boolean);
 if(query.fromMs!==scope.fromMs||query.toMs!==scope.toMs||!Environment.sameQuestionEnvironment(query,scope)||required.some(expected=>!leaves.some(p=>p.field?.kind===expected.field.kind&&p.field.key===expected.field.key&&p.op===expected.op&&(p.value??'')===(expected.value??''))))throw Error('Widget question differs from the selected service scope.');
}

export function create(widget,data,context){
 if(!widget?.query||widget.measurement||!data?.question||!['observed','no_data'].includes(data.state))throw Error('Read a coherent ordinary record widget before comparing.');
 const query=project(widget.query);
 if(context)assertService(widget.query,context);
 if(identity(widget.query)!==identity(data.question))throw Error('Widget result differs from its fixed question.');
 const baseline=CapturedCount.previousQuery(query);
 const composition=Composition.setComparison(Composition.create(query,Declaration.create(query)),CapturedCount.createQuery(query,baseline,'count'),query);
 return {version:1,title:'Captured records · previous window',queries:[query],composition};
}
