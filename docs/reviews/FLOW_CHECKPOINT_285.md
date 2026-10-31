# Investigation-flow checkpoint — 285

> Delivery update: 285 subsequently passed the checks recorded in
> [VALIDATION](../../VALIDATION.md). The review below records its original
> source/screenshot scope and pre-validation status.

2026-10-31. Source and documentation review of the working tree through284, with
285 recent-population support in progress. **285 is pending final validation.**
This review did not run browser tests, capture new screenshots, trial vendor
products or conduct a user study. Executed shipment evidence remains in
[VALIDATION](../../VALIDATION.md). The recommendations reconcile the
[280 product](PRODUCT_CHECKPOINT_280.md) and
[architecture](ARCHITECTURE_CHECKPOINT_280.md) checkpoints with the delivered cuts.

## Product judgment

The first ordinary record investigation loop is now coherent enough to build on:
start with Activity, retain an exact fixed question, inspect records and aggregates,
select time, preserve a reusable definition or an evidence-backed explanation, and
reopen without an unexpected read. The remaining285 entry gate should finish this
record-scope cut. Do not immediately follow it with another broad selector sweep,
a generic component migration or measurement-package codec expansion.

The next meaningful outcome should be **“Make this signal's current investigation
reusable.”** The normal recent canvas can own two signal queries, while the authored
composition intentionally owns one. Signal tabs only change what is displayed;
they do not create a focused one-query authoring context. Close that reachability
gap before adding another authored chart capability. Then let the saved timeline
answer **“When did latency change?”** using existing exact interval measures. Neither
step needs a new evaluator or chart engine.

Preserve the stream-first entry. Authoring, source plans and custody details should
remain contextual disclosures, not additional controls above the first records.
The recent fixed canvas is an explicit alternative for a fresh aligned overview;
it should not replace the live monitor or silently start a subscription.

## What this review checked

| Flow | Current source-supported behavior | Remaining boundary |
| --- | --- | --- |
| Activity and fixed scope |283 preserves typed Activity selectors through publication, applied-scope discovery and staged reopen.284 preserves typed record steps through sharing, pinning and same-tab recovery. |285 removes the separate recent-population gate; its planner, labels and delayed responses still require final validation. Measurement and service/native scope families remain separate capabilities. |
| Authored composition |279–281 declare records, aggregate and an optional count timeline under one fixed query. Paging holds aggregate/time inputs; interval keys stage a shared exact window. |`TimelineViewDeclaration` and `timeline-declaration.mjs` admit count only, even when `timeDataset` has duration/numeric cells. No authored comparison child or rolling subscription grammar is claimed. |
| Reuse |280 applies a compatible whole composition while retaining the destination query and held page. Width, signal, grouping and quantity compatibility remain explicit. |Recent signal tabs do not reduce a two-query owner to one; whole-composition reuse and authoring are consequently not directly reachable from the usual two-signal canvas. |
| Preservation |282 pins an exact saved-view ID and finding;284 pins an exact question step. Existing captures and retained-evidence cards keep independent custody. |Save reusable question, Save exploration view, Add question, Add saved view and Capture evidence remain separate entry controls whose distinctions users must learn. |
| Inspection and comparison |Existing record/trace components and ordinary distribution/summary comparisons support investigation. |Do not build another waterfall or imply the authored stack can already save every specialized baseline/native comparison. |

Reviewed anchors include `ui/stream-entry.mjs`, `ui/exploration.mjs`,
`ui/population-workspace.mjs`, `ui/documents.mjs`, `ui/question-protocol.mjs`,
`ui/count-timeline.mjs`, `ui/scoped-breakdown.mjs`,
`workbench/view_timeline.go`, `workbench/steps.go`,
`exploration/view_dataset.go`, `exploration/view_measures.go`, and the281–284
storyboards/contracts. This is a cross-flow review, not a claim to have audited all
repository features or production deployment controls.

## Prioritized next cuts

### 1. Focus one recent signal into the existing authored view

Add a contextual **Customize this signal** or equivalent action to the selected
recent signal, without another top-level entry or new whole-screen layout model.
It should stage a one-query child retaining that signal's exact typed environment,
predicates, grouping, measures and fixed window. Keep the original multi-signal
parent paused and returnable. The action must not silently delete the companion
from the parent, reset its filters or import old action authority.

For an eligible ordinary grouped query, the child can use the existing bounded
records-plus-aggregate declaration. Reuse the current record column choices and
checked display defaults; do not force histogram/native/multi-field profiles into
that contract. Use the root query at offset zero, not the currently paged row
request. A companion often has no grouping: in that case stage the exact focused
query first and let the user choose grouping through the existing editor before
composing. Never silently add `groupBy: service` while claiming query preservation;
any default grouping proposal is a disclosed staged query change requiring Run. A source-free authored configuration can be staged without
pretending the child has newly admitted evidence. Run explicitly acquires its
inputs, after which normal timeline configuration, save/reopen and whole-view
reuse are reachable. If a current display/selection cannot be carried, explain the
specific limitation before changing context.

This follows concrete code boundaries: recent signal buttons only update
`selectedSignal`; `mountCompositionEditor` requires exactly one query plus a
current authored display; range/facet drilldowns retain the existing query list.
The Reusable display checkbox can author each signal independently but cannot
make a two-query stack compatible with the one-query composition. Do not solve
this by weakening the composition contract or adding hidden per-child queries.

Acceptance: ordinary two-signal recent canvas with independent predicates → focus
spans → staged single-query view → explicit Run → save a records/aggregate/time
composition → reopen it → return to the unchanged paused parent. Check source
request accounting, exact scope and filter preservation, field-column validation,
no stale actions, and explicit handling of pending record selections. Unrelated
investigation notes must remain intact; focusing a signal is not permission to
replace their draft.
This is the recommended immediate next feature.

### 2. Saved duration timeline beside records

Extend the existing temporal declaration to query-supported duration measures,
starting with count, mean and p95. The parent query remains the sole authority for
`summary: span-duration`, fixed bounds and bucket width. Selecting a display
measure over already supplied cells must be local; enabling a missing source
measure stages the query and requires explicit Run.

Reuse the canonical exact value/support/method handling from263 and the existing
passive temporal display path. `timeViewDataset` already appends
`ViewMeasureFields` and per-cell support, while `ScopedBreakdown` already exposes
query-supported interval measures. Replace the count-specific assumptions in the
small authored presenter rather than inventing scalar percentile calculations.
Keep H1 range ownership, unknown-gap refusal, complete-interval checks and held
paging unchanged. Display duration in human units with exact details available.
Do not average p95 values, label counts as rates, or use record-page coverage to
qualify interval measures. Numeric measures can follow the same profile once the
duration journey proves it; signed values need the existing safe table fallback.

Acceptance: uneven real span traffic with a short latency burst → saved
records/aggregate/timeline composition → count/p95 switch on held inputs → select
burst interval → explicit Run → narrower records and latency → save/reopen/reuse.
Check exact quantile method/support, no-support intervals, units, independent
clocks and source-request counts. This follows the main-flow authoring entry above.

### 3. One contextual preservation entry, retaining distinct objects

After the temporal measure cut, consolidate the visible preservation choices in
the existing Explore/investigation flow. A compact contextual Preserve control can
offer **Save view**, **Add to investigation** and **Capture evidence**, with a short
object-specific explanation only when the user chooses an action. Keep existing
APIs, document cards, recovery and immutable IDs; do not introduce a new universal
saved-object store or a second destination picker.

Choose the object from context: a question card preserves query intent; a view card
preserves its authored representation; a capture preserves independently acquired
evidence. Do not silently autosave, recapture or replace a pending note. Existing
controls may remain internally specialized while their entry and status feedback
become consistent. The goal is fewer decisions before the user can preserve a
finding, not eliminating honest differences between definitions and evidence.

Acceptance: one scoped question can be preserved with or without authored layout,
and with separately requested evidence, using one discoverable entry. Every
resulting card clearly identifies what reopening does. Cover pending-note/recovery
and stale-save decisions in the existing owner tests rather than creating a new
state framework.

### Later: compare a saved scoped finding with another checked population

Then design one authored comparison profile around existing comparison operands
and custody, initially a fixed record population against a separately identified
baseline. Make the user outcome explicit: distinguish a suspect population from
its prior window or rest-of-parent population without losing its selected view.
Reuse the existing comparison math and exact scope checks. Decide whether the
baseline is a held capture reference or a query to run before changing the wire;
those are different acquisition semantics. Do not add an opaque baseline field to
a passive child or combine independently paged rows into an atomic capture.

This needs a small product/contract proposal before implementation. It is more
useful than broad template proliferation or a layout editor, but is not a blocker
for the duration timeline above.

## Simplification and architectural boundaries

- Keep one application query owner per fixed composition. `exploration.Compile`
  owns semantics, source admission owns supplied inputs, H1 owns action authority,
  and passive components own rendering. A saved declaration grants no read rights.
- Reuse one checked record environment/question protocol across handoffs. Further
  field-by-field migrations should be driven by a real blocked journey, not by
  removing every capability gate at once.
- Extract a small population transition/preservation helper only when the next
  feature would duplicate existing draft, owner and acknowledgement rules. The
  large `population-workspace.mjs` is a maintenance pressure, but a registry or
  controller rewrite before a user outcome would repeat the earlier migration
  spiral.
- Keep live Start/Pause, Show recent activity, fixed Run and definition reopen
  explicit in their effect. The recent action authorizes a fresh fixed plan and
  reads; opening configuration stages; a live connection has its own rolling
  window. These differences should appear in compact status/verbs, not repeated
  instructional paragraphs across every child.
- Keep source/aspect coverage separate from health. Missing support is not zero,
  captured zero is not healthy, and equal environment/time does not establish
  identity, causality or an atomic read.

## Process lessons and deferrals

The recent checks exposed test-driver assumptions about hidden Edit analysis
controls, ordinary Go omission of empty arrays and SSE timing completion. Preserve
those distinctions in executable journeys: open the actual controls, compare known
wire aliases narrowly, and independently observe completion after the stream closes.
A tail-connection count is not a count of internal SQL polls. Keep compact request
ledgers, model traces and documentation frames; bulky successful campaign copies
are disposable.

Close285 with its real typed-scope journey and focused regressions, then advance the
focused-signal authoring outcome above. Defer arbitrary grids/nesting, a new chart vendor,
marketplaces, full WYSIWYG, multiplayer and broad native/package scope adoption.
Production identity/source isolation, quotas/retention and durable alarms remain
separate product workstreams; this flow review does not establish their readiness.

## Visual follow-up on the285 journey

Viewed the actual `285-typed-recent-population.png` and
`285-selected-population.png` frames produced by the parent run. This is screenshot
inspection, not another browser execution; final285 validation is still pending.
The fixture has one displayed record per signal, so its empty main area does not
establish a general high-volume density problem.

1. The checked current/legacy policy, half-open UTC window, signal tabs and record
   context are visible and readable. The selected-window frame preserves that
   hierarchy. No material visual blocker to the scope handoff is apparent.
2. The expanded Fields and facets sidebar runs much farther down than the single
   record, with large actions, repeated explanatory text and a small one-value
   distribution. Prefer a compact collapsed summary until a facet interaction needs
   it, and shorten state-irrelevant prompts. This is useful accompanying polish;
   it does not replace the focused-signal authoring task. Put Customize this signal
   contextually near the active signal/records rather than adding another open
   control section above the data.
3. The duration overview repeats question marks across unsupported intervals.
   Preserve those unknowns, but consider a quieter contiguous unknown band/status
   with exact interval detail on inspection. Do not replace absent support with
   zero or infer a broad chart redesign from this sparse fixture.
