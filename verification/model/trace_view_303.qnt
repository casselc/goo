module trace_view_303 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var keeps:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,keeps'=0}
 action parent=all{phase==0,phase'=1,reads'=3,pages'=0,saves'=0,keeps'=0}
 action inspectTrace=all{phase==1,phase'=2,reads'=5,pages'=0,saves'=0,keeps'=0}
 action inspectorPage=all{phase==2,phase'=3,reads'=7,pages'=0,saves'=0,keeps'=0}
 action stage=all{phase==3,phase'=4,reads'=7,pages'=0,saves'=0,keeps'=0}
 action read=all{phase==4,phase'=5,reads'=7,pages'=1,saves'=0,keeps'=0}
 action list=all{phase==5,phase'=6,reads'=7,pages'=1,saves'=0,keeps'=0}
 action keep=all{phase==6,phase'=7,reads'=7,pages'=1,saves'=0,keeps'=1}
 action page=all{phase==7,phase'=8,reads'=7,pages'=2,saves'=0,keeps'=1}
 action save=all{phase==8,phase'=9,reads'=7,pages'=2,saves'=1,keeps'=1}
 action stale=all{phase==9,phase'=10,reads'=7,pages'=2,saves'=1,keeps'=1}
 action back=all{phase==10,phase'=11,reads'=7,pages'=2,saves'=1,keeps'=1}
 action reopen=all{phase==11,phase'=12,reads'=7,pages'=2,saves'=1,keeps'=1}
 action rerun=all{phase==12,phase'=13,reads'=7,pages'=3,saves'=1,keeps'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=3 and saves>=0 and saves<=1 and keeps>=0 and keeps<=1 and phase>=0 and phase<=13
}
