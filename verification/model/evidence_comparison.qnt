module evidence_comparison {
 const INITIAL: int
 var reads: int
 var compatible: bool
 var delta: bool
 var coherent: bool
 action init = all {reads'=INITIAL,compatible'=false,delta'=false,coherent'=true}
 action inspect = all {reads'=reads,compatible'=false,delta'=false,coherent'=true}
 action compare = all {reads'=reads,compatible'=true,delta'=true,coherent'=true}
 action refuse = all {reads'=reads,compatible'=false,delta'=false,coherent'=true}
 val safety = coherent and reads==0 and (delta implies compatible)
}
