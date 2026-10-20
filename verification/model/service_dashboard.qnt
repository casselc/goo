module service_dashboard {
 const INITIAL: int
 var coherent: bool
 var phase: int
 action init = all {coherent'=true,phase'=INITIAL}
 action switch_view = all {phase==0,phase'=phase+1,coherent'=true}
 action inspect_question = all {phase==1,phase'=phase+1,coherent'=true}
 action stage_question = all {phase==2,phase'=phase+1,coherent'=true}
 action capture = all {phase==3,phase'=phase+1,coherent'=true}
 action reopen = all {phase==4,phase'=phase+1,coherent'=true}
 val safety = coherent
}
