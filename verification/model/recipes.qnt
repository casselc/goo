module recipes {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action preview = all {phase==0,phase'=1,coherent'=true}
 action execute = all {phase==1,phase'=2,coherent'=true}
 action capture = all {phase==2,phase'=3,coherent'=true}
 action document = all {phase==3,phase'=4,coherent'=true}
 action reuse = all {phase==4,phase'=5,coherent'=true}
 val safety = coherent
}
