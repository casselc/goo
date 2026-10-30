// Projected host execution fencing. This model checks observed local effects;
// it proves no source authenticity or completeness of telemetry collection.
// fence counts observed lifecycle boundaries; it is NOT Session.generation.
module scope_actions_266 {
 const MAX_READS: int
 var fence: int
 var reads: int
 var staged: int
 var eligible: bool
 var accepted: bool
 var started: bool
 action init = all {fence'=0,reads'=0,staged'=0,eligible'=false,accepted'=false,started'=false}
 action source = all {not(started),fence'=1,reads'=1,staged'=0,eligible'=true,accepted'=false,started'=true}
 action rerun = all {started,eligible,fence'=fence+1,reads'=reads+1,staged'=staged,eligible'=true,accepted'=false,started'=started}
 action stale = all {started,eligible,fence'=fence,reads'=reads,staged'=staged,eligible'=eligible,accepted'=false,started'=started}
 action suspend = all {started,eligible,fence'=fence+1,reads'=reads,staged'=staged,eligible'=false,accepted'=false,started'=started}
 action suspended = all {started,not(eligible),fence'=fence,reads'=reads,staged'=staged,eligible'=false,accepted'=false,started'=started}
 action resume = all {started,not(eligible),fence'=fence+1,reads'=reads+1,staged'=staged,eligible'=true,accepted'=false,started'=started}
 action stage = all {started,eligible,fence'=fence+1,reads'=reads,staged'=staged+1,eligible'=false,accepted'=true,started'=started}
 val safety = all {fence>=0,reads>=0,reads<=MAX_READS,staged>=0,staged<=1,accepted implies staged == 1 and not(eligible),not(started) implies fence==0 and reads==0 and staged==0,started implies fence>=1 and reads>=1}
}
