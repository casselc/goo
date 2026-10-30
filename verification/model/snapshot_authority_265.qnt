// Projected local ownership model for the four-step browser journey.
// This proves no authentication, source truth or general query completeness.
module snapshot_authority_265 {
 const ORIGINAL_ROWS: int
 var sourceReads: int
 var rows: int
 var complete: bool
 var entry: str
 var started: bool
 action init = all {sourceReads'=0,rows'=0,complete'=false,entry'="none",started'=false}
 action source = all {not(started),sourceReads'=1,rows'=ORIGINAL_ROWS,complete'=true,entry'="source",started'=true}
 action copy = all {started,sourceReads'=sourceReads,rows'=1,complete'=false,entry'="held",started'=started}
 action subset = all {started,sourceReads'=sourceReads,rows'=1,complete'=false,entry'="projection",started'=started}
 action restore = all {started,sourceReads'=sourceReads,rows'=ORIGINAL_ROWS,complete'=true,entry'="source",started'=started}
 val safety = all {
  ORIGINAL_ROWS == 2,
  sourceReads >= 0 and sourceReads <= 1,
  started implies sourceReads == 1,
  not(started) implies rows == 0 and not(complete) and entry == "none",
  entry == "source" implies rows == ORIGINAL_ROWS and complete,
  entry == "held" or entry == "projection" implies rows == 1 and not(complete)
 }
}
