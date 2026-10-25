module trace_search_ownership {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_initial=all{phase==0,phase'=1,source'=2,coherent'=true}
 action observe_closed_editor=all{phase==1,phase'=2,source'=2,coherent'=true}
 action observe_closed_editor_rerun=all{phase==2,phase'=3,source'=3,coherent'=true}
 action observe_closed=all{phase==3,phase'=4,source'=3,coherent'=true}
 action observe_hidden=all{phase==4,phase'=5,source'=3,coherent'=true}
 action observe_held=all{phase==5,phase'=6,source'=4,coherent'=true}
 action observe_superseded=all{phase==6,phase'=7,source'=4,coherent'=true}
 action observe_fresh=all{phase==7,phase'=8,source'=5,coherent'=true}
 val safety=coherent and source>=0
}
