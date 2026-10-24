module contextual_numeric_plot_ownership {
 const INITIAL:int
 var phase:int
 var reads:int
 var plans:int
 var protocols:int
 var writes:int
 var streams:int
 var coherent:bool
 action init=all{phase'=INITIAL,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=0,coherent'=true}
 action open=all{phase==0,phase'=1,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action navigation=all{phase==1,phase'=2,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action dirty_scope=all{phase==2,phase'=3,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=1,coherent'=true}
 action resume=all{phase==3,phase'=4,reads'=0,plans'=0,protocols'=0,writes'=0,streams'=3,coherent'=true}
 val safety=coherent and reads>=0 and plans>=0 and protocols>=0 and writes==0 and streams>=0
}
