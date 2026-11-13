// Per-record exception grouping owns only explicit reads. Activity preparation
// and independent actor campaigns are recorded separately by their own owners.
module exceptiongroups {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 action init = all {phase'=INITIAL,reads'=0,sql'=0}
 action scope = all {or {phase==0,phase==4},phase'=0,reads'=reads,sql'=sql}
 action read_groups = all {or {phase==0,phase==1},phase'=1,reads'=reads+1,sql'=sql+2}
 action failed_read = all {or {phase==0,phase==1},phase'=0,reads'=reads+1,sql'=sql+1}
 action refused_groups = all {or {phase==0,phase==1},phase'=0,reads'=reads+1,sql'=sql+2}
 action local_intent = all {phase>=0,phase<=5,phase'=phase,reads'=reads,sql'=sql}
 action stage_group = all {phase==1,phase'=2,reads'=reads,sql'=sql}
 action run_group = all {phase==2,phase'=3,reads'=reads+1,sql'=sql+3}
 action held_back = all {phase==3,phase'=1,reads'=reads,sql'=sql}
 action pending_read = all {or {phase==0,phase==1},phase'=5,reads'=reads+1,sql'=sql+1}
 action retire = all {phase>=0,phase<=5,phase'=4,reads'=reads,sql'=sql}
 val safety = and {phase>=0,phase<=5,reads>=0,sql>=0,sql>=reads}
}
