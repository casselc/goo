module tail {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action connect = all {phase==0,phase'=1,coherent'=true}
 action receive_batch = all {phase==1,phase'=2,coherent'=true}
 action retain = all {phase==2,phase'=3,coherent'=true}
 action pause = all {phase==3,phase'=4,coherent'=true}
 action resume = all {phase==4,phase'=5,coherent'=true}
 action scope = all {phase==5,phase'=6,coherent'=true}
 val safety = coherent
}
