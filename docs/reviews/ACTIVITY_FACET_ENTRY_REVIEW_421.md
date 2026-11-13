# Activity contextual facet entry review — 421

Status: bounded source/interaction contract and all six actual ready/counted
frames accepted. Final browser/model execution remains recorded in validation.
This reviewer owns no production changes and runs no heavy checks.

## Product boundary

Activity already has full-scope counted values behind Filters → observed field →
Load values for this scope. The new entry should connect an inspected qualified
field to those existing controls. Selecting a field must not execute a query,
change the applied selector, or treat the retained buffer as a population count.
An explicit load may pause Activity and acquire the existing count/group reads.

The latest completed Activity window may differ from the original observation
window of a retained inspected record. The load target must show its actual UTC
window before invocation. Preserve the complete applied predicates, exact typed
environment policy, signal, and field namespace. A matching key in resource and
record attributes is not the same field.

## Source review requirements

- Reuse the existing observed-field Load action. Resolve the exact qualified
  field, reconcile the local field search, and reveal native disclosure ancestry
  for focus. No synthetic field insertion or automatic source load is needed.
- Counted values admit string group fields. Numeric timestamp, duration and
  severity-number columns must refuse locally rather than offering a doomed
  source request. Field eligibility is independent of whether this particular
  inspected value can be represented by an exact-value predicate.
- Native invocation and current-owner checks must cover both the Load action and
  returned counted-value staging. Hidden, detached, closed, stale, inert or
  disabled controls cannot regain authority through a saved handler.
- A lookup response must match its original full question and read lease.
  Changed or dirty scope withdraws old values and prevents late adoption.
  Conserved groups, missing versus present-empty values and exact integer counts
  remain the existing counted-value contract.
- A contextual focus handoff should retain the inspector. Returning/closing its
  filter surface is local; source reads occur only through explicit actions.

## Source acceptance

The implementation reuses the shared attribute action menu, exact observed-field
Load control and counted-value renderer. Field search is reconciled locally;
no arbitrary field is inserted. `fieldReason` rejects unsupported/numeric columns
before querying, and field entry no longer depends on the current inspected
value fitting a predicate.

`activity-facet-entry.mjs` detaches the fixed question and original observation
window, labels the latest completed read explicitly, and compares the exact
question when checking currency. The Activity owner additionally binds the
inspector generation and applied selector, and the focused Load holds that
current-window lease. Native eligibility and current observed-field membership
protect Load. Counted-value staging/search also use native readiness; dirty scope
withdraws results through the existing lookup owner.

Closing Filters restores the original field action through the shared native
focus helper while its owner remains valid. A staged or changed scope instead
focuses Apply/environment; hidden/departed targets cannot steal focus. The
inspector and its original-window disclosure remain held throughout the local
handoff. No new source, cache or persisted view format was introduced.

The requested feedback refinement is resolved: failed entry returns false and
successful entry returns true, so the shared field action can show its local
unavailable message alongside the owner draft error. Successful field entry also
invalidates the prior counted result and aborts any pending lookup before a new
explicit load. Generation advances before cancellation; a late old response
cannot restore the prior field. Invalidation only clears counted data, preserving
the focused Load and its scope preview, and it does not acquire telemetry.

The live pilot exposed an important construction-order distinction. Render-time
availability must not depend on the new inspector content already being mounted
(or reference that content before its `const` initialization). The owner now
computes offered facet availability from epoch, clean completed scope and field
support. Actual invocation separately checks inspector containment/visibility,
selector ownership, epoch, native readiness and the exact window lease. This
separation is accepted: a rendered enabled action grants no authority by itself,
and it avoids disabling the initial action during detached construction.

The narrow-screen pilot found an automatic flex minimum on the owning Activity
Filters item. Its page rule now permits shrinkage with `min-inline-size: 0` and
`max-inline-size: 100%`. The shared scope preview has its own component stylesheet:
a block layout and overflow wrapping, with sufficient specificity to override the
page's generic inline-small rule. This keeps page placement and reusable text
presentation separate.

The focused Load control now references its scope preview through a unique
`aria-describedby` ID. Cleanup removes only that association when the selected
field or observed-field rendering changes, preserving unrelated descriptions.
The preview remains plain text. This source-level association is accepted;
actual mobile containment and focus identity remain browser/visual gates, and no
screen-reader execution is claimed.

No source blocker was found in the reviewed flow. This is a bounded inspection,
not a whole-product accessibility or screen-reader certification. Actual frames,
source budgets, generated controls and browser execution belong in validation.

## Actual visual acceptance

Independently inspected `activity-facet-entry-{ready,counted}-{1440,390,320}.png`
from the running app. All six frames are accepted; no visual blocker remains.

The ready state clearly focuses the existing Load control and shows the latest
completed Activity window in UTC (`Z`), including the explicit note that the
inspected record window differs. The preview wraps within the selected field at
320 and 390 pixels. Its font and spacing remain consistent with shared evidence
text rather than introducing another panel or status card.

The counted state identifies the qualified `attribute:cohort` field and full
matching population: 261 matching records and four values, with common 258,
present-empty 1, canary 1 and missing-key 1. Missing and present-empty remain
visually distinct. Counts stay aligned and readable at all inspected widths;
search, Read details and individual staging controls remain available. The
separate held-buffer note still says 100 shown records with incomplete
suggestions, so sampled observations are not relabeled as full-scope counts.

The narrow frames show contained controls and wrapped scope text after the flex
minimum fix. Native focused-control rings are visible; exact active-element
identity, keyboard activation, containment measurements, read budgets and stale
handler refusal remain browser assertions, not facts inferred from images.
Source inspection confirms the Load description association; this is not a
screen-reader or cross-engine certification.

## Next depth candidate

After exercising this entry, consider multi-value facet selection against a
retained counted snapshot: OR within one qualified field, AND with the existing
scope, then explicit staging/application. Current counted values stage one exact
value and scope drafting withdraws the lookup. Simply allowing multiple existing
buttons would create contradictory same-field AND predicates. Any deeper slice
needs a distinct bounded selection draft, missing/empty semantics, the existing
predicate-size limits, and no claim that held counts describe an edited scope.
This is a proposal, not implemented behavior or a second active queue.
