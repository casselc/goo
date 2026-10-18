module telemetry_packages {
 const INITIAL: int
 var phase: int
 var reusable: bool
 var immutable: bool
 action init = all {phase'=INITIAL,reusable'=true,immutable'=true}
 action draft = all {phase==0,phase'=1,reusable'=true,immutable'=true}
 action save = all {phase==1,phase'=2,reusable'=true,immutable'=true}
 action edit = all {phase==2,phase'=3,reusable'=true,immutable'=true}
 action reopen = all {phase==3,phase'=4,reusable'=true,immutable'=true}
 action apply = all {phase==4,phase'=5,reusable'=true,immutable'=true}
 val safety = reusable and immutable
}
