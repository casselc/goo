module metric_scopes432 {
  const INITIAL: int
  var phase: int
  var windowOK: bool
  var scopeOK: bool
  var detached: bool
  action init = all {phase' = INITIAL, windowOK' = false, scopeOK' = false, detached' = false}
  action catalog = all {phase == 0, phase' = 1, windowOK' = true, scopeOK' = false, detached' = false}
  action choose = all {phase == 1, windowOK, phase' = 2, windowOK' = windowOK, scopeOK' = true, detached' = false}
  action native_read = all {phase == 2, scopeOK, phase' = 3, windowOK' = windowOK, scopeOK' = true, detached' = true}
  val safety = (phase < 2 or scopeOK) and (phase < 3 or detached)
}
