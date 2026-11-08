module captured_trace_search_383 {
 const INITIAL: int
 var reads: int
 var captures: int
 action init = all { reads'=INITIAL, captures'=0 }
 action read = all { reads'=reads+1, captures'=captures }
 action capture = all { reads'=reads, captures'=captures+1 }
 action history = all { reads'=reads, captures'=captures }
 val safety = reads>=0 and captures>=0 and captures<=1
}
