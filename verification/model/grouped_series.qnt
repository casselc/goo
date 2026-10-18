module grouped_series {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action plan = all {phase==0,phase'=1,coherent'=true}
 action execute = all {phase==1,phase'=2,coherent'=true}
 action inspect = all {phase==2,phase'=3,coherent'=true}
 action capture = all {phase==3,phase'=4,coherent'=true}
 action reuse = all {phase==4,phase'=5,coherent'=true}
 action reopen = all {phase==5,phase'=6,coherent'=true}
 val safety = coherent
}
