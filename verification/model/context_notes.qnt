module context_notes {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action observe = all {phase==0,phase'=1,coherent'=true}
 action draft = all {phase==1,phase'=2,coherent'=true}
 action revise = all {phase==2,phase'=3,coherent'=true}
 action branch = all {phase==3,phase'=4,coherent'=true}
 action reopen = all {phase==4,phase'=5,coherent'=true}
 val safety = coherent
}
