// Bounded local authoring authority projection. No source truth, CEL meaning,
// authentication, browser focus or global concurrency claims.
module cel_authoring436 {
 const SOURCE_READS: int
 var ready: bool
 var pending: bool
 var active: bool
 var alive: bool
 var lastValid: bool
 var accepted: bool
 var reads: int
 action init = all {ready'=false,pending'=false,active'=true,alive'=true,lastValid'=false,accepted'=false,reads'=0}
 action edit = all {alive,ready'=false,pending'=false,active'=active,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action check = all {alive,active,ready'=false,pending'=true,active'=active,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action accept = all {alive,active,pending,ready'=true,pending'=false,active'=active,alive'=alive,lastValid'=true,accepted'=true,reads'=reads}
 action failed = all {alive,active,pending,ready'=false,pending'=false,active'=active,alive'=alive,lastValid'=lastValid,accepted'=true,reads'=reads}
 action stale = all {ready'=ready,pending'=pending,active'=active,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action hide = all {alive,active,ready'=false,pending'=false,active'=false,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action show = all {alive,not(active),ready'=false,pending'=false,active'=true,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action clear = all {alive,active,ready'=true,pending'=false,active'=active,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action reparent = all {alive,ready'=false,pending'=false,active'=active,alive'=alive,lastValid'=lastValid,accepted'=false,reads'=reads}
 action dispose = all {alive,ready'=false,pending'=false,active'=false,alive'=false,lastValid'=lastValid,accepted'=false,reads'=reads}
 val safety = all {SOURCE_READS==0,reads==SOURCE_READS,ready implies alive and active, pending implies alive and active,not(ready and pending)}
}
