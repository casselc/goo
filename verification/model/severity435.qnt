module severity435 {
  const INITIAL: int
  var phase: int
  var reads: int
  var scopeOK: bool
  var bandsOK: bool
  var viewOK: bool
  action init = all {phase' = INITIAL, reads' = 0, scopeOK' = true, bandsOK' = true, viewOK' = true}
  action loaded = all {phase == 0, phase' = 1, reads' = reads + 1, scopeOK' = true, bandsOK' = true, viewOK' = true}
  action volume = all {phase == 1, phase' = 2, reads' = reads + 1, scopeOK' = true, bandsOK' = true, viewOK' = true}
  action band = all {or {phase == 2, phase == 3}, phase' = 3, reads' = reads, scopeOK' = scopeOK, bandsOK' = bandsOK, viewOK' = true}
  action stale_refused = all {phase == 3, phase' = phase, reads' = reads, scopeOK' = scopeOK, bandsOK' = bandsOK, viewOK' = viewOK}
  action explain = all {phase == 3, phase' = phase, reads' = reads, scopeOK' = scopeOK, bandsOK' = bandsOK, viewOK' = viewOK}
  action stage_filter = all {phase == 3, phase' = 4, reads' = reads, scopeOK' = true, bandsOK' = bandsOK, viewOK' = viewOK}
  action apply = all {phase == 4, phase' = 5, reads' = reads + 1, scopeOK' = true, bandsOK' = bandsOK, viewOK' = viewOK}
  action retire = all {phase == 5, phase' = 6, reads' = reads, scopeOK' = scopeOK, bandsOK' = bandsOK, viewOK' = viewOK}
  val safety = and {phase >= 0, phase <= 6, reads >= 0, reads <= 3, scopeOK, bandsOK, viewOK}
}
