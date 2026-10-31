// Only explicit Runs acquire fixed records; template preview/application are configuration.
module authored_template_296 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var saves:int
 var staged:bool
 action init=all{phase'=0,reads'=0,saves'=0,staged'=true}
 action service=all{phase==0,phase'=1,reads'=1,saves'=0,staged'=false}
 action customize=all{phase==1,phase'=2,reads'=reads,saves'=saves,staged'=true}
 action readSource=all{phase==2,phase'=3,reads'=reads+1,saves'=saves,staged'=false}
 action saveRegular=all{phase==3,phase'=4,reads'=reads,saves'=saves+1,staged'=false}
 action disabledProfile=all{phase==4,phase'=5,reads'=reads,saves'=saves,staged'=false}
 action author=all{phase==5,phase'=6,reads'=reads,saves'=saves,staged'=false}
 action saveTemplate=all{phase==6,phase'=7,reads'=reads,saves'=saves+1,staged'=false}
 action target=all{phase==7,phase'=8,reads'=reads+1,saves'=saves,staged'=true}
 action groupTarget=all{phase==8,phase'=9,reads'=reads,saves'=saves,staged'=true}
 action readTarget=all{phase==9,phase'=10,reads'=reads+1,saves'=saves,staged'=false}
 action preview=all{phase==10,phase'=11,reads'=reads,saves'=saves,staged'=false}
 action apply=all{phase==11,phase'=12,reads'=reads,saves'=saves,staged'=false}
 action readApplied=all{phase==12,phase'=13,reads'=reads+1,saves'=saves,staged'=false}
 action saveDestination=all{phase==13,phase'=14,reads'=reads,saves'=saves+1,staged'=false}
 action reopen=all{phase==14,phase'=15,reads'=reads,saves'=saves,staged'=true}
 action rerun=all{phase==15,phase'=16,reads'=reads+1,saves'=saves,staged'=false}
 val safety=reads>=0 and reads<=MAX_READS and saves>=0 and saves<=3 and phase>=0 and phase<=16
}
