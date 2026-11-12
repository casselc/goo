// Only the optional aggregate owner is modeled; stream polls and an explicit
// explanation Run remain separate authorities.
module activitylogs {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var adoptions: int
 action init = all {phase'=INITIAL,reads'=0,sql'=0,adoptions'=0}
 action show = all {phase==0,phase'=1,reads'=reads+1,sql'=sql+3,adoptions'=adoptions}
 action complete = all {phase==1,phase'=2,reads'=reads,sql'=sql,adoptions'=adoptions+1}
 action refresh = all {or {phase==2,phase==3},phase'=1,reads'=reads+1,sql'=sql+3,adoptions'=adoptions}
 action freeze = all {phase==2,phase'=3,reads'=reads,sql'=sql,adoptions'=adoptions}
 action choose_interval = all {phase==3,phase'=3,reads'=reads,sql'=sql,adoptions'=adoptions}
 action explain = all {phase==3,phase'=3,reads'=reads,sql'=sql,adoptions'=adoptions}
 action return_parent = all {phase==3,phase'=3,reads'=reads,sql'=sql,adoptions'=adoptions}
 action retire = all {phase>0,phase'=4,reads'=reads,sql'=sql,adoptions'=adoptions}
 action late_result = all {phase==4,phase'=4,reads'=reads,sql'=sql,adoptions'=adoptions}
 val safety = and {phase>=0,phase<=4,reads>=0,sql==reads*3,adoptions>=0,adoptions<=reads,reads-adoptions<=1}

}
