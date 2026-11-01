module paired_window_309 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var compares:int
 var saves:int
 action init=all{phase'=0,reads'=0,pages'=0,compares'=0,saves'=0}
 action narrow=all{phase==0,phase'=1,reads'=3,pages'=1,compares'=0,saves'=0}
 action parent=all{phase==1,phase'=2,reads'=4,pages'=1,compares'=0,saves'=1}
 action restCompared=all{phase==2,phase'=3,reads'=6,pages'=1,compares'=1,saves'=1}
 action promote=all{phase==3,phase'=4,reads'=6,pages'=1,compares'=1,saves'=1}
 action save=all{phase==4,phase'=5,reads'=6,pages'=1,compares'=1,saves'=2}
 action originalRun=all{phase==5,phase'=6,reads'=8,pages'=1,compares'=2,saves'=2}
 action invalidDraft=all{phase==6,phase'=7,reads'=8,pages'=1,compares'=2,saves'=2}
 action discardDraft=all{phase==7,phase'=8,reads'=8,pages'=1,compares'=2,saves'=2}
 action stageWindow=all{phase==8,phase'=9,reads'=8,pages'=1,compares'=2,saves'=2}
 action stale=all{phase==9,phase'=10,reads'=8,pages'=1,compares'=2,saves'=2}
 action saveWindow=all{phase==10,phase'=11,reads'=8,pages'=1,compares'=2,saves'=3}
 action reopen=all{phase==11,phase'=12,reads'=8,pages'=1,compares'=2,saves'=3}
 action newRun=all{phase==12,phase'=13,reads'=10,pages'=1,compares'=3,saves'=3}
 action encodeArtifact=all{phase==13,phase'=14,reads'=10,pages'=1,compares'=3,saves'=3}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and compares>=0 and compares<=3 and saves>=0 and saves<=3 and phase>=0 and phase<=14
}
