# 428 paired capture review

Status: architecture, production source, focused tests and responsive-frame review
accepted; final integrated browser/model acceptance pending. This report will distinguish independent
review from the parent-owned integration verdict.

## Existing boundary

The implementation should adopt `CaptureStore.createComparison`, not introduce a
new capture engine or reinterpret a browser-held result as stored evidence.
`CaptureQuestion.Validate` checks both valid fixed record queries, the same signal
and environment policy, and one identical categorical breakdown. Distribution
captures have a separate numeric/histogram exception; that does not broaden the
ordinary fixed-count entry in this milestone. Inspection paging is normalized
server-side without dropping predicates or fixed bounds.

Publication orders baseline then selection. Each successful child is independently
stored before cohort comparison and pair publication. A failed second read or
failed pair publication can leave useful children in Saved. There is no rollback,
transactional source snapshot or automatic retry. The UI must retain its current
investigation and make an explicit retry's fresh reads clear.

This general pair stores a `SavedComparison` containing both child IDs, complete
cohort request and checked result. The separate `/api/captures/comparisons`
`EvidenceComparison` receipt freezes a different decision over ordered capture
references and currently withholds ordinary numeric comparison across differing
selectors. A selected/rest pair must not be substituted into that receipt family
or weaken its compatibility rules. The existing general capture-card path is the
appropriate investigation attachment.

## Interaction recommendations sent to implementation

- Keep **Capture selected query** distinct from **Capture comparison**. Opening
  either form is local configuration plus existing metadata target discovery;
  only confirmation authorizes fresh source reads and immutable publication.
- Preview baseline and selection with exact typed environment, predicates,
  grouping and UTC interval. Preserve operand order and label both independently.
  The form captures fresh evidence, not the displayed pair verbatim.
- Use the existing native capture dialog, supplied-data scope summaries and
  controls. Keep exact query detail available without putting JSON or long prose
  before the main action.
- Add native dialog-open/visible/disabled and pending guards to the shared confirm
  boundary. The previous `if (!draft)` check alone allowed a retained hidden
  callback or duplicate pending activation to initiate another capture request.
- Bind the new origin to its semantic population owner/revision. Do not test a
  background opener's native readiness after `showModal`, because modal behavior
  intentionally inerts that background. Departure/edits before confirmation
  refuse; departure after confirmation retires display adoption. Existing explicit
  publication/attachment may finish without navigating or replacing a newer view.
- Pair errors must mention independently stored children may remain available;
  do not imply rollback or silently start another source read.

No new Go API, artifact family, query language or generalized permission layer is
needed for this slice. Source completeness stays unknown and per-child source
clocks remain distinct from the pair's stored receipt time. No-data and diagnostic
states remain evidence, not synthesized healthy/failing populations.

## Checks still required

Review the final production diff and focused tests; run the real generic OTLP
journey with fresh/late data, source refusal, immutable pair/child reopen,
attachment, departure and retained return. Inspect 1440/390/320 frames, native
keyboard/disabled behavior and no-read local operations. Native model/property
acceptance must use actual completed reports and rejecting controls. Physical
assistive-technology and cross-engine acceptance are not implied by source review.

## First implementation and frame review

The production helper now explicitly limits this entry to ordinary Count
comparisons, including a check of the selected declaration's measure and refusal
of duration-summary/status/numeric/specialized profiles. Reusing the broader
`eligible` predicate alone was insufficient: it also allowed duration-summary
queries. That finding was returned to implementation and corrected.

The new pair entry forwards a detached complete ordered question and a semantic
origin callback rather than a serialized request token. Confirmation checks the
native open dialog and enabled control, while owner generation/revision protects
its background population. The preview uses shared evidence-summary elements,
UTC formatting and human-readable complete selectors, not a JSON editor.

Independent focused execution of paired preservation, preservation ownership and
capture route ownership passes all three file runners with zero failures/skips.
The new Go tests use the real existing comparison creator for second-source
refusal/limit and no-data/diagnostic states. The pair-publication-failure test
creates real children then exercises the real immutable publisher; it is not a
claim of an end-to-end injected filesystem fault inside `createComparison`.

Actual 1440/390/320 preview frames were inspected. Operand labels, environment,
UTC bounds and distinct predicates remain readable without horizontal clipping.
Desktop confirmation fits; narrow dialogs scroll internally, so Save/Cancel
keyboard reachability remains an actual journey requirement. The inherited
destination controls and unconditional revision explanation consume avoidable
space before operands; hide irrelevant controls/notes in measured consolidation
429 rather than claim this existing form is fully compact already.

The first real current-investigation attachment exposed a deeper existing
navigation boundary: attachment may publish/mutate its explicitly chosen draft,
but must not navigate, scroll, focus or replace a document after the capture's
display lease retires. An optional completion-time navigation predicate is the
bounded correction. For saved revision attachment, withholding navigation must
also withhold `show` adoption, not merely call `show(doc, false)`, because `show`
replaces the editor's current document and cards even without a route change.
Final review of that correction and the completed browser/model report is pending.

## Attachment correction source acceptance

The completed bounded patch passes a navigation predicate through the existing
attachment owner. The saved-revision branch evaluates it after its asynchronous
reads/publication and guards `show` entirely. The current-draft `pin` and
`linkCapture` paths guard navigation, show, scroll and focus; no-navigation fresh
draft initialization propagates the option. The explicitly requested reference
can still be attached without a late capture completion replacing another screen.
Owned current-draft attachment focuses the visible editable document title control;
capture-only completion focuses the saved capture heading. These are distinct
destinations, so a hidden capture heading is not the attachment success target.

The contradictory selected-only group note is corrected, and unsupported pair
profiles now expose a visible refusal rather than relying only on disabled-button
tooltips. No remaining source-review blocker is identified. Actual attachment
callback tests and the final real journey/native verdict remain separate gates.

## Final source, focused tests and frames

The final empty-document attachment initializes with navigation withheld, then
makes one navigation decision after appending the authorized card. This avoids
initialization retiring its own completion lease. Four tests extract and execute
the actual document-owner callbacks: retired current-draft mutation, one owned
initialization/navigation, saved-revision departure across both awaited operations,
and linked-question evidence without retired navigation. They verify original
document/cards/route/status/focus retention where applicable.

Independent final rerun passes **16 tests, zero failures/skips** in 0.403 seconds:
paired configuration/preview admission, actual preservation owner, capture route
owner, and document attachment owner. No remaining source or focused-test blocker
is identified.

Final runtime preview and stored-pair frames at 1440/390/320 were inspected.
Selectors use native disclosure; operand labels and UTC intervals remain readable.
The saved comparison reuses common comparison rows and shared-scale bars. Narrow
layouts retain contained vertical scrolling and no horizontal clipping; final
actual keyboard completion is still a journey gate, not inferred from screenshots.

Review found an inherited presentation issue: frozen general pair operands showed
the underlying comparison dataset's `browser-held` origin even though the page is
stored evidence with immutable child receipts. The bounded adapter correction is
accepted: it uses the existing `data-first` / `external-context` display hooks and
supplies frozen Baseline/Selected context with original per-side source clocks,
counts, bases and exact child IDs. The immutable result and unknown coverage stay
unchanged. Ordinary render resets the stored context; saved-pair opening binds it
after its internal render. No child request or guessed custody timestamp is added.
An independent supplied-context test passes and verifies unchanged input, both IDs,
source clocks and no-data semantics. Refreshed actual frames complete the presentation review below. The parent owns
the integrated browser/native verdict separately.

## Final refreshed presentation acceptance

The final runtime frames were inspected independently after the frozen-context
correction:

- [Preview 1440](../journeys/screenshots/428-pair-preview-1440.png),
  [390](../journeys/screenshots/428-pair-preview-390.png),
  [320](../journeys/screenshots/428-pair-preview-320.png).
- [Stored pair 1440](../journeys/screenshots/428-stored-pair-1440.png),
  [390](../journeys/screenshots/428-stored-pair-390.png),
  [320](../journeys/screenshots/428-stored-pair-320.png).

The stored comparison now shows Baseline and Selected, shared-scale record bars
and separate Frozen evidence context. The misleading browser-held label is absent
from the visible saved-pair presentation. Readable UTC bounds and unknown source
coverage remain explicit; exact reference/source details stay disclosed. Both
compact and narrow layouts preserve contained scrolling and readable controls.
No source or visual blocker remains. The inherited capture-only destination
verbosity is an appropriate bounded follow-up for measured consolidation 429.

This review accepts source, focused callback tests and final frames. It does not
substitute for the parent-owned completed Bombadil/native journey verdict, physical
assistive-technology testing or cross-engine acceptance.
