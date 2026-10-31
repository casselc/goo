module record_step_handoff_284 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 var pinned:bool
 action init=all{phase'=0,reads'=0,staged'=true,pinned'=true}
 action open=all{phase==0,phase'=1,reads'=reads+0,staged'=true,pinned'=true}
 action start=all{phase==1,phase'=2,reads'=reads+1,staged'=false,pinned'=true}
 action analyze=all{phase==2,phase'=3,reads'=reads+0,staged'=true,pinned'=true}
 action execute=all{phase==3,phase'=4,reads'=reads+1,staged'=false,pinned'=true}
 action share=all{phase==4,phase'=5,reads'=reads+0,staged'=false,pinned'=true}
 action link=all{phase==5,phase'=6,reads'=reads+0,staged'=true,pinned'=true}
 action pin=all{phase==6,phase'=7,reads'=reads+0,staged'=true,pinned'=true}
 action recover=all{phase==7,phase'=8,reads'=reads+0,staged'=true,pinned'=true}
 action savedoc=all{phase==8,phase'=9,reads'=reads+0,staged'=true,pinned'=true}
 action question=all{phase==9,phase'=10,reads'=reads+0,staged'=true,pinned'=true}
 action rerun=all{phase==10,phase'=11,reads'=reads+1,staged'=false,pinned'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=11 and pinned
}
