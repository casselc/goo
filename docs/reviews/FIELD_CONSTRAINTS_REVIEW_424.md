# Field-constraint integration review — 424

Accepted for the bounded architecture/source review below. Final execution
results belong in [validation](../../VALIDATION.md); this review did not run
builds or browser/model tests.

Reviewed the new integration in `ui/filter-editor.mjs` and the Activity callbacks
in `ui/stream-entry.mjs`. The reviewer authored the separate pure planner and
focused UI tests; this is a review of the composing integration, not an
independent endorsement of that planner.

- **Scope binding:** the supplied whole record question is validated before the
  editor projects its existing `omitempty` representation. Both entry and action
  use that same projection. The live tree must match complete canonical content,
  with the original typed environment, signal, fixed window and other metadata
  retained. Empty strings and integer scale zero do not accidentally withdraw
  otherwise valid actions. Clause order/path remain significant.
- **Custody:** each action checks the current editor generation, owner lease,
  native control visibility/readiness and current predicate. Exact resolution
  uses a freshly reconstructed question, rather than comparing only to a frozen
  copy. Re-render, mutation and observation withdrawal retire the panels; old
  callbacks cannot regain authority when equivalent content returns.
- **Semantics:** Edit resolves the original clause path to the existing operator
  control. Direct Remove is restricted to conjunctive ancestry; OR branches have
  Edit only. The integration does not invent Boolean complements, remove the
  first matching field, or discard mixed-field groups.
- **Atomic staging:** a validated removal changes the tree once and emits the
  existing change contract. Activity's `draftScope(false)` invalidates counted
  values and pending reads, clears suggestions, pauses source work and preserves
  held observations. Only explicit Apply acquires the new Activity window.
- **Focus and layering:** the shared editor owns passive supplied-question
  presentation and local changes. The Activity adapter owns the new-window
  message and focus to its existing Apply control. The standalone editor has a
  connected Add-condition fallback. Existing disclosure/native readiness guards
  prevent hidden, inert, disabled and closed controls from revealing or staging.
  No source acquisition or full retained-row snapshot was added to these paths.

The initial normalization mismatch was corrected before acceptance: planning
against authored `value:""` while resolving against omitted-value editor output
would refuse a valid unchanged clause. Focused tests now cover explicit/omitted
empty values and integer scale-zero variants through actual Edit and Remove.

No remaining source-level blocker found. The review does not establish remote
source counts, final screenshot quality or model-trace validity; those require
the separately orchestrated actual-source journeys and model checks. This slice
is contextual Activity entry into existing controls, not a generic automated
Boolean-broadening engine or a new independent query owner.

## Independent UX/accessibility follow-up

Accepted within this bounded source and rendered-frame review. This reviewer
prepared the design/storyboard and did not author the production planner or UI
integration. No builds, browser interactions, screen-reader runs or automated
contrast measurements were independently executed by this reviewer.

Reviewed actual `field-constraints-{1440,390,320}.png` and
`field-constraints-staged-390.png`, plus the final shared editor and Activity
callback source. Compact clause rows fit at all three widths without an added
card hierarchy. Each row keeps the exact qualified field, quoted value and
All/Any ancestry visible. The mixed OR clause visibly offers Edit only with a
short explanation; the conjunctive clause has the separate Remove action.

Native disclosure summaries and action names retain their visible labels and
include qualified-field/ordinal context. Description IDs connect actions to the
exact clause text. Edit targets the existing operator control; staged removal
returns to the existing Apply control, whose strong focus ring is visible in the
390px frame. The staged message clearly says Apply uses a new Activity window.
The original counted result remains identified separately with its captured total
and unknown coverage; source review confirms it withdraws on staging.

The existing observed-fields viewport remains a bounded scroll area. Later field
actions can lie below its visible edge when the clause list is expanded; native
keyboard focus reveals those controls. This is an accepted composition limit,
not evidence that every field action fits simultaneously on a narrow screen.
No remaining UX/accessibility blocker was established in this review. Final
source budgets, raw-source membership, actor completion and model outcomes belong
in the root-orchestrated validation, not this visual acceptance.
