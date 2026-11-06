// Only the three observed explicit Run gestures acquire telemetry. The model
// covers the observed editor protocol; DOM and data authority are checked first.
module fixed_question_358 {
 const INITIAL:int
 var phase:int
 var opened:bool
 var reads:int
 var coherent:bool
 action init=all{phase'=INITIAL,opened'=false,reads'=0,coherent'=true}
 action edit_clean=all{phase==1,phase'=1,opened'=true,reads'=reads,coherent'=true}
 action hide_clean=all{phase==1,opened,phase'=1,opened'=false,reads'=reads,coherent'=true}
 action edit_invalid=all{phase==1,phase'=2,opened'=true,reads'=reads,coherent'=true}
 action invalid_run_blocked=all{phase==2,phase'=phase,opened'=true,reads'=reads,coherent'=true}
 action draft_return=all{phase==2,phase'=phase,opened'=true,reads'=reads,coherent'=true}
 action valid_fix=all{phase==2,phase'=3,opened'=true,reads'=reads,coherent'=true}
 action run_valid=all{phase==3,phase'=1,opened'=false,reads'=reads+1,coherent'=true}
 action edit_before_cancel=all{phase==1,phase'=phase,opened'=true,reads'=reads,coherent'=true}
 action cancel_read=all{phase==1,opened,phase'=4,opened'=true,reads'=reads+1,coherent'=true}
 action cancel_reading=all{phase==4,phase'=5,opened'=true,reads'=reads,coherent'=true}
 action late_result=all{phase==5,phase'=phase,opened'=true,reads'=reads,coherent'=true}
 action run_recover=all{phase==5,phase'=1,opened'=false,reads'=reads+1,coherent'=true}
 action inspect=all{phase==1,phase'=phase,opened'=opened,reads'=reads,coherent'=true}
 action applied_return=all{phase==1,reads==3,phase'=6,opened'=false,reads'=reads,coherent'=true}
 val safety=and{phase>=1,phase<=6,reads>=0,reads<=3,coherent,or{phase!=2,opened},or{phase!=5,opened}}
}
