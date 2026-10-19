module question_session {
 const INITIAL: int
 var generation: int
 var phase: int
 var published: int
 var coherent: bool
 action init = all {generation'=INITIAL,phase'=0,published'=-1,coherent'=true}
 action edit = all {generation'=generation+1,phase'=1,published'=published,coherent'=true}
 action start = all {generation'=generation+1,phase'=2,published'=published,coherent'=true}
 action publish = all {phase==2,generation'=generation,phase'=3,published'=generation,coherent'=true}
 action reject = all {generation'=generation,phase'=phase,published'=published,coherent'=true}
 action cancel = all {generation'=generation+1,phase'=4,published'=published,coherent'=true}
 action failedRead = all {phase==2,generation'=generation,phase'=5,published'=published,coherent'=true}
 val safety = all {coherent,published<=generation,phase==3 implies published==generation}
}
