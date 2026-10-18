# Two-instance investigations and bounded-skew comparison

The default lab now runs two actual woven checkout instances: `operation-app`
(`woven-demo`, loopback 8090) and `operation-peer` (`woven-peer`, loopback 8092).
Both emit standard OTLP to the same collector and ClickHouse, with distinct
`service.instance.id`. `GOO_INSTANCE_ID` selects the bounded sample identity.
The fleet panel shows both members and retains strict exact-interval totals.

## Explicit comparison policy

```yaml
fleet:
  name: checkout-coverage
  ratio: operation-errors
  members: [woven-demo, woven-peer]
  comparison: {mode: bounded-skew, maxSkew: 1s}
```

Comparison is optional. Without it, only the original strict fleet semantics
apply. The planner accepts `bounded-skew` and a positive tolerance up to one
second. It contributes to package revisions and comparison/fleet evidence bases.
The API exposes `fleet.result.comparison`; its core policy uses decimal
`maxSkewNS`, and its series retain actual member windows and member counts.

Independent SDK collection windows rarely share exact nanosecond endpoints.
Goo offers **side-by-side comparison**, not approximate fleet aggregation. Every
successful comparison requires a valid whole interval from every member, with
both pairwise start spread and pairwise end spread within the declared tolerance.
The output labels those rows `comparable`, reports both spreads and retains each
member's original from/to, calls, errors, fraction and percentage. It has **no
combined calls/errors/percentage fields**. The exact fleet result remains unknown
where those same intervals do not have identical boundaries.

The measured spread is a property of reported timestamps, not a bound on host
clock error. Clock accuracy, sampling and ingestion completeness remain
unverified. Overlapping but different windows may contain different events;
comparable does not mean identical populations or statistical equivalence.

## Matching and evidence limits

Members and stream identities are sorted canonically. Within a compatible stream,
the lexically first observed member supplies reference windows. For each valid
reference interval, a deterministic greedy matcher selects the nearest valid whole
interval per member, breaking ties by earlier time. Each candidate must keep the
pairwise endpoint spread within tolerance for the whole chosen set. Candidates
are reserved only when the whole row succeeds; a member interval is never reused
in another successful comparison. This is conservative matching, not a maximum
coverage optimizer. It can leave potential matches unsupported.

Unknown source intervals, missing outcomes, absent compatible streams, gaps,
resets, flags and clipped boundaries do not become valid candidates. Unmatched
rows stay unknown, with member diagnostics and overlapping underlying counter
causes where available. A partially known row can retain a member's own evidence;
it never becomes a fleet population. Policy changes and late evidence produce new
bases without changing earlier results. No interpolation, timestamp shifting,
proportional allocation, extrapolation or averaging percentages occurs.

## Live fault journey

The sample app includes **Lab telemetry fault injection**. Pausing metric export
intentionally drops collected metric exports while HTTP requests and the separate
trace exporter continue. Resuming preserves cumulative counter state. This is an
in-memory local lab control (`POST /demo/metrics?mode=on|off`), not a production
instrumentation feature or durable audit mechanism. Default ports remain loopback.

[Journey 17](journeys/README.md#17-two-live-instances-imbalance-loss-and-restart)
creates more successful traffic on one instance and errors on the other, compares
actual windows, pauses the peer's metrics while a request succeeds, waits beyond
the declared maximum counter gap, resumes export and verifies the gap stays
unknown. It then restarts only the owned Compose peer and verifies the new counter
epoch does not become a supported comparison. A `finally` block resumes exports.
Existing volumes, queue entries and the other services are retained. This restart
journey is intended for the owned local Compose stack.

The UI keeps exact totals and bounded comparisons in separate sections. An
explanation includes actual nanosecond timestamps, policy, measured spreads,
individual populations, counter causes and comparison basis. It never presents a
bounded-skew row as a common-window fleet percentage.

## Verification and remaining work

Go race tests cover actual-window/count preservation, tolerance rejection,
unknowns, stable order/policy bases and non-reuse. Hegel runs 100 state/swarm
sequences of up to 30 transitions through calls, restart, gap, loss and skew.
Quint validates six observed aligned/skew/outside/loss/reset/recovered outcomes
and rejects invented exact populations. These bounded semantic checks supplement
the live browser journey; they are not a proof of arbitrary production histories.

Trusted discovery, membership activation windows, clock-health bounds, automatic
common-window aggregation, fleet histogram UI and warehouse lowering remain
future work. The existing exact histogram merge API is unchanged. See
[coverage contracts](FLEET_COVERAGE.md) and [diagram 8](ARCHITECTURE_DIAGRAMS.md#8-two-live-instances-and-bounded-comparisons).
