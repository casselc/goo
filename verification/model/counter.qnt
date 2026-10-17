module counter {
 const INITIAL: int
 var value: int
 var epoch: int
 var known: bool
 var delta: int
 action init = all {value'=INITIAL,epoch'=1,known'=false,delta' = -1}
 action increase = all {
  value'=value+(if (epoch==1) 3 else 2),epoch'=epoch,known'=true,delta'=(if (epoch==1) 3 else 2)
 }
 action restart = all { value'=2,epoch'=2,known'=false,delta' = -1 }
 action gap = all { value'=value+5,epoch'=epoch,known'=false,delta' = -1 }
 action late_repair = all {value'=value,epoch'=epoch,known'=true,delta'=5}
 val safety = (known and delta>=0) or (not(known) and delta == -1)
}
