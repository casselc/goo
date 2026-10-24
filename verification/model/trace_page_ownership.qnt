module trace_page_ownership {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_closed_controls=all{phase==0,phase'=1,source'=3,coherent'=true}
 action observe_hidden_controls=all{phase==1,phase'=2,source'=3,coherent'=true}
 action observe_return_current=all{phase==2,phase'=3,source'=3,coherent'=true}
 val safety=coherent and source>=0
}
