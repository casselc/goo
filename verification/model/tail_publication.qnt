module tail_publication {
 const INITIAL: int
 var phase: int
 var durable: bool
 var refused: bool
 action init = all {phase'=INITIAL,durable'=false,refused'=false}
 action publish = all {phase==0,phase'=1,durable'=true,refused'=false}
 action recover = all {phase==1,phase'=2,durable'=true,refused'=false}
 action conflict = all {phase==2,phase'=3,durable'=true,refused'=true}
 action absent = all {phase==3,phase'=4,durable'=true,refused'=true}
 val safety = phase==0 or durable
}
