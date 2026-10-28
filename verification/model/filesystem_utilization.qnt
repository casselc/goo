module filesystem_utilization {
 const MAX_READS:int
 var reads:int
 var plans:int
 var phase:int
 action init = all {reads'=0,plans'=0,phase'=0}
 action defaults = all {phase==0,reads'=3,plans'=1,phase'=1}
 action bytes = all {phase==1,reads'=reads+2,plans'=plans,phase'=2}
 action utilization = all {phase==2,reads'=reads+1,plans'=plans,phase'=3}
 action local = all {phase==3,reads'=reads,plans'=plans,phase'=phase}
 action parent = local
 action advance(s:int):bool = all {phase==s,reads'=reads+1,plans'=plans,phase'=s+1}
 action native = advance(3)
 action capture = advance(4)
 val safety = reads>=0 and reads<=MAX_READS and plans<=1
}
