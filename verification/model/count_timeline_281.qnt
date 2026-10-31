module count_timeline_281 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 var held:bool
 action init=all{phase'=0,reads'=0,staged'=true,held'=true}
 action read=all{phase==0,phase'=1,reads'=reads+1,staged'=false,held'=true}
 action configure=all{phase==1,phase'=2,reads'=reads+0,staged'=true,held'=true}
 action execute=all{phase==2,phase'=3,reads'=reads+1,staged'=false,held'=true}
 action page=all{phase==3,phase'=4,reads'=reads+1,staged'=false,held'=true}
 action stage=all{phase==4,phase'=5,reads'=reads+0,staged'=true,held'=true}
 action narrow=all{phase==5,phase'=6,reads'=reads+1,staged'=false,held'=true}
 action reset=all{phase==6,phase'=7,reads'=reads+0,staged'=true,held'=true}
 action save=all{phase==7,phase'=8,reads'=reads+0,staged'=true,held'=true}
 action reopen=all{phase==8,phase'=9,reads'=reads+0,staged'=true,held'=true}
 action rerun=all{phase==9,phase'=10,reads'=reads+1,staged'=false,held'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=10 and held
}
