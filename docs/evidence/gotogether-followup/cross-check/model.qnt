// One bounded coordinator; identity/custody checkpoints, not telemetry completeness.
module cross {
  var phase: int
  var subject: str
  var trace: str
  var environment: str
  var count: int
  var members: str
  var captureId: str
  var digest: str
  action init = all {phase'=0, subject'="", trace'="", environment'="", count'=0, members'="", captureId'="", digest'=""}
  action generate(id: str): bool = all {phase==0, id!="", phase'=1, subject'=id, trace'=trace, environment'=environment, count'=count, members'=members, captureId'=captureId, digest'=digest}
  action collect(id: str, tid: str, env: str, n: int, ids: str): bool = all {phase==1,id==subject,tid!="",env!="",n>0,n<=100,ids!="",phase'=2,subject'=subject,trace'=tid,environment'=env,count'=n,members'=ids,captureId'=captureId,digest'=digest}
  action inspect(id: str, tid: str, env: str, n: int, ids: str): bool = all {phase==2,id==subject,tid==trace,env==environment,n==count,ids==members,phase'=3,subject'=subject,trace'=trace,environment'=environment,count'=count,members'=members,captureId'=captureId,digest'=digest}
  action capture(id: str, hash: str): bool = all {phase==3,id!="",hash!="",phase'=4,subject'=subject,trace'=trace,environment'=environment,count'=count,members'=members,captureId'=id,digest'=hash}
  action reopen(id: str, hash: str): bool = all {phase==4,id==captureId,hash==digest,phase'=5,subject'=subject,trace'=trace,environment'=environment,count'=count,members'=members,captureId'=captureId,digest'=digest}
  val safety = phase>=0 and phase<=5 and (phase<2 or count>0)
  val done = phase==5
}
