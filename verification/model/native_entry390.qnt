module native_entry390 {
  const SERVICE_READS: int
  var reads: int
  var stores: int
  var writes: int
  var finished: bool
  action init = all { reads'=0, stores'=0, writes'=0, finished'=false }
  action local = all { not(finished), reads'=reads, stores'=stores, writes'=writes, finished'=false }
  action read = all { not(finished), reads'=reads+1, stores'=stores, writes'=writes, finished'=false }
  action openSaved = all { not(finished), reads'=reads, stores'=stores+1, writes'=writes, finished'=false }
  action serviceLoad = all { not(finished), reads'=reads+SERVICE_READS, stores'=stores+1, writes'=writes, finished'=false }
  action finish = all { not(finished), reads'=reads, stores'=stores, writes'=writes, finished'=true }
  val safety = reads>=0 and stores>=0 and stores<=4 and writes==0
}
