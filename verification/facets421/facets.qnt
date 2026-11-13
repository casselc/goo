// Contextual selection and native disclosure are local. Only explicit counted
// loads acquire an independent full-scope count and complete group read.
module facets {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var counted: int
 var pending: bool
 var current: bool
 action init = all {phase'=INITIAL,reads'=0,sql'=0,counted'=0,pending'=false,current'=true}
 action local = all {phase'=phase,reads'=reads,sql'=sql,counted'=counted,pending'=pending,current'=current}
 action inspect = all {phase==0,current,phase'=1,reads'=reads,sql'=sql,counted'=counted,pending'=false,current'=current}
 action facet = all {or {phase==1,phase==2,phase==3},current,not(pending),phase'=2,reads'=reads,sql'=sql,counted'=counted,pending'=false,current'=current}
 action load = all {phase==2,current,not(pending),phase'=3,reads'=reads+1,sql'=sql+2,counted'=counted+1,pending'=false,current'=current}
 action stage = all {phase==3,current,phase'=4,reads'=reads,sql'=sql,counted'=counted,pending'=false,current'=false}
 action begin_load = all {phase==2,current,not(pending),phase'=phase,reads'=reads+1,sql'=sql+2,counted'=counted+1,pending'=true,current'=current}
 action finish_load = all {phase==2,current,pending,phase'=3,reads'=reads,sql'=sql,counted'=counted,pending'=false,current'=current}
 action retire = all {phase'=4,reads'=reads,sql'=sql,counted'=counted,pending'=false,current'=false}
 val safety = and {phase>=0,phase<=4,reads>=0,sql==reads*2,counted==reads}
}
