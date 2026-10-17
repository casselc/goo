module histogram {
 const INITIAL: int
 var known: bool
 var count: int
 action init = all {known'=false,count'=INITIAL}
 action increase = all {known'=true,count'=3}
 action restart = all {known'=false,count' = -1}
 action gap = all {known'=false,count' = -1}
 action late_repair = all {known'=true,count'=3}
 action merge = all {known'=true,count'=6}
 action bad_merge = all {known'=false,count' = -1}
 val safety = (known and count>=0) or (not(known) and count == -1)
}
