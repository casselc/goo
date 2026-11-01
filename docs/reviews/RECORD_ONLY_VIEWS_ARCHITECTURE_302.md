# Records-only views — architecture review 302

Source checkpoint: `f98e05e` (301). This review defines the accepted 302 boundary;
executed checks belong in VALIDATION.md rather than being inferred from source.

## Decision and scope

Add exactly one new composition form: one visible records child. Existing
aggregate-plus-records compositions with optional timeline/comparison retain their
current two-to-four-role rules. A records child plus timeline or comparison without
an aggregate remains outside this cut. Do not implement this as an aggregate that
happens to be hidden: the absent aspect must actually be absent from the declared
view and its acquisition plan.

The new form is for ordinary raw logs and spans. Table and List apply to either;
Span waterfall remains spans-only, including hidden/imported declarations. Keep
UTC/summary requirements, all eight qualified fields, role uniqueness, pinned kind
versions, visibility and strict closed-object validation. The new single child
must be visible. It retains the existing `goo.composition/1` stack and
`goo.record-page/1`; no source, store, pager or display kind is added.

The selected first-cut policy rejects numeric bindings, numeric histograms,
reported-edge bindings and grouped time-series on the new record-only declaration,
consistently in Go, JS and OpenAPI. It never silently removes those directives.
Existing aggregate-based forms keep their current support. Valid summary/status,
bucket width and absent/one/multiple ordinary grouping fields may remain in the
question, but the record-only acquisition supplies no aggregate/timeline aspects
for those directives. The question, bounds, predicates, environment policy and
delivery settings remain exact.

## Validator dependencies that must become explicit

Two existing invariants were previously inherited from the mandatory aggregate:

1. `ViewComposition.Validate` reached `exploration.Compile` through aggregate
   validation. Add explicit query compilation at the composition boundary so a
   directly validated/reused record-only composition cannot bypass bounds,
   predicate, field or window checks. A records declaration alone validates only
   its own scope/kind/columns; it is not the query compiler.
2. `ViewTemplateBinding.Validate` relied on composition validation to require an
   aggregate. Preserve the authored template's destination-category contract with
   an explicit aggregate-and-records requirement in Go and the OpenAPI template
   branch. JS `templateView` already requires both. A grouped record-only view must
   not accidentally become a category template merely because `groupBy` has one
   field. Ordinary saved-view reuse is the supported path in 302.

Schema conditions must admit one records child or the established aggregate forms,
not every arbitrary one-child stack. Test hidden-only, duplicate, missing-records,
nonaggregate timeline/comparison and metrics cases as refusals. Keep wire null and
unknown-field admission as strict as current handlers.

## Acquisition and source authority

`Source.FetchRecordPage` already reads independent plain count and bounded rows;
its record response intentionally omits Fields, Groups and all aggregate, summary,
histogram and time datasets. The application should use this existing profile for
record-only Run and pages when admitted, with `SourceSnapshots.adoptRecordPage`,
instead of manufacturing a full exploration wrapper. This preserves server-held
selection custody and the same exact result token/question/basis mapping.

The existing profile internally suppresses unused aggregate work but compiles the
original question for row projection. Do not clear grouping or other retained
question fields in the application to obtain the optimization: that would change
stored scope and can remove derived row columns. The underlying page profile is
more capable than the intentionally bounded 302 declaration profile; that is not
permission to silently broaden the first-cut public contract.

Measure actual source statements with the existing SQL recorder for ungrouped
logs, exact-trace spans and a valid multi-group question. The expected narrow
profile is count plus rows (two statements), with no group, time or inventory
statement. This is an acquisition-shape claim, not production latency or complete
upstream coverage. No source endpoint or new query evaluator is necessary.

Paging must choose the effective record result, token, question, source basis and
clock for inspection/publication. The configured root remains page zero for saved
views. With an aggregate composition, sibling aggregate/timeline results remain
held; a records-only view has no such sibling to synthesize. Its record count is
still an independent read, never inferred from the displayed row count.

## Native entry and saved reuse

Customize an ungrouped or multi-group logs/spans scope into this record-only form
without first adding a dummy grouping. Reuse the existing selected-signal fixed
child and application controls rather than introducing a second raw-record editor.
The record kind/field controls, current pending-selection guards, Inspect/Keep and
explicit cross-service Expand trace remain the existing owners.

A saved record-only composition has no aggregate display declarations. The current
catalog returns early when `Composition.declarations(view)` is empty; move that
return after the composition Apply branch, otherwise valid saved record-only views
remain unreachable. Reuse validates the source and target, keeps the destination
query and supplied data, and transfers presentation only. Same-signal and field
validation remain mandatory; a spans waterfall cannot be applied to logs.

Do not fabricate a missing aggregate from a record-page count. If a user applies an
aggregate-bearing composition to a record-only result with an otherwise compatible
question, its newly needed aspect must become explicitly unavailable/staged for
Run rather than crashing or receiving a witness from record data. The inverse
transition can reuse a previously admitted record aspect locally without claiming
that its old aggregate remains part of the new view.

## Acceptance and review

Use real OTel logs through ungrouped Activity → Customize → List/qualified fields
→ Inspect/Keep → native Save/reopen/Run → Apply composition to another ungrouped
log scope. Prove no grouping is added and no source query or rows are copied from
the saved specimen. Use an exact-trace spans scope with grouping absent or unchanged
→ Waterfall → Keep/page/save; the 301 original-index and parent semantics persist.

Assert declaration absence and actual response absence of aggregates, not merely
hidden DOM. Cover preserved typed environment policy, summary/status/bucket and
multi-group query metadata; malformed specialized profile and category-template
admission; stale/reinserted controls; pending record workflows; held return and
explicit acquisition. Extend actual Go/JS property checks and observed lifecycle
replay with negative controls. Existing grouped, timeline, comparison and authored
category-template journeys are required regressions.

A later Trace inspector → Open as view handoff can reuse this foundation. Existing
trace findings already retain both population and expanded trace scopes; a second
trace-save wrapper is not the core gap being solved here.

## Implementation source review

Reviewed the 302 Go declaration/template changes, generated OpenAPI builder,
JavaScript composition validator and workspace routing against the contract above.
The explicit query compilation and category-template aggregate requirements are
present. Raw declarations reject specialized profiles while retaining ordinary
question metadata. The source-profile test records exactly count plus rows for
ungrouped logs, exact-trace spans and retained multi-group/time metadata, and
checks that unacquired aspects remain absent. These are source-review findings;
executed check results belong in `VALIDATION.md`.

The workspace routes record-only Run through the existing page endpoint and
admission function, replaces the effective record result after paging, and retains
ordinary catalog Apply despite an empty aggregate-declaration collection. The
shared predicate validator also validates raw predicates. Rendering an aggregate
or timeline after applying a richer composition to held narrow data reports that
the aspect was not acquired and requires explicit Run.

One review finding requires a command boundary as well as an honest placeholder:
comparison preparation must reject a held narrow record page until Run obtains
the full aggregate aspect. A count field alone is not the existing comparison
operand authority. The UI must disable Prepare and its command must independently
guard the missing aspect; stored-baseline preparation must not quietly promote
narrow count/rows into a full population observation.

The comparison finding is resolved in source: `countCurrent` now requires an
admitted aggregate snapshot, `countCommand` independently requires that operand,
and both initial and synchronized Prepare controls disable when it is absent.
The missing-aggregate status directs the user to explicit Run. Final architecture
source review has no remaining blockers. `git diff --check` passed; browser and
model execution are recorded by the implementation/verification owners.

## Recommended next outcome: trace inspector handoff (303)

302 verification now includes the native 19-step journey, observed model replay,
and the inverse Apply/Prepare negative control, as reported by the implementation
and verification owners. Final source review preserves the acquisition and
comparison boundaries above; no further architecture blocker was found.

Add **Open as view** to the trace inspector, staging an ordinary records-only
span waterfall through the existing workspace owner. `ui/record-results.mjs`
already exports `traceQuestion(origin, traceID, spanID, signal, offset)`: it retains
typed environment policy and the original fixed bounds while deliberately replacing
search predicates with exact trace identity. Use the exact trace question at offset
zero, not the currently paged `tracePage.question`. The inspector's extra grouping,
summary and status metadata need not be added merely to open a raw waterfall.

The missing piece is an application-owned callback from the inspector to the
workspace's existing stage/child flow, carrying the validated question, ordinary
records presentation and opener. The passive waterfall remains unchanged. Staging
must perform no source read, copy no held rows/basis/result token, and require
explicit Run through the 302 page profile. This creates a configurable, saveable
view; the existing `TraceJourney.finding` continues to own separately captured
population/expanded-trace evidence without a second trace finding wrapper.

Guard the handoff with the active trace-context generation, visible dialog and
current application owner. Suspend or close the inspector before entering its
child workspace, retain predictable Back/focus behavior, and retire reinserted
controls after context changes. Cover exact trace identity, typed environment,
unchanged fixed bounds, page-offset reset, no acquisition until Run, native
save/reopen, and stale opener rejection. Prefer this small owner seam over a new
URL payload protocol or source-snapshot import path; those add authority and
lifecycle mechanisms that this outcome does not require.
