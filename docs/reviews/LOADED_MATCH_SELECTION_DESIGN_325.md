# Select loaded matches — design 325

Status: frozen source and all three actual frames accepted; final native/actor completion is reported separately by verification.
Read with repository guidance, the current roadmap and
[Find design 324](HELD_FIND_DESIGN_324.md).

## User task

A responder finds a suspicious cluster in held Activity logs or spans, then keeps
those original observations together. Reuse the existing Find, exact selection
and explicit Keep flow. This is not a full-population query, new search grammar,
backend batch endpoint or new capture owner.

## One contextual action

For a nonempty valid Find, place **Select N loaded matches** beside the matched/
loaded count in the existing compact record controls. N is the current matching
occurrence count, including matches already selected; do not imply N new items.
No action is needed for empty Find. Clearing Find does not clear selection.

The action requires an explicitly paused stream. While live it is disabled with
**Pause to select matches**, pointing to the existing Pause owner. Do not auto-pause
or fetch a replacement buffer. Zero matches disables it with **No loaded matches**.
If every match is already selected, disable with **All N matches selected**.
Invalid Find hides the action and keeps its existing validation feedback.
Avoid stacking these mutually exclusive explanations into a prose block.

## Atomic original-occurrence union

Activation snapshots current matching occurrences under the current paused
buffer/selector/connection. Union with the existing selection, including hidden
items, using original occurrence keys. Repeated trace/span IDs or equal text are
not deduplication authority. Never toggle off an already selected match.

Before mutation, validate the entire union: at most 20 items, exact preservation
proofs unexpired/available, same checked selector and connection. A failure leaves
all selection state unchanged. Never silently take the first 20, replace hidden
selections, or initiate a source read to repair an invalid proof.

An overflow should explain the task boundary: **N matches would make U selected;
limit 20. Refine Find or clear selection.** Expired/unavailable proofs should state
that the matching observations can no longer be kept and direct users to the
existing explicit fresh-read/Resume route. Do not claim missing source telemetry
or automatically reconnect. Exact failure copy can reuse the existing selection
error vocabulary where it gives the same actionable meaning.

The existing selection panel reports the resulting total and hidden selected
count. If the batch action owned focus, successful selection moves focus to the
existing enabled **Keep selected** action, without scrolling or activating it.
The completed batch action may now be disabled/replaced, so it is not a valid
focus-restoration target. If Keep is ineligible, focus the stable selection status
with programmatic tabindex=-1, or Find if no such status seam exists. A rejected
batch preserves focus on its still-current enabled action, or restores the
corresponding current action/error status after replacement. Scope or pending
updates must not steal focus from unrelated controls.
Existing **Keep** remains a separate explicit write with original references,
window, membership proofs and custody semantics. Selection itself is zero IO.

## Ownership and stale controls

New arrivals, eviction, query edits, clear/reset, reconnect and pause transitions
must not let a detached or reattached old button act on a newer corpus. Require a
visible current action and the current Find/buffer/scope/selection lease, then
revalidate readiness at activation. A stale refusal has no selection or source
effect. Do not invent another selection owner; implement batch preflight through
the existing TailSelection boundary.

Hidden individual row actions retain their existing visibility guards. The batch
action deliberately selects the current visible matches; it does not make hidden
row buttons actionable. Paused return and inspection still retain their original
observation window and current Find/selection ownership.

## Acceptance

Actual logs and spans should demonstrate a multi-match selection with an existing
hidden item, duplicate selection without removal, original-ref Keep and unchanged
source question/read counts. Exercise zero, empty, live refusal, already-selected,
20-item union, overflow, expired/missing proof and reattached stale action with
atomic negative assertions. State/property tests should cover union idempotence,
order independence, no partial mutation and exact occurrence identity.

Review three real frames for the contextual action, hidden-selection union and
explicit kept result; separate visual count/label proof from native source-budget,
focus and original-ref custody assertions. Do not add layout or library machinery
unless the actual journey reveals a blocker. No product/browser edits are part of
this design review.

## Frozen-source review

Source accepted pending actual frames/native results. The contextual action stays
inside the existing record controls and uses 32px targets, 44px for coarse pointer.
Readiness explains one reason at a time: no loaded matches, pause required,
already selected, an active saved selection, or precise preflight failure.
Unavailable/expired proofs and over-20 unions are rejected before mutation with
explicit recovery guidance. Repeated valid matches do not toggle selections off.

The existing TailSelection now owns pure whole-union planning and commit; both
individual and batch additions use that boundary. It validates existing hidden
selections as well as new matches. Command revision, current button identity,
actual visibility, pause state and current corpus signature guard activation;
source query and Keep remain separate owners.

Successful keyboard activation transfers focus to enabled Keep, stable selection
status or Find as contracted. A disabled replacement's reason is programmatically
focusable; unrelated refresh does not steal focus. Source acceptance does not
substitute for actual keyboard/custody/source-budget or compact-frame results.
No product/browser edits were made for this review.

## Actual-frame review

All three actual frames accepted; final runner/actor completion is a separate
verification outcome. `325-overflow-refused.png` shows 21 of 24 loaded matches,
one existing hidden selection and the disabled batch action explaining that the
union would be 22, above the limit of 20. The existing selection remains one;
there is no misleading partial-first-20 success or full-population count.

`325-hidden-selection-union.png` shows two of 24 loaded matches selected alongside
one hidden original, giving three selected. Both visible occurrence controls say
Selected and the batch reason says All 2 matches selected. Existing Keep selected
remains the next explicit action. The distinction between match count and total
selection is clear without another toolbar or modal.

`325-span-union-kept.png` shows Kept 3 observations, two of three loaded matches
and one selected not shown. The saved-selection guard explains why batch addition
is unavailable until the existing selection is finished or cleared. The actual
span rows retain UTC and 12.000 ms values. Record fields happens to be expanded
in this frame; it is an existing disclosure, not additional batch-flow controls.

These desktop frames demonstrate labels/counts/refusal and held-versus-kept state.
Atomic negative behavior, exact original-reference Keep, source-read budgets,
keyboard focus and stale-control refusal depend on the native/property assertions,
not on screenshot appearance alone. No semantic/accessibility blocker was found,
and no further layout changes are requested. No product/browser edits were made.
