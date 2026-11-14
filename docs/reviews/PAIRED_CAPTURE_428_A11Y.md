# 428 paired capture: accessibility and web standards review

Status: final source and responsive-frame review accepted. The parent-owned
complete browser/native verdict remains a separate integration gate.

The bounded review covers the existing native capture dialog and the new
fixed-population comparison entry. It does not infer whole-product accessibility
conformance or physical assistive-technology behavior from source inspection.

## Interaction requirements

- Keep native `details`, `button`, labelled controls and modal `dialog` semantics.
  The dialog title must name the capture task. Opening moves focus inside; Cancel
  and Escape restore a connected, visible opener when one remains available.
- Preview baseline and selection as distinct semantic sections with readable
  environment, fixed UTC bounds, grouping and predicates. Full exact detail can
  be disclosed; a JSON block alone is not an adequate confirmation summary.
- Keep single-question capture visibly distinct from full comparison capture.
  Confirmation performs fresh independent reads; the modal must say so without
  implying that the existing on-screen observations are archived verbatim.
- Pending confirmation must disable duplicate publication, expose busy state and
  retain Cancel. Native disabled/hidden/modal state and origin ownership must
  protect directly retained callbacks as well as ordinary pointer events.
- Failure keeps the form and retry explicit. Announce a concise refusal through
  the existing status region; explain that independently stored children may
  remain, without claiming rollback. Departure retires display adoption rather
  than undoing an already authorized publication.
- At 320/390/1440 widths, preserve readable operand order, wrapping labels,
  reachable actions and contained overflow. Keep exact times available while
  favoring readable UTC presentation.

## Initial findings

The existing dialog is named by `captureDraftHeading`; title/destination fields
have native labels and status uses `role="status"`. The initial source focused
the title on open and used native Cancel/Escape behavior.

The initial confirm handler lacked native pending/visibility checks for direct
retained invocation, and the dialog did not expose `aria-busy`. These findings
were returned to implementation before final review. Final acceptance below will
record their resolution and actual browser evidence separately.

## Follow-up source review

The first implementation resolves the initial interaction findings: confirmation
requires an open dialog and a native eligible control, pending confirmation is
disabled and marked busy, and the optional origin callback is checked before
publication and display adoption. Cancel/Escape share the same retirement and
safe focus-return path. Current successful capture-only adoption focuses the saved heading. An explicit
attachment that opens the investigation draft must focus that visible document
title control instead; retired completion cannot move focus or replace the workspace.

Baseline and Selected use semantic headings and the existing supplied-data
`goo-evidence-summary` element. Fixed UTC intervals are readable with exact detail;
environment policy, grouping and literal predicate scope are retained. CSS stacks
the two operands below 700px. The initial contradictory single-population note
under both capture choices was corrected to distinguish both questions from the
selected-query action.

Independent focused execution passed **13 tests, zero failures/skips** across
paired-preservation, preservation-owner and preservation modules. This execution
uses the actual production helper/controller bodies; actual modal keyboard and
responsive acceptance remains a separate browser gate.

## First real-frame inspection

The first generic-runtime frames at 1440, 390 and 320px show readable operand
headings, UTC bounds and wrapped literal selectors with no horizontal clipping.
Narrow dialogs use native vertical scrolling; final browser acceptance must retain
keyboard/action reachability below the visible second operand. The first frames
still contain the old surrounding single-population note, so final regenerated
frames are required after the corrected copy is embedded.

Actual initial browser checks pass title focus, Cancel and Escape restoration.
The journey caught a focus mismatch after an explicit current-draft attachment:
that action opens the document, so the final target must be the visible document
title control rather than the now-hidden saved heading. Final runtime focus assertions
are required. Separately, the inherited attachment flow must not navigate after
the initiating display lease has retired.


## Final independent acceptance

The corrected source resolves the reported findings. Long valid selectors are
kept in their own native disclosure, so the shared summary's fact-size budget
cannot blank the entire operand preview. Full selector text wraps without
truncating its meaning. The evidence note distinguishes capturing both questions
from capturing only the selected query.

Attachment navigation is evaluated at the actual handoff. Current capture-only
adoption focuses the saved title heading; attachment to an editable investigation
focuses its visible title input. Retired attachments may complete their explicit
reference write but cannot force navigation or focus. Empty-draft initialization
is source-local and does not self-invalidate the final navigation decision.
Independent final execution passes **16 tests, zero failures/skips**: four paired
scope/preview tests, four actual document attachment owner tests, seven
preservation-owner tests, and the capture-route owner regression.

All six regenerated preview/stored-pair frames at 1440/390/320px were inspected.
Operand headings, UTC bounds, literal selectors and stored comparison values are
readable. There is no horizontal clipping. Narrow dialogs and saved pages require
normal vertical scrolling; this is not hidden or unreachable content. The final
journey additionally checks native modal keyboard reachability and destination
focus, separately from this frame inspection.

No remaining source or visual blocker is identified. Nonblocking consolidation
work: reduce unrelated destination prose and disabled question-link controls when
Capture only is selected. This is a bounded follow-up, not a request for another
capture architecture or display shell. Physical AT,
speech output, touch ergonomics and cross-engine testing are not claimed.


## Final frozen-runtime follow-up

The final stored-pair display now supplies the existing comparison-context
component with Frozen evidence, original per-side source clocks/bases and exact
immutable child IDs. It removes inherited browser-held origin wording without
inventing a child custody timestamp or mutating stored evidence. The additional
independent stored-context regression passes, bringing this review's focused
checks to **17 passing tests, zero failures/skips**.

The final regenerated preview and stored-pair frames remain readable at 320, 390
and 1440px. The clearer shared-scale comparison retains baseline/selected labels,
counts and distinct colors; frozen context is disclosed separately with exact
scope. The final actual journey reports title/Cancel/Escape focus, 320px keyboard
Tab reachability to Save and Cancel, editable-draft title focus, and preservation
of the departed page's hash/focus after late attachment completion. Those browser
assertions close the interaction findings; the parent still owns the full
browser/Bombadil/native campaign verdict, which this review does not infer solely
from frames or focused tests.
