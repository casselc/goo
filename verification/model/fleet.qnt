module fleet {
 const INITIAL: int
 var state: str
 var calls: int
 var errors: int
 var supported: int
 action init = all {state'="unknown",calls'=INITIAL,errors'=INITIAL,supported'=0}
 action aligned = all {state'="ok",calls'=110,errors'=12,supported'=2}
 action missing = all {state'="unknown",calls' = -1,errors' = -1,supported'=1}
 action skew = all {state'="unknown",calls' = -1,errors' = -1,supported'=0}
 action reset = all {state'="unknown",calls' = -1,errors' = -1,supported'=1}
 action idle = all {state'="no_data",calls'=0,errors'=0,supported'=2}
 val safety = (state=="ok" and supported==2 and calls>0 and errors>=0 and errors<=calls) or (state=="no_data" and supported==2 and calls==0 and errors==0) or (state=="unknown" and supported<2 and calls == -1 and errors == -1)
}
