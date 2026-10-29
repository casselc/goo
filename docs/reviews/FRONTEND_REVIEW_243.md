# Independent frontend review — 243

Reviewed 2026-10-29 against `65ac193`. This review owns only this file. No runtime
code, live telemetry, captures or documents were changed. It rechecks the current
roadmap, component/source-owner code and recent 239–242 deliveries, uses the root
reviewer's real-OTLP whole-task counterexample, and adds one supplied-data browser
probe against the deployed component gallery. It is not customer research, a full
accessibility audit, a cross-browser certification or a production benchmark.

## Assessment and reconciliation

The stream-first direction and passive native web components remain appropriate.
Keep the completed 239–242 scope/inspection/density/catalog work closed. Signal
switching now retains compatible scope and stages incompatible fields; Activity
has an explicit reveal/focus/return owner; row and attribute actions reuse shared
controls; saved capture search has honest metadata-versus-content semantics.
These changes address the main earlier entrance findings without introducing a
framework or universal state layer.

The next frontend gate should cover the **boundary between Activity and supporting
trace context**, followed by one bounded keyboard/update convention for dense
plots. Neither another dashboard family nor a broad CSS/component rewrite is the
most valuable next frontend task. Whole-task acceptance catches a seam that the
individual Activity inspection and catalog journeys could not.

## F1 — High task priority: trace close and parent return have separate owners

**Source-confirmed mechanism; real mobile pointer counterexample supplied by the
root reviewer.** At 390×844 after scoped Activity Logs → Spans → Open trace,
`Return to paused activity` remains in the ordinary `.workspace-header` while
`#exploreContext` is a fixed, full-viewport sheet with z-index 30. Its content
intercepts a pointer attempt on the underlying parent-return control. Root's
hit-test measured the return rect at x227/y146,147×28.5px, with a SUMMARY inside the
trace sheet at its center. The root's normal visible Close-trace-inspector →
Return-to-paused-activity workaround completes the task with no extra reads.
This is an occluded direct parent-return path, not proof that all return is
impossible.

The source explains the seam:

- `ui/product.css:513` defines a fixed z-index 30 trace inspector; the ≤700px rule at
  `:692` sets `inset:0` and `width:100vw`. `.workspace-header` has no overlay owner.
- `ui/stream-entry.mjs` remembers the Activity opener/retained view and hides that
  owner while calling `onOpenTrace`.
- `ui/app.mjs:135` does not pass that opener through to the trace owner.
- `ui/record-results.mjs:47,51` focuses its own close control, then closes only the
  sheet. When the Activity opener has no rendered rect, its caller-provided
  fallback is the fixed editor's `#exploreRun` (`ui/exploration.mjs`), not the
  retained Activity action.

The trace sheet is `role=complementary`, not a modal dialog, despite covering the
whole viewport on mobile. Source does not establish background inertness or a
contained keyboard policy. Root is separately collecting close/focus/Tab geometry;
this report does not claim a completed live focus-trap audit.

**Bounded response:** define an Activity-owned close/back intent for a trace opened
from Activity. Put a reachable parent-return action in the active sheet, or make
that invocation's Close return to its retained Activity parent. A trace opened from
fixed results must still close back to fixed results. Keep the explicit fresh-read
and broader-trace scope wording. Do not fix this by globally lifting the entire
workspace header above every overlay, or by changing all passive inspectors.

Acceptance: pointer/touch and keyboard return from the visible trace surface at
320/390/700px and desktop; retained rows, selector, inspection and opener remain
coherent; no source reads on close/return; withdrawn opener gets a visible fallback.
Exercise the trace's normal close, parent return, Escape and resize path. Declare
whether its full-screen mobile presentation is modal; users must not traverse
invisible covered controls. Keep input/disclosure Escape semantics consistent with
Activity's explicit native-control guards.

## F2 — Medium: dense plot keyboard interaction remains linear in point count

**Fresh supplied-data observation plus source.** On `/ui-components`, a valid
256-point `goo-value-timeline` created 256 ordinary button marks. Each receives the
browser's normal button tab stop. `ui/bucket-heatmap.mjs` similarly creates one
button per supplied cell; its Shift+arrow range handling does not replace ordinary
Tab traversal. This confirms the earlier keyboard-density concern, not a WCAG
failure inferred from target size or a measured failure of every chart consumer.

Recent Activity tests use direct `focus()` on an intended row/attribute before
Enter. That proves activation, visible details and return; it does not measure
sequential navigation effort from the scope bar to the desired record. Timestamp,
service-filter, Inspect, Select and trace controls can still multiply Tab stops.

**Bounded response:** first establish one shared plot-navigation policy: enter a
plot once, move a keyed active observation/cell with arrows and Home/End, explicitly
select/anchor/range, then Tab out. Include an exact-value alternative for overlapping
marks. Preserve unknown intervals and every independently supplied observation.
Do not turn every display table into a complex grid or hide all record commands.

Acceptance should include a real Tab/Shift+Tab task from a known entry control,
256-point/multiple-series and heatmap fixtures, reverse ranges, overlap, no-data,
touch scroll versus brush, and source-free selection. Record counts of deliberate
key presses for the task, not a raw count of visible DOM candidates.

## F3 — Medium: replacing supplied data loses local interaction state

**Fresh isolated browser reproduction, zero `/api` requests.** The gallery probe
focused a timeline mark, then assigned an identical `.data` value. The old mark
became disconnected and focus was not preserved. A `goo-latency` disclosure opened
and focused before an identical `.summary` assignment was closed afterward, with
focus lost. `ui/value-timeline.mjs:34–39` and `ui/components.mjs`'s GooLatency.render
replace the child tree. By contrast timeline selected-key-only updates synchronize
existing nodes, and record summary/context components already retain disclosure
openness across rendering.

This is a confirmed supplied-component update contract inconsistency. It does not
prove an observed production live-refresh disruption. The Activity latency owner
has an optional automatic cadence (`ui/activity-latency.mjs`); the intended SSE/live
component reuse makes the missing policy material before broader adoption.

**Bounded response:** define keyed-update behavior and owner policy together.
Retain focus/selection/open disclosures when the same item remains; if the active
item disappears, use an explicit nearby fallback/announcement or freeze updates
while interacting. Never acquire focus merely because passive data arrived.
Application owners keep refresh/source state and can deliberately pause live
updates during an investigation. Do not add polling to custom-element lifecycle.

Accept identical-data updates, reordered keys, insertion/removal, changed state,
detached assignment/reconnect, retained disclosure and an interaction during an
actual owner's refresh. Combine with F2 only where it reduces duplicated key/focus
logic; avoid a general reactive rendering framework.

## Whole-product coherence and remaining limits

- Preservation outcomes should remain distinct: current live selection, a fresh
  fixed-question capture, authored reuse and an investigation document. The new
  catalog solves finding stored captures; it does not make an optional exact-item
  retention capability equivalent to a baseline bookmark. Reuse existing owners
  and destinations when refining those labels/actions; do not introduce another
  generic Save toolbar.
- The named-environment prerequisite still limits ordinary OTel adoption. UI
  diagnostics/discovery and an explicit unreported-scope design need product and
  source-adapter agreement. An empty result must not imply the source is empty,
  and blank environment must not silently become a wildcard.
- The separate legacy trace/surrounding-context shell is the most relevant shared
  primitive opportunity. Extract its layout/close/focus intent only when fixing a
  named invocation, preserving each application's source owner. The existing
  passive inspector and retained-resource-return contracts should not be undone.
- No current startup/module-count, throughput, heap or frame-time benchmark was
  performed. Earlier 137-module/DOM counts must not be relabeled current evidence.
  A global bundler/cache/framework migration is unsupported by this review.
- Chromium is the only browser probed here. Existing desktop/mobile journeys and
  the prior component audit remain supporting history, not new conformance runs.

## Suggested frontend sequence

1. Fix and review the visible trace → parent return contract (F1) as part of the
   complete Activity → supporting signals → preserve → find/reopen task.
2. Address the adoption/preservation product decisions alongside bounded core
   correctness fixes; they change what users can accomplish, not only layout.
3. Implement one keyed plot keyboard/update contract (F2/F3) before multiplying
   live widgets or claiming general SSE-ready interactions.
4. Continue comparison support and reusable trace-question work against named
   tasks. Add components only when that task needs them; retain the current passive
   display/app-owner split and explicit source-read accounting.
