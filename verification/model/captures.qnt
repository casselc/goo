module captures {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action save = all {phase==0,phase'=1,coherent'=true}
 action page = all {phase==1,phase'=2,coherent'=true}
 action refresh = all {phase==2,phase'=3,coherent'=true}
 action reopen = all {phase==3,phase'=4,coherent'=true}
 val safety = coherent
}
