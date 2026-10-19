module native_windows {
 const INITIAL: int
 var reads: int
 var allowed: bool
 var coherent: bool
 action init = all {reads'=INITIAL,allowed'=false,coherent'=true}
 action compare = all {reads'=reads,allowed'=true,coherent'=true}
 action refuse = all {reads'=reads,allowed'=false,coherent'=true}
 action reopen = all {reads'=reads,allowed'=true,coherent'=true}
 val safety = coherent and reads==INITIAL
}
