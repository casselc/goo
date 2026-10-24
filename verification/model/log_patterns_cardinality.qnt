module log_patterns_cardinality {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var streams:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action high_cardinality_stage=all{phase==0,phase'=1,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action high_cardinality_refused=all{phase==1,phase'=2,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 val safety=coherent and reads>=0 and plans==0 and protocols==0 and writes==0 and streams==0
}
