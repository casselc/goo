// A single coordinator's causal checkpoints; not a model of browser scheduling
// or proof that captured telemetry is complete.
module cross_app {
  const INITIAL: int
  var phase: int
  var subject: str
  var trace: str
  var environment: str
  var fromMs: int
  var toMs: int
  var count: int
  var members: str

  action init = all {
    phase' = INITIAL, subject' = "", trace' = "", environment' = "",
    fromMs' = 0, toMs' = 0, count' = 0, members' = ""
  }
  action accepted(id: str): bool = all {
    phase == 0, id != "", phase' = 1, subject' = id,
    trace' = trace, environment' = environment, fromMs' = fromMs,
    toMs' = toMs, count' = count, members' = members
  }
  action source_ready(id: str, traceID: str, env: str, lower: int, upper: int, n: int, ids: str): bool = all {
    phase == 1, id == subject, traceID != "", env != "",
    lower >= 0, upper > lower, upper - lower <= 7200000,
    n > 0, n <= 100, ids != "",
    phase' = 2, subject' = subject, trace' = traceID, environment' = env,
    fromMs' = lower, toMs' = upper, count' = n, members' = ids
  }
  action displayed(id: str, traceID: str, env: str, lower: int, upper: int, n: int, ids: str): bool = all {
    phase == 2, id == subject, traceID == trace, env == environment,
    lower == fromMs, upper == toMs, n == count, ids == members,
    phase' = 3, subject' = subject, trace' = trace, environment' = environment,
    fromMs' = fromMs, toMs' = toMs, count' = count, members' = members
  }
  action explored(steps: int, passed: bool): bool = all {
    phase == 3, steps > 0, passed,
    phase' = 4, subject' = subject, trace' = trace, environment' = environment,
    fromMs' = fromMs, toMs' = toMs, count' = count, members' = members
  }
  val safety = phase >= 0 and phase <= 4 and (phase < 2 or (count > 0 and toMs > fromMs))
}
