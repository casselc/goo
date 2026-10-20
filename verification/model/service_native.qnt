module service_native {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action automatic = all {phase==0,phase'=1,coherent'=true}
 action values = all {phase==1,phase'=2,coherent'=true}
 action browse = all {phase==2,phase'=3,coherent'=true}
 action preview = all {phase==3,phase'=4,coherent'=true}
 action evaluate = all {phase==4,phase'=5,coherent'=true}
 action capture = all {phase==5,phase'=6,coherent'=true}
 action reopen = all {phase==6,phase'=7,coherent'=true}
 action gauge = all {phase==7,phase'=8,coherent'=true}
 val safety = coherent
}
