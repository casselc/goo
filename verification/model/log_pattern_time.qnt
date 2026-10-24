module log_pattern_time {
 const INITIAL:int
 var phase:int
 var sourceOperations:int
 var protocols:int
 var writes:int
 var captureGets:int
 var coherent:bool
 action init=all{phase'=INITIAL,sourceOperations'=0,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action current_stage=all{phase==0,phase'=1,sourceOperations'=0,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action current_run=all{phase==1,phase'=2,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action previous_stage=all{phase==2,phase'=3,sourceOperations'=1,protocols'=0,writes'=0,captureGets'=0,coherent'=true}
 action previous_run=all{phase==3,phase'=4,sourceOperations'=3,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action previous_local_metrics=all{phase==4,phase'=5,sourceOperations'=3,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action side_stage=all{phase==5,phase'=6,sourceOperations'=3,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action side_run=all{phase==6,phase'=7,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action side_back=all{phase==7,phase'=8,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action parent_return=all{phase==8,phase'=9,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action saved_scope_edit=all{phase==9,phase'=10,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=0,coherent'=true}
 action saved_scope_load=all{phase==10,phase'=11,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=1,coherent'=true}
 action saved_stage=all{phase==11,phase'=12,sourceOperations'=4,protocols'=1,writes'=0,captureGets'=1,coherent'=true}
 action saved_run=all{phase==12,phase'=13,sourceOperations'=6,protocols'=2,writes'=0,captureGets'=1,coherent'=true}
 action saved_local_metrics=all{phase==13,phase'=14,sourceOperations'=6,protocols'=2,writes'=0,captureGets'=1,coherent'=true}
 action view_save=all{phase==14,phase'=15,sourceOperations'=6,protocols'=2,writes'=1,captureGets'=1,coherent'=true}
 action capture_fresh=all{phase==15,phase'=16,sourceOperations'=7,protocols'=2,writes'=1,captureGets'=3,coherent'=true}
 action capture_reopen=all{phase==16,phase'=17,sourceOperations'=7,protocols'=2,writes'=1,captureGets'=6,coherent'=true}
 action view_reopen=all{phase==17,phase'=18,sourceOperations'=7,protocols'=2,writes'=1,captureGets'=6,coherent'=true}
 action view_run=all{phase==18,phase'=19,sourceOperations'=9,protocols'=3,writes'=1,captureGets'=6,coherent'=true}
 action original_unchanged=all{phase==19,phase'=20,sourceOperations'=9,protocols'=3,writes'=1,captureGets'=6,coherent'=true}
 val safety=coherent and sourceOperations>=0 and protocols>=0 and writes>=0 and captureGets>=0
}
