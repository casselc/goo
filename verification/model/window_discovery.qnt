module window_discovery {
 const INITIAL: int
 var reads: int
 var phase: int
 var coherent: bool
 action init = all {reads'=INITIAL,phase'=0,coherent'=true}
 action discover = all {reads'=reads,phase'=1,coherent'=true}
 action choose_support = all {(phase==1 or phase==2),reads'=reads,phase'=2,coherent'=true}
 action capture = all {phase==2,reads'=reads+2,phase'=3,coherent'=true}
 action reopen = all {(phase==3 or phase==4),reads'=reads,phase'=4,coherent'=true}
 val safety = coherent and reads>=INITIAL
}
