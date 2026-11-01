module add_breakdown_304 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var captures:int
 var compares:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,captures'=0,compares'=0}
 action raw=all{phase==0,phase'=1,reads'=3,pages'=0,saves'=0,captures'=0,compares'=0}
 action rawRead=all{phase==1,phase'=2,reads'=3,pages'=1,saves'=0,captures'=0,compares'=0}
 action discard=all{phase==2,phase'=3,reads'=3,pages'=1,saves'=0,captures'=0,compares'=0}
 action add=all{phase==3,phase'=4,reads'=3,pages'=1,saves'=0,captures'=0,compares'=0}
 action grouped=all{phase==4,phase'=5,reads'=4,pages'=1,saves'=0,captures'=0,compares'=0}
 action save=all{phase==5,phase'=6,reads'=4,pages'=1,saves'=1,captures'=0,compares'=0}
 action capture=all{phase==6,phase'=7,reads'=4,pages'=1,saves'=1,captures'=1,compares'=0}
 action lateRead=all{phase==7,phase'=8,reads'=5,pages'=1,saves'=1,captures'=1,compares'=0}
 action compare=all{phase==8,phase'=9,reads'=5,pages'=1,saves'=1,captures'=1,compares'=1}
 action reopen=all{phase==9,phase'=10,reads'=5,pages'=1,saves'=1,captures'=1,compares'=1}
 action rerun=all{phase==10,phase'=11,reads'=6,pages'=1,saves'=1,captures'=1,compares'=1}
 action targetRaw=all{phase==11,phase'=12,reads'=9,pages'=2,saves'=1,captures'=1,compares'=1}
 action targetGrouped=all{phase==12,phase'=13,reads'=10,pages'=2,saves'=1,captures'=1,compares'=1}
 action apply=all{phase==13,phase'=14,reads'=10,pages'=2,saves'=1,captures'=1,compares'=1}
 action saveTarget=all{phase==14,phase'=15,reads'=10,pages'=2,saves'=2,captures'=1,compares'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=2 and saves>=0 and saves<=2 and captures>=0 and captures<=1 and compares>=0 and compares<=1 and phase>=0 and phase<=15
}
