// Configuration custody never implies an acquired or frozen telemetry operand.
module reuse {
 const INITIAL: int
 var phase: int
 var reads: int
 var sql: int
 var counted: int
 var compares: int
 var viewPosts: int
 var viewGets: int
 var captures: int
 var runs: int
 var saves: int
 var opens: int
 var receipt: bool
 action init = all {phase'=INITIAL,reads'=0,sql'=0,counted'=0,compares'=0,viewPosts'=0,viewGets'=0,captures'=0,runs'=0,saves'=0,opens'=0,receipt'=false}
 action local = all {phase'=phase,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action inspect = all {phase==0,phase'=1,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action facet = all {phase==1,phase'=2,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action load_values = all {phase==2,phase'=3,reads'=reads+1,sql'=sql+2,counted'=counted+1,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action choose_values = all {phase==3,phase'=4,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action stage_comparison = all {phase==4,phase'=5,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action execute_comparison = all {or {phase==5,phase==8},phase'=6,reads'=reads+2,sql'=sql+6,counted'=counted,compares'=compares+1,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs+1,saves'=saves,opens'=opens,receipt'=receipt}
 action preserve_late = all {or {phase==5,phase==6,phase==7},phase'=7,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts+1,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves+1,opens'=opens,receipt'=false}
 action preserve_view = all {or {phase==5,phase==6,phase==7},phase'=7,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts+1,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves+1,opens'=opens,receipt'=true}
 action open_definition = all {all {phase==7,receipt},phase'=8,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets+1,captures'=captures,runs'=runs,saves'=saves,opens'=opens+1,receipt'=receipt}
 action return_held = all {phase==6,phase'=9,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=receipt}
 action retire_context = all {phase==9,phase'=10,reads'=reads,sql'=sql,counted'=counted,compares'=compares,viewPosts'=viewPosts,viewGets'=viewGets,captures'=captures,runs'=runs,saves'=saves,opens'=opens,receipt'=false}
 val safety = and {phase>=0,phase<=10,reads==counted+2*runs,sql==2*counted+6*runs,compares==runs,viewPosts==saves,viewGets==opens,captures==0,opens<=saves,counted>=0,runs>=0,saves>=0,opens>=0}
}
