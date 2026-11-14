module source_capabilities_431 {
  const INITIAL: int
  var phase: int
  var canUse: bool
  var missing: bool
  action init = all { phase' = INITIAL, canUse' = false, missing' = false }
  action metadata_ok = all { phase == 0, phase' = 1, canUse' = false, missing' = false }
  action metadata_missing = all { phase == 0, phase' = 2, canUse' = false, missing' = true }
  action metadata_refused = all { phase == 0, phase' = 2, canUse' = false, missing' = false }
  action access_ok = all { phase == 1, phase' = 2, canUse' = true, missing' = false }
  action access_missing = all { phase == 1, phase' = 2, canUse' = false, missing' = true }
  action access_refused = all { phase == 1, phase' = 2, canUse' = false, missing' = false }
  action publish_compatible = all { phase == 2, canUse, not(missing), phase' = 3, canUse' = canUse, missing' = missing }
  action publish_absent = all { phase == 2, missing, not(canUse), phase' = 3, canUse' = canUse, missing' = missing }
  action publish_incompatible = all { phase == 1, not(canUse), not(missing), phase' = 3, canUse' = canUse, missing' = missing }
  action publish_unreadable = all { (phase == 1 or phase == 2), not(canUse), not(missing), phase' = 3, canUse' = canUse, missing' = missing }
  val safety = not(canUse and missing)
}
