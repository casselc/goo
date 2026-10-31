module saved_view_finding_282 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 var pinned:bool
 action init=all{phase'=0,reads'=0,staged'=true,pinned'=false}
 action read=all{phase==0,phase'=1,reads'=reads+1,staged'=false,pinned'=false}
 action dirty=all{phase==1,phase'=2,reads'=reads,staged'=false,pinned'=false}
 action saveview=all{phase==2,phase'=3,reads'=reads,staged'=false,pinned'=false}
 action attach=all{phase==3,phase'=4,reads'=reads,staged'=false,pinned'=true}
 action savedoc=all{phase==4,phase'=5,reads'=reads,staged'=false,pinned'=true}
 action reopen=all{phase==5,phase'=6,reads'=reads,staged'=false,pinned'=true}
 action definition=all{phase==6,phase'=7,reads'=reads,staged'=true,pinned'=true}
 action execute=all{phase==7,phase'=8,reads'=reads+1,staged'=false,pinned'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=8
}
