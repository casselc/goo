// Finite observed read effects, not arbitrary concurrent UI schedules.
module resource_defaults {
 const MAX_READS:int
 var reads:int
 var plans:int
 var compositions:int
 var histories:int
 var phase:int
 action init = all {reads'=0,plans'=0,compositions'=0,histories'=0,phase'=0}
 action update(r:int,p:int,c:int,h:int,s:int):bool = all {reads'=r,plans'=p,compositions'=c,histories'=h,phase'=s}
 action usage = all {phase==0,update(4,1,2,2,1)}
 action capacity = all {phase==1,update(reads+2,plans,compositions+1,histories+1,2)}
 action local = all {phase==2,update(reads,plans,compositions,histories,phase)}
 action ambiguous = all {phase==2,update(reads+2,plans+1,compositions+1,histories+1,3)}
 val safety = reads<=MAX_READS and reads>=0 and histories==compositions
}
