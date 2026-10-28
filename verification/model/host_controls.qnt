module host_controls {
 const MAX_READS:int
 var reads:int
 var plans:int
 var phase:int
 action init = all {reads'=0,plans'=0,phase'=0}
 action defaults = all {phase==0,reads'=3,plans'=1,phase'=1}
 action advance(s:int,n:int):bool = all {phase==s,reads'=reads+n,plans'=plans,phase'=s+1}
 action network = advance(1,2)
 action disk = advance(2,2)
 action bytes = advance(3,2)
 action utilization = advance(4,1)
 action local = all {phase==5,reads'=reads,plans'=plans,phase'=phase}
 action refresh = advance(5,2)
 action native = advance(6,1)
 action capture = advance(7,1)
 val safety = reads>=0 and reads<=MAX_READS and plans<=1
}
