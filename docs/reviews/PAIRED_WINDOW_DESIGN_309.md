# Shared window for a saved fixed comparison — design review 309

Status: implementation source and all three actual browser frames accepted.
Read with [saved cohort design](SAVED_COHORT_DESIGN_308.md).

## User outcome

A responder reopens a useful saved comparison and asks the same question about a
new time window. Both fixed operands must move together. The task should not
require exporting YAML, editing two clocks and importing another definition.
This cut updates the existing owner/configuration; it adds no live subscription,
comparison editor or parameter framework.

## One contextual window control

Reuse the population's existing scope/window controls in the fixed-comparison
context. A compact **Comparison window** disclosure can contain UTC From/To and
**Stage window** where the existing controls cannot express a paired edit. Keep
it near the applied/configured scope, not above the live Activity stream or in a
new toolbar. State once that the window applies to both fixed questions.

Prefer existing explicit fresh-window behavior where applicable: resolve one
fresh end time and the existing duration once, then apply that exact range to
both operands. It must not compute separate now values or start a rolling query.
Do not add sliders, presets or a generic operand selector for this first task.

## Staged execution and continuity

Editing controls changes only a draft. Stage validates one half-open UTC range,
updates both checked questions, resets page offsets and retires both observed
operands. Retain every unrelated predicate, typed environment policy, grouping,
measure, record kind and qualified field. The baseline remains generically named;
a new shared window does not create a partition/complement proof.

Show a concise confirmation such as **Both question windows staged. Run to read
them independently.** Focus the existing Run population view control. Stage and
fresh-window configuration perform no telemetry read, comparison execution,
capture or publication. Explicit Run acquires the selected and baseline questions
independently and displays their actual read clocks/coverage, as in 308.

Reject invalid/reversed/out-of-budget ranges before mutating either question.
Preserve valid current configuration if validation fails; identify the error near
the controls. Pending record work, unapplied source/filter drafts and existing
owner guards must not be bypassed by a second window mutation path. Prevent a
stale/detached prior control from rebasing a different view.

## Preserve and reuse

Save stores the newly configured pair through the ordinary Preserve flow. Export,
CLI validation/import and reopening retain that same explicit fixed range. Run
is still required after reopen. Existing saved revisions are unchanged; the user
must explicitly save a new definition if they want the edited range preserved.

This does not enable copying a fixed pair into another service/environment via
Apply composition. That remains a separate refused operation until a checked
scope-binding transform exists. It also does not shift captured evidence or the
temporary held-parent comparison: those observations keep their original meaning.

## Acceptance as a real task

Reopen a saved fixed pair, stage a different explicit window and verify both
questions use identical bounds with all other scope/presentation intact. Stage
without source/write effects; Run independently and compare real observations
that differ between the old and new windows. Verify UTC clocks and record-page
counts remain legible.

Exercise the existing fresh-window path if supported, proving now resolves once
for the pair. Reject reversed/invalid ranges and stale controls without partially
changing either operand. Save and reopen the rebased pair with no read; export
and validate it using the same checked artifact flow. Preserve parent return and
pending Keep/draft behavior. Update the executable storyboard and actual compact
frames rather than creating a separate paired-window demo surface.

## Implementation source review

The source reuses the existing Expand time window disclosure as Comparison window
for the fixed-pair context. UTC labels include an exclusive end. Stage window
updates the checked pair; Discard window draft restores the admitted state.
Tracked first edits block Run/Save until resolved. The existing fresh-window and
timeline gestures route through the same paired transition, with fresh time
resolved once. No stream-entry toolbar, separate editor or new source owner was
added. Root documentation matches this interaction. Source UX review has no
blocker; actual frame review is recorded below.

## Next substantive outcome — declarative scope controls

After a saved comparison can move to a useful window, broaden the grammar toward
its original whole-interface purpose rather than adding another comparison
measure or window preset. Current portable compositions contribute displays;
the service/team's useful scope controls still live outside that declaration.

Recommend one categorical **facet control child** in the existing composition.
A team can declare a meaningful qualified field (for example HTTP route) beside
its breakdown and records. Reuse existing field/value discovery and typed
selection intents, the application's selection tray and explicit Run. The
component stays passive: no fetch, private query state or untracked filtering.
Selected values change the checked scope through the existing owner, while
unrelated predicates/environment/window remain intact.

The first cut can target an ordinary single-query view. Do not silently rebind a
fixed comparison's baseline or invent generic multi-operand parameter handling.
Keep missing/empty distinctions and count-source/coverage semantics; discovery
must not turn a supplied page into exhaustive facet values. Showing a contributed
control must not secretly execute a values query. Source acquisition stays an
explicit host action under the current policy.

Success is one real service/team view with its contextual facet, chart and records
saved/exported/imported together, then used to isolate a population through the
same staged scope/Run flow. This advances “the entire UI is expressible in the
grammar” with one useful control rather than a new editor, registry or collection
of specialized profiles. Design and architecture agree this is the next outcome
to assess after 309.


## Actual 309 visual acceptance

All three actual frames are accepted. The
[original fixed pair](../journeys/screenshots/309-original-fixed-pair.png) shows
three selected canaries against a 104-record baseline (100 stable, two missing,
two empty), with separate visible UTC read clocks and a shared 0–100 scale.
The [staged window](../journeys/screenshots/309-both-windows-staged.png) shows the
new checked UTC bounds and explicitly says both questions are staged for
independent Run. Run and Use fresh comparison window are the existing compact
actions; Preserve and View settings remain collapsed, with no stale comparison
presented as a result for the new range.

The [rebased comparison](../journeys/screenshots/309-rebased-fixed-pair.png) shows
five selected canaries against 105 baseline records (101 stable, two missing,
two empty), retaining the generic fixed-baseline meaning, separate read clocks
and common 0–101 scale. No extra editor or above-stream controls appear. No visual
blocker remains.

The verification owner reports native main assertions passed for invalid draft /
Discard, a genuinely reattached old form refusal, saved/reopened paired bounds
and fresh actual counts. Observed-model and campaign completion remain with the
verification report. The frames establish presentation, while read/write absence
and lifecycle correctness depend on those separate checks.
