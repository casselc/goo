module population_drilldown {
 const INITIAL: int
 var depth: int
 var reads: int
 var coherent: bool
 action init = all {depth'=0, reads'=INITIAL, coherent'=true}
 action read = all {depth'=depth, reads'=reads+1, coherent'=true}
 action stage = all {depth<3, depth'=depth+1, reads'=reads, coherent'=true}
 action back = all {depth>0, depth'=depth-1, reads'=reads, coherent'=true}
 action history = all {depth'=depth, reads'=reads, coherent'=true}
 val safety = coherent and depth>=0 and depth<=3 and reads>=0
}
