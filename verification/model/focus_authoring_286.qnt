module focus_authoring_286 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 var plans:int
 action init=all{phase'=0,reads'=0,staged'=true,plans'=0}
 action open=all{phase==0,phase'=1,reads'=reads+0,staged'=true,plans'=0}
 action recent=all{phase==1,phase'=2,reads'=reads+2,staged'=false,plans'=1}
 action pageRecords=all{phase==2,phase'=3,reads'=reads+1,staged'=false,plans'=1}
 action focusLogs=all{phase==3,phase'=4,reads'=reads+0,staged'=true,plans'=1}
 action authorLogs=all{phase==4,phase'=5,reads'=reads+0,staged'=true,plans'=1}
 action readLogs=all{phase==5,phase'=6,reads'=reads+1,staged'=false,plans'=1}
 action saveLogs=all{phase==6,phase'=7,reads'=reads+0,staged'=false,plans'=1}
 action returnParent=all{phase==7,phase'=8,reads'=reads+0,staged'=false,plans'=1}
 action focusSpans=all{phase==8,phase'=9,reads'=reads+0,staged'=true,plans'=1}
 action groupSpans=all{phase==9,phase'=10,reads'=reads+0,staged'=true,plans'=1}
 action authorSpans=all{phase==10,phase'=11,reads'=reads+0,staged'=true,plans'=1}
 action readSpans=all{phase==11,phase'=12,reads'=reads+1,staged'=false,plans'=1}
 action saveSpans=all{phase==12,phase'=13,reads'=reads+0,staged'=false,plans'=1}
 action reopen=all{phase==13,phase'=14,reads'=reads+0,staged'=true,plans'=1}
 action finish=all{phase==14,phase'=15,reads'=reads+1,staged'=false,plans'=1}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=15 and plans>=0 and plans<=1
}
