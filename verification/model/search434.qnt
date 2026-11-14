module search434 {
  const INITIAL: int
  var phase: int
  var reads: int
  var scopeOK: bool
  var windowOK: bool
  var literalOK: bool
  var heldOK: bool
  action init = all {phase' = INITIAL, reads' = 0, scopeOK' = true, windowOK' = true, literalOK' = true, heldOK' = true}
  action loaded = all {phase == 0, phase' = 1, reads' = reads + 1, scopeOK' = true, windowOK' = true, literalOK' = true, heldOK' = true}
  action find = all {phase == 1, phase' = phase, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, literalOK' = literalOK, heldOK' = heldOK}
  action dirty_refused = all {phase == 1, phase' = phase, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, literalOK' = literalOK, heldOK' = heldOK}
  action stage = all {phase == 1, phase' = 2, reads' = reads, scopeOK' = true, windowOK' = true, literalOK' = true, heldOK' = true}
  action read_source = all {phase == 2, phase' = 3, reads' = reads + 1, scopeOK' = true, windowOK' = true, literalOK' = true, heldOK' = true}
  action return_held = all {phase == 3, phase' = 4, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, literalOK' = literalOK, heldOK' = true}
  action retire = all {phase == 4, phase' = 5, reads' = reads, scopeOK' = scopeOK, windowOK' = windowOK, literalOK' = literalOK, heldOK' = heldOK}
  val safety = and {phase >= 0, phase <= 5, reads >= 0, reads <= 2, scopeOK, windowOK, literalOK, heldOK}
}
