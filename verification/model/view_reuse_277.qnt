// Reuse changes presentation only; source acquisition remains explicit.
module view_reuse_277 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var reused:bool
 action init=all{phase'=0,reads'=0,reused'=false}
 action population=all{phase==0,phase'=1,reads'=3,reused'=false}
 action comparison=all{phase==1,phase'=2,reads'=reads+1,reused'=false}
 action inspect=all{phase==2,phase'=3,reads'=reads,reused'=false}
 action apply=all{phase==3,phase'=4,reads'=reads,reused'=true}
 action save=all{phase==4,phase'=5,reads'=reads,reused'=true}
 action reopen=all{phase==5,phase'=6,reads'=reads,reused'=true}
 action rerun=all{phase==6,phase'=7,reads'=reads+2,reused'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=7
}
