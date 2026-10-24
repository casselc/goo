module activity_explanation {
 const INITIAL:int
 var phase:int
 var sources:int
 var promotions:int
 var plans:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,promotions'=0,plans'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,sources'=0,promotions'=0,plans'=0,coherent'=true}
 action open=all{phase==1,phase'=2,sources'=0,promotions'=0,plans'=0,coherent'=true}
 action edit_fields=all{phase==2,phase'=3,sources'=0,promotions'=0,plans'=0,coherent'=true}
 action previous_run=all{phase==3,phase'=4,sources'=1,promotions'=0,plans'=0,coherent'=true}
 action brush=all{phase==4,phase'=5,sources'=1,promotions'=0,plans'=0,coherent'=true}
 action brush_run=all{phase==5,phase'=6,sources'=2,promotions'=0,plans'=0,coherent'=true}
 action saved_scope_stage=all{phase==6,phase'=7,sources'=2,promotions'=0,plans'=0,coherent'=true}
 action saved_scope_load=all{phase==7,phase'=8,sources'=2,promotions'=0,plans'=0,coherent'=true}
 action saved_scope_run=all{phase==8,phase'=9,sources'=3,promotions'=0,plans'=0,coherent'=true}
 action late_arrival=all{phase==9,phase'=10,sources'=3,promotions'=0,plans'=0,coherent'=true}
 action save_lost_ack=all{phase==10,phase'=11,sources'=3,promotions'=1,plans'=0,coherent'=true}
 action save_retry=all{phase==11,phase'=12,sources'=3,promotions'=2,plans'=0,coherent'=true}
 action filter_stage=all{phase==12,phase'=13,sources'=3,promotions'=2,plans'=1,coherent'=true}
 action filter_open=all{phase==13,phase'=14,sources'=3,promotions'=2,plans'=1,coherent'=true}
 action filter_run=all{phase==14,phase'=15,sources'=4,promotions'=2,plans'=1,coherent'=true}
 action filter_return=all{phase==15,phase'=16,sources'=4,promotions'=2,plans'=1,coherent'=true}
 action reopen=all{phase==16,phase'=17,sources'=4,promotions'=2,plans'=1,coherent'=true}
 action reuse_stage=all{phase==17,phase'=18,sources'=4,promotions'=2,plans'=1,coherent'=true}
 action reuse_run=all{phase==18,phase'=19,sources'=5,promotions'=2,plans'=1,coherent'=true}
 val safety=coherent and sources>=0 and promotions>=0 and plans>=0
}
