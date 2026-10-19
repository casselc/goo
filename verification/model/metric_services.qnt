module metric_services {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action discover = all {phase==0,phase'=1,coherent'=true}
 action browse = all {phase==1,phase'=2,coherent'=true}
 action preview = all {phase==2,phase'=3,coherent'=true}
 action execute = all {phase==3,phase'=4,coherent'=true}
 action returnLocal = all {phase==4,phase'=5,coherent'=true}
 val safety = coherent
}
