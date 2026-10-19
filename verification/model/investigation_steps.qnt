module investigation_steps {
 const INITIAL: int
 var phase: int
 var reads: int
 var coherent: bool
 action init = all {phase'=0,reads'=INITIAL,coherent'=true}
 action save = all {phase==0,phase'=1,reads'=reads,coherent'=true}
 action reopen = all {phase==1,phase'=2,reads'=reads,coherent'=true}
 action executeRead = all {phase==2,phase'=3,reads'=reads+1,coherent'=true}
 val safety = coherent and (phase<3 implies reads==0) and (phase==3 implies reads==1)
}
