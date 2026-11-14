module tail_refusal433 {
  const INITIAL: int
  var phase: int
  var reads: int
  var policyOK: bool
  var scopeOK: bool
  var heldOK: bool
  var ownerOK: bool
  var noWiden: bool
  action init = all {phase' = INITIAL, reads' = 0, policyOK' = true, scopeOK' = true, heldOK' = true, ownerOK' = true, noWiden' = true}
  action missing = all {phase == 0, phase' = 1, reads' = reads + 1, policyOK' = true, scopeOK' = true, heldOK' = true, ownerOK' = true, noWiden' = true}
  action quiet = all {phase == 1, phase' = 2, reads' = reads, policyOK' = policyOK, scopeOK' = scopeOK, heldOK' = heldOK, ownerOK' = ownerOK, noWiden' = noWiden}
  action schema = all {phase == 2, phase' = 3, reads' = reads + 1, policyOK' = true, scopeOK' = true, heldOK' = true, ownerOK' = true, noWiden' = true}
  action recover = all {phase == 3, phase' = 4, reads' = reads + 1, policyOK' = true, scopeOK' = true, heldOK' = true, ownerOK' = true, noWiden' = true}
  action fail_held = all {phase == 4, phase' = 5, reads' = reads, policyOK' = true, scopeOK' = true, heldOK' = true, ownerOK' = true, noWiden' = true}
  action inspect = all {phase == 5, phase' = 6, reads' = reads, policyOK' = policyOK, scopeOK' = scopeOK, heldOK' = heldOK, ownerOK' = true, noWiden' = noWiden}
  action refuse_dirty = all {phase == 6, phase' = 7, reads' = reads, policyOK' = policyOK, scopeOK' = scopeOK, heldOK' = heldOK, ownerOK' = true, noWiden' = noWiden}
  action refuse_hidden = all {phase == 7, phase' = 8, reads' = reads, policyOK' = policyOK, scopeOK' = scopeOK, heldOK' = heldOK, ownerOK' = true, noWiden' = noWiden}
  val safety = and {phase >= 0, phase <= 8, reads >= 0, reads <= 3, policyOK, scopeOK, heldOK, ownerOK, noWiden}
}

module tail_probe433 {
  const INITIAL: int
  const MODE: int
  var phase: int
  var reads: int
  var retryMode: int
  var scopeOK: bool
  var held: bool
  action init = all {phase' = INITIAL, reads' = 0, retryMode' = MODE, scopeOK' = true, held' = false}
  action refused = all {phase == 0, phase' = 1, reads' = reads + 1, retryMode' = MODE, scopeOK' = true, held' = false}
  action recovered = all {phase == 1, phase' = 2, reads' = reads + 1, retryMode' = retryMode, scopeOK' = scopeOK, held' = true}
  val safety = and {phase >= 0, phase <= 2, reads >= 0, reads <= 2, retryMode == MODE, scopeOK, phase < 2 or held}
}
