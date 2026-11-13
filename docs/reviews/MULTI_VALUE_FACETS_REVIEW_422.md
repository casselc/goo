# Multi-value counted facet review — 422

Status: bounded source contract and all five actual frames accepted. Execution
and model results remain recorded separately in validation. This reviewer owns no production changes and runs no
heavy checks.

## Product and semantic boundary

The existing counted result describes one qualified field in a fixed applied
scope. Multiple choices should mean OR within that field, AND with the retained
scope. A prior predicate on the same field remains part of that scope; this is
subset selection, not an implicit replacement or broadening operation.

Selection, local search and Clear are local drafts. One explicit Stage commits
a validated predicate atomically. Apply acquires a new Activity observation;
selection must not refresh source counts, compute a new total or grant coverage.
Counts keep their original window and read basis until retired.

The review checks these boundaries:

- Bind the ready result's field to its sole question grouping and signal. Preserve
  resource/attribute namespace, missing keys, present empty and literal lookalikes.
- Compare the editor's current predicate with the counted request's original
  predicate. A changed local draft cannot silently become the merge basis.
- Check current owner/read leases on checkbox, Clear and Stage, even before a
  later sync removes stale controls. Replacing data retires the old selection.
- Validate the whole prospective AST, including existing predicate nodes/depth
  and UTF-8 value bounds, before changing the tree or emitting onChange.
- Keep selections across loaded-value search and show the total selected count,
  including hidden choices. Unsupported values remain visible but disabled.
- Preserve native keyboard, disabled/inert/closed-action refusal and local return
  behavior; refused checkbox events restore the authoritative checked state.

## Source acceptance

The pure helper validates each exact string/missing value and the complete
combined AST before returning a detached candidate. It flattens an existing AND
without dropping its clauses, preserves existing same-field restrictions, and
uses an OR only for multiple chosen values. Empty and missing remain separate.
No caller input is mutated on success or refusal.

The source owner clones the original counted question into ready data and
supplies the current generation/scope lease separately from that serializable
result. The renderer checks the sole grouping field, signal, canonical original
predicate and live lease. Dirty manual edits, data replacement and owner
withdrawal retire selection authority. Selection comes only from the copied
result's items; Clear and loaded search do not stage or read. Hidden choices are
retained and explicitly counted. Stage validates again before one tree assignment
and one callback handoff. Old callbacks cannot reuse a retired result.

The shared UI uses native checkboxes, an explicit Stage action, a disabled empty
selection state and role=status feedback. A refused checkbox event restores the
stored checked state. Each checkbox describes its exact original-scope count;
Stage describes the selected-count/status text. Component-owned styles retain
native checkbox geometry, wrapping labels and aligned tabular counts. Actual
mobile containment, keyboard focus and source/read effects remain live gates.

No source blocker was found in this bounded inspection. It is not a whole-product
accessibility or concurrency certification.

## Actual-frame acceptance

Independently inspected the selected-state 1440/390/320 pixel frames, filtered
390 pixel frame and applied 390 pixel frame from the running app. No visual
blocker was found.

The two selected choices are missing key and present empty string. The literal
string `"Missing key"` stays separate and unchecked. Native checkboxes, clear
labels, aligned counts and the explicit Stage/Clear actions remain readable at
all widths; the 320 pixel action stack stays inside its owner. The original
261-record scope and five value counts remain visible while two choices are
selected, without an invented selected-population total.

The filtered frame correctly says `2 selected · 2 hidden by search` while only
the unselected canary is shown. Stage therefore remains an explicit handoff of
the retained choices rather than silently substituting the visible result. The
applied frame shows the retained predicate AND `(attribute cohort = "" OR
attribute cohort is missing)`, followed by two loaded records. It does not retain
old counted values as if they described the changed scope.

Source inspection confirms each checkbox's original-scope count description and
Stage's selected-state description. Visible focus is clear on the loaded-value
search. Exact focus identity, zero-read selection/staging, real predicate outcome
and stale-handler refusal are browser assertions; screenshots alone do not prove
those properties. This bounded Chromium/source review is not screen-reader,
cross-browser or whole-product accessibility certification.

## Candidate next consolidation

A bounded 423 follow-up can reconcile render-time availability with fresh
invocation state and reduce repeated retained-row snapshots during inspection.
`stream-entry` currently calls `tail.snapshot()` inside `groupable()`, used by
per-field availability callbacks. This is a profiling hypothesis, not a measured
performance claim. Measure cloning/allocation with representative retained
records, then share one snapshot per render where useful while preserving fresh
invocation leases. Retain first/second-inspection, hidden initial mount, changed
scope and return-focus conformance checks. No global action registry, broad
component rewrite or source-engine change is needed.
