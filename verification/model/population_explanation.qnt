module population_explanation {
 const INITIAL: int
 var phase:int
 var sources:int
 var promotions:int
 var plans:int
 var filters:int
 var coherent:bool
 action init=all{phase'=INITIAL,sources'=0,promotions'=0,plans'=0,filters'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action parent=all{phase==1,phase'=2,sources'=sources+1,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action select_range=all{phase==2,phase'=3,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action edit_fields=all{phase==3,phase'=4,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action explain=all{phase==4,phase'=5,sources'=sources+1,promotions'=promotions,plans'=plans+1,filters'=filters,coherent'=true}
 action late_arrival=all{phase==5,phase'=6,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action save_lost_ack=all{phase==6,phase'=7,sources'=sources,promotions'=promotions+1,plans'=plans,filters'=filters,coherent'=true}
 action save_retry=all{phase==7,phase'=8,sources'=sources,promotions'=promotions+1,plans'=plans,filters'=filters,coherent'=true}
 action reopen=all{phase==8,phase'=9,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action reuse_stage=all{phase==9,phase'=10,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action reuse_run=all{phase==10,phase'=11,sources'=sources+1,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action filter_stage=all{phase==11,filters<3,phase'=12,sources'=sources,promotions'=promotions,plans'=plans+1,filters'=filters,coherent'=true}
 action filter_open=all{phase==12,phase'=13,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action filter_run=all{phase==13,phase'=14,sources'=sources+1,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action filter_return=all{phase==14,phase'=11,sources'=sources,promotions'=promotions,plans'=plans,filters'=filters+1,coherent'=true}
 action high_parent=all{phase==11,filters==3,phase'=15,sources'=sources+1,promotions'=promotions,plans'=plans,filters'=filters,coherent'=true}
 action high_explain=all{phase==15,phase'=16,sources'=sources+1,promotions'=promotions,plans'=plans+1,filters'=filters,coherent'=true}
 val safety=coherent and sources>=0 and promotions>=0 and plans>=0 and filters<=3
}
