// Source counts belong to the applied scope. Checkbox/search/clear/stage are
// local; only Load and Apply acquire independent observations.
module facets {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var counted: int
 var applied: int
 var selected: int
 var pending: bool
 var current: bool
 action init = all {phase'=INITIAL,reads'=0,sql'=0,counted'=0,applied'=0,selected'=0,pending'=false,current'=true}
 action local = all {phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=selected,pending'=pending,current'=current}
 action inspect = all {phase==0,current,phase'=1,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=false,current'=current}
 action facet = all {phase==1,current,phase'=2,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=false,current'=current}
 action load = all {phase==2,current,not(pending),phase'=3,reads'=reads+1,sql'=sql+2,counted'=counted+1,applied'=applied,selected'=0,pending'=false,current'=current}
 action select_two = all {phase==3,current,not(pending),phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=2,pending'=pending,current'=current}
 action clear = all {phase==3,current,phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=pending,current'=current}
 action stage = all {phase==3,current,selected>0,not(pending),phase'=4,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=false,current'=false}
 action apply = all {phase==4,not(pending),phase'=5,reads'=reads+1,sql'=sql+1,counted'=counted,applied'=applied+1,selected'=0,pending'=false,current'=true}
 action begin_load = all {phase==2,current,not(pending),phase'=phase,reads'=reads+1,sql'=sql+2,counted'=counted+1,applied'=applied,selected'=0,pending'=true,current'=current}
 action finish_load = all {phase==2,current,pending,phase'=3,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=false,current'=current}
 action retire = all {phase'=6,reads'=reads,sql'=sql,counted'=counted,applied'=applied,selected'=0,pending'=false,current'=false}
 val safety = and {phase>=0,phase<=6,reads==counted+applied,sql==counted*2+applied,counted>=0,applied>=0,selected>=0,selected<=1000}
}
