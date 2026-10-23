module tail_membership {
 const INITIAL: int
 var phase: int
 var coherent: bool
 var count: int
 var available: bool
 var denied: bool
 action init = all {phase'=INITIAL,coherent'=true,count'=0,available'=false,denied'=false}
 action issue = all {phase==0,phase'=1,coherent'=true,count'=0,available'=true,denied'=false}
 action resolve_members = all {phase==1,phase'=2,coherent'=true,count'=2,available'=true,denied'=false}
 action expire = all {phase==2,phase'=3,coherent'=true,count'=0,available'=false,denied'=false}
 action refuse = all {phase==3,phase'=4,coherent'=true,count'=0,available'=false,denied'=true}
 val safety = coherent and (available or count==0) and (not(denied) or not(available))
}
