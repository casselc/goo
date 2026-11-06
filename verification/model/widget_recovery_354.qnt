module widget_recovery_354 {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action start = all {phase==0,phase'=1,coherent'=true}
 action source_failed = all {phase==1,phase'=2,coherent'=true}
 action retry = all {phase==2,phase'=3,coherent'=true}
 action succeed = all {phase==3,phase'=4,coherent'=true}
 action retire = all {phase==3,phase'=5,coherent'=true}
 action late = all {phase==5,phase'=5,coherent'=true}
 action refused = all {phase>0,phase'=phase,coherent'=true}
 action finish = all {or {phase==4,phase==5},phase'=6,coherent'=true}
 val safety = and {phase>=0,phase<=6,coherent}
}
