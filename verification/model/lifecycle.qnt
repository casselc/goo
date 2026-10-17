// Bounded custody-basis model for one accepted subject and a late-delivered
// on-time terminal. This does not model arbitrary lifecycle payloads or storage.
module lifecycle {
  const MAX_SNAPSHOTS: int
  type Snapshot = { eligible: int, good: int, bad: int, unknown: bool }
  var accepted: bool
  var terminalSeen: bool
  var basis: int
  var snapshots: List[Snapshot]

  val eligible = if (accepted) 1 else 0
  val good = if (accepted and terminalSeen and basis > 95) 1 else 0
  val bad = eligible - good
  val unknown = not(accepted) and terminalSeen and basis > 95

  action init = all {
    accepted' = false, terminalSeen' = false, basis' = 80, snapshots' = List()
  }
  action accept = all {
    not(accepted), accepted' = true, terminalSeen' = terminalSeen,
    basis' = basis, snapshots' = snapshots
  }
  action terminal = all {
    not(terminalSeen), terminalSeen' = true, accepted' = accepted,
    basis' = basis, snapshots' = snapshots
  }
  action duplicate = all {
    accepted or terminalSeen,
    accepted' = accepted, terminalSeen' = terminalSeen,
    basis' = basis, snapshots' = snapshots
  }
  action advance95 = all {
    basis < 95, basis' = 95, accepted' = accepted,
    terminalSeen' = terminalSeen, snapshots' = snapshots
  }
  action advance96 = all {
    basis < 96, basis' = 96, accepted' = accepted,
    terminalSeen' = terminalSeen, snapshots' = snapshots
  }
  action advance100 = all {
    basis < 100, basis' = 100, accepted' = accepted,
    terminalSeen' = terminalSeen, snapshots' = snapshots
  }
  action snapshot = all {
    snapshots.length() < MAX_SNAPSHOTS,
    snapshots' = snapshots.append({ eligible: eligible, good: good, bad: bad, unknown: unknown }),
    accepted' = accepted, terminalSeen' = terminalSeen, basis' = basis
  }
  action step = any { accept, terminal, duplicate, advance95, advance96, advance100, snapshot }
  val safety = all {
    basis >= 80, basis <= 100, good <= eligible, bad + good == eligible,
    snapshots.indices().forall(i => snapshots.nth(i).good <= snapshots.nth(i).eligible and snapshots.nth(i).bad + snapshots.nth(i).good == snapshots.nth(i).eligible)
  }
}

module lifecycleInstance {
  import lifecycle(MAX_SNAPSHOTS = 3).*
}
