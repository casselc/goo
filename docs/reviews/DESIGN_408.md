# Whole-product design review — 408 checkpoint

Reviewed the current capability index, 405 reconciliation, 406/407 contracts,
shared display code and the actual frames listed below. 408's export/reference
work is in progress at review time. This review ran no browser, accessibility,
performance or vendor-parity campaign; screenshots establish layout, not keyboard
or screen-reader correctness. The proposals below are not an adopted queue.

## Product judgment

Keep the product loop: **see current telemetry → isolate a population → follow
reported evidence → preserve or reuse the question**. The useful distinction is
between a question, a supplied read and immutable saved evidence. It should be
clear through actions and compact state, rather than repeated explanatory text.

No checkpoint-blocking visual regression was found. Activity remains recognizably
a stream; fixed results and Saved now put real rows ahead of their editors. The
new rates and histogram presentations fit the existing scope/Run model. Do not
replace this with a dashboard home, general canvas or another top-level builder.

The remaining friction is cumulative chrome and interaction vocabulary across
owners, rather than a lack of chart types. Preserve the same visual identity and
controls across Activity, Explore, Services, Saved and authored views while keeping
their different source/custody authority explicit.

## Evidence across the product

| Surface | Inspected evidence | Finding |
| --- | --- | --- |
| Activity | `352-unreported-activity-390.png`, `333-activity-comparison-entry.png`; `LIVE_ACTIVITY.md` | Rows, applied scope, pause and loaded counts are visible before question authoring. Expanded Investigate exposes six similarly weighted buttons; this is a contextual menu, not a reason for another permanent toolbar. |
| Fixed records | `384-ordinary-desktop.png` | Count/scope/Run are compact, but a four-record result consumes most of a desktop viewport. Selection, Inspect and Expand trace form a vertical action stack in every row. |
| Trace search | `403-trace-continuation.png` | Conditions collapse correctly and continuation has named fresh-read actions. The full source basis hash is inline between useful counts and paging; it belongs in the existing read-facts disclosure. |
| Services | `385-outgoing-http.png`, `390-final-service-histogram-1280.png` | Destination table combines identity, quantitative bars and actions well. Expanded native data repeats stream/coverage summaries and frames above the plot, consuming the available widget width. |
| Saved | `384-saved-record-desktop.png` | Evidence state and Open fixed question are clear. Stored records repeat the tall row-action layout; fixing shared record presentation should help both owners. |
| Authored histograms | `407-authored-histogram-390.png` | Histogram/Table and complete-only controls stay together. Three nested borders, two near-identical distribution headings and wrapped explanatory notes compete with the plot. Only a small part of a wide bin set is visible; scrolling needs a stronger affordance. |
| Review/annotation | `CAPTURE_RECORD_REVIEW.md`; `ui/capture-record-review.mjs` | Exact evidence targeting and local judgments are honest. Reusing a vocabulary requires pasting an exact task ID; there is no contextual discovery path. A single-record workflow also offers no explicit next/previous review occurrence. |

Screenshot paths above are relative to `docs/journeys/screenshots/`. The 384 frame
named Activity depicts an applied fixed question, so it was not treated as proof
of the current live-entry layout. Older frames are historical samples of shared
layout, not a claim that every present route was re-executed.

## Highest-value proposals

### 1. Consolidate quantitative view chrome

Use one chart frame/title/compact scope-state row per visible result. Move hashes,
full stream identity, read clocks and repeated mathematical caveats into existing
native disclosures. Retain literal units, unknown/partial state and the distinction
between captured count and valid numeric values in the visible summary.

Evidence is concrete in `ui/scope-view.mjs` (frame plus quality plus inner chart),
`ui/numeric-distribution.mjs` (outer adapter), and expanded native service widgets.
Apply the smallest shared density/profile adjustment rather than page-specific
negative margins or hiding uncertainty.

Acceptance: the same histogram and native interval remain understandable at 390 px
and desktop with one visible title; the plot starts higher at natural entry, not
only after a screenshot scroll. Wide charts expose a named, keyboard-usable scroll
region or equivalent existing control. Exact values and read facts stay reachable.
Check ordinary Explore, Services and Saved; no source or declaration changes.

### 2. Discover review vocabularies in context

Add a bounded task chooser inside the existing Review example disclosure, retaining
exact-ID load as an advanced/reproducible path. Show task title, vocabulary summary
and identity. Choosing a task does not label a record or fetch telemetry. Avoid a
new review navigation area or globally selected task that leaks between inspectors.

Confirmed seam: `ui/capture-record-review.mjs` creates Exact task ID / Load task /
Declare a task; the present server contract exposes task creation and exact load.
A bounded metadata-list contract is a prerequisite, not an assumed existing API.

Acceptance: another visit can find an existing vocabulary, inspect its labels and
select it for one exact captured occurrence. Pending chooser results and retained
callbacks cannot replace a different inspector/task. Empty/error states preserve
local drafts. No new custody target, negative-label inference or publication.

### 3. Deepen the existing captured-occurrence review flow

After task discovery, offer next/previous occurrence within one explicitly selected
immutable capture. Reuse the same task while showing the occurrence index, record
identity, event time and judgment state. Changing occurrence must not copy a draft
judgment or silently save it. Do not call this a training set, queue-wide review or
current consensus; export retains the existing historical-manifest contract.

Acceptance: review two distinct occurrences with the same reported span/trace IDs
without merging them; navigate away with an unsaved draft through an explicit,
consistent choice. Each saved judgment binds `{captureId,index}` and original
receipt. Uncertain/Abstain remain separate from labels. Keyboard focus returns to
the occurrence navigation control; telemetry reads remain zero.

This is a proposed narrow depth item, not permission to add kept-observation
custody or a second target identity. Those require a separate admission decision.

### 4. Standardize fresh, held and local display actions

407's real journey found both the wrong resolver and a completion callback that
rejected its temporarily disabled opener. It also found a distinct local Back
path and held-render renewal seam. Users should not have to understand those
owners to know which action is local and which reads fresh data.

Adopt a small vocabulary and shared visual placement: **View** changes presentation;
**Run** reads; **Inspect** opens held detail; named trace/context actions identify a
fresh read; local Back returns to the named parent. Held plots retain local view
controls, show one short paused notice and expose no live range actions until Run.
Do not make every button label a full disclaimer.

Acceptance: ordinary numeric, authored histogram, native interval and Saved detail
follow the same action/focus conventions while retaining their distinct guards.
Exercise genuine disabled/pending/hidden controls and late callbacks; do not weaken
source leases to make a UI label uniform. Consolidate repeated owner mechanics only
where architecture review identifies a common contract.

### 5. Compact shared record actions

Keep one row's UTC timestamp, service, record summary and primary action readable
without the per-row three-level button stack seen in fixed/Saved frames. Selection
should align with its row; secondary Expand trace can use an existing action
pattern. Native table semantics and keyboard-visible actions stay intact.

Acceptance: representative ordinary and Saved rows show materially more records in
the same viewport without shrinking text or removing meaningful fields. At 390 px,
identity and the primary action remain associated. Test long names, missing/empty
values, explicit checkbox labels and return focus. This is shared presentation
work, not a new table framework or data acquisition path.

## Suggested cadence and boundaries

A reasonable finite sequence is **breadth: contextual review-task discovery →
depth: review another exact occurrence in the same capture → consolidation:
quantitative chrome and source-action conventions**. Planner should weigh this
against the product/architecture review before adoption. Compact record actions
can be a later candidate, not a fourth automatic milestone.

If broader observability functionality is preferred over annotation depth, keep
that choice explicit; do not relabel an existing delivered capability as missing.
Exponential histograms, grouped percentiles, warehouse lowering, arbitrary kinds,
new stores and multi-user presence retain their independent semantic/platform
prerequisites. Do not restart the large roadmap from historical proposals.
