module metric_discovery {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL, coherent'=true}
 action discover = all {phase==0,phase'=1,coherent'=true}
 action build = all {phase==1,phase'=2,coherent'=true}
 action execute = all {phase==2,phase'=3,coherent'=true}
 val safety = coherent
}
