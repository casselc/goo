module discovery {
 const INITIAL: int
 var phase: int
 var bounded: bool
 var reusable: bool
 action init = all {phase'=INITIAL,bounded'=true,reusable'=true}
 action surround = all {phase==0,phase'=1,bounded'=true,reusable'=true}
 action trace = all {phase==1,phase'=2,bounded'=true,reusable'=true}
 action save = all {phase==2,phase'=3,bounded'=true,reusable'=true}
 val safety = bounded and reusable
}
