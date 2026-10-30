// Settled effect ownership, not authentication, concurrency or whole-UI liveness.
// current/baseline are harness ordinals of observed populations, not source IDs.
module summary_comparison_264 {
  const LIMIT: int
  var reads: int
  var comparisons: int
  var current: int
  var baseline: int
  action init = all { reads' = 0, comparisons' = 0, current' = 0, baseline' = 0 }
  action applyRun = all {
    current < LIMIT,
    comparisons < LIMIT * 2,
    reads' = reads + 1,
    current' = current + 1,
    baseline' = baseline,
    comparisons' = if (baseline > 0) comparisons + 1 else comparisons
  }
  action pin = all {
    current > 0,
    comparisons < LIMIT * 2,
    reads' = reads,
    current' = current,
    baseline' = current,
    comparisons' = comparisons + 1
  }
  action stage = all { current > 0, reads' = reads, comparisons' = comparisons, current' = current, baseline' = baseline }
  action inspect = all { baseline > 0, reads' = reads, comparisons' = comparisons, current' = current, baseline' = baseline }
  action reject = all { current > 0, reads' = reads, comparisons' = comparisons, current' = current, baseline' = baseline }
  action step = any { applyRun, pin, stage, inspect, reject }
  val safety = reads == current and current >= 0 and current <= LIMIT and baseline >= 0 and baseline <= current and comparisons >= 0 and comparisons <= LIMIT * 2 and (baseline > 0 or comparisons == 0)
}
