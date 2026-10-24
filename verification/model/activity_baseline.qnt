module activity_baseline {
 const INITIAL: int
 var reads: int
 var chosen: bool
 var coherent: bool
 action init = all {reads'=INITIAL,chosen'=false,coherent'=true}
 action read = all {reads'=reads+1,chosen'=chosen,coherent'=true}
 action choose = all {reads'=reads,chosen'=true,coherent'=true}
 action clear = all {reads'=reads,chosen'=false,coherent'=true}
 val safety = coherent and reads>=0
}
