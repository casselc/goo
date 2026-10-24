module log_pattern_time_empty {
 const INITIAL:int
 var phase:int
 var sourceOperations:int
 var protocols:int
 var writes:int
 var captureGets:int
 var coherent:bool
 action init=all{phase'=INITIAL,sourceOperations'=0,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action current_run=all{phase==0,phase'=1,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action empty_stage=all{phase==1,phase'=2,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action empty_run=all{phase==2,phase'=3,sourceOperations'=3,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 val safety=coherent and sourceOperations>=0 and protocols>=0 and writes>=0 and captureGets>=0
}
