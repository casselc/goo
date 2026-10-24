module contextual_numeric_plot {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var streams:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action activity_start=all{phase==0,phase'=1,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action activity_cancel=all{phase==1,phase'=2,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action activity_stage=all{phase==2,phase'=3,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action activity_run=all{phase==3,phase'=4,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action fixed_same_field_cancel=all{phase==4,phase'=5,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action resource_stage=all{phase==5,phase'=6,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action resource_run=all{phase==6,phase'=7,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action discovery_stage=all{phase==7,phase'=8,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action discovery_run=all{phase==8,phase'=9,reads'=3,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action brush=all{phase==9,phase'=10,reads'=3,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action selected_run=all{phase==10,phase'=11,reads'=4,plans'=1,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action back_brush=all{phase==11,phase'=12,reads'=4,plans'=1,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action comparison_stage=all{phase==12,phase'=13,reads'=4,plans'=2,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action comparison_run=all{phase==13,phase'=14,reads'=6,plans'=2,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action capture=all{phase==14,phase'=15,reads'=8,plans'=2,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action saved_stage=all{phase==15,phase'=16,reads'=8,plans'=2,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action saved_run=all{phase==16,phase'=17,reads'=9,plans'=2,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action return_paused=all{phase==17,phase'=18,reads'=9,plans'=2,protocols'=1,writes'=0,streams'=1,coherent'=true}
 val safety=coherent and reads>=0 and plans>=0 and protocols>=0 and writes==0 and streams>=0
}
