# Counter display reuse review — 389

Initial architecture/usability review, November 9, 2026. The contract below was
agreed with the parent and implementation agents before adding the counter arm.
Implementation source has been inspected and its source blockers are closed.
The parent reports refreshed runtime/image and required-model acceptance passed,
with zero skips. No bounded 389 review gate remains pending. Chronological
contract/source follow-ups follow.
This reviewer has read source and documentation; no tests or heavy jobs were run.

## Agreed applicability

Extend the existing `NativeView` owner, store and routes. Admit unbucketed,
per-stream, service-bound `counter` questions accepted by the current measurement
compiler. Resource-bound counters, stock/state/lifecycle families, consumer-group
grain and histogram time buckets remain outside this arm.

Counter configuration is a timeline or table over **evaluated intervals**, not
gauge samples or reconstructed counts:

```json
{
  "selection": {"kind": "evaluated-intervals"},
  "kind": {"name": "timeline", "version": 1},
  "display": {
    "category": "time",
    "value": "counter-rate",
    "requiresComplete": false
  }
}
```

Reuse requires the same measurement family, per-stream grain, exact literal
instrument unit and normalized counter domain. In authored questions the existing
empty/omitted `counterDomain` selects the integer kernel; `nonnegative-decimal`
selects the separate decimal kernel. Equal units do not permit cross-domain reuse.
The catalog exposes the normalized names `nonnegative-integer` and
`nonnegative-decimal` in a required counter-only `counterDomain` summary field.
Those catalog names do not expand the authored question's domain vocabulary.

The current evaluator already fixes the supported rate/reset semantics:
monotonic cumulative or delta exports, exact interval bounds, no subtraction
across epochs, unknown reset/decrease/gap/flag/boundary intervals and no
interpolation. Temporality is reported source evidence, not an authored field in
the question. No universal policy descriptor, arbitrary reset mode or new rate
engine is needed. Unsupported policy fields must be refused rather than silently
ignored.

## Calculation and custody boundaries

Reuse transfers display intent only. The **whole destination question** wins,
including service and exact stream, environment/namespace, unit/domain, fixed
bounds and `maxGapMs`. Donor values, selected interval, clocks, evidence basis and
freshness/gap policy must not transfer. Reuse remains unsaved configuration,
without a donor content ID. Apply clears prior evidence and selection and waits
for explicit Run; Browse and Inspect do not read telemetry.

Catalog compatibility is a metadata hint, not loaded/applicable configuration.
Inspect must compare the returned full view to the exact requested ID and catalog
summary, including normalized domain. Otherwise a stale or dishonest integer
summary could describe a decimal artifact even when final destination checks are
otherwise correct. Counter domain must be absent on gauge/histogram summaries.

## Honest presentation

Reuse the existing supplied-data timeline and table components. A valid zero
interval remains numeric zero; unknown intervals have no rate and remain gaps or
unavailable marks with their reasons. Show exact delta strings and optional
`exactPerSecond` where the evaluator supplies it. Do not manufacture an exact-rate
field for the integer arm, derive a rate from displayed totals, or sum independent
streams. Base units remain literal; a displayed rate uses that unit per second.

Preserve evaluated reset/epoch-change reasons and the existing source evidence
that carries epoch starts. Do not invent an epoch identity absent from the
interval DTO. Decimal Float64 projections and exact rational text retain their
different meanings. Limits on rendered intervals do not change evaluation or
capture totals. No continuity, complete source ingestion or application-health
claim follows from `requiresComplete: false`.

## Bounded acceptance

- Go/JavaScript admission agrees on the closed counter arm, default versus
  explicit decimal domain, explicit false completeness and unsupported-field
  refusal. Gauge/histogram behavior remains covered.
- Immutable save/load/list round trips retain exact bytes and content identity;
  donor and destination caller objects are detached from the returned view.
- Generated applications preserve the complete destination question and reject
  equal-unit cross-domain, wrong-unit, family, grain and policy substitutions.
- A real save → distinct destination → Inspect/Apply → explicit Run journey
  demonstrates different gap policy, zero, reset/epoch and unknown intervals.
  Raw source/evaluation checks, supplied-data paint and read budgets are separate
  from finite model replay.
- Actual storyboards and frames show the distinction between configuration and
  evidence. Completed browser outcomes and negative controls are required;
  clicks or a conforming prefix alone are insufficient.

The [389 roadmap](../ROADMAP_2026_11_09_CHECKPOINT_387.md#389--depth-reuse-evaluated-counter-displays)
owns scope. Native layout consolidation remains 390, rather than an additional
counter-specific layout or component library in this cut.

## Final source inspection

The frozen Go implementation in `workbench/native_views.go` uses the measurement
compiler as its question gate, adds the agreed evaluated-interval/rate arm, and
compares the authored domain as well as family, grain and exact unit during reuse.
The existing detach helper preserves caller isolation and transfers only the
complete destination question. Existing store and HTTP routes remain the owners.
Closed JSON admission allows `counterDomain` only on counter questions by field
presence, refusing null, wrong case and unrelated policy fields. Listing emits a
normalized counter-only summary domain from verified full artifacts.

The browser source in `native-view-declaration.mjs` normalizes explicit empty
integer domain to omission for identity comparisons without changing the authored
configuration returned to callers. Compatibility checks require domain as well
as family/unit. `native-view-reuse.mjs` includes domain in Find/summary text and
checks the loaded view against that exact catalog summary before enabling Apply.
Existing owner/generation and explicit Run controls remain intact.

`renderCounterDisplay` consumes existing evaluated intervals and supplies either
the common timeline or one compact table. Known zero stays zero; reset/gap
intervals have no numeric rate. The table preserves exact delta and optional
decimal rational rate text, and local switching does not recompute the kernel.
Supported service counters no longer render an extra legacy interval table under
the selected view. Resource-bound/unsupported-question callers retain their
existing presentation path and do not gain native-display applicability.

Inspected tests cover both domain/kind arms, immutable round trips, detached
whole-destination scope/policy, exact request/calculation identity, wire presence,
cross-domain refusal, HTTP source-read exclusion, supplied zero/reset/gap/exact
rate paint and local kind switching. The parent reports focused Go/HTTP checks
passed; this reviewer did not run them. Full UI/generated/runtime/model checks
remain separately owned acceptance gates and are not implied by source review.

## Schema default-value alignment

The parent identified that the shared authored `Definition.counterDomain` schema
enumerated only `nonnegative-decimal`, while the Go string field and validation
also accept the empty default. Allowing `""` alongside `nonnegative-decimal` in
that shared string enum matches implementation semantics and the agreed native
counter input contract. It does not add a new policy or a second schema engine.
Output `omitempty` still omits the default. Null remains invalid, the normalized
catalog name `nonnegative-integer` remains invalid as an authored value, and the
native-view conditional field-presence gate still rejects `counterDomain` on
gauge/histogram questions even when its string value is empty. No concern was
identified with this narrowly scoped schema correction.

## Actual table wording follow-up

The parent visually inspected the actual table and identified a presentation
ambiguity inherited from the generic table's null default, `Not recorded`.
Counter exports can be recorded while a rate/delta is unavailable because of an
excessive gap, reset, decrease or tainted epoch. The integer kernel also does not
emit an `exactPerSecond` field for an otherwise valid rate. Neither situation
should imply missing source records.

A counter-local presentation correction is appropriate: unavailable derived
delta/rate cells say **Unavailable**; a known interval with no supplied exact-rate
field says **Not supplied**. Keep numeric zero, raw null/omitted DTO values,
calculation, clocks and source basis unchanged. The generic table default and
other families need no change. Focused DOM wording checks and refreshed actual
table frames are the bounded acceptance for this display correction; no new
evaluation or policy mechanism is warranted.

The final source implements that correction only in `renderCounterDisplay`'s
cell projection. Unknown intervals supply three explicit `Unavailable` labels;
known intervals retain their supplied delta and numeric rate, with
`exactPerSecond ?? "Not supplied"` for the optional exact field. This preserves
recorded zero and changes neither input objects nor generic table behavior.
Inspected DOM regressions assert decimal zero/exact-zero, three unavailable
cells for reset/gap rows, integer zero with unsupplied exact rate, no misleading
`Not recorded` text, unchanged series and no additional reads. The parent reports
19 focused tests passed in 522.8 ms. This reviewer did not rerun them. The wording
source blocker is closed; refreshed actual table frames remain a separate gate.

## Refreshed actual runtime acceptance

The parent reports the final actual browser journey passed in 1.1 minutes with
11 measured facts. The UI performed three explicit native reads; a separate
decimal HTTP check performed one read. The physical source bank contains 17
rows, including two decimal rows supporting an exact `1/4` rate. The final actor
produced 687 snapshots through 11 phases, with one source read, three metadata
operations and zero writes. It independently observed current ownership, exact
intervals and points, the unavailable rail, selected interval and unknown
selection outcome.

Wire admission passed for 28 genuine responses and 12 request bodies: 23 HTTP
200, three 201 and two 400 responses. The parent visually accepted all four fresh
main frames and the final actor JPEG, including the corrected table labels.
These runtime/image results are parent-reported; this reviewer inspected source
and documentation but did not rerun the browser or independently inspect those
final images. The refreshed wording acceptance gate is closed on that reported
evidence. Final required-model confirmation remains pending and is not implied
by a browser or wire pass.

## Required-model confirmation

The parent confirms the independent required-Quint ledger passed in 3.97 seconds
(package 3.977 seconds), with zero skips: 11 observed facts became 11 model steps.
Three UI reads, one direct decimal check and one actor read account for five
physical SQL requests; typed parameters and read-only query identity matched.
Ten semantic mutants and an Apply-causes-source-read negative model control were
rejected. These results are attributed to the separate verification run and were
not rerun by this reviewer. All source, wording, runtime/image and required-model
gates are closed within the finite 389 scope.
