module trace_comparison {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_reference_search=all{phase==0,phase'=1,source'=1,coherent'=true}
 action observe_reference_expand=all{phase==1,phase'=2,source'=3,coherent'=true}
 action observe_pin=all{phase==2,phase'=3,source'=3,coherent'=true}
 action observe_current_search=all{phase==3,phase'=4,source'=4,coherent'=true}
 action observe_current_expand=all{phase==4,phase'=5,source'=6,coherent'=true}
 action observe_compare=all{phase==5,phase'=6,source'=6,coherent'=true}
 action observe_pair_attributes=all{phase==6,phase'=7,source'=6,coherent'=true}
 action observe_capture_reopen=all{phase==7,phase'=8,source'=6,coherent'=true}
 val safety=coherent and source>=0
}
