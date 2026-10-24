module error_triage {
 const INITIAL:int
 const TRACE_READS:int
 var phase:int
 var sources:int
 var streams:int
 var writes:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,streams'=0,writes'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,sources'=0,streams'=0,writes'=0,coherent'=true}
 action open=all{phase==1,phase'=2,sources'=0,streams'=1,writes'=0,coherent'=true}
 action held_inspect=all{phase==2,phase'=3,sources'=0,streams'=1,writes'=0,coherent'=true}
 action presets=all{phase==3,phase'=4,sources'=0,streams'=1,writes'=0,coherent'=true}
 action triage_stage=all{phase==4,phase'=5,sources'=0,streams'=1,writes'=0,coherent'=true}
 action triage_run=all{phase==5,phase'=6,sources'=1,streams'=1,writes'=0,coherent'=true}
 action fixed_inspect=all{phase==6,phase'=7,sources'=1,streams'=1,writes'=0,coherent'=true}
 action trace_inspect=all{phase==7 and TRACE_READS>0,phase'=8,sources'=1+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action capture=all{((phase==7 and TRACE_READS==0) or phase==8),phase'=9,sources'=2+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action reopen=all{phase==9,phase'=10,sources'=2+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action captured_question_stage=all{phase==10,phase'=11,sources'=2+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action captured_question_run=all{phase==11,phase'=12,sources'=3+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action group_stage=all{phase==12,phase'=13,sources'=3+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action group_run=all{phase==13,phase'=14,sources'=4+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action group_return=all{phase==14,phase'=15,sources'=4+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action parent_again=all{phase==15,phase'=16,sources'=5+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action service_stage=all{phase==16,phase'=17,sources'=5+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action service_run=all{phase==17,phase'=18,sources'=6+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action previous=all{phase==18,phase'=19,sources'=7+TRACE_READS,streams'=1,writes'=0,coherent'=true}
 action save_question=all{phase==19,phase'=20,sources'=7+TRACE_READS,streams'=1,writes'=1,coherent'=true}
 action notes=all{phase==20,phase'=21,sources'=7+TRACE_READS,streams'=1,writes'=2,coherent'=true}
 action capture_reload=all{phase==21,phase'=22,sources'=7+TRACE_READS,streams'=1,writes'=2,coherent'=true}
 val safety=coherent and sources>=0 and streams>=0 and writes>=0
}
