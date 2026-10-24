module activity_explanation_ownership {
 const INITIAL:int
 var phase:int
 var sources:int
 var plans:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,plans'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,sources'=0,plans'=0,coherent'=true}
 action open=all{phase==1,phase'=2,sources'=0,plans'=0,coherent'=true}
 action resume_held=all{phase==2,phase'=3,sources'=1,plans'=0,coherent'=true}
 action resume_released=all{phase==3,phase'=4,sources'=1,plans'=0,coherent'=true}
 action resume_reopen=all{phase==4,phase'=5,sources'=1,plans'=0,coherent'=true}
 action navigation_held=all{phase==5,phase'=6,sources'=2,plans'=0,coherent'=true}
 action navigation_released=all{phase==6,phase'=7,sources'=2,plans'=0,coherent'=true}
 action navigation_reopen=all{phase==7,phase'=8,sources'=2,plans'=0,coherent'=true}
 action scope_held=all{phase==8,phase'=9,sources'=3,plans'=0,coherent'=true}
 action scope_released=all{phase==9,phase'=10,sources'=3,plans'=0,coherent'=true}
 action scope_reopen=all{phase==10,phase'=11,sources'=3,plans'=0,coherent'=true}
 action load_held=all{phase==11,phase'=12,sources'=3,plans'=0,coherent'=true}
 action load_return=all{phase==12,phase'=13,sources'=3,plans'=0,coherent'=true}
 action load_fresh=all{phase==13,phase'=14,sources'=3,plans'=0,coherent'=true}
 action fresh_run=all{phase==14,phase'=15,sources'=4,plans'=0,coherent'=true}
 action filter_held=all{phase==15,phase'=16,sources'=4,plans'=1,coherent'=true}
 action filter_released=all{phase==16,phase'=17,sources'=4,plans'=1,coherent'=true}
 val safety=coherent and sources>=0 and plans>=0
}
