# Whole-product incident-flow usability review — 312

Status: current running-UI/source review complete; bounded 312 fixes and native
regression acceptance are recorded separately below when available.

## Review basis and limits

Reviewed the running app at `http://127.0.0.1:8088` on 2026-11-02 using independent
Chromium contexts at 1280×900 and 390×844. Exercised Activity pause, service
discovery, lab-web incoming HTTP overview, Customize, explicit Run, scope/measure
editing, method values, staged selection, selected-versus-rest child, comparison
promotion, trace close/return and the Activity Views disclosure. These were
unsaved interactions with real lab telemetry; no Save, Capture, Keep, annotation,
import or other durable write was submitted. The main store was not modified.

Also inspected actual 303–311 journey frames and current source/contracts. Existing
native journeys establish Save/import/reopen behavior; this reviewer did not
republish definitions merely to repeat those checks. Browser screenshots used for
this review are transient `/tmp/goo312-*.png` scratch files, not new permanent
evidence bundles. Findings below distinguish live DOM observation from historical
frame/source evidence. No competitor instance or user study was run.

## Product judgment

The initial experience now matches the user's direction: a compact live stream,
UTC event times, service context, Inspect/trace/selection actions and collapsed
scope/investigation controls. The product has a connected, useful investigation
model. The next gains come from making its transitions reliable and giving one
active workspace clear control ownership, rather than adding another chart,
service profile or authoring surface.

The actual fresh entry showed 89 buffered log records and only Pause stream as
the primary execution control. At 390px the document width stayed 390px; records
became readable cards without whole-page horizontal overflow. Pausing and closing
an Activity trace restored focus to its Open trace button and kept Activity paused.
Those are useful foundations to preserve.

## Ranked findings

### 1. An offered comparison promotion fails at the model boundary — fix in 312

Confirmed in the actual running UI. Discovery → lab-web → Customize HTTP route
view → Run produces a method-facet/route-chart/records composition. Its default
Span status and errors option is enabled, so it is **not** directly an eligible
ordinary count cohort profile. Explicitly disabling that measure under the
population's Measures · spans and running again makes the ordinary count path
available.

After Read values, Include POST, Compare with rest and Run population view, the
child visibly offers **Include comparison in view**. Clicking it yields the actual
`.population-status` text **“Facets do not support comparison compositions yet.”**
The exact clicked path performs no durable write. This is not a hypothetical
schema concern: a reasonable investigation reaches a working comparison but
cannot preserve its learned relation using the offered action.

Use the bounded explicit conversion agreed with architecture: **Include comparison
without facet**, with **“Facet control removed; applied filters retained.”** Keep
one action, not a modal or second Save wizard. The pure transition removes only
the facet control, retains selected/baseline predicates, chart and records,
handles template removal explicitly, stages configuration and focuses Run. Never
silently drop the facet or its applied filters. Guarding the unsupported click
alone would avoid an error but leave the user's preservation task unfinished.

Tray readiness/current-owner and retired lookup behavior belong to the same
bounded task reliability cut where architecture identifies concrete failures.
They are not a separate user-facing settings initiative.

### 2. Two query editors remain visible when one population owns the work — next 313

Confirmed in live DOM and the 1280px screenshot after service Customize/Run. Above
the active Population workspace remain the root Signal/Environment/Service
controls, Run question, Filters, Measures and chart, Reusable questions/package
and a disabled Explore linked population button. The active workspace separately
has Run population view, fixed scope, Preserve, View settings and the contributed
facet/chart/records. This makes the user reason about two editing/execution
contexts at once.

The DOM also exposes Return to paused activity, Back to service, Back to question,
Back to single question and, in a child, Back to parent population. These have
real different owners, but their simultaneous presentation does not explain the
relationship. In the inspected 900px viewport the preceding editor and workspace
chrome pushed the useful route chart below the viewport; this is structural
competition, not a request for smaller font.

Consolidate around the active population canvas. Hide or collapse dormant root
editing controls while that canvas owns the visit; retain one clear scope summary
and owner-appropriate Back. Keep advanced original-question editing reachable
through a deliberate existing exit rather than leave duplicate Run controls
visible. Preserve exact query/return/focus behavior and evidence boundaries;
remove presentation duplication, not the source-owner separation.

### 3. Reused presentation titles can contradict destination grouping — include in 313

Confirmed in the actual [311 destination frame](../journeys/screenshots/311-typed-template-destination.png):
the retained literal method facet correctly shows PUT five and POST two, while the
chart groups the destination service yet keeps the authored title **HTTP server
routes**. Current template semantics correctly retain destination grouping and
literal facet field. The stale default title nevertheless tells a human the
chart is about routes.

Distinguish authored explicit titles from generated defaults and surface actual
field/scope compactly. A context-aware default or qualified subtitle can solve
this without a new title API or silently overwriting a deliberately authored
name. Handle this with canvas/context cohesion, not another template profile.

### 4. Execution vocabulary is accurate but fragmented — reconcile within the canvas

Live controls use Run question and Run population view; staged facet children in
existing 310 frames use Run selected population, while the live selected cohort
child used Run population view. Read values and Read fixed baseline are deliberate
extra source reads. Preserve separates Question, Evidence and View definition
well when opened, but authors must first understand which workspace they are in.

Keep the source distinctions. Prefer one clear primary Run for the active canvas,
with brief local facts for extra explicit reads; avoid renaming every operation
to Apply or adding confirmation prose. Existing Preserve groups correctly explain
configuration versus fresh capture and should remain collapsed until requested.
The live opened panel used roughly 577px, which is reasonable for explicit
preservation but should not become permanent above-data content.

## Incident-flow coverage and remaining limits

The running service overview supplied actual lab-web server spans, route aggregates,
compact RED/duration values and a checked Customize action. The route Customize
button lay around document Y=1629 after the overview: discoverable by scrolling,
but not a quick top-level action. Do not add another global button now; first
consolidate the canvas reached by that action.

Historical native frames demonstrate portable team definitions, exact question
opening versus current-scope presentation reuse, missing/empty cohorts, trace
waterfalls, fixed-pair persistence and shared windows. Those semantics are strong.
They do not establish that a first-time user can find the same chain unaided.
The next task should exercise one actual incident across those surfaces, not just
a component showcase.

OTel scope boundaries remain useful: HTTP defaults require reported Server and
current nonempty method; recorded Error is not business failure, native duration
is not a raw span percentile, and supplied rows are not full populations. Do not
trade those facts away while reducing prose. Put exact clocks/support in one
local context, with short visible meaning where it changes the decision.

## Finite recommendation

Complete the offered promotion and concrete owner-readiness defects in 312, with
actual failure-to-success coverage. Then deliver 313 active-canvas cohesion:
dormant root editor consolidation, owner-appropriate return and honest generated
context/title. Assess the entire incident/reopen task after that, and select the
next substantive capability from observed task failure. Defer automatic facet
expansion to every family, generic operand editors, a second catalog and another
chart vendor. No additional permanent evidence archive is needed for this review.

## 312 bounded resolution and actual acceptance

This records the original 312 resolution. [337](../PAIRED_FACETS.md) supersedes
its control-removal step: fresh fixed comparisons now retain the facet, and the
maintained 312 journey/frame follow that current contract.

The historical findings above remain the review record. The offered-promotion
failure is now corrected through the agreed explicit conversion. The actual
[comparison action frame](../journeys/screenshots/312-comparison-without-facet.png)
shows **Include comparison without facet** directly below **“Removes the facet
control; applied filters, chart and records stay unchanged.”** The staged/no-read
state and selected-only-save note remain distinct. This makes the consequence
clear without a modal, duplicate action or silent loss of predicates.

The final native verification owner reports the full isolated-store incident
journey green in 38.9 seconds, including explicit status/projection conversion to
ordinary counts, the new action retaining WHERE/List while removing only the
facet, source-free staging and admitted fixed-pair Save. Architecture separately
accepted owner readiness/lookup disposal and count-certainty corrections. These
are bounded fixes to the investigation task, not added user configuration.

All four final actual frames are visually accepted. The
[Activity incident](../journeys/screenshots/312-activity-incident.png) starts from
three readable logs and an attribute inspector; the
[correlated trace](../journeys/screenshots/312-correlated-trace.png) shows three
spans with distinct 120/30/50 ms durations and recorded errors, without claiming
complete trace coverage. The [reopened view](../journeys/screenshots/312-reopened-trace.png)
retains the three-record Waterfall, UTC, source-page facts and native expansion
controls. The fourth frame establishes the corrected preservation affordance.
Static frames show interpretation and control clarity; the native/model report
establishes effects, retention and terminal completion.

No visual blocker remains for this bounded 312 round. The dormant-editor/return
stack and generated-title/context findings remain the substantive 313 canvas
cohesion queue; they were not silently marked resolved by the promotion fix.
