// Authored time views retain destination acquisition; only explicit Run/Page reads.
module time_template_299 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var staged:bool
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,staged'=true}
 action service=all{phase==0,phase'=1,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action customize=all{phase==1,phase'=2,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action stageSource=all{phase==2,phase'=3,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action readSource=all{phase==3,phase'=4,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action author=all{phase==4,phase'=5,reads'=reads,pages'=pages,saves'=saves,staged'=false}
 action saveTemplate=all{phase==5,phase'=6,reads'=reads,pages'=pages,saves'=saves+1,staged'=false}
 action target=all{phase==6,phase'=7,reads'=reads+1,pages'=pages,saves'=saves,staged'=true}
 action groupTarget=all{phase==7,phase'=8,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action readTarget=all{phase==8,phase'=9,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action missing=all{phase==9,phase'=10,reads'=reads,pages'=pages,saves'=saves,staged'=false}
 action wrongStage=all{phase==10,phase'=11,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action wrongRead=all{phase==11,phase'=12,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action mismatch=all{phase==12,phase'=13,reads'=reads,pages'=pages,saves'=saves,staged'=false}
 action matchingStage=all{phase==13,phase'=14,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action matchingRead=all{phase==14,phase'=15,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action preview=all{phase==15,phase'=16,reads'=reads,pages'=pages,saves'=saves,staged'=false}
 action apply=all{phase==16,phase'=17,reads'=reads,pages'=pages,saves'=saves,staged'=false}
 action readApplied=all{phase==17,phase'=18,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 action page=all{phase==18,phase'=19,reads'=reads,pages'=pages+1,saves'=saves,staged'=false}
 action saveDestination=all{phase==19,phase'=20,reads'=reads,pages'=pages,saves'=saves+1,staged'=false}
 action reopen=all{phase==20,phase'=21,reads'=reads,pages'=pages,saves'=saves,staged'=true}
 action rerun=all{phase==21,phase'=22,reads'=reads+1,pages'=pages,saves'=saves,staged'=false}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and saves>=0 and saves<=2 and phase>=0 and phase<=22
}
