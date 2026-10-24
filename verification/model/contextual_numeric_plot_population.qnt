module contextual_numeric_plot_population {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var streams:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_stage=all{phase==0,phase'=1,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_run=all{phase==1,phase'=2,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_record_stage=all{phase==2,phase'=3,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_record_run=all{phase==3,phase'=4,reads'=3,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_record_back=all{phase==4,phase'=5,reads'=3,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_field_stage=all{phase==5,phase'=6,reads'=3,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_field_run=all{phase==6,phase'=7,reads'=4,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action population_field_back=all{phase==7,phase'=8,reads'=4,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 val safety=coherent and reads>=0 and plans==0 and protocols==0 and writes==0 and streams==0
}
