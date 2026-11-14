# 427 exact saved-view recall review

Status: source, focused-test and responsive visual review accepted.
Full real-driver/model acceptance is reported separately by its owner.
This report distinguishes independent source/focused checks from the parent-owned
real browser/model acceptance.

## Contract under review

The existing Recent / Pinned journal may retain a bounded immutable saved-view
reference and display title. It must not serialize source, query operands,
composition, rows, receipts or private authoring state into browser history.
The immutable ID identifies the entry; title is a human-readable hint, not
meaning or authority. Pin, filter, remove and journal persistence remain local.

Stage resolves that exact reference through the existing definition reader. The
complete admitted view must survive: all configured queries, typed environment
policy, fixed bounds, composition/comparison operands and checked CEL authoring.
Staging is configuration loading, not telemetry execution or evidence capture.
The existing explicit Run remains the source-read owner.

## Early findings sent to implementation

1. Existing `population.load` destructively clears the held pane, children and
   live state before the metadata request. A missing or incompatible reference
   must instead leave the current investigation intact. Resolve/admit a detached
   candidate first; only the current private invocation may adopt it.
2. `Composition.view` is not complete root/query admission for non-composition
   views. The browser must check supplied shape/queries/contracts before any
   parent restoration or pane replacement, while the existing Go store remains
   authoritative for content ID integrity and source recompilation. Do not
   reproduce Go JSON hashing in the browser.
3. Keep must restore focus to a connected retained draft/opener. Explicit adoption
   should focus the new scope heading rather than a removed picker button or an
   accidental execution button. Native labels/disabled states must reflect pending
   or retired actions; a saved title cannot stand in for resolved scope facts.
4. A route/hash comparison alone cannot stop A → B → A revival. Resolve, failure,
   dirty-choice, direct callback and metadata publication need private invocation
   ownership. Cancellation is cleanup, not proof of current authority.

These are bounded corrections within the existing catalog and presentation owners;
no new save surface, permission coordinator, evaluator or journal data format for
full definitions is needed.

## Acceptance evidence

Source review confirms detached admission before adoption, immutable reference identity,
private picker and caller epochs, source-inert staging, self-remember refresh
suppression and explicit dirty-choice ownership. Peer Stage buttons now supersede
the pending invocation rather than appearing enabled while doing nothing.

Independent focused execution: history and picker file runners pass. After
fixture and diagnostic updates, all 11 saved-view route-owner tests and all three
4366 same-ID/current-view recall tests pass independently (zero skips). This includes
same-ID Keep/Discard, retired callbacks, route A → B → A, pending departure,
explicit retry, suspension/disposal and reentrant invalidation. The earlier
failures were returned to the owner and resolved before this rerun. The missing
`prepareSavedRecall` concern was resolved: it is the investigation-trail API,
not a missing exploration export.

The five new 427 controller tests and five existing CEL preservation integration
tests also pass independently, with zero skips. They exercise detached failures,
whole admitted definition adoption, same-ID supersession and A → B → A, parent
edits, aborted prompts, incompatible shapes and valid/invalid source recall.

The dirty prompt now focuses its safe Keep action. It restores a connected
previous investigation control or opener; adoption focuses the staged scope
heading. Abort removes the prompt and settles its waiting choice. The native library follow-up checks the currently acknowledged saved ID rather
than the incoming target ID; its conflict is mounted in the visible population
panel instead of the canvas-inert root library host. The additional clean CEL
A → B regression passes independently.

No remaining source-review or responsive visual blocker is identified. The final
independent rerun passed 24 controller/regression cases: 427 (5), same/current
saved-view recall (3), route owner (11) and CEL preservation (5), zero skips.
History/picker runners passed earlier in this review.

The actual runtime screenshots at 1440, 390 and 320 pixels were inspected:

- [Desktop](../journeys/screenshots/427-recent-pinned-1440.png)
- [390 px](../journeys/screenshots/427-recent-pinned-390.png)
- [320 px](../journeys/screenshots/427-recent-pinned-320.png)

Titles, revision hints and action groups remain readable at each width; narrow
controls wrap and staged configuration remains distinct from explicit Run.
No page-width overflow is visible. The held-tab list scroll boundary partially
shows its disabled Current control; it is an internal list viewport, not missing
content. A bounded future polish item is to bring desktop row actions closer to
their titles instead of spanning the full content width. This does not block 427.

The parent-owned full real driver/model acceptance remains separate; this review
does not infer its success from screenshots or focused controller tests.
Physical assistive-technology and cross-engine checks are not claimed by this
review.
