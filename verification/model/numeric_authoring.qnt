module numeric_authoring {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action stage=all{phase==0,phase'=1,reads'=0,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action author=all{phase==1,phase'=2,reads'=0,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_gte=all{phase==2,phase'=3,reads'=1,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action invalid_and_repair=all{phase==3,phase'=4,reads'=1,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_negative=all{phase==4,phase'=5,reads'=2,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_resource=all{phase==5,phase'=6,reads'=3,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_valid=all{phase==6,phase'=7,reads'=4,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_missing=all{phase==7,phase'=8,reads'=5,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_empty=all{phase==8,phase'=9,reads'=6,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action filter_invalid=all{phase==9,phase'=10,reads'=7,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action parent_stage=all{phase==10,phase'=11,reads'=7,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action parent_run=all{phase==11,phase'=12,reads'=8,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action brush=all{phase==12,phase'=13,reads'=8,plans'=0,protocols'=0,writes'=0,coherent'=true}
 action selected_run=all{phase==13,phase'=14,reads'=9,plans'=1,protocols'=0,writes'=0,coherent'=true}
 action back_and_brush=all{phase==14,phase'=15,reads'=9,plans'=1,protocols'=0,writes'=0,coherent'=true}
 action comparison_stage=all{phase==15,phase'=16,reads'=9,plans'=2,protocols'=0,writes'=0,coherent'=true}
 action comparison_run=all{phase==16,phase'=17,reads'=11,plans'=2,protocols'=1,writes'=0,coherent'=true}
 action capture=all{phase==17,phase'=18,reads'=13,plans'=2,protocols'=1,writes'=0,coherent'=true}
 action reopen=all{phase==18,phase'=19,reads'=13,plans'=2,protocols'=1,writes'=0,coherent'=true}
 action package_save=all{phase==19,phase'=20,reads'=13,plans'=2,protocols'=1,writes'=1,coherent'=true}
 action package_use=all{phase==20,phase'=21,reads'=14,plans'=2,protocols'=1,writes'=1,coherent'=true}
 val safety=coherent and reads>=0 and plans>=0 and protocols>=0 and writes>=0
}
