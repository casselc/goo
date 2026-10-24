module log_patterns {
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
 action patterns_stage=all{phase==1,phase'=2,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action patterns_run=all{phase==2,phase'=3,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action sort_local=all{phase==3,phase'=4,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action raw_inspection=all{phase==4,phase'=5,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action pattern_trend_stage=all{phase==5,phase'=6,reads'=1,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action pattern_trend_run=all{phase==6,phase'=7,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action pattern_back=all{phase==7,phase'=8,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action comparison_stage=all{phase==8,phase'=9,reads'=2,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action comparison_run=all{phase==9,phase'=10,reads'=4,plans'=0,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action capture=all{phase==10,phase'=11,reads'=6,plans'=0,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action reopen=all{phase==11,phase'=12,reads'=6,plans'=0,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action reuse_stage=all{phase==12,phase'=13,reads'=6,plans'=0,protocols'=1,writes'=0,streams'=1,coherent'=true}
 action reuse_save=all{phase==13,phase'=14,reads'=6,plans'=0,protocols'=1,writes'=1,streams'=1,coherent'=true}
 action reuse_run=all{phase==14,phase'=15,reads'=7,plans'=0,protocols'=1,writes'=1,streams'=1,coherent'=true}
 action return_paused=all{phase==15,phase'=16,reads'=7,plans'=0,protocols'=1,writes'=1,streams'=1,coherent'=true}
 val safety=coherent and reads>=0 and plans==0 and protocols>=0 and writes>=0 and streams>=0 and streams<=1
}
