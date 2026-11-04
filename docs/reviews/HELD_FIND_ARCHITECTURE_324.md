# Local Find over held Activity observations — architecture 324

Status: final source review accepted; generated state checks pass.

The general-purpose outcome is to locate an interesting occurrence already held
in the Activity log/span buffer. Find is a local presentation operation, not a new
query/filter grammar, source population, completeness claim or evidence capture.
Keep the existing TailBuffer, keyed row nodes, TailSelection and stream owner.

## Deliberate search scope

Use a bounded, case-insensitive literal substring, matched independently against
explicit searchable fields. Reuse RecordPresentation.describe for summary/status/
severity and qualified service presentation, reported trace/span/parent IDs and,
configured RecordColumns field values through their
existing value/presentation helpers. Core human UTC/duration and valid reported IDs are included; state the actual scope beside the control.
Do not search raw JSON, hidden arbitrary attribute maps, action/menu labels,
membership tokens or preservation metadata. No regex, token grammar, locale
collation or concatenated cross-field matching is needed. Clarify whether UTC
is included; do not make search silently depend on DOM formatting.

Configured field changes recompute matching with the same text. Missing, null
and empty values contribute no implicit search term; literal strings such as Null
or Missing remain actual searchable strings. String values are searched directly,
not through presentation-added JSON quotes. Chosen timestamp/duration columns use
the human formatter rather than raw nanosecond epochs. Other present nonstring
values may use existing display presentation. The literal query is well-formed,
at most 256 UTF-16 units and 1,024 UTF-8 bytes, with deterministic lowercase and
no trimming, regex, token interpretation or field-to-field bridge.

## Projection and lifecycle

Hide/show retained keyed row DOM nodes; do not rebuild record components or prune
TailBuffer on each keystroke. Appends and evictions reconcile against the original
buffer and then apply the local projection. A hidden row's retained action handler
must refuse through the existing actionVisible/current checks. Snapshot.rows and
all source/selection/explanation consumers remain the complete loaded buffer,
not the matching subset.

Display matching/loaded counts explicitly. Observed and evicted counters retain
their existing connection meaning; none is a query total. Keep no-match local
state separate from no telemetry/no data/source unavailable. Find must neither
mutate selector predicates nor create a new tail request, abort, reconnect,
pause/resume, capture or saved Activity definition change.

Same-scope reconnect can clear the old buffer as before while retaining local
Find. Explicit new applied scope/signal/stage/start and close/reset should clear
Find deliberately, with code paths distinct from connection-reset buffer clearing.
Clear Find only clears its text/projection; it does not clear the buffer, selection,
query filters, history or preservation request.

## Selection, focus and source-effect hazards

TailSelection deliberately holds detached exact occurrence references independently
of TailBuffer. Selected rows may already be evicted; Find must not silently discard
hidden or evicted selections. Show selected-hidden/not-shown information and retain
Keep/expiry semantics. Clearing selection remains explicit. Existing dismiss focus
must not select a newly hidden row; use a visible current action or Find control.

The viewport currently pauses the source when scrolling away from the bottom, and
row keyboard focus pauses for inspection. Hiding rows/programmatic follow changes
must not accidentally trigger that source mutation. Find input focus stays outside
the row viewport, editing preserves its focus/caret, and matching projection does
not autoscroll the user's typing context. Existing row inspection still pauses and
returns to its own retained opener. Restore focus safely when a focused row ceases
to be visible or is evicted, without reviving stale row/service/Keep authority.

## Verification scope

Generated state tests can use existing TailBuffer/TailSelection and the pure local
matcher over append, duplicate append, Find edits, layout change, eviction,
selection and reset. Check an independent reference set, bounded rows/bytes,
exact selected occurrence retention and no record/query mutation. DOM/native tests
must establish stable row identity, keyboard/inspector/Keep continuity, honest
match/loaded/selected counts, no-match recovery, logs and spans, source-free Find
while streaming and explicit scope reset behavior. No new source executor or
browser-state framework is justified for this local feature.

## Implemented projection and state verification

The selected helpers live in existing RecordPresentation: findQuery, findValues
and matchesLoaded. They use bounded literal lowercase and actual present strings,
human UTC/duration, valid reported IDs and configured field values. Missing/null/
empty chosen values add no implicit terms; literal strings remain searchable
without presentation-added quotes. RecordColumns field actions now recheck actual
visibility before refinement, closing the hidden-row retained-action seam.

The tail owner hides existing keyed rows and leaves snapshot rows complete.
Matched/loaded reporting is separate from connection observed/evicted counters.
Selection references remain independent of visibility/eviction with an explicit
not-shown count; Clear selection uses a visible opener or Find fallback. Field
changes project before recomputing matches, and stream batches retain the existing
follow behavior. Local hide-induced scroll is fenced by its exact resulting
position rather than a broad timer that swallows user scrolling. Reconnect buffer
reset retains Find; explicit prepare/reset clears it.

`TestHegelHeldFindState324` passes 30 generated state cases in 3.112 seconds, zero
skips. Each executes 15–30 generated append/duplicate/Find/layout/select/signal/reset
commands plus a deterministic select→hide→evict prefix. An independent reference
checks loaded/matching keys, bounds/counters and detached selected occurrences.
Additional literal, hidden-token, cross-field, UTC/duration/ID, value-presence and
invalid-query controls pass. No DOM or source-cancellation result is inferred from
this pure state lane; browser/owner checks are separate in VALIDATION.

Final source review found a missing focus fallback when programmatic Find hides a
currently focused row. Capture the original row before projection and focus the
visible Find control with preventScroll if that row is hidden; ordinary typing,
column controls and inspector focus must not be displaced. This local correction
is pending final source inspection, not a need for another source owner.


Final source inspection accepts the focus correction: applyFind captures the
active row before projection and focuses Find only if that row becomes hidden,
belongs to this tail body and Find is actually visible. Other tail/column/inspector
focus is preserved and preventScroll avoids a new source effect. This is a local
DOM change after the passing pure state run; no matcher or state semantics changed.
No remaining architecture blocker was found. Actual browser stability, scroll
cancellation and telemetry budgets remain independently recorded in VALIDATION.

The final small lifecycle correction retains Find on manual Resume as well as
same-scope automatic reconnect; new prepare/applied scope/signal and close/reset
still clear it. Resume changes the observation window, not the local display
intent. A separate projection correction adds the actual visible compact UTC
string to the corpus. Full ISO nanosecond precision did not contain a copied
HH:mm:ss.SSSZ needle because of its additional fractional digits. The generated
fixture now includes nonzero submillisecond nanoseconds and derives the compact
needle independently using integer millisecond division plus Date formatting.
The final 30-case execution below supersedes the earlier focused time.

Final `TestHegelHeldFindState324` passes all 30 generated cases in 4.619 seconds,
zero skips, including the independent compact UTC/submillisecond control. Source
review of the lifecycle/projection corrections remains accepted.
