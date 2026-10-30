// Observed fixed-bin/data/effect projection; not source truth or private proof.
module scoped_histogram_269 {
 const ORIGINAL_TOTAL: int
 var reads: int
 var plans: int
 var total: int
 var valid: int
 var selected: int
 var edge: str
 var started: bool
 action init = all {reads'=0,plans'=0,total'=0,valid'=0,selected'=0,edge'="",started'=false}
 action source = all {not(started),reads'=1,plans'=0,total'=ORIGINAL_TOTAL,valid'=13,selected'=1,edge'="9223372036854775808",started'=true}
 action brush = all {started,reads'=reads,plans'=plans,total'=total,valid'=valid,selected'=selected,edge'=edge,started'=started}
 action table = brush
 action rerun = all {started,reads'=reads+1,plans'=plans,total'=total,valid'=valid,selected'=selected,edge'=edge,started'=started}
 action stale = brush
 action choose = brush
 action execute = all {started,reads'=reads+1,plans'=plans+1,total'=selected,valid'=selected,selected'=selected,edge'=edge,started'=started}
 val safety = all {ORIGINAL_TOTAL==18,reads>=0,reads<=3,plans>=0,plans<=1,valid>=0,valid<=total,selected>=0,selected<=valid,started implies edge=="9223372036854775808"}
}
