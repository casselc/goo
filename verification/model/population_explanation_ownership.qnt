module population_explanation_ownership {
 const INITIAL:int
 var phase:int
 var sources:int
 var plans:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,plans'=0,coherent'=true}
 action parent=all{phase==0,phase'=1,sources'=sources+1,plans'=plans,coherent'=true}
 action edit_fields=all{phase==1,phase'=2,sources'=sources,plans'=plans,coherent'=true}
 action run_held=all{phase==2,phase'=3,sources'=sources+1,plans'=plans+1,coherent'=true}
 action cancel=all{phase==3,phase'=4,sources'=sources,plans'=plans,coherent'=true}
 action return=all{phase==4,phase'=5,sources'=sources,plans'=plans,coherent'=true}
 action fresh_run=all{phase==5,phase'=6,sources'=sources+1,plans'=plans+1,coherent'=true}
 action filter_held=all{phase==6,phase'=7,sources'=sources,plans'=plans+1,coherent'=true}
 action filter_return=all{phase==7,phase'=8,sources'=sources,plans'=plans,coherent'=true}
 action filter_fresh=all{phase==8,phase'=9,sources'=sources,plans'=plans+1,coherent'=true}
 action stage_side=all{phase==9,phase'=10,sources'=sources,plans'=plans,coherent'=true}
 action run_side=all{phase==10,phase'=11,sources'=sources+1,plans'=plans,coherent'=true}
 val safety=coherent and sources>=0 and plans>=0
}
