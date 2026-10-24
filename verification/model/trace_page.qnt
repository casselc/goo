module trace_page {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_run=all{phase==0,phase'=1,source'=1,coherent'=true}
 action observe_expand=all{phase==1,phase'=2,source'=3,coherent'=true}
 action observe_group_highlight=all{phase==2,phase'=3,source'=3,coherent'=true}
 action observe_group_clear=all{phase==3,phase'=4,source'=3,coherent'=true}
 action observe_row_select=all{phase==4,phase'=5,source'=4,coherent'=true}
 action observe_capture=all{phase==5,phase'=6,source'=5,coherent'=true}
 action observe_reopen=all{phase==6,phase'=7,source'=5,coherent'=true}
 val safety=coherent and source>=0
}
