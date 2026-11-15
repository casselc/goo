// Paired Activity transport ownership, not source completeness or cross-host order.
module paired_activity444 {
  const INITIAL: int
  var logsReads: int
  var spansReads: int
  var logsActive: bool
  var spansActive: bool
  // Logical owner generations; private abort-counter increments are not modeled.
  var logsGeneration: int
  var spansGeneration: int
  var away: bool
  var held: bool

  action init = all { logsReads'=INITIAL, spansReads'=INITIAL, logsActive'=false,
    spansActive'=false, logsGeneration'=0, spansGeneration'=0, away'=false, held'=true }
  action stage = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=false, spansActive'=false, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration+1, away'=away, held'=true }
  // A request can settle into a refused lane. It never controls its sibling.
  action start(l: bool, s: bool): bool = all { not(away), logsReads'=logsReads+1, spansReads'=spansReads+1,
      logsActive'=l, spansActive'=s, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration+1, away'=away, held'=true }
  action resumeLogs(running: bool): bool = all { not(away), logsReads'=logsReads+1, spansReads'=spansReads,
      logsActive'=running, spansActive'=spansActive, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration, away'=away, held'=true }
  action resumeSpans(running: bool): bool = all { not(away), logsReads'=logsReads, spansReads'=spansReads+1,
      logsActive'=logsActive, spansActive'=running, logsGeneration'=logsGeneration, spansGeneration'=spansGeneration+1, away'=away, held'=true }
  action pauseLogs = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=false, spansActive'=spansActive, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration, away'=away, held'=true }
  action pauseSpans = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=logsActive, spansActive'=false, logsGeneration'=logsGeneration, spansGeneration'=spansGeneration+1, away'=away, held'=true }
  action pauseBoth = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=false, spansActive'=false, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration+1, away'=away, held'=true }
  action navigate = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=false, spansActive'=false, logsGeneration'=logsGeneration+1, spansGeneration'=spansGeneration+1, away'=true, held'=true }
  action back = all { away, logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=false, spansActive'=false, logsGeneration'=logsGeneration, spansGeneration'=spansGeneration, away'=false, held'=true }
  action observe = all { logsReads'=logsReads, spansReads'=spansReads,
    logsActive'=logsActive, spansActive'=spansActive, logsGeneration'=logsGeneration, spansGeneration'=spansGeneration, away'=away, held'=true }
  action step = {
    nondet l = Set(true,false).oneOf()
    nondet s = Set(true,false).oneOf()
    any {stage, start(l,s), resumeLogs(l), resumeSpans(s), pauseLogs, pauseSpans,
      pauseBoth, navigate, back, observe}
  }
  val safety = held and logsReads>=0 and spansReads>=0 and
    logsGeneration>=logsReads and spansGeneration>=spansReads and
    (not(away) or (not(logsActive) and not(spansActive)))
}

module paired_activity444_instance {
  import paired_activity444(INITIAL=0).*
}
