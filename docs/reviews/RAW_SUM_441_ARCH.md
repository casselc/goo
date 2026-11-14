# 441 raw sum observation architecture review

## Contract decision

Use a separate typed `POST /sum-observations` operation and
`goo/raw-sum-observations/v1` profile. The existing gauge operation remains a
gauge operation. This avoids an untyped union dispatcher or a fake evaluated
result while allowing shared scalar normalization and local inspection.

A sum's exact binding includes reported temporality and monotonicity. Both must
be explicit, including false. Discovery groups these fields and reclassifies
cumulative nonmonotonic sums as `stock`; therefore raw inspection must accept the
canonical `stock` kind as well as `counter`. Accepting counters alone would leave
a real OTel sum family without an observation path. Stock is exactly tempo 2 /
monotonic false; every other representable sum tuple is counter. These are
source-selection identities, not evaluator admission.

Admit reported int32 zero/unrecognized temporalities and label them
diagnostically. Reject malformed, absent or mismatched identity, because the read
cannot independently establish the requested tuple. Preserve malformed scalar,
flags or start-time facts diagnostically when representable; end time remains
mandatory and in-window. Never merge variants, infer continuity, convert rates
or manufacture producer identity.

## Review boundaries

- Validate exact scope and complete instrument before source work.
- Read only the sum family; schema/access metadata is not value readiness.
- Include temporality and monotonicity in parameterized SQL and independently
  validate returned rows against both.
- Retain duplicates, raw source projection, attributes, units and reported time.
- Keep 10,000-row and 8 MiB source/result limits atomic.
- Keep evaluator preflight, native All refusal and custody boundaries intact.
- Reuse local display/lifecycle only with the same explicit-read, cancellation,
  retained-bank and stale-result authority contracts.
- New profile labels must say sum; nonmonotonic values are not automatically
  counters, rates, valid stock values or health evidence.

## Validation status

Independent source review accepts the frozen backend implementation. The request
requires every tuple field explicitly and enforces the discovery stock rule.
SQL binds temporality as Int32 and monotonicity as Bool, reads only the sum
source, and retains the fixed half-open window. Returned metadata uses non-null
typed fields and is independently matched before publication. Unknown reported
temporalities add unspecified/unrecognized diagnostics; malformed or mismatched
metadata returns no partial result.

Neutral raw observation extraction retains the existing gauge wire and shared
strict end-time, scalar/flags/start diagnostics, duplicate keys and result
budgets. It introduces no evaluator or persistence dependency. The shared UI
owner selects the family-specific endpoint/profile, validates reported tuple
fields and preserves its explicit-read/local and suspend/discard boundaries.

Independent `go test ./measurements -run 'Raw(Gauge438|Sum441)' -count=1`
passes **0.620s** on Go 1.27.3. The focused bank covers explicit required fields,
canonical stock identity, unknown int32 temporality extremes, atomic metadata
refusal, scalar states, both byte budgets, row limit, wrong source family and
canceled/unsupported zero-read behavior. No architecture blocker remains.

Parent-owned API generation, real OTLP journey, generated properties and UX
acceptance must be reported separately; this source review does not claim those
checks ran or that raw source rows are complete or fresh.
