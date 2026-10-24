module trace_comparison_ownership {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_pair=all{phase==0,phase'=1,source'=6,coherent'=true}
 action observe_closed_pair=all{phase==1,phase'=2,source'=6,coherent'=true}
 action observe_hidden_actions=all{phase==2,phase'=3,source'=6,coherent'=true}
 action observe_return_stage=all{phase==3,phase'=4,source'=6,coherent'=true}
 action observe_returned_compare=all{phase==4,phase'=5,source'=8,coherent'=true}
 action observe_pending_actions=all{phase==5,phase'=6,source'=12,coherent'=true}
 action observe_stale_response=all{phase==6,phase'=7,source'=12,coherent'=true}
 action observe_fresh_accepted=all{phase==7,phase'=8,source'=15,coherent'=true}
 val safety=coherent and source>=0
}
