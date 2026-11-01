module authored_cohort_307 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var compares:int
 var saves:int
 action init=all{phase'=0,reads'=0,pages'=0,compares'=0,saves'=0}
 action narrow=all{phase==0,phase'=1,reads'=3,pages'=1,compares'=0,saves'=0}
 action parent=all{phase==1,phase'=2,reads'=4,pages'=1,compares'=0,saves'=1}
 action include=all{phase==2,phase'=3,reads'=4,pages'=1,compares'=0,saves'=1}
 action stageParent=all{phase==3,phase'=4,reads'=4,pages'=1,compares'=0,saves'=1}
 action parentCompared=all{phase==4,phase'=5,reads'=5,pages'=1,compares'=1,saves'=1}
 action backParent=all{phase==5,phase'=6,reads'=5,pages'=1,compares'=1,saves'=1}
 action stale=all{phase==6,phase'=7,reads'=5,pages'=1,compares'=1,saves'=1}
 action excludeFacet=all{phase==7,phase'=8,reads'=5,pages'=2,compares'=1,saves'=1}
 action stageRest=all{phase==8,phase'=9,reads'=5,pages'=2,compares'=1,saves'=1}
 action restCompared=all{phase==9,phase'=10,reads'=7,pages'=2,compares'=2,saves'=1}
 action save=all{phase==10,phase'=11,reads'=7,pages'=2,compares'=2,saves'=2}
 action backRest=all{phase==11,phase'=12,reads'=7,pages'=2,compares'=2,saves'=2}
 action reopen=all{phase==12,phase'=13,reads'=7,pages'=2,compares'=2,saves'=2}
 action rerun=all{phase==13,phase'=14,reads'=8,pages'=2,compares'=2,saves'=2}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=2 and compares>=0 and compares<=2 and saves>=0 and saves<=2 and phase>=0 and phase<=14
}
