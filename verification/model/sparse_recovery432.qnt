module sparse_recovery432 {
  const INITIAL: int
  var phase: int
  var reads: int
  var scopeOK: bool
  var windowOK: bool
  var clockFresh: bool
  var heldOK: bool
  action init = all {phase' = INITIAL, reads' = 0, scopeOK' = false, windowOK' = false, clockFresh' = false, heldOK' = false}
  action complete_poll = all {phase == 0, phase' = 1, reads' = reads + 1, scopeOK' = true, windowOK' = true, clockFresh' = false, heldOK' = false}
  action refuse_draft = all {phase == 1, phase' = phase, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, clockFresh' = clockFresh, heldOK' = heldOK}
  action stage_wide = all {phase == 1, phase' = 2, reads' = reads, scopeOK' = true, windowOK' = true, clockFresh' = false, heldOK' = false}
  action read_wide = all {phase == 2, phase' = 3, reads' = reads + 1, scopeOK' = true, windowOK' = true, clockFresh' = true, heldOK' = false}
  action reread_wide = all {phase == 3, phase' = 4, reads' = reads + 1, scopeOK' = true, windowOK' = true, clockFresh' = true, heldOK' = false}
  action return_held = all {phase == 4, phase' = 5, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, clockFresh' = clockFresh, heldOK' = true}
  action resume_recent = all {phase == 5, phase' = 6, reads' = reads + 1, scopeOK' = true, windowOK' = true, clockFresh' = clockFresh, heldOK' = heldOK}
  action retire_context = all {phase == 6, phase' = 7, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, clockFresh' = clockFresh, heldOK' = heldOK}
  val safety = and {phase >= 0, phase <= 7, reads >= 0, reads <= 4, phase < 1 or (scopeOK and windowOK), phase < 3 or clockFresh, phase < 5 or heldOK}
}
