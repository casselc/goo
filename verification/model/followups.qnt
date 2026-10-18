module followups {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action preview = all {phase==0,phase'=1,coherent'=true}
 action duration = all {phase==1,phase'=2,coherent'=true}
 action trace = all {phase==2,phase'=3,coherent'=true}
 action logs = all {phase==3,phase'=4,coherent'=true}
 action resources = all {phase==4,phase'=5,coherent'=true}
 action gap = all {phase==5,phase'=6,coherent'=true}
 action reopen = all {phase==6,phase'=7,coherent'=true}
 val safety = coherent
}
