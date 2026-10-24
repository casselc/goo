module log_patterns_ownership {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var streams:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action original_run=all{phase==0,phase'=1,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action hidden_owner=all{phase==1,phase'=2,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action dirty_owner=all{phase==2,phase'=3,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 val safety=coherent and reads>=0 and plans==0 and protocols==0 and writes==0 and streams==0
}
