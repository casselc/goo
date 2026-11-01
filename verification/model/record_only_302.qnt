module record_only_302 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var keeps:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,keeps'=0}
 action sourceFixed=all{phase==0,phase'=1,reads'=1,pages'=0,saves'=0,keeps'=0}
 action sourceCustomize=all{phase==1,phase'=2,reads'=3,pages'=0,saves'=0,keeps'=0}
 action sourceRead=all{phase==2,phase'=3,reads'=3,pages'=1,saves'=0,keeps'=0}
 action listInspect=all{phase==3,phase'=4,reads'=3,pages'=1,saves'=0,keeps'=0}
 action keepLog=all{phase==4,phase'=5,reads'=3,pages'=1,saves'=0,keeps'=1}
 action saveLog=all{phase==5,phase'=6,reads'=3,pages'=1,saves'=1,keeps'=1}
 action pageLog=all{phase==6,phase'=7,reads'=3,pages'=2,saves'=1,keeps'=1}
 action reopenLog=all{phase==7,phase'=8,reads'=3,pages'=2,saves'=1,keeps'=1}
 action rerunLog=all{phase==8,phase'=9,reads'=3,pages'=3,saves'=1,keeps'=1}
 action targetRead=all{phase==9,phase'=10,reads'=6,pages'=4,saves'=1,keeps'=1}
 action applyLog=all{phase==10,phase'=11,reads'=6,pages'=4,saves'=1,keeps'=1}
 action saveTarget=all{phase==11,phase'=12,reads'=6,pages'=4,saves'=2,keeps'=1}
 action spanRead=all{phase==12,phase'=13,reads'=9,pages'=5,saves'=2,keeps'=1}
 action waterfall=all{phase==13,phase'=14,reads'=9,pages'=5,saves'=2,keeps'=1}
 action keepSpan=all{phase==14,phase'=15,reads'=9,pages'=5,saves'=2,keeps'=2}
 action pageSpan=all{phase==15,phase'=16,reads'=9,pages'=6,saves'=2,keeps'=2}
 action saveSpan=all{phase==16,phase'=17,reads'=9,pages'=6,saves'=3,keeps'=2}
 action reopenSpan=all{phase==17,phase'=18,reads'=9,pages'=6,saves'=3,keeps'=2}
 action rerunSpan=all{phase==18,phase'=19,reads'=9,pages'=7,saves'=3,keeps'=2}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=7 and saves>=0 and saves<=3 and keeps>=0 and keeps<=2 and phase>=0 and phase<=19
}
