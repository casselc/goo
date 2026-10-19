module service_browser {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all {phase'=INITIAL,coherent'=true}
 action browse = all {phase==0,phase'=1,coherent'=true}
 action follow = all {phase==1,phase'=2,coherent'=true}
 action returnLocal = all {phase==2,phase'=3,coherent'=true}
 action preview = all {phase==3,phase'=4,coherent'=true}
 val safety = coherent
}
