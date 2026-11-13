// Local editing preserves held observations. Only Load and Apply acquire data.
module constraints {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var counted: int
 var applied: int
 var current: bool
 var conjunctive: bool
 action init = all {phase'=INITIAL,reads'=0,sql'=0,counted'=0,applied'=0,current'=true,conjunctive'=true}
 action local = all {phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=current,conjunctive'=conjunctive}
 action inspect = all {phase==0,current,phase'=1,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=current,conjunctive'=conjunctive}
 action context = all {or {phase==1,phase==2,phase==3},current,phase'=2,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=current,conjunctive'=conjunctive}
 action edit = all {or {phase==2,phase==3},current,phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=current,conjunctive'=conjunctive}
 action choose_or = all {or {phase==2,phase==3},current,phase'=phase,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=current,conjunctive'=false}
 action load = all {phase==2,current,phase'=3,reads'=reads+1,sql'=sql+2,counted'=counted+1,applied'=applied,current'=current,conjunctive'=conjunctive}
 action remove = all {or {phase==2,phase==3},current,conjunctive,phase'=4,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=false,conjunctive'=conjunctive}
 action apply = all {phase==4,not(current),phase'=5,reads'=reads+1,sql'=sql+1,counted'=counted,applied'=applied+1,current'=true,conjunctive'=conjunctive}
 action retire = all {phase'=6,reads'=reads,sql'=sql,counted'=counted,applied'=applied,current'=false,conjunctive'=conjunctive}
 val safety = and {phase>=0,phase<=6,reads==counted+applied,sql==counted*2+applied,counted>=0,applied>=0}
}
