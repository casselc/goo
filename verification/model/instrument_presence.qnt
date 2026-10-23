module instrument_presence {
 const INITIAL: int
 var phase: int
 var coherent: bool
 var reads: int
 var effects: int
 action init = all {phase'=INITIAL,coherent'=true,reads'=0,effects'=0}
 action discover = all {phase==0,phase'=1,coherent'=true,reads'=1,effects'=0}
 action inspect = all {phase==1,phase'=2,coherent'=true,reads'=reads,effects'=effects}
 val safety = coherent and effects==0 and reads<=1
}
