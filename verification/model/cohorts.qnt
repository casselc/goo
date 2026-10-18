module cohorts {
 const INITIAL: int
 var phase: int
 var scope: bool
 var honest: bool
 action init = all {phase'=INITIAL,scope'=true,honest'=true}
 action pin = all {phase==0,phase'=1,scope'=true,honest'=true}
 action selection = all {phase==1,phase'=2,scope'=true,honest'=true}
 action empty = all {phase==2,phase'=3,scope'=true,honest'=true}
 action changed = all {phase==3,phase'=4,scope'=true,honest'=true}
 action scopeChange = all {phase==4,phase'=5,scope'=true,honest'=true}
 val safety = scope and honest
}
