module trace_page_graphs {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_missing_parent=all{phase==0,phase'=1,source'=3,coherent'=true}
 action observe_outside_parent=all{phase==1,phase'=2,source'=6,coherent'=true}
 action observe_grandchild_outside_child=all{phase==2,phase'=3,source'=9,coherent'=true}
 action observe_duplicate_identity=all{phase==3,phase'=4,source'=12,coherent'=true}
 action observe_parent_cycle=all{phase==4,phase'=5,source'=15,coherent'=true}
 action observe_bounded_page=all{phase==5,phase'=6,source'=20,coherent'=true}
 val safety=coherent and source>=0
}
