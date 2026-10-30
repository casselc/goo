// Observed display/effect projection; no private scope-authority claim.
module duration_histogram_271 {
 const GROUP_TOTAL:int
 var reads:int
 var pairs:int
 var started:bool
 action init=all{reads'=0,pairs'=0,started'=false}
 action source=all{not(started),reads'=1,pairs'=0,started'=true}
 action brush=all{started,reads'=reads,pairs'=pairs,started'=started}
 action manual=brush
 action execute=all{started,reads'=reads+1,pairs'=pairs,started'=started}
 action rerun=execute
 action stale=brush
 action rest=brush
 action compare=all{started,reads'=reads+2,pairs'=pairs+1,started'=started}
 val safety=all{GROUP_TOTAL==11,reads>=0,reads<=5,pairs>=0,pairs<=1}
}
