# Authored cohort comparison architecture — 307

Status: bounded contract recommendation; implementation review pending. The task is
select a population in an authored count breakdown and compare it with its parent
or the remainder, using existing query/comparison and retained-child owners.

## Existing seam and missing capability

`partitionQueries` already builds selected and complement predicates over the
original parent scope, preserving typed environment and fixed bounds and resetting
page offsets. It distinguishes missing keys from present values and enforces the
predicate budget. Native selection comparison uses `Distributions.snapshot`,
`setBaseline`, the current exploration executor and existing Compare rendering.
The rest action stages a child with a baseline question for explicit independent
reads. Authored charts currently disable both actions, and their adapter's
`setBaseline`/`select` methods are deliberately no-ops. Removing the authored
eligibility check alone would expose a broken operation.

## Bounded implementation outcome

Support ordinary authored **count** Bars/Table over one explicit grouping, backed
by a genuinely admitted full aggregate snapshot. Keep numeric/histogram, duration,
reported-edge and special profiles outside this first task rather than silently
comparing their records as count. Record-only narrow data cannot be promoted to an
aggregate authority just because it has a count field. A changed-between-reads
result, pending edits/pages/record actions or an inactive owner remains unavailable.

Open the existing retained population child for authored comparison, leaving the
parent question, authored AST, template and held results unchanged. Reuse existing
native count comparison presentation/executor rather than teaching the passive
ScopedBreakdown renderer to fetch or creating a second captured-comparison system.
A child can map compatible count presentation defaults; it does not need a new
persisted authored cohort declaration for this task.

For **selection versus parent**, retain the admitted full parent aggregate as an
in-memory baseline through the existing snapshot/setBaseline semantics. Explicit
Run reads the selected question; the parent side remains that held observation,
with its captured clock and basis. The selected and parent populations overlap;
this is not a partition or a claim that counts sum to the parent.

For **selection versus rest**, reuse `partitionQueries` and the existing child
baseline-question configuration. Staging performs no read. Explicit Run reads
selected and remainder independently under the original parent predicates,
environment and bounds. The predicates are disjoint by construction; independent
source reads do not establish a simultaneous exhaustive physical snapshot or
complete upstream coverage. Preserve null/missing versus present empty values.

## Ownership and custody

Any held-baseline handoff is a narrow application-only input to the current child,
not a stored view field, source import endpoint or portable artifact clause. Check
its exact question, admitted aggregate basis and originating owner before use.
Do not copy a row page into a fabricated dataset, synthesize source witnesses or
create captures merely to make the comparison API accept an operand. Save may
store baseline question intent using existing rules; held data itself is not saved
or exported as a definition. Make any later rerun behavior explicit rather than
silently claiming the old observation survives a reconstructed saved view.

The parent/selection lease must capture generation, held result and selection,
and stale/reinserted controls must not operate on a replacement parent. Existing
child depth, pending source/filter/record/page, inactive and live-mode restrictions
remain. Back restores the paused parent and the originating action predictably,
without a new source read.

## Acceptance

Start with the actual 304 authored service breakdown, select one or more values,
then exercise parent and remainder comparisons. Prove source-free staging, exact
predicate/environment/window preservation, one selected read with held parent vs
two independent reads for remainder, genuine aggregate operands and truthful
labels. Return to the unchanged authored parent and save/reuse its declaration.
Include missing, empty and multi-value selections, stale selection/control,
narrow-page or changed-state refusal and read failure. Reuse existing property,
observed lifecycle and real browser storyboards rather than fixture-only UI data.

## Implementation seam refinement

The selected child may preserve the authored composition and records kind instead
of downgrading to a flat native view. An ephemeral application adapter in the
existing child/executor can hold the parent operand or fixed remainder question,
read full selected/remainder results through the existing executor, then use the
existing `/compare` admission and shared measure/group comparison components.
This refinement preserves authoring continuity without adding a persisted cohort
AST or a standalone comparison workbench. It must not use authored adapter no-op
baseline methods or alter the captured-baseline owner.

If the transient pair and remainder query are not persisted, Save/export must say
that they preserve selected view configuration only, with comparison omitted.
Reopening cannot claim to restore the ephemeral pair. Initially refuse compositions
already containing a captured comparison to avoid ambiguous competing comparison
preparation/shared-pending semantics; expand that interaction only deliberately.

## Supplied group profile compatibility

Implementation inspection found a real existing API gap: optional group-count
comparison reused `CompatibleCapturedCountQueries`, whose exact predicate rule is
correct for captured-root reuse but excludes selected/parent/remainder cohorts.
The browser group validator inherited the same constraint. Keeping the planned UI
while dropping its group frames would leave the core comparison task incomplete.

Add a dedicated ordinary count-cohort compatibility helper for this supplied
profile. Validate both ordinary single-group count queries, preserve exact typed
environment, signal, grouping and remaining profile directives, and disregard
predicate differences plus the already documented window/delivery normalization.
Use it only for optional group-frame generation/admission. Keep captured-count
root/template compatibility unchanged, including its exact-predicate requirement.
No request flag or additional source authority is necessary: the optional profile
is a checked rendering of supplied operands, not a claim of shared root or disjoint
partition. UI provenance/overlap labels remain application responsibilities.

Update the older optional-profile test: a filter-only difference now legitimately
produces group comparison, while a genuine unsupported profile difference must
still leave broad comparison available without a group frame. Retain malformed
group/claimed-accounting refusal and add direct Go/browser parity controls.

## Final implementation source review

The selected-count module requires a private admitted ranked aggregate snapshot,
its exhaustive groups witness, exact query context, source basis and captured clock.
Normal changed-between-reads results do not receive that witness from the backend,
so they cannot become exact held operands. Narrow pages and cloned/fabricated
snapshot objects are refused. The plan freezes selected/parent/rest question intent
and uses supplied count/group cohorts only; no rows or source snapshot are sent as
an execution instruction.

The authored child retains composition and record options. Explicit Run reads the
selected scope, then the existing count-controller pipeline reads remainder only
when required and imports `/compare` through existing admission. Generation,
serial, intent and current-result guards fence completions. Scope/profile changes,
including adding a captured comparison, retire the context; compatible presentation
changes remain local. Save/export omit transient operands/pair, with visible notes.

Authored selection actions capture generation, draft revision and selection value;
current field edits rebuild the tray so old actions cannot acquire a new selection.
The special Back path renews authored/timeline actions and reissues current tray
controls, resolving the same parent/rest action for focus rather than leaving only
retired handlers. The parent remains held and unchanged. Source review has no
remaining architecture blockers; actual browser/model/API outcomes belong in the
validation record.

## Recommended next outcome: save the remainder relationship (308)

The next concrete authoring gap is durable selected-versus-rest intent. In 307 the
user can discover a useful contrast but Save/export deliberately retain only the
selected view. That is honest bounded behavior, yet it loses the central question
when a team wants to reopen or share the investigation. Prefer addressing that task
before another visualization or compatibility profile.

Start with remainder comparisons only. Persist two fixed question intents through
one bounded fixed-query operand variant of the existing count-comparison role and
executor. Reuse current ordinary count cohort compatibility, exact typed
environment/window/group/profile validation and explicit independent reads. Do not
allow composition plus legacy flat breakdown state to become competing owners.
This is a deliberate query-plan input for one existing role, not a generic recipe,
dataset-role framework or new registry. Save and portable artifacts contain
questions, never rows, witnesses or source basis. Reopen stages them and explicit
Run acquires both fresh observations.

Held-parent comparisons must stay ephemeral in this cut: saving a parent query and
silently re-reading it would change a comparison against the held observation into
a different question. A future explicit capture or fresh-baseline conversion can
make that distinction visible; it should not be an automatic Save side effect.

Reusing presentation into another current scope is a separate concern. A portable
query-pair file can be imported and Open can stage its original questions, but
ordinary Apply must not copy the source baseline predicate/environment/window into
a destination. Initially refuse query-pair presentation reuse or require exact
scope equivalence; defer cross-scope rebinding until there is an explicit parameter
or transform contract. A real journey should select an anomalous group, compare
with remainder, Save/export, import/reopen, Run both scopes and recover the same
question relationship with new read clocks.
