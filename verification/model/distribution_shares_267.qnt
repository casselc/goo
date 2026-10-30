// Projection of observed UI effects, not private source/ownership proof.
// Reads count explicit Explore requests, not internal SQL statements.
module distribution_shares_267 {
 const ORIGINAL_TOTAL: int
 var reads: int
 var denominator: int
 var workerPercent: int
 var staged: bool
 var started: bool
 action init = all {reads'=0,denominator'=0,workerPercent'=0,staged'=false,started'=false}
 action source = all {not(started),reads'=1,denominator'=ORIGINAL_TOTAL,workerPercent'=25,staged'=false,started'=true}
 action find = all {started,reads'=reads,denominator'=denominator,workerPercent'=workerPercent,staged'=staged,started'=started}
 action sort = find
 action measure = find
 action rerun = all {started,not(staged),reads'=reads+1,denominator'=denominator,workerPercent'=workerPercent,staged'=false,started'=started}
 action stale = find
 action stage = all {started,not(staged),reads'=reads,denominator'=denominator,workerPercent'=workerPercent,staged'=true,started'=started}
 action execute = all {started,staged,reads'=reads+1,denominator'=1,workerPercent'=100,staged'=false,started'=started}
 val safety = all {ORIGINAL_TOTAL==4,reads>=0,reads<=3,denominator>=0,workerPercent>=0,workerPercent<=100,staged implies reads==2,started implies denominator>0}
}
