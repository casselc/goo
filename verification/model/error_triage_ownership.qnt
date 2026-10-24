module error_triage_ownership {
 const INITIAL:int
 var phase:int
 var sources:int
 var streams:int
 var writes:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,streams'=0,writes'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,sources'=0,streams'=0,writes'=0,coherent'=true}
 action open=all{phase==1,phase'=2,sources'=0,streams'=1,writes'=0,coherent'=true}
 action dirty_open=all{phase==2,phase'=3,sources'=0,streams'=1,writes'=0,coherent'=true}
 action applied_old_open=all{phase==3,phase'=4,sources'=0,streams'=2,writes'=0,coherent'=true}
 action reopen=all{phase==4,phase'=5,sources'=0,streams'=2,writes'=0,coherent'=true}
 action hidden_open=all{phase==5,phase'=6,sources'=0,streams'=2,writes'=0,coherent'=true}
 action return=all{phase==6,phase'=7,sources'=0,streams'=2,writes'=0,coherent'=true}
 action resumed_old_open=all{phase==7,phase'=8,sources'=0,streams'=3,writes'=0,coherent'=true}
 val safety=coherent and sources==0 and writes==0
}
