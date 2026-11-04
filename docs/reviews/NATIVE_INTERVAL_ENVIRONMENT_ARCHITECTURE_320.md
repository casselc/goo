# Typed native service interval values — architecture 320

Status: final source review accepted; generated execution recorded below.

Extend the existing service instrument journey to typed service-bound per-stream
counters and explicit histograms. The value is an actual interval delta/rate or
native distribution under the same checked environment as the service overview,
not another metadata-only discovery surface. Keep the 319 acquisition, package,
capture, retained-result and navigation owners.

## Admission and kernel boundaries

Typed Definition selectors admit gauge, counter and histogram only, with per-stream
grain and no resource binding. Stock, state, lifecycle, consumer-group and special
host/entity kernels remain refused before source IO. Legacy environment strings
retain current-key-only semantics and their established families.

CounterDomain remains the existing independent optional policy: ordinary counters
do not acquire a decimal requirement, and an explicit nonnegative-decimal domain
must still satisfy its existing counter/per-stream validation. Preserve monotonic,
delta/cumulative temporality, predecessor lookback, reset and gap behavior. Typed
selectors change stream membership, not arithmetic or interval boundaries.

Discovery retains the existing monotonic-counter and temporality 1/2 guards.
Explicit histogram discovery retains its existing duration-unit s/ms restriction;
unsupported stock or histogram metadata does not gain an executable Definition.
No exponential-histogram source or new fleet-percentile interpretation is added.
Existing histogramBucketMs and exact-stream checks retain their bounds and native
bucket semantics. Histograms preserve count, explicit bounds, bucket counts,
unit and reset/temporality diagnostics; do not average stream percentiles.

## Exact policy and retained-state seams

The 319 source compiler, decoded row membership and exact-stream validation already
share checked EnvironmentSelection. The extension should widen family admission,
not add SQL branches or a second policy evaluator. Raw resources remain part of
stream identity: current and legacy maps matching one effective name stay distinct.

The frontend has three closed admission gates to widen deliberately:
`native-record-context.mjs`, `question-protocol.mjs`, and the histogram-specific
`measurement-scope.mjs`. Histogram questions must retain the selector in their
closed shape, use shared environment selection for exact-stream membership, and
continue refusing typed consumer-group questions. Interval, time-bucket and window
selectors carry the complete original question and exact native bounds; they are
not rewritten to a scalar environment or inferred from a label.

Retained histogram witnesses and comparison/capture validation already bind full
question/Definition JSON. Preserve that identity through source adapters, packages,
question steps and common Scope; a changed selector must invalidate compatibility.
Existing native interval→record-context staging must carry the selector and exact
context interval, stage without IO, and require the existing explicit Run.

## Capability and acceptance

Service capability wording should state supported service per-stream metric families
rather than gauge-only. Declared entity bindings retain their separate current-only
named gate. Discovery unsupportedReason remains authoritative for capability
eligibility; no unsupported Definition is offered as executable preview.

Acceptance needs real counter and explicit histogram values through the existing
service picker, policy membership and raw-stream separation, interval and histogram
selection/retention, package and step round trips, typed interval→record question,
and refused unsupported resources/families before IO. Generated cross-language
checks should include present-empty current keys shadowing legacy names and
foreign-policy rows. Existing kernels and source request budgets remain unchanged;
actual execution is recorded in VALIDATION rather than inferred from this review.


## Final source review

The final backend change widens only typed service family admission and discovery
capability. Existing SQL, decoded membership, exact-stream validation and metric
kernels remain shared. Typed histogram duration units are explicitly restricted
to s/ms; unsupported resource/grain/family definitions still refuse. Dashboard
capability now describes service-per-stream support and retains separate declared
entity gating. Optional counterDomain validation is unchanged.

The three frontend closed gates now retain typed counter/histogram questions,
while histogram exact-stream membership uses the same checked environment policy.
Histogram selector bounds, bucket anchoring, full stream pinning and original
question identity remain intact. No source executor, inferred scalar environment,
stream merge or new snapshot authority was introduced.

A dedicated verification file generates 30 counter and 30 histogram cases. It
checks all four supported policy forms and namespace variants, matching current/
legacy resources with otherwise identical stream identity, exact one-second
counter delta/rate and unknown counter decreases, delta/cumulative histogram
bucket subtraction/count/unit, foreign-policy rejection, Go/JS native interval
context agreement, histogram selector/pinning refusal and caller/result policy
detachment. The final focused execution, including the full-stream pinning counterexamples,
passed in 4.717 seconds (60 generated cases, zero skips).
