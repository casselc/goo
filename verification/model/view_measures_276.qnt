// Saved measure intent and explicit source reads; upstream completeness is not modeled.
module view_measures_276 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var measure:int
 var blocked:bool
 action init=all{phase'=0,reads'=0,measure'=0,blocked'=false}
 action staged=all{phase==0,phase'=1,reads'=reads,measure'=measure,blocked'=blocked}
 action read=all{phase==1,phase'=2,reads'=reads+1,measure'=measure,blocked'=blocked}
 action numeric=all{phase==2,phase'=3,reads'=reads,measure'=1,blocked'=false}
 action duration=all{phase==3,phase'=4,reads'=reads,measure'=2,blocked'=false}
 action invalid=all{phase==4,phase'=5,reads'=reads,measure'=measure,blocked'=true}
 action apply=all{phase==5,phase'=6,reads'=reads,measure'=measure,blocked'=false}
 action save=all{phase==6,phase'=7,reads'=reads,measure'=measure,blocked'=false}
 action reopen=all{phase==7,phase'=8,reads'=reads,measure'=measure,blocked'=false}
 action rerun=all{phase==8,phase'=9,reads'=reads+1,measure'=measure,blocked'=false}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=9
}
