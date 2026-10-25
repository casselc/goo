module trace_search {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_parent_run=all{phase==0,phase'=1,source'=1,coherent'=true}
 action observe_stage=all{phase==1,phase'=2,source'=1,coherent'=true}
 action observe_search=all{phase==2,phase'=3,source'=2,coherent'=true}
 action observe_witness_details=all{phase==3,phase'=4,source'=2,coherent'=true}
 action observe_open_cross=all{phase==4,phase'=5,source'=4,coherent'=true}
 action observe_pin_cross=all{phase==5,phase'=6,source'=4,coherent'=true}
 action observe_open_same=all{phase==6,phase'=7,source'=6,coherent'=true}
 action observe_compare=all{phase==7,phase'=8,source'=6,coherent'=true}
 action observe_open_repeated=all{phase==8,phase'=9,source'=8,coherent'=true}
 action observe_next_trace_page=all{phase==9,phase'=10,source'=10,coherent'=true}
 val safety=coherent and source>=0
}
