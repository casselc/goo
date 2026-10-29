module record_environment_flow {
  const INITIAL: int
  var phase: int
  var preserved: bool
  var detached: bool
  var distinct: bool
  action init = all { phase' = INITIAL, preserved' = true, detached' = true, distinct' = true }
  action construct_tail = all { phase == 0, phase' = 1, preserved' = true, detached' = true, distinct' = true }
  action fixed = all { phase == 1, phase' = 2, preserved' = true, detached' = true, distinct' = true }
  action normalized = all { phase == 2, phase' = 3, preserved' = true, detached' = true, distinct' = true }
  val safety = preserved and detached and distinct
}
