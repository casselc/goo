module fresh_comparison {
 const INITIAL: int
 var reads: int
 var coherent: bool
 action init = all {reads'=INITIAL,coherent'=true}
 action preview = all {reads'=reads,coherent'=true}
 action choose_prior = all {reads'=reads,coherent'=true}
 action capture = all {reads'=reads+1,coherent'=true}
 action retry = all {reads'=reads,coherent'=true}
 action save = all {reads'=reads,coherent'=true}
 action reopen = all {reads'=reads,coherent'=true}
 val safety = coherent
}
