// Observed display/effect projection for one independently measured group.
// No private source-proof, ownership or upstream collection claim is modeled.
module status_fractions_268 {
 const INITIAL_POPULATION: int
 const GROUP_POPULATION: int
 var reads: int
 var population: int
 var numerator: int
 var denominator: int
 var percent: int
 var staged: bool
 var started: bool
 action init = all {reads'=0,population'=0,numerator'=0,denominator'=0,percent'=0,staged'=false,started'=false}
 action source = all {not(started),reads'=1,population'=INITIAL_POPULATION,numerator'=1,denominator'=GROUP_POPULATION,percent'=1,staged'=false,started'=true}
 action find = all {started,reads'=reads,population'=population,numerator'=numerator,denominator'=denominator,percent'=percent,staged'=staged,started'=started}
 action sort = find
 action table = find
 action rerun = all {started,not(staged),reads'=reads+1,population'=population,numerator'=numerator,denominator'=denominator,percent'=percent,staged'=false,started'=started}
 action stale = find
 action stage = all {started,not(staged),reads'=reads,population'=population,numerator'=numerator,denominator'=denominator,percent'=percent,staged'=true,started'=started}
 action execute = all {started,staged,reads'=reads+1,population'=GROUP_POPULATION,numerator'=numerator,denominator'=denominator,percent'=percent,staged'=false,started'=started}
 val safety = all {INITIAL_POPULATION==103,GROUP_POPULATION==100,reads>=0,reads<=3,numerator>=0,numerator<=denominator,percent>=0,percent<=100,percent*denominator==numerator*100,started implies denominator==GROUP_POPULATION,staged implies reads==2}
}
