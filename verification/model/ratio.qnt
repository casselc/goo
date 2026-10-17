module ratio {
 const INITIAL: int
 var state: str
 var calls: int
 var errors: int
 action init = all {state'="unknown",calls' = INITIAL,errors' = INITIAL}
 action traffic = all {state'="ok",calls'=10,errors'=2}
 action missing = all {state'="unknown",calls' = -1,errors' = -1}
 action idle = all {state'="no_data",calls'=0,errors'=0}
 action reset = all {state'="unknown",calls' = -1,errors' = -1}
 action gap = all {state'="unknown",calls' = -1,errors' = -1}
 action repair = all {state'="ok",calls'=3,errors'=3}
 val safety = (state=="ok" and calls>0 and errors>=0 and errors<=calls) or (state=="no_data" and calls==0 and errors==0) or (state=="unknown" and calls == -1 and errors == -1)
}
