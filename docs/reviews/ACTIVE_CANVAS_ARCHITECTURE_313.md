# Active canvas consolidation — architecture 313

Status: architecture source review accepted; executed checks are tracked in
`VALIDATION.md`. Consolidate
visible controls and return behavior around the current canvas without merging the
Activity, fixed-question, population, retained-child or trace execution owners.

## Visibility and local return

When a population or retained child owns the visit, hide dormant root question
controls using an application-shell visibility state. Preserve their DOM, drafts,
query history and suspended owners; do not rebuild them merely to conceal controls.
Do not hide a root-only editor unless the active canvas supplies a supported local
editor or a clear route back to it. Read-only contexts retain their existing limits.

Expose one primary local return for the current child/population. Returning is a
local owner transition with zero source reads; it restores held data and predictable
focus, without silently applying drafts or starting Activity. The preexisting
paused Activity anchor remains the Activity owner's responsibility. A secondary
service/Activity context link may remain, but do not introduce several competing
Back actions or a new navigation registry.

Scope actions, saved-view catalog and Preserve continue to use the active
application owner. CSS visibility alone does not authorize a detached control;
existing handler readiness/current-generation checks remain. No new executor or
universal state container is needed. Use a small adapter only where the actual
root/canvas visibility role is shared by real callers.

## Titles and capability explanation

An authored saved/display title is literal. Do not guess whether a title is a
replaceable default by matching words or prefixes. A computed default needs an
explicit declaration/default contract, or may appear as a separate nonpersisted
scope subtitle. Actual service/category/group context should be visible without
rewriting the user's title or mutating their saved AST.

For an unavailable cohort action, display the actual bounded reason near that
action lane (for example span-status projection is outside ordinary count cohort
support). Do not silently clear measures, guess a different profile or replace all
specific reasons with a generic unavailable banner. A user can deliberately edit
scope/profile through the existing appropriate owner.

## Trace read cancellation

The trace inspector owns its independent context and read generations. Close,
replacement trace, context leave and hide must invalidate old actions/results and
cancel outstanding source fetches through its existing fetch adapter signal seam.
Do not abort the parent question merely to cancel the trace. Apply the same
ownership to trace pages/linked logs or surrounding context if sharing that read
owner, so replacement requests cannot waste work or publish stale results.

Cancellation prevents further UI publication but does not prove a source never ran.
Keep normal close/Activity-return callbacks distinct from trace-to-view handoff,
which stages configuration while preserving the paused return owner. Reopening or
Back must not automatically reacquire canceled data.

## Acceptance and review

Exercise Activity → service Customize → facet/cohort → trace → staged view and
Back. Assert one active editor/canvas, root draft/history retention, exact scope and
literal titles, truthful capability reasons, zero-read local return and preserved
paused Activity. Start trace/linked-log work then close/replace/leave; assert request
abort and stale result/action refusal without parent cancellation. Inspect keyboard
focus and narrow-width layouts in actual browser frames. Keep source review and
executed browser/model/request checks explicit in validation.


## Settled title/capability contract

The new shipped HTTP aggregate omits its display title, deliberately using the
existing renderer's category/measure default. The view/catalog title and every
explicit authored display title stay literal and byte-for-byte unchanged. Show a
separate compact qualified group/measure context even beside a literal title. No
new title metadata language or heuristic replacement is introduced.

A common pure count-cohort `eligibilityReason` returns the same specific projection
restriction used by selection validation; the tray displays that reason near the
unsupported comparison actions. The local population return is labeled **Edit
original question**, keeping its existing source-free owner transition.

## Source review: visibility, capability and cancellation

The visibility implementation retains original DOM and records inert state only
on first canvas entry. A visit revision fences queued focus restoration, so a later
entry cannot be disturbed by an older return callback. Nested canvas transitions
keep the root hidden; local exit restores its prior inert state and visible saved
focus or root Run fallback without executing a query.

The shared pure count eligibility reason preserves existing admission policy,
while the tray shows the first actual profile/value/comparison restriction beside
unavailable actions. Current command readiness remains unchanged. This is an
explanation improvement, not automatic profile repair.

Trace, linked-log and surrounding reads use three mount-local controller slots
forwarded through existing fetch adapters. Replacement/close/hide/dispose abort
owned requests, generation checks still fence results, and `throwIfAborted` refuses
late fulfillment from an uncooperative adapter. Finally clears only its matching
controller. Abort errors remain silent, and a retired trace cannot start its
follow-up log read. The parent query controller is untouched. Source review accepts
these lanes with no architecture blocker; actual execution belongs in validation.


## Final title and active-owner review

The frozen visibility lane also prevents dormant root UI events from mutating
source state while leaving public navigation methods intact. Active-canvas-aware
focus fallback avoids restoring focus into hidden root controls. Root drafts and
DOM remain retained; no query/session/executor replacement was introduced.

`ScopedBreakdown.boundContext` preserves every explicit nonempty title literally
and derives a default only for absent/empty display title. It always supplies a
separate qualified group/measure context. Only the new shipped HTTP aggregate's
default title was omitted; view/catalog and existing authored titles remain
unchanged. Final source review accepts all 313 lanes with no remaining architecture
blocker. `git diff --check` passed; actual browser/request/model checks remain in
validation.

Nested-return source review accepts the deepest-child-only Back correction. CSS
hides only an outer retained wrapper's direct Back/evidence note when it contains
a nested child; the nodes remain connected and paused. The outer Back handler also
refuses while a nested child exists, so invoking a hidden/reinserted ancestor
control cannot skip the active owner. The deepest Back retains the existing local
close/focus behavior and performs no source read. Native execution status remains
separate; environment asset-load stalls are not evidence of application correctness
or failure.
