// Deliberate version-discovery/comparison actions. Unrelated service overview
// reads and a separate browser actor belong to separate owners.
module serviceversions {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var pairs: int
 action init = all {phase'=INITIAL,reads'=0,sql'=0,pairs'=0}
 action discover = all {phase<4,phase'=1,reads'=reads+1,sql'=sql+2,pairs'=pairs}
 action failed_discovery = all {phase<4,phase'=0,reads'=reads+1,sql'=sql+1,pairs'=pairs}
 action local_intent = all {phase<4,phase'=phase,reads'=reads,sql'=sql,pairs'=pairs}
 action stage_pair = all {phase==1,phase'=2,reads'=reads,sql'=sql,pairs'=pairs}
 action run_pair = all {phase==2,phase'=3,reads'=reads+2,sql'=sql+6,pairs'=pairs+1}
 action return_service = all {phase==3,phase'=1,reads'=reads,sql'=sql,pairs'=pairs}
 action pending_discovery = all {phase<4,phase'=5,reads'=reads+1,sql'=sql+1,pairs'=pairs}
 action retire = all {or {phase<4,phase==5},phase'=4,reads'=reads,sql'=sql,pairs'=pairs}
 val safety = and {phase>=0,phase<=5,reads>=0,sql>=0,pairs>=0,reads>=pairs*2,sql>=pairs*6}
}
