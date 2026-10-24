module investigation_draft {
 const INITIAL: int
 var pending: bool
 var unknown: bool
 var coherent: bool
 action init = all {pending'=INITIAL==1,unknown'=false,coherent'=true}
 action checkpoint = all {not(unknown),pending'=true,unknown'=false,coherent'=true}
 action reload = all {pending'=pending,unknown'=unknown,coherent'=true}
 action restore = all {pending,not(unknown),pending'=true,unknown'=false,coherent'=true}
 action uncertain = all {pending,pending'=true,unknown'=true,coherent'=true}
 action restoreUnknown = all {pending,unknown,pending'=true,unknown'=true,coherent'=true}
 action editUnknown = all {pending,unknown,pending'=true,unknown'=true,coherent'=true}
 action lateAck = all {pending,unknown,pending'=true,unknown'=true,coherent'=true}
 action saveSuccess = all {pending,pending'=false,unknown'=false,coherent'=true}
 action discard = all {pending'=false,unknown'=false,coherent'=true}
 action invalid = all {pending'=false,unknown'=false,coherent'=true}
 val safety = coherent and (not(unknown) or pending)
}
