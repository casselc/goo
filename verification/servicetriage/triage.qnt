// Reads here belong only to the explicit triage operation. Existing service
// detail reads are outside this owner's ledger and are not relabeled as local.
module triage {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var adoptions: int
 action init = all {phase'=INITIAL,reads'=0,sql'=0,adoptions'=0}
 action catalog = all {phase==0,phase'=1,reads'=reads,sql'=sql,adoptions'=adoptions}
 action measure = all {phase==1,phase'=2,reads'=reads+2,sql'=sql+4,adoptions'=adoptions+1}
 action sort = all {phase==2,phase'=phase,reads'=reads,sql'=sql,adoptions'=adoptions}
 action search = all {phase==2,phase'=phase,reads'=reads,sql'=sql,adoptions'=adoptions}
 action return_catalog = all {phase==2,phase'=phase,reads'=reads,sql'=sql,adoptions'=adoptions}
 action retire = all {phase==2,phase'=3,reads'=reads,sql'=sql,adoptions'=adoptions}
 action late_result = all {phase==3,phase'=phase,reads'=reads,sql'=sql,adoptions'=adoptions}
 val safety = and {phase>=0,phase<=3,reads>=0,reads<=2,sql==reads*2,adoptions>=0,adoptions<=1}
}
