module native_navigation {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action inventory = all {phase==0,phase'=1,coherent'=true}
 action page = all {phase==1,phase'=2,coherent'=true}
 action review = all {phase==2,phase'=3,coherent'=true}
 action empty = all {phase==3,phase'=4,coherent'=true}
 action find = all {phase==4,phase'=5,coherent'=true}
 action clear = all {phase==5,phase'=6,coherent'=true}
 action restore = all {phase==6,phase'=7,coherent'=true}
 action focus = all {phase==7,phase'=8,coherent'=true}
 val safety = coherent
}
