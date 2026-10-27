// Finite observed macrosteps only. phase is an internal ordering guard;
// actual history withdrawal/publication and raw point/range identity live in Go.
module container_history {
 const MAX_READS:int
 var reads:int
 var plans:int
 var compositions:int
 var histories:int
 var phase:int
 action init = all {reads'=0,plans'=0,compositions'=0,histories'=0,phase'=0}
 action update(r:int,p:int,c:int,h:int,s:int):bool = all {reads'=r,plans'=p,compositions'=c,histories'=h,phase'=s}
 action metadata = all {phase==0,update(3,2,0,0,1)}
 action both_histories = all {phase==1,update(reads+6,plans,compositions+2,histories+2,2)}
 action local_point = all {phase==2,update(reads,plans,compositions,histories,phase)}
 action local_range = all {phase==2 or phase==3,update(reads,plans,compositions,histories,3)}
 action context_run_back = all {phase==3,update(reads+1,plans,compositions,histories,phase)}
 action native_stage = all {phase==3,update(reads,plans+1,compositions,histories,4)}
 action native_run_capture = all {phase==4,update(reads+2,plans,compositions,histories,5)}
 action frozen = all {phase==5,update(reads,plans,compositions,histories,6)}
 action hold_history = all {phase==0,update(3,1,1,1,7)}
 action withdraw = all {phase==7,update(reads,plans,compositions,histories,8)}
 action late_ignored = all {phase==8,update(reads,plans,compositions,histories,8)}
 action sibling_load = all {phase==7,update(reads+1,plans,compositions+1,histories+1,9)}
 action held_finishes = all {phase==9,update(reads,plans,compositions,histories,10)}
 val safety = reads<=MAX_READS and reads>=0 and plans>=0 and compositions>=0 and histories>=0 and histories<=compositions
}
