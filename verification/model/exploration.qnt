module exploration {
 const INITIAL: int
 var phase: int
 var scope: bool
 var population: bool
 var exact: bool
 action init = all {phase'=INITIAL,scope'=true,population'=true,exact'=true}
 action search = all {phase'=1,scope'=true,population'=true,exact'=true}
 action page = all {phase==1,phase'=1,scope'=true,population'=true,exact'=true}
 action trace = all {phase==1,phase'=2,scope'=true,population'=true,exact'=true}
 action logs = all {phase==2,phase'=3,scope'=true,population'=true,exact'=true}
 val safety = scope and population and exact
}
