module question_links {
 const INITIAL: int
 var reads: int
 var coherent: bool
 action init = all { reads'=INITIAL, coherent'=true }
 action restore = all { reads'=reads, coherent'=true }
 action read = all { reads'=reads+1, coherent'=true }
 action history = all { reads'=reads, coherent'=true }
 val safety = coherent and reads>=0
}
