module log_pattern_time_ownership {
 const INITIAL:int
 var phase:int
 var sourceOperations:int
 var protocols:int
 var writes:int
 var captureGets:int
 var coherent:bool
 action init=all{phase'=INITIAL,sourceOperations'=0,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action current_run=all{phase==0,phase'=1,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action failure_stage=all{phase==1,phase'=2,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action failure_delivered=all{phase==2,phase'=3,sourceOperations'=3,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action held_run=all{phase==3,phase'=4,sourceOperations'=5,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action held_return=all{phase==4,phase'=5,sourceOperations'=5,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action recovered_run=all{phase==5,phase'=6,sourceOperations'=7,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action stale_side=all{phase==6,phase'=7,sourceOperations'=7,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 val safety=coherent and sourceOperations>=0 and protocols>=0 and writes>=0 and captureGets>=0
}
