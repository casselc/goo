// Whole composition reuse transfers intent, never source authority or paging.
module whole_composition_280 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var composed:bool
 var aggregateHeld:bool
 action init=all{phase'=0,reads'=0,composed'=false,aggregateHeld'=true}
 action read=all{phase==0,phase'=1,reads'=reads+1,composed'=false,aggregateHeld'=true}
 action blocked=all{phase==1,phase'=2,reads'=reads,composed'=false,aggregateHeld'=true}
 action apply=all{phase==2,phase'=3,reads'=reads,composed'=true,aggregateHeld'=true}
 action page=all{phase==3,phase'=4,reads'=reads+1,composed'=true,aggregateHeld'=true}
 action inspect=all{phase==4,phase'=5,reads'=reads,composed'=true,aggregateHeld'=true}
 action save=all{phase==5,phase'=6,reads'=reads,composed'=true,aggregateHeld'=true}
 action reopen=all{phase==6,phase'=7,reads'=reads,composed'=true,aggregateHeld'=true}
 action rerun=all{phase==7,phase'=8,reads'=reads+1,composed'=true,aggregateHeld'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=8
}
