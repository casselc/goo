// Finite owner protocol: explicit acquisition versus local invocation/return.
// Source and SQL counters are separate; configuration preparation is source-free.
module actions {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var pairs: int
 var pending: bool
 var current: bool
 action init = all {phase'=INITIAL,reads'=0,sql'=0,pairs'=0,pending'=false,current'=true}
 action local = all {phase'=phase,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action stage_compare = all {phase==0,phase'=1,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action run_compare = all {phase==1,phase'=2,reads'=reads+2,sql'=sql+6,pairs'=pairs+1,pending'=pending,current'=current}
 action return_service = all {phase==2,phase'=0,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action stage_records = all {phase==0,current,phase'=3,reads'=reads,sql'=sql,pairs'=pairs,pending'=false,current'=current}
 action read_records = all {phase==3,phase'=4,reads'=reads+1,sql'=sql+4,pairs'=pairs,pending'=pending,current'=current}
 action inspect_record = all {phase==4,phase'=5,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action read_trace = all {phase==5,phase'=6,reads'=reads+2,sql'=sql+7,pairs'=pairs,pending'=pending,current'=current}
 action return_trace = all {phase==6,phase'=5,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action row_trace = all {or {phase==4,phase==5},phase'=7,reads'=reads+2,sql'=sql+7,pairs'=pairs,pending'=pending,current'=current}
 action return_row = all {phase==7,phase'=4,reads'=reads,sql'=sql,pairs'=pairs,pending'=pending,current'=current}
 action begin_customize = all {phase==0,current,not(pending),phase'=phase,reads'=reads,sql'=sql,pairs'=pairs,pending'=true,current'=current}
 action complete_customize = all {phase==0,current,pending,phase'=3,reads'=reads,sql'=sql,pairs'=pairs,pending'=false,current'=current}
 action retire_customize = all {phase==0,pending,phase'=phase,reads'=reads,sql'=sql,pairs'=pairs,pending'=false,current'=false}
 val safety = and {phase>=0,phase<=7,reads>=0,sql>=0,pairs>=0}
}
