# Native gauge display review — milestone 386

Read-only architecture and usability source review, November 9, 2026. No runtime,
heavy tests or production edits were performed by this reviewer. Final source, visual and runtime acceptance is recorded in the follow-ups below.

## Shared contract

The implementation chooses one `NativeView` Go/API/UI owner rather than a second
gauge manager. `question.definition.kind` and the closed declaration distinguish
histogram export-interval histogram/table displays from gauge retained-observations
timeline/table displays. Gauge `requiresComplete: false` must be explicit; it does
not promise continuous or complete upstream collection. Metadata summaries name
family and selection, preventing a same-unit table from implying cross-family
compatibility. Public routes use `/native-views`; existing private
`native-histogram-views` storage remains in place without moving, rewriting or
rehashing retained artifacts. Specialized histogram projection stays separate.

`ReuseNativeView` validates both questions, requires the same family/grain and exact
unit, and detaches the entire destination question. Mutable services, attributes,
namespace, environment selection and exact-stream maps are copied. The UI likewise
validates the returned destination question, clears the previous result and waits
for explicit Run. Browse, load, save and Apply are configuration operations, not
telemetry acquisition. Gauge units must be nonempty and compiler-approved, matching the existing
measurement contract; this milestone does not add unitless admission.

The gauge renderer reuses the existing native timeline and projects the same
retained samples as a table. A local cached switch does not acquire data. Both
projections retain stale/unknown observations and the latest-200 disclosure;
selection remains a reported observation rather than a histogram interval.
Freshness is evaluated against the fixed question end, not wall clock. The new
gauge validator is separate from histogram interval/exemplar selectors, preserving
those boundaries.

## Candidate source blockers

1. `workbench/native_views.go` exact-stream wire allowlist omits
   `resourceSchemaURL`, which `metrics.Stream` and the UI accept as real identity.
   A nonempty reported resource schema URL must survive save/load/reuse; stripping
   it would weaken exact-stream authority. Parent has assigned a correction.
2. The new general `nativeObject` helper rejects every null property. This exceeds
   the intended closed-shape validation and disagrees with existing UI/base
   canonical forms: source `attributes:null`, `namespace:null`, `exactStream:null`
   and exact-stream `attributes:null` are legitimate omitted/empty representations.
   Preserve those optional semantics while explicitly refusing null required
   fields, missing/null `requiresComplete`, forbidden source resource authority and
   unrelated measurement policies. An exact-stream null resource map still cannot
   establish mandatory service/instance identity. Paired old histogram/new gauge
   wire tests should cover these representations.

3. `nativeQuestionWire` decodes the definition into a map without an exact-case
   closed-key check. Go's typed `DisallowUnknownFields` accepts case-insensitive
   field aliases that the UI refuses. Use a finite service gauge/histogram
   definition key guard; keep unrelated arm fields forbidden by presence, including
   null. No generic validation framework is needed.

The UI owner has also identified minor copy corrections: the shared question
refusal still describes only export intervals, and loaded unitless configurations
can print `undefined`. These should be corrected in the pending batch; they are
not grounds for changing the measurement contract.

## Remaining acceptance

After the source corrections, verify ordinary histogram artifacts remain readable,
nonempty resource schema identity survives all configuration operations, optional
null forms agree across Go/JavaScript, and malformed or cross-family declarations
remain refused. Actual runtime must show gauge save/browse/load/Apply, unchanged
destination authority, explicit-run acquisition, supplied-data Timeline/Table
switching and stale/unknown handling. Unit and source review are not substitutes
for that journey.


## Final source follow-up

The three candidate blockers above are closed in the frozen source. The exact
stream allowlist now includes `resourceSchemaURL`; optional null attribute maps,
namespace and exact-stream selectors retain their baseline meaning, while null
map values remain refused. Source resource authority remains explicitly forbidden.
The definition now has a finite exact-case key allowlist, excluding unrelated arm
configuration even when null. The separate UI gauge validator now also requires a
nonempty unit, matching Go compilation rather than broadening calculation policy.

The parent reports full root Go and final targeted workbench/command tests passed.
This reviewer inspected the corrections but did not rerun those checks. No
remaining source blocker was identified at this checkpoint. Generated API, final
UI checks and actual runtime acceptance remain pending. The shared question
refusal and loaded-unit label corrections are also present in the final UI source.


## Visual semantics correction and refreshed runtime

The first runtime exposed a visual blocker in the shared value timeline: unknown
observations had no numeric value but their question marks were positioned at the
numeric maximum. Truthful detail text did not remove that positional ambiguity.
The bounded component correction now places unknown/no-data marks in a labelled
“Unavailable observations · no numeric value” rail outside the numeric plot.
Timestamp X positions, button selection and selected-state synchronization remain
intact; numeric domain and zero baseline remain independent of unavailable marks.
The focused component regression checks these boundaries and input immutability.

This reviewer inspected both the final source and the refreshed
`386-gauge-retained-timeline.png`. The frame visibly separates the unknown marker
from numeric values, shows stale recorded zero on the zero baseline, and exposes
selected unknown detail without inventing a value. This closes the visual blocker.
The parent reports 1,378 frontend checks passed and the refreshed actual main
journey plus wide/narrow raw-DOM checks passed. The actor was still finishing at
this checkpoint; no final actor result is asserted here.

The 390-pixel identity/diagnostic/save-control stack remains verbose and pushes
the chart below the initial frame. This is recorded for the next whole-product
simplification pass, not represented as an above-the-fold success or a blocker to
the corrected data semantics.

## Final runtime and geometry follow-up

The parent inspected the final refreshed desktop/narrow frames and actor JPEG.
An intermediate rail version inherited global section padding, shifting physical
time positions despite matching percentage attributes. Scoped primitive resets
remove that inset; unknown intervals also retain the numeric track's one-pixel
minimum rather than an oversized hit-target width. Final raw DOM admission checks
all four actual marker centers against the common numeric time axis: maximum
error 0.015625px desktop and 0.010417px narrow, with equal track left/width.

The journey owner reports final gauge Playwright 56.6s, existing histogram
regression 11.6s, 814 actor states with observed exact question/selection/rail and
no extra reads, 25 actual responses / 7 valid request bodies plus a rejected counter
body. The verification owner independently admits the final source, bytes and
geometry and passes required Quint in 3.873s, zero skips: 11 measured steps,
14 rejected semantic mutants and the rejected read-on-Apply model control.
No runtime or source blocker remains at this checkpoint. The earlier candidate
and failed intermediate reports are not substituted for these final outcomes.
