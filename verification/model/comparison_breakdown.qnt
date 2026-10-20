module comparison_breakdown {
 const INITIAL: int
 var reads: int
 var coherent: bool
 action init = all {reads'=INITIAL,coherent'=true}
 action preview = all {reads'=reads,coherent'=true}
 action capture_pair = all {reads'=reads+2,coherent'=true}
 action focus = all {reads'=reads,coherent'=true}
 action save = all {reads'=reads,coherent'=true}
 action reopen = all {reads'=reads,coherent'=true}
 val safety = coherent
}
