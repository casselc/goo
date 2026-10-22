module stream_entry {
 const INITIAL: int
 var reads: int
 var paused: bool
 var gap: bool
 var coherent: bool
 action init = all {reads'=INITIAL,paused'=false,gap'=false,coherent'=true}
 action connect = all {reads'=reads+1,paused'=false,gap'=false,coherent'=true}
 action pause = all {reads'=reads,paused'=true,gap'=gap,coherent'=true}
 action hold = all {reads'=reads,paused'=paused,gap'=gap,coherent'=true}
 action stored = all {reads'=reads,paused'=true,gap'=false,coherent'=true}
 action resume = all {paused,reads'=reads+1,paused'=false,gap'=true,coherent'=true}
 val safety = coherent and reads>=0
}
