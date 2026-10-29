// Extends verification/model/lifecycle.qnt's frozen-custody contract.
// Two tenants share their subject and immutable event IDs. T=100, D=10,
// L=3, W=30; both accepted events are at 60 and terminals at 70.
// Deliveries before freeze are visible at cutoff 96. Later deliveries are
// at 97/98 and appear only in the correction at cutoff 100.
module generations {
  type View = { a: str, b: str, impaired: bool }
  type Counts = { eligible: int, good: int, bad: int, unknown: int }
  type State = { accepted: Set[int], successes: Set[int], failures: Set[int],
    phase: int, original: View, corrected: View, selected: int, aggregate: Counts,
    mixedRejected: bool, scopeRejected: bool, done: bool }
  var s: State
  pure val EMPTY = { a: "absent", b: "absent", impaired: false }
  pure val UNCHECKED = { eligible: -1, good: -1, bad: -1, unknown: -1 }
  pure def classification(accepted: Set[int], successes: Set[int], failures: Set[int], tenant: int): str =
    if (not(accepted.contains(tenant))) "absent"
    else if (successes.contains(tenant) and failures.contains(tenant)) "unknown"
    else if (successes.contains(tenant)) "good" else "bad"
  pure def projection(accepted: Set[int], successes: Set[int], failures: Set[int]): View = {
    a: classification(accepted,successes,failures,0),
    b: classification(accepted,successes,failures,1),
    impaired: Set(0,1).exists(t =>
      (successes.contains(t) and failures.contains(t)) or
      (not(accepted.contains(t)) and (successes.contains(t) or failures.contains(t)))),
  }
  pure def counts(v: View): Counts = {
    eligible: (if (v.a != "absent") 1 else 0) + (if (v.b != "absent") 1 else 0),
    good: (if (v.a == "good") 1 else 0) + (if (v.b == "good") 1 else 0),
    bad: (if (v.a == "bad") 1 else 0) + (if (v.b == "bad") 1 else 0),
    unknown: (if (v.a == "unknown") 1 else 0) + (if (v.b == "unknown") 1 else 0),
  }
  val current = projection(s.accepted,s.successes,s.failures)
  val view = if (s.phase == 0) current else if (s.selected == 0) s.original else s.corrected
  val totals = counts(view)
  pure def quality(v: View): str = if (v.impaired) "unknown" else if (counts(v).eligible == 0) "no_data" else "ok"
  val state = quality(view)

  val ratioTicks = if (state != "ok") -1 else if (totals.good == 0) 0 else if (totals.good == totals.eligible) 2 else 1

  action init = s' = { accepted: Set(), successes: Set(), failures: Set(), phase: 0,
    original: EMPTY, corrected: EMPTY, selected: -1, aggregate: UNCHECKED,
    mixedRejected: true, scopeRejected: true, done: false }
  action accept(tenant: int): bool = all {
    s.phase < 2, not(s.accepted.contains(tenant)), Set(0,1).contains(tenant),
    s' = { ...s, accepted: s.accepted.union(Set(tenant)) },
  }
  action terminal(tenant: int, successful: bool): bool = all {
    s.phase < 2, Set(0,1).contains(tenant),
    not(if (successful) s.successes.contains(tenant) else s.failures.contains(tenant)),
    s' = if (successful) { ...s, successes: s.successes.union(Set(tenant)) }
         else { ...s, failures: s.failures.union(Set(tenant)) },
  }
  action duplicate(tenant: int): bool = all {
    s.phase < 2, Set(0,1).contains(tenant),
    s.accepted.contains(tenant) or s.successes.contains(tenant) or s.failures.contains(tenant), s' = s,
  }
  action freeze = all {
    s.phase == 0, s' = { ...s, phase: 1, original: current, selected: 0 },
  }
  action correct = all {
    s.phase == 1, s' = { ...s, phase: 2, corrected: current, selected: 1 },
  }
  action selectGeneration(generation: int): bool = all {
    not(s.done), s.phase > 0, generation == 0 or (generation == 1 and s.phase == 2),
    s' = { ...s, selected: generation, aggregate: counts(if (generation == 0) s.original else s.corrected) },
  }
  action mixed = all {
    not(s.done), s.phase == 2, counts(s.original).eligible > 0, counts(s.corrected).eligible > 0,
    s' = { ...s, mixedRejected: true },
  }
  action changedScope = all {
    not(s.done), s.phase > 0, s' = { ...s, scopeRejected: true },
  }
  action finish = all {
    not(s.done), s.phase == 2, s.aggregate.eligible >= 0, s' = { ...s, done: true },
  }
  action idle = all { s.done, s' = s }
  action step = {
    nondet tenant = Set(0,1).oneOf()
    nondet successful = Set(true,false).oneOf()
    nondet generation = Set(0,1).oneOf()
    any { accept(tenant), terminal(tenant,successful), duplicate(tenant), freeze,
      correct, selectGeneration(generation), mixed, changedScope, finish, idle }
  }
  val safety = totals.eligible <= 2 and totals.good + totals.bad + totals.unknown == totals.eligible
  val complete = s.done
  val correctedUnknown = s.phase == 2 and (s.corrected.a == "unknown" or s.corrected.b == "unknown")
}
