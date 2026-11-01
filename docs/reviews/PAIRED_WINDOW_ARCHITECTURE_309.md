# Paired fixed-window authoring — architecture 309

Status: source review accepted; executed checks are tracked in `VALIDATION.md`.
Saved query comparisons
must remain practical to rerun in a new fixed window without hand-editing two
questions or silently breaking the shared-context invariant. No new query profile,
endpoint, persistence owner or live subscription is needed.

## Pure transition

Accept one validated fixed-query comparison view plus explicit integer UTC bounds.
Clone both root and baseline questions, changing only `fromMs`, `toMs` and root
paging offset zero. Preserve exact predicates, typed environment policy, grouping,
measure directives, limits and every presentation/record option. Validate the
resulting root and comparison together before publishing either mutation. The
old view must remain unchanged when validation fails.

Use existing question limits (positive bounded fixed interval) and existing
composition/timeline/bucket constraints. A wider window can exceed an existing
interval budget; reject that edit clearly rather than silently altering bucket
width or removing a timeline. A fresh-window shortcut samples one clock once and
retains the current shared duration for both questions. This is fixed-window
rebasing, not a live question or automatic time advance after Save.

## Existing application owner

Stage through the existing root-query draft path, clearing held result authority,
prepared pair and pending count work before Run. Reuse current count serial and
controller retirement; late selected/baseline/comparison responses must not publish
into the rebased draft. Source-free staging cannot relabel old data as observations
for the new bounds. Explicit Run performs the existing independent selected and
baseline reads with fresh clocks; Save/export persist only the rebased intent.

The form must use existing draft/error registries so first edits block Save/Apply
while Add/Apply or Discard remains usable. Discard restores the current admitted
pair without a read. Capture the current generation/root/declaration or equivalent
checked draft lease, retire detached/reinserted controls after replacement, and
retain the established pending page/selection/record workflow guards.

## Scope gestures and consistency

Root-only scope actions cannot be left enabled with a fixed baseline. The current
Fresh window action, timeline brush, reset-window and restore-start controls must
either route through the validated paired transition or remain explicitly disabled.
Do not derive a former baseline from `original[]`, which currently records only the
root. Shared-window editing does not authorize grouping/profile rebinding or
predicate replacement; those remain separate bounded behavior.

## Acceptance and review

Open a saved fixed pair, stage a new UTC interval or fresh window, prove both
questions retain every non-window field and no request occurs, then Save/export
and explicitly Run. Verify both actual request bounds and new independent clocks.
Reject invalid UTC ranges and incompatible existing timeline/bucket budgets without
mutation. Exercise first-edit Save blocking, Discard, pending/late baseline and
comparison responses, stale control reattachment, and existing captured-comparison
behavior. Keep the browser storyboard and observed model aligned with these real
user operations, distinguishing source review from executed validation.

## Final source review

`rebaseFixedComparison` validates the current fixed pair, clones both questions,
changes only bounds/offset and validates the resulting composition before return.
The explicit UTC editor uses existing draft/error registries; its own draft remains
finishable/discardable while ordinary Save/Apply/Run are blocked. Its lease includes
current generation, exact root and baseline identity, query-variant check and
connected active owner. Pending baseline/comparison/page/record work blocks edits.

Staging uses the existing root owner: `clearViews` retires count serial/controller
and prepared pair, then generation/source cancellation, held-result removal and
`markDraft` retire saved identity and prior response authority. Timeline brushing
and window reset use the same paired helper; full restore-start is disabled because
no original baseline AST is retained. The existing Fresh action now becomes **Use fresh comparison window** for fixed
pairs: it samples `Date.now()` once, retains the current duration and stages both
questions through the same validated paired helper. Its current-owner and pending
guards match explicit editing; it performs no source read. Bucket/timeline validation
is retained without silent width adjustment.

Source review has no remaining architecture blockers. `git diff --check` passed;
actual native lifecycle and property/model outcomes belong in validation.
