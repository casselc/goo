module trace_comparison_partial {
 const INITIAL:int
 var phase:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,source'=0,coherent'=true}
 action observe_problem_pin=all{phase==0,phase'=1,source'=3,coherent'=true}
 action observe_problem_compare=all{phase==1,phase'=2,source'=6,coherent'=true}
 action observe_partial_compare=all{phase==2,phase'=3,source'=9,coherent'=true}
 val safety=coherent and source>=0
}
