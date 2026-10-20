module service_namespace {
 const INITIAL: int
 var coherent: bool
 var phase: int
 action init = all {coherent'=true,phase'=INITIAL}
 action discover = all {phase==0,phase'=1,coherent'=true}
 action open = all {phase==1,phase'=2,coherent'=true}
 action contract = all {phase==2,phase'=3,coherent'=true}
 val safety = coherent
}
