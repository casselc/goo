// A counted read supplies choices, never comparison operands.
module comparison {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var counted: int
 var compares: int
 var runs: int
 action init = all {phase'=INITIAL,reads'=0,sql'=0,counted'=0,compares'=0,runs'=0}
 action local = all {phase'=phase,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action inspect = all {phase==0,phase'=1,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action facet = all {phase==1,phase'=2,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action load = all {phase==2,phase'=3,reads'=reads+1,sql'=sql+2,counted'=counted+1,compares'=compares,runs'=runs}
 action choose_values = all {or {phase==3,phase==4},phase'=4,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action stage = all {phase==4,phase'=5,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action execute_comparison = all {phase==5,phase'=6,reads'=reads+2,sql'=sql+6,counted'=counted,compares'=compares+1,runs'=runs+1}
 action back = all {phase==6,phase'=4,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 action retire = all {phase'=7,reads'=reads,sql'=sql,counted'=counted,compares'=compares,runs'=runs}
 val safety = and {phase>=0,phase<=7,reads==counted+2*runs,sql==2*counted+6*runs,compares==runs,counted>=0,runs>=0}
}
