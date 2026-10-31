module recent_scope_285 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 var plans:int
 action init=all{phase'=0,reads'=0,staged'=true,plans'=0}
 action open=all{phase==0,phase'=1,reads'=reads+0,staged'=true,plans'=0}
 action recent=all{phase==1,phase'=2,reads'=reads+2,staged'=false,plans'=1}
 action preview=all{phase==2,phase'=3,reads'=reads+0,staged'=false,plans'=1}
 action link=all{phase==3,phase'=4,reads'=reads+0,staged'=false,plans'=1}
 action child=all{phase==4,phase'=5,reads'=reads+0,staged'=true,plans'=1}
 action execute=all{phase==5,phase'=6,reads'=reads+2,staged'=false,plans'=1}
 action save=all{phase==6,phase'=7,reads'=reads+0,staged'=false,plans'=1}
 action reopen=all{phase==7,phase'=8,reads'=reads+0,staged'=true,plans'=1}
 action rerun=all{phase==8,phase'=9,reads'=reads+2,staged'=false,plans'=1}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=9 and plans>=0 and plans<=1
}
